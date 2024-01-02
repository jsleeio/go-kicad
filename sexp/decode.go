package sexp

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Decode populates a struct passed in t (which must be a pointer to a struct)
// if (and only if) the top-level file type keyword matches that given in
// typeName.
//
// Kicad file formats all start with a top-level tuple whose first value
// gives the file type and whose remaining values are key/value tuples. this
// function is the main way to parse such a file into a struct.
func Decode(r io.Reader, typeName string, t interface{}) error {
	v := reflect.ValueOf(t)

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return &InvalidDecodeError{v.Type()}
	}

	v = decodeIndirect(v)

	if v.Type().Kind() != reflect.Struct {
		return fmt.Errorf("Decode target must be pointer to struct, not %s", v.Type())
	}

	s := NewScanner(r)

	open := s.Read()
	if open.Type != LEFT {
		return fmt.Errorf("must start with LEFT; got %s", open.Type)
	}

	typeTok := s.Read()
	if typeTok.Type != RAW_STRING {
		return fmt.Errorf("first element must be RAW_STRING; got %s", typeTok.Type)
	}

	if typeTok.Data != typeName {
		return fmt.Errorf("want filetype %q but got %q", typeName, typeTok.Data)
	}

	err := decodeSequenceIntoStruct(s, v, RIGHT)
	if err != nil {
		return err
	}
	s.Read() // consume closing paren
	return nil
}

// DecodeSimple writes a single value based on a sequence read from the given
// reader. It can be used to decode isolated values, but Decode must be used
// to decode the usual kicad convention of having a top-level tuple that is
// a type name followed by a sequence of fields.
func DecodeSimple(r io.Reader, t interface{}) error {
	s := NewScanner(r)
	return decodeIntoValue(s, reflect.ValueOf(t))
}

func decodeIntoValue(s *Scanner, v reflect.Value) error {
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return &InvalidDecodeError{v.Type()}
	}

	v = decodeIndirect(v)

	switch v.Kind() {
	case reflect.String:
		return decodeString(s, v)
	case reflect.Int, reflect.Uint:
		return decodeInt(s, v)
	case reflect.Bool:
		return decodeBool(s, v)
	case reflect.Float64:
		return decodeFloat(s, v)
	case reflect.Slice:
		return decodeSlice(s, v)
	case reflect.Map:
		return decodeMap(s, v)
	case reflect.Struct:
		return decodeStruct(s, v)
	default:
		return &InvalidDecodeError{v.Type()}
	}
}

// decodeSkip skips the next value, leaving the scanner pointing at the
// beginning of the following value. If the next value is a tuple then the
// entire tuple (including any nested tuples) is skipped.
func decodeSkip(s *Scanner) error {
	next := s.Peek()

	if next.Type == LEFT {
		// We need to count open/close parens until we get back to
		// our initial nesting level.
		nest := 0
		for {
			token := s.Read()
			switch token.Type {
			case LEFT:
				nest++
			case RIGHT:
				nest--
				if nest == 0 {
					return nil
				}
			case EOF:
				return fmt.Errorf("unexpected EOF while skipping tuple")
			}
		}
	}

	if next.Type == RIGHT || next.Type == EOF {
		return fmt.Errorf("no value to skip! found %s", next.Type)
	}

	s.Read() // consume single-token value

	return nil
}

func decodeString(s *Scanner, v reflect.Value) error {
	next := s.Peek()

	switch next.Type {
	case RAW_STRING:
		v.SetString(next.Data)
	case QUOTE_STRING:
		str, err := unquoteString(next.Data)
		if err != nil {
			return err
		}
		v.SetString(str)
	default:
		return fmt.Errorf(
			"unexpected %s while decoding into string",
			next.Type,
		)
	}

	s.Read() // consume the token

	return nil
}

func decodeInt(s *Scanner, v reflect.Value) error {
	next := s.Peek()

	switch next.Type {
	case RAW_STRING:
		switch v.Kind() {
		// TODO: kicad additionally supports exponents
		case reflect.Int:
			val, err := strconv.ParseInt(next.Data, 10, 64)
			if err != nil {
				return err
			}
			v.SetInt(val)
		case reflect.Uint:
			base := 10
			// kicad sometimes emits numeric values in hex, with underscores
			token := strings.ReplaceAll(next.Data, "_", "")
			if strings.HasPrefix(token, "0x") {
				base = 16
				token = strings.TrimPrefix(token, "0x")
			}
			val, err := strconv.ParseUint(token, base, 64)
			if err != nil {
				return err
			}
			v.SetUint(val)
		default:
			// should never happen because this is handled in caller
			panic("invalid decodeInt target")
		}
	default:
		return fmt.Errorf(
			"unexpected %s while decoding into int",
			next.Type,
		)
	}

	s.Read() // consume the token

	return nil
}

func decodeFloat(s *Scanner, v reflect.Value) error {
	next := s.Peek()

	switch next.Type {
	case RAW_STRING:
		val, err := strconv.ParseFloat(next.Data, 64)
		if err != nil {
			return err
		}
		v.SetFloat(val)
	default:
		return fmt.Errorf(
			"unexpected %s while decoding into float",
			next.Type,
		)
	}

	s.Read() // consume the token

	return nil
}

func decodeBool(s *Scanner, v reflect.Value) error {
	next := s.Peek()

	switch next.Type {
	case RAW_STRING:
		val, err := strconv.ParseBool(next.Data)
		if err != nil {
			return err
		}
		v.SetBool(val)
	default:
		return fmt.Errorf(
			"unexpected %s while decoding into bool",
			next.Type,
		)
	}

	s.Read() // consume the token

	return nil
}

func decodeSlice(s *Scanner, v reflect.Value) error {
	next := s.Peek()
	if next.Type != LEFT {
		return fmt.Errorf(
			"slice value cannot begin with %s", next.Type,
		)
	}
	s.Read() // consume parenthesis

	empty := reflect.MakeSlice(v.Type(), 0, 2)
	v.Set(empty)

	err := decodeSequenceIntoSlice(s, v, RIGHT)
	if err != nil {
		return err
	}
	s.Read() // consume closing paren
	return nil
}

func decodeSequenceIntoSlice(s *Scanner, v reflect.Value, endType TokenType) error {
	ret := v

	elemType := v.Type().Elem()
	for {
		next := s.Peek()
		if next.Type == endType {
			break
		}
		if next.Type == EOF {
			return fmt.Errorf(
				"unexpected EOF while decoding slice value",
			)
		}

		elem := reflect.New(elemType)
		err := decodeIntoValue(s, elem)
		if err != nil {
			return err
		}

		ret = reflect.Append(ret, elem.Elem())
	}

	v.Set(ret)

	return nil
}

func decodeMap(s *Scanner, v reflect.Value) error {
	next := s.Peek()
	if next.Type != LEFT {
		return fmt.Errorf(
			"map value cannot begin with %s", next.Type,
		)
	}
	s.Read() // consume parenthesis

	ret := reflect.MakeMap(v.Type())

	keyType := v.Type().Key()
	valType := v.Type().Elem()
	for {
		next := s.Peek()
		if next.Type == RIGHT {
			s.Read() // consume parenthesis
			break
		}
		if next.Type == EOF {
			return fmt.Errorf(
				"unexpected EOF while decoding slice value",
			)
		}

		if next.Type != LEFT {
			return fmt.Errorf(
				"map entry must be tuple, but got %s", next.Type,
			)
		}

		s.Read() // consume open paren

		key := reflect.New(keyType)
		val := reflect.New(valType)

		err := decodeIntoValue(s, key)
		if err != nil {
			return err
		}

		if s.Peek().Type == RIGHT {
			return fmt.Errorf("map entry tuples must have two elements")
		}
		if s.Peek().Type == EOF {
			return fmt.Errorf("unexpected EOF while decoding map entry")
		}

		err = decodeIntoValue(s, val)
		if err != nil {
			return err
		}

		if s.Peek().Type != RIGHT {
			return fmt.Errorf("map entry tuples must have two elements")
		}
		s.Read() // Consume closing paren

		ret.SetMapIndex(key.Elem(), val.Elem())
	}

	v.Set(ret)

	return nil
}

func decodeStruct(s *Scanner, v reflect.Value) error {
	next := s.Peek()
	if next.Type != LEFT {
		return fmt.Errorf(
			"struct value cannot begin with %s", next.Type,
		)
	}
	s.Read() // consume parenthesis

	ty := v.Type()
	ret := reflect.New(ty)
	v.Set(ret.Elem())

	err := decodeSequenceIntoStruct(s, v, RIGHT)
	if err != nil {
		return err
	}
	s.Read() // consume closing paren
	return nil
}

func decodeSequenceIntoStruct(s *Scanner, v reflect.Value, endType TokenType) error {
	ty := v.Type()
	type Field struct {
		Index int
		Flat  bool
		Multi bool
	}

	var posFields []*Field
	nameFields := make(map[string]*Field)
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)
		tag, tagSet := field.Tag.Lookup("kicad")
		if !tagSet {
			continue
		}

		parts := strings.Split(tag, ",")
		key := parts[0]
		flags := parts[1:]
		fieldDef := &Field{
			Index: i,
		}
		for _, flag := range flags {
			switch flag {
			case "flat":
				fieldDef.Flat = true
			case "multi":
				fieldDef.Multi = true
			default:
				return fmt.Errorf(
					"invalid kicad decode flag %q on %s",
					flag, field.Name,
				)
			}
		}

		chkType := field.Type
		if fieldDef.Multi {
			if chkType.Kind() != reflect.Slice {
				return fmt.Errorf("'multi' flag used on non-slice field %s", field.Name)
			}
			chkType = chkType.Elem()
		}

		if fieldDef.Flat {
			kind := chkType.Kind()
			if kind != reflect.Slice && kind != reflect.Struct {
				return fmt.Errorf("'flat' flag cannot be used on non-slice, non-struct field %s", field.Name)
			}
		}

		if key == "" {
			posFields = append(posFields, fieldDef)
		} else {
			nameFields[key] = fieldDef
		}

	}

	for {
		next := s.Peek()
		if next.Type == endType {
			break
		}
		if next.Type == EOF {
			return fmt.Errorf("unexpected EOF decoding struct value")
		}

		var fieldDef *Field
		needClose := false
		if len(posFields) > 0 {
			fieldDef = posFields[0]
			posFields = posFields[1:]
		} else {
			if next.Type != LEFT {
				return fmt.Errorf(
					"named struct field must start with LEFT, but got %s",
					next.Type,
				)
			}
			s.Read() // consume parenthesis

			label := s.Peek()
			if label.Type != RAW_STRING {
				return fmt.Errorf(
					"struct name must be RAW_STRING, but got %s",
					label.Type,
				)
			}
			s.Read() // consume label

			fieldDef = nameFields[label.Data]
			needClose = true
		}

		var fieldValue reflect.Value
		var fieldType reflect.Type
		var valType reflect.Type
		var tv reflect.Value

		if fieldDef == nil {
			err := decodeSkip(s)
			if err != nil {
				return err
			}
			goto Done
		}

		fieldValue = v.Field(fieldDef.Index)
		fieldType = fieldValue.Type()

		valType = fieldType
		if fieldDef.Multi {
			// Multi fields are slices of the target type, which we
			// append to for each new instance. Thus our value type
			// is the slice's element type.
			valType = fieldType.Elem()
		}

		tv = reflect.New(valType)

		if fieldDef.Flat {
			// For "Flat" we are expecting the elements of a slice or the
			// fields of a struct to appear directly after the field name,
			// without an additional wrapping tuple.
			switch valType.Kind() {
			case reflect.Struct:
				err := decodeSequenceIntoStruct(s, tv.Elem(), RIGHT)
				if err != nil {
					return err
				}
			case reflect.Slice:
				err := decodeSequenceIntoSlice(s, tv.Elem(), RIGHT)
				if err != nil {
					return err
				}
			default:
				// Should never happen due to validation above
				panic("non-slice and non-struct flat target")
			}
		} else {
			err := decodeIntoValue(s, tv)
			if err != nil {
				return err
			}
		}

		if fieldDef.Multi {
			fieldValue.Set(reflect.Append(fieldValue, tv.Elem()))
		} else {
			fieldValue.Set(tv.Elem())
		}

	Done:

		if needClose {
			close := s.Read()
			if close.Type != RIGHT {
				return fmt.Errorf(
					"missing closing paren for struct tuple; got %s",
					close.Type,
				)
			}
		}
	}

	if len(posFields) > 0 {
		// If there is one remaining field and it is a "flat" field then
		// this is acceptable since it is allowed to "consume" the zero
		// remaining values.
		if len(posFields) != 1 || !posFields[0].Flat {
			return fmt.Errorf(
				"insufficient values for positional fields %#v",
				posFields,
			)
		}
	}

	return nil
}

// decodeIndirect deals with pointer values by allocating pointers as
// needed to reach the final value.
func decodeIndirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}

	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && e.Elem().Kind() == reflect.Ptr {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}

	return v
}

func unquoteString(raw string) (string, error) {
	ret := make([]byte, 0, len(raw)-2)

	// Trim off the enclosing quote markers first
	raw = raw[1 : len(raw)-1]

	for ; len(raw) > 0; raw = raw[1:] {
		switch raw[0] {
		case '\\':
			// we should be guaranteed that a lone backslash never occurs at
			// the end of the string, since otherwise the scanner would've
			// treated it as escaping the closing quote.
			switch raw[1] {
			case 'a':
				ret = append(ret, 0x07)
			case 'b':
				ret = append(ret, 0x08)
			case 'f':
				ret = append(ret, 0x0c)
			case 'n':
				ret = append(ret, '\n')
			case 'r':
				ret = append(ret, '\r')
			case 't':
				ret = append(ret, '\t')
			case 'v':
				ret = append(ret, 0x0b)
			case 'x':
				// Hex character escape requires at least one more digit
				if len(raw) < 3 || !isHexDigit(raw[2]) {
					ret = append(ret, '\\')
					continue
				}

				var digits string
				if len(raw) > 3 && isHexDigit(raw[3]) {
					digits = raw[2:4]
				} else {
					digits = raw[2:3]
				}

				val, _ := strconv.ParseUint(digits, 16, 16)
				ret = append(ret, byte(val))

				raw = raw[len(digits):]
			case '0', '1', '2', '3', '4', '5', '6', '7':
				// Octal character escape requires at least one more digit
				if len(raw) < 2 || !isOctalDigit(raw[1]) {
					ret = append(ret, '\\')
					continue
				}

				var digits string
				if len(raw) > 3 && isOctalDigit(raw[2]) && isOctalDigit(raw[3]) {
					digits = raw[1:4]
				} else if len(raw) > 2 && isOctalDigit(raw[2]) {
					digits = raw[1:3]
				} else {
					digits = raw[1:2]
				}

				val, err := strconv.ParseUint(digits, 8, 16)
				if err != nil {
					// should never happen
					panic(err)
				}
				ret = append(ret, byte(val))

				raw = raw[len(digits)-1:]
			default:
				// Treat invalid escapes as literal backslashes
				ret = append(ret, '\\')
				continue
			}
			raw = raw[1:]
		default:
			ret = append(ret, raw[0])
		}
	}

	return string(ret), nil
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f')
}

func isOctalDigit(b byte) bool {
	return (b >= '0' && b <= '7')
}

// InvalidDecodeError is an error that indicates that a given type is not
// a valid target for a decode.
type InvalidDecodeError struct {
	Type reflect.Type
}

func (e *InvalidDecodeError) Error() string {
	if e.Type == nil {
		return "kicad sexp: can't decode into nil"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "kicad sexp: can't decode into " + e.Type.String()
	}
	return "kicad sexp: can't decode into nil " + e.Type.String() + ")"
}

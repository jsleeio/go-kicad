package sexp

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

func Decode(r io.Reader, t interface{}) error {
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
	default:
		return &InvalidDecodeError{v.Type()}
	}
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
			val, err := strconv.ParseUint(next.Data, 10, 64)
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

	ret := reflect.MakeSlice(v.Type(), 0, 2)

	elemType := v.Type().Elem()
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

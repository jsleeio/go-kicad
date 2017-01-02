package sexp

import (
	"errors"
	"io"
)

// Writer is a low-level utility for writing KiCad S-Expression files.
// It could be considered as the writing equivalent of Scanner, appending
// to the given writer in terms of the raw tokens though with some basic
// smarts to produce human-friendly indentation.
type Writer struct {
	w      io.Writer
	parens int
	indent int

	nextDelim       delimType
	writtenOneValue bool
}

type delimType int

const (
	delimNone   delimType = 0
	delimSpace  delimType = 1
	delimIndent delimType = 2
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// BeginTuple writes the open parenthesis that begins a tuple.
//
// An error is returned if the underlying byte writer signals an error.
func (w *Writer) BeginTuple() error {
	if w.indent > 0 {
		err := w.newline()
		if err != nil {
			return err
		}
	} else if w.writtenOneValue {
		return errors.New("can't begin a second top-level value")
	}
	err := w.delimiter()
	if err != nil {
		return err
	}
	_, err = w.w.Write([]byte{'('})
	if err == nil {
		w.indent++
		w.parens++
	}
	w.nextDelim = delimNone
	return err
}

// EndTuple writes the closing parenthesis that ends a tuple.
//
// An error is returned if there aren't any open tuples to close or if the
// underlying byte writer signals an error.
func (w *Writer) EndTuple() error {
	if w.nextDelim != delimSpace {
		w.delimiter()
	}
	_, err := w.w.Write([]byte{')'})
	if err != nil {
		return err
	}
	w.nextDelim = delimSpace
	w.indent--
	w.parens--
	if w.parens < 0 {
		return errors.New("unbalanced tuple delimiters")
	}
	w.recordWrittenOneValue()
	return err
}

// WriteRawString writes the given string to the document verbatim as a
// raw string token. It's the callers responsibility to ensure that the
// given string is valid as a raw string, which includes making sure It
// doesn't include any whitespace or parenthesis characters.
func (w *Writer) WriteRawString(str string) error {
	if w.writtenOneValue {
		return errors.New("can't begin a second top-level value")
	}

	err := w.delimiter()
	if err != nil {
		return err
	}

	_, err = w.w.Write([]byte(str))
	if err != nil {
		return err
	}
	w.nextDelim = delimSpace
	w.recordWrittenOneValue()
	return nil
}

func (w *Writer) WriteQuoteString(str string) error {
	if w.writtenOneValue {
		return errors.New("can't begin a second top-level value")
	}

	err := w.delimiter()
	if err != nil {
		return err
	}

	_, err = w.w.Write([]byte{'"'})
	if err != nil {
		return err
	}
	for i := 0; i < len(str); i++ {
		ch := str[i]
		var wrb []byte
		switch ch {
		case '\n':
			wrb = []byte{'\\', 'n'}
		case '\r':
			wrb = []byte{'\\', 'r'}
		case '\t':
			wrb = []byte{'\\', 't'}
		case '\v':
			wrb = []byte{'\\', 'v'}
		case '"':
			wrb = []byte{'\\', '"'}
		case '\\':
			wrb = []byte{'\\', '\\'}
		default:
			wrb = []byte{ch}
		}
		_, err = w.w.Write(wrb)
		if err != nil {
			return err
		}
	}
	_, err = w.w.Write([]byte{'"'})
	w.nextDelim = delimSpace
	w.recordWrittenOneValue()
	return err
}

// WriteString writes the given string as a raw string if possible or as
// a quoted string otherwise.
func (w *Writer) WriteString(str string) error {
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case 10, 13, 32, 8, 0, '(', ')', '#':
			return w.WriteQuoteString(str)
		}
	}

	return w.WriteRawString(str)
}

// WriteToken writes a token produced by the scanner. The token is assumed
// to be something the scanner would produce, so if a caller is manually
// constructing the token it's the caller's responsibility to ensure that it
// is valid and consistent with the scanner's behavior.
func (w *Writer) WriteToken(token Token) error {
	switch token.Type {
	case LEFT:
		return w.BeginTuple()
	case RIGHT:
		return w.BeginTuple()
	default:
		// Since we know that token.Data is valid token data, we can just
		// write it out as a raw string. Only the left and right parens
		// need to be treated as special because we want to keep track of
		// our indent level for formatting purposes.
		return w.WriteRawString(token.Data)
	}
}

// Close signals the end of the s-expression structure.
//
// An error is returned if there are any unclosed tuples or if the underlying
// byte writer signals an error when asked to close.
func (w *Writer) Close() error {
	if w.parens > 0 {
		return errors.New("unbalanced tuple delimiters")
	}
	if !w.writtenOneValue {
		return errors.New("no value written")
	}
	if closer, ok := w.w.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (w *Writer) newline() error {
	// Don't newline if we're already at the start of a line
	if w.nextDelim == delimIndent {
		return nil
	}

	_, err := w.w.Write([]byte{'\n'})
	if err != nil {
		return err
	}

	w.nextDelim = delimIndent

	return nil
}

func (w *Writer) delimiter() error {
	switch w.nextDelim {
	case delimNone:
		return nil
	case delimSpace:
		_, err := w.w.Write([]byte{' '})
		if err != nil {
			return err
		}
	case delimIndent:
		for i := 0; i < w.indent; i++ {
			_, err := w.w.Write([]byte{' ', ' '})
			if err != nil {
				return err
			}
		}
	}

	w.nextDelim = delimNone

	return nil
}

func (w *Writer) recordWrittenOneValue() {
	if w.parens == 0 {
		w.writtenOneValue = true
	}
}

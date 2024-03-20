// Package sexp implements a scanner and decoder for S-expressions as used in
// various kinds of Kicad files.
//
// It does *not* provide Go structs matching the Kicad file formats; those are
// elsewhere.
package sexp

import (
	"bufio"
	"fmt"
	"io"
)

// Scanner contains the state for an S-expression token scanner.
type Scanner struct {
	s      *bufio.Scanner
	peeked *Token
	lines  int
	eof    bool
}

// Token contains information about a single S-expression token, such as left
// or right parentheses, a quoted string, a number, and so on.
type Token struct {
	Type TokenType
	Data string
}

// TokenType stores a token type identifier
//
//go:generate stringer -type=TokenType
type TokenType rune

const (
	// RAWSTRING tokens are strings that aren't quoted. Example: abc-def-123
	RAWSTRING TokenType = 'B'
	// QUOTESTRING tokens are strings that are quoted. Example: "abc-def-123"
	QUOTESTRING TokenType = 'Q'

	// NUMBER tokens are numbers in hexadecimal, decimal or floating point forms.
	// Examples: 0xdeadface, 123, -1, 3.1459
	NUMBER TokenType = 'N'
	// RIGHT tokens are a single closing parenthesis: )
	RIGHT TokenType = ')'
	// LEFT tokens are a single opening parenthesis: (
	LEFT TokenType = '('

	// EOF tokens are a synthetic token that indicates end of stream
	EOF TokenType = '␄'

	// INVALID tokens indicate an error state
	INVALID TokenType = '�'
)

// GoString returns the name of the Go constant indicating a token type
func (t TokenType) GoString() string {
	return fmt.Sprintf("sexp.%s", t.String())
}

type scanError byte

// Error() returns a specialized error from the Scanner
func (b scanError) Error() string {
	return fmt.Sprintf("invalid byte %q", byte(b))
}

// NewScanner creates a new scanner that finds tokens in the given reader.
func NewScanner(r io.Reader) *Scanner {
	ret := &Scanner{
		lines: 1,
		s:     bufio.NewScanner(r),
	}
	ret.s.Split(ret.findToken)
	return ret
}

// Peek find the next token in the stream and returns it without consuming it.
// Subsequent calls to Peek will return the same token until Read is called,
// which will then consume the token and allow a new token to be peeked.
func (s *Scanner) Peek() Token {
	if s.peeked != nil {
		return *s.peeked
	}
	if s.eof {
		return Token{
			Type: EOF,
			Data: "",
		}
	}

	data := ""
	for data == "" {
		if !s.s.Scan() {
			s.eof = true

			if s.s.Err() != nil {
				invData := ""
				if inv, ok := s.s.Err().(scanError); ok {
					invData = string(inv)
				}
				return Token{
					Type: INVALID,
					Data: invData,
				}
			}

			// Now that we have s.eof set, we can call this function
			// recursively to get the EOF token immediately, since we have nothing
			// else to return here.
			return s.Peek()
		}

		// Text might still be empty if we're skipping whitespace/comments, so
		// we keep trying until we get something non-empty or until we hit EOF.
		data = s.s.Text()
	}

	// Classify the token
	tokenType := RAWSTRING
	switch {
	case data == "(" || data == ")":
		// The parens are their own token value, so we can cheat here
		tokenType = TokenType(data[0])
	case data[0] == '"':
		tokenType = QUOTESTRING
	}

	s.peeked = &Token{
		Type: tokenType,
		Data: data,
	}

	return *s.peeked
}

// Read returns the next token from the stream, if available
func (s *Scanner) Read() Token {
	token := s.Peek()
	if token.Type != EOF {
		s.peeked = nil
	}
	return token
}

// findToken implements the bufio.SplitFunc interface for lexing S-expressions
func (s *Scanner) findToken(data []byte, eof bool) (advance int, token []byte, err error) {
	{
		size, skipData, skipErr := s.scanIrrelevant(data, eof)
		if size != 0 || skipData != nil || skipErr != nil {
			return size, []byte{}, skipErr
		}
	}
	if len(data) == 0 {
		return 0, nil, nil
	}
	next := data[0]
	switch {
	case next == '(' || next == ')':
		return 1, data[:1], nil
	case next == '"':
		return s.scanString(data, eof)
	default:
		// Everything else is treated as a raw token
		return s.scanRaw(data, eof)
	}
}

func (s *Scanner) scanIrrelevant(data []byte, eof bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		// tell the scanner to read more data
		return 0, nil, nil
	}
	switch data[0] {
	case 10, 13, 32, 9, 0:
		return s.scanWhitespace(data, eof)
	case '#':
		return s.scanComment(data, eof)
	}
	return 0, nil, nil
}

func (s *Scanner) scanWhitespace(data []byte, eof bool) (int, []byte, error) {
	size := 0
	b := data
Bytes:
	for {
		if len(b) == 0 {
			break Bytes
		}

		next := b[0]
		b = b[1:]

		switch next {
		case 10:
			size++
			s.lines++
		case 0, 9, 13, 32:
			size++
		default:
			break Bytes
		}
	}
	return size, nil, nil
}

func (s *Scanner) scanComment(data []byte, eof bool) (int, []byte, error) {
	size := 1
	b := data[1:] // skip initial # symbol
Bytes:
	for {
		if len(b) == 0 {
			break Bytes
		}

		next := b[0]
		b = b[1:]
		size++

		switch next {
		case 10, 13:
			break Bytes
		}
	}

	return size, nil, nil
}

func (s *Scanner) scanString(data []byte, eof bool) (int, []byte, error) {
	advance := 1
	b := data[1:]

	escape := false
Bytes:
	for {
		if len(b) == 0 {
			if eof {
				return 0, nil, fmt.Errorf("line %d: unexpected EOF in string %q", s.lines, data)
			}

			// Request more bytes
			return 0, nil, nil
		}

		next := b[0]
		advance++
		b = b[1:]

		switch {
		case escape:
			escape = false
		case next == '\\':
			escape = true
		case next == '"':
			// Done!
			break Bytes
		}
	}

	return advance, data[:advance], nil
}

func (s *Scanner) scanRaw(data []byte, eof bool) (int, []byte, error) {
	advance := 0
	b := data

Bytes:
	for {
		if len(b) == 0 {
			if eof {
				break Bytes
			} else {
				// Request more bytes
				return 0, nil, nil
			}
		}

		next := b[0]
		b = b[1:]

		switch next {
		case 10, 13, 32, 9, 0, '(', ')', '#':
			break Bytes
		}

		advance++
	}

	return advance, data[:advance], nil
}

package sexp

import (
	"reflect"
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	tests := []struct {
		Input string
		Want  []Token
	}{
		{
			``,
			[]Token{
				{EOF, ""},
			},
		},
		{
			`    `,
			[]Token{
				{EOF, ""},
			},
		},
		{
			`# comment`,
			[]Token{
				{EOF, ""},
			},
		},
		{
			"# comment\n#\n#comment",
			[]Token{
				{EOF, ""},
			},
		},
		{
			`()`,
			[]Token{
				{LEFT, `(`},
				{RIGHT, `)`},
				{EOF, ""},
			},
		},
		{
			`""`,
			[]Token{
				{QUOTE_STRING, `""`},
				{EOF, ""},
			},
		},
		{
			`"hello"`,
			[]Token{
				{QUOTE_STRING, `"hello"`},
				{EOF, ""},
			},
		},
		{
			`"hello\nworld"`,
			[]Token{
				{QUOTE_STRING, `"hello\nworld"`},
				{EOF, ""},
			},
		},
		{
			`"hello\xffworld"`,
			[]Token{
				{QUOTE_STRING, `"hello\xffworld"`},
				{EOF, ""},
			},
		},
		{
			`"hello\"world"`,
			[]Token{
				{QUOTE_STRING, `"hello\"world"`},
				{EOF, ""},
			},
		},
		{
			`baz`,
			[]Token{
				{RAW_STRING, `baz`},
				{EOF, ""},
			},
		},
		{
			`Resistors_SMD:R_1206_HandSoldering`,
			[]Token{
				{RAW_STRING, `Resistors_SMD:R_1206_HandSoldering`},
				{EOF, ""},
			},
		},
		{
			`(foo (bar "baz") (boz 12))`,
			[]Token{
				{LEFT, `(`},
				{RAW_STRING, `foo`},
				{LEFT, `(`},
				{RAW_STRING, `bar`},
				{QUOTE_STRING, `"baz"`},
				{RIGHT, `)`},
				{LEFT, `(`},
				{RAW_STRING, `boz`},
				{RAW_STRING, `12`},
				{RIGHT, `)`},
				{RIGHT, `)`},
				{EOF, ""},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			scanner := NewScanner(strings.NewReader(test.Input))
			got := make([]Token, 0, 8)
			for {
				token := scanner.Read()
				got = append(got, token)
				if token.Type == EOF {
					break
				}
			}
			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf(
					"incorrect token stream\ngot:  %#v\nwant: %#v",
					got, test.Want,
				)
			}
		})
	}
}

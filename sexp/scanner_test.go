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
				{QUOTESTRING, `""`},
				{EOF, ""},
			},
		},
		{
			`"hello"`,
			[]Token{
				{QUOTESTRING, `"hello"`},
				{EOF, ""},
			},
		},
		{
			`"hello\nworld"`,
			[]Token{
				{QUOTESTRING, `"hello\nworld"`},
				{EOF, ""},
			},
		},
		{
			`"hello\xffworld"`,
			[]Token{
				{QUOTESTRING, `"hello\xffworld"`},
				{EOF, ""},
			},
		},
		{
			`"hello\"world"`,
			[]Token{
				{QUOTESTRING, `"hello\"world"`},
				{EOF, ""},
			},
		},
		{
			`baz`,
			[]Token{
				{RAWSTRING, `baz`},
				{EOF, ""},
			},
		},
		{
			`Resistors_SMD:R_1206_HandSoldering`,
			[]Token{
				{RAWSTRING, `Resistors_SMD:R_1206_HandSoldering`},
				{EOF, ""},
			},
		},
		{
			" (foo ( bar \"baz\" ) (boz 12 ) ) ",
			[]Token{
				{LEFT, `(`},
				{RAWSTRING, `foo`},
				{LEFT, `(`},
				{RAWSTRING, `bar`},
				{QUOTESTRING, `"baz"`},
				{RIGHT, `)`},
				{LEFT, `(`},
				{RAWSTRING, `boz`},
				{RAWSTRING, `12`},
				{RIGHT, `)`},
				{RIGHT, `)`},
				{EOF, ""},
			},
		},
		{
			"\t(foo\t(\tbar\t\"baz\"\t)\t(boz\t12\t)\t)\t",
			[]Token{
				{LEFT, `(`},
				{RAWSTRING, `foo`},
				{LEFT, `(`},
				{RAWSTRING, `bar`},
				{QUOTESTRING, `"baz"`},
				{RIGHT, `)`},
				{LEFT, `(`},
				{RAWSTRING, `boz`},
				{RAWSTRING, `12`},
				{RIGHT, `)`},
				{RIGHT, `)`},
				{EOF, ""},
			},
		},
		{
			"\n  (foo\n  (\n  bar\n  \"baz\"\n  )\n  (boz\n  12\n  )\n  )\n  ",
			[]Token{
				{LEFT, `(`},
				{RAWSTRING, `foo`},
				{LEFT, `(`},
				{RAWSTRING, `bar`},
				{QUOTESTRING, `"baz"`},
				{RIGHT, `)`},
				{LEFT, `(`},
				{RAWSTRING, `boz`},
				{RAWSTRING, `12`},
				{RIGHT, `)`},
				{RIGHT, `)`},
				{EOF, ""},
			},
		},
		{
			"\n(foo\n(\nbar\n\"baz\"\n)\n(boz\n12\n)\n)\n",
			[]Token{
				{LEFT, `(`},
				{RAWSTRING, `foo`},
				{LEFT, `(`},
				{RAWSTRING, `bar`},
				{QUOTESTRING, `"baz"`},
				{RIGHT, `)`},
				{LEFT, `(`},
				{RAWSTRING, `boz`},
				{RAWSTRING, `12`},
				{RIGHT, `)`},
				{RIGHT, `)`},
				{EOF, ""},
			},
		},
		{
			`(foo (bar "baz") (boz 12))`,
			[]Token{
				{LEFT, `(`},
				{RAWSTRING, `foo`},
				{LEFT, `(`},
				{RAWSTRING, `bar`},
				{QUOTESTRING, `"baz"`},
				{RIGHT, `)`},
				{LEFT, `(`},
				{RAWSTRING, `boz`},
				{RAWSTRING, `12`},
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

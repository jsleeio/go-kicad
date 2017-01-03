package sexp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestDecode_valid(t *testing.T) {
	sptr := func(s string) *string {
		return &s
	}
	iptr := func(i int) *int {
		return &i
	}
	uiptr := func(i uint) *uint {
		return &i
	}
	bptr := func(i bool) *bool {
		return &i
	}
	fptr := func(i float64) *float64 {
		return &i
	}
	slicePtr := func(s []string) *[]string {
		return &s
	}
	sliceSlicePtr := func(s [][]string) *[][]string {
		return &s
	}

	tests := []struct {
		Input  string
		Target interface{}
		Want   interface{}
	}{
		{
			Input:  `hello`,
			Target: sptr(""),
			Want:   sptr("hello"),
		},
		{
			Input:  `"hello"`,
			Target: sptr(""),
			Want:   sptr("hello"),
		},
		{
			Input:  `"hello\nworld"`,
			Target: sptr(""),
			Want:   sptr("hello\nworld"),
		},
		{
			Input:  `"hello\r\nworld"`,
			Target: sptr(""),
			Want:   sptr("hello\r\nworld"),
		},
		{
			Input:  `"hello\tworld"`,
			Target: sptr(""),
			Want:   sptr("hello\tworld"),
		},
		{
			Input:  `"hello\x20world"`,
			Target: sptr(""),
			Want:   sptr("hello world"),
		},
		{
			Input:  `"hello\40world"`,
			Target: sptr(""),
			Want:   sptr("hello world"),
		},
		{
			Input:  `"hello\040world"`,
			Target: sptr(""),
			Want:   sptr("hello world"),
		},
		{
			Input:  `500`,
			Target: sptr(""),
			Want:   sptr("500"),
		},
		{
			Input:  `true`,
			Target: sptr(""),
			Want:   sptr("true"),
		},
		{
			Input:  `500`,
			Target: iptr(0),
			Want:   iptr(500),
		},
		{
			Input:  `-500`,
			Target: iptr(0),
			Want:   iptr(-500),
		},
		{
			Input:  `500`,
			Target: uiptr(0),
			Want:   uiptr(500),
		},
		{
			Input:  `true`,
			Target: bptr(false),
			Want:   bptr(true),
		},
		{
			Input:  `false`,
			Target: bptr(true),
			Want:   bptr(false),
		},
		{
			Input:  `1.2`,
			Target: fptr(0.0),
			Want:   fptr(1.2),
		},
		{
			Input:  `-1.2`,
			Target: fptr(0.0),
			Want:   fptr(-1.2),
		},
		{
			Input:  `-0.5`,
			Target: fptr(0.0),
			Want:   fptr(-0.5),
		},
		{
			Input:  `()`,
			Target: slicePtr([]string(nil)),
			Want:   slicePtr([]string{}),
		},
		{
			Input:  `(hello "world")`,
			Target: slicePtr([]string(nil)),
			Want:   slicePtr([]string{"hello", "world"}),
		},
		{
			Input:  `(hello world I like pizza)`,
			Target: slicePtr([]string(nil)),
			Want:   slicePtr([]string{"hello", "world", "I", "like", "pizza"}),
		},
		{
			Input:  `((hello world) () (I like pizza))`,
			Target: sliceSlicePtr([][]string(nil)),
			Want:   sliceSlicePtr([][]string{{"hello", "world"}, {}, {"I", "like", "pizza"}}),
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%s into %T", test.Input, test.Target)
		t.Run(testName, func(t *testing.T) {
			reader := strings.NewReader(test.Input)
			err := Decode(reader, test.Target)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			got := test.Target
			want := test.Want
			if !reflect.DeepEqual(got, want) {
				t.Errorf(
					"incorrect result\ngot:  %swant: %s",
					spew.Sdump(got), spew.Sdump(want),
				)
			}
		})
	}
}

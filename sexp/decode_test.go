package sexp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestDecode_valid(t *testing.T) {
	type PCBGeneral struct {
		Links int `kicad:"links"`
		Nets  int `kicad:"nets"`
	}

	type PCBNet struct {
		Index int    `kicad:""`
		Name  string `kicad:""`
	}

	type PCBLayer struct {
		Index int      `kicad:""`
		Name  string   `kicad:""`
		Type  string   `kicad:""`
		Flags []string `kicad:",flat"`
	}

	type PCBNetClass struct {
		Name        string   `kicad:""`
		Description string   `kicad:""`
		Clearance   float64  `kicad:"clearance"`
		Nets        []string `kicad:"add_net,multi"`
	}

	type PCB struct {
		Version    int           `kicad:"version"`
		General    PCBGeneral    `kicad:"general,flat"`
		Page       string        `kicad:"page"`
		Nets       []PCBNet      `kicad:"net,multi,flat"`
		Layers     []PCBLayer    `kicad:"layers,flat"`
		NetClasses []PCBNetClass `kicad:"net_class,multi,flat"`
	}

	tests := []struct {
		Input  string
		FileTy string
		Target interface{}
		Want   interface{}
	}{
		{
			Input:  `(kicad_pcb)`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want:   &PCB{},
		},
		{
			Input:  `(kicad_pcb (page "USLetter"))`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want: &PCB{
				Page: "USLetter",
			},
		},
		{
			Input:  `(kicad_pcb (page "USLetter") (general))`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want: &PCB{
				Page: "USLetter",
			},
		},
		{
			Input:  `(kicad_pcb (page "USLetter") (general (links 10)))`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want: &PCB{
				Page: "USLetter",
				General: PCBGeneral{
					Links: 10,
				},
			},
		},
		{
			Input:  `(kicad_pcb (net 1 "Foo") (net 3 "Baz"))`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want: &PCB{
				Nets: []PCBNet{
					{
						Index: 1,
						Name:  "Foo",
					},
					{
						Index: 3,
						Name:  "Baz",
					},
				},
			},
		},
		{
			Input:  `(kicad_pcb (layers))`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want:   &PCB{},
		},
		{
			Input:  `(kicad_pcb (layers (1 F.Cu signal) (2 B.Cu power hide)))`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want: &PCB{
				Layers: []PCBLayer{
					{
						Index: 1,
						Name:  "F.Cu",
						Type:  "signal",
					},
					{
						Index: 2,
						Name:  "B.Cu",
						Type:  "power",
						Flags: []string{"hide"},
					},
				},
			},
		},
		{
			Input:  `(kicad_pcb (net_class Default "The default"))`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want: &PCB{
				NetClasses: []PCBNetClass{
					{
						Name:        "Default",
						Description: "The default",
					},
				},
			},
		},
		{
			Input:  `(kicad_pcb (net_class a b (add_net foo) (add_net bar)))`,
			FileTy: "kicad_pcb",
			Target: &PCB{},
			Want: &PCB{
				NetClasses: []PCBNetClass{
					{
						Name:        "a",
						Description: "b",
						Nets:        []string{"foo", "bar"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%s into %T", test.Input, test.Target)
		t.Run(testName, func(t *testing.T) {
			reader := strings.NewReader(test.Input)
			err := Decode(reader, test.FileTy, test.Target)
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

func TestDecodeSimple_valid(t *testing.T) {
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
	mapPtr := func(s map[string]string) *map[string]string {
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
			Input:  `0xdeadbeef`,
			Target: uiptr(0),
			Want:   uiptr(0xdeadbeef),
		},
		{
			Input:  `0xfaded_face`,
			Target: uiptr(0),
			Want:   uiptr(0xfadedface),
		},
		{
			Input:  `0xfeed_face_cafe`,
			Target: uiptr(0),
			Want:   uiptr(0xfeedfacecafe),
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
		{
			Input:  `((greeting "hello world")(pizza_topping cheese))`,
			Target: mapPtr(map[string]string{}),
			Want: mapPtr(map[string]string{
				"greeting":      "hello world",
				"pizza_topping": "cheese",
			}),
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%s into %T", test.Input, test.Target)
		t.Run(testName, func(t *testing.T) {
			reader := strings.NewReader(test.Input)
			err := DecodeSimple(reader, test.Target)
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

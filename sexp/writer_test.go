package sexp

import (
	"bytes"
	"testing"
)

func TestWriter_valid(t *testing.T) {
	bw := bytes.NewBufferString("")
	w := NewWriter(bw)

	w.BeginTuple()
	w.WriteRawString("export")
	w.BeginTuple()
	w.WriteRawString("version")
	w.WriteString("D")
	w.EndTuple()
	w.BeginTuple()
	w.WriteRawString("design")
	w.BeginTuple()
	w.WriteRawString("tool")
	w.WriteQuoteString(`go-kicad test "foo"`)
	w.EndTuple()
	w.EndTuple()
	w.BeginTuple()
	w.WriteRawString("components")
	w.BeginTuple()
	w.WriteRawString("comp")
	w.BeginTuple()
	w.WriteRawString("value")
	w.WriteString("1k")
	w.EndTuple()
	w.BeginTuple()
	w.WriteRawString("footprint")
	w.WriteString("Resistors SMD:R 1206 HandSoldering")
	w.EndTuple()
	w.EndTuple()
	w.EndTuple()
	w.EndTuple()
	err := w.Close()
	if err != nil {
		t.Errorf("error on close: %s", err)
	}

	got := bw.String()
	want := `(export
  (version D)
  (design
    (tool "go-kicad test \"foo\""))
  (components
    (comp
      (value 1k)
      (footprint "Resistors SMD:R 1206 HandSoldering"))))`
	if got != want {
		t.Errorf("incorrect result\ngot:  %s\nwant: %s", got, want)
	}
}

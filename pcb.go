package kicad

import (
	"io"
	"os"

	"github.com/apparentlymart/go-kicad/sexp"
)

// ReadPCB reads a stream containing a pcbnew PCB document and returns a
// PCB structure describing it.
//
// The PCB structure is not a comprehensive representation of the pcbnew
// file format, so overwriting the original file using WritePCB with the
// returned object is a lossy operation.
func ReadPCB(r io.Reader) (*PCB, error) {
	pcb := &PCB{}
	err := sexp.Decode(r, "kicad_pcb", pcb)
	return pcb, err
}

// ReadPCBFile is a convenience wrapper around ReadPCB that takes a filename
// and opens the given file for reading before calling ReadPCB.
func ReadPCBFile(filename string) (*PCB, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return ReadPCB(f)
}

// PCB represents a KiCad pcbnew PCB document.
type PCB struct {
	Version int        `kicad:"version"`
	Host    string     `kicad:"host"`
	General PCBGeneral `kicad:"general"`
}

type PCBGeneral struct {
	Links      int         `kicad:"links"`
	NoConnects int         `kicad:"no_connects"`
	Area       BoundingBox `kicad:"area"`
	Thickness  float64     `kicad:"thickness"`
	Drawings   int         `kicad:"drawings"`
	Tracks     int         `kicad:"tracks"`
	Zones      int         `kicad:"zones"`
	Modules    int         `kicad:"modules"`
	Nets       int         `kicad:"nets"`
}

type Position struct {
	X float64 `kicad:""`
	Y float64 `kicad:""`
}

type PositionAngle struct {
	X     float64 `kicad:""`
	Y     float64 `kicad:""`
	Angle float64 `kicad:""`
}

type BoundingBox struct {
	X1 float64 `kicad:""`
	Y1 float64 `kicad:""`
	X2 float64 `kicad:""`
	Y2 float64 `kicad:""`
}

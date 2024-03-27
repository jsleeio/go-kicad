// Package kicad defines structures reflecting the Kicad 8.0 file format
package kicad

import (
	"fmt"
	"io"
	"os"
	"strconv"

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
	Version          string         `kicad:"version"`
	Generator        string         `kicad:"generator"`
	GeneratorVersion string         `kicad:"generator_version"`
	General          PCBGeneral     `kicad:"general,flat"`
	Paper            string         `kicad:"paper"`
	Layers           []Layer        `kicad:"layers,flat"`
	Setup            Setup          `kicad:"setup,flat"`
	Nets             []Net          `kicad:"net,flat,multi"`
	Footprints       []Footprint    `kicad:"footprint,flat,multi"`
	Segments         []Segment      `kicad:"segment,flat,multi"`
	GraphicsLines    []GraphicsLine `kicad:"gr_line,flat,multi"`
}

// PCBGeneral ...
type PCBGeneral struct {
	Thickness       float64 `kicad:"thickness"`
	LegacyTeardrops bool    `kicad:"legacy_teardrops"`
}

// Layer ...
type Layer struct {
	Index   int      `kicad:""`
	Name    string   `kicad:""`
	Kind    string   `kicad:""`
	Aliases []string `kicad:",flat"`
}

// Setup ...
type Setup struct {
	PadToMaskClearance                 float64           `kicad:"pad_to_mask_clearance"`
	AllowSoldermaskBridgesInFootprints bool              `kicad:"allow_soldermask_bridges_in_footprints"`
	AuxAxisOrigin                      Position          `kicad:"aux_axis_origin,flat"`
	GridOrigin                         Position          `kicad:"grid_origin,flat"`
	PCBPlotParameters                  PCBPlotParameters `kicad:"pcbplotparams,flat"`
}

// PCBPlotParameters ...
type PCBPlotParameters struct {
	LayerSelection              uint    `kicad:"layerselection"`
	PlotOnAllLayersSelection    uint    `kicad:"plot_on_all_layers_selection"`
	DisableApertureMacros       bool    `kicad:"disableapertmacros"`
	UseGerberExtensions         bool    `kicad:"usegerberextensions"`
	UseGerberAttributes         bool    `kicad:"usegerberattributes"`
	UseGerberAdvancedAttributes bool    `kicad:"usegerberadvancedattributes"`
	CreateGerberJobFile         bool    `kicad:"creategerberjobfile"`
	DashedLineDashRatio         float64 `kicad:"dashed_line_dash_ratio"`
	DashedLineGapRatio          float64 `kicad:"dashed_line_gap_ratio"`
	SVGPrecision                int     `kicad:"svgprecision"`
	PlotFrameRef                bool    `kicad:"plotframeref"`
	ViasOnMask                  bool    `kicad:"viasonmask"`
	Mode                        int     `kicad:"mode"`
	UseAuxOrigin                bool    `kicad:"useauxorigin"`
	HPGLPenNumber               int     `kicad:"hpglpennumber"`
	HPGLPenSpeed                int     `kicad:"hpglpenspeed"`
	HPGLPenDiameter             float64 `kicad:"hpglpendiameter"`
	PDFFrontFPPropertyPopups    bool    `kicad:"pdf_front_fp_property_popups"`
	PDFBackFPPropertyPopups     bool    `kicad:"pdf_back_fp_property_popups"`
	DXFPolygonMode              bool    `kicad:"dxfpolygonmode"`
	DXFImperialUnits            bool    `kicad:"dxfimperialunits"`
	DXFUsePCBNewFont            bool    `kicad:"dxfusepcbnewfont"`
	PostscriptNegative          bool    `kicad:"psnegative"`
	PostscriptA4Output          bool    `kicad:"psa4output"`
	PlotReference               bool    `kicad:"plotreference"`
	PlotValue                   bool    `kicad:"plotvalue"`
	PlotFPText                  bool    `kicad:"plotfptext"`
	PlotInvisibleText           bool    `kicad:"plotinvisibletext"`
	SketchPadsOnFab             bool    `kicad:"sketchpadsonfab"`
	SubtractMaskFromSilk        bool    `kicad:"subtractmaskfromsilk"`
	OutputFormat                int     `kicad:"outputformat"`
	Mirror                      bool    `kicad:"mirror"`
	DrillShape                  int     `kicad:"drillshape"`
	ScaleSelection              int     `kicad:"scaleselection"`
	OutputDirectory             string  `kicad:"outputdirectory"`
}

// Footprint ...
type Footprint struct {
	Name        string            `kicad:""`
	Layer       string            `kicad:"layer"`
	UUID        string            `kicad:"uuid"`
	At          PositionAngle     `kicad:"at,flat"`
	Description string            `kicad:"descr"`
	Tags        string            `kicad:"tags"`
	Properties  []Property        `kicad:"property,flat,multi"`
	Path        string            `kicad:"path"`
	SheetFile   string            `kicad:"sheetfile"`
	Attribute   []string          `kicad:"attr,flat"`
	Circles     []FootprintCircle `kicad:"fp_circle,flat,multi"`
	Texts       []FootprintText   `kicad:"fp_text,flat,multi"`
	Pads        []FootprintPad    `kicad:"pad,flat,multi"`
	Groups      []Group           `kicad:"group,flat,multi"`
	properties  map[string]string
}

// PropertiesMap collects the various properties associated with a footprint into
// a map and returns it
func (f *Footprint) PropertiesMap() map[string]string {
	if f.properties != nil {
		return f.properties
	}
	f.properties = make(map[string]string)
	for _, p := range f.Properties {
		f.properties[p.Name] = p.Text
	}
	return f.properties
}

// PropertyOrDefaultString returns the string value of a property if it is
// present, otherwise returns a provided default value
func (f *Footprint) PropertyOrDefaultString(key, defaultValue string) string {
	if v, found := f.PropertiesMap()[key]; found && v != "" {
		return v
	}
	return defaultValue
}

// PropertyOrDefaultInt attempts to convert a property value, if present, to
// int, returning an error if the conversion failed. If the property is not
// present, the provided default value is returned.
func (f *Footprint) PropertyOrDefaultInt(key string, defaultValue int) (int, error) {
	v, found := f.PropertiesMap()[key]
	if !found {
		return defaultValue, nil
	}
	iv, err := strconv.Atoi(v)
	if err != nil {
		return 0.0, fmt.Errorf("PropertyOrDefaultFloat: invalid float value for key %s: %q", key, v)
	}
	return iv, nil
}

// PropertyOrDefaultFloat attempts to convert a property value, if present, to
// float64, returning an error if the conversion failed. If the property is not
// present, the provided default value is returned.
func (f *Footprint) PropertyOrDefaultFloat(key string, defaultValue float64) (float64, error) {
	v, found := f.PropertiesMap()[key]
	if !found {
		return defaultValue, nil
	}
	fv, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0.0, fmt.Errorf("PropertyOrDefaultFloat: invalid float value for key %s: %q", key, v)
	}
	return fv, nil
}

// MountedOnBack returns true if the footprint is on the back side of the PCB
func (f Footprint) MountedOnBack() bool {
	return f.Layer == "B.Cu"
}

// MountedOnFront returns true if the footprint is on the front side of the PCB
func (f Footprint) MountedOnFront() bool {
	return f.Layer == "F.Cu"
}

// Property ...
type Property struct {
	Name     string        `kicad:""`
	Text     string        `kicad:""`
	At       PositionAngle `kicad:"at,flat"`
	Layer    string        `kicad:"layer"`
	Unlocked bool          `kicad:"unlocked"`
	Hide     bool          `kicad:"hide"`
	UUID     string        `kicad:"uuid"`
	Effects  Effects       `kicad:"effects,flat"`
}

// Stroke ...
type Stroke struct {
	Width float64 `kicad:"width"`
	Type  string  `kicad:"type"`
}

// FootprintCircle ...
type FootprintCircle struct {
	Center Position `kicad:"center,flat"`
	End    Position `kicad:"end,flat"`
	Stroke Stroke   `kicad:"stroke,flat"`
	Fill   string   `kicad:"fill"`
	Layer  string   `kicad:"layer"`
	UUID   string   `kicad:"uuid"`
}

// FootprintText ...
type FootprintText struct {
	Kind     string        `kicad:""`
	Text     string        `kicad:""`
	At       PositionAngle `kicad:"at,flat"`
	Unlocked bool          `kicad:"unlocked"`
	Layer    string        `kicad:"layer"`
	UUID     string        `kicad:"uuid"`
	Effects  Effects       `kicad:"effects,flat"`
}

// FootprintPad ...
type FootprintPad struct {
	Name               string        `kicad:""`
	Kind               string        `kicad:""`
	Shape              string        `kicad:""`
	At                 PositionAngle `kicad:"at,flat"`
	Size               Size          `kicad:"size,flat"`
	Drill              float64       `kicad:"drill"`
	Layers             []string      `kicad:"layers,flat"`
	RemoveUnusedLayers bool          `kicad:"remove_unused_layers"`
	Net                Net           `kicad:"net,flat"`
	PinFunction        string        `kicad:"pinfunction"`
	PinType            string        `kicad:"pintype"`
	UUID               string        `kicad:"uuid"`
}

// Group ...
type Group struct {
	Name    string   `kicad:""`
	UUID    string   `kicad:"uuid"`
	Members []string `kicad:"members,flat"`
}

// GraphicsLine ...
type GraphicsLine struct {
	Start  Position `kicad:"start,flat"`
	End    Position `kicad:"end,flat"`
	UUID   string   `kicad:"uuid"`
	Stroke Stroke   `kicad:"stroke,flat"`
	Layer  string   `kicad:"layer"`
}

// Segment ...
type Segment struct {
	Start Position `kicad:"start,flat"`
	End   Position `kicad:"end,flat"`
	UUID  string   `kicad:"uuid"`
	Width float64  `kicad:"width"`
	Net   string   `kicad:"net"`
	Layer string   `kicad:"layer"`
}

// Effects ...
type Effects struct {
	Font Font `kicad:"font,flat"`
}

// Font ...
type Font struct {
	Size      Size    `kicad:"size,flat"`
	Thickness float64 `kicad:"thickness"`
}

// Size ...
type Size struct {
	Width  float64 `kicad:""`
	Height float64 `kicad:""`
}

// Net describes a single net in the netlist
type Net struct {
	Index int      `kicad:""`
	Name  string   `kicad:""`
	Flags []string `kicad:",flat"`
}

// Position ...
type Position struct {
	X float64 `kicad:""`
	Y float64 `kicad:""`
}

// PositionAngle ...
type PositionAngle struct {
	X         float64   `kicad:""`
	Y         float64   `kicad:""`
	Remainder []float64 `kicad:",flat"` // because sometimes angle is absent :-(
}

// Angle ...
func (pa PositionAngle) Angle() float64 {
	if len(pa.Remainder) < 1 {
		return 0.0
	}
	return pa.Remainder[0]
}

// BoundingBox ...
type BoundingBox struct {
	X1 float64 `kicad:""`
	Y1 float64 `kicad:""`
	X2 float64 `kicad:""`
	Y2 float64 `kicad:""`
}

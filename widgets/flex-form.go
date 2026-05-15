package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// FlexForm is the WidgetForm for *Flex.
//
// Flex is a container, but its Add method takes no per-child layout
// arguments — children are positioned by alignment and spacing on the
// container itself — so FlexForm satisfies core.WidgetForm only and
// does not provide a LayoutForm.
type FlexForm struct {
	ComponentForm

	Vertical  bool   `group:"layout" label:"Vertical"`
	Alignment string `group:"layout" label:"Alignment" control:"select" options:"start,center,end,stretch,left,right"`
	Spacing   int    `group:"layout" label:"Spacing"`
}

func (f *FlexForm) Name() string  { return "Flex" }
func (f *FlexForm) Group() string { return "container" }
func (f *FlexForm) Help() string  { return "Linear container with alignment and spacing" }

func (f *FlexForm) Load(w core.Widget) {
	x := w.(*Flex)
	f.ComponentForm.Load(&x.Component)
	f.Vertical = x.Flag(core.FlagVertical)
	f.Alignment = x.alignment.String()
	f.Spacing = x.spacing
}

func (f *FlexForm) Store(w core.Widget) {
	x := w.(*Flex)
	f.ComponentForm.Store(&x.Component)
	x.SetFlag(core.FlagVertical, f.Vertical)
	x.alignment = parseAlignment(f.Alignment)
	x.spacing = f.Spacing
}

func (f *FlexForm) New() core.Widget {
	x := NewFlex("", "", parseAlignment(f.Alignment), f.Spacing)
	if f.Vertical {
		x.SetFlag(core.FlagVertical, true)
	}
	f.Store(x)
	return x
}

func (f *FlexForm) Validate(field string) error { return nil }

// Emit writes the Flex constructor onto an existing chain. Bypasses
// EmitFrame because the constructor name itself is computed (HFlex
// vs VFlex) from Vertical, and because we want to suppress the
// FlagVertical entry in the standard chain (the VFlex path sets it
// implicitly). The class prefix and the rest of the standard chain
// (Hint / Skip / Hidden / Disabled / style) are reused as
// primitives.
func (f *FlexForm) Emit(w io.Writer, mode string) error {
	if err := f.CheckBuilderMode(mode); err != nil {
		return err
	}
	f.EmitClassPrefix(w)
	ctor := "HFlex"
	if f.Vertical {
		ctor = "VFlex"
	}
	fmt.Fprintf(w, "%s(%q, %s, %d).\n",
		ctor, f.ID, alignmentConst(f.Alignment), f.Spacing)
	// FlagVertical is implicitly set by VFlex, so a separate Flag
	// chain entry would be redundant. Everything else from the
	// standard ComponentForm chain is welcome.
	if f.HintW != 0 || f.HintH != 0 {
		fmt.Fprintf(w, "Hint(%d, %d).\n", f.HintW, f.HintH)
	}
	if f.Skip {
		fmt.Fprintf(w, "Flag(FlagSkip, true).\n")
	}
	if f.Hidden {
		fmt.Fprintf(w, "Flag(FlagHidden, true).\n")
	}
	if f.Disabled {
		fmt.Fprintf(w, "Flag(FlagDisabled, true).\n")
	}
	f.EmitStyle(w)
	return nil
}

// alignmentConst maps an Alignment.String() value back to the
// exported core constant identifier so generated code reads as
// "core.Center" or "Center" inside the widgets package.
func alignmentConst(s string) string {
	switch s {
	case "start":
		return "Start"
	case "left":
		return "Left"
	case "center":
		return "Center"
	case "right":
		return "Right"
	case "end":
		return "End"
	case "stretch":
		return "Stretch"
	default:
		return "Default"
	}
}

// parseAlignment is the inverse of Alignment.String. Empty / unknown
// values fall back to Default rather than failing the form Store, so
// editing freshly-created forms never panics.
func parseAlignment(s string) core.Alignment {
	switch s {
	case "start":
		return core.Start
	case "left":
		return core.Left
	case "center":
		return core.Center
	case "right":
		return core.Right
	case "end":
		return core.End
	case "stretch":
		return core.Stretch
	default:
		return core.Default
	}
}

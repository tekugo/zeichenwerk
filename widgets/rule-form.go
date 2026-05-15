package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// RuleForm is the WidgetForm for *Rule. The Builder exposes Rule via
// the orientation-specific HRule and VRule constructors; the form
// switches between them via the Horizontal flag.
type RuleForm struct {
	ComponentForm

	Horizontal  bool   `group:"general" label:"Horizontal"`
	BorderStyle string `group:"general" label:"Style" control:"border"`
}

func (f *RuleForm) Name() string  { return "Rule" }
func (f *RuleForm) Group() string { return "leaf" }
func (f *RuleForm) Help() string  { return "Horizontal or vertical visual divider" }

func (f *RuleForm) Load(w core.Widget) {
	r := w.(*Rule)
	f.ComponentForm.Load(&r.Component)
	f.Horizontal = r.horizontal
	f.BorderStyle = r.style
}

func (f *RuleForm) Store(w core.Widget) {
	r := w.(*Rule)
	f.ComponentForm.Store(&r.Component)
	r.horizontal = f.Horizontal
	r.style = f.BorderStyle
	if f.Horizontal {
		r.id = "hrule"
		r.hheight = 1
		r.hwidth = 0
	} else {
		r.id = "vrule"
		r.hwidth = 1
		r.hheight = 0
	}
}

func (f *RuleForm) New() core.Widget {
	var r *Rule
	if f.Horizontal {
		r = NewHRule("", f.BorderStyle)
	} else {
		r = NewVRule("", f.BorderStyle)
	}
	f.Store(r)
	return r
}

func (f *RuleForm) Validate(field string) error { return nil }

// Emit picks HRule or VRule based on the orientation flag. Rule's
// Builder constructors take no id parameter (they hard-code "hrule" /
// "vrule"), so the form's ID field is dropped from the call.
func (f *RuleForm) Emit(w io.Writer, mode string) error {
	if err := f.CheckBuilderMode(mode); err != nil {
		return err
	}
	f.EmitClassPrefix(w)
	ctor := "HRule"
	if !f.Horizontal {
		ctor = "VRule"
	}
	fmt.Fprintf(w, "%s(%q).\n", ctor, f.BorderStyle)
	f.EmitChain(w)
	f.EmitStyle(w)
	return nil
}

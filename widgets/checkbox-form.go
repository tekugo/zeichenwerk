package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// CheckboxForm is the WidgetForm for *Checkbox.
//
// Checkbox state is split between ComponentForm (Hidden / Disabled /
// Skip flags) and the per-widget surface here: the label text, the
// initial checked state, and the read-only flag. The checked state is
// part of FlagChecked but is treated as a value rather than a flag in
// the form because it's the primary thing a user reasons about.
type CheckboxForm struct {
	ComponentForm

	Text     string `group:"general" label:"Text"`
	Checked  bool   `group:"value" label:"Checked"`
	Readonly bool   `group:"flags" label:"Read-only"`
}

func (f *CheckboxForm) Name() string  { return "Checkbox" }
func (f *CheckboxForm) Group() string { return "input" }
func (f *CheckboxForm) Help() string  { return "Toggleable boolean input with label" }

func (f *CheckboxForm) Load(w core.Widget) {
	c := w.(*Checkbox)
	f.ComponentForm.Load(&c.Component)
	f.Text = c.text
	f.Checked = c.Flag(core.FlagChecked)
	f.Readonly = c.Flag(core.FlagReadonly)
}

func (f *CheckboxForm) Store(w core.Widget) {
	c := w.(*Checkbox)
	f.ComponentForm.Store(&c.Component)
	c.text = f.Text
	c.SetFlag(core.FlagChecked, f.Checked)
	c.SetFlag(core.FlagReadonly, f.Readonly)
}

func (f *CheckboxForm) New() core.Widget {
	c := NewCheckbox("", "", "", false)
	f.Store(c)
	return c
}

func (f *CheckboxForm) Validate(field string) error { return nil }

// Emit writes the Checkbox constructor onto an existing chain. The
// Checked argument is emitted directly into the constructor; the
// Read-only flag follows as a Flag entry because there is no
// chained Builder setter for it.
func (f *CheckboxForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		if _, err := fmt.Fprintf(w, "Checkbox(%q, %q, %t).\n", f.ID, f.Text, f.Checked); err != nil {
			return err
		}
		if f.Readonly {
			if _, err := fmt.Fprintf(w, "Flag(FlagReadonly, true).\n"); err != nil {
				return err
			}
		}
		return nil
	})
}

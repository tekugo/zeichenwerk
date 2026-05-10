package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// SwitcherForm is the WidgetForm for *Switcher. Switcher's Add does
// not consume per-child parameters, so the form satisfies WidgetForm
// only. Selected drives which pane is visible after the form is
// loaded; out-of-range indices are silently clamped.
type SwitcherForm struct {
	ComponentForm

	Selected int `group:"value" label:"Selected"`
}

func (f *SwitcherForm) Name() string  { return "Switcher" }
func (f *SwitcherForm) Group() string { return "container" }
func (f *SwitcherForm) Help() string  { return "Container that shows one named pane at a time" }

func (f *SwitcherForm) Load(w core.Widget) {
	s := w.(*Switcher)
	f.ComponentForm.Load(&s.Component)
	f.Selected = s.selected
}

func (f *SwitcherForm) Store(w core.Widget) {
	s := w.(*Switcher)
	f.ComponentForm.Store(&s.Component)
	if f.Selected < 0 {
		f.Selected = 0
	}
	if f.Selected >= len(s.panes) {
		if len(s.panes) > 0 {
			f.Selected = len(s.panes) - 1
		} else {
			f.Selected = 0
		}
	}
	s.selected = f.Selected
}

func (f *SwitcherForm) New() core.Widget {
	s := NewSwitcher("", "")
	f.Store(s)
	return s
}

func (f *SwitcherForm) Validate(field string) error { return nil }

// Emit writes the Switcher constructor. The "connect to last Tabs"
// behaviour is a Builder-time concern that the form cannot capture
// without ambient state; the codegen path always emits connect=false
// and leaves wiring to the user.
func (f *SwitcherForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Switcher(%q, false).\n", f.ID)
		return err
	}); err != nil {
		return err
	}
	if f.Selected > 0 {
		fmt.Fprintf(w, "// TODO: Select(%d) — no Builder setter\n", f.Selected)
	}
	return nil
}

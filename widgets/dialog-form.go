package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// DialogForm is the WidgetForm for *Dialog. Dialog is a single-child
// container with a title bar; Add does not consume per-child
// parameters, so the form satisfies WidgetForm only.
type DialogForm struct {
	ComponentForm

	Title string `group:"general" label:"Title"`
}

func (f *DialogForm) Name() string  { return "Dialog" }
func (f *DialogForm) Group() string { return "container" }
func (f *DialogForm) Help() string  { return "Single-child overlay container with title bar" }

func (f *DialogForm) Load(w core.Widget) {
	d := w.(*Dialog)
	f.ComponentForm.Load(&d.Component)
	f.Title = d.title
}

func (f *DialogForm) Store(w core.Widget) {
	d := w.(*Dialog)
	f.ComponentForm.Store(&d.Component)
	d.title = f.Title
}

func (f *DialogForm) New() core.Widget {
	d := NewDialog("", "", f.Title)
	f.Store(d)
	return d
}

func (f *DialogForm) Validate(field string) error { return nil }

func (f *DialogForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Dialog(%q, %q).\n", f.ID, f.Title)
		return err
	})
}

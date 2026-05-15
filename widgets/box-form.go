package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// BoxForm is the WidgetForm for *Box.
//
// Box is a single-child container with an optional inline title.
// There is no per-child layout form because Box's Add ignores
// parameters (it always replaces the single slot), so BoxForm only
// implements WidgetForm — not ContainerForm.
type BoxForm struct {
	ComponentForm

	Title string `group:"general" label:"Title"`
}

func (f *BoxForm) Name() string  { return "Box" }
func (f *BoxForm) Group() string { return "container" }
func (f *BoxForm) Help() string  { return "Single-child container with optional title" }

func (f *BoxForm) Load(w core.Widget) {
	b := w.(*Box)
	f.ComponentForm.Load(&b.Component)
	f.Title = b.Title
}

func (f *BoxForm) Store(w core.Widget) {
	b := w.(*Box)
	f.ComponentForm.Store(&b.Component)
	b.Title = f.Title
}

func (f *BoxForm) New() core.Widget {
	b := NewBox("", "", "")
	f.Store(b)
	return b
}

func (f *BoxForm) Validate(field string) error { return nil }

func (f *BoxForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Box(%q, %q).\n", f.ID, f.Title)
		return err
	})
}

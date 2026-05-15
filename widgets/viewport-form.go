package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// ViewportForm is the WidgetForm for *Viewport. Viewport is a
// single-child container; its Add does not consume per-child
// parameters, so the form satisfies WidgetForm only.
type ViewportForm struct {
	ComponentForm

	Title string `group:"general" label:"Title"`
}

func (f *ViewportForm) Name() string  { return "Viewport" }
func (f *ViewportForm) Group() string { return "container" }
func (f *ViewportForm) Help() string  { return "Scrollable single-child container" }

func (f *ViewportForm) Load(w core.Widget) {
	v := w.(*Viewport)
	f.ComponentForm.Load(&v.Component)
	f.Title = v.Title
}

func (f *ViewportForm) Store(w core.Widget) {
	v := w.(*Viewport)
	f.ComponentForm.Store(&v.Component)
	v.Title = f.Title
}

func (f *ViewportForm) New() core.Widget {
	v := NewViewport("", "", f.Title)
	f.Store(v)
	return v
}

func (f *ViewportForm) Validate(field string) error { return nil }

func (f *ViewportForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Viewport(%q, %q).\n", f.ID, f.Title)
		return err
	})
}

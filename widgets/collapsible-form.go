package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// CollapsibleForm is the WidgetForm for *Collapsible. The widget is
// a single-child container whose Add does not consume per-child
// parameters, so the form satisfies WidgetForm only.
type CollapsibleForm struct {
	ComponentForm

	Title    string `group:"general" label:"Title"`
	Expanded bool   `group:"general" label:"Expanded"`
}

func (f *CollapsibleForm) Name() string  { return "Collapsible" }
func (f *CollapsibleForm) Group() string { return "container" }
func (f *CollapsibleForm) Help() string  { return "Single-child container with a clickable header" }

func (f *CollapsibleForm) Load(w core.Widget) {
	c := w.(*Collapsible)
	f.ComponentForm.Load(&c.Component)
	f.Title = c.title
	f.Expanded = c.expanded
}

func (f *CollapsibleForm) Store(w core.Widget) {
	c := w.(*Collapsible)
	f.ComponentForm.Store(&c.Component)
	c.title = f.Title
	c.expanded = f.Expanded
}

func (f *CollapsibleForm) New() core.Widget {
	c := NewCollapsible("", "", f.Title, f.Expanded)
	f.Store(c)
	return c
}

func (f *CollapsibleForm) Validate(field string) error { return nil }

func (f *CollapsibleForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Collapsible(%q, %q, %t).\n", f.ID, f.Title, f.Expanded)
		return err
	})
}

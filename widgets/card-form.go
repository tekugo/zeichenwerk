package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// CardForm is the WidgetForm for *Card.
//
// Card holds two slots — content and footer — but Add does not
// take parameters; the first Add becomes content, the second
// becomes footer. Because order is determined by the call sequence
// and not by an explicit cell coordinate, CardForm satisfies
// WidgetForm only and does not provide a LayoutForm.
type CardForm struct {
	ComponentForm

	Title string `group:"general" label:"Title"`
}

func (f *CardForm) Name() string  { return "Card" }
func (f *CardForm) Group() string { return "container" }
func (f *CardForm) Help() string  { return "Two-slot container with title and optional footer" }

func (f *CardForm) Load(w core.Widget) {
	c := w.(*Card)
	f.ComponentForm.Load(&c.Component)
	f.Title = c.Title
}

func (f *CardForm) Store(w core.Widget) {
	c := w.(*Card)
	f.ComponentForm.Store(&c.Component)
	c.Title = f.Title
}

func (f *CardForm) New() core.Widget {
	c := NewCard("", "", "")
	f.Store(c)
	return c
}

func (f *CardForm) Validate(field string) error { return nil }

func (f *CardForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Card(%q, %q).\n", f.ID, f.Title)
		return err
	})
}

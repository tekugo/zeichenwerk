package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// DigitsForm is the WidgetForm for *Digits. Digits is a leaf widget
// that renders a string as ASCII-art glyphs.
type DigitsForm struct {
	ComponentForm

	Text string `group:"general" label:"Text"`
}

func (f *DigitsForm) Name() string  { return "Digits" }
func (f *DigitsForm) Group() string { return "leaf" }
func (f *DigitsForm) Help() string  { return "Large ASCII-art glyph display" }

func (f *DigitsForm) Load(w core.Widget) {
	d := w.(*Digits)
	f.ComponentForm.Load(&d.Component)
	f.Text = d.Text
}

func (f *DigitsForm) Store(w core.Widget) {
	d := w.(*Digits)
	f.ComponentForm.Store(&d.Component)
	d.Text = f.Text
}

func (f *DigitsForm) New() core.Widget {
	d := NewDigits("", "", f.Text)
	f.Store(d)
	return d
}

func (f *DigitsForm) Validate(field string) error { return nil }

func (f *DigitsForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Digits(%q, %q).\n", f.ID, f.Text)
		return err
	})
}

package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// ButtonForm is the WidgetForm for *Button.
//
// Button has a single editable property beyond ComponentForm — its
// label text. Activation handlers cannot be round-tripped through
// codegen, so Emit produces a comment-stub for any registered
// EvtActivate handler in the same place as Static does for its
// Alignment field.
type ButtonForm struct {
	ComponentForm

	Text string `group:"general" label:"Text"`
}

func (f *ButtonForm) Name() string  { return "Button" }
func (f *ButtonForm) Group() string { return "leaf" }
func (f *ButtonForm) Help() string  { return "Clickable button with a label" }

func (f *ButtonForm) Load(w core.Widget) {
	b := w.(*Button)
	f.ComponentForm.Load(&b.Component)
	f.Text = b.text
}

func (f *ButtonForm) Store(w core.Widget) {
	b := w.(*Button)
	f.ComponentForm.Store(&b.Component)
	b.text = f.Text
}

func (f *ButtonForm) New() core.Widget {
	b := NewButton("", "", "")
	f.Store(b)
	return b
}

func (f *ButtonForm) Validate(field string) error { return nil }

func (f *ButtonForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Button(%q, %q).\n", f.ID, f.Text)
		return err
	})
}

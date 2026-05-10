package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// StyledForm is the WidgetForm for *Styled. Styled is read-only at
// runtime; the static editing surface is the Markdown source. Scroll
// state is runtime-only.
type StyledForm struct {
	ComponentForm

	Text string `group:"general" label:"Text"`
}

func (f *StyledForm) Name() string  { return "Styled" }
func (f *StyledForm) Group() string { return "leaf" }
func (f *StyledForm) Help() string  { return "Markdown-styled text display" }

func (f *StyledForm) Load(w core.Widget) {
	s := w.(*Styled)
	f.ComponentForm.Load(&s.Component)
	f.Text = s.text
}

func (f *StyledForm) Store(w core.Widget) {
	s := w.(*Styled)
	f.ComponentForm.Store(&s.Component)
	s.SetText(f.Text)
}

func (f *StyledForm) New() core.Widget {
	s := NewStyled("", "", f.Text)
	f.Store(s)
	return s
}

func (f *StyledForm) Validate(field string) error { return nil }

func (f *StyledForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Styled(%q, %q).\n", f.ID, f.Text)
		return err
	})
}

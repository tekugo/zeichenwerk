package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// StaticForm is the WidgetForm for *Static. It exists in the widgets
// package so Load and Store can read and write Static's fields directly.
type StaticForm struct {
	ComponentForm

	Text      string `group:"general" label:"Text"`
	Alignment string `group:"general" label:"Alignment" control:"select" options:"left,center,right"`
}

func (f *StaticForm) Name() string  { return "Static" }
func (f *StaticForm) Group() string { return "leaf" }
func (f *StaticForm) Help() string  { return "Non-interactive text label" }

func (f *StaticForm) Load(w core.Widget) {
	s := w.(*Static)
	f.ComponentForm.Load(&s.Component)
	f.Text = s.Text
	f.Alignment = s.Alignment
}

func (f *StaticForm) Store(w core.Widget) {
	s := w.(*Static)
	f.ComponentForm.Store(&s.Component)
	s.Text = f.Text
	if f.Alignment != "" {
		s.Alignment = f.Alignment
	}
}

func (f *StaticForm) New() core.Widget {
	s := NewStatic("", "", "")
	f.Store(s)
	return s
}

func (f *StaticForm) Validate(field string) error { return nil }

// Emit writes Static's call shape onto an in-progress chain by
// delegating the standard prefix / chain / style machinery to
// EmitFrame and supplying just the constructor in the body. Each
// chain element ends with ".\n" — see ComponentForm.EmitClassPrefix
// for the convention. Alignment is emitted as a TODO comment after
// the frame because the Builder API has no chained setter for it.
func (f *StaticForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Static(%q, %q).\n", f.ID, f.Text)
		return err
	}); err != nil {
		return err
	}
	if f.Alignment != "" && f.Alignment != "left" {
		fmt.Fprintf(w, "// TODO: SetAlignment(%q) — no Builder setter\n", f.Alignment)
	}
	return nil
}

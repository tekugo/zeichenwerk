package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// EditorForm is the WidgetForm for *Editor. Buffer contents and
// cursor / selection state are runtime-only; the static editing
// surface here covers tab handling, line numbers, auto-indent, and
// the read-only flag.
type EditorForm struct {
	ComponentForm

	TabWidth    int  `group:"general" label:"Tab Width"`
	UseSpaces   bool `group:"general" label:"Insert Spaces"`
	LineNumbers int  `group:"general" label:"Line Number Width"`
	AutoIndent  bool `group:"general" label:"Auto-indent"`
	Readonly    bool `group:"flags" label:"Read-only"`
}

func (f *EditorForm) Name() string  { return "Editor" }
func (f *EditorForm) Group() string { return "input" }
func (f *EditorForm) Help() string  { return "Multi-line text editor" }

func (f *EditorForm) Load(w core.Widget) {
	e := w.(*Editor)
	f.ComponentForm.Load(&e.Component)
	f.TabWidth = e.tab
	f.UseSpaces = e.spaces
	f.LineNumbers = e.numbers
	f.AutoIndent = e.indent
	f.Readonly = e.disabled
}

func (f *EditorForm) Store(w core.Widget) {
	e := w.(*Editor)
	f.ComponentForm.Store(&e.Component)
	if f.TabWidth > 0 {
		e.tab = f.TabWidth
	}
	e.spaces = f.UseSpaces
	e.numbers = f.LineNumbers
	e.indent = f.AutoIndent
	e.disabled = f.Readonly
}

func (f *EditorForm) New() core.Widget {
	e := NewEditor("", "")
	f.Store(e)
	return e
}

func (f *EditorForm) Validate(field string) error { return nil }

// Emit writes the Editor constructor. The configuration setters
// (SetTabWidth, UseSpaces, ShowLineNumbers, SetAutoIndent,
// SetReadOnly) are not chained on the Builder, so changes from the
// defaults emit as TODO comments after the frame.
func (f *EditorForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Editor(%q).\n", f.ID)
		return err
	}); err != nil {
		return err
	}
	if f.TabWidth != 0 && f.TabWidth != 4 {
		fmt.Fprintf(w, "// TODO: SetTabWidth(%d) — no Builder setter\n", f.TabWidth)
	}
	if f.UseSpaces {
		fmt.Fprintf(w, "// TODO: UseSpaces(true) — no Builder setter\n")
	}
	if f.LineNumbers > 0 {
		fmt.Fprintf(w, "// TODO: ShowLineNumbers(%d) — no Builder setter\n", f.LineNumbers)
	}
	if !f.AutoIndent {
		fmt.Fprintf(w, "// TODO: SetAutoIndent(false) — no Builder setter\n")
	}
	if f.Readonly {
		fmt.Fprintf(w, "// TODO: SetReadOnly(true) — no Builder setter\n")
	}
	return nil
}

package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// TerminalForm is the WidgetForm for *Terminal. The cell buffers and
// ANSI parser state are runtime-only; the static editing surface
// covers the title, auto-wrap mode, and cursor visibility.
type TerminalForm struct {
	ComponentForm

	Title      string `group:"general" label:"Title"`
	AutoWrap   bool   `group:"general" label:"Auto Wrap"`
	ShowCursor bool   `group:"general" label:"Show Cursor"`
}

func (f *TerminalForm) Name() string  { return "Terminal" }
func (f *TerminalForm) Group() string { return "leaf" }
func (f *TerminalForm) Help() string  { return "ANSI terminal emulator" }

func (f *TerminalForm) Load(w core.Widget) {
	t := w.(*Terminal)
	f.ComponentForm.Load(&t.Component)
	f.Title = t.title
	f.AutoWrap = t.autoWrap
	f.ShowCursor = t.showCursor
}

func (f *TerminalForm) Store(w core.Widget) {
	t := w.(*Terminal)
	f.ComponentForm.Store(&t.Component)
	t.title = f.Title
	t.autoWrap = f.AutoWrap
	t.showCursor = f.ShowCursor
}

func (f *TerminalForm) New() core.Widget {
	t := NewTerminal("", "")
	f.Store(t)
	return t
}

func (f *TerminalForm) Validate(field string) error { return nil }

func (f *TerminalForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Terminal(%q).\n", f.ID)
		return err
	}); err != nil {
		return err
	}
	if !f.AutoWrap {
		fmt.Fprintf(w, "// TODO: AutoWrap = false — no Builder setter\n")
	}
	if !f.ShowCursor {
		fmt.Fprintf(w, "// TODO: ShowCursor = false — no Builder setter\n")
	}
	return nil
}

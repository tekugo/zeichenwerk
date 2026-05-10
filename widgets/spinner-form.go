package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// SpinnerForm is the WidgetForm for *Spinner. Sequence is the raw
// space-separated frame string Spinner stores internally; named
// sequences from the Spinners map are exposed via the SequenceName
// dropdown for convenience.
type SpinnerForm struct {
	ComponentForm

	Sequence string `group:"general" label:"Sequence"`
}

func (f *SpinnerForm) Name() string  { return "Spinner" }
func (f *SpinnerForm) Group() string { return "leaf" }
func (f *SpinnerForm) Help() string  { return "Animated spinner cycling through Unicode frames" }

func (f *SpinnerForm) Load(w core.Widget) {
	s := w.(*Spinner)
	f.ComponentForm.Load(&s.Component)
	f.Sequence = strings.Join(s.sequence, " ")
}

func (f *SpinnerForm) Store(w core.Widget) {
	s := w.(*Spinner)
	f.ComponentForm.Store(&s.Component)
	s.SetSequence(f.Sequence)
}

func (f *SpinnerForm) New() core.Widget {
	s := NewSpinner("", "", f.Sequence)
	f.Store(s)
	return s
}

func (f *SpinnerForm) Validate(field string) error { return nil }

func (f *SpinnerForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Spinner(%q, %q).\n", f.ID, f.Sequence)
		return err
	})
}

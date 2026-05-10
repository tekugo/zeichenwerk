package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// ProgressForm is the WidgetForm for *Progress. Total = 0 puts the
// widget in indeterminate mode; the form preserves that distinction.
type ProgressForm struct {
	ComponentForm

	Horizontal bool `group:"general" label:"Horizontal"`
	Value      int  `group:"value" label:"Value"`
	Total      int  `group:"value" label:"Total"`
}

func (f *ProgressForm) Name() string  { return "Progress" }
func (f *ProgressForm) Group() string { return "leaf" }
func (f *ProgressForm) Help() string  { return "Determinate or indeterminate progress indicator" }

func (f *ProgressForm) Load(w core.Widget) {
	p := w.(*Progress)
	f.ComponentForm.Load(&p.Component)
	f.Horizontal = p.horizontal
	f.Value = p.value
	f.Total = p.total
}

func (f *ProgressForm) Store(w core.Widget) {
	p := w.(*Progress)
	f.ComponentForm.Store(&p.Component)
	p.horizontal = f.Horizontal
	p.SetTotal(f.Total)
	p.Set(f.Value)
}

func (f *ProgressForm) New() core.Widget {
	p := NewProgress("", "", f.Horizontal)
	f.Store(p)
	return p
}

func (f *ProgressForm) Validate(field string) error { return nil }

// Emit writes the Progress constructor. The Total / Value mutators
// (SetTotal, Set) are not chained on the Builder, so non-default
// values emit as TODO comments.
func (f *ProgressForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Progress(%q, %t).\n", f.ID, f.Horizontal)
		return err
	}); err != nil {
		return err
	}
	if f.Total > 0 {
		fmt.Fprintf(w, "// TODO: SetTotal(%d) — no Builder setter\n", f.Total)
	}
	if f.Value != 0 {
		fmt.Fprintf(w, "// TODO: Set(%d) — no Builder setter\n", f.Value)
	}
	return nil
}

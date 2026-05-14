package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// SliderForm is the WidgetForm for *Slider used by the designer. The form
// exposes the range, step, and current value as editable integer fields.
type SliderForm struct {
	ComponentForm

	Minimum int `group:"value" label:"Min"`
	Maximum int `group:"value" label:"Max"`
	Step    int `group:"value" label:"Step"`
	Value   int `group:"value" label:"Value"`
}

func (f *SliderForm) Name() string  { return "Slider" }
func (f *SliderForm) Group() string { return "input" }
func (f *SliderForm) Help() string  { return "Horizontal int range input — compact at h=1, rounded box at h≥2" }

func (f *SliderForm) Load(w core.Widget) {
	s := w.(*Slider)
	f.ComponentForm.Load(&s.Component)
	f.Minimum = s.min
	f.Maximum = s.max
	f.Step = s.step
	f.Value = s.value
}

func (f *SliderForm) Store(w core.Widget) {
	s := w.(*Slider)
	f.ComponentForm.Store(&s.Component)
	s.SetMin(f.Minimum)
	s.SetMax(f.Maximum)
	s.SetStep(f.Step)
	s.Set(f.Value)
}

func (f *SliderForm) New() core.Widget {
	s := NewSlider("", "")
	f.Store(s)
	return s
}

func (f *SliderForm) Validate(field string) error {
	if f.Maximum < f.Minimum {
		return fmt.Errorf("max (%d) must be ≥ min (%d)", f.Maximum, f.Minimum)
	}
	if f.Step < 1 {
		return fmt.Errorf("step must be ≥ 1")
	}
	return nil
}

// Emit writes the Slider constructor. Range/Step/Value mutators are not
// chained on the Builder, so non-default values emit as TODO comments.
func (f *SliderForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Slider(%q).\n", f.ID)
		return err
	}); err != nil {
		return err
	}
	if f.Minimum != 0 || f.Maximum != 100 {
		fmt.Fprintf(w, "// TODO: SetMin(%d) / SetMax(%d) — no Builder setter\n", f.Minimum, f.Maximum)
	}
	if f.Step != 1 {
		fmt.Fprintf(w, "// TODO: SetStep(%d) — no Builder setter\n", f.Step)
	}
	if f.Value != 0 {
		fmt.Fprintf(w, "// TODO: Set(%d) — no Builder setter\n", f.Value)
	}
	return nil
}

package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// RadioForm is the WidgetForm for *Radio. Options are edited as a
// comma-separated list of "value:text" pairs, identical to SelectForm,
// so the same option-string syntax works for both widgets.
type RadioForm struct {
	ComponentForm

	OptionsRaw string `group:"value" label:"Options (value:text, comma-separated)"`
	Selected   string `group:"value" label:"Selected"`
}

func (f *RadioForm) Name() string  { return "Radio" }
func (f *RadioForm) Group() string { return "input" }
func (f *RadioForm) Help() string  { return "Mutually-exclusive choice from a fixed list shown inline" }

func (f *RadioForm) Load(w core.Widget) {
	r := w.(*Radio)
	f.ComponentForm.Load(&r.Component)
	parts := make([]string, len(r.options))
	for i, o := range r.options {
		if o.value == o.text {
			parts[i] = o.value
		} else {
			parts[i] = o.value + ":" + o.text
		}
	}
	f.OptionsRaw = strings.Join(parts, ", ")
	if len(r.options) > 0 && r.index >= 0 && r.index < len(r.options) {
		f.Selected = r.options[r.index].value
	}
}

func (f *RadioForm) Store(w core.Widget) {
	r := w.(*Radio)
	f.ComponentForm.Store(&r.Component)
	r.options = parseSelectOptions(f.OptionsRaw)
	r.index = 0
	if f.Selected != "" {
		for i, o := range r.options {
			if o.value == f.Selected {
				r.index = i
				break
			}
		}
	}
}

func (f *RadioForm) New() core.Widget {
	args := selectArgs(parseSelectOptions(f.OptionsRaw))
	r := NewRadio("", "", args...)
	f.Store(r)
	return r
}

func (f *RadioForm) Validate(field string) error { return nil }

// Emit writes the Radio constructor with each option as a pair of
// alternating value/text arguments. The Selected field has no chained
// Builder setter and emits as a TODO comment after the frame.
func (f *RadioForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		opts := parseSelectOptions(f.OptionsRaw)
		args := fmt.Sprintf("%q", f.ID)
		for _, o := range opts {
			args += fmt.Sprintf(", %q, %q", o.value, o.text)
		}
		_, err := fmt.Fprintf(w, "Radio(%s).\n", args)
		return err
	}); err != nil {
		return err
	}
	if f.Selected != "" {
		fmt.Fprintf(w, "// TODO: Select(%q) — no Builder setter\n", f.Selected)
	}
	return nil
}

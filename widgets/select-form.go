package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// SelectForm is the WidgetForm for *Select. Options are edited as a
// comma-separated list of "value:text" pairs. A bare entry without a
// colon is treated as a value where the display text matches the
// value, mirroring the convention used by BuildFormGroup's "select"
// control.
type SelectForm struct {
	ComponentForm

	OptionsRaw string `group:"value" label:"Options (value:text, comma-separated)"`
	Selected   string `group:"value" label:"Selected"`
}

func (f *SelectForm) Name() string  { return "Select" }
func (f *SelectForm) Group() string { return "input" }
func (f *SelectForm) Help() string  { return "Dropdown selection from a fixed list of options" }

func (f *SelectForm) Load(w core.Widget) {
	s := w.(*Select)
	f.ComponentForm.Load(&s.Component)
	parts := make([]string, len(s.options))
	for i, o := range s.options {
		if o.value == o.text {
			parts[i] = o.value
		} else {
			parts[i] = o.value + ":" + o.text
		}
	}
	f.OptionsRaw = strings.Join(parts, ", ")
	if len(s.options) > 0 && s.index >= 0 && s.index < len(s.options) {
		f.Selected = s.options[s.index].value
	}
}

func (f *SelectForm) Store(w core.Widget) {
	s := w.(*Select)
	f.ComponentForm.Store(&s.Component)
	s.options = parseSelectOptions(f.OptionsRaw)
	s.index = 0
	if f.Selected != "" {
		for i, o := range s.options {
			if o.value == f.Selected {
				s.index = i
				break
			}
		}
	}
}

func (f *SelectForm) New() core.Widget {
	args := selectArgs(parseSelectOptions(f.OptionsRaw))
	s := NewSelect("", "", args...)
	f.Store(s)
	return s
}

func (f *SelectForm) Validate(field string) error { return nil }

// Emit writes the Select constructor with each option as a pair of
// alternating value/text arguments. The Selected field has no
// chained Builder setter and emits as a TODO comment after the
// frame.
func (f *SelectForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		opts := parseSelectOptions(f.OptionsRaw)
		args := fmt.Sprintf("%q", f.ID)
		for _, o := range opts {
			args += fmt.Sprintf(", %q, %q", o.value, o.text)
		}
		_, err := fmt.Fprintf(w, "Select(%s).\n", args)
		return err
	}); err != nil {
		return err
	}
	if f.Selected != "" {
		fmt.Fprintf(w, "// TODO: Select(%q) — no Builder setter\n", f.Selected)
	}
	return nil
}

// parseSelectOptions splits a comma-separated "value:text" list into
// option records. Bare entries (no ':') reuse the value as the
// display text.
func parseSelectOptions(raw string) []option {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]option, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		if i := strings.IndexByte(t, ':'); i >= 0 {
			out = append(out, option{value: strings.TrimSpace(t[:i]), text: strings.TrimSpace(t[i+1:])})
		} else {
			out = append(out, option{value: t, text: t})
		}
	}
	return out
}

// selectArgs flattens an []option into the alternating value/text
// strings expected by NewSelect's variadic.
func selectArgs(opts []option) []string {
	out := make([]string, 0, len(opts)*2)
	for _, o := range opts {
		out = append(out, o.value, o.text)
	}
	return out
}

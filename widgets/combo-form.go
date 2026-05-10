package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// ComboForm is the WidgetForm for *Combo. Items are edited as a
// comma-separated string; the current value follows the same
// convention as InputForm.
type ComboForm struct {
	ComponentForm

	ItemsRaw string `group:"value" label:"Items (comma-separated)"`
	Value    string `group:"value" label:"Value"`
}

func (f *ComboForm) Name() string  { return "Combo" }
func (f *ComboForm) Group() string { return "input" }
func (f *ComboForm) Help() string  { return "Combo box with free-text input and suggestion list" }

func (f *ComboForm) Load(w core.Widget) {
	c := w.(*Combo)
	f.ComponentForm.Load(&c.Component)
	f.ItemsRaw = strings.Join(c.items, ", ")
	f.Value = c.value
}

func (f *ComboForm) Store(w core.Widget) {
	c := w.(*Combo)
	f.ComponentForm.Store(&c.Component)
	c.items = parseItems(f.ItemsRaw)
	c.value = f.Value
}

func (f *ComboForm) New() core.Widget {
	c := NewCombo("", "", parseItems(f.ItemsRaw))
	f.Store(c)
	return c
}

func (f *ComboForm) Validate(field string) error { return nil }

// Emit writes the Combo constructor. The Value field has no chained
// Builder setter and is emitted as a TODO comment after the frame.
func (f *ComboForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		items := parseItems(f.ItemsRaw)
		args := fmt.Sprintf("%q", f.ID)
		if len(items) > 0 {
			quoted := make([]string, len(items))
			for i, it := range items {
				quoted[i] = fmt.Sprintf("%q", it)
			}
			args = fmt.Sprintf("%q, %s", f.ID, strings.Join(quoted, ", "))
		}
		_, err := fmt.Fprintf(w, "Combo(%s).\n", args)
		return err
	}); err != nil {
		return err
	}
	if f.Value != "" {
		fmt.Fprintf(w, "// TODO: Set(%q) — no Builder setter\n", f.Value)
	}
	return nil
}

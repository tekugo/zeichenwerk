package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// TabsForm is the WidgetForm for *Tabs. Tab names are edited as a
// comma-separated string; Selected is the active tab index after the
// form is stored. The Tabs widget tracks a separate "highlight"
// (focus-only) index that is runtime-only and not captured here.
type TabsForm struct {
	ComponentForm

	TabsRaw  string `group:"value" label:"Tabs (comma-separated)"`
	Selected int    `group:"value" label:"Selected"`
}

func (f *TabsForm) Name() string  { return "Tabs" }
func (f *TabsForm) Group() string { return "leaf" }
func (f *TabsForm) Help() string  { return "Tab navigation widget" }

func (f *TabsForm) Load(w core.Widget) {
	t := w.(*Tabs)
	f.ComponentForm.Load(&t.Component)
	f.TabsRaw = strings.Join(t.tabs, ", ")
	f.Selected = t.selected
}

func (f *TabsForm) Store(w core.Widget) {
	t := w.(*Tabs)
	f.ComponentForm.Store(&t.Component)
	t.tabs = parseItems(f.TabsRaw)
	if f.Selected < 0 || f.Selected >= len(t.tabs) {
		t.selected = 0
		t.index = 0
	} else {
		t.selected = f.Selected
		t.index = f.Selected
	}
}

func (f *TabsForm) New() core.Widget {
	t := NewTabs("", "")
	for _, name := range parseItems(f.TabsRaw) {
		t.Add(name)
	}
	f.Store(t)
	return t
}

func (f *TabsForm) Validate(field string) error { return nil }

func (f *TabsForm) Emit(w io.Writer, mode string) error {
	if err := f.EmitFrame(w, mode, func() error {
		names := parseItems(f.TabsRaw)
		args := fmt.Sprintf("%q", f.ID)
		for _, n := range names {
			args += fmt.Sprintf(", %q", n)
		}
		_, err := fmt.Fprintf(w, "Tabs(%s).\n", args)
		return err
	}); err != nil {
		return err
	}
	if f.Selected > 0 {
		fmt.Fprintf(w, "// TODO: Set(%d) — no Builder setter\n", f.Selected)
	}
	return nil
}

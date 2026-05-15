package widgets

import (
	"fmt"
	"io"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// ListForm is the WidgetForm for *List.
//
// List exposes its items as a comma-separated string in the editor
// surface — the same convention BuildFormGroup uses for slice-of-
// string fields. ItemsRaw round-trips through Items / strings.Split
// so the editor doesn't need a list-of-strings widget. Selection
// state, scroll offset and the multi-selection set are runtime
// state, not part of the static form.
type ListForm struct {
	ComponentForm

	ItemsRaw    string `group:"value" label:"Items (comma-separated)"`
	Numbers     bool   `group:"display" label:"Show line numbers"`
	Scrollbar   bool   `group:"display" label:"Show scrollbar"`
	QuickSearch bool   `group:"display" label:"Quick search"`
}

func (f *ListForm) Name() string  { return "List" }
func (f *ListForm) Group() string { return "input" }
func (f *ListForm) Help() string  { return "Selectable list of items" }

func (f *ListForm) Load(w core.Widget) {
	l := w.(*List)
	f.ComponentForm.Load(&l.Component)
	f.ItemsRaw = strings.Join(l.items, ", ")
	f.Numbers = l.numbers
	f.Scrollbar = l.scrollbar
	f.QuickSearch = l.quickSearch
}

func (f *ListForm) Store(w core.Widget) {
	l := w.(*List)
	f.ComponentForm.Store(&l.Component)
	l.items = parseItems(f.ItemsRaw)
	l.numbers = f.Numbers
	l.scrollbar = f.Scrollbar
	l.quickSearch = f.QuickSearch
	if l.index >= len(l.items) {
		l.index = 0
	}
}

func (f *ListForm) New() core.Widget {
	l := NewList("", "", parseItems(f.ItemsRaw))
	f.Store(l)
	return l
}

func (f *ListForm) Validate(field string) error { return nil }

// Emit writes the List constructor with each item as a separate
// argument so generated code stays readable. Display flags follow
// as field-style assignments because List has no chained Builder
// setters for them yet — when those land, this falls back to
// chained calls automatically.
func (f *ListForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		items := parseItems(f.ItemsRaw)
		quoted := make([]string, len(items))
		for i, it := range items {
			quoted[i] = fmt.Sprintf("%q", it)
		}
		args := f.ID
		if len(items) > 0 {
			args = fmt.Sprintf("%q, %s", f.ID, strings.Join(quoted, ", "))
		} else {
			args = fmt.Sprintf("%q", f.ID)
		}
		_, err := fmt.Fprintf(w, "List(%s).\n", args)
		return err
	})
}

// parseItems splits a comma-separated user-edited string into the
// item list. Empty / whitespace-only entries are dropped so an
// accidental trailing comma doesn't produce a phantom blank item.
func parseItems(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

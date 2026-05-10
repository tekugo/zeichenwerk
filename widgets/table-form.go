package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// TableForm is the WidgetForm for *Table. The TableProvider is a
// runtime dependency injected by the host application; the static
// editing surface only captures the navigation mode and the inner /
// outer grid line flags.
type TableForm struct {
	ComponentForm

	CellNav bool `group:"general" label:"Cell Navigation"`
	Inner   bool `group:"display" label:"Inner Grid Lines"`
	Outer   bool `group:"display" label:"Outer Grid Lines"`
}

func (f *TableForm) Name() string  { return "Table" }
func (f *TableForm) Group() string { return "leaf" }
func (f *TableForm) Help() string  { return "Tabular data view with row or cell navigation" }

func (f *TableForm) Load(w core.Widget) {
	t := w.(*Table)
	f.ComponentForm.Load(&t.Component)
	f.CellNav = t.cellNav
	f.Inner = t.inner
	f.Outer = t.outer
}

func (f *TableForm) Store(w core.Widget) {
	t := w.(*Table)
	f.ComponentForm.Store(&t.Component)
	t.cellNav = f.CellNav
	t.inner = f.Inner
	t.outer = f.Outer
}

func (f *TableForm) New() core.Widget {
	t := NewTable("", "", emptyTableProvider{}, f.CellNav)
	f.Store(t)
	return t
}

func (f *TableForm) Validate(field string) error { return nil }

// Emit writes the Table constructor with a placeholder provider
// identifier. The TableProvider is application-level state that
// codegen cannot synthesise, so the user is expected to replace
// "tableProvider" with the actual variable.
func (f *TableForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Table(%q, tableProvider /* TODO */, %t).\n", f.ID, f.CellNav)
		return err
	})
}

// emptyTableProvider is a zero-row, zero-column TableProvider used
// as the default when a Table is constructed by the form's New —
// the caller is expected to replace it with a real provider before
// using the widget.
type emptyTableProvider struct{}

func (emptyTableProvider) Length() int              { return 0 }
func (emptyTableProvider) Columns() []TableColumn   { return nil }
func (emptyTableProvider) Str(row, col int) string  { return "" }

package widgets

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/tekugo/zeichenwerk/core"
)

// GridForm is the WidgetForm for *Grid. It also implements
// inspector.ContainerForm by providing LayoutForm so per-child cell
// coordinates can be edited and round-tripped through codegen.
//
// Rows and Columns are exposed as comma-separated strings (e.g.
// "-1, 2, 10") because the property panel's BuildFormGroup does not
// render []int slices today. Negative values are fractional, zero
// uses the child's hint, positive values are fixed sizes — same
// semantics as the underlying Grid. Changing the *count* of values
// resizes the live grid via Grid.Resize: cells whose origin falls
// outside the new bounds are dropped, cells whose span overshoots
// are clamped, and freshly-added rows / columns default to
// fractional sizing (-1).
type GridForm struct {
	ComponentForm

	Rows    string `group:"layout" label:"Rows"    width:"24"`
	Columns string `group:"layout" label:"Columns" width:"24"`
	Lines   bool   `group:"layout" label:"Lines"`
}

func (f *GridForm) Name() string  { return "Grid" }
func (f *GridForm) Group() string { return "container" }
func (f *GridForm) Help() string  { return "Grid container with row/column layout" }

func (f *GridForm) Load(w core.Widget) {
	g := w.(*Grid)
	f.ComponentForm.Load(&g.Component)
	f.Rows = intsCSV(g.rows)
	f.Columns = intsCSV(g.columns)
	f.Lines = g.lines
}

func (f *GridForm) Store(w core.Widget) {
	g := w.(*Grid)
	f.ComponentForm.Store(&g.Component)

	rows, rowsOk := parseIntCSV(f.Rows)
	cols, colsOk := parseIntCSV(f.Columns)

	// Resize first if either count changed; Resize keeps
	// surviving size-config values intact and pads new entries
	// with -1, so the subsequent copy below only overwrites
	// what the user typed.
	newRows := len(g.rows)
	if rowsOk && len(rows) > 0 {
		newRows = len(rows)
	}
	newCols := len(g.columns)
	if colsOk && len(cols) > 0 {
		newCols = len(cols)
	}
	if newRows != len(g.rows) || newCols != len(g.columns) {
		g.Resize(newRows, newCols)
	}

	if rowsOk && len(rows) == len(g.rows) {
		copy(g.rows, rows)
	}
	if colsOk && len(cols) == len(g.columns) {
		copy(g.columns, cols)
	}
	g.lines = f.Lines
}

func (f *GridForm) New() core.Widget {
	rows, _ := parseIntCSV(f.Rows)
	cols, _ := parseIntCSV(f.Columns)
	if len(rows) == 0 {
		rows = []int{-1}
	}
	if len(cols) == 0 {
		cols = []int{-1}
	}
	g := NewGrid("", "", len(rows), len(cols), f.Lines)
	copy(g.rows, rows)
	copy(g.columns, cols)
	f.ComponentForm.Store(&g.Component)
	g.lines = f.Lines
	return g
}

func (f *GridForm) Validate(field string) error { return nil }

// LayoutForm returns a fresh per-child layout form already loaded with
// child's cell coordinates on parent.
func (f *GridForm) LayoutForm(parent core.Container, child core.Widget) core.LayoutForm {
	lf := &GridLayoutForm{}
	lf.Load(parent, child)
	return lf
}

// Emit writes Grid's call shape onto a chain that already exists.
// Children are emitted by the codegen walker after this returns; the
// walker is responsible for the closing ".End()". Rows and Columns
// emit after the standard ComponentForm chain (Hint / flags /
// style) because they are kind-specific tail methods on the Grid.
func (f *GridForm) Emit(w io.Writer, mode string) error {
	rows, _ := parseIntCSV(f.Rows)
	cols, _ := parseIntCSV(f.Columns)

	if err := f.EmitFrame(w, mode, func() error {
		_, err := fmt.Fprintf(w, "Grid(%q, %d, %d, %t).\n",
			f.ID, len(rows), len(cols), f.Lines)
		return err
	}); err != nil {
		return err
	}
	if !defaultGridSizes(rows) {
		fmt.Fprintf(w, "Rows(%s).\n", intsCSV(rows))
	}
	if !defaultGridSizes(cols) {
		fmt.Fprintf(w, "Columns(%s).\n", intsCSV(cols))
	}
	return nil
}

// GridLayoutForm captures one child's cell coordinates on a Grid.
type GridLayoutForm struct {
	X int `group:"position" label:"Column"`
	Y int `group:"position" label:"Row"`
	W int `group:"position" label:"Col Span"`
	H int `group:"position" label:"Row Span"`
}

// Load reads child's cell from grid.
func (f *GridLayoutForm) Load(parent core.Container, child core.Widget) {
	g, ok := parent.(*Grid)
	if !ok {
		return
	}
	for _, c := range g.cells {
		if c.content == child {
			f.X, f.Y, f.W, f.H = c.x, c.y, c.w, c.h
			return
		}
	}
}

// Store mutates grid's cell entry for child to match the form. If child
// is not in grid, Store is a no-op.
func (f *GridLayoutForm) Store(parent core.Container, child core.Widget) {
	g, ok := parent.(*Grid)
	if !ok {
		return
	}
	for _, c := range g.cells {
		if c.content == child {
			c.x, c.y, c.w, c.h = f.X, f.Y, f.W, f.H
			return
		}
	}
}

func (f *GridLayoutForm) Validate(field string) error { return nil }

// Emit writes the Cell prefix that precedes the child's constructor.
// Indentation is gofmt's responsibility; trailing ".\n" continues
// the chain to the child's call.
func (f *GridLayoutForm) Emit(w io.Writer, mode string) error {
	switch mode {
	case "builder":
		fmt.Fprintf(w, "Cell(%d, %d, %d, %d).\n", f.X, f.Y, f.W, f.H)
		return nil
	case "compose":
		return fmt.Errorf("compose mode not implemented")
	}
	return fmt.Errorf("unknown mode %q", mode)
}

// ---- helpers ----

// defaultGridSizes reports whether sizes is the all-fractional default
// (every entry == -1). NewGrid initializes both axes to that, so when
// neither was customized we don't need to emit a .Rows / .Columns call.
func defaultGridSizes(sizes []int) bool {
	for _, s := range sizes {
		if s != -1 {
			return false
		}
	}
	return true
}

// intsCSV returns "1, 2, 3" formatting for variadic int arguments.
func intsCSV(ints []int) string {
	parts := make([]string, len(ints))
	for i, n := range ints {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(parts, ", ")
}

// parseIntCSV parses a comma-separated decimal-integer string into a
// slice. Leading/trailing whitespace and whitespace around each
// element are tolerated. Returns ok=false (and a nil slice) if any
// element fails to parse, so callers that store back can decide
// whether to leave existing state alone. An empty input yields an
// empty slice with ok=true.
func parseIntCSV(s string) ([]int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, true
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}

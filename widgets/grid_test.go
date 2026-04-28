package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

// Ensure Grid implements Container interface.
var _ Container = (*Grid)(nil)

func TestNewGrid(t *testing.T) {
	g := NewGrid("grid1", "cls", 3, 4, true)

	if g.ID() != "grid1" {
		t.Errorf("ID() = %q; want %q", g.ID(), "grid1")
	}
	if len(g.rows) != 3 {
		t.Errorf("rows count = %d; want 3", len(g.rows))
	}
	if len(g.columns) != 4 {
		t.Errorf("columns count = %d; want 4", len(g.columns))
	}
	if !g.lines {
		t.Error("expected lines = true")
	}
	// All rows and columns default to -1 (flexible sizing)
	for i, r := range g.rows {
		if r != -1 {
			t.Errorf("rows[%d] = %d; want -1 (flexible)", i, r)
		}
	}
	for i, c := range g.columns {
		if c != -1 {
			t.Errorf("columns[%d] = %d; want -1 (flexible)", i, c)
		}
	}
	if len(g.cells) != 0 {
		t.Errorf("expected 0 cells initially, got %d", len(g.cells))
	}
}

func TestNewGrid_Separators(t *testing.T) {
	g := NewGrid("grid", "", 2, 3, true)
	for i := range 2 {
		for j := range 3 {
			if g.separators[i][j] != GridB {
				t.Errorf("separators[%d][%d] = %d; want %d (GridB)", i, j, g.separators[i][j], GridB)
			}
		}
	}
}

func TestGrid_Add(t *testing.T) {
	g := NewGrid("grid", "", 3, 3, false)
	c := NewComponent("c1", "")
	g.Add(c, 0, 0, 1, 1)

	if len(g.cells) != 1 {
		t.Errorf("expected 1 cell, got %d", len(g.cells))
	}
	cell := g.cells[0]
	if cell.x != 0 || cell.y != 0 {
		t.Errorf("cell position = (%d,%d); want (0,0)", cell.x, cell.y)
	}
	if cell.w != 1 || cell.h != 1 {
		t.Errorf("cell span = %dx%d; want 1x1", cell.w, cell.h)
	}
}

func TestGrid_Add_SetsParent(t *testing.T) {
	g := NewGrid("grid", "", 2, 2, false)
	c := NewComponent("c", "")
	g.Add(c, 0, 0, 1, 1)
	if c.Parent() != g {
		t.Error("Add() should set the child's parent to the grid")
	}
}

func TestGrid_Add_ZeroSpanDefaultsToOne(t *testing.T) {
	g := NewGrid("grid", "", 3, 3, false)
	c := NewComponent("c", "")
	g.Add(c, 1, 1, 0, 0)

	cell := g.cells[0]
	if cell.w != 1 || cell.h != 1 {
		t.Errorf("zero span should default to 1x1, got %dx%d", cell.w, cell.h)
	}
}

func TestGrid_Add_ClipsColumnPosition(t *testing.T) {
	g := NewGrid("grid", "", 3, 3, false)
	c := NewComponent("c", "")
	g.Add(c, 5, 0, 1, 1) // x=5 out of bounds for 3 columns

	cell := g.cells[0]
	if cell.x != 2 {
		t.Errorf("out-of-bounds x should clip to last column (2), got %d", cell.x)
	}
}

func TestGrid_Add_ClipsRowPosition(t *testing.T) {
	g := NewGrid("grid", "", 3, 3, false)
	c := NewComponent("c", "")
	g.Add(c, 0, 5, 1, 1) // y=5 out of bounds for 3 rows

	cell := g.cells[0]
	if cell.y != 2 {
		t.Errorf("out-of-bounds y should clip to last row (2), got %d", cell.y)
	}
}

func TestGrid_Add_ClipsColumnSpan(t *testing.T) {
	g := NewGrid("grid", "", 3, 3, false)
	c := NewComponent("c", "")
	g.Add(c, 1, 0, 5, 1) // span 5 from col 1 exceeds 3-column grid

	cell := g.cells[0]
	if cell.w != 2 { // 3 - 1 = 2
		t.Errorf("column span should clip to 2, got %d", cell.w)
	}
}

func TestGrid_Add_ClipsRowSpan(t *testing.T) {
	g := NewGrid("grid", "", 3, 3, false)
	c := NewComponent("c", "")
	g.Add(c, 0, 1, 1, 5) // span 5 from row 1 exceeds 3-row grid

	cell := g.cells[0]
	if cell.h != 2 { // 3 - 1 = 2
		t.Errorf("row span should clip to 2, got %d", cell.h)
	}
}

func TestGrid_Add_MultipleWidgets(t *testing.T) {
	g := NewGrid("grid", "", 2, 2, false)
	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	g.Add(c1, 0, 0, 1, 1)
	g.Add(c2, 1, 1, 1, 1)

	if len(g.cells) != 2 {
		t.Errorf("expected 2 cells, got %d", len(g.cells))
	}
}

func TestGrid_Children(t *testing.T) {
	g := NewGrid("grid", "", 3, 3, false)
	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	g.Add(c1, 0, 0, 1, 1)
	g.Add(c2, 1, 1, 1, 1)

	children := g.Children()
	if len(children) != 2 {
		t.Errorf("Children() returned %d items; want 2", len(children))
	}
	if children[0].ID() != "c1" || children[1].ID() != "c2" {
		t.Errorf("Children() returned wrong widgets")
	}
}

func TestGrid_Children_Empty(t *testing.T) {
	g := NewGrid("grid", "", 2, 2, false)
	children := g.Children()
	if len(children) != 0 {
		t.Errorf("Children() on empty grid returned %d items; want 0", len(children))
	}
}

func TestGrid_Columns(t *testing.T) {
	g := NewGrid("grid", "", 1, 3, false)
	g.Columns(10, 20, 30)

	if g.columns[0] != 10 || g.columns[1] != 20 || g.columns[2] != 30 {
		t.Errorf("columns = %v; want [10 20 30]", g.columns)
	}
}

func TestGrid_Columns_WrongCount(t *testing.T) {
	g := NewGrid("grid", "", 1, 3, false)
	original := make([]int, len(g.columns))
	copy(original, g.columns)

	g.Columns(10, 20) // wrong: grid has 3 columns

	for i, c := range g.columns {
		if c != original[i] {
			t.Errorf("columns should be unchanged after wrong-count call; columns[%d] = %d, want %d", i, c, original[i])
		}
	}
}

func TestGrid_Rows(t *testing.T) {
	g := NewGrid("grid", "", 3, 1, false)
	g.Rows(5, 10, 15)

	if g.rows[0] != 5 || g.rows[1] != 10 || g.rows[2] != 15 {
		t.Errorf("rows = %v; want [5 10 15]", g.rows)
	}
}

func TestGrid_Rows_WrongCount(t *testing.T) {
	g := NewGrid("grid", "", 3, 1, false)
	original := make([]int, len(g.rows))
	copy(original, g.rows)

	g.Rows(5, 10) // wrong: grid has 3 rows

	for i, r := range g.rows {
		if r != original[i] {
			t.Errorf("rows should be unchanged after wrong-count call; rows[%d] = %d, want %d", i, r, original[i])
		}
	}
}

func TestGrid_Layout_EqualFractionalColumns(t *testing.T) {
	g := NewGrid("grid", "", 1, 2, false)
	g.SetBounds(0, 0, 60, 20)

	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	g.Add(c1, 0, 0, 1, 1)
	g.Add(c2, 1, 0, 1, 1)
	g.Layout()

	_, _, w1, _ := c1.Bounds()
	_, _, w2, _ := c2.Bounds()

	// 2 equal fractions: 60 / 2 = 30 each (last gets remainder)
	if w1 != 30 {
		t.Errorf("c1 width = %d; want 30", w1)
	}
	if w2 != 30 {
		t.Errorf("c2 width = %d; want 30", w2)
	}
}

func TestGrid_Layout_EqualFractionalRows(t *testing.T) {
	g := NewGrid("grid", "", 2, 1, false)
	g.SetBounds(0, 0, 30, 40)

	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	g.Add(c1, 0, 0, 1, 1)
	g.Add(c2, 0, 1, 1, 1)
	g.Layout()

	_, _, _, h1 := c1.Bounds()
	_, _, _, h2 := c2.Bounds()

	// 2 equal fractions: 40 / 2 = 20 each (last gets remainder)
	if h1 != 20 {
		t.Errorf("c1 height = %d; want 20", h1)
	}
	if h2 != 20 {
		t.Errorf("c2 height = %d; want 20", h2)
	}
}

func TestGrid_Layout_FixedColumns(t *testing.T) {
	g := NewGrid("grid", "", 1, 2, false)
	g.SetBounds(0, 0, 60, 20)
	g.Columns(20, 40)

	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	g.Add(c1, 0, 0, 1, 1)
	g.Add(c2, 1, 0, 1, 1)
	g.Layout()

	x1, _, w1, _ := c1.Bounds()
	x2, _, w2, _ := c2.Bounds()

	if x1 != 0 {
		t.Errorf("c1 x = %d; want 0", x1)
	}
	if w1 != 20 {
		t.Errorf("c1 width = %d; want 20", w1)
	}
	if x2 != 20 {
		t.Errorf("c2 x = %d; want 20", x2)
	}
	if w2 != 40 {
		t.Errorf("c2 width = %d; want 40", w2)
	}
}

func TestGrid_Layout_FixedRows(t *testing.T) {
	g := NewGrid("grid", "", 2, 1, false)
	g.SetBounds(0, 0, 30, 40)
	g.Rows(15, 25)

	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	g.Add(c1, 0, 0, 1, 1)
	g.Add(c2, 0, 1, 1, 1)
	g.Layout()

	_, y1, _, h1 := c1.Bounds()
	_, y2, _, h2 := c2.Bounds()

	if y1 != 0 {
		t.Errorf("c1 y = %d; want 0", y1)
	}
	if h1 != 15 {
		t.Errorf("c1 height = %d; want 15", h1)
	}
	if y2 != 15 {
		t.Errorf("c2 y = %d; want 15", y2)
	}
	if h2 != 25 {
		t.Errorf("c2 height = %d; want 25", h2)
	}
}

func TestGrid_Layout_CellPositions(t *testing.T) {
	g := NewGrid("grid", "", 2, 2, false)
	g.SetBounds(0, 0, 60, 40)
	g.Columns(30, 30)
	g.Rows(20, 20)

	topLeft := NewComponent("tl", "")
	topRight := NewComponent("tr", "")
	bottomLeft := NewComponent("bl", "")
	bottomRight := NewComponent("br", "")

	g.Add(topLeft, 0, 0, 1, 1)
	g.Add(topRight, 1, 0, 1, 1)
	g.Add(bottomLeft, 0, 1, 1, 1)
	g.Add(bottomRight, 1, 1, 1, 1)
	g.Layout()

	checkBounds := func(name string, widget *Component, wantX, wantY, wantW, wantH int) {
		x, y, w, h := widget.Bounds()
		if x != wantX || y != wantY || w != wantW || h != wantH {
			t.Errorf("%s bounds = (%d,%d,%d,%d); want (%d,%d,%d,%d)",
				name, x, y, w, h, wantX, wantY, wantW, wantH)
		}
	}

	checkBounds("topLeft", topLeft, 0, 0, 30, 20)
	checkBounds("topRight", topRight, 30, 0, 30, 20)
	checkBounds("bottomLeft", bottomLeft, 0, 20, 30, 20)
	checkBounds("bottomRight", bottomRight, 30, 20, 30, 20)
}

func TestGrid_Layout_ColumnSpan(t *testing.T) {
	g := NewGrid("grid", "", 1, 3, false)
	g.SetBounds(0, 0, 60, 20)
	g.Columns(20, 20, 20)
	g.Rows(20)

	c := NewComponent("spanning", "")
	g.Add(c, 0, 0, 2, 1) // span 2 columns
	g.Layout()

	_, _, w, _ := c.Bounds()
	if w != 40 { // 20 + 20
		t.Errorf("2-column span width = %d; want 40", w)
	}
}

func TestGrid_Layout_RowSpan(t *testing.T) {
	g := NewGrid("grid", "", 3, 1, false)
	g.SetBounds(0, 0, 30, 60)
	g.Rows(20, 20, 20)
	g.Columns(30)

	c := NewComponent("spanning", "")
	g.Add(c, 0, 0, 1, 2) // span 2 rows
	g.Layout()

	_, _, _, h := c.Bounds()
	if h != 40 { // 20 + 20
		t.Errorf("2-row span height = %d; want 40", h)
	}
}

func TestGrid_Layout_GridLines_ReducesAvailableSpace(t *testing.T) {
	// With lines enabled, grid line pixels are subtracted from available space.
	// 2 columns: 1 grid line between them.
	g := NewGrid("grid", "", 1, 2, true)
	g.SetBounds(0, 0, 61, 10) // 30 + 1 (line) + 30 = 61

	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	g.Add(c1, 0, 0, 1, 1)
	g.Add(c2, 1, 0, 1, 1)
	g.Layout()

	_, _, w1, _ := c1.Bounds()
	_, _, w2, _ := c2.Bounds()

	// Available = 61 - 1 (line) = 60, split equally = 30 each
	if w1 != 30 {
		t.Errorf("c1 width = %d; want 30", w1)
	}
	if w2 != 30 {
		t.Errorf("c2 width = %d; want 30", w2)
	}
}

func TestGrid_Layout_GridLines_ColumnPositions(t *testing.T) {
	g := NewGrid("grid", "", 1, 2, true)
	g.SetBounds(0, 0, 61, 10)
	g.Columns(30, 30)

	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	g.Add(c1, 0, 0, 1, 1)
	g.Add(c2, 1, 0, 1, 1)
	g.Layout()

	x1, _, _, _ := c1.Bounds()
	x2, _, _, _ := c2.Bounds()

	if x1 != 0 {
		t.Errorf("c1 x = %d; want 0", x1)
	}
	// c2 starts after c1 (30) + 1 grid line = 31
	if x2 != 31 {
		t.Errorf("c2 x = %d; want 31", x2)
	}
}

func TestGrid_GridB_Constant(t *testing.T) {
	if GridB != GridH|GridV {
		t.Errorf("GridB (%d) should equal GridH|GridV (%d)", GridB, GridH|GridV)
	}
}

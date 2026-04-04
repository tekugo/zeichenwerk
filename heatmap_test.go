package zeichenwerk

import "testing"

// ---- constructor -----------------------------------------------------------

func TestHeatmap_Defaults(t *testing.T) {
	h := NewHeatmap("h", "", 3, 4)
	if h.rows != 3 || h.cols != 4 {
		t.Errorf("rows=%d cols=%d; want 3,4", h.rows, h.cols)
	}
	if h.cellWidth != 1 {
		t.Errorf("cellWidth=%d; want 1", h.cellWidth)
	}
	for r := 0; r < 3; r++ {
		for c := 0; c < 4; c++ {
			if h.values[r][c] != 0 {
				t.Errorf("values[%d][%d]=%v; want 0", r, c, h.values[r][c])
			}
		}
	}
}

// ---- Hint ------------------------------------------------------------------

func TestHeatmap_Hint_NoLabels(t *testing.T) {
	h := NewHeatmap("h", "", 4, 7)
	w, ht := h.Hint()
	// cols*(cellWidth+1)-1 = 7*2-1 = 13; height = 4
	if w != 13 || ht != 4 {
		t.Errorf("Hint()=(%d,%d); want (13,4)", w, ht)
	}
}

func TestHeatmap_Hint_WithRowLabels(t *testing.T) {
	h := NewHeatmap("h", "", 4, 7)
	h.SetRowLabels([]string{"Mon", "Tue", "Wed", "Thu"}) // width 3 → lcw=4
	w, ht := h.Hint()
	// 4 + 7*2-1 = 4+13 = 17; height = 4
	if w != 17 || ht != 4 {
		t.Errorf("Hint()=(%d,%d); want (17,4)", w, ht)
	}
}

func TestHeatmap_Hint_WithColLabels(t *testing.T) {
	h := NewHeatmap("h", "", 4, 7)
	h.SetColLabels([]string{"A", "B", "C", "D", "E", "F", "G"})
	w, ht := h.Hint()
	// no row labels: lcw=0; cols=7*2-1=13; lrh=1 → height=5
	if w != 13 || ht != 5 {
		t.Errorf("Hint()=(%d,%d); want (13,5)", w, ht)
	}
}

func TestHeatmap_Hint_CellWidth2(t *testing.T) {
	h := NewHeatmap("h", "", 2, 3)
	h.SetCellWidth(2)
	w, ht := h.Hint()
	// 3*(2+1)-1 = 8; height=2
	if w != 8 || ht != 2 {
		t.Errorf("Hint()=(%d,%d); want (8,2)", w, ht)
	}
}

// ---- Data methods ----------------------------------------------------------

func TestHeatmap_SetValue_GetValue(t *testing.T) {
	h := NewHeatmap("h", "", 3, 3)
	h.SetValue(1, 2, 7.5)
	if h.Value(1, 2) != 7.5 {
		t.Errorf("Value(1,2)=%v; want 7.5", h.Value(1, 2))
	}
}

func TestHeatmap_SetValue_OutOfBounds_IsNoop(t *testing.T) {
	h := NewHeatmap("h", "", 2, 2)
	h.SetValue(5, 5, 99) // should not panic
	if h.Value(0, 0) != 0 {
		t.Errorf("out-of-bounds SetValue mutated the grid")
	}
}

func TestHeatmap_SetRow(t *testing.T) {
	h := NewHeatmap("h", "", 2, 3)
	h.SetRow(0, []float64{1, 2, 3})
	if h.Value(0, 0) != 1 || h.Value(0, 1) != 2 || h.Value(0, 2) != 3 {
		t.Errorf("SetRow(0,…) values not set correctly")
	}
}

func TestHeatmap_SetAll(t *testing.T) {
	h := NewHeatmap("h", "", 2, 2)
	h.SetAll([][]float64{{1, 2}, {3, 4}})
	if h.Value(1, 1) != 4 {
		t.Errorf("SetAll: Value(1,1)=%v; want 4", h.Value(1, 1))
	}
}

func TestHeatmap_SetAll_DimensionMismatch_Panics(t *testing.T) {
	h := NewHeatmap("h", "", 2, 2)
	defer func() {
		if recover() == nil {
			t.Error("SetAll with wrong dimensions should panic")
		}
	}()
	h.SetAll([][]float64{{1, 2, 3}}) // wrong row count → panic
}

// ---- Render: colour selection ----------------------------------------------

// newHeatmapRenderer builds a recording screen and renderer with preset styles.
func newHeatmapRenderer() (*cellScreen, *Renderer) {
	cs := newCellScreen()
	return cs, NewRenderer(cs, NewTheme())
}

func renderHeatmap(h *Heatmap, w, ht int) *cellScreen {
	cs, r := newHeatmapRenderer()
	h.SetBounds(0, 0, w, ht)
	h.Render(r)
	return cs
}

// setHeatmapStyles injects known hex colours so tests can assert exact values.
func setHeatmapStyles(h *Heatmap) {
	h.SetStyle("", NewStyle("").WithColors("$fg0", "$bg0"))
	h.SetStyle("zero", NewStyle("zero").WithColors("#111111", "#222222"))
	h.SetStyle("max", NewStyle("max").WithColors("#eeeeee", "#aaaaaa"))
	h.SetStyle("mid", NewStyle("mid").WithColors("", ""))
	h.SetStyle("header", NewStyle("header").WithColors("#888888", "#000000"))
}

func TestHeatmap_AllZero_UsesZeroStyle(t *testing.T) {
	h := NewHeatmap("h", "", 2, 2)
	setHeatmapStyles(h)
	cs := renderHeatmap(h, 3, 2) // 2 cols * 2 - 1 = 3 wide, 2 rows high

	// Every cell (0,0), (2,0), (0,1), (2,1) should have zero bg
	for _, pos := range [][2]int{{0, 0}, {2, 0}, {0, 1}, {2, 1}} {
		bg := cs.bg
		_ = bg
		// Check fg recorded at each cell position
		fg := cs.fgs[pos]
		if fg != "#111111" {
			t.Errorf("cell %v fg=%q; want #111111 (zero style)", pos, fg)
		}
	}
}

func TestHeatmap_SingleNonZero_IsMax(t *testing.T) {
	h := NewHeatmap("h", "", 1, 1)
	setHeatmapStyles(h)
	h.SetValue(0, 0, 5)
	cs, r := newHeatmapRenderer()
	h.SetBounds(0, 0, 1, 1)
	h.Render(r)

	fg := cs.fgs[[2]int{0, 0}]
	if fg != "#eeeeee" {
		t.Errorf("single non-zero cell: fg=%q; want #eeeeee (max style)", fg)
	}
}

func TestHeatmap_MaxCell_UsesMaxStyle(t *testing.T) {
	h := NewHeatmap("h", "", 1, 2)
	setHeatmapStyles(h)
	h.SetValue(0, 0, 5)
	h.SetValue(0, 1, 10)
	// Width: 2*(1+1)-1 = 3
	cs, r := newHeatmapRenderer()
	h.SetBounds(0, 0, 3, 1)
	h.Render(r)

	fg := cs.fgs[[2]int{2, 0}] // col 1 at x=2
	if fg != "#eeeeee" {
		t.Errorf("max cell fg=%q; want #eeeeee", fg)
	}
}

func TestHeatmap_MidCell_IsInterpolated(t *testing.T) {
	h := NewHeatmap("h", "", 1, 2)
	h.SetStyle("", NewStyle("").WithColors("", ""))
	h.SetStyle("zero", NewStyle("zero").WithColors("#000000", "#000000"))
	h.SetStyle("max", NewStyle("max").WithColors("#ffffff", "#ffffff"))
	h.SetStyle("mid", NewStyle("mid").WithColors("", ""))
	h.SetStyle("header", NewStyle("header").WithColors("", ""))

	// v=5, max=10 → frac=0.5 → lerp(0,255,0.5)=128=0x80 → #808080
	h.SetValue(0, 0, 5)
	h.SetValue(0, 1, 10)
	cs, r := newHeatmapRenderer()
	h.SetBounds(0, 0, 3, 1)
	h.Render(r)

	fg := cs.fgs[[2]int{0, 0}] // col 0 (the mid cell)
	if fg != "#808080" {
		t.Errorf("mid cell fg=%q; want #808080 (interpolated)", fg)
	}
}

// ---- Render: label placement -----------------------------------------------

func TestHeatmap_RowLabels_Offset(t *testing.T) {
	h := NewHeatmap("h", "", 2, 1)
	h.SetStyle("", NewStyle("").WithColors("", ""))
	h.SetStyle("zero", NewStyle("zero").WithColors("cell-fg", ""))
	h.SetStyle("max", NewStyle("max").WithColors("cell-fg", ""))
	h.SetStyle("mid", NewStyle("mid").WithColors("", ""))
	h.SetStyle("header", NewStyle("header").WithColors("hdr-fg", ""))

	h.SetRowLabels([]string{"A", "B"}) // lcw = 1+1 = 2
	// Width: lcw + 1*(1+1)-1 = 2+1 = 3; height = 2
	cs, r := newHeatmapRenderer()
	h.SetBounds(0, 0, 3, 2)
	h.Render(r)

	// Row label at (0,0) should be "A"
	if cs.cells[[2]int{0, 0}] != "A" {
		t.Errorf("row label at (0,0)=%q; want A", cs.cells[[2]int{0, 0}])
	}
	// Cell at (2,0) (after lcw=2) should use cell fg
	if cs.fgs[[2]int{2, 0}] != "cell-fg" {
		t.Errorf("cell (2,0) fg=%q; want cell-fg", cs.fgs[[2]int{2, 0}])
	}
}

func TestHeatmap_ColLabels_Offset(t *testing.T) {
	h := NewHeatmap("h", "", 1, 2)
	h.SetStyle("", NewStyle("").WithColors("", ""))
	h.SetStyle("zero", NewStyle("zero").WithColors("cell-fg", ""))
	h.SetStyle("max", NewStyle("max").WithColors("cell-fg", ""))
	h.SetStyle("mid", NewStyle("mid").WithColors("", ""))
	h.SetStyle("header", NewStyle("header").WithColors("hdr-fg", ""))

	h.SetColLabels([]string{"X", "Y"}) // lrh = 1
	// Width: 2*(1+1)-1=3; height: 1+1=2
	cs, r := newHeatmapRenderer()
	h.SetBounds(0, 0, 3, 2)
	h.Render(r)

	// Col label "X" at (0,0), "Y" at (2,0)
	if cs.cells[[2]int{0, 0}] != "X" {
		t.Errorf("col label at (0,0)=%q; want X", cs.cells[[2]int{0, 0}])
	}
	if cs.cells[[2]int{2, 0}] != "Y" {
		t.Errorf("col label at (2,0)=%q; want Y", cs.cells[[2]int{2, 0}])
	}
	// Data row should be at y=1
	if cs.fgs[[2]int{0, 1}] != "cell-fg" {
		t.Errorf("cell (0,1) fg=%q; want cell-fg", cs.fgs[[2]int{0, 1}])
	}
}

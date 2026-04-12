package zeichenwerk

import (
	"testing"
)

// newTestTiles creates a Tiles with a no-op render function and sets bounds
// so that Content() returns useful values.
func newTestTiles(items []any, tileW, tileH, contentW, contentH int) *Tiles {
	t := NewTiles("t", "", func(_ *Renderer, _, _, _, _, _ int, _ any, _, _ bool) {}, tileW, tileH)
	// Set bounds large enough to contain the content.
	t.SetBounds(0, 0, contentW, contentH)
	if len(items) > 0 {
		t.SetItems(items)
	}
	return t
}

func items(n int) []any {
	s := make([]any, n)
	for i := range s {
		s[i] = i
	}
	return s
}

// ---- Constructor -----------------------------------------------------------

func TestTilesPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for tileWidth=0")
		}
	}()
	NewTiles("t", "", nil, 0, 1)
}

func TestTilesDefaults(t *testing.T) {
	ti := NewTiles("t", "", nil, 4, 3)
	if ti.Selected() != -1 {
		t.Errorf("initial index = %d, want -1", ti.Selected())
	}
	if !ti.scrollbar {
		t.Error("scrollbar should be true by default")
	}
	if !ti.Flag(FlagFocusable) {
		t.Error("tiles must be focusable")
	}
}

// ---- cols ------------------------------------------------------------------

func TestTilesCols(t *testing.T) {
	ti := newTestTiles(nil, 10, 3, 40, 20)
	if got := ti.cols(); got != 4 {
		t.Errorf("cols() = %d, want 4 (contentWidth=40, tileWidth=10)", got)
	}
}

func TestTilesColsMinOne(t *testing.T) {
	// content width smaller than tile width → cols() must return 1
	ti := newTestTiles(nil, 20, 3, 10, 20)
	if got := ti.cols(); got != 1 {
		t.Errorf("cols() = %d, want 1 when content narrower than tile", got)
	}
}

func TestTilesColsAfterResize(t *testing.T) {
	ti := newTestTiles(nil, 10, 3, 40, 20)
	if ti.cols() != 4 {
		t.Fatalf("precondition: cols() != 4")
	}
	ti.SetBounds(0, 0, 20, 20)
	if got := ti.cols(); got != 2 {
		t.Errorf("cols() = %d after resize to w=20, want 2", got)
	}
}

// ---- rows ------------------------------------------------------------------

func TestTilesRows(t *testing.T) {
	// 10 items, 4 cols → 3 rows (ceil(10/4))
	ti := newTestTiles(items(10), 10, 3, 40, 20)
	if got := ti.rows(); got != 3 {
		t.Errorf("rows() = %d, want 3", got)
	}
}

func TestTilesRowsEmpty(t *testing.T) {
	ti := newTestTiles(nil, 10, 3, 40, 20)
	if got := ti.rows(); got != 0 {
		t.Errorf("rows() = %d for empty items, want 0", got)
	}
}

// ---- SetItems --------------------------------------------------------------

func TestTilesSetItemsResetsIndex(t *testing.T) {
	ti := newTestTiles(items(5), 10, 3, 40, 20)
	ti.Select(4)
	ti.SetItems(items(3))
	if ti.Selected() != 0 {
		t.Errorf("index = %d after SetItems, want 0", ti.Selected())
	}
	if ti.offsetRow != 0 {
		t.Errorf("offsetRow = %d after SetItems, want 0", ti.offsetRow)
	}
}

func TestTilesSetItemsEmptySetsIndexMinus1(t *testing.T) {
	ti := newTestTiles(items(5), 10, 3, 40, 20)
	ti.SetItems([]any{})
	if ti.Selected() != -1 {
		t.Errorf("index = %d after empty SetItems, want -1", ti.Selected())
	}
}

// ---- Move: column (reading order) -----------------------------------------

func TestTilesMoveRightOneCol(t *testing.T) {
	ti := newTestTiles(items(8), 10, 3, 40, 20) // 4 cols
	ti.Select(0)
	ti.Move(0, 1)
	if ti.Selected() != 1 {
		t.Errorf("after Move(0,1) from 0: index = %d, want 1", ti.Selected())
	}
}

func TestTilesMoveRightWrapsToNextRow(t *testing.T) {
	// 4 cols: item 3 is last in row 0; Move(0,1) should go to item 4 (row 1)
	ti := newTestTiles(items(8), 10, 3, 40, 20)
	ti.Select(3)
	ti.Move(0, 1)
	if ti.Selected() != 4 {
		t.Errorf("after Move(0,1) from 3 (last of row 0): index = %d, want 4", ti.Selected())
	}
}

func TestTilesMoveLeftWrapsFromFirstColToPrevRow(t *testing.T) {
	// item 4 is first in row 1; Move(0,-1) should go to item 3 (last of row 0)
	ti := newTestTiles(items(8), 10, 3, 40, 20)
	ti.Select(4)
	ti.Move(0, -1)
	if ti.Selected() != 3 {
		t.Errorf("after Move(0,-1) from 4 (first of row 1): index = %d, want 3", ti.Selected())
	}
}

func TestTilesMoveRightClampsAtLastItem(t *testing.T) {
	ti := newTestTiles(items(8), 10, 3, 40, 20)
	ti.Select(7) // last item
	ti.Move(0, 1)
	if ti.Selected() != 7 {
		t.Errorf("after Move(0,1) at last item: index = %d, want 7 (clamped)", ti.Selected())
	}
}

func TestTilesMoveLeftClampsAtFirstItem(t *testing.T) {
	ti := newTestTiles(items(8), 10, 3, 40, 20)
	ti.Select(0)
	ti.Move(0, -1)
	if ti.Selected() != 0 {
		t.Errorf("after Move(0,-1) at first item: index = %d, want 0 (clamped)", ti.Selected())
	}
}

// ---- Move: row navigation --------------------------------------------------

func TestTilesMoveDown(t *testing.T) {
	// 4 cols: item 0 is row 0 col 0; Move(1,0) → row 1 col 0 = item 4
	ti := newTestTiles(items(12), 10, 3, 40, 20)
	ti.Select(0)
	ti.Move(1, 0)
	if ti.Selected() != 4 {
		t.Errorf("after Move(1,0) from 0: index = %d, want 4", ti.Selected())
	}
}

func TestTilesMoveDownKeepsColumn(t *testing.T) {
	// 4 cols: item 2 is row 0 col 2; Move(1,0) → row 1 col 2 = item 6
	ti := newTestTiles(items(12), 10, 3, 40, 20)
	ti.Select(2)
	ti.Move(1, 0)
	if ti.Selected() != 6 {
		t.Errorf("after Move(1,0) from 2: index = %d, want 6", ti.Selected())
	}
}

func TestTilesMoveUp(t *testing.T) {
	// item 4 → row 1 col 0; Move(-1,0) → row 0 col 0 = item 0
	ti := newTestTiles(items(12), 10, 3, 40, 20)
	ti.Select(4)
	ti.Move(-1, 0)
	if ti.Selected() != 0 {
		t.Errorf("after Move(-1,0) from 4: index = %d, want 0", ti.Selected())
	}
}

func TestTilesMoveDownClampsAtLastRow(t *testing.T) {
	// 10 items, 4 cols → rows: 0,1,2. item 8 is row 2 col 0.
	ti := newTestTiles(items(10), 10, 3, 40, 20)
	ti.Select(8) // row 2, col 0
	ti.Move(1, 0)
	// target would be row 3 col 0 = index 12, past end → clamp to last = 9
	if ti.Selected() != 9 {
		t.Errorf("after Move(1,0) from row 2: index = %d, want 9 (clamped)", ti.Selected())
	}
}

func TestTilesMoveUpClampsAtFirstRow(t *testing.T) {
	ti := newTestTiles(items(8), 10, 3, 40, 20)
	ti.Select(2) // row 0
	ti.Move(-1, 0)
	if ti.Selected() != 2 {
		t.Errorf("after Move(-1,0) at first row: index = %d, want 2 (clamped)", ti.Selected())
	}
}

// ---- Scroll / adjust -------------------------------------------------------

func TestTilesAdjustScrollsDownWhenSelectionBelowViewport(t *testing.T) {
	// contentH=6, tileH=3 → 2 visible rows. Items in 4 cols.
	// Select item 8 (row 2) → offsetRow should adjust to 1 so row 2 is visible.
	ti := newTestTiles(items(12), 10, 3, 40, 6)
	ti.Select(8) // row 2
	if ti.offsetRow != 1 {
		t.Errorf("offsetRow = %d after selecting row 2, want 1", ti.offsetRow)
	}
}

func TestTilesAdjustScrollsUpWhenSelectionAboveViewport(t *testing.T) {
	ti := newTestTiles(items(12), 10, 3, 40, 6) // 2 visible rows
	ti.Select(8)                                 // scroll down
	ti.Select(0)                                 // scroll back up
	if ti.offsetRow != 0 {
		t.Errorf("offsetRow = %d after selecting row 0, want 0", ti.offsetRow)
	}
}

func TestTilesColChangePreservesOffsetRow(t *testing.T) {
	// Start with 4 cols, scroll to offsetRow=1, then resize to 2 cols.
	// offsetRow should be preserved (adjust re-clamps if needed).
	ti := newTestTiles(items(12), 10, 3, 40, 6) // 4 cols, 2 visible rows
	ti.Select(8)                                 // row 2 → offsetRow=1
	if ti.offsetRow != 1 {
		t.Fatalf("precondition: offsetRow=%d, want 1", ti.offsetRow)
	}
	ti.SetBounds(0, 0, 20, 6) // now 2 cols; row(8) in 2-col grid = 4
	// offsetRow is still 1; adjust will re-evaluate on next Select/Move.
	// The test just checks offsetRow wasn't zeroed out by the resize itself.
	if ti.offsetRow != 1 {
		t.Errorf("offsetRow = %d after resize, want 1 (preserved)", ti.offsetRow)
	}
}

// ---- First / Last ----------------------------------------------------------

func TestTilesFirst(t *testing.T) {
	ti := newTestTiles(items(5), 10, 3, 40, 20)
	ti.SetDisabled([]int{0})
	ti.First()
	if ti.Selected() != 1 {
		t.Errorf("First() with item 0 disabled: index = %d, want 1", ti.Selected())
	}
}

func TestTilesLast(t *testing.T) {
	ti := newTestTiles(items(5), 10, 3, 40, 20)
	ti.SetDisabled([]int{4})
	ti.Last()
	if ti.Selected() != 3 {
		t.Errorf("Last() with item 4 disabled: index = %d, want 3", ti.Selected())
	}
}

// ---- PageUp / PageDown -----------------------------------------------------

func TestTilesPageDown(t *testing.T) {
	// 4 cols, tileH=3, contentH=6 → 2 visible rows → PageDown = Move(2,0)
	ti := newTestTiles(items(20), 10, 3, 40, 6)
	ti.Select(0)
	ti.PageDown()
	// Should move 2 rows down: row 0 col 0 → row 2 col 0 = index 8
	if ti.Selected() != 8 {
		t.Errorf("PageDown from 0: index = %d, want 8", ti.Selected())
	}
}

func TestTilesPageUp(t *testing.T) {
	ti := newTestTiles(items(20), 10, 3, 40, 6)
	ti.Select(8) // row 2
	ti.PageUp()
	// Should move 2 rows up: row 2 col 0 → row 0 col 0 = index 0
	if ti.Selected() != 0 {
		t.Errorf("PageUp from 8: index = %d, want 0", ti.Selected())
	}
}

// ---- Render ----------------------------------------------------------------

func TestTilesRenderCallsRenderFunction(t *testing.T) {
	calls := 0
	render := func(_ *Renderer, _, _, _, _, _ int, _ any, _, _ bool) {
		calls++
	}
	ti := NewTiles("t", "", render, 10, 3)
	ti.SetItems(items(8))

	theme := NewTheme()
	AddUnicodeBorders(theme)
	theme.SetColors(map[string]string{"$fg": "#ffffff", "$bg": "#000000"})
	theme.SetStyles(
		NewStyle("").WithColors("$fg", "$bg").WithMargin(0).WithPadding(0),
		NewStyle("tiles").WithColors("$fg", "$bg"),
	)
	ti.Apply(theme)

	cs := newCellScreen()
	ren := NewRenderer(cs, theme)
	ti.SetBounds(0, 0, 40, 9) // 4 cols × 3 rows visible (9/3=3 rows)
	ti.Render(ren)

	// 3 visible rows × 4 cols = 12 slots, but only 8 items
	if calls != 8 {
		t.Errorf("render called %d times, want 8", calls)
	}
}

func TestTilesRenderPassesCorrectArgs(t *testing.T) {
	type call struct {
		x, y, w, h, index int
		selected           bool
	}
	var calls []call
	render := func(_ *Renderer, x, y, w, h, index int, _ any, selected, _ bool) {
		calls = append(calls, call{x, y, w, h, index, selected})
	}
	ti := NewTiles("t", "", render, 10, 3)
	ti.SetItems(items(6))

	theme := NewTheme()
	AddUnicodeBorders(theme)
	theme.SetColors(map[string]string{"$fg": "#ffffff", "$bg": "#000000"})
	theme.SetStyles(
		NewStyle("").WithColors("$fg", "$bg").WithMargin(0).WithPadding(0),
		NewStyle("tiles").WithColors("$fg", "$bg"),
	)
	ti.Apply(theme)

	cs := newCellScreen()
	ren := NewRenderer(cs, theme)
	ti.SetBounds(0, 0, 40, 9) // 4 cols, 3 visible rows
	ti.Render(ren)

	if len(calls) != 6 {
		t.Fatalf("render called %d times, want 6", len(calls))
	}
	// Item 0: slot (0,0), tileW=10, tileH=3, selected=true (default index=0)
	c0 := calls[0]
	if c0.x != 0 || c0.y != 0 || c0.w != 10 || c0.h != 3 {
		t.Errorf("item 0 slot = (%d,%d,%d,%d), want (0,0,10,3)", c0.x, c0.y, c0.w, c0.h)
	}
	if !c0.selected {
		t.Error("item 0 should be selected (default index=0)")
	}
	// Item 5: row 1 col 1 → slotX=10, slotY=3
	c5 := calls[5]
	if c5.x != 10 || c5.y != 3 {
		t.Errorf("item 5 slot pos = (%d,%d), want (10,3)", c5.x, c5.y)
	}
	if c5.selected {
		t.Error("item 5 should not be selected")
	}
}

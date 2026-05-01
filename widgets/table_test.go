package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// helper: build a small ArrayTableProvider for most tests
func makeProvider() *ArrayTableProvider {
	return NewArrayTableProvider(
		[]string{"Name", "Age", "City"},
		[][]string{
			{"Alice", "30", "New York"},
			{"Bob", "25", "Berlin"},
			{"Carol", "40", "Tokyo"},
			{"Dave", "35", "Paris"},
			{"Eve", "28", "Madrid"},
		},
	)
}

// ── Constructor ───────────────────────────────────────────────────────────────

func TestTable_Defaults(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	if !tbl.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
	row, col := tbl.Selected()
	if row != 0 {
		t.Errorf("Selected row = %d; want 0", row)
	}
	if col != -1 {
		t.Errorf("Selected col = %d in row mode; want -1", col)
	}
}

func TestTable_CellNav_Defaults(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), true)
	row, col := tbl.Selected()
	if row != 0 {
		t.Errorf("Selected row = %d; want 0", row)
	}
	if col != 0 {
		t.Errorf("Selected col = %d in cell mode; want 0", col)
	}
}

// ── Hint ─────────────────────────────────────────────────────────────────────

func TestTable_Hint_MatchesProvider(t *testing.T) {
	p := makeProvider()
	tbl := NewTable("t", "", p, false)
	w, h := tbl.Hint()
	if h != p.Length() {
		t.Errorf("Hint height = %d; want %d (row count)", h, p.Length())
	}
	// width = sum of column widths + separators
	expected := 0
	for i, col := range p.Columns() {
		if i > 0 {
			expected++
		}
		expected += col.Width
	}
	if w != expected {
		t.Errorf("Hint width = %d; want %d", w, expected)
	}
}

// ── Set ───────────────────────────────────────────────────────────────────────

func TestTable_Set_UpdatesProvider(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	p2 := NewArrayTableProvider([]string{"X"}, [][]string{{"a"}, {"b"}})
	tbl.Set(p2)
	_, h := tbl.Hint()
	if h != 2 {
		t.Errorf("Hint height = %d after Set; want 2", h)
	}
}

// ── SetSelected ───────────────────────────────────────────────────────────────

func TestTable_SetSelected_InRange(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	ok := tbl.SetSelected(3, 0)
	if !ok {
		t.Fatal("SetSelected(3,0) returned false; want true")
	}
	row, _ := tbl.Selected()
	if row != 3 {
		t.Errorf("Selected row = %d; want 3", row)
	}
}

func TestTable_SetSelected_OutOfRange_ReturnsFalse(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	ok := tbl.SetSelected(99, 0)
	if ok {
		t.Error("SetSelected(99,0) should return false for out-of-range row")
	}
}

func TestTable_SetSelected_NegativeRow_ReturnsFalse(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	ok := tbl.SetSelected(-1, 0)
	if ok {
		t.Error("SetSelected(-1,0) should return false")
	}
}

func TestTable_SetSelected_CellMode_SetsColumn(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), true)
	tbl.SetSelected(2, 1)
	row, col := tbl.Selected()
	if row != 2 {
		t.Errorf("Selected row = %d; want 2", row)
	}
	if col != 1 {
		t.Errorf("Selected col = %d; want 1", col)
	}
}

func TestTable_SetSelected_CellMode_ClampsColumn(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), true)
	tbl.SetSelected(0, 99)
	_, col := tbl.Selected()
	if col != 2 { // 3 columns → max index 2
		t.Errorf("Selected col = %d after out-of-range; want 2 (clamped)", col)
	}
}

// ── Keyboard — row mode ───────────────────────────────────────────────────────

func TestTable_Keyboard_Down(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.handleKey(BuildKey(tcell.KeyDown))
	row, _ := tbl.Selected()
	if row != 1 {
		t.Errorf("Selected row = %d after Down; want 1", row)
	}
}

func TestTable_Keyboard_Up(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(2, 0)
	tbl.handleKey(BuildKey(tcell.KeyUp))
	row, _ := tbl.Selected()
	if row != 1 {
		t.Errorf("Selected row = %d after Up; want 1", row)
	}
}

func TestTable_Keyboard_Down_AtEnd_NoOp(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(4, 0) // last row (5 rows, index 4)
	tbl.handleKey(BuildKey(tcell.KeyDown))
	row, _ := tbl.Selected()
	if row != 4 {
		t.Errorf("Selected row = %d after Down at end; want 4", row)
	}
}

func TestTable_Keyboard_Up_AtStart_NoOp(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.handleKey(BuildKey(tcell.KeyUp))
	row, _ := tbl.Selected()
	if row != 0 {
		t.Errorf("Selected row = %d after Up at start; want 0", row)
	}
}

func TestTable_Keyboard_Home(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(3, 0)
	tbl.handleKey(BuildKey(tcell.KeyHome))
	row, _ := tbl.Selected()
	if row != 0 {
		t.Errorf("Selected row = %d after Home; want 0", row)
	}
}

func TestTable_Keyboard_End(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.handleKey(BuildKey(tcell.KeyEnd))
	row, _ := tbl.Selected()
	if row != 4 {
		t.Errorf("Selected row = %d after End; want 4", row)
	}
}

func TestTable_Keyboard_PageDown(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 6) // pageSize = max(1, 6-3) = 3
	tbl.handleKey(BuildKey(tcell.KeyPgDn))
	row, _ := tbl.Selected()
	if row <= 0 {
		t.Errorf("Selected row = %d after PageDown; want > 0", row)
	}
}

func TestTable_Keyboard_PageUp(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 6)
	tbl.SetSelected(4, 0)
	tbl.handleKey(BuildKey(tcell.KeyPgUp))
	row, _ := tbl.Selected()
	if row >= 4 {
		t.Errorf("Selected row = %d after PageUp; want < 4", row)
	}
}

// ── Keyboard — EvtActivate (Enter) ────────────────────────────────────────────

func TestTable_Keyboard_Enter_DispatchesActivate(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(1, 0)
	firedRow := -1
	tbl.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		firedRow = data[0].(int)
		return true
	})
	tbl.handleKey(BuildKey(tcell.KeyEnter))
	if firedRow != 1 {
		t.Errorf("EvtActivate row = %d; want 1", firedRow)
	}
}

func TestTable_Keyboard_Enter_IncludesRowData(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(0, 0)
	var firedData []string
	tbl.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		firedData = data[1].([]string)
		return true
	})
	tbl.handleKey(BuildKey(tcell.KeyEnter))
	if len(firedData) == 0 || firedData[0] != "Alice" {
		t.Errorf("EvtActivate data = %v; want [Alice 30 New York]", firedData)
	}
}

// ── Keyboard — EvtSelect (Space) ─────────────────────────────────────────────

func TestTable_Keyboard_Space_DispatchesSelect(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(2, 0)
	firedRow := -1
	tbl.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		firedRow = data[0].(int)
		return true
	})
	tbl.handleKey(tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone))
	if firedRow != 2 {
		t.Errorf("EvtSelect row = %d after Space; want 2", firedRow)
	}
}

// ── Keyboard — cell navigation mode ──────────────────────────────────────────

func TestTable_CellNav_Right(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), true)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(0, 0)
	tbl.handleKey(BuildKey(tcell.KeyRight))
	_, col := tbl.Selected()
	if col != 1 {
		t.Errorf("Selected col = %d after Right; want 1", col)
	}
}

func TestTable_CellNav_Left(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), true)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(0, 2)
	tbl.handleKey(BuildKey(tcell.KeyLeft))
	_, col := tbl.Selected()
	if col != 1 {
		t.Errorf("Selected col = %d after Left; want 1", col)
	}
}

func TestTable_CellNav_Right_AtEnd_NoOp(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), true)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(0, 2) // last col
	tbl.handleKey(BuildKey(tcell.KeyRight))
	_, col := tbl.Selected()
	if col != 2 {
		t.Errorf("Selected col = %d after Right at end; want 2", col)
	}
}

func TestTable_CellNav_Home_MovesToFirstCol(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), true)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(1, 2)
	tbl.handleKey(BuildKey(tcell.KeyHome))
	_, col := tbl.Selected()
	if col != 0 {
		t.Errorf("Selected col = %d after Home in cell mode; want 0", col)
	}
}

func TestTable_CellNav_End_MovesToLastCol(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), true)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetSelected(1, 0)
	tbl.handleKey(BuildKey(tcell.KeyEnd))
	_, col := tbl.Selected()
	if col != 2 {
		t.Errorf("Selected col = %d after End in cell mode; want 2", col)
	}
}

// ── Offset ────────────────────────────────────────────────────────────────────

func TestTable_SetOffset_Clamped(t *testing.T) {
	tbl := NewTable("t", "", makeProvider(), false)
	tbl.SetBounds(0, 0, 40, 10)
	tbl.SetOffset(-5, -5)
	ox, oy := tbl.Offset()
	if ox < 0 || oy < 0 {
		t.Errorf("Offset() = (%d,%d) after negative set; want >= 0", ox, oy)
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestTable_Render_ShowsHeaderAndData(t *testing.T) {
	p := NewArrayTableProvider(
		[]string{"Name"},
		[][]string{{"Alice"}, {"Bob"}},
	)
	tbl := NewTable("t", "", p, false)
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	tbl.SetBounds(0, 0, 10, 6)
	tbl.Render(r)

	// Header "Name" starts at (0,0) — row 0 is inside the component border area
	// Content area starts at row 2 (header + separator)
	found := false
	for x := 0; x < 10; x++ {
		if cs.Get(x, 0) == "N" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'N' (from header 'Name') to appear in row 0")
	}
}

// ── ImplementsInterface ───────────────────────────────────────────────────────

func TestTable_ImplementsWidget(t *testing.T) {
	var _ Widget = (*Table)(nil)
}

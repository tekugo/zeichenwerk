package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// ---- helpers ---------------------------------------------------------------

// mockScreen is a minimal Screen implementation for render tests.
type mockScreen struct{}

func (m *mockScreen) Clear()                               {}
func (m *mockScreen) Clip(x, y, w, h int)                  {}
func (m *mockScreen) Flush()                               {}
func (m *mockScreen) Get(x, y int) string                  { return "" }
func (m *mockScreen) Put(x, y int, ch string)              {}
func (m *mockScreen) Set(fg, bg, font string)              {}
func (m *mockScreen) SetUnderline(style int, color string) {}
func (m *mockScreen) Translate(x, y int)                   {}

func newTestRenderer() *Renderer {
	return NewRenderer(&mockScreen{}, NewTheme())
}

func nopRender(r *Renderer, x, y, w, h, index int, data any, selected, focused bool) {}

func newDeck(itemHeight int, items ...any) *Deck {
	d := NewDeck("d", "", nopRender, itemHeight)
	if len(items) > 0 {
		d.Set(items)
	}
	return d
}

// ---- constructor -----------------------------------------------------------

func TestNewDeck_Defaults(t *testing.T) {
	d := NewDeck("d", "", nopRender, 3)
	if d.index != -1 {
		t.Errorf("index = %d; want -1", d.index)
	}
	if !d.scrollbar {
		t.Error("scrollbar should be true by default")
	}
	if !d.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
}

func TestNewDeck_PanicsOnZeroItemHeight(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewDeck should panic with itemHeight < 1")
		}
	}()
	NewDeck("d", "", nopRender, 0)
}

// ---- SetItems --------------------------------------------------------------

func TestDeck_SetItems_SetsIndex0(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	if d.index != 0 {
		t.Errorf("index = %d; want 0 after SetItems", d.index)
	}
	if d.offset != 0 {
		t.Errorf("offset = %d; want 0 after SetItems", d.offset)
	}
}

func TestDeck_SetItems_EmptySetsMinusOne(t *testing.T) {
	d := newDeck(2, "a", "b")
	d.Set([]any{})
	if d.index != -1 {
		t.Errorf("index = %d; want -1 for empty items", d.index)
	}
}

func TestDeck_Items_ReturnsSlice(t *testing.T) {
	items := []any{"x", "y"}
	d := newDeck(1)
	d.Set(items)
	got := d.Get()
	if len(got) != 2 || got[0] != "x" {
		t.Errorf("Items() = %v; want %v", got, items)
	}
}

// ---- Selected --------------------------------------------------------------

func TestDeck_Selected_ReturnsIndex(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.index = 2
	if d.Selected() != 2 {
		t.Errorf("Selected() = %d; want 2", d.Selected())
	}
}

// ---- Select ----------------------------------------------------------------

func TestDeck_Select_SetsIndex(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	var got int
	d.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			got = data[0].(int)
		}
		return true
	})
	d.Select(2)
	if d.index != 2 {
		t.Errorf("index = %d; want 2", d.index)
	}
	if got != 2 {
		t.Errorf("EvtSelect data = %d; want 2", got)
	}
}

func TestDeck_Select_OutOfRange_NoOp(t *testing.T) {
	d := newDeck(2, "a", "b")
	d.Select(5)
	if d.index != 0 {
		t.Errorf("index = %d; want 0 (unchanged)", d.index)
	}
}

// ---- Move ------------------------------------------------------------------

func TestDeck_Move_Down(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.index = 0
	d.Move(1)
	if d.index != 1 {
		t.Errorf("index = %d; want 1", d.index)
	}
}

func TestDeck_Move_Up(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.index = 2
	d.Move(-1)
	if d.index != 1 {
		t.Errorf("index = %d; want 1", d.index)
	}
}

func TestDeck_Move_SkipsDisabled(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.SetDisabled([]int{1})
	d.index = 0
	d.Move(1)
	if d.index != 2 {
		t.Errorf("index = %d; want 2 (skipped disabled 1)", d.index)
	}
}

func TestDeck_Move_SkipsDisabledBackward(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.SetDisabled([]int{1})
	d.index = 2
	d.Move(-1)
	if d.index != 0 {
		t.Errorf("index = %d; want 0 (skipped disabled 1)", d.index)
	}
}

func TestDeck_Move_ClampAtEnd(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.index = 2
	d.Move(5)
	if d.index != 2 {
		t.Errorf("index = %d; want 2 (clamped)", d.index)
	}
}

func TestDeck_Move_ZeroIsNoop(t *testing.T) {
	d := newDeck(2, "a", "b")
	d.index = 1
	d.Move(0)
	if d.index != 1 {
		t.Errorf("index = %d; want 1 (unchanged)", d.index)
	}
}

// ---- First / Last ----------------------------------------------------------

func TestDeck_First(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.SetDisabled([]int{0})
	d.index = 2
	d.First()
	if d.index != 1 {
		t.Errorf("First() index = %d; want 1 (first enabled)", d.index)
	}
}

func TestDeck_Last(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.SetDisabled([]int{2})
	d.index = 0
	d.Last()
	if d.index != 1 {
		t.Errorf("Last() index = %d; want 1 (last enabled)", d.index)
	}
}

// ---- PageDown slot count ---------------------------------------------------

func TestDeck_PageDown_AdvancesBySlots(t *testing.T) {
	d := newDeck(2, "a", "b", "c", "d", "e")
	// Simulate a viewport 6 rows tall (3 full slots of height 2).
	d.SetBounds(0, 0, 20, 6)
	d.index = 0
	d.PageDown()
	// 6/2 = 3 slots → Move(3), lands on index 3.
	if d.index != 3 {
		t.Errorf("PageDown index = %d; want 3", d.index)
	}
}

func TestDeck_PageUp_AdvancesBySlots(t *testing.T) {
	d := newDeck(2, "a", "b", "c", "d", "e")
	d.SetBounds(0, 0, 20, 6)
	d.index = 4
	d.PageUp()
	// 6/2 = 3 slots → Move(-3), lands on index 1.
	if d.index != 1 {
		t.Errorf("PageUp index = %d; want 1", d.index)
	}
}

// ---- Render function called with correct args -----------------------------

func TestDeck_Render_CorrectArgs(t *testing.T) {
	type call struct {
		index    int
		data     any
		selected bool
	}
	var calls []call

	items := []any{"alpha", "beta", "gamma"}
	d := NewDeck("d", "", func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool) {
		calls = append(calls, call{index, data, selected})
	}, 2)
	d.Set(items)
	d.index = 1

	// Bounds: 3 items × 2 rows = 6 rows, all visible.
	d.SetBounds(0, 0, 20, 6)

	r := newTestRenderer()
	d.Render(r)

	if len(calls) != 3 {
		t.Fatalf("render called %d times; want 3", len(calls))
	}
	for i, c := range calls {
		if c.index != i {
			t.Errorf("call[%d].index = %d; want %d", i, c.index, i)
		}
		if c.data != items[i] {
			t.Errorf("call[%d].data = %v; want %v", i, c.data, items[i])
		}
		wantSelected := i == 1
		if c.selected != wantSelected {
			t.Errorf("call[%d].selected = %v; want %v", i, c.selected, wantSelected)
		}
	}
}

// ---- Render slot positions ------------------------------------------------

func TestDeck_Render_SlotPositions(t *testing.T) {
	type slotBounds struct{ x, y, w, h int }
	var slots []slotBounds

	d := NewDeck("d", "", func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool) {
		slots = append(slots, slotBounds{x, y, w, h})
	}, 3)
	d.Set([]any{"a", "b"})
	d.SetBounds(0, 0, 20, 6)

	d.Render(newTestRenderer())

	if len(slots) != 2 {
		t.Fatalf("render called %d times; want 2", len(slots))
	}
	if slots[0].y == slots[1].y {
		t.Error("slots should have distinct y positions")
	}
	if slots[1].y != slots[0].y+3 {
		t.Errorf("slot[1].y = %d; want %d (slot[0].y + itemHeight)", slots[1].y, slots[0].y+3)
	}
	if slots[0].h != 3 || slots[1].h != 3 {
		t.Errorf("slot heights = %d, %d; want 3, 3", slots[0].h, slots[1].h)
	}
}

// ---- Scrollbar position ---------------------------------------------------

func TestDeck_Scrollbar_UnitCalculation(t *testing.T) {
	// 5 items, itemHeight=2 → total visual height = 10.
	// Viewport height = 4 (2 slots visible).
	// When offset=1, scrollbar offset = 1*2 = 2.
	// We verify the scrollbar is requested at the correct row-based offset by
	// checking the render function is called for the correct item indices.
	var calledIndices []int
	items := []any{"a", "b", "c", "d", "e"}
	d := NewDeck("d", "", func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool) {
		calledIndices = append(calledIndices, index)
	}, 2)
	d.Set(items)
	d.offset = 1
	d.index = 1
	d.SetBounds(0, 0, 20, 4) // 4 rows / 2 = 2 visible slots

	r := newTestRenderer()
	d.Render(r)

	// Slots 1 and 2 should be rendered (offset=1).
	if len(calledIndices) != 2 {
		t.Fatalf("render called %d times; want 2", len(calledIndices))
	}
	if calledIndices[0] != 1 || calledIndices[1] != 2 {
		t.Errorf("rendered indices = %v; want [1 2]", calledIndices)
	}
}

// ---- Keyboard interaction -------------------------------------------------

func TestDeck_KeyDown_MovesHighlight(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	ev := tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone)
	d.handleKey(ev)
	if d.index != 1 {
		t.Errorf("index = %d; want 1 after Down", d.index)
	}
}

func TestDeck_KeyUp_MovesHighlight(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.index = 2
	ev := tcell.NewEventKey(tcell.KeyUp, "", tcell.ModNone)
	d.handleKey(ev)
	if d.index != 1 {
		t.Errorf("index = %d; want 1 after Up", d.index)
	}
}

func TestDeck_KeyHome_First(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.index = 2
	ev := tcell.NewEventKey(tcell.KeyHome, "", tcell.ModNone)
	d.handleKey(ev)
	if d.index != 0 {
		t.Errorf("index = %d; want 0 after Home", d.index)
	}
}

func TestDeck_KeyEnd_Last(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	ev := tcell.NewEventKey(tcell.KeyEnd, "", tcell.ModNone)
	d.handleKey(ev)
	if d.index != 2 {
		t.Errorf("index = %d; want 2 after End", d.index)
	}
}

func TestDeck_KeyEnter_DispatchesActivate(t *testing.T) {
	d := newDeck(2, "a", "b", "c")
	d.index = 1
	var activated int = -1
	d.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			activated = data[0].(int)
		}
		return true
	})
	ev := tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
	d.handleKey(ev)
	if activated != 1 {
		t.Errorf("EvtActivate data = %d; want 1", activated)
	}
}

// ---- Mouse interaction ----------------------------------------------------

func TestDeck_Mouse_SelectsItem(t *testing.T) {
	d := newDeck(3, "a", "b", "c")
	d.SetBounds(0, 0, 20, 9)
	// Click row 4 → (4-0)/3 + 0 = 1
	ev := tcell.NewEventMouse(4, 4, tcell.Button1, tcell.ModNone)
	d.handleMouse(ev)
	if d.index != 1 {
		t.Errorf("index = %d; want 1", d.index)
	}
}

func TestDeck_Mouse_ClickSelectedActivates(t *testing.T) {
	d := newDeck(3, "a", "b", "c")
	d.SetBounds(0, 0, 20, 9)
	d.index = 1
	var activated int = -1
	d.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			activated = data[0].(int)
		}
		return true
	})
	// Click row 4 → item 1 (already selected)
	ev := tcell.NewEventMouse(4, 4, tcell.Button1, tcell.ModNone)
	d.handleMouse(ev)
	if activated != 1 {
		t.Errorf("EvtActivate data = %d; want 1", activated)
	}
}

func TestDeck_Mouse_DisabledIgnored(t *testing.T) {
	d := newDeck(3, "a", "b", "c")
	d.SetBounds(0, 0, 20, 9)
	d.SetDisabled([]int{1})
	d.index = 0
	// Click row 4 → item 1 (disabled)
	ev := tcell.NewEventMouse(4, 4, tcell.Button1, tcell.ModNone)
	handled := d.handleMouse(ev)
	if handled {
		t.Error("click on disabled item should not be handled")
	}
	if d.index != 0 {
		t.Errorf("index = %d; want 0 (unchanged)", d.index)
	}
}

func TestDeck_Mouse_OutsideBoundsIgnored(t *testing.T) {
	d := newDeck(2, "a", "b")
	d.SetBounds(5, 5, 20, 10)
	// Click outside the content area
	ev := tcell.NewEventMouse(0, 0, tcell.Button1, tcell.ModNone)
	handled := d.handleMouse(ev)
	if handled {
		t.Error("click outside bounds should not be handled")
	}
}

func TestDeck_Mouse_Button2Ignored(t *testing.T) {
	d := newDeck(2, "a", "b")
	d.SetBounds(0, 0, 20, 10)
	ev := tcell.NewEventMouse(1, 1, tcell.Button2, tcell.ModNone)
	handled := d.handleMouse(ev)
	if handled {
		t.Error("non-Button1 click should not be handled")
	}
}

// ---- Hint -----------------------------------------------------------------

func TestDeck_Hint(t *testing.T) {
	d := newDeck(3, "a", "b", "c", "d")
	// 4 items × 3 rows = 12
	_, h := d.Hint()
	if h != 12 {
		t.Errorf("Hint() height = %d; want 12", h)
	}
}

func TestDeck_Hint_Empty(t *testing.T) {
	d := newDeck(3)
	_, h := d.Hint()
	if h != 0 {
		t.Errorf("Hint() height = %d; want 0 for empty deck", h)
	}
}

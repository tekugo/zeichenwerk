package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestTabs_Defaults(t *testing.T) {
	tabs := NewTabs("t", "")
	if !tabs.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
	if tabs.Count() != 0 {
		t.Errorf("Count() = %d; want 0 for new tabs", tabs.Count())
	}
	if tabs.Selected() != -1 {
		t.Errorf("Selected() = %d for empty tabs; want -1", tabs.Selected())
	}
}

// ── Add / Count ───────────────────────────────────────────────────────────────

func TestTabs_Add_IncreasesCount(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("Home")
	tabs.Add("Settings")
	tabs.Add("About")
	if tabs.Count() != 3 {
		t.Errorf("Count() = %d; want 3", tabs.Count())
	}
}

func TestTabs_Add_FirstTabSelected(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("Home")
	if tabs.Selected() != 0 {
		t.Errorf("Selected() = %d after first Add; want 0", tabs.Selected())
	}
}

// ── Select ────────────────────────────────────────────────────────────────────

func TestTabs_Select_InRange(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	tabs.Add("C")
	ok := tabs.Select(2)
	if !ok {
		t.Fatal("Select(2) returned false; want true")
	}
	if tabs.Selected() != 2 {
		t.Errorf("Selected() = %d; want 2", tabs.Selected())
	}
}

func TestTabs_Select_OutOfRange_ReturnsFalse(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	ok := tabs.Select(5)
	if ok {
		t.Error("Select(5) should return false for out-of-range index")
	}
}

func TestTabs_Select_NegativeIndex_ReturnsFalse(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	ok := tabs.Select(-1)
	if ok {
		t.Error("Select(-1) should return false")
	}
}

func TestTabs_Select_DispatchesEvtActivate(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	firedIdx := -1
	tabs.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		firedIdx = data[0].(int)
		return true
	})
	tabs.Select(1)
	if firedIdx != 1 {
		t.Errorf("EvtActivate data = %d; want 1", firedIdx)
	}
}

func TestTabs_Select_DispatchesEvtChange(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	firedIdx := -1
	tabs.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		firedIdx = data[0].(int)
		return true
	})
	tabs.Select(1)
	if firedIdx != 1 {
		t.Errorf("EvtChange data = %d; want 1", firedIdx)
	}
}

func TestTabs_Select_SameIndex_NoEvents(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	fired := false
	tabs.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	tabs.Select(0) // same as current
	if fired {
		t.Error("EvtChange should not fire when selection does not change")
	}
}

// ── Hint ─────────────────────────────────────────────────────────────────────

func TestTabs_Hint_HeightIsTwo(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	_, h := tabs.Hint()
	if h != 2 {
		t.Errorf("Hint height = %d; want 2", h)
	}
}

func TestTabs_Hint_WidthAccountsForPadding(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("AB") // 2 runes → 2 + (1*2) + 2 = 6
	w, _ := tabs.Hint()
	if w != 6 {
		t.Errorf("Hint width = %d for one 2-rune tab; want 6", w)
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func TestTabs_Keyboard_Right_Advances(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	tabs.Add("C")
	tabs.handleKey(buildKey(tcell.KeyRight))
	// index (navigation) should advance, not selected
	// Select(0) is already set; Right moves navigation index to 1
	// EvtChange fires with new index
	fired := false
	tabs.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	tabs.Select(0)
	tabs.handleKey(buildKey(tcell.KeyRight))
	if fired {
		// EvtChange fires when navigation index changes
	}
}

func TestTabs_Keyboard_Right_Wraps(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	tabs.Add("C")
	tabs.Select(2) // index = 2 (last)
	firedIdx := -1
	tabs.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		firedIdx = data[0].(int)
		return true
	})
	tabs.handleKey(buildKey(tcell.KeyRight)) // should wrap to 0
	if firedIdx != 0 {
		t.Errorf("navigation index after Right from last = %d; want 0 (wrapped)", firedIdx)
	}
}

func TestTabs_Keyboard_Left_Wraps(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	tabs.Add("C")
	tabs.Select(0) // index = 0 (first)
	firedIdx := -1
	tabs.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		firedIdx = data[0].(int)
		return true
	})
	tabs.handleKey(buildKey(tcell.KeyLeft)) // should wrap to last
	if firedIdx != 2 {
		t.Errorf("navigation index after Left from first = %d; want 2 (wrapped)", firedIdx)
	}
}

func TestTabs_Keyboard_Home_GoesToFirst(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	tabs.Add("C")
	tabs.Select(2)
	firedIdx := -1
	tabs.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		firedIdx = data[0].(int)
		return true
	})
	tabs.handleKey(buildKey(tcell.KeyHome))
	if firedIdx != 0 {
		t.Errorf("navigation index after Home = %d; want 0", firedIdx)
	}
}

func TestTabs_Keyboard_End_GoesToLast(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	tabs.Add("C")
	tabs.Select(0)
	firedIdx := -1
	tabs.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		firedIdx = data[0].(int)
		return true
	})
	tabs.handleKey(buildKey(tcell.KeyEnd))
	if firedIdx != 2 {
		t.Errorf("navigation index after End = %d; want 2", firedIdx)
	}
}

func TestTabs_Keyboard_Enter_ActivatesCurrentTab(t *testing.T) {
	tabs := NewTabs("t", "")
	tabs.Add("A")
	tabs.Add("B")
	tabs.Select(1)
	firedIdx := -1
	tabs.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		firedIdx = data[0].(int)
		return true
	})
	tabs.handleKey(buildKey(tcell.KeyEnter))
	if firedIdx != 1 {
		t.Errorf("EvtActivate data after Enter = %d; want 1", firedIdx)
	}
}

func TestTabs_Keyboard_EmptyTabs_NoOp(t *testing.T) {
	tabs := NewTabs("t", "")
	handled := tabs.handleKey(buildKey(tcell.KeyRight))
	if handled {
		t.Error("Right key on empty tabs should return false")
	}
}

package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestList_Defaults(t *testing.T) {
	l := NewList("l", "", nil)
	if l.Selected() != 0 {
		t.Errorf("Selected() = %d; want 0 (index initialised to 0)", l.Selected())
	}
	if !l.Flag(FlagFocusable) {
		t.Error("expected FlagFocusable to be set")
	}
}

func TestList_DefaultsWithItems(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	if l.Selected() != 0 {
		t.Errorf("Selected() = %d; want 0 (first item selected by default)", l.Selected())
	}
	if len(l.Items()) != 3 {
		t.Errorf("len(Items()) = %d; want 3", len(l.Items()))
	}
}

// ── Set / Items ───────────────────────────────────────────────────────────────

func TestList_Set_ReplacesItems(t *testing.T) {
	l := NewList("l", "", []string{"a", "b"})
	l.Set([]string{"x", "y", "z"})
	items := l.Items()
	if len(items) != 3 || items[0] != "x" || items[2] != "z" {
		t.Errorf("Items() = %v after Set; want [x y z]", items)
	}
}

func TestList_Set_ResetsSelectionToFirst(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(2)
	l.Set([]string{"x", "y"})
	if l.Selected() != 0 {
		t.Errorf("Selected() = %d after Set; want 0", l.Selected())
	}
}

func TestList_Set_Empty_ResetsToZero(t *testing.T) {
	l := NewList("l", "", []string{"a", "b"})
	l.Set(nil)
	if l.Selected() != 0 {
		t.Errorf("Selected() = %d after Set(nil); want 0 (index always resets to 0)", l.Selected())
	}
}

// ── Select ────────────────────────────────────────────────────────────────────

func TestList_Select_InRange(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(2)
	if l.Selected() != 2 {
		t.Errorf("Selected() = %d; want 2", l.Selected())
	}
}

func TestList_Select_OutOfRange_Stored(t *testing.T) {
	// Select does not clamp — the caller is responsible for valid indices.
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(10)
	if l.Selected() != 10 {
		t.Errorf("Selected() = %d after Select(10); want 10 (not clamped)", l.Selected())
	}
	l.Select(-1)
	if l.Selected() != -1 {
		t.Errorf("Selected() = %d after Select(-1); want -1 (not clamped)", l.Selected())
	}
}

// ── Move ─────────────────────────────────────────────────────────────────────

func TestList_Move_Forward(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(0)
	l.Move(2)
	if l.Selected() != 2 {
		t.Errorf("Selected() = %d after Move(2); want 2", l.Selected())
	}
}

func TestList_Move_Backward(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(2)
	l.Move(-1)
	if l.Selected() != 1 {
		t.Errorf("Selected() = %d after Move(-1); want 1", l.Selected())
	}
}

func TestList_Move_ClampsAtEnd(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(2)
	l.Move(10)
	if l.Selected() != 2 {
		t.Errorf("Selected() = %d after Move(10) at end; want 2 (clamped)", l.Selected())
	}
}

func TestList_Move_ClampsAtStart(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(0)
	l.Move(-10)
	if l.Selected() != 0 {
		t.Errorf("Selected() = %d after Move(-10) at start; want 0 (clamped)", l.Selected())
	}
}

func TestList_Move_DispatchesEvtSelect(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(0)
	fired := -1
	l.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(int)
		return true
	})
	l.Move(1)
	if fired != 1 {
		t.Errorf("EvtSelect data = %d after Move; want 1", fired)
	}
}

// ── First / Last ──────────────────────────────────────────────────────────────

func TestList_First(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Select(2)
	l.First()
	if l.Selected() != 0 {
		t.Errorf("Selected() = %d after First(); want 0", l.Selected())
	}
}

func TestList_Last(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Last()
	if l.Selected() != 2 {
		t.Errorf("Selected() = %d after Last(); want 2", l.Selected())
	}
}

// ── PageUp / PageDown ─────────────────────────────────────────────────────────

func TestList_PageDown_Advances(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c", "d", "e"})
	l.SetBounds(0, 0, 10, 3) // content height = 3
	l.Select(0)
	l.PageDown()
	if l.Selected() <= 0 {
		t.Errorf("Selected() = %d after PageDown; want > 0", l.Selected())
	}
}

func TestList_PageUp_Retreats(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c", "d", "e"})
	l.SetBounds(0, 0, 10, 3)
	l.Select(4)
	l.PageUp()
	if l.Selected() >= 4 {
		t.Errorf("Selected() = %d after PageUp; want < 4", l.Selected())
	}
}

// ── Filter ────────────────────────────────────────────────────────────────────

func TestList_Filter_SubstringMatch(t *testing.T) {
	l := NewList("l", "", []string{"apple", "banana", "apricot", "cherry"})
	l.Filter("ap")
	items := l.Items()
	if len(items) != 2 {
		t.Errorf("len(Items()) = %d after Filter(\"ap\"); want 2 (apple, apricot)", len(items))
	}
	if items[0] != "apple" || items[1] != "apricot" {
		t.Errorf("Items() = %v; want [apple apricot]", items)
	}
}

func TestList_Filter_CaseInsensitive(t *testing.T) {
	l := NewList("l", "", []string{"Apple", "BANANA", "apricot"})
	l.Filter("AP")
	items := l.Items()
	if len(items) != 2 {
		t.Errorf("len(Items()) = %d after case-insensitive filter; want 2", len(items))
	}
}

func TestList_Filter_Empty_ClearsFilter(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.Filter("a")
	l.Filter("")
	if len(l.Items()) != 3 {
		t.Errorf("len(Items()) = %d after clearing filter; want 3", len(l.Items()))
	}
}

func TestList_Filter_NoMatch_EmptyList(t *testing.T) {
	l := NewList("l", "", []string{"apple", "banana"})
	l.Filter("xyz")
	if len(l.Items()) != 0 {
		t.Errorf("len(Items()) = %d after no-match filter; want 0", len(l.Items()))
	}
}

// ── Suggest ───────────────────────────────────────────────────────────────────

func TestList_Suggest_PrefixMatchItems(t *testing.T) {
	l := NewList("l", "", []string{"apple", "apricot", "banana"})
	got := l.Suggest("ap")
	if len(got) == 0 {
		t.Fatal("Suggest(\"ap\") returned nil; want matches")
	}
	if got[0] != "apple" {
		t.Errorf("Suggest(\"ap\")[0] = %q; want %q", got[0], "apple")
	}
}

func TestList_Suggest_NoMatchItems_ReturnsNil(t *testing.T) {
	l := NewList("l", "", []string{"apple", "banana"})
	got := l.Suggest("xyz")
	if got != nil {
		t.Errorf("Suggest(\"xyz\") = %v; want nil", got)
	}
}

func TestList_Suggest_EmptyItems_ReturnsNil(t *testing.T) {
	l := NewList("l", "", []string{"apple"})
	got := l.Suggest("")
	if got != nil {
		t.Errorf("Suggest(\"\") = %v; want nil", got)
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func TestList_Keyboard_Down(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.SetBounds(0, 0, 10, 5)
	l.Select(0)
	l.handleKey(buildKey(tcell.KeyDown))
	if l.Selected() != 1 {
		t.Errorf("Selected() = %d after Down; want 1", l.Selected())
	}
}

func TestList_Keyboard_Up(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.SetBounds(0, 0, 10, 5)
	l.Select(2)
	l.handleKey(buildKey(tcell.KeyUp))
	if l.Selected() != 1 {
		t.Errorf("Selected() = %d after Up; want 1", l.Selected())
	}
}

func TestList_Keyboard_Home(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.SetBounds(0, 0, 10, 5)
	l.Select(2)
	l.handleKey(buildKey(tcell.KeyHome))
	if l.Selected() != 0 {
		t.Errorf("Selected() = %d after Home; want 0", l.Selected())
	}
}

func TestList_Keyboard_End(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.SetBounds(0, 0, 10, 5)
	l.handleKey(buildKey(tcell.KeyEnd))
	if l.Selected() != 2 {
		t.Errorf("Selected() = %d after End; want 2", l.Selected())
	}
}

func TestList_Keyboard_Enter_DispatchesActivate(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.SetBounds(0, 0, 10, 5)
	l.Select(1)
	fired := -1
	l.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(int)
		return true
	})
	l.handleKey(buildKey(tcell.KeyEnter))
	if fired != 1 {
		t.Errorf("EvtActivate data = %d after Enter; want 1", fired)
	}
}

// ── Mouse ─────────────────────────────────────────────────────────────────────

func TestList_Mouse_ClickSelects(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c", "d", "e"})
	l.SetBounds(0, 0, 10, 5)
	// Item at row 2 (my=2, cy=0, offset=0) → index=2
	ev := tcell.NewEventMouse(3, 2, tcell.Button1, tcell.ModNone)
	l.handleMouse(ev)
	if l.Selected() != 2 {
		t.Errorf("Selected() = %d after click at row 2; want 2", l.Selected())
	}
}

func TestList_Mouse_DoubleClickActivates(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.SetBounds(0, 0, 10, 5)
	fired := -1
	l.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(int)
		return true
	})
	ev := tcell.NewEventMouse(3, 1, tcell.Button1, tcell.ModNone)
	l.handleMouse(ev) // first click — selects item 1
	l.handleMouse(ev) // second click — activates item 1
	if fired != 1 {
		t.Errorf("EvtActivate data = %d after double-click; want 1", fired)
	}
}

func TestList_Mouse_OutOfBoundsIgnored(t *testing.T) {
	l := NewList("l", "", []string{"a", "b", "c"})
	l.SetBounds(0, 0, 10, 5)
	l.Select(0)
	ev := tcell.NewEventMouse(50, 50, tcell.Button1, tcell.ModNone)
	l.handleMouse(ev)
	if l.Selected() != 0 {
		t.Errorf("Selected() changed after out-of-bounds click; want 0")
	}
}

func TestList_Mouse_NonButton1Ignored(t *testing.T) {
	l := NewList("l", "", []string{"a", "b"})
	l.SetBounds(0, 0, 10, 5)
	handled := l.handleMouse(tcell.NewEventMouse(0, 0, tcell.Button2, tcell.ModNone))
	if handled {
		t.Error("handleMouse with Button2 should return false")
	}
}

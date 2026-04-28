package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestSwitcher_Defaults(t *testing.T) {
	s := NewSwitcher("sw", "")
	if len(s.Children()) != 0 {
		t.Errorf("Children() = %d; want 0 for new switcher", len(s.Children()))
	}
}

// ── Add ───────────────────────────────────────────────────────────────────────

func TestSwitcher_Add_FirstPaneVisible(t *testing.T) {
	s := NewSwitcher("sw", "")
	p1 := NewStatic("p1", "", "Pane 1")
	if err := s.Add(p1); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if p1.Flag(FlagHidden) {
		t.Error("first pane should be visible (FlagHidden=false)")
	}
}

func TestSwitcher_Add_SecondPaneHidden(t *testing.T) {
	s := NewSwitcher("sw", "")
	p1 := NewStatic("p1", "", "Pane 1")
	p2 := NewStatic("p2", "", "Pane 2")
	s.Add(p1)
	s.Add(p2)
	if !p2.Flag(FlagHidden) {
		t.Error("second pane should be hidden (FlagHidden=true)")
	}
}

func TestSwitcher_Add_NilReturnsError(t *testing.T) {
	s := NewSwitcher("sw", "")
	err := s.Add(nil)
	if err == nil {
		t.Error("Add(nil) should return ErrChildIsNil")
	}
}

func TestSwitcher_Add_SetsParent(t *testing.T) {
	s := NewSwitcher("sw", "")
	p := NewStatic("p", "", "")
	s.Add(p)
	if p.Parent() != s {
		t.Error("parent of added pane should be the switcher")
	}
}

// ── Children ─────────────────────────────────────────────────────────────────

func TestSwitcher_Children_ReturnsAll(t *testing.T) {
	s := NewSwitcher("sw", "")
	s.Add(NewStatic("p1", "", "A"))
	s.Add(NewStatic("p2", "", "B"))
	s.Add(NewStatic("p3", "", "C"))
	if len(s.Children()) != 3 {
		t.Errorf("Children() len = %d; want 3", len(s.Children()))
	}
}

// ── Select ────────────────────────────────────────────────────────────────────

func TestSwitcher_Select_ByIndex_HidesPrevious(t *testing.T) {
	s := NewSwitcher("sw", "")
	p1 := NewStatic("p1", "", "A")
	p2 := NewStatic("p2", "", "B")
	s.Add(p1)
	s.Add(p2)
	s.Select(1)
	if !p1.Flag(FlagHidden) {
		t.Error("previously visible pane should be hidden after Select(1)")
	}
	if p2.Flag(FlagHidden) {
		t.Error("newly selected pane should be visible")
	}
}

func TestSwitcher_Select_ByID(t *testing.T) {
	s := NewSwitcher("sw", "")
	p1 := NewStatic("first", "", "A")
	p2 := NewStatic("second", "", "B")
	s.Add(p1)
	s.Add(p2)
	s.Select("second")
	if !p1.Flag(FlagHidden) {
		t.Error("p1 should be hidden after Select(\"second\")")
	}
	if p2.Flag(FlagHidden) {
		t.Error("p2 should be visible after Select(\"second\")")
	}
}

func TestSwitcher_Select_ByID_Unknown_NoOp(t *testing.T) {
	s := NewSwitcher("sw", "")
	p1 := NewStatic("p1", "", "A")
	s.Add(p1)
	s.Select("nonexistent")
	if p1.Flag(FlagHidden) {
		t.Error("visible pane should remain visible after selecting unknown ID")
	}
}

func TestSwitcher_Select_DispatchesEvtHideAndEvtShow(t *testing.T) {
	s := NewSwitcher("sw", "")
	p1 := NewStatic("p1", "", "A")
	p2 := NewStatic("p2", "", "B")
	s.Add(p1)
	s.Add(p2)

	hideFired := false
	showFired := false
	p1.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		hideFired = true
		return true
	})
	p2.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		showFired = true
		return true
	})
	s.Select(1)
	if !hideFired {
		t.Error("EvtHide should fire on previously visible pane")
	}
	if !showFired {
		t.Error("EvtShow should fire on newly visible pane")
	}
}

// ── Hint ─────────────────────────────────────────────────────────────────────

func TestSwitcher_Hint_MaxOfChildren(t *testing.T) {
	s := NewSwitcher("sw", "")
	p1 := NewStatic("p1", "", "Hello")
	p2 := NewStatic("p2", "", "Hi there, world!")
	p1.SetHint(5, 3)
	p2.SetHint(16, 1)
	s.Add(p1)
	s.Add(p2)
	w, h := s.Hint()
	if w != 16 {
		t.Errorf("Hint width = %d; want 16 (max of children)", w)
	}
	if h != 3 {
		t.Errorf("Hint height = %d; want 3 (max of children)", h)
	}
}

func TestSwitcher_Hint_EmptyIsZero(t *testing.T) {
	s := NewSwitcher("sw", "")
	w, h := s.Hint()
	if w != 0 || h != 0 {
		t.Errorf("Hint() = (%d,%d) for empty switcher; want (0,0)", w, h)
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestSwitcher_Render_ShowsSelectedPane(t *testing.T) {
	s := NewSwitcher("sw", "")
	p1 := NewStatic("p1", "", "AAA")
	p2 := NewStatic("p2", "", "BBB")
	s.Add(p1)
	s.Add(p2)
	s.SetBounds(0, 0, 10, 1)
	s.Select(1)

	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	s.Render(r)

	got := cs.Get(0, 0) + cs.Get(1, 0) + cs.Get(2, 0)
	if got != "BBB" {
		t.Errorf("rendered text = %q; want %q (selected pane content)", got, "BBB")
	}
}

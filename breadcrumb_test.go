package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestBreadcrumb_Defaults(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	if bc.selected != -1 {
		t.Errorf("selected = %d; want -1", bc.selected)
	}
	if bc.first != 0 {
		t.Errorf("first = %d; want 0", bc.first)
	}
	if bc.separator == "" {
		t.Error("separator should have a default value")
	}
	if bc.overflow == "" {
		t.Error("overflow should have a default value")
	}
	if !bc.Flag(FlagFocusable) {
		t.Error("expected FlagFocusable to be set")
	}
}

// ── Hint ──────────────────────────────────────────────────────────────────────

func TestBreadcrumb_Hint_Empty(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	w, h := bc.Hint()
	if w != 0 {
		t.Errorf("Hint() width = %d; want 0 for empty breadcrumb", w)
	}
	if h < 1 {
		t.Errorf("Hint() height = %d; want >= 1", h)
	}
}

func TestBreadcrumb_Hint_Natural(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.separator = " > " // 3 chars
	bc.SetSegments([]string{"Home", "Projects", "zeichenwerk"})
	// "Home" (4) + " > " (3) + "Projects" (8) + " > " (3) + "zeichenwerk" (11) = 29
	w, _ := bc.Hint()
	if w != 29 {
		t.Errorf("Hint() width = %d; want 29", w)
	}
}

func TestBreadcrumb_Hint_SingleSegment(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.separator = " / "
	bc.SetSegments([]string{"Home"})
	w, _ := bc.Hint()
	if w != 4 {
		t.Errorf("Hint() width = %d; want 4", w)
	}
}

// ── SetSegments / Push / Pop / Truncate ───────────────────────────────────────

func TestBreadcrumb_SetSegments_ClampsSelected(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c"})
	bc.selected = 2
	bc.SetSegments([]string{"x"})
	if bc.selected != 0 {
		t.Errorf("selected = %d after shrink; want 0", bc.selected)
	}
	if bc.first != 0 {
		t.Errorf("first = %d after SetSegments; want 0", bc.first)
	}
}

func TestBreadcrumb_SetSegments_ResetsFirst(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.first = 3
	bc.SetSegments([]string{"a", "b"})
	if bc.first != 0 {
		t.Errorf("first = %d after SetSegments; want 0", bc.first)
	}
}

func TestBreadcrumb_Push(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.Push("a")
	bc.Push("b")
	if len(bc.segments) != 2 {
		t.Errorf("len(segments) = %d; want 2", len(bc.segments))
	}
	if bc.segments[1] != "b" {
		t.Errorf("segments[1] = %q; want %q", bc.segments[1], "b")
	}
}

func TestBreadcrumb_Pop_ReturnsLast(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c"})
	seg := bc.Pop()
	if seg != "c" {
		t.Errorf("Pop() = %q; want %q", seg, "c")
	}
	if len(bc.segments) != 2 {
		t.Errorf("len(segments) = %d after Pop; want 2", len(bc.segments))
	}
}

func TestBreadcrumb_Pop_ClampsSelected(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c"})
	bc.selected = 2
	bc.Pop()
	if bc.selected != 1 {
		t.Errorf("selected = %d after Pop; want 1", bc.selected)
	}
}

func TestBreadcrumb_Pop_Empty(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	seg := bc.Pop()
	if seg != "" {
		t.Errorf("Pop() on empty = %q; want empty string", seg)
	}
}

func TestBreadcrumb_Truncate(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c", "d"})
	bc.Truncate(1)
	if len(bc.segments) != 2 {
		t.Errorf("len(segments) = %d after Truncate(1); want 2", len(bc.segments))
	}
	if bc.segments[1] != "b" {
		t.Errorf("segments[1] = %q; want %q", bc.segments[1], "b")
	}
}

func TestBreadcrumb_Truncate_ClampsSelected(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c", "d"})
	bc.selected = 3
	bc.Truncate(1)
	if bc.selected != 1 {
		t.Errorf("selected = %d after Truncate(1); want 1", bc.selected)
	}
}

// ── Select ────────────────────────────────────────────────────────────────────

func TestBreadcrumb_Select_Clamps(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c"})
	bc.Select(10)
	if bc.selected != 2 {
		t.Errorf("selected = %d after Select(10); want 2", bc.selected)
	}
	bc.Select(-5)
	if bc.selected != 0 {
		t.Errorf("selected = %d after Select(-5); want 0", bc.selected)
	}
}

func TestBreadcrumb_Select_DecreasesFirst(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c", "d"})
	bc.first = 2
	bc.selected = 2
	bc.Select(0)
	if bc.first != 0 {
		t.Errorf("first = %d after Select(0) with first=2; want 0", bc.first)
	}
}

func TestBreadcrumb_Select_DoesNotIncreaseFirst(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c", "d"})
	bc.first = 1
	bc.selected = 1
	bc.Select(3)
	if bc.first != 1 {
		t.Errorf("first = %d after selecting forward; want 1 (unchanged)", bc.first)
	}
}

func TestBreadcrumb_Select_EmptySegments(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.Select(0)
	if bc.selected != -1 {
		t.Errorf("selected = %d on empty breadcrumb; want -1", bc.selected)
	}
}

func TestBreadcrumb_Select_DispatchesEvtSelect(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c"})
	fired := -1
	bc.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(int)
		return true
	})
	bc.Select(2)
	if fired != 2 {
		t.Errorf("EvtSelect data = %d; want 2", fired)
	}
}

// ── computeFirstVis ───────────────────────────────────────────────────────────

func TestBreadcrumb_ComputeFirstVis_AllFit(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.separator = " / " // 3 chars
	bc.overflow = "…"    // 1 char
	bc.SetSegments([]string{"a", "b", "c"})
	// "a / b / c" = 1+3+1+3+1 = 9
	start := bc.computeFirstVis(20)
	if start != 0 {
		t.Errorf("computeFirstVis(20) = %d; want 0", start)
	}
}

func TestBreadcrumb_ComputeFirstVis_Collapses(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.separator = " / " // 3 chars
	bc.overflow = "…"    // 1 char
	bc.SetSegments([]string{"Home", "Projects", "zeichenwerk", "spec"})
	// Test with width just enough for last 2: "… / zeichenwerk / spec" = 1+3+11+3+4 = 22
	start := bc.computeFirstVis(22)
	if start != 2 {
		t.Errorf("computeFirstVis(22) = %d; want 2", start)
	}
}

func TestBreadcrumb_ComputeFirstVis_AtLeastOne(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.separator = " / "
	bc.overflow = "…"
	bc.SetSegments([]string{"very-long-segment-one", "very-long-segment-two", "very-long-segment-three"})
	// Width too narrow to fit anything properly — should still show the last segment.
	start := bc.computeFirstVis(3)
	if start != len(bc.segments)-1 {
		t.Errorf("computeFirstVis(3) = %d; want %d (last segment)", start, len(bc.segments)-1)
	}
}

// ── EvtActivate via Enter ─────────────────────────────────────────────────────

func TestBreadcrumb_Enter_DispatchesActivate(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c"})
	bc.selected = 1
	fired := -1
	bc.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(int)
		return true
	})
	bc.handleKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	if fired != 1 {
		t.Errorf("EvtActivate data = %d after Enter; want 1", fired)
	}
}

// ── EvtFocus auto-select ──────────────────────────────────────────────────────

func TestBreadcrumb_Focus_AutoSelectsLast(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c"})
	// selected starts at -1 after SetSegments (no prior selection)
	bc.selected = -1
	bc.Dispatch(bc, EvtFocus)
	if bc.selected != 2 {
		t.Errorf("selected = %d after EvtFocus with selected=-1; want 2 (last)", bc.selected)
	}
}

func TestBreadcrumb_Focus_NoAutoSelectWhenAlreadySelected(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.SetSegments([]string{"a", "b", "c"})
	bc.selected = 1
	bc.Dispatch(bc, EvtFocus)
	if bc.selected != 1 {
		t.Errorf("selected = %d after EvtFocus with selected=1; want 1 (unchanged)", bc.selected)
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func newBreadcrumbRenderer() (*cellScreen, *Renderer) {
	cs := newCellScreen()
	return cs, NewRenderer(cs, NewTheme())
}

func renderBreadcrumb(bc *Breadcrumb, w, h int) *cellScreen {
	cs, r := newBreadcrumbRenderer()
	bc.SetBounds(0, 0, w, h)
	bc.Render(r)
	return cs
}

func TestBreadcrumb_Render_SegmentsPresent(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.separator = " / "
	bc.overflow = "..."
	bc.SetSegments([]string{"Home", "spec"})
	cs := renderBreadcrumb(bc, 40, 1)
	// Expect "Home" to appear at x=0
	got := ""
	for i := 0; i < 4; i++ {
		ch := cs.Get(i, 0)
		if ch == "" {
			ch = " "
		}
		got += ch
	}
	if got != "Home" {
		t.Errorf("cols 0-3 = %q; want %q", got, "Home")
	}
}

func TestBreadcrumb_Render_OverflowPrefix(t *testing.T) {
	bc := NewBreadcrumb("bc", "")
	bc.separator = "/"
	bc.overflow = "…"
	bc.first = 1
	bc.SetSegments([]string{"very-long", "spec"})
	// Width 6: "…/spec" = 1+1+4 = 6
	cs := renderBreadcrumb(bc, 6, 1)
	if cs.Get(0, 0) != "…" {
		t.Errorf("col 0 = %q; want overflow marker %q", cs.Get(0, 0), "…")
	}
}

func TestBreadcrumb_Render_SeparatorStyle(t *testing.T) {
	theme := NewTheme()
	theme.SetStyles(
		NewStyle("breadcrumb/separator").WithColors("sepfg", "sepbg"),
		NewStyle("breadcrumb/segment").WithColors("segfg", "segbg"),
	)
	bc := NewBreadcrumb("bc", "")
	bc.Apply(theme)
	bc.separator = "/"
	bc.SetSegments([]string{"a", "b"})

	cs, r := newBreadcrumbRenderer()
	bc.SetBounds(0, 0, 10, 1)
	bc.Render(r)

	// "a" at col 0, "/" at col 1, "b" at col 2
	sepFg := cs.fgs[[2]int{1, 0}]
	if sepFg != "sepfg" {
		t.Errorf("separator fg = %q at col 1; want %q", sepFg, "sepfg")
	}
}

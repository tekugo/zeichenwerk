package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestSlider_Defaults(t *testing.T) {
	s := NewSlider("s", "")
	if s.Value() != 0 {
		t.Errorf("Value() = %d; want 0", s.Value())
	}
	if s.Min() != 0 {
		t.Errorf("Min() = %d; want 0", s.Min())
	}
	if s.Max() != 100 {
		t.Errorf("Max() = %d; want 100", s.Max())
	}
	if s.Step() != 1 {
		t.Errorf("Step() = %d; want 1", s.Step())
	}
	if !s.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
}

// ── Set / clamp ───────────────────────────────────────────────────────────────

func TestSlider_Set_ClampsLow(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(-50)
	if s.Value() != 0 {
		t.Errorf("Value() = %d; want 0 (clamped to min)", s.Value())
	}
}

func TestSlider_Set_ClampsHigh(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(500)
	if s.Value() != 100 {
		t.Errorf("Value() = %d; want 100 (clamped to max)", s.Value())
	}
}

func TestSlider_Set_DispatchesChange(t *testing.T) {
	s := NewSlider("s", "")
	var got int
	fired := false
	s.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		got = data[0].(int)
		fired = true
		return true
	})
	s.Set(42)
	if !fired {
		t.Fatal("EvtChange should fire on Set")
	}
	if got != 42 {
		t.Errorf("EvtChange payload = %d; want 42", got)
	}
}

func TestSlider_Set_NoDispatchWhenUnchanged(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(20)
	count := 0
	s.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		count++
		return true
	})
	s.Set(20)
	if count != 0 {
		t.Errorf("EvtChange fired %d times; want 0 when value unchanged", count)
	}
}

// ── Bounds setters ────────────────────────────────────────────────────────────

func TestSlider_SetMax_ReclampsValue(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(80)
	s.SetMax(50)
	if s.Value() != 50 {
		t.Errorf("Value() = %d after SetMax(50); want 50", s.Value())
	}
}

func TestSlider_SetMin_ReclampsValue(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(20)
	s.SetMin(40)
	if s.Value() != 40 {
		t.Errorf("Value() = %d after SetMin(40); want 40", s.Value())
	}
}

func TestSlider_SetStep_ClampsToOne(t *testing.T) {
	s := NewSlider("s", "")
	s.SetStep(0)
	if s.Step() != 1 {
		t.Errorf("Step() = %d after SetStep(0); want 1", s.Step())
	}
	s.SetStep(-5)
	if s.Step() != 1 {
		t.Errorf("Step() = %d after SetStep(-5); want 1", s.Step())
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func TestSlider_Keyboard_RightIncreases(t *testing.T) {
	s := NewSlider("s", "")
	s.SetStep(5)
	s.handleKey(BuildKey(tcell.KeyRight))
	if s.Value() != 5 {
		t.Errorf("Value() = %d after KeyRight; want 5", s.Value())
	}
}

func TestSlider_Keyboard_LeftDecreases(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(20)
	s.SetStep(5)
	s.handleKey(BuildKey(tcell.KeyLeft))
	if s.Value() != 15 {
		t.Errorf("Value() = %d after KeyLeft; want 15", s.Value())
	}
}

func TestSlider_Keyboard_HomeEnd(t *testing.T) {
	s := NewSlider("s", "")
	s.handleKey(BuildKey(tcell.KeyEnd))
	if s.Value() != 100 {
		t.Errorf("End should jump to max; got %d", s.Value())
	}
	s.handleKey(BuildKey(tcell.KeyHome))
	if s.Value() != 0 {
		t.Errorf("Home should jump to min; got %d", s.Value())
	}
}

func TestSlider_Keyboard_VimKeys(t *testing.T) {
	s := NewSlider("s", "")
	s.SetStep(3)
	s.handleKey(BuildRune("l"))
	if s.Value() != 3 {
		t.Errorf("l should increase by step; got %d", s.Value())
	}
	s.handleKey(BuildRune("h"))
	if s.Value() != 0 {
		t.Errorf("h should decrease by step; got %d", s.Value())
	}
}

func TestSlider_Keyboard_Readonly_Ignored(t *testing.T) {
	s := NewSlider("s", "")
	s.SetFlag(FlagReadonly, true)
	if s.handleKey(BuildKey(tcell.KeyRight)) {
		t.Error("readonly slider should not handle keys")
	}
	if s.Value() != 0 {
		t.Errorf("Value() = %d; want 0 (readonly)", s.Value())
	}
}

// ── Mouse ─────────────────────────────────────────────────────────────────────

func TestSlider_Mouse_Click_Compact_LeftEdge(t *testing.T) {
	s := NewSlider("s", "")
	s.SetBounds(0, 0, 21, 1)
	s.handleMouse(tcell.NewEventMouse(0, 0, tcell.Button1, tcell.ModNone))
	if s.Value() != 0 {
		t.Errorf("click at left edge → %d; want 0", s.Value())
	}
}

func TestSlider_Mouse_Click_Compact_RightEdge(t *testing.T) {
	s := NewSlider("s", "")
	s.SetBounds(0, 0, 21, 1)
	s.handleMouse(tcell.NewEventMouse(20, 0, tcell.Button1, tcell.ModNone))
	if s.Value() != 100 {
		t.Errorf("click at right edge → %d; want 100", s.Value())
	}
}

func TestSlider_Mouse_Click_Compact_Middle(t *testing.T) {
	s := NewSlider("s", "")
	s.SetBounds(0, 0, 21, 1) // 21 cells, middle is column 10 → 50%
	s.handleMouse(tcell.NewEventMouse(10, 0, tcell.Button1, tcell.ModNone))
	if s.Value() != 50 {
		t.Errorf("click at middle → %d; want 50", s.Value())
	}
}

func TestSlider_Mouse_Click_Box_LeftEdgeIgnoresCorner(t *testing.T) {
	s := NewSlider("s", "")
	s.SetBounds(0, 0, 12, 2) // box style; inner track is columns 1..10
	// Click on left corner column (x=0) — clamped into inner track at col 0 → min.
	s.handleMouse(tcell.NewEventMouse(0, 0, tcell.Button1, tcell.ModNone))
	if s.Value() != 0 {
		t.Errorf("box click on left corner → %d; want 0", s.Value())
	}
}

func TestSlider_Mouse_OutOfBounds_Ignored(t *testing.T) {
	s := NewSlider("s", "")
	s.SetBounds(0, 0, 10, 1)
	if s.handleMouse(tcell.NewEventMouse(50, 50, tcell.Button1, tcell.ModNone)) {
		t.Error("click outside bounds should not be handled")
	}
}

// ── Render: compact ───────────────────────────────────────────────────────────

func sliderTestTheme(t *testing.T) *Theme {
	t.Helper()
	theme := NewTheme()
	theme.SetStrings(map[string]string{
		"slider.compact.track":    "━",
		"slider.compact.thumb":    "┃",
		"slider.box.top-left":     "╭",
		"slider.box.top-right":    "╮",
		"slider.box.bottom-left":  "╰",
		"slider.box.bottom-right": "╯",
		"slider.box.horizontal":   "─",
		"slider.box.thumb-top":    "╥",
		"slider.box.thumb-bottom": "╨",
	})
	return theme
}

func TestSlider_Render_Compact_ThumbAtMin(t *testing.T) {
	s := NewSlider("s", "")
	s.Apply(sliderTestTheme(t))
	cs := NewTestScreen()
	rd := NewRenderer(cs, sliderTestTheme(t))
	s.SetBounds(0, 0, 11, 1)
	s.Render(rd)
	if got := cs.Get(0, 0); got != "┃" {
		t.Errorf("col 0 = %q; want thumb ┃ at min", got)
	}
	if got := cs.Get(10, 0); got != "━" {
		t.Errorf("col 10 = %q; want track ━", got)
	}
}

func TestSlider_Render_Compact_ThumbAtMax(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(100)
	s.Apply(sliderTestTheme(t))
	cs := NewTestScreen()
	rd := NewRenderer(cs, sliderTestTheme(t))
	s.SetBounds(0, 0, 11, 1)
	s.Render(rd)
	if got := cs.Get(10, 0); got != "┃" {
		t.Errorf("col 10 = %q; want thumb ┃ at max", got)
	}
	if got := cs.Get(0, 0); got != "━" {
		t.Errorf("col 0 = %q; want track ━", got)
	}
}

func TestSlider_Render_Compact_ThumbAtMiddle(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(50)
	s.Apply(sliderTestTheme(t))
	cs := NewTestScreen()
	rd := NewRenderer(cs, sliderTestTheme(t))
	s.SetBounds(0, 0, 11, 1)
	s.Render(rd)
	if got := cs.Get(5, 0); got != "┃" {
		t.Errorf("col 5 = %q; want thumb ┃ at 50%%", got)
	}
}

// ── Render: box ───────────────────────────────────────────────────────────────

func TestSlider_Render_Box_CornersAndThumb(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(50)
	s.Apply(sliderTestTheme(t))
	cs := NewTestScreen()
	rd := NewRenderer(cs, sliderTestTheme(t))
	s.SetBounds(0, 0, 11, 2)
	s.Render(rd)
	// Corners
	if got := cs.Get(0, 0); got != "╭" {
		t.Errorf("top-left = %q; want ╭", got)
	}
	if got := cs.Get(10, 0); got != "╮" {
		t.Errorf("top-right = %q; want ╮", got)
	}
	if got := cs.Get(0, 1); got != "╰" {
		t.Errorf("bottom-left = %q; want ╰", got)
	}
	if got := cs.Get(10, 1); got != "╯" {
		t.Errorf("bottom-right = %q; want ╯", got)
	}
	// Inner width = 9 (cols 1..9). Thumb at 50% → col 4 within inner → x=5.
	if got := cs.Get(5, 0); got != "╥" {
		t.Errorf("thumb-top at x=5 = %q; want ╥", got)
	}
	if got := cs.Get(5, 1); got != "╨" {
		t.Errorf("thumb-bottom at x=5 = %q; want ╨", got)
	}
}

func TestSlider_Render_Box_CenteredVertically(t *testing.T) {
	s := NewSlider("s", "")
	s.Apply(sliderTestTheme(t))
	cs := NewTestScreen()
	rd := NewRenderer(cs, sliderTestTheme(t))
	// Height 4 → box should occupy rows 1 and 2 (offset (4-2)/2 = 1).
	s.SetBounds(0, 0, 11, 4)
	s.Render(rd)
	if got := cs.Get(0, 1); got != "╭" {
		t.Errorf("top-left at row 1 = %q; want ╭ (centered in h=4)", got)
	}
	if got := cs.Get(0, 2); got != "╰" {
		t.Errorf("bottom-left at row 2 = %q; want ╰ (centered in h=4)", got)
	}
	// Rows 0 and 3 are inside the content area, so the component paints
	// the background there (spaces) — what matters is they hold no border
	// glyphs.
	for _, r := range []int{0, 3} {
		if got := cs.Get(0, r); got == "╭" || got == "╰" {
			t.Errorf("row %d contains box glyph %q; should be padding", r, got)
		}
	}
}

// ── Summary ───────────────────────────────────────────────────────────────────

func TestSlider_Summary(t *testing.T) {
	s := NewSlider("s", "")
	s.Set(42)
	if s.Summary() == "" {
		t.Error("Summary() should not be empty")
	}
}

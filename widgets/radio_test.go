package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestRadio_Defaults(t *testing.T) {
	r := NewRadio("r", "", "a", "Alpha", "b", "Beta")
	if !r.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
	if r.Value() != "a" {
		t.Errorf("Value() = %q; want %q (first option)", r.Value(), "a")
	}
}

func TestRadio_Empty_Safe(t *testing.T) {
	r := NewRadio("r", "")
	if r.Value() != "" {
		t.Errorf("Value() on empty radio = %q; want \"\"", r.Value())
	}
	if r.Text() != "" {
		t.Errorf("Text() on empty radio = %q; want \"\"", r.Text())
	}
	if r.Summary() != "" {
		t.Errorf("Summary() on empty radio = %q; want \"\"", r.Summary())
	}
}

// ── Select / Value / Text ─────────────────────────────────────────────────────

func TestRadio_SelectByValue(t *testing.T) {
	r := NewRadio("r", "", "a", "Alpha", "b", "Beta", "c", "Gamma")
	r.Select("b")
	if r.Value() != "b" {
		t.Errorf("Value() = %q; want %q", r.Value(), "b")
	}
	if r.Text() != "Beta" {
		t.Errorf("Text() = %q; want %q", r.Text(), "Beta")
	}
}

func TestRadio_Select_UnknownValue_ResetsToFirst(t *testing.T) {
	r := NewRadio("r", "", "a", "Alpha", "b", "Beta")
	r.Select("b")
	r.Select("nonexistent")
	if r.Value() != "a" {
		t.Errorf("Value() after Select(unknown) = %q; want %q", r.Value(), "a")
	}
}

func TestRadio_Select_Silent(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B")
	fired := false
	r.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	r.Select("b")
	if fired {
		t.Error("Select() should not dispatch EvtChange — programmatic updates are silent")
	}
}

// ── Hint ──────────────────────────────────────────────────────────────────────

func TestRadio_Hint_HeightMatchesOptionCount(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B", "c", "C")
	_, h := r.Hint()
	if h != 3 {
		t.Errorf("Hint height = %d; want 3", h)
	}
}

func TestRadio_Hint_WidthFromLongestOption(t *testing.T) {
	r := NewRadio("r", "", "a", "Hi", "b", "Hello World")
	w, _ := r.Hint()
	// Longest label = 11; prefix budget = 4 → 15
	if w != 15 {
		t.Errorf("Hint width = %d; want 15 (longest text + 4)", w)
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func TestRadio_Keyboard_Down_MovesSelection(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B", "c", "C")
	r.handleKey(BuildKey(tcell.KeyDown))
	if r.Value() != "b" {
		t.Errorf("Value() after Down = %q; want %q", r.Value(), "b")
	}
}

func TestRadio_Keyboard_Down_AtLastClamps(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B")
	r.Select("b")
	r.handleKey(BuildKey(tcell.KeyDown))
	if r.Value() != "b" {
		t.Errorf("Down at last option should not wrap; got %q", r.Value())
	}
}

func TestRadio_Keyboard_Up_MovesSelection(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B")
	r.Select("b")
	r.handleKey(BuildKey(tcell.KeyUp))
	if r.Value() != "a" {
		t.Errorf("Value() after Up = %q; want %q", r.Value(), "a")
	}
}

func TestRadio_Keyboard_Home_End(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B", "c", "C")
	r.handleKey(BuildKey(tcell.KeyEnd))
	if r.Value() != "c" {
		t.Errorf("End should select last option; got %q", r.Value())
	}
	r.handleKey(BuildKey(tcell.KeyHome))
	if r.Value() != "a" {
		t.Errorf("Home should select first option; got %q", r.Value())
	}
}

func TestRadio_Keyboard_DispatchesEvtChange(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B")
	var got string
	r.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		got = data[0].(string)
		return true
	})
	r.handleKey(BuildKey(tcell.KeyDown))
	if got != "b" {
		t.Errorf("EvtChange data = %q; want %q", got, "b")
	}
}

func TestRadio_Keyboard_Readonly_Ignored(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B")
	r.SetFlag(FlagReadonly, true)
	handled := r.handleKey(BuildKey(tcell.KeyDown))
	if handled {
		t.Error("readonly radio should not handle navigation keys")
	}
	if r.Value() != "a" {
		t.Error("selection should not change while readonly")
	}
}

// ── Mouse ─────────────────────────────────────────────────────────────────────

func TestRadio_Mouse_Click_SelectsRow(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B", "c", "C")
	r.SetBounds(0, 0, 10, 3)
	r.handleMouse(tcell.NewEventMouse(2, 1, tcell.Button1, tcell.ModNone))
	if r.Value() != "b" {
		t.Errorf("Click on row 1 should select 2nd option; got %q", r.Value())
	}
}

func TestRadio_Mouse_OutOfBounds_Ignored(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B")
	r.SetBounds(0, 0, 10, 2)
	handled := r.handleMouse(tcell.NewEventMouse(50, 50, tcell.Button1, tcell.ModNone))
	if handled {
		t.Error("click outside bounds should not be handled")
	}
}

func TestRadio_Mouse_BelowLastRow_Ignored(t *testing.T) {
	r := NewRadio("r", "", "a", "A", "b", "B")
	r.SetBounds(0, 0, 10, 5)
	// Row 3 has no option (only 2 options exist)
	handled := r.handleMouse(tcell.NewEventMouse(2, 3, tcell.Button1, tcell.ModNone))
	if handled {
		t.Error("click on empty row should not be handled")
	}
}

// ── Glyph width handling ──────────────────────────────────────────────────────

func TestRadio_Render_WithTripleCharGlyph(t *testing.T) {
	r := NewRadio("r", "", "a", "Alpha", "b", "Beta")
	theme := NewTheme()
	theme.SetStrings(map[string]string{
		"radio.on":  "(•)",
		"radio.off": "( )",
	})
	r.Apply(theme)
	cs := NewTestScreen()
	rd := NewRenderer(cs, theme)
	r.SetBounds(0, 0, 20, 2)
	r.Render(rd)
	// First row should start with "(•)" because option 0 is selected.
	got := cs.Get(0, 0) + cs.Get(1, 0) + cs.Get(2, 0)
	if got != "(•)" {
		t.Errorf("row 0 glyph = %q; want %q", got, "(•)")
	}
	// Second row should start with "( )".
	got = cs.Get(0, 1) + cs.Get(1, 1) + cs.Get(2, 1)
	if got != "( )" {
		t.Errorf("row 1 glyph = %q; want %q", got, "( )")
	}
}

func TestRadio_Render_WithSingleCharGlyph(t *testing.T) {
	r := NewRadio("r", "", "a", "Alpha", "b", "Beta")
	theme := NewTheme()
	theme.SetStrings(map[string]string{
		"radio.on":  "◉",
		"radio.off": "○",
	})
	r.Apply(theme)
	cs := NewTestScreen()
	rd := NewRenderer(cs, theme)
	r.SetBounds(0, 0, 20, 2)
	r.Render(rd)
	if got := cs.Get(0, 0); got != "◉" {
		t.Errorf("row 0 glyph = %q; want %q", got, "◉")
	}
	if got := cs.Get(0, 1); got != "○" {
		t.Errorf("row 1 glyph = %q; want %q", got, "○")
	}
}

// ── Summary ───────────────────────────────────────────────────────────────────

func TestRadio_Summary(t *testing.T) {
	r := NewRadio("r", "", "key", "Label")
	if r.Summary() == "" {
		t.Error("Summary() should not be empty")
	}
}

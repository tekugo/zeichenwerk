package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestHRule_ID(t *testing.T) {
	r := NewHRule("", "thin")
	if r.ID() != "hrule" {
		t.Errorf("ID() = %q; want %q", r.ID(), "hrule")
	}
}

func TestVRule_ID(t *testing.T) {
	r := NewVRule("", "thin")
	if r.ID() != "vrule" {
		t.Errorf("ID() = %q; want %q", r.ID(), "vrule")
	}
}

// ── Hint ─────────────────────────────────────────────────────────────────────

func TestHRule_Hint_FixedHeight(t *testing.T) {
	r := NewHRule("", "thin")
	_, h := r.Hint()
	if h != 1 {
		t.Errorf("HRule Hint height = %d; want 1", h)
	}
}

func TestHRule_Hint_WidthIsZero(t *testing.T) {
	r := NewHRule("", "thin")
	w, _ := r.Hint()
	if w != 0 {
		t.Errorf("HRule Hint width = %d; want 0 (fills available space)", w)
	}
}

func TestVRule_Hint_FixedWidth(t *testing.T) {
	r := NewVRule("", "thin")
	w, _ := r.Hint()
	if w != 1 {
		t.Errorf("VRule Hint width = %d; want 1", w)
	}
}

func TestVRule_Hint_HeightIsZero(t *testing.T) {
	r := NewVRule("", "thin")
	_, h := r.Hint()
	if h != 0 {
		t.Errorf("VRule Hint height = %d; want 0 (fills available space)", h)
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestHRule_Render_DrawsLine(t *testing.T) {
	theme := NewTheme()
	theme.SetBorders(map[string]*Border{"default": {InnerH: "-", InnerV: "|"}})
	rule := NewHRule("", "thin")
	cs := NewTestScreen()
	renderer := NewRenderer(cs, theme)
	rule.SetBounds(0, 0, 5, 1)
	rule.Render(renderer)

	// r.Line is called with w-2 = 3 repetitions starting at (x, y) = (0, 0).
	// At least one cell should contain a line character.
	found := false
	for x := 0; x < 5; x++ {
		if cs.Get(x, 0) != "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("HRule render should draw at least one cell")
	}
}

func TestVRule_Render_DrawsLine(t *testing.T) {
	theme := NewTheme()
	theme.SetBorders(map[string]*Border{"default": {InnerH: "-", InnerV: "|"}})
	rule := NewVRule("", "thin")
	cs := NewTestScreen()
	renderer := NewRenderer(cs, theme)
	rule.SetBounds(0, 0, 1, 5)
	rule.Render(renderer)

	found := false
	for y := 0; y < 5; y++ {
		if cs.Get(0, y) != "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("VRule render should draw at least one cell")
	}
}

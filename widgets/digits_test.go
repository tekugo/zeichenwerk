package widgets

import (
	"testing"
	"unicode/utf8"

	. "github.com/tekugo/zeichenwerk/core"
)

// ── Constructor / Hint ────────────────────────────────────────────────────────

func TestDigits_Hint_HeightIsThree(t *testing.T) {
	d := NewDigits("d", "", "0")
	_, h := d.Hint()
	if h != 3 {
		t.Errorf("Hint height = %d; want 3", h)
	}
}

func TestDigits_Hint_WidthFromPattern(t *testing.T) {
	d := NewDigits("d", "", "0")
	w, _ := d.Hint()
	// '0' pattern[0] = "╭──╮" → 4 runes
	expected := utf8.RuneCountInString("╭──╮")
	if w != expected {
		t.Errorf("Hint width = %d for \"0\"; want %d (rune width of pattern[0])", w, expected)
	}
}

func TestDigits_Hint_MultiCharWidth(t *testing.T) {
	d := NewDigits("d", "", "00")
	w0, _ := NewDigits("d", "", "0").Hint()
	w2, _ := d.Hint()
	if w2 != w0*2 {
		t.Errorf("Hint width = %d for \"00\"; want %d (2× single char)", w2, w0*2)
	}
}

func TestDigits_Hint_EmptyString(t *testing.T) {
	d := NewDigits("d", "", "")
	w, h := d.Hint()
	if w != 0 || h != 3 {
		t.Errorf("Hint() = (%d,%d) for empty string; want (0,3)", w, h)
	}
}

func TestDigits_Hint_AllDigits(t *testing.T) {
	for _, ch := range "0123456789" {
		d := NewDigits("d", "", string(ch))
		w, h := d.Hint()
		if w <= 0 {
			t.Errorf("Hint width = %d for %q; want > 0", w, string(ch))
		}
		if h != 3 {
			t.Errorf("Hint height = %d for %q; want 3", h, string(ch))
		}
	}
}

// ── Set ───────────────────────────────────────────────────────────────────────

func TestDigits_Set_UpdatesText(t *testing.T) {
	d := NewDigits("d", "", "0")
	d.Set("123")
	if d.Text != "123" {
		t.Errorf("Text = %q after Set; want %q", d.Text, "123")
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestDigits_Render_DrawsPattern(t *testing.T) {
	d := NewDigits("d", "", "0")
	w, _ := d.Hint()
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	d.SetBounds(0, 0, w+2, 5) // enough room
	d.Render(r)

	// Row 0 of '0' pattern is "╭──╮"; expect '╭' at first cell
	if cs.Get(0, 0) != "╭" {
		t.Errorf("render row 0 col 0 = %q; want %q", cs.Get(0, 0), "╭")
	}
}

func TestDigits_Render_ThreeRowsUsed(t *testing.T) {
	d := NewDigits("d", "", "8") // '8' has non-empty content on all 3 rows
	w, _ := d.Hint()
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	d.SetBounds(0, 0, w+2, 5)
	d.Render(r)

	// Rows 0, 1, 2 should all have content
	for row := 0; row < 3; row++ {
		if cs.Get(0, row) == "" {
			t.Errorf("row %d is empty; expect digit pattern content", row)
		}
	}
}

func TestDigits_Render_InsufficientHeight_NoOp(t *testing.T) {
	d := NewDigits("d", "", "0")
	w, _ := d.Hint()
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	d.SetBounds(0, 0, w+2, 2) // only 2 rows — Render returns early
	d.Render(r)
	// Should not panic; nothing verifiable about content
}

func TestDigits_Render_UnknownCharSkipped(t *testing.T) {
	d := NewDigits("d", "", "0?0") // '?' is not in the digit map
	w0, _ := NewDigits("d", "", "0").Hint()
	// '?' not in digits map, so hint width is 2×w0 (only two known chars)
	wExpected := 2 * w0
	w, _ := d.Hint()
	if w != wExpected {
		t.Errorf("Hint width = %d for \"0?0\"; want %d (unknown char skipped)", w, wExpected)
	}
}

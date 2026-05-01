package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestStatic_Defaults(t *testing.T) {
	s := NewStatic("s", "", "hello")
	if s.Text != "hello" {
		t.Errorf("Text = %q; want %q", s.Text, "hello")
	}
	if s.Alignment != "left" {
		t.Errorf("Alignment = %q; want %q", s.Alignment, "left")
	}
}

// ── Hint ─────────────────────────────────────────────────────────────────────

func TestStatic_Hint_RuneCount(t *testing.T) {
	s := NewStatic("s", "", "Hello")
	w, h := s.Hint()
	if w != 5 {
		t.Errorf("Hint width = %d; want 5", w)
	}
	if h != 1 {
		t.Errorf("Hint height = %d; want 1", h)
	}
}

func TestStatic_Hint_MultiByte(t *testing.T) {
	s := NewStatic("s", "", "über") // 4 runes
	w, _ := s.Hint()
	if w != 4 {
		t.Errorf("Hint width = %d; want 4 (rune count)", w)
	}
}

func TestStatic_Hint_Empty(t *testing.T) {
	s := NewStatic("s", "", "")
	w, h := s.Hint()
	if w != 0 || h != 1 {
		t.Errorf("Hint() = (%d,%d); want (0,1)", w, h)
	}
}

// ── Set ───────────────────────────────────────────────────────────────────────

func TestStatic_Set_String(t *testing.T) {
	s := NewStatic("s", "", "old")
	s.Set("new")
	if s.Text != "new" {
		t.Errorf("Text = %q after Set; want %q", s.Text, "new")
	}
}

func TestStatic_Set_NonString_Formatted(t *testing.T) {
	s := NewStatic("s", "", "")
	s.Set(42)
	if s.Text != "42" {
		t.Errorf("Text = %q after Set(42); want %q", s.Text, "42")
	}
}

// ── SetAlignment ─────────────────────────────────────────────────────────────

func TestStatic_SetAlignment(t *testing.T) {
	s := NewStatic("s", "", "text")
	s.SetAlignment("right")
	if s.Alignment != "right" {
		t.Errorf("Alignment = %q; want %q", s.Alignment, "right")
	}
}

// ── Summary ───────────────────────────────────────────────────────────────────

func TestStatic_Summary_Short(t *testing.T) {
	s := NewStatic("s", "", "hello")
	if s.Summary() != "hello" {
		t.Errorf("Summary() = %q; want %q", s.Summary(), "hello")
	}
}

func TestStatic_Summary_TruncatesAt60(t *testing.T) {
	long := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // 62 chars
	s := NewStatic("s", "", long)
	sum := s.Summary()
	runes := []rune(sum)
	if len(runes) > 61 { // 60 + "…"
		t.Errorf("Summary() len = %d; want ≤ 61 (60 chars + ellipsis)", len(runes))
	}
	if runes[len(runes)-1] != '…' {
		t.Errorf("Summary() last char = %q; want ellipsis", string(runes[len(runes)-1]))
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestStatic_Render_ShowsText(t *testing.T) {
	s := NewStatic("s", "", "Hi")
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	s.SetBounds(0, 0, 10, 1)
	s.Render(r)
	got := cs.Get(0, 0) + cs.Get(1, 0)
	if got != "Hi" {
		t.Errorf("rendered = %q; want %q", got, "Hi")
	}
}

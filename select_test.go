package zeichenwerk

import "testing"

// ── Constructor ───────────────────────────────────────────────────────────────

func TestSelect_Defaults(t *testing.T) {
	s := NewSelect("s", "", "v1", "Option 1", "v2", "Option 2")
	if !s.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
	// First option should be selected by default
	if s.Value() != "v1" {
		t.Errorf("Value() = %q; want %q (first option)", s.Value(), "v1")
	}
}

func TestSelect_Text_MatchesFirstOption(t *testing.T) {
	s := NewSelect("s", "", "v1", "Option 1", "v2", "Option 2")
	if s.Text() != "Option 1" {
		t.Errorf("Text() = %q; want %q", s.Text(), "Option 1")
	}
}

// ── Select / Value / Text ─────────────────────────────────────────────────────

func TestSelect_SelectByValue(t *testing.T) {
	s := NewSelect("s", "", "v1", "Alpha", "v2", "Beta", "v3", "Gamma")
	s.Select("v2")
	if s.Value() != "v2" {
		t.Errorf("Value() = %q after Select(\"v2\"); want %q", s.Value(), "v2")
	}
	if s.Text() != "Beta" {
		t.Errorf("Text() = %q after Select(\"v2\"); want %q", s.Text(), "Beta")
	}
}

func TestSelect_Select_UnknownValue_ResetsToFirst(t *testing.T) {
	s := NewSelect("s", "", "v1", "Alpha", "v2", "Beta")
	s.Select("v2")
	s.Select("nonexistent")
	// Implementation resets to index 0 for unknown values
	if s.Value() != "v1" {
		t.Errorf("Value() = %q after Select(unknown); want %q", s.Value(), "v1")
	}
}

func TestSelect_Select_LastOption(t *testing.T) {
	s := NewSelect("s", "", "a", "Apple", "b", "Banana", "c", "Cherry")
	s.Select("c")
	if s.Value() != "c" {
		t.Errorf("Value() = %q; want %q", s.Value(), "c")
	}
	if s.Text() != "Cherry" {
		t.Errorf("Text() = %q; want %q", s.Text(), "Cherry")
	}
}

// ── Hint ─────────────────────────────────────────────────────────────────────

func TestSelect_Hint_WidthFromLongestOption(t *testing.T) {
	s := NewSelect("s", "", "v1", "Hi", "v2", "Hello World")
	w, h := s.Hint()
	// Longest text is "Hello World" = 11 runes; + 2 for dropdown marker
	if w != 13 {
		t.Errorf("Hint width = %d; want 13 (longest text + 2)", w)
	}
	if h != 1 {
		t.Errorf("Hint height = %d; want 1", h)
	}
}

func TestSelect_Hint_SingleOption(t *testing.T) {
	s := NewSelect("s", "", "v", "OK") // "OK" = 2 runes → 2 + 2 = 4
	w, _ := s.Hint()
	if w != 4 {
		t.Errorf("Hint width = %d; want 4", w)
	}
}

// ── Summary ───────────────────────────────────────────────────────────────────

func TestSelect_Summary(t *testing.T) {
	s := NewSelect("s", "", "key", "Label")
	sum := s.Summary()
	if sum == "" {
		t.Error("Summary() should not be empty")
	}
}

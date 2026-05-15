package core

import "testing"

func TestParseInsetsCSV(t *testing.T) {
	cases := []struct {
		in   string
		want [4]int
		ok   bool
	}{
		// Empty maps to NoInsets.
		{"", [4]int{}, true},
		{"   ", [4]int{}, true},

		// 1 value: applies to all four sides.
		{"5", [4]int{5, 5, 5, 5}, true},
		{"  5  ", [4]int{5, 5, 5, 5}, true},

		// 2 values: top/bottom, left/right.
		{"1 2", [4]int{1, 2, 1, 2}, true},
		{"1, 2", [4]int{1, 2, 1, 2}, true},
		{"1,2", [4]int{1, 2, 1, 2}, true},

		// 3 values: top, left/right, bottom.
		{"1 2 3", [4]int{1, 2, 3, 2}, true},
		{"1, 2, 3", [4]int{1, 2, 3, 2}, true},

		// 4 values: top, right, bottom, left.
		{"1 2 3 4", [4]int{1, 2, 3, 4}, true},
		{"1, 2, 3, 4", [4]int{1, 2, 3, 4}, true},

		// Errors: non-numeric, too many, mid-edit garbage.
		{"a", [4]int{}, false},
		{"1 b", [4]int{}, false},
		{"1 2 3 4 5", [4]int{}, false},
		{"1,", [4]int{1, 1, 1, 1}, true}, // trailing comma is whitespace
	}
	for _, tc := range cases {
		got, ok := parseInsets(tc.in)
		if ok != tc.ok {
			t.Errorf("parseInsetsCSV(%q): ok = %v, want %v", tc.in, ok, tc.ok)
			continue
		}
		if got.Array() != tc.want {
			t.Errorf("parseInsetsCSV(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestFormatInsets(t *testing.T) {
	cases := []struct {
		in   *Insets
		want string
	}{
		// Nil and all-zero collapse to "" (NoInsets).
		{nil, ""},
		{&Insets{}, "0"},

		// Symmetric collapses.
		{&Insets{Top: 5, Right: 5, Bottom: 5, Left: 5}, "5"},
		{&Insets{Top: 1, Right: 2, Bottom: 1, Left: 2}, "1 2"},
		{&Insets{Top: 1, Right: 2, Bottom: 3, Left: 2}, "1 2 3"},
		{&Insets{Top: 1, Right: 2, Bottom: 3, Left: 4}, "1 2 3 4"},
	}
	for _, tc := range cases {
		got := tc.in.String()
		if got != tc.want {
			t.Errorf("formatInsets(%+v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestStyleForm_RoundTrip(t *testing.T) {
	// Build a style with overrides; Load → Store → Load should
	// preserve everything across the cycle.
	src := NewStyle("test").
		WithForeground("$fg1").
		WithBackground("$bg2").
		WithBorder("round").
		WithPadding(1, 2, 3, 4).
		WithMargin(5, 10).
		WithFont("bold")

	var f StyleForm
	f.Load(src)

	// Render a fresh style from the form and re-load to compare.
	dst := f.Store(NewStyle(""))
	var f2 StyleForm
	f2.Load(dst)

	if f.Foreground != f2.Foreground {
		t.Errorf("Foreground: %q != %q", f.Foreground, f2.Foreground)
	}
	if f.Background != f2.Background {
		t.Errorf("Background: %q != %q", f.Background, f2.Background)
	}
	if f.Border != f2.Border {
		t.Errorf("Border: %q != %q", f.Border, f2.Border)
	}
	if f.Padding != f2.Padding {
		t.Errorf("Padding: %q != %q", f.Padding, f2.Padding)
	}
	if f.Margin != f2.Margin {
		t.Errorf("Margin: %q != %q", f.Margin, f2.Margin)
	}
	if f.Font != f2.Font {
		t.Errorf("Font: %q != %q", f.Font, f2.Font)
	}
}

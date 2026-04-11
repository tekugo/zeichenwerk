package zeichenwerk

import "testing"

// ── parseHex ─────────────────────────────────────────────────────────────────

func TestParseHex_Valid(t *testing.T) {
	cases := []struct {
		in   string
		r, g, b uint8
	}{
		{"#000000", 0, 0, 0},
		{"#ffffff", 255, 255, 255},
		{"#ff0000", 255, 0, 0},
		{"#00ff00", 0, 255, 0},
		{"#0000ff", 0, 0, 255},
		{"#1a2b3c", 0x1a, 0x2b, 0x3c},
	}
	for _, tc := range cases {
		r, g, b := parseHex(tc.in)
		if r != tc.r || g != tc.g || b != tc.b {
			t.Errorf("parseHex(%q) = (%d,%d,%d); want (%d,%d,%d)",
				tc.in, r, g, b, tc.r, tc.g, tc.b)
		}
	}
}

func TestParseHex_InvalidFormat_ReturnsZero(t *testing.T) {
	cases := []string{"", "red", "#fff", "#gggggg", "1a2b3c"}
	for _, in := range cases {
		r, g, b := parseHex(in)
		if r != 0 || g != 0 || b != 0 {
			t.Errorf("parseHex(%q) = (%d,%d,%d); want (0,0,0) for invalid input", in, r, g, b)
		}
	}
}

// ── fmtHex ───────────────────────────────────────────────────────────────────

func TestFmtHex(t *testing.T) {
	cases := []struct {
		r, g, b uint8
		want    string
	}{
		{0, 0, 0, "#000000"},
		{255, 255, 255, "#ffffff"},
		{255, 0, 0, "#ff0000"},
		{0x1a, 0x2b, 0x3c, "#1a2b3c"},
	}
	for _, tc := range cases {
		got := fmtHex(tc.r, tc.g, tc.b)
		if got != tc.want {
			t.Errorf("fmtHex(%d,%d,%d) = %q; want %q", tc.r, tc.g, tc.b, got, tc.want)
		}
	}
}

// ── dimColor ─────────────────────────────────────────────────────────────────

func TestDimColor_Half(t *testing.T) {
	got := dimColor("#ffffff", 0.5)
	// Each channel: round(255 * 0.5) = 128 = 0x80
	if got != "#808080" {
		t.Errorf("dimColor(\"#ffffff\", 0.5) = %q; want %q", got, "#808080")
	}
}

func TestDimColor_ZeroFactor(t *testing.T) {
	got := dimColor("#aabbcc", 0)
	if got != "#000000" {
		t.Errorf("dimColor(\"#aabbcc\", 0) = %q; want %q", got, "#000000")
	}
}

func TestDimColor_FullFactor(t *testing.T) {
	got := dimColor("#1a2b3c", 1.0)
	if got != "#1a2b3c" {
		t.Errorf("dimColor(\"#1a2b3c\", 1.0) = %q; want unchanged %q", got, "#1a2b3c")
	}
}

func TestDimColor_NonHex_Passthrough(t *testing.T) {
	got := dimColor("red", 0.5)
	if got != "red" {
		t.Errorf("dimColor(\"red\", 0.5) = %q; want %q (non-hex passthrough)", got, "red")
	}
}

func TestDimColor_EmptyString_Passthrough(t *testing.T) {
	got := dimColor("", 0.5)
	if got != "" {
		t.Errorf("dimColor(\"\", 0.5) = %q; want empty", got)
	}
}

// ── lerpColor ────────────────────────────────────────────────────────────────

func TestLerpColor_AtZero_ReturnsA(t *testing.T) {
	got := lerpColor("#ff0000", "#0000ff", 0)
	if got != "#ff0000" {
		t.Errorf("lerpColor(t=0) = %q; want %q", got, "#ff0000")
	}
}

func TestLerpColor_AtOne_ReturnsB(t *testing.T) {
	got := lerpColor("#ff0000", "#0000ff", 1)
	if got != "#0000ff" {
		t.Errorf("lerpColor(t=1) = %q; want %q", got, "#0000ff")
	}
}

func TestLerpColor_AtHalf(t *testing.T) {
	// #000000 → #ffffff at t=0.5: each channel ≈ 128 = 0x80
	got := lerpColor("#000000", "#ffffff", 0.5)
	if got != "#808080" {
		t.Errorf("lerpColor(t=0.5) = %q; want %q", got, "#808080")
	}
}

func TestLerpColor_BelowZero_ClampsToA(t *testing.T) {
	got := lerpColor("#ff0000", "#0000ff", -1)
	if got != "#ff0000" {
		t.Errorf("lerpColor(t=-1) = %q; want %q (clamped to a)", got, "#ff0000")
	}
}

func TestLerpColor_AboveOne_ClampsToB(t *testing.T) {
	got := lerpColor("#ff0000", "#0000ff", 2)
	if got != "#0000ff" {
		t.Errorf("lerpColor(t=2) = %q; want %q (clamped to b)", got, "#0000ff")
	}
}

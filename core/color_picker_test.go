package core

import (
	"math"
	"math/rand"
	"testing"
)

// ── ParseHexColor ───────────────────────────────────────────────────────────

func TestParseHexColor_SixDigits(t *testing.T) {
	r, g, b, ok := ParseHexColor("#1a2b3c")
	if !ok {
		t.Fatalf("ParseHexColor(#1a2b3c) ok=false; want true")
	}
	if r != 0x1a || g != 0x2b || b != 0x3c {
		t.Errorf("ParseHexColor(#1a2b3c) = (%d,%d,%d); want (26,43,60)", r, g, b)
	}
}

func TestParseHexColor_ThreeDigits_DoublesEachDigit(t *testing.T) {
	r3, g3, b3, ok3 := ParseHexColor("#abc")
	r6, g6, b6, ok6 := ParseHexColor("#aabbcc")
	if !ok3 || !ok6 {
		t.Fatalf("expected both forms to parse")
	}
	if r3 != r6 || g3 != g6 || b3 != b6 {
		t.Errorf("#abc and #aabbcc should be equal; got (%d,%d,%d) vs (%d,%d,%d)", r3, g3, b3, r6, g6, b6)
	}
	if r3 != 0xaa || g3 != 0xbb || b3 != 0xcc {
		t.Errorf("#abc → (%d,%d,%d); want (170,187,204)", r3, g3, b3)
	}
}

func TestParseHexColor_Invalid(t *testing.T) {
	cases := []string{"", "red", "1a2b3c", "#", "#1", "#12", "#12345", "#1234567", "#gggggg", "#ggg"}
	for _, in := range cases {
		_, _, _, ok := ParseHexColor(in)
		if ok {
			t.Errorf("ParseHexColor(%q) ok=true; want false", in)
		}
	}
}

// ── RGB ↔ HSL round trip ───────────────────────────────────────────────────

func TestRGBToHSL_KnownValues(t *testing.T) {
	cases := []struct {
		r, g, b uint8
		h, s, l float64
	}{
		{0, 0, 0, 0, 0, 0},
		{255, 255, 255, 0, 0, 100},
		{255, 0, 0, 0, 100, 50},
		{0, 255, 0, 120, 100, 50},
		{0, 0, 255, 240, 100, 50},
		{128, 128, 128, 0, 0, 50.196},
	}
	for _, tc := range cases {
		h, s, l := RGBToHSL(tc.r, tc.g, tc.b)
		if math.Abs(h-tc.h) > 0.5 || math.Abs(s-tc.s) > 0.5 || math.Abs(l-tc.l) > 0.5 {
			t.Errorf("RGBToHSL(%d,%d,%d) = (%.2f,%.2f,%.2f); want (%.2f,%.2f,%.2f)",
				tc.r, tc.g, tc.b, h, s, l, tc.h, tc.s, tc.l)
		}
	}
}

func TestHSLToRGB_KnownValues(t *testing.T) {
	cases := []struct {
		h, s, l float64
		r, g, b uint8
	}{
		{0, 0, 0, 0, 0, 0},
		{0, 0, 100, 255, 255, 255},
		{0, 100, 50, 255, 0, 0},
		{120, 100, 50, 0, 255, 0},
		{240, 100, 50, 0, 0, 255},
	}
	for _, tc := range cases {
		r, g, b := HSLToRGB(tc.h, tc.s, tc.l)
		if r != tc.r || g != tc.g || b != tc.b {
			t.Errorf("HSLToRGB(%.0f,%.0f,%.0f) = (%d,%d,%d); want (%d,%d,%d)",
				tc.h, tc.s, tc.l, r, g, b, tc.r, tc.g, tc.b)
		}
	}
}

func TestRGBHSLRoundTrip_RandomColors(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 200; i++ {
		r := uint8(rng.Intn(256))
		g := uint8(rng.Intn(256))
		b := uint8(rng.Intn(256))

		h, s, l := RGBToHSL(r, g, b)
		r2, g2, b2 := HSLToRGB(h, s, l)

		// Allow a 1-channel rounding error in either direction.
		dr := int(r) - int(r2)
		dg := int(g) - int(g2)
		db := int(b) - int(b2)
		if abs(dr) > 1 || abs(dg) > 1 || abs(db) > 1 {
			t.Errorf("round trip drift for (%d,%d,%d): got (%d,%d,%d) via HSL(%.2f,%.2f,%.2f)",
				r, g, b, r2, g2, b2, h, s, l)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ── Contrast ratio ─────────────────────────────────────────────────────────

func TestContrastRatio_BlackWhite(t *testing.T) {
	got := ContrastRatio(0, 0, 0, 255, 255, 255)
	if math.Abs(got-21) > 0.01 {
		t.Errorf("ContrastRatio(black,white) = %.4f; want ~21.0", got)
	}
}

func TestContrastRatio_Symmetric(t *testing.T) {
	a := ContrastRatio(255, 255, 255, 0, 0, 0)
	b := ContrastRatio(0, 0, 0, 255, 255, 255)
	if math.Abs(a-b) > 1e-9 {
		t.Errorf("ContrastRatio is not symmetric: %.6f vs %.6f", a, b)
	}
}

func TestContrastRatio_SameColor_IsOne(t *testing.T) {
	got := ContrastRatio(0x1a, 0x2b, 0x3c, 0x1a, 0x2b, 0x3c)
	if math.Abs(got-1) > 1e-9 {
		t.Errorf("ContrastRatio(c,c) = %.6f; want 1.0", got)
	}
}

func TestContrastRatio_Bounds(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 100; i++ {
		r1 := uint8(rng.Intn(256))
		g1 := uint8(rng.Intn(256))
		b1 := uint8(rng.Intn(256))
		r2 := uint8(rng.Intn(256))
		g2 := uint8(rng.Intn(256))
		b2 := uint8(rng.Intn(256))
		ratio := ContrastRatio(r1, g1, b1, r2, g2, b2)
		if ratio < 1 || ratio > 21+1e-6 {
			t.Errorf("contrast ratio out of bounds: %.4f", ratio)
		}
	}
}

package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// DimColor scales each RGB channel of a hex colour by factor, producing a
// darker (factor < 1) or unchanged (factor == 1) variant. Input must be in
// the lowercase "#RRGGBB" form; anything that does not match that exact
// shape (including 8-digit RGBA, 3-digit shorthand, named colours, or
// non-hex theme variables like "$primary") is returned verbatim so callers
// can pass theme colour strings without defensive branching.
//
// The function does not clamp factor, so values greater than 1 brighten the
// colour subject to uint8 overflow semantics and negative values produce
// garbage. Each channel is rounded to the nearest integer.
func DimColor(hex string, factor float64) string {
	if !strings.HasPrefix(hex, "#") || len(hex) != 7 {
		return hex
	}
	r, g, b := parseHex(hex)
	return fmtHex(
		uint8(math.Round(float64(r)*factor)),
		uint8(math.Round(float64(g)*factor)),
		uint8(math.Round(float64(b)*factor)),
	)
}

// LerpColor performs per-channel linear interpolation between two hex
// colours a and b. The parameter t is clamped to [0, 1]: values at or below
// 0 return a unchanged, values at or above 1 return b unchanged, and values
// in between are mixed in RGB space.
//
// Both endpoints are expected to be "#RRGGBB" strings. Non-hex inputs are
// treated as black by the internal parser — this is harmless at the exact
// endpoints (which short-circuit to the input) but produces unspecified
// results in the interior of the range, so callers should sanitise their
// inputs first.
func LerpColor(a, b string, t float64) string {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	ar, ag, ab := parseHex(a)
	br, bg, bb := parseHex(b)
	lerp := func(x, y uint8) uint8 {
		return uint8(float64(x) + t*(float64(y)-float64(x)) + 0.5)
	}
	return fmtHex(lerp(ar, br), lerp(ag, bg), lerp(ab, bb))
}

// fmtHex formats (r, g, b) components as a lowercase #RRGGBB string.
func fmtHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// FormatHex formats (r, g, b) components as a lowercase #RRGGBB string.
func FormatHex(r, g, b uint8) string {
	return fmtHex(r, g, b)
}

// parseHex converts a #RRGGBB string to its (r, g, b) uint8 components.
// Returns (0, 0, 0) for any string that is not exactly 7 characters starting
// with '#'.
func parseHex(hex string) (uint8, uint8, uint8) {
	var r, g, b uint64
	if len(hex) == 7 {
		r, _ = strconv.ParseUint(hex[1:3], 16, 8)
		g, _ = strconv.ParseUint(hex[3:5], 16, 8)
		b, _ = strconv.ParseUint(hex[5:7], 16, 8)
	}
	return uint8(r), uint8(g), uint8(b)
}

// ParseHexColor parses a colour string in `#RGB` or `#RRGGBB` form into its
// (r, g, b) uint8 components. The 3-digit form expands each digit by
// duplication (`#abc` → `#aabbcc`). The function reports parse errors so
// callers can react (e.g. show an error style on an input field). Whitespace
// is not trimmed; the caller is expected to pass a clean string.
func ParseHexColor(hex string) (r, g, b uint8, ok bool) {
	if len(hex) < 4 || hex[0] != '#' {
		return 0, 0, 0, false
	}
	switch len(hex) {
	case 4:
		rv, err1 := strconv.ParseUint(hex[1:2], 16, 8)
		gv, err2 := strconv.ParseUint(hex[2:3], 16, 8)
		bv, err3 := strconv.ParseUint(hex[3:4], 16, 8)
		if err1 != nil || err2 != nil || err3 != nil {
			return 0, 0, 0, false
		}
		return uint8(rv*16 + rv), uint8(gv*16 + gv), uint8(bv*16 + bv), true
	case 7:
		rv, err1 := strconv.ParseUint(hex[1:3], 16, 8)
		gv, err2 := strconv.ParseUint(hex[3:5], 16, 8)
		bv, err3 := strconv.ParseUint(hex[5:7], 16, 8)
		if err1 != nil || err2 != nil || err3 != nil {
			return 0, 0, 0, false
		}
		return uint8(rv), uint8(gv), uint8(bv), true
	default:
		return 0, 0, 0, false
	}
}

// RGBToHSL converts an RGB colour (each channel in [0, 255]) to HSL with
// H in [0, 360), S and L in [0, 100]. Inputs outside the byte range are not
// possible because uint8 already clamps them.
func RGBToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	maxc := math.Max(rf, math.Max(gf, bf))
	minc := math.Min(rf, math.Min(gf, bf))
	l = (maxc + minc) / 2

	if maxc == minc {
		return 0, 0, l * 100
	}

	d := maxc - minc
	if l < 0.5 {
		s = d / (maxc + minc)
	} else {
		s = d / (2 - maxc - minc)
	}

	switch maxc {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = 2 + (bf-rf)/d
	default:
		h = 4 + (rf-gf)/d
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	if h >= 360 {
		h -= 360
	}
	return h, s * 100, l * 100
}

// HSLToRGB converts an HSL colour (H in [0, 360), S and L in [0, 100]) to
// RGB with each channel rounded to the nearest integer in [0, 255]. Inputs
// outside the canonical ranges are wrapped/clamped: H is reduced modulo 360,
// S and L are clamped to [0, 100].
func HSLToRGB(h, s, l float64) (r, g, b uint8) {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	if s < 0 {
		s = 0
	} else if s > 100 {
		s = 100
	}
	if l < 0 {
		l = 0
	} else if l > 100 {
		l = 100
	}

	sf := s / 100
	lf := l / 100

	if sf == 0 {
		v := uint8(math.Round(lf * 255))
		return v, v, v
	}

	var q float64
	if lf < 0.5 {
		q = lf * (1 + sf)
	} else {
		q = lf + sf - lf*sf
	}
	p := 2*lf - q
	hk := h / 360

	hueToRGB := func(p, q, t float64) float64 {
		if t < 0 {
			t += 1
		}
		if t > 1 {
			t -= 1
		}
		switch {
		case t < 1.0/6.0:
			return p + (q-p)*6*t
		case t < 1.0/2.0:
			return q
		case t < 2.0/3.0:
			return p + (q-p)*(2.0/3.0-t)*6
		default:
			return p
		}
	}

	rf := hueToRGB(p, q, hk+1.0/3.0)
	gf := hueToRGB(p, q, hk)
	bf := hueToRGB(p, q, hk-1.0/3.0)

	return uint8(math.Round(rf * 255)), uint8(math.Round(gf * 255)), uint8(math.Round(bf * 255))
}

// RelativeLuminance returns the WCAG 2.1 relative luminance of an sRGB
// colour. The result is in [0, 1].
func RelativeLuminance(r, g, b uint8) float64 {
	channel := func(c uint8) float64 {
		v := float64(c) / 255
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b)
}

// ContrastRatio returns the WCAG 2.1 contrast ratio between two RGB colours.
// The result is in [1, 21]; values >= 4.5 satisfy AA for normal text and >=
// 7 satisfy AAA. The function is symmetric in its arguments.
func ContrastRatio(fr, fg, fb, br, bg, bb uint8) float64 {
	l1 := RelativeLuminance(fr, fg, fb)
	l2 := RelativeLuminance(br, bg, bb)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

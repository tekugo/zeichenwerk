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

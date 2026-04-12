package zeichenwerk

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// dimColor multiplies each RGB channel of a #RRGGBB colour by factor ∈ [0, 1],
// returning a darker version. Non-hex strings are returned unchanged.
func dimColor(hex string, factor float64) string {
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

// fmtHex formats (r, g, b) components as a lowercase #RRGGBB string.
func fmtHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// lerpColor linearly interpolates between two #RRGGBB colours.
// t=0 returns a, t=1 returns b; values outside [0, 1] are clamped.
// Non-hex strings are returned unchanged at the nearest end of the range.
func lerpColor(a, b string, t float64) string {
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

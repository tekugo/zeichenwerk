package core

import (
	"fmt"
	"math"
	"time"
)

// Humanize formats n with three significant digits and a K/M/B suffix.
//
//	999       →  "999"
//	1_000     →  "1.00K"
//	15_300    →  "15.3K"
//	999_000   →  "999K"
//	1_000_000 →  "1.00M"
func Humanize[T Number](n T) string {
	f := float64(n)
	switch {
	case f >= 1_000_000_000:
		return threeDigits(f/1_000_000_000) + "B"
	case f >= 1_000_000:
		return threeDigits(f/1_000_000) + "M"
	case f >= 1_000:
		return threeDigits(f/1_000) + "K"
	default:
		if f == math.Trunc(f) {
			return fmt.Sprintf("%d", int64(f))
		}
		return threeDigits(f)
	}
}

// FormatDuration formats d as a compact human-readable string.
//
//	30s  →  "30s"
//	75s  →  "1m 15s"
//	90m  →  "1h 30m"
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// threeDigits formats v with exactly three significant digits.
func threeDigits(v float64) string {
	switch {
	case v < 10:
		return fmt.Sprintf("%.2f", v)
	case v < 100:
		return fmt.Sprintf("%.1f", v)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

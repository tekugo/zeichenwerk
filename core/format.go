package core

import (
	"fmt"
	"math"
	"time"
)

// Humanize formats n as a compact string with approximately three
// significant digits and a decimal SI-style suffix (K for thousands, M for
// millions, B for billions). It is intended for quick on-screen display of
// counts and byte-like magnitudes where precision is less important than
// readability.
//
// Values below 1000 are rendered without a suffix: integral values keep
// their exact representation, while non-integral values are formatted with
// three significant digits. Negative numbers fall into the unscaled branch
// because none of the magnitude thresholds match; very large negative
// numbers therefore print in full rather than being scaled.
//
// The type parameter T is constrained to Number, which covers the built-in
// integer and floating-point types.
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

// FormatDuration formats d as a compact human-readable string using the two
// most significant time units. The duration is first rounded to whole
// seconds; then the output is chosen by the largest non-zero unit:
//   - at least one hour   → "Hh Mm"
//   - at least one minute → "Mm Ss"
//   - otherwise           → "Ss"
//
// Sub-second components are discarded by the rounding step, and units
// smaller than the one below the leading unit are omitted entirely (for
// example a three-hour duration is shown as "3h 0m", not "3h 0m 12s").
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

// threeDigits formats v with approximately three significant digits by
// choosing the number of fractional digits based on v's magnitude:
// values under 10 use two decimals, values under 100 use one, and values
// of 100 or more are shown without a fractional part.
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

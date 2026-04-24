package core

import (
	"testing"
	"time"
)

func TestHumanize(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "0"},
		{"small integer", 1, "1"},
		{"small integer just below threshold", 999, "999"},
		{"small non-integer under ten", 1.5, "1.50"},
		{"small non-integer under hundred", 12.5, "12.5"},
		{"small non-integer under thousand", 123.4, "123"},

		{"exactly one thousand", 1_000, "1.00K"},
		{"one point five K", 1_500, "1.50K"},
		{"fifteen point three K", 15_300, "15.3K"},
		{"hundred K", 100_000, "100K"},
		{"just below one million", 999_000, "999K"},

		{"exactly one million", 1_000_000, "1.00M"},
		{"one point five million", 1_500_000, "1.50M"},
		{"fifteen million", 15_000_000, "15.0M"},
		{"hundred million", 100_000_000, "100M"},

		{"exactly one billion", 1_000_000_000, "1.00B"},
		{"two and a half billion", 2_500_000_000, "2.50B"},

		{"negative integer falls through", -500, "-500"},
		{"negative non-integer falls through", -1.5, "-1.50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Humanize(tt.input); got != tt.expected {
				t.Errorf("Humanize(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHumanize_IntegerTypes(t *testing.T) {
	if got := Humanize(int(1_000)); got != "1.00K" {
		t.Errorf("Humanize(int) = %q, want %q", got, "1.00K")
	}
	if got := Humanize(int8(42)); got != "42" {
		t.Errorf("Humanize(int8) = %q, want %q", got, "42")
	}
	if got := Humanize(int32(15_300)); got != "15.3K" {
		t.Errorf("Humanize(int32) = %q, want %q", got, "15.3K")
	}
	if got := Humanize(int64(1_000_000_000)); got != "1.00B" {
		t.Errorf("Humanize(int64) = %q, want %q", got, "1.00B")
	}
	if got := Humanize(float32(1_500)); got != "1.50K" {
		t.Errorf("Humanize(float32) = %q, want %q", got, "1.50K")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"zero", 0, "0s"},
		{"one second", 1 * time.Second, "1s"},
		{"thirty seconds", 30 * time.Second, "30s"},
		{"fifty-nine seconds", 59 * time.Second, "59s"},

		{"exactly one minute", 1 * time.Minute, "1m 0s"},
		{"seventy-five seconds", 75 * time.Second, "1m 15s"},
		{"fifty-nine minutes", 59 * time.Minute, "59m 0s"},

		{"exactly one hour", 1 * time.Hour, "1h 0m"},
		{"ninety minutes", 90 * time.Minute, "1h 30m"},
		{"three hours", 3 * time.Hour, "3h 0m"},
		{"three hours seconds dropped", 3*time.Hour + 12*time.Second, "3h 0m"},

		{"sub-second rounds up", 1500 * time.Millisecond, "2s"},
		{"sub-second rounds down", 499 * time.Millisecond, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDuration(tt.input); got != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestThreeDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "0.00"},
		{"under ten", 1.234, "1.23"},
		{"just under ten", 9.999, "10.00"},
		{"exactly ten", 10, "10.0"},
		{"under hundred", 42.5, "42.5"},
		{"just under hundred", 99.95, "100.0"},
		{"exactly hundred", 100, "100"},
		{"over hundred rounds", 123.4, "123"},
		{"over hundred rounds up", 123.6, "124"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := threeDigits(tt.input); got != tt.expected {
				t.Errorf("threeDigits(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

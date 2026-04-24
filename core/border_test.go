package core

import (
	"testing"
)

// TestBorder_Horizontal verifies that Border.Horizontal correctly reports the
// total horizontal space (in cells) consumed by the left and right border
// sides. It exercises empty borders, ASCII and Unicode single-character
// borders, asymmetric borders where one side is missing, and multi-character
// sides to ensure the rune count — rather than byte length — is used.
func TestBorder_Horizontal(t *testing.T) {
	tests := []struct {
		name     string
		border   Border
		expected int
	}{
		{
			name:     "Empty border",
			border:   Border{},
			expected: 0,
		},
		{
			name: "Simple ASCII border",
			border: Border{
				Left:  "|",
				Right: "|",
			},
			expected: 2,
		},
		{
			name: "Unicode border",
			border: Border{
				Left:  "\u2502", // │
				Right: "\u2502", // │
			},
			expected: 2,
		},
		{
			name: "Mixed border (one side missing)",
			border: Border{
				Left:  "|",
				Right: "",
			},
			expected: 1,
		},
		{
			name: "Wide border chars (double width not handled by rune count but count is 1)",
			border: Border{
				Left:  "W",
				Right: "W",
			},
			expected: 2,
		},
		{
			name: "Multiple chars per side",
			border: Border{
				Left:  "||",
				Right: "||",
			},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.border.Horizontal(); got != tt.expected {
				t.Errorf("Border.Horizontal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestBorder_Vertical verifies that Border.Vertical correctly reports the
// total vertical space (in cells) consumed by the top and bottom border
// sides. It covers empty borders, ASCII and Unicode borders, asymmetric
// borders with a missing side, and long horizontal border strings to confirm
// that each side contributes at most a single row to the vertical extent.
func TestBorder_Vertical(t *testing.T) {
	tests := []struct {
		name     string
		border   Border
		expected int
	}{
		{
			name:     "Empty border",
			border:   Border{},
			expected: 0,
		},
		{
			name: "Simple ASCII border",
			border: Border{
				Top:    "-",
				Bottom: "-",
			},
			expected: 2,
		},
		{
			name: "Unicode border",
			border: Border{
				Top:    "\u2500", // ─
				Bottom: "\u2500", // ─
			},
			expected: 2,
		},
		{
			name: "Mixed border (one side missing)",
			border: Border{
				Top:    "-",
				Bottom: "",
			},
			expected: 1,
		},
		{
			name: "Long border line (should be capped at 1)",
			border: Border{
				Top:    "-----",
				Bottom: "-----",
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.border.Vertical(); got != tt.expected {
				t.Errorf("Border.Vertical() = %v, want %v", got, tt.expected)
			}
		})
	}
}

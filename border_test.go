package next

import (
	"testing"
)

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

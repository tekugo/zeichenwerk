package next

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/stretchr/testify/assert"
)

func TestTcellScreen_Put(t *testing.T) {
	// Setup
	mockScreen, err := NewMockScreen() // Initialize embedded screen
	if err != nil {
		t.Fatal(err)
	}

	// Create the wrapper
	screen := NewTcellScreen(mockScreen)

	// Define test cases
	tests := []struct {
		name         string
		clipX, clipY int
		clipW, clipH int
		putX, putY   int
		expectedAbsX int
		expectedAbsY int
		shouldCall   bool
	}{
		{
			name:         "Basic Put inside clip",
			clipX:        10,
			clipY:        5,
			clipW:        20,
			clipH:        10,
			putX:         2,
			putY:         3,
			expectedAbsX: 12, // 10 + 2
			expectedAbsY: 8,  // 5 + 3
			shouldCall:   true,
		},
		{
			name:         "Put at 0,0 inside clip",
			clipX:        5,
			clipY:        5,
			clipW:        10,
			clipH:        10,
			putX:         0,
			putY:         0,
			expectedAbsX: 5,
			expectedAbsY: 5,
			shouldCall:   true,
		},
		{
			name:       "Put outside clip (negative)",
			clipX:      10,
			clipY:      10,
			clipW:      10,
			clipH:      10,
			putX:       -1,
			putY:       0,
			shouldCall: false,
		},
		{
			name:       "Put outside clip (too large)",
			clipX:      10,
			clipY:      10,
			clipW:      10,
			clipH:      10,
			putX:       10, // Width is 10, so valid indices are 0-9
			putY:       0,
			shouldCall: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// execute
			screen.Clear()
			screen.Clip(tc.clipX, tc.clipY, tc.clipW, tc.clipH)
			screen.Put(tc.putX, tc.putY, "X")
			screen.Flush()

			// verify
			if tc.shouldCall {
				cell, _, _ := mockScreen.Get(tc.expectedAbsX, tc.expectedAbsY)
				assert.Equal(t, "X", cell)
			} else {
				cell, _, _ := mockScreen.Get(tc.expectedAbsX, tc.expectedAbsY)
				assert.Equal(t, " ", cell)
				cell, _, _ = mockScreen.Get(tc.putX+tc.clipX, tc.putY+tc.clipY)
				assert.Equal(t, " ", cell)
			}
		})
	}
}

func TestTcellScreen_Style(t *testing.T) {
	// Setup
	mockScreen, err := NewMockScreen() // Initialize embedded screen
	if err != nil {
		t.Fatal(err)
	}
	screen := NewTcellScreen(mockScreen).(*TcellScreen)

	tests := []struct {
		name       string
		fg         string
		bg         string
		font       string
		expectedFg int32
		expectedBg int32
		expectedAt tcell.AttrMask
	}{
		{
			name:       "Basic Red/Blue",
			fg:         "red",
			bg:         "blue",
			font:       "normal",
			expectedFg: 0xff0000,
			expectedBg: 0x0000ff,
			expectedAt: tcell.AttrNone,
		},
		{
			name:       "Bold Attribute",
			fg:         "",
			bg:         "",
			font:       "bold",
			expectedFg: -1,
			expectedBg: -1,
			expectedAt: tcell.AttrBold,
		},
		{
			name:       "Combined Attributes",
			fg:         "#00ff00",
			bg:         "",
			font:       "bold italic",
			expectedFg: 0x00ff00,
			expectedBg: -1,
			expectedAt: tcell.AttrBold | tcell.AttrItalic,
		},
		{
			name:       "Reset to Normal",
			fg:         "",
			bg:         "",
			font:       "normal",
			expectedFg: -1,
			expectedBg: -1,
			expectedAt: tcell.AttrNone,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// execute
			screen.Set(tc.fg, tc.bg, tc.font)

			// verify
			assert.Equal(t, tc.expectedFg, screen.style.GetForeground().Hex(), "Foreground color mismatch")
			assert.Equal(t, tc.expectedBg, screen.style.GetBackground().Hex(), "Background color mismatch")
			assert.Equal(t, tc.expectedAt, screen.style.GetAttributes(), "Attributes mismatch")
		})
	}
}

func TestTcellScreen_Set(t *testing.T) {
	// Setup
	mockScreen, err := NewMockScreen() // Initialize embedded screen
	if err != nil {
		t.Fatal(err)
	}
	screen := NewTcellScreen(mockScreen)

	tests := []struct {
		name       string
		fg         string
		bg         string
		font       string
		expectedFg int32
		expectedBg int32
		expectedAt tcell.AttrMask
	}{
		{
			name:       "Basic Red/Blue",
			fg:         "red",
			bg:         "blue",
			font:       "normal",
			expectedFg: 0xff0000,
			expectedBg: 0x0000ff,
			expectedAt: tcell.AttrNone,
		},
		{
			name:       "Bold Attribute",
			fg:         "",
			bg:         "",
			font:       "bold",
			expectedFg: -1,
			expectedBg: -1,
			expectedAt: tcell.AttrBold,
		},
		{
			name:       "Combined Attributes",
			fg:         "#33ff33",
			bg:         "",
			font:       "bold italic",
			expectedFg: 0x33ff33,
			expectedBg: -1,
			expectedAt: tcell.AttrBold | tcell.AttrItalic,
		},
		{
			name:       "Reset to Normal",
			fg:         "",
			bg:         "",
			font:       "normal",
			expectedFg: -1,
			expectedBg: -1,
			expectedAt: tcell.AttrNone,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// execute
			screen.Clear()
			screen.Set(tc.fg, tc.bg, tc.font)
			screen.Put(10, 10, "X")
			screen.Flush()

			t.Logf("Foreground Color %s: %x\n", tc.fg, color.GetColor(tc.fg).Hex())
			t.Logf("Background Color %s: %x\n", tc.bg, color.GetColor(tc.bg).Hex())

			// verify by checking the style of the cell at 10,10
			ch, style, _ := mockScreen.Get(10, 10)

			assert.Equal(t, "X", ch, "Character mismatch")
			assert.Equal(t, tc.expectedFg, style.GetForeground().Hex(), "Foreground color mismatch")
			assert.Equal(t, tc.expectedBg, style.GetBackground().Hex(), "Background color mismatch")

			// Mask out attributes not handled or defaults if necessary, but usually exact match works
			assert.Equal(t, tc.expectedAt, style.GetAttributes(), "Attributes mismatch")
		})
	}
}

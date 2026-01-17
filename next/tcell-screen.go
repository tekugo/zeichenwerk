package next

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// TcellScreen is a concrete implementation of the Screen interface using TCell v3.
// It provides a clipped and translated view into the terminal screen, allowing
// for windowed rendering of components.
type TcellScreen struct {
	screen              tcell.Screen // underlying tcell screen instance
	style               tcell.Style  // current style state for drawing
	x, y, width, height int          // clipping region and translation offset
}

// MewTcellScreen creates a new TcellScreen instance.
//
// Parameters:
//   - screen: The underlying tcell screen instance to use.
func NewTcellScreen(screen tcell.Screen) Screen {
	return &TcellScreen{
		screen: screen,
		style:  tcell.StyleDefault,
	}
}

// Clear clears the entire underlying screen.
// Note: This operation ignores the current clipping region and clears everything.
func (t *TcellScreen) Clear() {
	t.screen.Clear()
}

// Clip sets the clipping region and translation origin for subsequent operations.
// The (x, y) coordinates become the new (0, 0) origin for Put and Get.
//
// Parameters:
//   - x, y: The top-left corner of the clipping region (Absolute Screen Coordinates).
//   - width, height: The dimensions of the clipping region.
func (t *TcellScreen) Clip(x, y, width, height int) {
	t.x = x
	t.y = y
	t.width = width
	t.height = height
}

// Flush synchronizes the internal buffer with the actual terminal display.
// This should be called to make recent drawing operations visible to the user.
func (t *TcellScreen) Flush() {
	t.screen.Show()
}

// Get retrieves the character at the specified position relative to the clipping origin.
//
// Parameters:
//   - x, y: Position relative to the top-left of the current clipping region.
//
// Returns:
//   - string: The character at the position (might include combining characters).
func (t *TcellScreen) Get(x, y int) string {
	ch, _, _ := t.screen.Get(x+t.x, y+t.y)
	return ch
}

// Put places a character at the specified position relative to the clipping origin.
// drawing is clipped to the current region dimensions.
//
// Parameters:
//   - x, y: Position relative to the top-left of the current clipping region.
//   - ch: The character to display.
func (t *TcellScreen) Put(x, y int, ch string) {
	// Check if the point is within the clipping region (relative coordinates check)
	if x >= 0 && (t.width == 0 || x < t.width) && y >= 0 && (t.height == 0 || y < t.height) {
		t.screen.Put(x+t.x, y+t.y, ch, t.style)
	}
}

// Set updates the current drawing style (foreground, background, font attributes).
//
// Parameters:
//   - foreground: Color name or hex string (e.g., "red", "#ff0000").
//   - background: Color name or hex string.
//   - font: Space-separated list of attributes (e.g., "bold italic blink").
func (t *TcellScreen) Set(foreground, background, font string) {
	next := tcell.StyleDefault

	if background != "" {
		next = next.Background(color.GetColor(background))
	}

	if foreground != "" {
		next = next.Foreground(color.GetColor(foreground))
	}

	for part := range strings.SplitSeq(font, " ") {
		option := strings.ToLower(strings.TrimSpace(part))
		switch option {
		case "blink":
			next = next.Blink(true)
		case "bold":
			next = next.Bold(true)
		case "normal":
			next = next.Blink(false).Bold(false).Italic(false).Underline(false).StrikeThrough(false)
		case "italic":
			next = next.Italic(true)
		case "strikethrough":
			next = next.StrikeThrough(true)
		case "underline":
			next = next.Underline(true)
		}
	}

	t.style = next
}

// Style retrieves the style at the specified position relative to the clipping origin.
//
// Parameters:
//   - x, y: Position relative to the top-left of the current clipping region.
//
// Returns:
//   - tcell.Style: The style at the position.
func (t *TcellScreen) Style(x, y int) tcell.Style {
	_, style, _ := t.screen.Get(x+t.x, y+t.y)
	fmt.Printf("Style at %d,%d: %v\n", x+t.x, y+t.y, style)
	return style
}

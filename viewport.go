package zeichenwerk

import "github.com/gdamore/tcell/v2"

// Viewport represents a clipped rendering area within a larger screen or surface.
// It acts as a bounded window that restricts drawing operations to a specific
// rectangular region, providing clipping functionality for widgets and layouts.
//
// Features:
//   - Bounded rendering area with automatic clipping
//   - Coordinate translation and bounds checking
//   - Screen content access within the viewport bounds
//   - Protection against drawing outside the designated area
//
// Viewports are commonly used for:
//   - Implementing scrollable content areas
//   - Creating bounded drawing regions for widgets
//   - Clipping content to prevent overflow
//   - Implementing windowed or paned interfaces
//
// The viewport enforces strict bounds checking and will panic if attempts
// are made to draw outside the designated clipping area.
type Viewport struct {
	screen              Screen // The underlying screen or surface to draw on
	x, y, width, height int    // The viewport's position and dimensions (clipping bounds)
	tx, ty              int    // Translation offset for coordinate mapping
}

// NewViewport creates a new viewport with the specified bounds on the given screen.
// The viewport will clip all drawing operations to the specified rectangular area,
// preventing content from being rendered outside the designated bounds.
//
// Parameters:
//   - screen: The underlying screen or surface to draw on
//   - x: The x-coordinate of the viewport's top-left corner
//   - y: The y-coordinate of the viewport's top-left corner
//   - w: The width of the viewport in characters/cells
//   - h: The height of the viewport in characters/cells
//
// Returns:
//   - *Viewport: A new viewport instance with the specified bounds
//
// Example usage:
//
//	viewport := NewViewport(screen, 10, 5, 80, 24)  // 80x24 area at (10,5)
func NewViewport(screen Screen, x, y, w, h int) *Viewport {
	return &Viewport{screen: screen, x: x, y: y, width: w, height: h}
}

// GetContent retrieves the content at the specified coordinates from the underlying screen.
// This method provides read access to screen content without bounds checking,
// allowing inspection of content anywhere on the underlying screen surface.
//
// Parameters:
//   - x: The x-coordinate to read from
//   - y: The y-coordinate to read from
//
// Returns:
//   - rune: The primary character at the specified position
//   - []rune: Any combining characters at the position
//   - tcell.Style: The style information for the character
//   - int: The character width (for multi-cell characters)
func (v *Viewport) GetContent(x, y int) (rune, []rune, tcell.Style, int) {
	return v.screen.GetContent(x, y)
}

// SetContent sets the content at the specified coordinates with strict bounds checking.
// This method enforces the viewport's clipping bounds and will panic if an attempt
// is made to draw outside the designated viewport area.
//
// The bounds checking ensures that:
//   - x must be >= viewport.x and < viewport.x + viewport.width
//   - y must be >= viewport.y and <= viewport.y + viewport.height
//
// Parameters:
//   - x: The x-coordinate to write to (must be within viewport bounds)
//   - y: The y-coordinate to write to (must be within viewport bounds)
//   - primary: The primary character to set
//   - combining: Any combining characters to apply
//   - style: The style information for the character
//
// Panics:
//   - If the coordinates are outside the viewport's clipping area
func (v *Viewport) SetContent(x, y int, primary rune, combining []rune, style tcell.Style) {
	if x >= v.x && x < v.x+v.width && y >= v.y && y <= v.y+v.height {
		v.screen.SetContent(x, y, primary, combining, style)
	} else {
		panic("outside clipping area")
	}
}

// Bounds returns the viewport's position and dimensions.
// This provides access to the clipping area boundaries for bounds checking
// or coordinate calculations.
//
// Returns:
//   - int: x-coordinate of the viewport's top-left corner
//   - int: y-coordinate of the viewport's top-left corner
//   - int: width of the viewport in characters/cells
//   - int: height of the viewport in characters/cells
func (v *Viewport) Bounds() (int, int, int, int) {
	return v.x, v.y, v.width, v.height
}

// Contains checks if the specified coordinates are within the viewport's bounds.
// This is useful for validating coordinates before attempting to draw content.
//
// Parameters:
//   - x: The x-coordinate to check
//   - y: The y-coordinate to check
//
// Returns:
//   - bool: true if the coordinates are within bounds, false otherwise
func (v *Viewport) Contains(x, y int) bool {
	return x >= v.x && x < v.x+v.width && y >= v.y && y <= v.y+v.height
}

// Translate sets the coordinate translation offset for the viewport.
// This allows for coordinate mapping between different coordinate systems,
// useful for implementing scrolling or panning functionality.
//
// Parameters:
//   - tx: The x-axis translation offset
//   - ty: The y-axis translation offset
func (v *Viewport) Translate(tx, ty int) {
	v.tx = tx
	v.ty = ty
}

// Translation returns the current coordinate translation offset.
// This can be used to retrieve the current scrolling or panning position.
//
// Returns:
//   - int: The x-axis translation offset
//   - int: The y-axis translation offset
func (v *Viewport) Translation() (int, int) {
	return v.tx, v.ty
}

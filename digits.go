package zeichenwerk

// Digits represents a widget that displays text using large ASCII art-style characters.
// It renders digits (0-9), letters (A-Z), and various symbols using Unicode box-drawing
// characters to create visually prominent display elements.
//
// Features:
//   - Large-format character rendering (approximately 4x3 character cells per digit)
//   - Support for digits 0-9, letters A-Z, and common symbols (.,#:)
//   - Unicode box-drawing characters for clean appearance
//   - Non-focusable by design (display-only widget)
//   - Automatic theming support through the style system
//
// Supported characters:
//   - Digits: 0, 1, 2, 3, 4, 5, 6, 7, 8, 9
//   - Letters: A, B, C, D, E, F (hexadecimal support)
//   - Symbols: . (period), , (comma), # (hash), : (colon)
//
// Use cases:
//   - Digital clock displays
//   - Numeric counters and indicators
//   - Hexadecimal value displays
//   - Status displays requiring high visibility
//   - Dashboard elements and scoreboards
//
// The Digits widget is particularly useful when you need to display numeric
// information prominently, such as in monitoring dashboards, digital clocks,
// or any interface where large, easily readable numbers are important.
type Digits struct {
	BaseWidget
	// Text contains the string to be rendered as large digits.
	// Each character in the string will be rendered as a large ASCII art
	// representation using approximately 4x3 character cells.
	// Unsupported characters may be rendered as blank space or placeholder.
	Text string
}

// NewDigits creates a new digits widget for displaying large ASCII art-style text.
// The widget is non-focusable by design since it's purely for display purposes.
//
// The widget automatically sets a size hint based on the text length:
// - Width: len(text) * 4 characters (each digit uses ~4 character cells)
// - Height: 3 characters (standard height for large digits)
//
// Parameters:
//   - id: Unique identifier for the digits widget
//   - text: Text content to display as large digits/characters
//
// Returns:
//   - *Digits: A new digits widget instance
//
// Example usage:
//
//	// Display current time
//	timeDigits := NewDigits("clock", "12:34")
//
//	// Display hexadecimal value
//	hexDigits := NewDigits("hex-display", "DEADBEEF")
//
//	// Display counter value
//	counterDigits := NewDigits("counter", "00042")
//
// The widget integrates with the theming system and can be styled using
// the "digits" selector in theme configurations.
func NewDigits(id string, text string) *Digits {
	return &Digits{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Text:       text,
	}
}

// SetText updates the text content of the digits widget.
// This method allows dynamic updating of the displayed content after the widget
// has been created. The widget will automatically re-render with the new text
// on the next render cycle.
//
// Parameters:
//   - text: New text content to display as large digits/characters
//
// Example usage:
//
//	// Update a clock display
//	clockWidget.SetText(time.Now().Format("15:04"))
//
//	// Update a counter
//	counterWidget.SetText(fmt.Sprintf("%05d", currentValue))
func (d *Digits) SetText(text string) {
	d.Text = text
}

// GetText returns the current text content of the digits widget.
// This method provides access to the currently displayed text.
//
// Returns:
//   - string: Current text content being displayed
func (d *Digits) GetText() string {
	return d.Text
}

// Render draws the digits widget using large ASCII art-style characters.
// This method is called by the framework during the rendering phase.
// The actual character rendering is handled by the renderer's renderDigits method.
//
// The rendering process:
// 1. Delegates to the renderer's renderDigits method
// 2. Each character is rendered as approximately 4x3 character cells
// 3. Uses Unicode box-drawing characters for clean appearance
// 4. Applies current theme styling
//
// Parameters:
//   - screen: Screen interface for drawing operations
func (d *Digits) Render(screen Screen) {
	// The actual rendering is handled by the renderer
	// This method is here for interface compliance and documentation
	// The real implementation is in renderer.go -> renderDigits()
}

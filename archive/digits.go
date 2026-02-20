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

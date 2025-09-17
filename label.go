package zeichenwerk

import "fmt"

// Label represents a text display widget that shows static or dynamic text content.
// It is a non-interactive widget primarily used for displaying information, captions,
// descriptions, or other textual content in the user interface.
//
// Features:
//   - Text content display with configurable alignment
//   - Support for dynamic text updates
//   - Configurable text alignment (left, center, right)
//   - Integration with the styling system for visual customization
//   - Non-interactive (doesn't handle user input events)
//
// Labels are commonly used for:
//   - Field labels and captions
//   - Status messages and information display
//   - Headers and titles
//   - Descriptive text in forms and dialogs
type Label struct {
	BaseWidget
	Text  string // The text content to display in the label
	Align string // Text alignment within the label bounds ("left", "center", "right")
}

// NewLabel creates a new label widget with the specified ID and text content.
// The label is initialized with default alignment (typically left-aligned)
// and can be customized after creation by setting the Align field.
//
// Parameters:
//   - id: Unique identifier for the label widget
//   - text: The initial text content to display
//
// Returns:
//   - *Label: A new label widget instance
//
// Example usage:
//
//	label := NewLabel("status", "Ready")
//	label.Align = "center"  // Optional: set text alignment
func NewLabel(id string, text string) *Label {
	return &Label{
		BaseWidget: BaseWidget{id: id},
		Text:       text,
		Align:      "left", // Set default alignment
	}
}

// Info returns a human-readable description of the label's current state.
// This includes both the outer bounds and inner content area dimensions,
// which is useful for debugging layout and styling issues.
//
// The format includes:
//   - Outer bounds: position and size including margins, borders, padding
//   - Content area: inner area available for text rendering
//   - Widget type identifier
//
// Returns:
//   - string: Formatted string with label bounds and content area information
func (l *Label) Info() string {
	x, y, w, h := l.Bounds()
	cx, cy, cw, ch := l.Content()
	return fmt.Sprintf("@%d.%d %d:%d (%d.%d %d:%d) label", x, y, w, h, cx, cy, cw, ch)
}

// SetText updates the label's text content and triggers a refresh.
// This is a convenience method for dynamically updating the label's
// displayed text, commonly used for status updates or dynamic content.
//
// Parameters:
//   - text: The new text content to display
func (l *Label) SetText(text string) {
	l.Text = text
	l.Refresh()
}

// SetAlignment sets the text alignment for the label.
// This controls how the text is positioned within the label's content area.
//
// Parameters:
//   - align: The alignment mode ("left", "center", "right")
func (l *Label) SetAlignment(align string) {
	l.Align = align
	l.Refresh()
}

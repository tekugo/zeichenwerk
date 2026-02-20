package next

import "unicode/utf8"

// Static represents a text display widget that shows static content.
// It is a non-interactive widget primarily used for displaying information, captions,
// descriptions, or other textual content in the user interface.
type Static struct {
	Component
	Text      string // The text content to display in the label
	Alignment string // Text alignment within the label bounds ("left", "center", "right")
}

// NewStatic creates a new static widget with the specified ID and text content.
// The static is initialized with default alignment (typically left-aligned)
// and can be customized after creation by setting the Align field.
//
// Parameters:
//   - id: Unique identifier for the static widget
//   - text: The initial text content to display
//
// Returns:
//   - *Static: A new static widget instance
func NewStatic(id string, text string) *Static {
	return &Static{
		Component: Component{id: id, hwidth: utf8.RuneCountInString(text), hheight: 1},
		Text:      text,
		Alignment: "left", // Set default alignment
	}
}

// Render renders the static widget to the screen using the Renderer.
func (s *Static) Render(r *Renderer) {
	s.Component.Render(r)
	cx, cy, cw, _ := s.Content()
	r.Text(cx, cy, s.Text, cw)
}

// SetText updates the static's text content and triggers a refresh.
// This is a convenience method for dynamically updating the static's
// displayed text, commonly used for status updates or dynamic content.
//
// Parameters:
//   - text: The new text content to display
func (s *Static) SetText(text string) {
	s.Text = text
	s.Refresh()
}

// SetAlignment sets the text alignment for the label.
// This controls how the text is positioned within the label's content area.
//
// Parameters:
//   - align: The alignment mode ("left", "center", "right")
func (s *Static) SetAlignment(align string) {
	s.Alignment = align
	s.Refresh()
}

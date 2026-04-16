package zeichenwerk

import (
	"fmt"
	"unicode/utf8"
)

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
//   - class: Style class
//   - text: The initial text content to display
//
// Returns:
//   - *Static: A new static widget instance
func NewStatic(id, class string, text string) *Static {
	return &Static{
		Component: Component{id: id, class: class},
		Text:      text,
		Alignment: "left", // Set default alignment
	}
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies a theme's styles to the component.
func (s *Static) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("static"))
}

// Hint returns the natural size of the static derived from its current text.
// If hwidth or hheight has been set explicitly, both are returned as-is.
func (s *Static) Hint() (int, int) {
	if s.hwidth != 0 || s.hheight != 0 {
		return s.hwidth, s.hheight
	}
	return utf8.RuneCountInString(s.Text), 1
}

// Render renders the static widget to the screen using the Renderer.
func (s *Static) Render(r *Renderer) {
	s.Component.Render(r)
	cx, cy, cw, _ := s.Content()
	if s.Flag(FlagRight) {
		textWidth := utf8.RuneCountInString(s.Text)
		cx = max(cx, cx+cw-textWidth)
		r.Text(cx, cy, s.Text, 0)
	} else {
		r.Text(cx, cy, s.Text, cw)
	}
}

// ---- Summarizer -----------------------------------------------------------

// Summary returns the static text for Dump output (truncated to 60 runes).
func (s *Static) Summary() string {
	r := []rune(s.Text)
	if len(r) > 60 {
		return string(r[:60]) + "…"
	}
	return s.Text
}

// ---- Setter ---------------------------------------------------------------

// Set updates the displayed text. Non-string values are formatted with
// fmt.Sprintf("%v", value).
func (s *Static) Set(value any) {
	if str, ok := value.(string); ok {
		s.Text = str
	} else {
		s.Text = fmt.Sprintf("%v", value)
	}
	s.Refresh()
}

// SetAlignment sets the text alignment for the label.
func (s *Static) SetAlignment(align string) {
	s.Alignment = align
	s.Refresh()
}

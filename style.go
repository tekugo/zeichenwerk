package zeichenwerk

import (
	"fmt"
)

var DefaultStyle = Style{
	parent:     nil,
	fixed:      true,
	background: "black",
	foreground: "white",
	font:       "",
	border:     "none",
	cursor:     "",
	margin:     NewInsets(0),
	padding:    NewInsets(0),
}

var NoInsets = &Insets{}

// Style defines the visual appearance and layout properties of TUI widgets.
// It provides a comprehensive styling system similar to CSS, allowing
// fine-grained control over widget appearance including colors, font,
// spacing and borders.
//
// The style system supports:
//   - Color management for foreground and background
//   - Border styling with various border types
//   - Spacing control through margins and padding
//   - Cursor styling for interactive widgets
//
// Styles can be cascaded (inherited) from other styles, allowing for
// hierarchical styling systems and theme inheritance. The box model follows
// CSS conventions with content area, padding, border, and margin layers.
//
// Box model (from inside out):
//  1. Content area (actual widget content)
//  2. Padding (inner spacing with background color)
//  3. Border (decorative border around padding)
//  4. Margin (outer transparent spacing)
type Style struct {
	selector   string  // Style selector
	parent     *Style  // Reference to the parent style
	fixed      bool    // Fixed flag for not modifiable styles
	background string  // Background color name or hex code (empty = transparent)
	foreground string  // Text/foreground color name or hex code
	font       string  // Font style
	border     string  // Border style identifier (empty = no border)
	cursor     string  // Cursor style for interactive widgets
	margin     *Insets // Outer spacing around the widget (transparent)
	padding    *Insets // Inner spacing within the widget (with background color)
}

// ---- Constructor Function -------------------------------------------------

func NewStyle(params ...string) *Style {
	if len(params) > 0 {
		return &Style{
			selector: params[0],
		}
	} else {
		return &Style{}
	}
}

// ---- Property Access Methods ----------------------------------------------

func (s *Style) Background() string {
	if s.background == "" && s.parent != nil {
		return s.parent.Background()
	} else {
		return s.background
	}
}

func (s *Style) Border() string {
	if s.border == "" && s.parent != nil {
		return s.parent.Border()
	} else {
		return s.border
	}
}

func (s *Style) Cursor() string {
	if s.cursor == "" && s.parent != nil {
		return s.parent.Cursor()
	} else {
		return s.cursor
	}
}

func (s *Style) Fixed() bool {
	return s.fixed
}

func (s *Style) Font() string {
	if s.font == "" && s.parent != nil {
		return s.parent.Font()
	} else {
		return s.font
	}
}

func (s *Style) Foreground() string {
	if s.foreground == "" && s.parent != nil {
		return s.parent.Foreground()
	} else {
		return s.foreground
	}
}

func (s *Style) Margin() *Insets {
	if s.margin == nil && s.parent != nil {
		return s.parent.Margin()
	} else if s.margin != nil {
		return s.margin
	} else {
		return NoInsets
	}
}

func (s *Style) Padding() *Insets {
	if s.padding == nil && s.parent != nil {
		return s.parent.Padding()
	} else if s.padding != nil {
		return s.padding
	} else {
		return NoInsets
	}
}

func (s *Style) Parent() *Style {
	return s.parent
}

// ---- Property Change Methods ----------------------------------------------

func (s *Style) WithBackground(background string) *Style {
	style := s.Modifiable()
	style.background = background
	return style
}

func (s *Style) WithBorder(border string) *Style {
	style := s.Modifiable()
	style.border = border
	return style
}

func (s *Style) WithColors(foreground, background string) *Style {
	style := s.Modifiable()
	style.foreground = foreground
	style.background = background
	return style
}

func (s *Style) WithCursor(cursor string) *Style {
	style := s.Modifiable()
	style.cursor = cursor
	return style
}

func (s *Style) WithFont(font string) *Style {
	style := s.Modifiable()
	style.font = font
	return style
}

func (s *Style) WithForeground(foreground string) *Style {
	style := s.Modifiable()
	style.foreground = foreground
	return style
}

func (s *Style) WithMargin(values ...int) *Style {
	style := s.Modifiable()
	style.margin = NewInsets(values...)
	return style
}

func (s *Style) WithPadding(values ...int) *Style {
	style := s.Modifiable()
	style.padding = NewInsets(values...)
	return style
}

func (s *Style) WithParent(parent *Style) *Style {
	style := s.Modifiable()
	style.parent = parent
	return style
}

// ---- Helper methods -------------------------------------------------------

// Horizontal calculates the total horizontal spacing consumed by this style.
// This includes left and right margins, padding, and border space.
// The calculation is used by layout algorithms to determine the total
// horizontal space required by a widget beyond its content area.
//
// Calculation:
//   - Margin left + margin right
//   - Padding left + padding right
//   - Border space (2 characters if border is present, 0 if not)
//
// Returns:
//   - int: Total horizontal spacing in characters/pixels
//
// Example:
//   - Margin: 2 left, 3 right
//   - Padding: 1 left, 1 right
//   - Border: present
//   - Result: 2 + 3 + 1 + 1 + 2 = 9
func (s *Style) Horizontal() int {
	var insets *Insets
	result := 0
	if insets = s.Margin(); insets != nil {
		result += insets.Horizontal()
	}
	if insets = s.Padding(); insets != nil {
		result += insets.Horizontal()
	}
	if s.Border() != "" && s.Border() != "none" {
		result += 2
	}
	return result
}

// Vertical calculates the total vertical spacing consumed by this style.
// This includes top and bottom margins, padding, and border space.
// The calculation is used by layout algorithms to determine the total
// vertical space required by a widget beyond its content area.
//
// Calculation:
//   - Margin top + margin bottom
//   - Padding top + padding bottom
//   - Border space (2 lines if border is present, 0 if not)
//
// Returns:
//   - int: Total vertical spacing in characters/pixels
//
// Example:
//   - Margin: 1 top, 2 bottom
//   - Padding: 1 top, 1 bottom
//   - Border: present
//   - Result: 1 + 2 + 1 + 1 + 2 = 7
func (s *Style) Vertical() int {
	var insets *Insets
	result := 0
	if insets = s.Margin(); insets != nil {
		result += insets.Vertical()
	}
	if insets = s.Padding(); insets != nil {
		result += insets.Vertical()
	}
	if s.Border() != "" && s.Border() != "none" {
		result += 2
	}
	return result
}

func (s *Style) Fix() *Style {
	s.fixed = true
	return s
}

func (s *Style) Info() string {
	return fmt.Sprintf(
		`Selector  :%s
Parent    : %s
Background: %s
Foreground: %s
Border    : %s
Cursor    : %s
Margin    : %s
Padding   : %s`,
		s.selector, "", s.Background(), s.Foreground(), s.Border(), s.Cursor(), s.Margin().Info(), s.Padding().Info())
}

func (s *Style) Modifiable() *Style {
	if s.fixed {
		return NewStyle("(custom)").WithParent(s)
	} else {
		return s
	}
}

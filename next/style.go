package next

import "fmt"

// DefaultStyle is the base style used when no other style is specified.
// It provides reasonable defaults for all style properties:
// - Fixed: true (cannot be modified directly)
// - Background: black
// - Foreground: white
// - Border: none
// - Font: empty (system default)
// - Cursor: empty
// - Margin: 0
// - Padding: 0
var DefaultStyle = Style{
	parent:     nil,
	fixed:      true,
	background: "black",
	foreground: "white",
	font:       "",
	border:     "none",
	cursor:     "",
	margin:     NoInsets,
	padding:    NoInsets,
}

// Style defines the visual appearance and layout properties of TUI widgets.
// It provides a comprehensive styling system similar to CSS, allowing
// fine-grained control over widget appearance including colors, font,
// spacing and borders.
//
// Styles can be cascaded (inherited) from other styles, allowing for
// hierarchical styling systems and theme inheritance. The box model follows
// CSS conventions with content area, padding, border, and margin layers. If
// a property is not set in a style, it will inherit the value from its parent.
type Style struct {
	selector   string  // Style selector
	parent     *Style  // Reference to the parent style
	fixed      bool    // Fixed flag for unmodifiable styles
	margin     *Insets // Outer spacing around the widget (transparent)
	padding    *Insets // Inner spacing inside the widget (with background color)
	border     string  // Border style
	shadow     string  // Shadow style
	background string  // Background color
	foreground string  // Foreground color
	font       string  // Font style
	cursor     string  // Cursor style
}

// NewStyle creates a new Style instance.
//
// Parameters:
//   - selector: Optional CSS-like selector string for debugging and identification.
//     If multiple strings are provided, only the first one is used.
//
// Returns:
//   - A pointer to the new Style instance.
func NewStyle(selector ...string) *Style {
	if len(selector) > 0 {
		return &Style{selector: selector[0]}
	} else {
		return &Style{}
	}
}

// ---- Property Access Methods ----------------------------------------------

// Background returns the background color of the style.
// If the background is not set in this style, it inherits from the parent.
// Returns an empty string if no background is set in the hierarchy.
func (s *Style) Background() string {
	if s.background == "" && s.parent != nil {
		return s.parent.Background()
	} else {
		return s.background
	}
}

// Border returns the border style.
// If the border is not set in this style, it inherits from the parent.
// Common values include "none", "single", "double", "rounded", "hidden".
// Returns an empty string if no border is set in the hierarchy.
func (s *Style) Border() string {
	if s.border == "" && s.parent != nil {
		return s.parent.Border()
	} else {
		return s.border
	}
}

// Cursor returns the cursor style.
// If the cursor is not set in this style, it inherits from the parent.
// Returns an empty string if no cursor is set in the hierarchy.
func (s *Style) Cursor() string {
	if s.cursor == "" && s.parent != nil {
		return s.parent.Cursor()
	} else {
		return s.cursor
	}
}

// Fixed returns whether this style is immutable.
// Fixed styles cannot be modified directly; modification methods will return
// a new Style instance wrapping the fixed style as a parent.
func (s *Style) Fixed() bool {
	return s.fixed
}

// Font returns the font style.
// If the font is not set in this style, it inherits from the parent.
// Returns an empty string if no font is set in the hierarchy.
func (s *Style) Font() string {
	if s.font == "" && s.parent != nil {
		return s.parent.Font()
	} else {
		return s.font
	}
}

// Foreground returns the foreground code/color of the style.
// If the foreground is not set in this style, it inherits from the parent.
// Returns an empty string if no foreground is set in the hierarchy.
func (s *Style) Foreground() string {
	if s.foreground == "" && s.parent != nil {
		return s.parent.Foreground()
	} else {
		return s.foreground
	}
}

// Margin returns the margin insets (outer spacing).
// If margin is not set in this style, it inherits from the parent.
// Returns NoInsets (all zeros) if no margin is set in the hierarchy.
func (s *Style) Margin() *Insets {
	if s.margin == nil && s.parent != nil {
		return s.parent.Margin()
	} else if s.margin != nil {
		return s.margin
	} else {
		return NoInsets
	}
}

// Padding returns the padding insets (inner spacing).
// If padding is not set in this style, it inherits from the parent.
// Returns NoInsets (all zeros) if no padding is set in the hierarchy.
func (s *Style) Padding() *Insets {
	if s.padding == nil && s.parent != nil {
		return s.parent.Padding()
	} else if s.padding != nil {
		return s.padding
	} else {
		return NoInsets
	}
}

// Parent returns the parent style of this style.
// Returns nil if this style has no parent (is a root style).
func (s *Style) Parent() *Style {
	return s.parent
}

// Shadow returns the shadow style.
// If the shadow is not set in this style, it inherits from the parent.
// Returns an empty string if no shadow is set in the hierarchy.
func (s *Style) Shadow() string {
	if s.shadow == "" && s.parent != nil {
		return s.parent.Shadow()
	} else {
		return s.shadow
	}
}

// ---- Property Change Methods ----------------------------------------------

// WithBackground sets the background color of the style.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithBackground(background string) *Style {
	style := s.Modifiable()
	style.background = background
	return style
}

// WithBorder sets the border style.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithBorder(border string) *Style {
	style := s.Modifiable()
	style.border = border
	return style
}

// WithColors sets both foreground and background colors.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithColors(foreground, background string) *Style {
	style := s.Modifiable()
	style.foreground = foreground
	style.background = background
	return style
}

// WithCursor sets the cursor style.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithCursor(cursor string) *Style {
	style := s.Modifiable()
	style.cursor = cursor
	return style
}

// WithFont sets the font style.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithFont(font string) *Style {
	style := s.Modifiable()
	style.font = font
	return style
}

// WithForeground sets the foreground color.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithForeground(foreground string) *Style {
	style := s.Modifiable()
	style.foreground = foreground
	return style
}

// WithMargin sets the margin (outer spacing).
// Values are interpreted as CSS margin: top, right, bottom, left.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithMargin(values ...int) *Style {
	style := s.Modifiable()
	style.margin = NewInsets(values...)
	return style
}

// WithPadding sets the padding (inner spacing).
// Values are interpreted as CSS padding: top, right, bottom, left.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithPadding(values ...int) *Style {
	style := s.Modifiable()
	style.padding = NewInsets(values...)
	return style
}

// WithParent sets the parent style for inheritance.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithParent(parent *Style) *Style {
	style := s.Modifiable()
	style.parent = parent
	return style
}

// WithShadow sets the shadow style.
// If the style is fixed, it returns a new child style with the change.
func (s *Style) WithShadow(shadow string) *Style {
	style := s.Modifiable()
	style.shadow = shadow
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

// Left calculates the total left spacing (margin + padding + border).
func (s *Style) Left() int {
	var insets *Insets
	left := 0
	if insets = s.Margin(); insets != nil {
		left += insets.Left
	}
	if insets = s.Padding(); insets != nil {
		left += insets.Left
	}
	if s.Border() != "" && s.Border() != "none" {
		left++
	}
	return left
}

// Top calculates the total top spacing (margin + padding + border).
func (s *Style) Top() int {
	var insets *Insets
	top := 0
	if insets = s.Margin(); insets != nil {
		top += insets.Top
	}
	if insets = s.Padding(); insets != nil {
		top += insets.Top
	}
	if s.Border() != "" && s.Border() != "none" {
		top++
	}
	return top
}

// Fix marks the style as immutable.
// Subsequent modifications will return a new derived style.
func (s *Style) Fix() *Style {
	s.fixed = true
	return s
}

// Info returns a formatted string containing all style properties
// for debugging purposes.
func (s *Style) Info() string {
	return fmt.Sprintf(
		`Selector  :%s
Parent    : %s
Background: %s
Foreground: %s
Border    : %s
Shadow    : %s
Cursor    : %s
Margin    : %s
Padding   : %s`,
		s.selector, "", s.Background(), s.Foreground(), s.Border(), s.Shadow(), s.Cursor(), s.Margin().Info(), s.Padding().Info())
}

// Modifiable returns a modifiable version of the style.
// If the style is already modifiable (not fixed), returns self.
// If the style is fixed, returns a new child style that inherits from this one.
// This is used internally by all "With..." methods.
func (s *Style) Modifiable() *Style {
	if s.fixed {
		return NewStyle("(custom)").WithParent(s)
	} else {
		return s
	}
}

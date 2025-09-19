package zeichenwerk

import (
	"fmt"
)

// Style defines the visual appearance and layout properties of TUI widgets.
// It provides a comprehensive styling system similar to CSS, allowing fine-grained
// control over widget appearance including colors, spacing, borders, and sizing hints.
//
// The style system supports:
//   - Color management for foreground and background
//   - Border styling with various border types
//   - Spacing control through margins and padding
//   - Size hints for layout algorithms
//   - Cursor styling for interactive widgets
//   - Rendering hints for specialized display modes
//
// Styles can be cascaded (inherited) from other styles, allowing for hierarchical
// styling systems and theme inheritance. The box model follows CSS conventions
// with content area, padding, border, and margin layers.
//
// Box model (from inside out):
//  1. Content area (actual widget content)
//  2. Padding (inner spacing with background color)
//  3. Border (decorative border around padding)
//  4. Margin (outer transparent spacing)
type Style struct {
	Background string  // Background color name or hex code (empty = transparent)
	Foreground string  // Text/foreground color name or hex code
	Font       string  // Font style
	Border     string  // Border style identifier (empty = no border)
	Cursor     string  // Cursor style for interactive widgets
	Margin     *Insets // Outer spacing around the widget (transparent)
	Padding    *Insets // Inner spacing within the widget (with background color)
	Width      int     // Preferred width hint for layout algorithms
	Height     int     // Preferred height hint for layout algorithms
	Render     string  // Special rendering mode hint for the widget
}

// NewStyle creates a new style with the specified foreground and background colors.
// All other style properties are initialized to their default values (empty/zero).
// This is a convenience constructor for creating basic color-only styles.
//
// Parameters:
//   - fg: Foreground (text) color name or hex code
//   - bg: Background color name or hex code
//
// Returns:
//   - *Style: A new style instance with the specified colors
//
// Example usage:
//
//	// Create a style with white text on blue background
//	style := NewStyle("white", "blue")
//
//	// Create a style with hex colors
//	style := NewStyle("#ffffff", "#0000ff")
func NewStyle(fg, bg string) *Style {
	return &Style{
		Background: bg,
		Foreground: fg,
	}
}

// Cascade applies properties from another style to this style, implementing
// style inheritance. Only non-empty/non-zero properties from the other style
// are applied, allowing for selective property overriding.
//
// This method implements a CSS-like cascading behavior where child styles
// can inherit and override parent style properties. The cascading follows
// these rules:
//   - Empty strings are not applied (preserves existing values)
//   - Zero values for Width/Height are not applied
//   - Nil pointers for Margin/Padding are not applied
//   - Non-empty/non-zero values override existing properties
//
// Parameters:
//   - other: The style to cascade from (can be nil for no-op)
//
// Example usage:
//
//	baseStyle := NewStyle("white", "blue")
//	baseStyle.SetBorder("solid")
//
//	childStyle := NewStyle("", "")  // Empty colors
//	childStyle.Cascade(baseStyle)   // Inherits white/blue colors and border
//	childStyle.SetForeground("red") // Override just the foreground
func (s *Style) Cascade(other *Style) {
	if other == nil {
		return
	}
	if other.Background != "" {
		s.Background = other.Background
	}
	if other.Foreground != "" {
		s.Foreground = other.Foreground
	}
	if other.Font != "" {
		s.Font = other.Font
	}
	if other.Border != "" {
		s.Border = other.Border
	}
	if other.Cursor != "" {
		s.Cursor = other.Cursor
	}
	if other.Margin != nil {
		copy := *other.Margin
		s.Margin = &copy
	}
	if other.Padding != nil {
		copy := *other.Padding
		s.Padding = &copy
	}
	if other.Width != 0 {
		s.Width = other.Width
	}
	if other.Height != 0 {
		s.Height = other.Height
	}
	if other.Render != "" {
		s.Render = other.Render
	}
}

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
	result := 0
	if s.Margin != nil {
		result += s.Margin.Horizontal()
	}
	if s.Padding != nil {
		result += s.Padding.Horizontal()
	}
	if s.Border != "" {
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
	result := 0
	if s.Margin != nil {
		result += s.Margin.Vertical()
	}
	if s.Padding != nil {
		result += s.Padding.Vertical()
	}
	if s.Border != "" {
		result += 2
	}
	return result
}

// Info returns a human-readable multi-line string representation of the style.
// This method is primarily used for debugging and development purposes,
// providing a comprehensive overview of all style properties.
//
// Returns:
//   - string: Formatted multi-line string with all style properties
//
// The output includes:
//   - Background and foreground colors
//   - Border and cursor styles
//   - Detailed margin and padding information
//   - Preferred size hints
//   - Rendering mode
func (s *Style) Info() string {
	return fmt.Sprintf("Background: %s\nForeground: %s\nBorder    : %s\nCursor    : %s\nMargin    : %s\nPadding   : %s\nPref. Size: %d x %d\nRender    : %s", s.Background, s.Foreground, s.Border, s.Cursor, s.Margin.Info(), s.Padding.Info(), s.Width, s.Height, s.Render)
}

// SetBackground sets the background color for the style and returns the style
// for method chaining. The background color is applied to the widget's content
// area and padding area, but not to margins.
//
// Parameters:
//   - background: Color name, hex code, or empty string for transparent
//
// Returns:
//   - *Style: The style instance for method chaining
//
// Example usage:
//
//	style.SetBackground("blue").SetForeground("white")
func (s *Style) SetBackground(background string) *Style {
	s.Background = background
	return s
}

// SetForeground sets the foreground (text) color for the style and returns
// the style for method chaining. The foreground color is used for text
// and other drawable content within the widget.
//
// Parameters:
//   - foreground: Color name, hex code, or empty string for default
//
// Returns:
//   - *Style: The style instance for method chaining
//
// Example usage:
//
//	style.SetForeground("#ffffff").SetBackground("#000000")
func (s *Style) SetForeground(foreground string) *Style {
	s.Foreground = foreground
	return s
}

// SetBorder sets the border style for the widget and returns the style
// for method chaining. The border is drawn around the padding area
// and affects the total space calculations.
//
// Parameters:
//   - border: Border style identifier or empty string for no border
//
// Returns:
//   - *Style: The style instance for method chaining
//
// Common border styles may include "solid", "dashed", "double", etc.
// The exact styles depend on the rendering implementation.
func (s *Style) SetBorder(border string) *Style {
	s.Border = border
	return s
}

// SetCursor sets the cursor style for interactive widgets and returns
// the style for method chaining. This affects how the cursor appears
// when the widget has focus.
//
// Parameters:
//   - cursor: Cursor style identifier or empty string for default
//
// Returns:
//   - *Style: The style instance for method chaining
//
// Example usage:
//
//	style.SetCursor("block").SetForeground("white")
func (s *Style) SetCursor(cursor string) *Style {
	s.Cursor = cursor
	return s
}

// SetFont sets the font style ad returns the style for method chaining.
// Font support highly depends on the terminal
//
// The font support is directly mapped to the tcell style. The following
// styles are supported at the moment:
//   - bold
//   - italic
//   - underline
//   - strikethrough
//
// Styles can be combined separated by comma, the underline style may be
// extended by double, curly, dotted or dashed and/or a color value.
//
// Parameters:
//   - font: Font style
//
// Returns:
//   - *Style: The style instance for method chaining
func (s *Style) SetFont(font string) *Style {
	s.Font = font
	return s
}

// SetRender sets a special rendering mode hint for the widget and returns
// the style for method chaining. This can be used to enable special
// rendering behaviors or optimizations.
//
// Parameters:
//   - render: Rendering mode identifier or empty string for default
//
// Returns:
//   - *Style: The style instance for method chaining
//
// The specific rendering modes depend on the widget implementation
// and rendering system capabilities.
func (s *Style) SetRender(render string) *Style {
	s.Render = render
	return s
}

// SetMargin sets the margin spacing around the widget and returns the style
// for method chaining. Margins create transparent space outside the border.
// The method accepts 1-4 values following CSS margin conventions.
//
// Value patterns:
//   - 1 value: all sides
//   - 2 values: vertical, horizontal
//   - 3 values: top, horizontal, bottom
//   - 4 values: top, right, bottom, left
//
// Parameters:
//   - values: Variable number of margin values
//
// Returns:
//   - *Style: The style instance for method chaining
//
// Example usage:
//
//	style.SetMargin(2)           // 2 on all sides
//	style.SetMargin(1, 2)        // 1 vertical, 2 horizontal
//	style.SetMargin(1, 2, 3, 4)  // top, right, bottom, left
func (s *Style) SetMargin(values ...int) *Style {
	if s.Margin == nil {
		s.Margin = &Insets{}
	}
	s.Margin.Set(values...)
	return s
}

// SetPadding sets the padding spacing inside the widget and returns the style
// for method chaining. Padding creates space between the border and content,
// filled with the background color. Accepts 1-4 values following CSS conventions.
//
// Value patterns:
//   - 1 value: all sides
//   - 2 values: vertical, horizontal
//   - 3 values: top, horizontal, bottom
//   - 4 values: top, right, bottom, left
//
// Parameters:
//   - values: Variable number of padding values
//
// Returns:
//   - *Style: The style instance for method chaining
//
// Example usage:
//
//	style.SetPadding(1)           // 1 on all sides
//	style.SetPadding(2, 4)        // 2 vertical, 4 horizontal
//	style.SetPadding(1, 2, 3, 4)  // top, right, bottom, left
func (s *Style) SetPadding(values ...int) *Style {
	if s.Padding == nil {
		s.Padding = &Insets{}
	}
	s.Padding.Set(values...)
	return s
}

// SetSize sets the preferred width and height hints for layout algorithms
// and returns the style for method chaining. These are suggestions to the
// layout system and may not be the final rendered size.
//
// Parameters:
//   - w: Preferred width in characters/pixels
//   - h: Preferred height in characters/pixels
//
// Returns:
//   - *Style: The style instance for method chaining
//
// Size hints are used by:
//   - Layout containers to determine space allocation
//   - Flexible layouts for initial size calculations
//   - Auto-sizing algorithms for preferred dimensions
//
// Example usage:
//
//	style.SetSize(80, 24)  // Prefer 80x24 character size
func (s *Style) SetSize(w, h int) *Style {
	s.Width = w
	s.Height = h
	return s
}

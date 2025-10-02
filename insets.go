package zeichenwerk

import "fmt"

// Insets represents spacing or padding values for all four sides of a rectangular area.
// It follows the CSS box model convention where values are specified in clockwise order:
// Top, Right, Bottom, Left.
//
// Insets are commonly used for:
//   - Widget padding (inner spacing)
//   - Widget margins (outer spacing)
//   - Border spacing and layout calculations
//   - Content area calculations within styled widgets
//
// The struct provides CSS-style shorthand methods for convenient configuration
// and utility methods for common layout calculations.
type Insets struct {
	Top, Right, Bottom, Left int // Spacing values for each side (clockwise from top)
}

// NewInsets creates a new Insets instance using CSS-style shorthand notation.
// This is a convenience constructor that uses the same value interpretation
// as the Set() method.
//
// Parameters:
//   - values: Variable number of integer values for inset configuration
//
// Returns:
//   - Insets: A new Insets instance with the specified values
//
// Examples:
//
//	NewInsets()           // All sides = 0
//	NewInsets(5)          // All sides = 5
//	NewInsets(10, 20)     // Top/Bottom = 10, Left/Right = 20
//	NewInsets(1, 2, 3, 4) // Top = 1, Right = 2, Bottom = 3, Left = 4
func NewInsets(values ...int) *Insets {
	insets := Insets{}
	insets.Set(values...)
	return &insets
}

// Info returns a human-readable string representation of the insets.
// The format follows the CSS convention: "(top right bottom left)".
// This is useful for debugging and logging inset values.
//
// Returns:
//   - string: Formatted string showing all four inset values
func (i *Insets) Info() string {
	return fmt.Sprintf("(%d %d %d %d)", i.Top, i.Right, i.Bottom, i.Left)
}

// Set configures the insets using CSS-style shorthand notation.
// This method follows the same conventions as CSS padding/margin properties,
// allowing for flexible specification of inset values.
//
// Value interpretation based on count:
//   - 0 values: All sides set to 0
//   - 1 value:  All sides set to the same value (uniform insets)
//   - 2 values: Top/Bottom = first value, Left/Right = second value
//   - 3 values: Top = first, Left/Right = second, Bottom = third
//   - 4+ values: Top, Right, Bottom, Left (clockwise from top)
//
// Parameters:
//   - values: Variable number of integer values for inset configuration
//
// Examples:
//
//	insets.Set()           // All sides = 0
//	insets.Set(5)          // All sides = 5
//	insets.Set(10, 20)     // Top/Bottom = 10, Left/Right = 20
//	insets.Set(1, 2, 3)    // Top = 1, Left/Right = 2, Bottom = 3
//	insets.Set(1, 2, 3, 4) // Top = 1, Right = 2, Bottom = 3, Left = 4
func (i *Insets) Set(values ...int) {
	switch len(values) {
	case 0:
		// All sides = 0
		i.Top = 0
		i.Right = 0
		i.Bottom = 0
		i.Left = 0

	case 1:
		// All sides = same value
		i.Top = values[0]
		i.Right = values[0]
		i.Bottom = values[0]
		i.Left = values[0]

	case 2:
		// Top/Bottom = first, Left/Right = second
		i.Top = values[0]
		i.Right = values[1]
		i.Bottom = values[0]
		i.Left = values[1]

	case 3:
		// Top = first, Left/Right = second, Bottom = third
		i.Top = values[0]
		i.Right = values[1]
		i.Bottom = values[2]
		i.Left = values[1]

	default:
		// Top, Right, Bottom, Left (clockwise from top)
		i.Top = values[0]
		i.Right = values[1]
		i.Bottom = values[2]
		i.Left = values[3]
	}
}

// Horizontal returns the total horizontal spacing (left + right).
// This is useful for calculating the total width consumed by horizontal insets.
//
// Returns:
//   - int: The sum of left and right inset values
func (i *Insets) Horizontal() int {
	return i.Left + i.Right
}

// Vertical returns the total vertical spacing (top + bottom).
// This is useful for calculating the total height consumed by vertical insets.
//
// Returns:
//   - int: The sum of top and bottom inset values
func (i *Insets) Vertical() int {
	return i.Top + i.Bottom
}

// Total returns the total spacing for both dimensions.
// This provides the total width and height consumed by the insets.
//
// Returns:
//   - int: Total horizontal spacing (left + right)
//   - int: Total vertical spacing (top + bottom)
func (i *Insets) Total() (int, int) {
	return i.Horizontal(), i.Vertical()
}

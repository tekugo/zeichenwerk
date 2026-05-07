package core

import (
	"fmt"
	"strconv"
	"strings"
)

// NoInsets is a shared, zero-valued Insets instance used as the canonical
// "no spacing" sentinel. It is returned by Style accessors when no margin
// or padding has been set anywhere in the cascade. Callers must treat it
// as read-only: mutating it through the pointer would affect every widget
// that relies on the sentinel.
var NoInsets = &Insets{}

// Insets represents spacing or padding values for all four sides of a rectangular area.
// It follows the CSS box model convention where values are specified in clockwise order:
// Top, Right, Bottom, Left.
//
// The struct provides CSS-style shorthand methods for convenient configuration
// and utility methods for common layout calculations.
type Insets struct {
	Top, Right, Bottom, Left int // Spacing values for each side (clockwise from top)
}

// NewInsets creates a new Insets instance using CSS-style shorthand notation.
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

// String reverses the CSS-shorthand expansion so the shortest
// equivalent variadic argument list is emitted (matches WithPadding /
// WithMargin's input expectations).
func (i *Insets) String(separator ...string) string {
	if i == nil {
		return ""
	}

	sep := " "
	if len(separator) > 0 {
		sep = separator[0]
	}

	switch {
	case i.Top == i.Right && i.Right == i.Bottom && i.Bottom == i.Left:
		return fmt.Sprintf("%d", i.Top)
	case i.Top == i.Bottom && i.Right == i.Left:
		return fmt.Sprintf("%d%s%d", i.Top, sep, i.Right)
	case i.Right == i.Left:
		return fmt.Sprintf("%d%s%d%s%d", i.Top, sep, i.Right, sep, i.Bottom)
	default:
		return fmt.Sprintf("%d%s%d%s%d%s%d", i.Top, sep, i.Right, sep, i.Bottom, sep, i.Left)
	}
}

// Parse parses a CSS-shorthand inset string keyed Top, Right, Bottom, Left.
// Whitespace and commas are both accepted as separators, so "1 2", "1, 2",
// "1,2", and "  1 ,  2 all yield the same result.
//
// Element-count semantics:
//
//	0 ("")           -> [0, 0, 0, 0]               (NoInsets)
//	1 ("a")          -> [a, a, a, a]
//	2 ("a b")        -> [a, b, a, b]               (top/bottom, left/right)
//	3 ("a b c")      -> [a, b, c, b]               (top, left/right, bottom)
//	4 ("a b c d")    -> [a, b, c, d]               (top, right, bottom, left)
//
// Returns ok=false on parse failure or unsupported element counts;
// callers (Store, Emit) keep the existing inset rather than
// applying a half-typed mid-edit value.
func (i *Insets) Parse(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		i.Set(0)
		return true
	}

	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	if len(fields) > 4 {
		return false
	}

	nums := make([]int, len(fields))
	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil {
			return false
		}
		nums[i] = n
	}
	i.Set(nums...)
	return true
}

func (i *Insets) IsZero() bool {
	return i == nil || (i.Top == 0 && i.Right == 0 && i.Bottom == 0 && i.Left == 0)
}

func (i *Insets) Array() [4]int {
	if i == nil {
		return [4]int{}
	} else {
		return [4]int{i.Top, i.Right, i.Bottom, i.Left}
	}
}

package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Flex is a container widget that arranges child widgets in a linear layout,
// either horizontally or vertically. It provides flexible sizing and alignment
// options for creating responsive and organized user interfaces.
//
// Features:
//   - Horizontal and vertical layout orientations
//   - Flexible sizing with fractional units (negative values)
//   - Multiple alignment options (start, center, end, stretch)
//   - Configurable spacing between child widgets
//   - Automatic size calculation based on child hints
//
// Layout behavior:
//   - Fixed sizes: Positive width/height values are treated as absolute sizes
//   - Auto sizes: Zero values use the widget's preferred size hint
//   - Flexible sizes: Negative values are treated as fractional units of remaining space
//   - Spacing: Applied between adjacent children (not at edges)
type Flex struct {
	BaseWidget
	children    []Widget // Child widgets managed by this flex container
	Orientation string   // Layout direction: "horizontal" or "vertical"
	Alignment   string   // Child alignment: "start", "center", "end", or "stretch"
	Spacing     int      // Pixels/characters of spacing between children
}

// NewFlex creates a new flex container widget with the specified configuration.
// The flex container will arrange its child widgets according to the provided
// orientation, alignment, and spacing parameters.
//
// Parameters:
//   - id: Unique identifier for the flex container
//   - orientation: Layout direction ("horizontal" or "vertical")
//   - alignment: How children are aligned ("start", "center", "end", "stretch")
//   - spacing: Number of pixels/characters between adjacent children
//
// Returns:
//   - *Flex: A new flex container widget instance
//
// Example usage:
//
//	// Create a horizontal flex with center alignment and 2-pixel spacing
//	flex := NewFlex("main-flex", "horizontal", "center", 2)
//
//	// Create a vertical flex with stretch alignment
//	vflex := NewFlex("sidebar", "vertical", "stretch", 1)
func NewFlex(id string, orientation, alignment string, spacing int) *Flex {
	return &Flex{
		BaseWidget:  BaseWidget{id: id},
		Orientation: orientation,
		Alignment:   alignment,
		Spacing:     spacing,
	}
}

// Add appends a new child widget to the flex container.
// The widget will be positioned according to the flex's orientation and alignment
// settings. The layout will need to be recalculated after adding widgets.
//
// Parameters:
//   - widget: The widget to add as a child of this flex container
func (f *Flex) Add(widget Widget) {
	f.children = append(f.children, widget)
}

// Children returns a slice of all child widgets in this flex container.
// The widgets are returned in the order they were added, which determines
// their layout order within the flex container.
//
// Returns:
//   - []Widget: A slice containing all child widgets
func (f *Flex) Children() []Widget {
	return f.children
}

// Find searches for a widget with the specified ID within this flex container
// and all of its descendant widgets. Uses the standard container search algorithm.
//
// Parameters:
//   - id: The unique identifier of the widget to find
//
// Returns:
//   - Widget: The widget with the matching ID, or nil if not found
func (f *Flex) Find(id string) Widget {
	return Find(f, id)
}

// FindAt searches for the widget located at the specified coordinates within
// this flex container and its child widgets. Uses the standard container
// coordinate-based search algorithm.
//
// Parameters:
//   - x: The x-coordinate to search at
//   - y: The y-coordinate to search at
//
// Returns:
//   - Widget: The widget at the specified coordinates, or nil if not found
func (f *Flex) FindAt(x, y int) Widget {
	return FindAt(f, x, y)
}

// Handle processes events for the flex container.
// The base implementation doesn't handle any events directly, as flex containers
// typically delegate event handling to their child widgets.
//
// Parameters:
//   - event: The tcell.Event to process
//
// Returns:
//   - bool: Always returns false as flex containers don't handle events directly
func (f *Flex) Handle(event tcell.Event) bool {
	return false
}

// Hint calculates and returns the preferred size for the flex container
// based on its children's size hints and the layout orientation.
//
// Size calculation:
//   - Horizontal layout: Width = sum of child widths + spacing, Height = max child height
//   - Vertical layout: Width = max child width, Height = sum of child heights + spacing
//   - Includes child margins, padding, and borders in calculations
//   - Accounts for spacing between children (but not at edges)
//
// Returns:
//   - int: Preferred width in characters/pixels
//   - int: Preferred height in characters/pixels
func (f *Flex) Hint() (int, int) {
	var width, height int

	if f.Orientation == "horizontal" {
		// Horizontal layout: sum widths, take max height
		for i, child := range f.children {
			cw, ch := child.Hint()
			cw += child.Style("").Horizontal()
			ch += child.Style("").Vertical()
			width += cw
			if i > 0 { // Add spacing between children (not before first)
				width += f.Spacing
			}
			if ch > height {
				height = ch
			}
		}
	} else {
		// Vertical layout: take max width, sum heights
		for i, child := range f.children {
			cw, ch := child.Hint()
			cw += child.Style("").Horizontal()
			ch += child.Style("").Vertical()
			height += ch
			if i > 0 { // Add spacing between children (not before first)
				height += f.Spacing
			}
			if cw > width {
				width = cw
			}
			f.Log("  Hint %s, width=%d, height=%d, total width=%d, height=%d", child.ID(), cw, ch, width, height)
		}
	}
	return width, height
}

// Info returns a human-readable description of the flex container's current state.
// This includes position, dimensions, content area, and container type information.
// Primarily used for debugging and development purposes.
//
// Returns:
//   - string: Formatted string with flex container information
func (f *Flex) Info() string {
	x, y, w, h := f.Bounds()
	cx, cy, cw, ch := f.Content()
	return fmt.Sprintf("@%d.%d %d:%d (%d.%d %d:%d) flex[%s,%s,%d]",
		x, y, w, h, cx, cy, cw, ch, f.Orientation, f.Alignment, f.Spacing)
}

// Layout arranges all child widgets within the flex container according to
// the specified orientation, alignment, and spacing configuration.
//
// Layout process:
//   - Calculates available space within the content area
//   - Distributes space among children based on their size hints and flex properties
//   - Positions children according to orientation (horizontal/vertical)
//   - Applies alignment settings for the cross-axis
//   - Handles flexible sizing with fractional units
//   - Recursively triggers layout on child containers
//
// The layout supports three sizing modes for children:
//   - Fixed: Positive values are treated as absolute sizes
//   - Auto: Zero values use the widget's preferred size hint
//   - Flexible: Negative values are fractional units of remaining space
func (f *Flex) Layout() {
	if f.Orientation == "horizontal" {
		f.layoutH()
	} else if f.Orientation == "vertical" {
		f.layoutV()
	}

	// Refresh and lay out children
	Layout(f)
}

// layoutH performs horizontal layout of child widgets.
// This method arranges children from left to right, calculating widths
// based on style hints and distributing remaining space among flexible children.
//
// Layout algorithm:
//  1. Calculate total fractions and remaining space after fixed-size children
//  2. Determine fraction size (remaining space / total fractions)
//  3. Position each child with calculated width and full container height
//  4. Apply spacing between adjacent children
//
// Width calculation:
//   - Fixed width (positive): Use exact value plus horizontal margins/padding
//   - Flexible width (negative): Use fractional units of remaining space
//   - Last flexible child gets any remaining pixels to avoid rounding errors
func (f *Flex) layoutH() {
	var style *Style
	var fraction int

	cx, cy, cw, ch := f.Content()

	// Calculate fractional size for 1 fraction unit
	fractions := 0
	last := -1
	rest := cw
	for i, child := range f.children {
		style = child.Style("")
		f.Log("  %T %s %d:%d", child, child.ID(), style.Width, style.Height)
		if style.Width < 0 {
			fractions -= style.Width // Accumulate fraction units (negative values)
			last = i
		} else if style.Width == 0 {
			pw, _ := child.Hint()
			rest = rest - pw - style.Horizontal()
		} else {
			rest = rest - style.Width - style.Horizontal()
		}
		if i > 0 {
			rest -= f.Spacing
		}
	}

	// Calculate space per fraction unit
	if fractions > 0 {
		fraction = rest / fractions
	}

	f.Log("Flex H %s children= %d, fractions=%d, fraction=%d, rest=%d", f.id, len(f.children), fractions, fraction, rest)

	// Position children from left to right
	x := cx
	for i, child := range f.children {
		style = child.Style("")

		// Calculate x position and width according to alignment
		wy, wh := align(f.Alignment, cy, cy+ch, style.Height+style.Vertical())
		if style.Width < 0 {
			if i != last {
				child.SetBounds(x, wy, -fraction*style.Width, wh)
			} else {
				// Last flexible child gets remaining space to handle rounding
				child.SetBounds(x, wy, rest, wh)
			}
			rest = rest - fraction*style.Width
		} else if style.Width == 0 {
			pw, _ := child.Hint()
			child.SetBounds(x, wy, pw+style.Horizontal(), wh)
		} else {
			child.SetBounds(x, wy, style.Width+style.Horizontal(), wh)
		}
		_, _, width, _ := child.Bounds()
		x += width + f.Spacing
	}
}

// layoutV performs vertical layout of child widgets.
// This method arranges children from top to bottom, calculating heights
// based on style hints and distributing remaining space among flexible children.
// It also applies horizontal alignment for each child.
//
// Layout algorithm:
//  1. Calculate total fractions and remaining space after fixed-size children
//  2. Determine fraction size (remaining space / total fractions)
//  3. Position each child with calculated height and aligned width
//  4. Apply spacing between adjacent children
//
// Height calculation:
//   - Fixed height (positive): Use exact value plus vertical margins/padding
//   - Auto height (zero): Use widget's preferred height hint
//   - Flexible height (negative): Use fractional units of remaining space
//   - Last flexible child gets any remaining pixels to avoid rounding errors
func (f *Flex) layoutV() {
	var style *Style
	var fraction int

	cx, cy, cw, ch := f.Content()

	// Calculate fractional size for 1 fraction unit
	fractions := 0
	last := -1
	rest := ch
	for i, child := range f.children {
		style = child.Style("")
		if style.Height < 0 {
			fractions -= style.Height // Accumulate fraction units (negative values)
			last = i
		} else if style.Height == 0 {
			_, ph := child.Hint()
			rest = rest - ph - style.Vertical()
		} else {
			rest = rest - style.Height - style.Vertical()
		}
		if i > 0 {
			rest -= f.Spacing
		}
	}

	// Calculate space per fraction unit
	if fractions > 0 {
		fraction = rest / fractions
	}

	// Position children from top to bottom
	y := cy
	for i, child := range f.children {
		style = child.Style("")

		// Calculate x position and width according to alignment
		wx, ww := align(f.Alignment, cx, cx+cw, style.Width+style.Horizontal())
		if style.Height < 0 {
			if i != last {
				child.SetBounds(wx, y, ww, -fraction*style.Height)
			} else {
				// Last flexible child gets remaining space to handle rounding
				child.SetBounds(wx, y, ww, rest)
			}
			rest = rest - fraction*style.Height
		} else if style.Height == 0 {
			_, ph := child.Hint()
			child.SetBounds(wx, y, ww, ph+style.Vertical())
		} else {
			child.SetBounds(wx, y, ww, style.Height+style.Vertical())
		}
		_, _, _, height := child.Bounds()
		y += height + f.Spacing
	}
}

// align calculates the position and size for a widget within a container
// based on the specified alignment mode. This function is used for cross-axis
// alignment in flex layouts (horizontal alignment in vertical flex, etc.).
//
// Alignment modes:
//   - "start": Align to the beginning of the available space
//   - "center": Center within the available space
//   - "end": Align to the end of the available space
//   - "stretch": Expand to fill the entire available space
//   - Default: Same as "start"
//
// Parameters:
//   - alignment: The alignment mode string
//   - start: The starting coordinate of the available space
//   - end: The ending coordinate of the available space
//   - size: The preferred size of the widget (ignored for "stretch")
//
// Returns:
//   - int: The calculated position coordinate
//   - int: The calculated size (may be modified for "stretch" alignment)
func align(alignment string, start, end, size int) (int, int) {
	space := end - start

	switch alignment {
	case "center":
		if space >= size {
			return start + (space-size)/2, size
		}
		return start + space/2, size
	case "end":
		return end - size, size
	case "stretch":
		return start, space
	default: // "start" or any other value
		return start, size
	}
}

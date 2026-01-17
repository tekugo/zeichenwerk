package next

// Flex is a container widget that arranges child widgets in a linear layout,
// either horizontally or vertically. It provides flexible sizing and alignment
// options for creating responsive and organized user interfaces.
//
// Layout behavior:
//   - Fixed sizes: Positive width/height values are treated as absolute sizes
//   - Auto sizes: Zero values use the widget's preferred size hint
//   - Flexible sizes: Negative values are treated as fractional units of remaining space
//   - Spacing: Applied between adjacent children (not at edges)
type Flex struct {
	Component
	children   []Widget // Child widgets managed by this flex container
	horizontal bool     // Whether the flex is horizontal or vertical
	alignment  string   // Child alignment: "start", "center", "end", or "stretch"
	spacing    int      // Pixels/characters of spacing between children
}

// NewFlex creates a new flex container widget with the specified configuration.
// The flex container will arrange its child widgets according to the provided
// orientation, alignment, and spacing parameters.
//
// Parameters:
//   - id: Unique identifier for the flex container
//   - horizontal: Whether the flex is horizontal or vertical
//   - alignment: How children are aligned ("start", "center", "end", "stretch")
//   - spacing: Number of pixels/characters between adjacent children
//
// Returns:
//   - *Flex: A new flex container widget instance
//
// Example usage:
//
//	// Create a horizontal flex with center alignment and 2-pixel spacing
//	flex := NewFlex("main-flex", true, "center", 2)
//
//	// Create a vertical flex with stretch alignment
//	vflex := NewFlex("sidebar", false, "stretch", 1)
func NewFlex(id string, horizontal bool, alignment string, spacing int) *Flex {
	return &Flex{
		Component:  Component{id: id},
		horizontal: horizontal,
		alignment:  alignment,
		spacing:    spacing,
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

	if f.horizontal {
		// Horizontal layout: sum widths, take max height
		for i, child := range f.children {
			cw, ch := child.Hint()
			cw += child.Style().Horizontal()
			ch += child.Style().Vertical()
			width += cw
			if i > 0 { // Add spacing between children (not before first)
				width += f.spacing
			}
			if ch > height {
				height = ch
			}
		}
	} else {
		// Vertical layout: take max width, sum heights
		for i, child := range f.children {
			cw, ch := child.Hint()
			cw += child.Style().Horizontal()
			ch += child.Style().Vertical()
			height += ch
			if i > 0 { // Add spacing between children (not before first)
				height += f.spacing
			}
			if cw > width {
				width = cw
			}
		}
	}
	return width, height
}

// Layout arranges all child widgets within the flex container according to
// the specified orientation, alignment, and spacing configuration.
func (f *Flex) Layout() {
	if f.horizontal {
		f.Log(f, "debug", "Flex horizontal layout %s @%d,%d %dx%d", f.id, f.x, f.y, f.width, f.height)
		f.layoutHorizontal()
	} else {
		f.Log(f, "debug", "Flex vertical layout %s @%d,%d %dx%d", f.id, f.x, f.y, f.width, f.height)
		f.layoutVertical()
	}

	// Refresh and lay out children
	Layout(f)
}

// layoutHorizontal performs horizontal layout of child widgets.
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
func (f *Flex) layoutHorizontal() {
	var style *Style
	var fraction int

	cx, cy, cw, ch := f.Content()

	// Calculate fractional size for 1 fraction unit
	fractions := 0
	last := -1
	rest := cw
	for i, child := range f.children {
		hw, _ := child.Hint()
		style = child.Style()

		if hw < 0 {
			fractions -= hw // Accumulate fraction units (negative values)
			last = i
		} else {
			rest = rest - hw - style.Horizontal()
		}
		if i > 0 {
			rest -= f.spacing
		}
	}

	// Calculate space per fraction unit
	if fractions > 0 {
		fraction = rest / fractions
	}

	// Position children from left to right
	x := cx
	for i, child := range f.children {
		hw, hh := child.Hint()
		style = child.Style()

		// Calculate x position and width according to alignment
		wy, wh := align(f.alignment, cy, cy+ch, hh+style.Vertical())
		if hw < 0 {
			if i != last {
				child.SetBounds(x, wy, -fraction*hw, wh)
			} else {
				// Last flexible child gets remaining space to handle rounding
				child.SetBounds(x, wy, rest, wh)
			}
			rest = rest - fraction*(-hw)
		} else {
			child.SetBounds(x, wy, hw+style.Horizontal(), wh)
		}
		_, _, width, _ := child.Bounds()
		x += width + f.spacing
	}
}

// layoutVertical performs vertical layout of child widgets.
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
func (f *Flex) layoutVertical() {
	var style *Style
	var fraction int

	cx, cy, cw, ch := f.Content()

	// Calculate fractional size for 1 fraction unit
	fractions := 0
	last := -1
	rest := ch
	for i, child := range f.children {
		_, hh := child.Hint()
		style = child.Style()

		if hh < 0 {
			fractions -= hh // Accumulate fraction units (negative values)
			last = i
		} else {
			rest = rest - hh - style.Vertical()
		}
		if i > 0 {
			rest -= f.spacing
		}
	}

	// Calculate space per fraction unit
	if fractions > 0 {
		fraction = rest / fractions
	}

	// Position children from top to bottom
	y := cy
	for i, child := range f.children {
		hw, hh := child.Hint()
		style = child.Style()

		// Calculate x position and width according to alignment
		wx, ww := align(f.alignment, cx, cx+cw, hw+style.Horizontal())
		if hh < 0 {
			if i != last {
				child.SetBounds(wx, y, ww, -fraction*hh)
				rest = rest - fraction*(-hh)
			} else {
				// Last flexible child gets remaining space to handle rounding
				child.SetBounds(wx, y, ww, rest)
			}
		} else {
			child.SetBounds(wx, y, ww, hh+style.Vertical())
		}
		_, _, _, height := child.Bounds()
		f.Log(f, "debug", "  %d %s hint=%d height=%d", i, child.ID(), hh, height)
		y += height + f.spacing
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

// Render draws the flex container and all its children.
func (f *Flex) Render(r *Renderer) {
	f.Component.Render(r)
	for _, child := range f.children {
		child.Render(r)
	}
}

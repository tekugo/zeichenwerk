package next

import "unicode/utf8"

// Border represents a border with all possible elements.
type Border struct {
	// Outer border elements - form the perimeter of widgets
	Top    string // Horizontal top border
	Right  string // Vertical right border
	Bottom string // Horizontal bottom border
	Left   string // Vertical left border

	// Corner elements - connect perpendicular border segments
	TopLeft     string // Top-left corner
	TopRight    string // Top-right corner
	BottomRight string // Bottom-right corner
	BottomLeft  string // Bottom-left corner

	// Outer T-connectors - join borders at T-junctions on the perimeter
	TopT    string // Top T-junction
	RightT  string // Right T-junction
	BottomT string // Bottom T-junction
	LeftT   string // Left T-junction

	// Inner grid elements - divide the widget into cells
	InnerH string // Horizontal inner grid element
	InnerV string // Vertical inner grid element
	InnerX string // Cross inner grid element

	// Inner T-connectors - join inner grid elements
	InnerTopT    string // Top T-junction
	InnerRightT  string // Right T-junction
	InnerBottomT string // Bottom T-junction
	InnerLeftT   string // Left T-junction
}

// Horizontal returns the width of the border.
func (b *Border) Horizontal() int {
	return utf8.RuneCountInString(b.Left) + utf8.RuneCountInString(b.Right)
}

// Vertical returns the height of the border.
func (b *Border) Vertical() int {
	return min(len(b.Top), 1) + min(len(b.Bottom), 1)
}

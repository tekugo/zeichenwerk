package zeichenwerk

// Box represents a container widget that wraps a single child widget with an optional title.
// It provides a bordered container that can display a title and contains exactly one child widget.
// The box automatically handles layout by fitting the child widget within its content area.
//
// Features:
//   - Single child widget containment
//   - Optional title display
//   - Automatic border and padding management
//   - Focus and hover state support
//   - Recursive widget search capabilities
//
// The Box widget is useful for grouping related UI elements, providing visual separation,
// and adding descriptive titles to sections of the interface. It serves as a fundamental
// building block for creating organized and visually structured layouts.
//
// Layout behavior:
//   - The child widget is positioned within the box's content area
//   - Content area excludes borders, margins, and padding
//   - The box's preferred size includes both its styling overhead and child requirements
type Box struct {
	BaseWidget
	Title string // The title text displayed in the box header (optional)
	child Widget // The single child widget contained within this box
}

// NewBox creates a new box container widget with the specified ID and title.
// The box is initialized as not focusable and ready to contain a single child widget.
// The title parameter can be an empty string if no title is desired.
//
// Parameters:
//   - id: Unique identifier for the box widget
//   - title: Optional title text to display in the box header
//
// Returns:
//   - *Box: A new box widget instance ready to contain a child widget
func NewBox(id, title string) *Box {
	return &Box{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Title:      title,
	}
}

// Add sets the child widget for this box container.
// Since Box is designed to contain exactly one child widget, calling Add
// will replace any previously set child widget. If you need to contain
// multiple widgets, consider using a layout container (like Flex or Grid)
// as the child widget.
//
// Parameters:
//   - widget: The widget to be contained within this box
func (b *Box) Add(widget Widget) {
	b.child = widget
}

// Children returns a slice containing the single child widget of this box.
// This method implements the Container interface requirement. The returned
// slice will contain exactly one element if a child has been added, or will
// be empty if no child widget has been set.
//
// Returns:
//   - []Widget: A slice containing the child widget, or empty slice if no child is set
//
// Note: The returned slice should not be modified directly. Use the Add method
// to set or change the child widget.
func (b *Box) Children(_ bool) []Widget {
	if b.child == nil {
		return []Widget{}
	}
	return []Widget{b.child}
}

// Find searches for a widget with the specified ID within this box and its child widget.
// The search is performed recursively, first checking if this box matches the ID,
// then delegating to the generic Find function which will search the child widget
// and any of its descendants if the child is also a container.
//
// Parameters:
//   - id: The unique identifier of the widget to find
//   - visible: Only look for visible children
//
// Returns:
//   - Widget: The widget with the matching ID, or nil if not found
//
// The search order is:
//  1. Check if this box's ID matches
//  2. Recursively search within the child widget (if it exists)
//  3. If child is a container, search its descendants
func (b *Box) Find(id string, visible bool) Widget {
	return Find(b, id, visible)
}

// FindAt searches for the widget located at the specified screen coordinates.
// This method is used for mouse interaction to determine which widget is
// positioned at a given point. The search includes this box and its child widget.
//
// Parameters:
//   - x: The x-coordinate to search at
//   - y: The y-coordinate to search at
//
// Returns:
//   - Widget: The widget at the specified coordinates, or nil if none found
//
// The search process:
//  1. Check if coordinates are within this box's bounds
//  2. If within bounds, recursively search the child widget
//  3. Return the most specific widget found at the coordinates
func (b *Box) FindAt(x, y int) Widget {
	return FindAt(b, x, y)
}

// Hint returns the box's preferred size for optimal display.
// The calculation includes the box's styling overhead (borders, padding, margins)
// plus the preferred size of its child widget. This ensures the box can
// accommodate both its visual styling and its contained content.
//
// Returns:
//   - int: Preferred width including styling overhead and child requirements
//   - int: Preferred height including styling overhead and child requirements
//
// Size calculation:
//  1. Start with the box's horizontal and vertical styling overhead
//  2. Add the child widget's preferred dimensions (if child exists)
//  3. Return the total required space for optimal display
func (b *Box) Hint() (int, int) {
	if b.child != nil {
		w, h := b.child.Hint()
		w += b.child.Style("").Horizontal()
		h += b.child.Style("").Vertical()
		return w, h
	} else {
		return 0, 0
	}
}

// Info returns an information string about the widget.
// This is mainly used for debugging purposes.
func (b *Box) Info() string {
	return "box [" + b.BaseWidget.Info() + "]"
}

// Layout positions and sizes the child widget within this box's content area.
// The child widget is given the full content area of the box, which excludes
// the space used by borders, padding, and margins. After positioning the child,
// the generic Layout function is called to handle any additional layout requirements.
//
// Layout process:
//  1. Calculate the content area (excluding styling overhead)
//  2. Position the child widget to fill the entire content area
//  3. Delegate to the generic Layout function for final processing
//
// The child widget will be positioned at the content area's top-left corner
// and sized to fill the available content space completely.
func (b *Box) Layout() {
	if b.child != nil {
		cx, cy, cw, ch := b.Content()
		b.child.SetBounds(cx, cy, cw, ch)
	}
	Layout(b)
}

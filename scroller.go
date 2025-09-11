package zeichenwerk

type Scroller struct {
	BaseWidget
	Title  string // Optional box title
	child  Widget // Child
	tx, ty int    // current offset / translation for viewport
}

// NewScroller creates a new scroll pane widget with the specified ID and title.
// The scroller is initialized as focusable and ready to contain a single child widget.
// The title parameter can be an empty string if no title is desired.
//
// Parameters:
//   - id: Unique identifier for the box widget
//   - title: Optional title text to display in the box header
//
// Returns:
//   - *Box: A new box widget instance ready to contain a child widget
func NewScroller(id, title string) *Scroller {
	return &Scroller{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Title:      title,
	}
}

// Add sets the child widget for this scroll pane.
// Since Box is designed to contain exactly one child widget, calling Add
// will replace any previously set child widget. If you need to contain
// multiple widgets, consider wrapping it in a layout container (like Flex
// or Grid) as the child widget.
//
// Parameters:
//   - widget: The widget to be contained within this scroll pane
func (s *Scroller) Add(widget Widget) {
	s.child = widget
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
func (s *Scroller) Children(_ bool) []Widget {
	if s.child == nil {
		return []Widget{}
	}
	return []Widget{s.child}
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
func (s *Scroller) Find(id string, visible bool) Widget {
	return Find(s, id, visible)
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
func (s *Scroller) FindAt(x, y int) Widget {
	return FindAt(s, x, y)
}

func (s *Scroller) Info() string {
	return "scroller [" + s.BaseWidget.Info() + "]"
}

// Layout positions and sizes the child widget within the content area.
// The child widget is always given its preferred width and height, as the
// scroller just show the part of the widget, which is in its content area.
// After positioning the child, the generic Layout function is called to
// handle any additional layout requirements.
//
// Layout process:
//  1. Get the preferred size of the child
//  2. Position the child widget at the top-left corner
//  3. Delegate to the generic Layout function for final processing
func (s *Scroller) Layout() {
	if s.child != nil {
		cx, cy, _, _ := s.Content()
		pw, ph := s.child.Hint()
		s.child.SetBounds(cx, cy, pw, ph)
	}
	Layout(s)
}

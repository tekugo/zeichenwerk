package zeichenwerk

type ThemeSwitch struct {
	BaseWidget
	theme Theme
	child Widget // The single child widget contained within this theme switcher
}

func NewThemeSwitch(id string, theme Theme) *ThemeSwitch {
	return &ThemeSwitch{
		BaseWidget: BaseWidget{id: id, focusable: false},
		theme:      theme,
	}
}

// Add sets the child widget for this container.
// Since THemeSwitch is designed to contain exactly one child widget, calling Add
// will replace any previously set child widget. If you need to contain
// multiple widgets, consider using a layout container (like Flex or Grid)
// as the child widget.
//
// Parameters:
//   - widget: The widget to be contained within this theme switch
func (t *ThemeSwitch) Add(widget Widget) {
	t.child = widget
}

// Children returns a slice containing the single child widget of this theme switch.
// This method implements the Container interface requirement. The returned
// slice will contain exactly one element if a child has been added, or will
// be empty if no child widget has been set.
//
// Returns:
//   - []Widget: A slice containing the child widget, or empty slice if no child is set
//
// Note: The returned slice should not be modified directly. Use the Add method
// to set or change the child widget.
func (t *ThemeSwitch) Children(_ bool) []Widget {
	if t.child == nil {
		return []Widget{}
	}
	return []Widget{t.child}
}

// Find searches for a widget with the specified ID within this contaier and its child widget.
// The search is performed recursively, first checking if this container matches the ID,
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
//  1. Check if this theme switch's ID matches
//  2. Recursively search within the child widget (if it exists)
//  3. If child is a container, search its descendants
func (t *ThemeSwitch) Find(id string, visible bool) Widget {
	return Find(t, id, visible)
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
func (t *ThemeSwitch) FindAt(x, y int) Widget {
	return FindAt(t, x, y)
}

// Hint returns the theme switch's preferred size for optimal display.
// The calculation includes the styling overhead (borders, padding, margins)
// plus the preferred size of its child widget. This ensures the theme switch
// can accommodate both its visual styling and its contained content.
//
// Returns:
//   - int: Preferred width including styling overhead and child requirements
//   - int: Preferred height including styling overhead and child requirements
//
// Size calculation:
//  1. Start with the theme switch's horizontal and vertical styling overhead
//  2. Add the child widget's preferred dimensions (if child exists)
//  3. Return the total required space for optimal display
func (t *ThemeSwitch) Hint() (int, int) {
	w := t.Style("").Horizontal()
	h := t.Style("").Vertical()
	if t.child != nil {
		cw, ch := t.child.Hint()
		w += cw
		h += ch
	}
	return w, h
}

func (t *ThemeSwitch) Info() string {
	return "theme-switch [" + t.BaseWidget.Info() + "]"
}

// Layout positions and sizes the child widget within the content area.
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
func (t *ThemeSwitch) Layout() {
	if t.child != nil {
		cx, cy, cw, ch := t.Content()
		t.child.SetBounds(cx, cy, cw, ch)
	}
	Layout(t)
}

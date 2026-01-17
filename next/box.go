package next

// Box represents a container widget that wraps a single child widget with an
// optional margin, border, title and padding. The box automatically handles
// layout by fitting the child widget within its content area.
//
// The Box widget is useful for grouping related UI elements, providing visual
// separation, and adding descriptive titles to sections of the interface. It
// serves as a fundamental building block for creating organized and visually
// structured layouts.
//
// Layout behavior:
//   - The child widget is positioned within the box's content area
//   - Content area excludes the box's borders, margins, and padding
//   - The box's preferred size is the child with its margin, border and padding
type Box struct {
	Component
	Title string // The title text displayed in the box header (optional)
	child Widget // The single child widget contained within this box
}

// NewBox creates a new box container widget with the specified ID and title.
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
		Component: Component{id: id},
		Title:     title,
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
func (b *Box) Children() []Widget {
	if b.child == nil {
		return []Widget{}
	}
	return []Widget{b.child}
}

// Hint returns the box's preferred content size for optimal display.
// This does not include the box's styling overhead (borders, padding, margins)
// but that of its child.
//
// Returns:
//   - int: Preferred width without box styling
//   - int: Preferred height without box styling
func (b *Box) Hint() (int, int) {
	if b.child != nil {
		w, h := b.child.Hint()
		style := b.child.Style()
		w += style.Horizontal()
		h += style.Vertical()
		return w, h
	} else {
		return 0, 0
	}
}

// Layout positions and sizes the child widget within this box's content area.
// The child widget is given the full content area of the box, which excludes
// the space used by borders, padding, and margins. After positioning the child,
// the layout of the child is also called, if it is a container itself.
func (b *Box) Layout() {
	if b.child != nil {
		cx, cy, cw, ch := b.Content()
		b.child.SetBounds(cx, cy, cw, ch)
	}
	Layout(b)
}

// Render renders the box and its child widget.
func (b *Box) Render(r *Renderer) {
	b.Component.Render(r)

	// Determine the style to use based on the widget state
	state := b.State()
	if state != "" {
		state = ":" + state
	}
	style := b.Style(state)

	if b.Title != "" {
		titleStyle := b.Style("title")
		r.Set(titleStyle.Foreground(), titleStyle.Background(), titleStyle.Font())

		// Use boxStyle margin for positioning to align with the border
		r.Text(b.x+style.Margin().Left+2, b.y+style.Margin().Top, " "+b.Title+" ", 0)
	}
	if b.child != nil {
		b.child.Render(r)
	}
}

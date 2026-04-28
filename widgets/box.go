package widgets

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// Box represents a container widget that wraps a single child widget with an
// optional margin, border, title and padding. The box automatically handles
// layout by fitting the child widget within its content area.
//
// The Box widget is useful for grouping related UI elements, providing visual
// separation, and adding descriptive titles to sections of the interface. It
// is a very simple container for just a single widget. Its main feature is
// the addition of a title inside the border area.
//
// Layout behavior:
//   - The child widget is positioned within the box's content area
//   - Content area excludes the box's borders, margins, and padding
//   - The preferred size is the child size plus margin, border and padding
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
func NewBox(id, class, title string) *Box {
	return &Box{
		Component: Component{id: id, class: class},
		Title:     title,
	}
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies a theme style to the component.
func (b *Box) Apply(theme *Theme) {
	theme.Apply(b, b.Selector("box"))
}

// Hint returns the box's preferred content size for optimal display.
// This does not include the box's styling overhead (borders, padding, margins)
// but that of its child. If a preferred size is set manually using SetHint,
// that size is returned.
//
// Returns:
//   - int: Preferred width without box styling
//   - int: Preferred height without box styling
func (b *Box) Hint() (int, int) {
	if b.hwidth != 0 || b.hheight != 0 {
		return b.hwidth, b.hheight
	} else if b.child != nil {
		w, h := b.child.Hint()
		style := b.child.Style()
		w += style.Horizontal()
		h += style.Vertical()
		return w, h
	} else {
		return 0, 0
	}
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

// ---- Setter ---------------------------------------------------------------

// Set sets the box value (title) in a generic way.
func (b *Box) Set(value string) {
	b.Title = value
	b.Refresh()
}

// ---- Container Methods ----------------------------------------------------

// Add sets the child widget for this box container.
// Since Box is designed to contain exactly one child widget, calling Add
// will replace any previously set child widget. If you need to contain
// multiple widgets, consider using a layout container (like Flex or Grid)
// as the child widget.
//
// Parameters:
//   - widget: The widget to be contained within this box
func (b *Box) Add(widget Widget, params ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	if b.child != nil {
		b.child.SetParent(nil) // clear old parent reference
	}
	b.child = widget
	b.child.SetParent(b)
	return nil
}

// Children returns a slice containing the single child widget of this box.
// This method implements the Container interface requirement. The returned
// slice will contain exactly one element if a child has been added, or will
// be empty if no child widget has been set.
//
// Returns:
//   - []Widget: A slice containing the child widget, or empty slice
func (b *Box) Children() []Widget {
	if b.child == nil {
		return []Widget{}
	}
	return []Widget{b.child}
}

// Layout positions and sizes the child widget within this box's content area.
// The child widget is given the full content area of the box, which excludes
// the space used by borders, padding, and margins. After positioning the child,
// the layout of the child is also called, if it is a container itself.
func (b *Box) Layout() error {
	if b.child != nil {
		cx, cy, cw, ch := b.Content()
		b.child.SetBounds(cx, cy, cw, ch)
	}
	return Layout(b)
}

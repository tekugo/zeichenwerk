package next

// Container represents a widget that can contain and manage child widgets.
// It extends the Widget interface with methods for managing child widget
// collections, enabling complex UI layouts and hierarchical structures.
type Container interface {
	Widget

	// Children returns a slice of all direct child widgets.
	// All children will be returned, whether they are visible or not.
	Children() []Widget

	// Layout arranges the child widgets in the container.
	Layout()
}

// Find searches for a widget by ID within a container hierarchy.
//
// Parameters:
//   - container: The container to search within
//   - id: The unique identifier of the widget to find
//
// Returns:
//   - Widget: The widget with the matching ID, or nil if not found
func Find(container Container, id string) Widget {
	if container.ID() == id {
		return container
	}
	for _, child := range container.Children() {
		if child.ID() == id {
			return child
		}
		inner, ok := child.(Container)
		if ok {
			widget := Find(inner, id)
			if widget != nil {
				return widget
			}
		}
	}
	return nil
}

// FindAll returns all widgets of the specified type found inside the container.
//
// Parameters:
//   - container: The container to search within
//   - visible: If true, search only visible children; if false, search all
//
// Returns:
//   - T: An array containing all found widgets, may be empty, but never nil
func FindAll[T any](container Container) []T {
	var result []T
	Traverse(container, func(widget Widget) {
		if val, ok := widget.(T); ok {
			result = append(result, val)
		}
	})
	return result
}

// FindAt locates a widget at specific coordinates within a container hierarchy.
//
// Parameters:
//   - container: The container to search within
//   - x: The x-coordinate to search at
//   - y: The y-coordinate to search at
//
// Returns:
//   - Widget: The most specific widget at coordinates, or nil if outside bounds
func FindAt(container Container, x, y int) Widget {
	cx, cy, cw, ch := container.Bounds()

	// Check if it is inside the bounds
	if x < cx || y < cy || x >= cx+cw || y >= cy+ch {
		return nil
	}

	for _, child := range container.Children() {
		visible := child.Flag("visible")
		if !visible {
			continue
		}
		cx, cy, cw, ch = child.Bounds()
		if x >= cx && y >= cy && x < cx+cw && y < cy+ch {
			inner, ok := child.(Container)
			if ok {
				widget := FindAt(inner, x, y)
				if widget != nil {
					return widget
				}
			}
			return child
		}
	}

	return container
}

// Layout recursively triggers layout on all child containers within the given container.
// Ensures the entire widget hierarchy is properly laid out after changes.
//
// Parameters:
//   - container: The container whose child containers should be laid out
func Layout(container Container) {
	for _, child := range container.Children() {
		if inner, ok := child.(Container); ok {
			inner.Layout()
		}
	}
}

// Traverse recursively visits all widgets in the container hierarchy and applies
// the given function to each widget. Uses depth-first traversal order.
//
// Parameters:
//   - container: The container to traverse
//   - fn: The function to apply to each widget
func Traverse(container Container, fn func(Widget)) {
	for _, child := range container.Children() {
		fn(child)
		if inner, ok := child.(Container); ok {
			Traverse(inner, fn)
		}
	}
}

package zeichenwerk

// Container represents a widget that can contain and manage child widgets.
// It extends the Widget interface with methods for managing child widget
// collections, enabling complex UI layouts and hierarchical structures.
//
// Core responsibilities:
//   - Child management: Add, remove, and organize child widgets
//   - Widget lookup: Find widgets by ID or screen coordinates
//   - Layout coordination: Position and size child widgets appropriately
//   - Event propagation: Route events to appropriate child widgets
//   - Hierarchy management: Maintain parent-child relationships
//
// Common implementations: Box, Flex, Grid, Stack, Switcher, Scroller, Tabs
type Container interface {
	Widget

	// Children returns a slice of all direct child widgets.
	// The visible parameter controls whether to include only visible children
	// or all children regardless of visibility state.
	//
	// Parameters:
	//   - visible: If true, return only visible children; if false, return all
	//
	// Returns:
	//   - []Widget: Slice containing the child widgets
	Children(bool) []Widget

	// Find searches for a widget with the specified ID within this container
	// and all descendant widgets. Performs recursive traversal of the widget tree.
	//
	// Parameters:
	//   - string: The unique identifier of the widget to find
	//   - bool: If true, search only visible children; if false, search all
	//
	// Returns:
	//   - Widget: The widget with the matching ID, or nil if not found
	Find(string, bool) Widget

	// FindAt locates the widget at the specified screen coordinates.
	// Used for mouse interaction to determine which widget is at a given point.
	//
	// Parameters:
	//   - int: The x-coordinate to search at
	//   - int: The y-coordinate to search at
	//
	// Returns:
	//   - Widget: The widget at the coordinates, or nil if none found
	FindAt(int, int) Widget

	// Layout arranges and positions all child widgets within this container
	// according to the container's layout algorithm and child size hints.
	//
	// Called when container size changes or when children are added/removed.
	Layout()
}

// Find searches for a widget by ID within a container hierarchy.
// Provides a default implementation that container types can use.
//
// Algorithm:
//   - Checks container's own ID first
//   - Searches direct children for matching ID
//   - Recursively searches child containers
//   - Returns first match found
//
// Parameters:
//   - container: The container to search within
//   - id: The unique identifier of the widget to find
//   - visible: If true, search only visible children; if false, search all
//
// Returns:
//   - Widget: The widget with the matching ID, or nil if not found
func Find(container Container, id string, visible bool) Widget {
	if container.ID() == id {
		return container
	}
	for _, child := range container.Children(visible) {
		if child.ID() == id {
			return child
		}
		inner, ok := child.(Container)
		if ok {
			widget := inner.Find(id, visible)
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
func FindAll[T any](container Container, visible bool) []T {
	var result []T
	Traverse(container, visible, func(widget Widget) {
		if val, ok := widget.(T); ok {
			result = append(result, val)
		}
	})
	return result
}

// FindAt locates a widget at specific coordinates within a container hierarchy.
// Provides a default implementation that container types can use.
//
// Algorithm:
//   - Checks if coordinates are within container bounds
//   - Searches children for widgets containing the coordinates
//   - Recursively searches child containers first (respects layering)
//   - Returns the most specific (deepest) widget found
//   - Falls back to container itself if no children match
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

	for _, child := range container.Children(true) {
		cx, cy, cw, ch = child.Bounds()
		if x >= cx && y >= cy && x < cx+cw && y < cy+ch {
			inner, ok := child.(Container)
			if ok {
				widget := inner.FindAt(x, y)
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
// Process:
//   - Iterates through all direct children
//   - Calls Layout() on each child container
//   - Skips non-container widgets (leaf widgets)
//
// Parameters:
//   - container: The container whose child containers should be laid out
func Layout(container Container) {
	for _, child := range container.Children(false) {
		if inner, ok := child.(Container); ok {
			inner.Layout()
		}
	}
}

// Traverse recursively visits all widgets in the container hierarchy and applies
// the given function to each widget. Uses depth-first traversal order.
//
// Common use cases:
//   - Collecting statistics about widgets
//   - Applying changes to all widgets
//   - Searching for widgets with specific properties
//   - Debugging widget trees
//
// Parameters:
//   - container: The container to traverse
//   - visible: If true, traverse only visible children; if false, traverse all
//   - fn: The function to apply to each widget
//
// Example usage:
//
//	// Count all widgets
//	count := 0
//	Traverse(container, true, func(w Widget) { count++ })
//
//	// Disable all buttons
//	Traverse(container, true, func(w Widget) {
//		if btn, ok := w.(*Button); ok { btn.SetEnabled(false) }
//	})
func Traverse(container Container, visible bool, fn func(Widget)) {
	for _, child := range container.Children(visible) {
		fn(child)
		if inner, ok := child.(Container); ok {
			Traverse(inner, visible, fn)
		}
	}
}

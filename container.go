package zeichenwerk

// Container represents a widget that can contain and manage child widgets.
// It extends the Widget interface with additional methods for managing
// a collection of child widgets, enabling the creation of complex UI layouts
// and hierarchical widget structures.
//
// Containers are responsible for:
//   - Managing child widget collections
//   - Providing widget lookup functionality by ID
//   - Coordinating layout and rendering of child widgets
//   - Propagating events and state changes to children
//
// Common container implementations include layouts like Stack, Grid, Flex,
// and other composite widgets that organize multiple child components.
type Container interface {
	Widget

	// Children returns a slice of all direct child widgets contained within this container.
	// The returned slice represents the current state of the container's child collection
	// and should not be modified directly. The order of widgets in the slice may be
	// significant for layout and rendering purposes depending on the container implementation.
	//
	// Parameeters:
	//   - bool: Flag if only visible children should be returned
	//
	// Returns:
	//   - []Widget: A slice containing all direct child widgets
	Children(bool) []Widget

	// Find searches for a widget with the specified ID within this container and
	// all of its descendant widgets. The search is typically performed recursively,
	// traversing the entire widget hierarchy rooted at this container.
	//
	// The search process:
	//   - First checks direct children for a matching ID
	//   - Recursively searches within child containers
	//   - Returns the first widget found with the matching ID
	//   - Returns nil if no widget with the specified ID is found
	//
	// Parameters:
	//   - string: The unique identifier of the widget to find
	//   - bool: Only look for visible children
	//
	// Returns:
	//   - Widget: The widget with the matching ID, or nil if not found
	Find(string, bool) Widget

	// FindAt searches for the widget located at the specified coordinates within
	// this container and its child widgets. This method is used for mouse interaction
	// to determine which widget is positioned at a given point on the screen.
	//
	// The search process:
	//   - Checks if the coordinates fall within this container's bounds
	//   - Recursively searches child widgets and containers at the coordinates
	//   - Returns the topmost/most specific widget at the given position
	//   - Considers widget layering and z-order where applicable
	//
	// Parameters:
	//   - int: The x-coordinate to search at
	//   - int: The y-coordinate to search at
	//
	// Returns:
	//   - Widget: The widget at the specified coordinates, or nil if no widget is found
	FindAt(int, int) Widget

	// Layout arranges and positions all child widgets within this container
	// according to the container's layout algorithm and any layout hints or
	// constraints associated with the child widgets.
	//
	// The layout process typically involves:
	//   - Calculating positions and sizes for all child widgets
	//   - Applying layout constraints and hints
	//   - Recursively laying out child containers
	//   - Updating widget bounds based on the layout algorithm
	//
	// This method should be called whenever the container's size changes
	// or when child widgets are added, removed, or modified in ways that
	// affect the layout.
	Layout()
}

// Find is a utility function that provides a default implementation for searching
// widgets by ID within a container hierarchy. This function can be used by
// concrete container implementations to implement their Find method.
//
// The search algorithm:
//   - Iterates through all direct children of the container
//   - Checks each child's ID for an exact match
//   - If a child is also a container, recursively searches within it
//   - Returns the first widget found with the matching ID
//   - Returns nil if no widget with the specified ID exists in the hierarchy
//
// Parameters:
//   - container: The container to search within
//   - id: The unique identifier of the widget to find
//   - visible: Find only visible children
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

// FindAt is a utility function that provides a default implementation for finding
// widgets at specific coordinates within a container hierarchy. This function can
// be used by concrete container implementations to implement their FindAt method.
//
// The search algorithm:
//   - First checks if the coordinates fall within the container's bounds
//   - Returns nil immediately if coordinates are outside the container
//   - Iterates through all direct children to find those containing the coordinates
//   - For child containers, recursively searches within them first
//   - Returns the most specific (deepest) widget found at the coordinates
//   - Falls back to returning the container itself if no children match
//
// This implementation respects widget layering by checking children in order,
// allowing later children to override earlier ones at the same coordinates.
//
// Parameters:
//   - container: The container to search within
//   - x: The x-coordinate to search at
//   - y: The y-coordinate to search at
//
// Returns:
//   - Widget: The most specific widget at the coordinates, or nil if outside bounds
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

// Layout is a utility function that recursively triggers layout on all child
// containers within the given container. This function provides a convenient
// way to ensure that the entire widget hierarchy is properly laid out.
//
// The function:
//   - Iterates through all direct children of the container
//   - Identifies which children are themselves containers
//   - Calls Layout() on each child container
//   - Does not affect non-container widgets (leaf widgets)
//
// This is typically used after making changes to the widget hierarchy
// or when the container's size has changed and all descendant containers
// need to recalculate their layouts.
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
// the given function to each widget. This utility function provides a convenient
// way to perform operations on all widgets within a container tree.
//
// The traversal algorithm:
//   - Visits all direct children of the container first
//   - Applies the function to each child widget
//   - For child containers, recursively traverses their children
//   - Uses depth-first traversal order
//   - Visits each widget exactly once
//
// Common use cases:
//   - Collecting statistics about widgets in the hierarchy
//   - Applying styling or configuration changes to all widgets
//   - Searching for widgets with specific properties
//   - Debugging and inspection of widget trees
//
// Parameters:
//   - container: The container to traverse
//   - visible: Traverse only the visible children
//   - fn: The function to apply to each widget in the hierarchy
//
// Example usage:
//
//	// Count all widgets in the hierarchy
//	count := 0
//	Traverse(rootContainer, func(w Widget) {
//	    count++
//	})
//
//	// Find all buttons and disable them
//	Traverse(rootContainer, func(w Widget) {
//	    if button, ok := w.(*Button); ok {
//	        button.SetEnabled(false)
//	    }
//	})
func Traverse(container Container, visible bool, fn func(Widget)) {
	for _, child := range container.Children(visible) {
		fn(child)
		if inner, ok := child.(Container); ok {
			Traverse(inner, visible, fn)
		}
	}
}

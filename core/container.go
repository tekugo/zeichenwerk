package core

// Container represents a widget that can contain and manage child widgets.
// It extends the Widget interface with methods for managing child widget
// collections, enabling complex UI layouts and hierarchical structures.
type Container interface {
	Widget

	// Adds a widget to the container
	Add(widget Widget, params ...any) error

	// Children returns a slice of all direct child widgets.
	// All children will be returned, whether they are visible or not.
	Children() []Widget

	// Layout arranges the child widgets in the container.
	Layout() error
}

package widgets

import (
	"fmt"
	"strings"

	. "github.com/tekugo/zeichenwerk/core"
)

// FindRoot traverses up the widget hierarchy to find the root instance.
// This utility function is useful for widgets that need access to UI-level
// functionality such as logging, popup management, or theme access.
//
// Parameters:
//   - widget: The starting widget to traverse up from
//
// Returns:
//   - *UI: The root UI instance, or nil if not found
func FindRoot(widget Widget) Root {
	current := widget
	for current != nil {
		if root, ok := current.(Root); ok {
			return root
		}
		current = current.Parent()
	}
	return nil
}

// Returns the widget ID, if the widget is not nil, `"<nil>"` otherwise.
func ID(widget Widget) string {
	if widget != nil {
		return widget.ID()
	} else {
		return "<nil>"
	}
}

// Redraw queues a single widget for redraw.
func Redraw(widget Widget) {
	root := FindRoot(widget)
	if root != nil {
		root.Redraw(widget)
	}
}

// Relayout walks up the parent chain to the root UI, re-runs the full layout
// top-down, and then queues a full screen repaint. Use this when a widget
// changes its own preferred size at runtime (e.g. Collapsible expand/collapse).
func Relayout(widget Widget) {
	root := FindRoot(widget)
	if root != nil {
		root.Layout()
		root.Refresh()
	}
}

// WidgetType returns a clean, human-readable string representation of the
// widget's type.
func WidgetType(widget Widget) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", widget), "*widgets.")
}

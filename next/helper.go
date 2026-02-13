package next

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// FindUI traverses up the widget hierarchy to find the root UI instance.
// This utility function is useful for widgets that need access to UI-level
// functionality such as logging, popup management, or theme access.
//
// Parameters:
//   - widget: The starting widget to traverse up from
//
// Returns:
//   - *UI: The root UI instance, or nil if not found
func FindUI(widget Widget) *UI {
	current := widget
	for current != nil {
		if ui, ok := current.(*UI); ok {
			return ui
		}
		current = current.Parent()
	}
	return nil
}

// OnKey registers a key event handler for the given widget.
func OnKey(widget Widget, handler func(Widget, *tcell.EventKey) bool) {
	if widget == nil {
		return
	}

	widget.On("key", func(widget Widget, event string, data ...any) bool {
		if len(data) != 1 {
			return false
		}
		if ev, ok := data[0].(*tcell.EventKey); ok {
			return handler(widget, ev)
		} else {
			return false
		}
	})
}

// OnMouse registers a mouse event handler for the given widget.
func OnMouse(widget Widget, handler func(Widget, *tcell.EventMouse) bool) {
	if widget == nil {
		return
	}

	widget.On("mouse", func(widget Widget, event string, data ...any) bool {
		if len(data) != 1 {
			return false
		}
		if ev, ok := data[0].(*tcell.EventMouse); ok {
			return handler(widget, ev)
		} else {
			return false
		}
	})
}

// Redraw queues a single widget for redraw.
func Redraw(widget Widget) {
	ui := FindUI(widget)
	if ui != nil {
		ui.Redraw(widget)
	}
}

// WidgetType returns a clean, human-readable string representation of the
// widget's type.
func WidgetType(widget Widget) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", widget), "*next.")
}

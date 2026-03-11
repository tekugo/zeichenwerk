package zeichenwerk

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

// Returns the widget ID, if the widget is not nil, `"<nil>"` otherwise.
func ID(widget Widget) string {
	if widget != nil {
		return widget.ID()
	} else {
		return "<nil>"
	}
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

// HandleKeyEvent registers a key event handler for a widget within a container.
func HandleKeyEvent(container Container, id string, fn func(Widget, *tcell.EventKey) bool) {
	widget := Find(container, id)
	if widget == nil {
		container.Log(container, "error", "Widget %s not found", id)
		return
	}
	OnKey(widget, fn)
}

// HandleListEvent registers a handler for List-specific events.
func HandleListEvent(container Container, id, event string, fn func(*List, string, int) bool) {
	widget := Find(container, id)
	if widget == nil {
		container.Log(container, "error", "Widget %s not found", id)
		return
	}
	list, ok := widget.(*List)
	if !ok {
		container.Log(container, "error", "Widget %s is not a List", id)
		return
	}
	widget.On(event, func(widget Widget, event string, data ...any) bool {
		if len(data) != 1 {
			container.Log(container, "error", "List event %s expected 1 data parameter, got %d", event, len(data))
			return false
		}
		index, ok := data[0].(int)
		if !ok {
			container.Log(container, "error", "List event %s data should be int, got %T", event, data[0])
			return false
		}
		return fn(list, event, index)
	})
}

// Update updates widget content based on type.
func Update(container Container, id string, value any) {
	widget := Find(container, id)
	if widget == nil {
		return
	}
	switch w := widget.(type) {
	case *Static:
		if str, ok := value.(string); ok {
			w.SetText(str)
		} else {
			w.SetText(fmt.Sprintf("%v", value))
		}
	case *List:
		if items, ok := value.([]string); ok {
			w.SetItems(items)
			// Select first item if available
			if len(items) > 0 {
				w.Dispatch("select", 0)
			}
		}
	case *Text:
		if lines, ok := value.([]string); ok {
			w.Set(lines)
		}
	}
}

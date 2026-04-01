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

// OnAccept registers an accept event handler for the given widget.
// The handler receives the accepted string value.
func OnAccept(widget Widget, handler func(string) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if value, ok := data[0].(string); ok {
			return handler(value)
		}
		return false
	})
}

// OnActivate registers an activate event handler for the given widget.
// The handler receives the index of the activated item.
func OnActivate(widget Widget, handler func(int) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if index, ok := data[0].(int); ok {
			return handler(index)
		}
		return false
	})
}

// OnChange registers a change event handler for the given widget.
// The handler receives the new value as a string.
func OnChange(widget Widget, handler func(string) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if value, ok := data[0].(string); ok {
			return handler(value)
		}
		return false
	})
}

// OnEnter registers an Enter event handler for the given widget.
func OnEnter(widget Widget, handler func(value string) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtEnter, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if value, ok := data[0].(string); ok {
			return handler(value)
		}
		return false
	})
}

// OnHide registers a hide event handler for the given widget.
func OnHide(widget Widget, handler func() bool) {
	if widget == nil {
		return
	}
	widget.On(EvtHide, func(_ Widget, _ Event, data ...any) bool {
		return handler()
	})
}

// OnKey registers a key event handler for the given widget.
func OnKey(widget Widget, handler func(*tcell.EventKey) bool) {
	if widget == nil {
		return
	}

	widget.On(EvtKey, func(_ Widget, _ Event, data ...any) bool {
		if len(data) != 1 {
			return false
		}
		if ev, ok := data[0].(*tcell.EventKey); ok {
			return handler(ev)
		} else {
			return false
		}
	})
}

// OnMouse registers a mouse event handler for the given widget.
func OnMouse(widget Widget, handler func(*tcell.EventMouse) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtMouse, func(_ Widget, _ Event, data ...any) bool {
		if len(data) != 1 {
			return false
		}
		if ev, ok := data[0].(*tcell.EventMouse); ok {
			return handler(ev)
		} else {
			return false
		}
	})
}

// OnSelect registers a select event handler for the given widget.
// The handler receives the index of the selected item.
func OnSelect(widget Widget, handler func(int) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if index, ok := data[0].(int); ok {
			return handler(index)
		}
		return false
	})
}

// OnShow registers a show event handler for the given widget.
func OnShow(widget Widget, handler func() bool) {
	if widget == nil {
		return
	}
	widget.On(EvtShow, func(_ Widget, _ Event, data ...any) bool {
		return handler()
	})
}

// Suggest returns a suggest function for use with Typeahead.SetSuggest.
// It performs case-insensitive prefix matching against the provided candidates.
func Suggest(candidates []string) func(string) []string {
	return func(text string) []string {
		if text == "" {
			return nil
		}
		lower := strings.ToLower(text)
		var matches []string
		for _, c := range candidates {
			if strings.HasPrefix(strings.ToLower(c), lower) {
				matches = append(matches, c)
			}
		}
		return matches
	}
}

// Redraw queues a single widget for redraw.
func Redraw(widget Widget) {
	ui := FindUI(widget)
	if ui != nil {
		ui.Redraw(widget)
	}
}

// Relayout walks up the parent chain to the root UI, re-runs the full layout
// top-down, and then queues a full screen repaint. Use this when a widget
// changes its own preferred size at runtime (e.g. Collapsible expand/collapse).
func Relayout(widget Widget) {
	ui := FindUI(widget)
	if ui != nil {
		ui.Layout()
		ui.Refresh()
	}
}

// WidgetType returns a clean, human-readable string representation of the
// widget's type.
func WidgetType(widget Widget) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", widget), "*zeichenwerk.")
}

// Update updates the content of the widget identified by id within container.
// The widget must implement the Setter interface; if it does not, the call
// is silently ignored.
func Update(container Container, id string, value any) {
	widget := Find(container, id)
	if widget == nil {
		return
	}
	if s, ok := widget.(Setter); ok {
		s.Set(value)
	}
}

package values

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// Update updates the content of the widget identified by id within container.
// The widget must implement Setter[T]; if it does not, the call is silently
// ignored.
func Update[T any](container Container, id string, value T) {
	widget := Find(container, id)
	if widget == nil {
		return
	}
	if s, ok := widget.(Setter[T]); ok {
		s.Set(value)
	}
}

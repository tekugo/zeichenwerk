package zeichenwerk

import (
	"reflect"
)

// Form is a container that binds to a Go struct and manages form controls.
// It provides data binding between struct fields and form widgets (Input, Checkbox).
// The form itself does not render any visual content; its child (typically a FormGroup
// or layout container) is rendered. The form's primary role is to coordinate data
// synchronization.
type Form struct {
	Component
	title string
	data  any    // pointer to struct
	child Widget // single child container (e.g., FormGroup)
}

// NewForm creates a new Form container with the specified ID, title, and data struct.
// The data parameter must be a pointer to a struct. The form will use reflection
// to bind struct fields to form controls based on struct tags.
//
// Parameters:
//   - id: Unique identifier for the form
//   - title: Optional title (may be used by Theme or displayed by FormGroup)
//   - data: Pointer to the struct containing form data
func NewForm(id, class, title string, data any) *Form {
	f := &Form{
		Component: Component{id: id, class: class},
		title:     title,
		data:      data,
	}
	return f
}

// Add sets the child widget for the form.
// Since Form is designed to contain exactly one child container (typically a FormGroup
// or a layout container like Flex/Grid), calling Add will replace any previously set
// child widget.
func (f *Form) Add(widget Widget) {
	if widget == nil {
		return
	}
	if f.child != nil {
		f.child.SetParent(nil)
	}
	f.child = widget
	f.child.SetParent(f)
}

// Apply applies a theme's styles to the component.
func (f *Form) Apply(theme *Theme) {
	theme.Apply(f, f.Selector("form"))
}

// Children returns a slice containing the child widget of the form.
// This method implements the Container interface. The returned slice will contain
// exactly one element if a child has been added, or will be empty if no child
// widget has been set.
func (f *Form) Children() []Widget {
	if f.child == nil {
		return []Widget{}
	}
	return []Widget{f.child}
}

func (f *Form) Hint() (int, int) {
	if f.child != nil {
		w, h := f.child.Hint()
		style := f.child.Style()
		w += style.Horizontal()
		h += style.Vertical()
		return w, h
	} else {
		return 0, 0
	}
}

// Layout positions the child widget within the form's content area.
// It sets the child's bounds to the full content area of the form and
// then recursively lays out the child if it is a container.
func (f *Form) Layout() {
	f.Log(f, "Debug", "Form Layout")
	if f.child != nil {
		cx, cy, cw, ch := f.Content()
		f.child.SetBounds(cx, cy, cw, ch)
		if container, ok := f.child.(Container); ok {
			container.Layout()
		}
	}
}

// Render does nothing; the form itself is not a visual widget.
// Its child (e.g., FormGroup) is responsible for rendering.
func (f *Form) Render(r *Renderer) {
	f.Component.Render(r)
	f.child.Render(r)
}

// Update returns an event handler that writes the changed value back to the
// struct field associated with the given reflect.Value. The returned handler
// can be used with widget.On("change", form.Update(fieldValue)).
//
// The handler supports:
//   - Input: sets the struct field to the new string
//   - Checkbox: sets the struct field to the new boolean
func (f *Form) Update(value reflect.Value) Handler {
	return func(widget Widget, event string, data ...any) bool {
		switch widget.(type) {
		case *Input:
			if len(data) > 0 {
				if str, ok := data[0].(string); ok {
					value.SetString(str)
				}
			}
		case *Checkbox:
			if len(data) > 0 {
				if b, ok := data[0].(bool); ok {
					value.SetBool(b)
				}
			}
		default:
			widget.Log(widget, "warn", "Unknown widget type to update")
		}
		return false // allow event to continue bubbling
	}
}

// Data returns the data struct pointer associated with this form.
func (f *Form) Data() any {
	return f.data
}

// Title returns the form's title.
func (f *Form) Title() string {
	return f.title
}

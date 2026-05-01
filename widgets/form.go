package widgets

import (
	"reflect"
	"strconv"
	"strings"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// Form is a container that binds to a Go struct and manages form controls.
// It provides data binding between struct fields and form widgets (Input, Checkbox).
// The form itself does not render any visual content; its child (typically a FormGroup
// or layout container) is rendered. The form's primary role is to coordinate data
// synchronization.
type Form struct {
	Component
	Data  any // pointer to struct
	title string
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
		Data:      data,
		title:     title,
	}
	return f
}

// Add sets the child widget for the form.
// Since Form is designed to contain exactly one child container (typically a FormGroup
// or a layout container like Flex/Grid), calling Add will replace any previously set
// child widget.
func (f *Form) Add(widget Widget, params ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	if f.child != nil {
		f.child.SetParent(nil)
	}
	f.child = widget
	f.child.SetParent(f)
	return nil
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
	if f.hwidth != 0 || f.hheight != 0 {
		return f.hwidth, f.hheight
	}
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
func (f *Form) Layout() error {
	f.Log(f, Debug, "Form Layout")
	if f.child != nil {
		cx, cy, cw, ch := f.Content()
		f.child.SetBounds(cx, cy, cw, ch)
		if container, ok := f.child.(Container); ok {
			return container.Layout()
		}
	}
	return nil
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
	return func(widget Widget, event Event, data ...any) bool {
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
			widget.Log(widget, Warning, "Unknown widget type to update")
		}
		return false // allow event to continue bubbling
	}
}

// BuildFormGroup populates group with form controls generated from the struct
// fields of form's data. name filters by the "group" struct tag; pass an empty
// string to include all fields regardless of grouping. theme is applied to
// each generated control. This is the public equivalent of the Builder's
// internal buildGroup method, intended for use by the compose package and
// other non-Builder code that constructs forms imperatively.
func BuildFormGroup(form *Form, group *FormGroup, name string, theme *Theme) {
	line := 0
	v := reflect.ValueOf(form.Data)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return
	}
	v = v.Elem()
	t := v.Type()

	for i := range v.NumField() {
		sf := t.Field(i)
		fv := v.Field(i)
		g := sf.Tag.Get("group")
		if name != "" && name != g {
			continue
		}
		label := sf.Tag.Get("label")
		if label == "-" {
			continue
		} else if label == "" {
			label = sf.Name
		}
		control := sf.Tag.Get("control")
		options := sf.Tag.Get("options")
		_, readonly := sf.Tag.Lookup("readonly")
		width, err := strconv.Atoi(sf.Tag.Get("width"))
		if err != nil {
			width = 10
		}
		if l, err := strconv.Atoi(sf.Tag.Get("line")); err == nil {
			line = l
		}

		widget := buildFormControl(control, sf.Name, "", fv, options, theme)
		if readonly {
			widget.SetFlag(FlagReadonly, true)
		}
		widget.SetHint(width, 1)
		widget.On(EvtChange, form.Update(fv))
		group.Add(widget, line, label)
		line++
	}
}

func buildFormControl(control, id, class string, v reflect.Value, options string, theme *Theme) Widget {
	if control == "" {
		switch v.Kind() {
		case reflect.Bool:
			control = "checkbox"
		default:
			control = "input"
		}
	}
	switch control {
	case "checkbox":
		w := NewCheckbox(id, class, id, v.Bool())
		w.Apply(theme)
		w.SetFlag(FlagChecked, v.Bool())
		return w
	case "password":
		w := NewInput(id, class, "", "", "*")
		w.SetFlag(FlagMasked, true)
		w.Apply(theme)
		w.Set(v.String())
		return w
	case "select":
		o := strings.Split(options, ",")
		w := NewSelect(id, class, o...)
		w.Apply(theme)
		w.Select(v.String())
		return w
	default:
		w := NewInput(id, class)
		w.Apply(theme)
		w.Set(v.String())
		return w
	}
}

// Title returns the form's title.
func (f *Form) Title() string {
	return f.title
}

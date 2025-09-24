package zeichenwerk

import "reflect"

// Form represents a form container that manages data binding between Go structs and form controls.
// It provides automatic form generation and data synchronization capabilities.
//
// The Form widget serves as a container for form groups and handles the binding of struct fields
// to UI controls. It supports automatic field validation, data updates, and event handling.
//
// Example usage:
//   data := &struct {
//     Name  string `label:"Full Name" width:"30"`
//     Email string `label:"Email Address" width:"40"`
//     Age   int    `label:"Age" width:"10"`
//   }{}
//   
//   form := NewForm("user-form", "User Registration", data)
//
// Struct tag attributes supported:
//   - label: Display label for the field (default: field name)
//   - width: Width of the input control in characters
//   - control: Type of control ("input", "checkbox", "password") - auto-detected if not specified
//   - group: Group name for organizing fields
//   - line: Line number within a group for layout control
//   - readOnly: Makes the field read-only (true/false)
type Form struct {
	BaseWidget
	Title string // Form title displayed in the container border
	Data  any    // Form data as pointer to struct - must be a pointer for updates to work
	child Widget // Child container (typically a FormGroup or layout container)
}

// Add sets the child widget for this form container.
// Typically used to add a FormGroup or layout container that holds the form controls.
//
// Parameters:
//   - widget: The child widget to add (usually a FormGroup)
func (f *Form) Add(widget Widget) {
	f.child = widget
}

func (f *Form) Children(_ bool) []Widget {
	if f.child != nil {
		return []Widget{f.child}
	} else {
		return []Widget{}
	}
}

func (f *Form) Find(id string, visible bool) Widget {
	return Find(f, id, visible)
}

func (f *Form) FindAt(x, y int) Widget {
	return FindAt(f, x, y)
}

func (f *Form) Hint() (int, int) {
	if f.child != nil {
		w, h := f.child.Hint()
		w += f.child.Style("").Horizontal()
		h += f.child.Style("").Vertical()
		style := f.Style("")
		if style != nil {
			style.Width = w
			style.Height = h
		}
		return w, h
	} else {
		return 0, 0
	}
}

func (f *Form) Layout() {
	if f.child != nil {
		cx, cy, cw, ch := f.Content()
		f.child.SetBounds(cx, cy, cw, ch)
	}
	Layout(f)
}

// Update creates an event handler function that synchronizes form control values 
// back to the bound struct field when the control's value changes.
//
// This method is typically called automatically during form construction to set up
// data binding between form controls and struct fields.
//
// Parameters:
//   - value: The reflect.Value of the struct field to update
//
// Returns:
//   - An event handler function that updates the struct field when triggered
//
// Supported control types:
//   - *Checkbox: Updates boolean fields
//   - *Input: Updates string fields
func (f *Form) Update(value reflect.Value) func(Widget, string, ...any) bool {
	return func(widget Widget, event string, values ...any) bool {
		switch widget := widget.(type) {
		case *Checkbox:
			value.SetBool(widget.Checked)
		case *Input:
			value.SetString(widget.Text)
		}
		return false
	}
}

// NewForm creates a new form widget with the specified ID, title, and data structure.
//
// The data parameter must be a pointer to a struct. The struct fields will be automatically
// converted to form controls based on their types and struct tags.
//
// Parameters:
//   - id: Unique identifier for the form widget
//   - title: Display title for the form (shown in border if styled)
//   - data: Pointer to struct containing the form data
//
// Returns:
//   - *Form: New form widget instance
//
// Example:
//   type User struct {
//     Name  string `label:"Full Name" width:"30"`
//     Email string `label:"Email" width:"40"`
//     Admin bool   `label:"Administrator"`
//   }
//   
//   user := &User{}
//   form := NewForm("user-form", "User Details", user)
func NewForm(id, title string, data any) *Form {
	return &Form{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Title:      title,
		Data:       data,
	}
}

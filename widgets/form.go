package widgets

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	. "github.com/tekugo/zeichenwerk/core"
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
//   - Input → string field: sets the struct field to the new string
//   - Input → int field:    parses decimal; ignores the change on parse error
//   - Checkbox → bool:      sets the struct field to the new boolean
//   - Select → string field: sets the struct field to the selected value
func (f *Form) Update(value reflect.Value) Handler {
	return func(widget Widget, event Event, data ...any) bool {
		switch widget.(type) {
		case *Input:
			if len(data) == 0 {
				return false
			}
			str, ok := data[0].(string)
			if !ok {
				return false
			}
			switch value.Kind() {
			case reflect.String:
				value.SetString(str)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if n, err := strconv.ParseInt(str, 10, 64); err == nil {
					value.SetInt(n)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if n, err := strconv.ParseUint(str, 10, 64); err == nil {
					value.SetUint(n)
				}
			}
		case *Checkbox:
			if len(data) > 0 {
				if b, ok := data[0].(bool); ok {
					value.SetBool(b)
				}
			}
		case *Select:
			// Select dispatches EvtChange with the selected
			// item's string value; we just store it. The field
			// kind is always string for Select-driven controls
			// (the form-control switch only emits a Select for
			// "select" / "border" tags, both string-typed).
			if len(data) == 0 {
				return false
			}
			str, ok := data[0].(string)
			if !ok {
				return false
			}
			if value.Kind() == reflect.String {
				value.SetString(str)
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
//
// Anonymous embedded struct fields are recursed into, so a form struct that
// embeds a base struct (e.g. ComponentForm) renders all the embedded fields
// inline as if they were declared on the outer struct. Unexported fields and
// fields whose Kind is unsupported by buildFormControl (slices, arrays,
// channels, maps, …) are skipped silently.
func BuildFormGroup(form *Form, group *FormGroup, name string, theme *Theme) {
	v := reflect.ValueOf(form.Data)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return
	}
	line := 0
	buildGroupFields(form, group, v.Elem(), name, theme, &line)
}

// BuildFormGroupAt populates group with controls for v's directly
// declared, exported fields without recursing into anonymous
// embedded structs. The intended use is rendering a multi-section
// property panel where each embedded struct level becomes its own
// FormGroup; callers iterate v's anonymous fields, calling
// BuildFormGroupAt once per level (passing each struct's reflect
// value separately).
//
// v must be a struct value whose path back to form.Data is
// addressable, otherwise the generated controls cannot write back
// through reflect. Calling on a value obtained via
// reflect.ValueOf(p).Elem() and a chain of Field() lookups satisfies
// that constraint.
//
// The "group" tag filter, label / control / options / readonly /
// width / line tag handling, and skip-unsupported-kinds rules are
// identical to BuildFormGroup at the same level.
func BuildFormGroupAt(form *Form, group *FormGroup, v reflect.Value, name string, theme *Theme) {
	if v.Kind() != reflect.Struct {
		return
	}
	line := 0
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		fv := v.Field(i)

		// Skip anonymous embedded structs at this level — they
		// belong to their own section.
		if sf.Anonymous && fv.Kind() == reflect.Struct {
			continue
		}
		if !sf.IsExported() {
			continue
		}
		if !supportedFieldKind(fv.Kind()) {
			continue
		}
		buildOneFieldControl(form, group, sf, fv, name, theme, &line)
	}
}

// buildOneFieldControl is the shared per-field rendering logic
// extracted so both BuildFormGroup (recursive) and BuildFormGroupAt
// (shallow) produce identical controls. Returns silently if the
// field is filtered out by the "group" tag or labelled "-".
func buildOneFieldControl(form *Form, group *FormGroup, sf reflect.StructField, fv reflect.Value, name string, theme *Theme, line *int) {
	g := sf.Tag.Get("group")
	if name != "" && name != g {
		return
	}
	label := sf.Tag.Get("label")
	if label == "-" {
		return
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
		*line = l
	}

	outer, bound := buildFormControl(control, sf.Name, "", fv, options, theme)
	if readonly {
		// FlagReadonly belongs on the bound widget — it's the
		// editable control (Input) whose key handler honours the
		// flag. The outer wrapper (e.g. an HFlex around an Input
		// + color preview) is layout-only.
		bound.SetFlag(FlagReadonly, true)
	}
	outer.SetHint(width, 1)
	bound.On(EvtChange, form.Update(fv))
	group.Add(outer, *line, label)
	*line++
}

// buildGroupFields walks v's exported fields and adds one control per
// renderable field to group. Anonymous embedded structs are flattened
// recursively. line is shared across the recursion so the outer FormGroup
// receives a single, contiguous sequence of lines.
func buildGroupFields(form *Form, group *FormGroup, v reflect.Value, name string, theme *Theme, line *int) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		fv := v.Field(i)

		// Recurse into anonymous embedded structs so embedded fields render
		// inline. The "group" filter applies to the recursed fields, not to
		// the embedded struct itself.
		if sf.Anonymous && fv.Kind() == reflect.Struct {
			buildGroupFields(form, group, fv, name, theme, line)
			continue
		}

		if !sf.IsExported() {
			continue
		}
		if !supportedFieldKind(fv.Kind()) {
			continue
		}
		buildOneFieldControl(form, group, sf, fv, name, theme, line)
	}
}

// supportedFieldKind reports whether the field's reflect.Kind has a
// matching control in buildFormControl. Unsupported kinds (Slice, Array,
// Map, Chan, Func, Interface, etc.) are skipped during form rendering.
func supportedFieldKind(k reflect.Kind) bool {
	switch k {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// buildFormControl returns two widgets per field:
//
//   - outer is the widget added to the FormGroup as the field's
//     entry. For most controls outer == bound; for composite
//     controls (color picker = Input + preview block) outer is the
//     containing HFlex.
//   - bound is the widget that receives the EvtChange wiring back to
//     form.Update. For composite controls bound is the inner
//     editable widget (Input / Checkbox / Select), so writeback
//     fires when the user edits the actual value rather than when
//     the wrapper container fires its own (non-existent) events.
//
// The two return values collapse to the same value for non-
// composite controls; callers that don't need the distinction can
// use either.
func buildFormControl(control, id, class string, v reflect.Value, options string, theme *Theme) (outer, bound Widget) {
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
		return w, w
	case "password":
		w := NewInput(id, class, "", "", "*")
		w.SetFlag(FlagMasked, true)
		w.Apply(theme)
		w.Set(v.String())
		return w, w
	case "select":
		// NewSelect's variadic takes alternating value/text
		// pairs. The "options" tag is a comma-separated list of
		// values; we display each value as its own label.
		o := strings.Split(options, ",")
		pairs := make([]string, 0, len(o)*2)
		for _, item := range o {
			item = strings.TrimSpace(item)
			pairs = append(pairs, item, item)
		}
		w := NewSelect(id, class, pairs...)
		w.Apply(theme)
		w.Select(v.String())
		return w, w
	case "border":
		// Options come from the active theme. Two pseudo-entries
		// front the list:
		//
		//   ""     — empty string. Means "inherit from parent
		//            style cascade"; the widget will pick up
		//            whatever Border is set higher up. Displayed
		//            as "(inherit)" so the user can tell it apart
		//            from a blank line.
		//   "none" — explicit no-border. Overrides any cascaded
		//            Border, even when a parent style sets one.
		//
		// Both values are first-class to the rendering side; the
		// difference matters when the user wants to opt OUT of
		// inherited borders without leaving the field blank.
		//
		// NewSelect's variadic takes alternating value/text
		// pairs, so each option contributes two strings.
		names := theme.BorderNames()
		pairs := make([]string, 0, (len(names)+2)*2)
		pairs = append(pairs, "", "(inherit)")
		pairs = append(pairs, "none", "none")
		for _, n := range names {
			pairs = append(pairs, n, n)
		}
		w := NewSelect(id, class, pairs...)
		w.Apply(theme)
		w.Select(v.String())
		return w, w
	case "color":
		// Composite: an Input the user types into, plus a small
		// Static block whose foreground tracks the current
		// value. The block updates on every Input change so the
		// user sees the colour while typing. EvtChange wiring
		// goes on the Input (bound), not on the HFlex (outer).
		//
		// The Input's hint is fractional (-1, 1) so the HFlex
		// gives it whatever space remains after the preview
		// block's fixed 2-column footprint. With the default
		// (0, 1) hint the HFlex would treat the Input as fixed
		// at zero width and the input would render invisible.
		input := NewInput(id, class)
		input.Apply(theme)
		input.SetHint(-1, 1)
		input.Set(v.String())

		block := NewStatic("color-preview-"+id, "", "◼")
		block.Apply(theme)
		block.SetHint(2, 1)

		updateBlock := func() {
			block.SetStyle("", NewStyle("").WithForeground(input.Get()))
			Redraw(block)
		}
		updateBlock()
		input.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
			updateBlock()
			return false
		})

		hflex := NewFlex("color-flex-"+id, "", Stretch, 1)
		hflex.Apply(theme)
		_ = hflex.Add(input)
		_ = hflex.Add(block)
		return hflex, input
	default:
		w := NewInput(id, class)
		w.Apply(theme)
		w.Set(formatFieldValue(v))
		return w, w
	}
}

// formatFieldValue renders v as the string the Input control should display.
// reflect.Value.String() formats non-string kinds as "<T Value>", which is
// useless for a UI; this helper handles ints/floats/bools explicitly.
func formatFieldValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	}
	return fmt.Sprint(v.Interface())
}

// Title returns the form's title.
func (f *Form) Title() string {
	return f.title
}

package designer

import "reflect"

// Kind describes one registered widget kind: the metadata the
// inspector needs (name / group / help) plus the reflect.Type the
// kind binds to and the factory that produces a fresh, unloaded
// WidgetForm. The inspector caches a Kind per registered widget so
// metadata lookups (KindNames, picker rendering) don't allocate a
// fresh form per call.
//
// Kind values are produced once at registration time. Designer
// validates the factory at Register-time (it calls Make and asserts
// the resulting form's New() returns a widget of Type), so a
// driver-side mismatch surfaces immediately rather than panicking
// during Load.
type Kind struct {
	// Name is the human-readable kind label (e.g. "Static",
	// "Grid"). Used in the Add-child picker and in trailing
	// End() // Kind#id markers in generated source. Backed by
	// WidgetForm.Name().
	Name string

	// Group categorises the kind for the picker dialog
	// ("leaf", "container", "input", "display", …).
	Group string

	// Help is a one-line tooltip shown alongside Name in the
	// picker.
	Help string

	// Type is the concrete pointer type of the widget this kind
	// edits — e.g. reflect.TypeOf((*widgets.Static)(nil)).
	// Designer uses Type as the registry key when looking up the
	// form for a given widget.
	Type reflect.Type

	// Make returns a fresh, unloaded form instance. Callers that
	// need a form for an existing widget should go through
	// Designer.FormFor, which calls Make and then Load(w).
	Make func() WidgetForm
}

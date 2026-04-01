package zeichenwerk

// Setter is a generic interface for setting a widget's value. What constitutes
// the value depends on the widget: a List accepts []string, a Checkbox accepts
// bool, etc. Returns true if the value was accepted, false if the type did not
// match.
type Setter interface {
	// Set sets the widget's value. Returns false if the type is not accepted.
	Set(value any) bool
}

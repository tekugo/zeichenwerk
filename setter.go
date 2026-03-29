package zeichenwerk

// Simple interface for setting a widget' value.
//
// What that value is depends on the widget and the information it contains.
type Setter interface {
	// Set sets the widget' value
	Set(value any) bool
}

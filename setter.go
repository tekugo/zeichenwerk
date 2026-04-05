package zeichenwerk

// Setter is a generic interface for setting a widget's value. What constitutes
// the value depends on the widget: a List accepts []string, a Checkbox accepts
// bool, etc.
type Setter[T any] interface {
	Set(value T)
}

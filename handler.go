package zeichenwerk

// Handler represents a function that handles events.
type Handler func(source Widget, event Event, data ...any) bool

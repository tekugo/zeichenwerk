package next

// Handler represents a function that handles events.
type Handler func(event string, data ...any) bool

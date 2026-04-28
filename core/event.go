package core

// Event is the name of a notification that widgets can dispatch and subscribe
// to. It is modelled as a string so that both built-in framework events and
// arbitrary application-defined events share a single, extensible name space
// without requiring a central registry.
//
// Events flow through the widget tree via Widget.Dispatch and are consumed by
// handlers registered with Widget.On. By convention names are single
// lowercase words ("click", "focus", "change", "blur"); compound names and
// separators such as hyphens, underscores, or camelCase are avoided. The
// canonical list of built-in events lives in widgets/events.go.
type Event string

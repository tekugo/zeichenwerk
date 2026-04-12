package zeichenwerk

// Summarizer is an optional interface that widgets can implement to provide a
// concise single-line content description for Dump output. Built-in widgets
// implement it directly; custom widgets may opt in by satisfying this interface.
type Summarizer interface {
	Summary() string
}

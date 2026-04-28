package core

// Summarizer is an optional capability interface that widgets may implement
// to expose a short, single-line description of their current content.
// Debugging and diagnostic tools (notably Dump output) query for this
// interface at runtime to enrich their output: widgets that satisfy it get
// a human-readable summary printed alongside their structural metadata,
// while widgets that don't simply fall back to generic formatting.
//
// Summaries should be concise — a handful of words or a truncated preview
// — and free of embedded newlines. The method is consulted frequently, so
// implementations should be cheap and avoid allocations where practical.
type Summarizer interface {
	Summary() string
}

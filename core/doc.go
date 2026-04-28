// Package core contains the foundational building blocks of the
// zeichenwerk TUI framework: the Widget and Container interfaces, the
// rendering adapter, the styling and theming system, the event model, and
// a handful of supporting data structures (gap buffer, ring buffer,
// stack, time series).
//
// Design goals:
//
//   - Self-contained. The only non-standard dependency is the renderer
//     package, which provides the primitive drawing surface; everything
//     else — geometry, styles, themes, event dispatch — is defined here so
//     that higher layers can be developed and tested without pulling in the
//     full widget set.
//   - Interface-first. Widgets and containers are defined as interfaces so
//     that concrete implementations in sibling packages can share common
//     helpers (Find, Traverse, Layout, ...) and so that applications can
//     plug in custom widgets.
//   - Stateless helpers. Cross-cutting functionality such as tree
//     traversal is exposed as free functions rather than methods to keep
//     the interfaces minimal.
//
// Concrete widgets and containers live in the widgets package; the main UI
// control that wires everything together lives in the root package.
//
// # Error handling convention
//
// The package draws a deliberate line between two failure classes:
//
//   - Programming invariant violations — conditions that indicate a bug in
//     the caller, such as peeking an empty stack, indexing past the end of
//     a ring buffer, or asking for the min of a zero-sized series — SHOULD
//     panic. These checks exist to catch mistakes during development; they
//     are not part of the normal control flow and callers are expected to
//     prevent them with preconditions (Empty, Size, Length, ...).
//   - Runtime conditions — failures that arise from legitimate-but-invalid
//     inputs such as a nil child handed to Container.Add — MUST return an
//     error (typically a MessageCode sentinel) so callers can branch on
//     them without recovering from a panic.
//
// New types in this package should follow the same rule: reserve panics
// for "this should never happen if the caller used the API correctly", and
// use errors for everything that a reasonable program might need to
// handle at runtime.
//
// # Concurrency model
//
// Types in this package are single-goroutine unless their documentation
// explicitly says otherwise. Widgets, containers, styles, themes, gap
// buffers, stacks, and time series all assume that a single goroutine
// owns them at any given time; callers that share instances between
// goroutines must provide their own synchronisation.
//
// The sole exception is RingBuffer, which is designed for live log-style
// use: its writes (Add, Clear) are protected by an internal mutex while
// reads (Get, Size) are intentionally lock-free so that consumers do not
// contend with producers. The trade-off is that a concurrent Add may
// partially overlap a Get on the same slot; callers that need a strictly
// consistent snapshot must synchronise externally.
package core

// Package renderer provides the low-level rendering abstraction used by the
// zeichenwerk framework. It is deliberately kept independent of the core
// widget, theme, and style types so that the same drawing primitives can be
// reused by tests, alternative front-ends, or tools that need terminal
// output without pulling in the full widget tree.
//
// The package exposes three layers:
//
//   - Screen — a minimal cell-oriented interface (Put, Get, Clear, Clip,
//     Set, Translate, Flush) that any concrete back-end must implement.
//     TcellScreen is the production implementation on top of tcell v3;
//     tests typically use the in-memory core.TestScreen.
//   - Renderer — a thin wrapper around a Screen that adds higher-level
//     drawing primitives (Text, Line, Fill, ScrollbarH/V, Border helpers)
//     built on top of cell Put/Get calls. All widget Render methods
//     receive a *Renderer.
//   - NewMockScreen — a helper that wires a tcell screen to a vt.MockTerm
//     so tests can exercise rendering without a real terminal.
//
// # Coordinate model
//
// Drawing coordinates passed to Put and Get are always relative to the
// current clip origin plus translation offset: the Screen implementation
// adds those offsets before touching the underlying back-end. Widgets
// therefore work in their local coordinate space while containers set up
// clip/translate to place them on the screen.
//
// # Style state
//
// Colour and font state installed by Set persists until the next Set
// call. Drawing primitives never take a style argument; callers must set
// the desired style before a batch of Put/Text/Fill operations.
package renderer

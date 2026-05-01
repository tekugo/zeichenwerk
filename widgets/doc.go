// Package widgets contains all concrete widget and container
// implementations provided by zeichenwerk. Under normal use there is no
// need to instantiate widgets directly; use the fluent Builder API
// (root package) or the functional composition API (package compose)
// instead.
//
// # Flat package layout
//
// All widgets live in a single package rather than being grouped into
// sub-packages (widgets/input, widgets/layout, ...). This is intentional:
//
//   - Every widget embeds the same Component base and shares the same
//     styling, event, and flag machinery. Splitting would force Component
//     and a dozen helper types into a shared sub-package that every
//     widget sub-package would need to re-export, adding ceremony
//     without reducing coupling.
//   - The fluent Builder and the composition API rely on dot-imports
//     (`. "github.com/tekugo/zeichenwerk/v2/widgets"`) so that builder
//     expressions remain terse. A split would force either fragile
//     aliases or callers that import six sub-packages.
//   - Discoverability is handled by naming conventions and grouped files
//     (`bar-chart.go`, `tree-fs.go`, etc.) rather than directory
//     structure; Go tooling (go doc, IDE symbol search) makes the flat
//     namespace navigable in practice.
//
// New widgets should be added as new files in this package rather than
// spawning sub-packages; the flat layout is a deliberate part of the
// library's public shape.
package widgets

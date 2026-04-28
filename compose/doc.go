// Package compose provides a functional composition API for building
// zeichenwerk terminal UIs.
//
// Every widget is represented as an [Option] — a plain function value that
// adds a widget to its parent and applies child options recursively. Options
// can be nested, stored, and passed around freely. The active theme flows
// through the tree automatically; no global state is used.
//
// Entry points are [UI] (builds a complete UI and returns [*zeichenwerk.UI])
// and [Build] (builds a single widget subtree, useful for screen functions
// passed to [Include]).
//
// Styling, layout, and event options ([Bg], [Fg], [Font], [Border], [Padding],
// [Margin], [Hint], [On], etc.) are applied to the widget that receives them,
// not to their children, so they can appear at any position in the option list.
//
// When direct widget access is needed after construction — for example to wire
// events imperatively, populate a [Tree], or start animations — retrieve the
// widget with [zeichenwerk.Find] and call its methods directly.
package compose

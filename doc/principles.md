# Design Principles

Each principle is tagged with the layer it applies to:

- **`[core]`** — contract of the `core` package (interfaces, geometry, flags, errors).
- **`[widgets]`** — convention followed by concrete widgets in the `widgets` package.
- **`[values]`** — convention in the `values` package (data binding helpers).
- **`[renderer]`** — convention of the low-level `renderer` package (Screen, Renderer).
- **`[all]`** — cross-cutting rule that applies across every layer.

---

- `[values]` Where applicable, widgets SHOULD implement the `Setter` interface
- `[widgets]` Set(...) MUST never fire EvtChange
- `[core]` Setting the Hint(...) to other than 0, 0 MUST return these values
- `[widgets]` A Hint of 0, 0 SHOULD calculate a preferred content size for the widget
- `[core]` `Hint` values encode size on each axis as: **positive** = fixed cells, **negative** = fractional weight (the magnitude is the weight relative to siblings), **zero** = automatic (take whatever space is available / let the widget compute its preferred size). Containers interpret these three cases consistently across layouts.
- `[core]` Content size MUST never contain the margin, border or padding
- `[core]` The render area of widgets normally is not clipped, so they MAY draw outside
  their content area (e.g. border, shadow, title, grid lines)
- `[widgets]` For all timed rendering or animations, widgets SHOULD extend `ANIMATION`
- `[widgets]` For Grid to cover the available area, at least one row and column MUST use
  fractional sizes (negative values)
- `[core]` Handler functions (On(...)) are called in reverse order, last one added is
  called first, so they can consume or alter behavior of widgets.
- `[core]` `Add(nil)` MUST return `ErrChildIsNil`
- `[core]` Single-child containers MUST replace the existing child on a second `Add()` call
- `[core]` `Children()` MUST return an empty slice, never nil
- `[core]` `Layout()` MUST only compute and apply bounds via `SetBounds()`; it MUST NOT produce visual output
- `[core]` `Render()` MUST NOT modify widget state; all geometry is already resolved by the time `Render()` is called
- `[widgets]` Use `Redraw(w)` for visual-only updates; use `Relayout(w)` when the widget's `Hint()` changes
- `[core]` `FlagHidden` suppresses rendering **and** input — hidden widgets are skipped
  by mouse hit-testing (`FindAt`) and tab-order focus traversal. Layout bounds
  are unaffected: the widget still occupies its slot, so revealing it later is
  cheap and does not disturb surrounding geometry.
- `[widgets]` Widgets that accept keyboard input MUST set `FlagFocusable` in their constructor
- `[widgets]` Every widget MUST implement `Apply(theme *Theme)` and call `theme.Apply(w, w.Selector("type"))`
- `[all]` Programming invariant violations (wrong dimensions, empty stacks) SHOULD panic; runtime conditions (nil child) MUST return an error
- `[core]` Styles registered in a Theme are immutable — `Theme.Add` calls `Fix()` on every style. `With...` methods on a fixed style return a derived child style that cascades from the fixed parent rather than mutating it in place.
- `[core]` `Theme.Get` never returns nil; style lookups that match nothing fall through to `DefaultStyle`. Every `Style` accessor (`Foreground`, `Border`, `Margin`, ...) therefore returns a concrete value, so renderers never have to nil-check the result.
- `[core]` `Container.Layout` MUST lay out its own direct subtree, including recursion into nested containers. The `core.Layout(c)` helper exists as a convenience for containers that want to recurse after computing their own child bounds, but inline recursion (`child.Layout()` per container child) is equally acceptable. A `Layout` implementation MUST NOT rely on an external caller to finish the job.
- `[widgets]` Event names are single lowercase words (e.g., `click`, `focus`, `change`). Compound names and separators (`before-`, `-changed`, camelCase, etc.) are avoided — see `widgets/events.go` for the canonical list.
- `[widgets]` `Widget.Refresh()` defaults to a full-screen refresh (the base `Component.Refresh()` bubbles up to the root). Widgets MAY override `Refresh()` to call `Redraw(w)` instead when they can guarantee that only their own bounds need repainting. The choice is a trade-off: full refresh is always correct but heavier; widget-only redraw is cheaper but only safe when nothing outside the widget's bounds depends on its state.
- `[core]` `Widget.ID()` uniqueness is the caller's responsibility — the framework does not enforce it. `Find` returns the first match in depth-first order, so duplicate IDs silently yield the wrong widget. Treat IDs as you would DOM element IDs: unique within the UI.
- `[renderer]` Drawing coordinates are local. Every `Put`/`Get` coordinate is mapped to absolute screen coordinates by adding the current clip origin and translation offset inside the `Screen` implementation. Widgets therefore work in their own coordinate space; containers install `Clip`/`Translate` to position them on the screen.
- `[renderer]` Style state is persistent, not per-draw. Drawing primitives (`Put`, `Text`, `Fill`, `Line`, `ScrollbarH`/`V`, ...) never take a style argument. Callers MUST invoke `Set` (and optionally `SetUnderline`) before a batch of draws; the active style remains in effect until the next `Set` call.
- `[renderer]` Theme variables are resolved above the renderer. `renderer.Renderer.Set` forwards colour strings verbatim to the back-end, so literal colours (`"red"`, `"#ff0000"`) work directly but `$`-prefixed theme variables do not. `core.Renderer` wraps `renderer.Renderer` and performs the variable lookup before delegating — widgets rendering through `core.Renderer` may use theme variables freely.
- `[renderer]` `Screen.Clip` replaces, it does not intersect. Installing a new clip discards the previous one; there is no implicit stacking. Widget code that needs to nest clips (for example a container clipping inside its parent's clip) MUST intersect rectangles explicitly before calling `Clip`.

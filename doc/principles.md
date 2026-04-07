# Design Principles

- Where applicable, widgets SHOULD implement the `Setter` interface
- Set(...) MUST never fire EvtChange
- Setting the Hint(...) to other than 0, 0 MUST return these values
- A Hint of 0, 0 SHOULD calculate a preferred content size for the widget
- Content size MUST never contain the margin, border or padding
- The render area of widgets normally is not clipped, so they MAY draw outside
  their content area (e.g. border, shadow, title, grid lines)
- For all timed rendering or animations, widgets SHOULD extend `ANIMATION`
- For Grid to cover the available area, at least one row and column MUST use
  fractional sizes (negative values)
- Handler functions (On(...)) are called in reverse order, last one added is
  called first, so they can consume or alter behavior of widgets.
- `Add(nil)` MUST return `ErrChildIsNil`
- Single-child containers MUST replace the existing child on a second `Add()` call
- `Children()` MUST return an empty slice, never nil
- `Layout()` MUST only compute and apply bounds via `SetBounds()`; it MUST NOT produce visual output
- `Render()` MUST NOT modify widget state; all geometry is already resolved by the time `Render()` is called
- Use `Redraw(w)` for visual-only updates; use `Relayout(w)` when the widget's `Hint()` changes
- `FlagHidden` is purely visual — layout and event dispatch are unaffected
- Widgets that accept keyboard input MUST set `FlagFocusable` in their constructor
- Every widget MUST implement `Apply(theme *Theme)` and call `theme.Apply(w, w.Selector("type"))`
- Programming invariant violations (wrong dimensions, empty stacks) SHOULD panic; runtime conditions (nil child) MUST return an error

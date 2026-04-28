# Events

Events are how user input and widget state changes are routed through the
widget tree. Widgets call `Dispatch(source, event, data...)`; handlers
register via `widget.On(event, handler)`.

## Handler signature

```go
type Handler func(source Widget, event Event, data ...any) bool
```

`source` is the widget that fired the event. `data` is event-specific —
sometimes a string, sometimes an `int` index, sometimes a
`*tcell.EventKey`. Return `true` to stop propagation.

## Event flow

1. **Generation** — user input (keyboard, mouse) or programmatic state
   change calls `Dispatch`.
2. **Bubbling** — the event propagates from the source widget up through
   its parent chain. Handlers on intermediate containers are invoked
   along the way; siblings never see the event.
3. **Termination** — propagation stops when a handler returns `true` or
   the chain reaches the root. (`Ctrl-Q`, `Ctrl-D`, `Tab`/`Shift-Tab`,
   `Esc` are handled by `UI.Handle` after the bubble completes — they
   don't go through the handler list.)

Multiple handlers on the same widget run in **reverse registration
order**: the most recently added runs first and may consume the event
before earlier handlers see it.

## Canonical event list

All constants live in [`widgets/events.go`](../widgets/events.go).

| Constant      | String value | Typical data | Description |
|---------------|--------------|--------------|-------------|
| `EvtAccept`   | `"accept"`   | `string`     | User accepted a suggested or pending value (e.g. Tab in a Typeahead) |
| `EvtActivate` | `"activate"` | varies       | User activated an item (Enter, double-click) |
| `EvtBlur`     | `"blur"`     | —            | Widget lost keyboard focus |
| `EvtChange`   | `"change"`   | varies       | Value or state changed (Input keystroke, Checkbox toggle, …) |
| `EvtClick`    | `"click"`    | —            | Mouse button-1 single click |
| `EvtClose`    | `"close"`    | —            | Popup layer is about to close (`UI.Close`) — last chance to clean up |
| `EvtEnter`    | `"enter"`    | `string`     | Enter pressed in an Input |
| `EvtFocus`    | `"focus"`    | —            | Widget gained keyboard focus |
| `EvtHide`     | `"hide"`     | —            | Widget became hidden (Switcher, Grow, …) |
| `EvtHover`    | `"hover"`    | `*tcell.EventMouse` | Mouse over widget |
| `EvtKey`      | `"key"`      | `*tcell.EventKey`   | Unhandled key event bubbling up |
| `EvtMode`     | `"mode"`     | `string`     | Editing mode changed (Canvas) |
| `EvtMouse`    | `"mouse"`    | `*tcell.EventMouse` | Raw mouse event |
| `EvtMove`     | `"move"`     | `int, int`   | Highlight or cursor position changed |
| `EvtPaste`    | `"paste"`    | `string`     | Text pasted |
| `EvtSelect`   | `"select"`   | varies       | Highlighted/selected item changed (before activation) |
| `EvtShow`     | `"show"`     | —            | Widget became visible |

## Per-widget data payloads

When the table above lists "varies", the exact `data` arguments are:

### `EvtActivate`

| Widget | Data | Notes |
|--------|------|-------|
| `Button` | `int` (always `0`) | Activate / Enter / Space / click |
| `List` | `int` | Item index |
| `Tabs` | `int` | Tab index (also fired on letter shortcut) |
| `Tree` | `*TreeNode` | Activated node |
| `Table` | `int, []string` | Row index, full row data |
| `BarChart` | `int` | Category index |
| `Breadcrumb` | `int` | Segment index |
| `Tiles` | `int` | Tile index |
| `Combo` | `string` | Confirmed value |
| `Deck` | `int` | Item index |
| `Typewriter` | `bool` (always `true`) | Animation completed |

### `EvtSelect`

| Widget | Data | Notes |
|--------|------|-------|
| `List` | `int` | Highlighted index |
| `Tree` | `*TreeNode` | Newly selected node |
| `Table` | `int, int` | Row index, column index (column is `-1` when not in cell-nav mode) |
| `BarChart` | `int` | Category index |
| `Breadcrumb` | `int` | Segment index |
| `Tiles` | `int` | Tile index |
| `Deck` | `int` | Item index |

### `EvtChange`

| Widget | Data | Notes |
|--------|------|-------|
| `Input` | `string` | Current text after the change |
| `Typeahead` / `Filter` | `string` | Same as Input |
| `Combo` | `string` | Current input text while popup is open |
| `Checkbox` | `bool` | New checked state |
| `Select` | `string` | New selected value |
| `Tabs` | `int` | New highlighted tab |
| `Editor` | — | No data; query content with `editor.Text()` |
| `Tree` | `*TreeNode` | Node expanded/collapsed |
| `Canvas` | — | Pixel/cell modified |
| `Typewriter` | `bool` (always `true`) | Reveal phase finished |

## Typed helpers

`widgets/event-helper.go` ships small wrappers that unwrap `data[0]` for
common cases. They take care of type-asserting and falling through when
the assertion fails.

```go
widgets.OnActivate(w, func(index int) bool       { … })   // EvtActivate (int payload)
widgets.OnSelect  (w, func(index int) bool       { … })   // EvtSelect   (int payload)
widgets.OnChange  (w, func(value string) bool    { … })   // EvtChange   (string payload)
widgets.OnAccept  (w, func(value string) bool    { … })   // EvtAccept
widgets.OnEnter   (w, func(value string) bool    { … })   // EvtEnter
widgets.OnKey     (w, func(*tcell.EventKey)  bool { … })  // EvtKey
widgets.OnMouse   (w, func(*tcell.EventMouse) bool { … }) // EvtMouse
widgets.OnHide    (w, func() bool                { … })   // EvtHide
widgets.OnShow    (w, func() bool                { … })   // EvtShow
```

The helper picks the **first** `data` element and asserts to its expected
type — your handler is silently skipped if the type doesn't match. That's
why `OnChange(checkbox, …)` (string handler) does *not* fire for
`Checkbox` (which dispatches `bool`); use the raw form when the payload
type is non-string:

```go
checkbox.On(widgets.EvtChange, func(_ Widget, _ Event, data ...any) bool {
    if checked, ok := data[0].(bool); ok {
        // …
    }
    return true
})
```

Similarly, `OnActivate` only catches `int`-payload activates — for
`Combo` (which sends `string`) and `Tree` / `Table` (which send richer
payloads), use the raw form and assert.

## Tips

- A handler on a *container* catches everything that bubbles up from its
  descendants and wasn't consumed earlier. This is the cleanest place for
  app-wide keyboard shortcuts.
- A handler that mutates widget state should call `widgets.Redraw(self)`
  if the change is purely visual, or `widgets.Relayout(self)` if it
  affects size. Many widgets (`List`, `Static`, `Editor`) call the right
  helper from their own setters; you only need these helpers when you
  reach into widget internals directly.
- Returning `true` consumes the event for **all** ancestors, not just
  the current handler. If you genuinely want to handle but allow others
  to also see, return `false`.

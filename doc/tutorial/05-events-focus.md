# 5. Events & focus

You build the tree, then you wire behaviour. This chapter is about that
second half.

## Handlers

Every widget exposes:

```go
widget.On(event Event, handler Handler)
```

with

```go
type Handler func(source Widget, event Event, data ...any) bool
```

`source` is the widget that fired the event. `data` is event-specific —
sometimes a string, sometimes an int index, sometimes a `*tcell.EventKey`.
The return value controls **propagation**: `true` consumes the event,
`false` lets it bubble up to the parent container.

The raw form is fine but tedious — every handler ends up type-asserting
`data[0]`. The `widgets` package ships typed helpers that unwrap the data
for you:

| Helper | Wraps | Handler signature |
|--------|-------|-------------------|
| `widgets.OnActivate(w, fn)` | `EvtActivate` | `func(index int) bool` |
| `widgets.OnSelect(w, fn)` | `EvtSelect` | `func(index int) bool` |
| `widgets.OnChange(w, fn)` | `EvtChange` (string-valued) | `func(value string) bool` |
| `widgets.OnEnter(w, fn)` | `EvtEnter` | `func(value string) bool` |
| `widgets.OnAccept(w, fn)` | `EvtAccept` | `func(value string) bool` |
| `widgets.OnKey(w, fn)` | `EvtKey` | `func(*tcell.EventKey) bool` |
| `widgets.OnMouse(w, fn)` | `EvtMouse` | `func(*tcell.EventMouse) bool` |
| `widgets.OnHide(w, fn)` | `EvtHide` | `func() bool` |
| `widgets.OnShow(w, fn)` | `EvtShow` | `func() bool` |

> **Caveat:** the helpers use a fixed data type. `Checkbox` dispatches
> `EvtChange` with a `bool`, not a string, so `OnChange` won't catch it —
> use the raw `widget.On(EvtChange, …)` and assert `data[0].(bool)`.

## The events that matter

| Event | Fires on… |
|-------|-----------|
| `EvtActivate` | Button click, List/Tree row Enter, Tab activated |
| `EvtSelect` | Highlighted row changed in List, Tree, Table |
| `EvtChange` | Input keystroke, Checkbox toggle, Tree expand/collapse |
| `EvtEnter` | Enter pressed in an Input |
| `EvtAccept` | User accepts a Typeahead suggestion |
| `EvtKey` | Unhandled key bubbles up — your last chance to catch it |
| `EvtMouse` | Raw mouse event |
| `EvtFocus` / `EvtBlur` | Widget gained / lost keyboard focus |
| `EvtShow` / `EvtHide` | Switcher pane became visible / hidden |

For the full canonical list see
[`widgets/events.go`](../../widgets/events.go).

## Propagation: reverse order, consumable

Two rules:

1. Multiple handlers for the same event run in **reverse registration
   order** — last one added runs first. This lets a downstream handler
   consume an event before earlier ones see it.
2. If any handler returns `true`, propagation stops — the event does not
   bubble to the parent container.

A common pattern is to register a key handler on a *container* to provide
shortcuts for everything inside it:

```go
widgets.OnKey(rootContainer, func(ev *tcell.EventKey) bool {
    switch ev.Key() {
    case tcell.KeyCtrlS:
        save()
        return true
    case tcell.KeyCtrlO:
        open()
        return true
    }
    return false
})
```

Children's own handlers run first; only unhandled keys reach the container.

## Example

[`examples/05-events/main.go`](examples/05-events/main.go) wires three
widgets to a single status line:

```go
echo := core.MustFind[*widgets.Static](ui, "echo")
status := core.MustFind[*widgets.Static](ui, "status")

// Live keystroke echo from an Input.
widgets.OnChange(core.MustFind[*widgets.Input](ui, "name"), func(value string) bool {
    echo.Set("→ " + value)
    return true
})

// Selection vs. activation in a List.
colours := core.MustFind[*widgets.List](ui, "colours")
widgets.OnSelect(colours, func(i int) bool {
    status.Set(fmt.Sprintf("selected: %s", colours.Items()[i]))
    return true
})
widgets.OnActivate(colours, func(i int) bool {
    status.Set(fmt.Sprintf("activated: %s !", colours.Items()[i]))
    return true
})
```

Run it:

```bash
go run ./doc/tutorial/examples/05-events
```

Tab moves between the input, the list, and the quit button. Inside the
list, ↑/↓ fires `EvtSelect`; Enter fires `EvtActivate`. They're different
events on purpose — selection is "hover with the keyboard," activation is
"commit."

## Focus

The framework manages focus for you:

- Widgets that accept input (`Input`, `Editor`, `Button`, `List`,
  `Checkbox`, `Select`, `Combo`, …) set `core.FlagFocusable` in their
  constructor.
- Tab / →/ ↓ moves to the next focusable widget; Shift-Tab / ← / ↑ moves
  to the previous.
- Focus wraps around at the ends.

You can drive focus programmatically:

```go
ui.Focus(core.MustFind[*widgets.Input](ui, "name"))   // jump to widget
ui.SetFocus("first")    // first focusable
ui.SetFocus("next")     // next in order
```

To exclude a focusable widget from Tab traversal but still allow
programmatic focus, set `core.FlagSkip`:

```go
input := core.MustFind[*widgets.Input](ui, "search")
input.SetFlag(core.FlagSkip, true)
```

To hide a widget completely (no rendering, no input, skipped by focus):

```go
input.SetFlag(core.FlagHidden, true)
```

The slot in the layout is preserved, so revealing it later doesn't shuffle
surrounding widgets.

## Updating after an event

When your handler changes widget content, the framework needs to know what
to repaint. Three options, cheapest first:

| Helper | Use when |
|--------|----------|
| `widgets.Redraw(w)` | Visual change only; the widget's `Hint()` didn't change. Repaints just `w`. |
| `widgets.Relayout(w)` | The change affects size — text grew, list got more items. Re-runs layout starting at the nearest ancestor that needs it. |
| `ui.Refresh()` | Full screen repaint. Always correct, rarely needed. |

Many widgets call the right helper themselves. For instance, `Static.Set`
queues a refresh; `Input` keystrokes redraw the input automatically. You
only need to call these helpers when you mutate widget state directly
(swapping a list's items, reaching into a custom widget's fields).

→ [Next: Widget tour](06-widget-tour.md)

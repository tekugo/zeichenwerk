# Events

Events are the primary mechanism for handling user interactions and widget state changes. Widgets emit events via `Dispatch(source, event, data...)` and register handlers via `On(event, handler)`.

## Event Flow

1. **Generation**: User input (keyboard, mouse) or widget state changes trigger events.
2. **Dispatch**: The widget calls `Dispatch`, which invokes registered handlers.
3. **Handling**: Handlers are called in reverse registration order (newest first). Returning `true` stops propagation; `false` continues.

## Events Reference

| Event | Widget(s) | Data | Description |
|-------|-----------|------|-------------|
| `"activate"` | List, Table, Tabs | see below | Item activated via Enter key |
| `"change"` | Canvas, Checkbox, Editor, Input, Select, Tabs | see below | Value or state changed |
| `"click"` | Button | — | Button activated via Enter, Space, or mouse |
| `"enter"` | Input | `string` | Enter key pressed |
| `"hide"` | Switcher | — | Pane became hidden |
| `"key"` | all widgets | `*tcell.EventKey` | Keyboard event dispatched to focused widget |
| `"mode"` | Canvas | `string` | Editing mode changed |
| `"mouse"` | all widgets | `*tcell.EventMouse` | Mouse event dispatched to widget under cursor |
| `"move"` | Canvas | `x, y int` | Cursor moved |
| `"select"` | List, Table | see below | Highlighted item changed |
| `"show"` | Switcher | — | Pane became visible |

### Event Data by Widget

**`"activate"`**
- `List`: `int` (index)
- `Table`: `int` (row index), `[]string` (row data)
- `Tabs`: `int` (tab index)

**`"change"`**
- `Canvas`: no data
- `Checkbox`: `bool` (checked state)
- `Editor`: no data
- `Input`: `string` (current text)
- `Select`: `string` (selected value)
- `Tabs`: `int` (highlighted tab index)

**`"select"`**
- `List`: `int` (index)
- `Table`: `int` (row index), `[]string` (row data)

## Helper Functions

Convenience wrappers in `helper.go` that unwrap event data into typed handler signatures:

- `OnActivate(widget, func(Widget, int) bool)` — `"activate"` events; receives item index
- `OnChange(widget, func(Widget, string) bool)` — `"change"` events; receives new value as string (Input, Select)
- `OnKey(widget, func(Widget, *tcell.EventKey) bool)` — `"key"` events
- `OnMouse(widget, func(Widget, *tcell.EventMouse) bool)` — `"mouse"` events
- `OnSelect(widget, func(Widget, int) bool)` — `"select"` events; receives item index

## Notes

- Return `true` from a handler to stop propagation, `false` to continue.
- Use type assertions to access event data directly: `text := data[0].(string)`.
- `OnChange` expects `string` in `data[0]`; it will not fire for Checkbox (`bool`) or Tabs (`int`) — handle those with `widget.On("change", ...)` directly.

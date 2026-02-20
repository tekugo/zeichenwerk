# Events

Events are the primary mechanism for handling user interactions and widget state changes. All widgets can emit events via `Dispatch(event, data...)` and handle events via `On(event, handler)`.

## Event Flow

1. **Generation**: User input (keyboard, mouse) or widget actions trigger events.
2. **Dispatch**: Events are sent to a target widget, then bubble up the parent chain.
3. **Handling**: Handlers are called in registration order. Returning `true` stops propagation; `false` continues.

## Events Reference

| Event | Widget(s) | Data | Description |
|-------|-----------|------|-------------|
| `activate` | Button, List, Table, Tabs | Varies by widget (see below) | Item/widget was activated (Enter key, click, or double-click equivalent). |
| `change` | Checkbox, Editor, Input, Tabs | Varies by widget (see below) | Widget value or selection changed. |
| `click` | Button | none | Button was activated via mouse click or keyboard. |
| `enter` | Input | `string` (current text) | Enter key pressed in input field. |
| `hover` | All widgets (system) | `*tcell.EventMouse` | Mouse cursor moved over widget. |
| `key` | All widgets (system) | `*tcell.EventKey` | Keyboard event dispatched to focused widget. |
| `paste` | All widgets (system) | `*tcell.EventPaste` | Paste event (terminal paste) received. |
| `select` | List, Table | Varies by widget (see below) | Selection changed to a different item/row. |

### Event Data by Widget

- **`activate`**:
  - `Button`: no data
  - `List`: `int` (index)
  - `Table`: `int` (row index), `[]string` (row data)
  - `Tabs`: `int` (selected tab index)
- **`change`**:
  - `Checkbox`: `bool` (checked state)
  - `Editor`: no data
  - `Input`: `string` (current text)
  - `Tabs`: `int` (highlighted tab index)
- **`select`**:
  - `List`: `int` (index)
  - `Table`: `int` (row index), `[]string` (row data)

## Helper Functions

Convenience wrappers in `helper.go`:

- `OnKey(widget, handler)` - Registers for `key` events with typed handler.
- `OnMouse(widget, handler)` - Registers for mouse-related events; note that the framework currently dispatches `hover` instead of `mouse`.

## Notes

- Events bubble from target widget up through parents until handled.
- Use type assertions to access event data: `text := data[0].(string)`.
- Return `true` to stop propagation, `false` to continue.
- All widgets implement `Widget` and support `On()` for any event, but the events listed above are the ones actually dispatched by specific widgets.

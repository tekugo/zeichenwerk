# Button

Clickable button with a text label.

**Constructor:** `NewButton(id, class, text string) *Button`

## Methods

- `Activate()` — programmatically triggers the activate event
- `Text() string` — current label text
- `Set(value string)` — replace label text (does not fire `EvtChange`)

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `int` (always `0`) | Enter, Space, or mouse click |

## Notes

Flags: `"focusable"`. States: `"pressed"`, `"focused"`, `"hovered"`.

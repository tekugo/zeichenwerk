# Input

Single-line text input field.

**Constructor:** `NewInput(id, class string, params ...string) *Input`

Params: `[0]` initial text, `[1]` placeholder, `[2]` mask character

## Methods

- `Clear()` — removes all text
- `Delete()` — backspace (deletes before cursor)
- `DeleteForward()` — delete (removes at cursor)
- `End()` — moves cursor to end
- `Insert(ch string)` — inserts character at cursor
- `Left()` — moves cursor left
- `Right()` — moves cursor right
- `SetMask(mask string)` — enables password masking
- `SetText(text string)` — replaces entire content
- `Start()` — moves cursor to beginning
- `Text() string` — returns current text

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `string` | Text modified |
| `"enter"` | `string` | Enter key pressed |

## Notes

Flags: `"focusable"`, `"masked"`, `"readonly"`

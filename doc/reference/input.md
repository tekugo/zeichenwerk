# Input

Single-line text input.

**Constructor:** `NewInput(id, class string, params ...string) *Input`

Params: `[0]` initial text, `[1]` placeholder, `[2]` mask character.

## Methods

- `Get() string` — current text
- `Set(text string)` — replace text (does not fire `EvtChange`)
- `Insert(ch string)` — insert text at cursor
- `Delete()` — backspace (delete before cursor)
- `DeleteForward()` — delete at cursor
- `Clear()` — remove all text
- `Left() / Right()` — move cursor by one rune
- `Start() / End()` — jump cursor to beginning/end
- `SetMask(mask string)` — set the mask character (used when `FlagMasked` is set)

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `string` | Text was modified |
| `"enter"` | `string` | Enter key pressed; data is the current value |

## Notes

Flags: `"focusable"` (default), `"masked"` (display mask character instead of real characters; useful for passwords), `"readonly"` (disable editing without removing focusability).

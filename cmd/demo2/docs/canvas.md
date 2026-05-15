# Canvas

Low-level pixel buffer for custom rendering with modal editing.

**Constructor:** `NewCanvas(id, class string, pages, width, height int) *Canvas`

## Methods

- `Cell(x, y int) *Cell` — returns cell at coordinates
- `Clear()` — clears all cells
- `Cursor() (x, y int, style string)` — returns cursor position
- `Fill(ch string, style *Style)` — fills entire buffer with character and style
- `Mode() string` — returns current mode
- `Resize(pages, rows, columns int)` — changes buffer dimensions
- `SetCell(x, y int, ch string, style *Style)` — sets cell content
- `SetCursor(x, y int)` — moves cursor to position
- `SetMode(mode string)` — sets current mode
- `SetPage(page int)` — sets current page
- `Size() (width, height int)` — returns buffer dimensions

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | — | Cell content modified |
| `"mode"` | `string` | Mode changed |
| `"move"` | `x, y int` | Cursor moved |

## Modes

`ModeNormal`, `ModeInsert`, `ModeDraw`, `ModeVisual`, `ModeCommand`, `ModePresent`

## Notes

Flags: `"focusable"`

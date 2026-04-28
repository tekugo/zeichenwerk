# Tiles

Wrapping grid of fixed-size tiles. Each tile is drawn by a user-supplied `ItemRender` callback. The number of columns is computed at render time from the content width: `cols = max(1, contentWidth / tileWidth)`.

**Constructor:** `NewTiles(id, class string, render ItemRender, tileWidth, tileHeight int) *Tiles`

`tileWidth` and `tileHeight` must be ≥ 1 (panics otherwise).

```go
type ItemRender func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool)
```

## Methods

- `Items() []any` — current items
- `SetItems(items []any)` — replace all items
- `SetDisabled(indices []int)` — mark item indices as non-selectable (skipped by navigation)
- `Selected() int` — current highlighted index (-1 if empty)
- `Select(index int)` — set highlighted index; fires `EvtSelect`
- `Move(dr, dc int)` — move selection by row/column delta
- `First() / Last()` — jump to first/last enabled item
- `PageUp() / PageDown()` — scroll by viewport height

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"select"` | `int` | Highlighted item index changed |
| `"activate"` | `int` | Item activated via Enter or double-click |

## Notes

Flags: `"focusable"`.

Navigation wraps between rows in reading order: moving right past the last column advances to the first column of the next row; moving up/down keeps the current column. A vertical scrollbar is drawn when the items don't fit; toggle via the internal `scrollbar` field (set during construction; no public setter).

`SetItems` accepts any slice — `[]string`, `[]MyStruct`, etc. — passed through `[]any`. The render callback gets the raw `data any` and is responsible for type-asserting.

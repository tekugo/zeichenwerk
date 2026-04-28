# Tiles

A scrollable 2D grid widget where every item occupies a fixed
`tileWidth × tileHeight` cell. Items fill left-to-right, wrapping to the next
row when the viewport width is exhausted. Rendering is delegated to a
`DeckRenderFunc` (same type as `Deck`) — one function, repositioned per
visible cell.

## Layout

```
┌──────────────────────────────────────┐
│ [Item 0] [Item 1] [Item 2] [Item 3]  │
│ [Item 4] [Item 5] [Item 6] [Item 7]  │
│ [Item 8] [Item 9]                    │
└──────────────────────────────────────┘
```

The number of columns is computed at render time:

```
cols = max(1, contentWidth / tileWidth)
```

It recalculates whenever the widget is resized. Rows with fewer items than
`cols` leave blank space at the right edge (no stretching).

## Structure

```go
type Tiles struct {
    Component
    render     DeckRenderFunc // Same render function type as Deck
    items      []any          // Data items
    disabled   []int          // Non-selectable item indices
    tileWidth  int            // Fixed cell width (>= 1)
    tileHeight int            // Fixed cell height (>= 1)
    index      int            // Highlighted item index (-1 if empty)
    offsetRow  int            // First visible row index (0-based)
    scrollbar  bool
}
```

## Constructor

```go
func NewTiles(id, class string, render DeckRenderFunc, tileWidth, tileHeight int) *Tiles
```

- Both dimensions must be >= 1; panics otherwise.
- `index = -1`, `scrollbar = true`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

## Derived values (computed, not stored)

```go
func (t *Tiles) cols() int   // contentWidth / tileWidth, min 1
func (t *Tiles) rows() int   // ceil(len(items) / cols)
func (t *Tiles) row(index int) int  // index / cols
func (t *Tiles) col(index int) int  // index % cols
```

These recalculate from current bounds on each call. `cols` depends on
`Content()`, so it is only valid after layout.

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetItems(items []any)` | Replaces all items; resets index to 0 (or -1 if empty), `offsetRow` to 0 |
| `Items() []any` | Returns the current items slice |
| `SetDisabled(indices []int)` | Replaces the disabled index list |

### Navigation

| Method | Description |
|--------|-------------|
| `Select(index int)` | Highlights item; adjusts scroll; dispatches `EvtSelect` |
| `Selected() int` | Returns highlighted index (-1 if none) |
| `Move(dr, dc int)` | Moves highlight by `dr` rows and `dc` columns, skipping disabled items, clamping to bounds |
| `First()` | Highlights first enabled item |
| `Last()` | Highlights last enabled item |
| `PageUp()` / `PageDown()` | Moves by the number of fully visible rows |

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `←` | `Move(0, -1)` |
| `→` | `Move(0, +1)` |
| `↑` | `Move(-1, 0)` |
| `↓` | `Move(+1, 0)` |
| `PgUp` / `PgDn` | `PageUp()` / `PageDown()` |
| `Home` / `End` | `First()` / `Last()` |
| `Enter` | Dispatch `EvtActivate` with current index |

`Move` in the column direction wraps between rows: moving right past the last
column of a row advances to the first column of the next row, and vice versa.
This matches the natural reading-order traversal of a grid.

## Mouse interaction

Map click position to item index:

```
clickedCol = (mouseX - contentX) / tileWidth
clickedRow = (mouseY - contentY) / tileHeight + offsetRow
clickedIndex = clickedRow * cols + clickedCol
```

If `clickedIndex` is valid and not disabled: select and dispatch `EvtSelect`.
Second click on already-selected item dispatches `EvtActivate`.

## Scrolling

Scrolling is row-based. `offsetRow` is the index of the first visible row.

`adjust()` ensures the highlighted item's row stays in the viewport:

```
visibleRows = contentHeight / tileHeight
if row(index) < offsetRow:
    offsetRow = row(index)
if row(index) >= offsetRow + visibleRows:
    offsetRow = row(index) - visibleRows + 1
```

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"select"` | `int` | Highlighted item index changed |
| `"activate"` | `int` | Enter pressed or item clicked twice |

## Rendering

```go
func (t *Tiles) Render(r *Renderer)
```

1. `t.Component.Render(r)` — draws background and border.
2. Compute `(cx, cy, cw, ch)` and derive `cols`, `visibleRows`, `tw`.
   Reserve the right column for the scrollbar when needed:
   `tw = cw - 1` if `scrollbar && t.rows() * tileHeight > ch`, else `tw = cw`.
3. Recalculate `cols = max(1, tw / tileWidth)`.
4. For each visible row `r` in `offsetRow … offsetRow + visibleRows - 1`:
   For each column `c` in `0 … cols-1`:
   - `itemIndex = r * cols + c`; skip if `>= len(items)`.
   - `slotX = cx + c * tileWidth`.
   - `slotY = cy + (r - offsetRow) * tileHeight`.
   - `r.Clip(slotX, slotY, tileWidth, tileHeight)`.
   - Call `t.render(renderer, slotX, slotY, tileWidth, tileHeight, itemIndex, t.items[itemIndex], itemIndex == t.index)`.
   - `r.Clip(cx, cy, tw, ch)` — restore.
5. Draw scrollbar if needed:
   `r.ScrollbarV(cx+tw, cy, ch, t.offsetRow*t.tileHeight, t.rows()*t.tileHeight)`.

### Partial last row

Items that don't fill the last row leave their cells empty (the background fill
from `Component.Render` covers them). No special handling needed.

### Column count change on resize

When the terminal is resized, `cols` changes automatically because it is
derived from `Content()`. `offsetRow` is preserved. `adjust()` is called after
resize to re-clamp the offset if the highlighted row would become invisible.

## Hint

```go
func (t *Tiles) Hint() (int, int)
```

- Width: manually set hint, or `tileWidth * defaultCols` where `defaultCols`
  is a configurable field (default 4). Gives the parent layout a reasonable
  starting width.
- Height: `t.rows() * tileHeight` — shows all items without scrolling if space
  allows.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"tiles"` | Container background and border |
| `":focused"` | Focused state (border highlight) |
| `":disabled"` | Disabled state |

Item-level styling is entirely the render function's responsibility.

## Comparison with Deck

| | `Deck` | `Tiles` |
|--|--------|---------|
| Layout | 1D vertical | 2D grid |
| Scroll | by item | by row |
| Navigation | `↑`/`↓` | `←`/`→`/`↑`/`↓` |
| Columns | 1 | `contentWidth / tileWidth` |
| Render function | `DeckRenderFunc` (shared type) | `DeckRenderFunc` (shared type) |

## Implementation plan

1. **`tiles.go`** — new file
   - Define `Tiles` struct and `NewTiles`.
   - Implement `cols`, `rows`, `row`, `col` helper methods.
   - Implement `SetItems`, `Items`, `SetDisabled`.
   - Implement `Select`, `Move`, `First`, `Last`, `PageUp`, `PageDown`,
     `adjust` (mirrors `Deck` but row-aware).
   - Implement `handleKey`, `handleMouse`, `Apply`, `Hint`, `Render`.

2. **`builder.go`** — add `Tiles` method
   ```go
   func (b *Builder) Tiles(id string, render DeckRenderFunc, tileWidth, tileHeight int) *Builder
   ```

3. **Tests** — `tiles_test.go`
   - `cols` recalculates correctly for different content widths.
   - `Move(0, +1)` wraps from last column of row N to first column of row N+1.
   - `Move(-1, 0)` on the first row of the viewport triggers scroll.
   - Column count change on resize preserves `offsetRow` and re-clamps.
   - Render clips each tile independently.
   - `SetItems` with empty slice sets `index = -1`.

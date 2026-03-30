# Deck

A scrollable list widget where every item occupies a fixed number of rows.
Rendering is delegated to a caller-supplied **render function** that receives
the renderer, slot bounds, item index, data, and selection state — no per-item
widget allocations. The uniform item height makes it suited for rich items such
as style pickers, contact cards, or navigation menus where each entry needs
multiple lines.

## Render function

```go
type ItemRender func(r *Renderer, x, y, w, h, index int, data any, selected bool)
```

Called once per visible slot immediately before the slot is drawn. `selected`
is `true` when the slot is the currently highlighted item. The function is
responsible for all visual content within the slot; `Deck` clips each slot
before calling it so overflow into adjacent slots is not possible.

The caller closes over any context needed (theme references, shared state, etc.).
For access to the deck's own styles, the function can close over the `*Deck`
and call `deck.Style(selector)`.

## Structure

```go
type Deck struct {
    Component
    render     ItemRender // Render function for each slot
    items      []any          // Data items, one per slot
    disabled   []int          // Indices of non-selectable items
    itemHeight int            // Fixed row count per item slot (>= 1)
    index      int            // Currently highlighted item index (-1 if empty)
    offset     int            // Index of first visible item
    scrollbar  bool           // Whether to draw a vertical scrollbar
}
```

## Constructor

```go
func NewDeck(id, class string, render ItemRender, itemHeight int) *Deck
```

- `itemHeight` must be >= 1; panics otherwise.
- `index = -1`, `scrollbar = true`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetItems(items []any)` | Replaces all items; resets `index` to 0 (or -1 if empty) and `offset` to 0; calls `Refresh()` |
| `Items() []any` | Returns the current items slice |
| `SetDisabled(indices []int)` | Replaces the disabled index list |

### Navigation

| Method | Description |
|--------|-------------|
| `Select(index int)` | Highlights item at index; adjusts scroll; dispatches `EvtSelect` |
| `Selected() int` | Returns the current highlight index (-1 if none) |
| `Move(count int)` | Moves highlight by count, skipping disabled items; clamps to bounds |
| `First()` | Highlights first enabled item |
| `Last()` | Highlights last enabled item |
| `PageUp()` / `PageDown()` | Moves by the number of fully visible slots |

## Keyboard interaction

Mirrors `List` exactly — the larger item height does not change the key model
since navigation works in item units.

| Key | Behaviour |
|-----|-----------|
| `↑` / `↓` | `Move(±1)` — skips disabled items |
| `PgUp` / `PgDn` | `PageUp()` / `PageDown()` |
| `Home` / `End` | `First()` / `Last()` |
| `Enter` | Dispatch `EvtActivate` with current index |

## Mouse interaction

A click at row `mouseY` within the content area maps to item index:

```
clickedItem = (mouseY - contentY) / itemHeight + offset
```

If `clickedItem` is valid and not disabled, select it and dispatch `EvtSelect`.
A second click on the already-selected item dispatches `EvtActivate`.

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"select"` | `int` | Highlighted item index changed |
| `"activate"` | `int` | Enter pressed or item clicked twice |

## Rendering

```go
func (d *Deck) Render(r *Renderer)
```

1. `d.Component.Render(r)` — draws background and border.
2. Compute content area `(cx, cy, cw, ch)`.
3. Reserve the rightmost column for the scrollbar when it will be shown:
   `tw = cw - 1` if `scrollbar && len(d.items)*d.itemHeight > ch`, else `tw = cw`.
4. Compute visible slot count: `slots = ch / d.itemHeight`.
5. For each slot `s` in `0 … slots-1`:
   - `itemIndex = d.offset + s`; stop if `itemIndex >= len(d.items)`.
   - `slotY = cy + s * d.itemHeight`.
   - `r.Clip(cx, slotY, tw, d.itemHeight)` — confine output to slot.
   - Call `d.render(r, cx, slotY, tw, d.itemHeight, itemIndex, d.items[itemIndex], itemIndex == d.index)`.
   - `r.Clip(cx, cy, tw, ch)` — restore content-area clip.
6. Draw scrollbar if needed:
   `r.ScrollbarV(cx+tw, cy, ch, d.offset*d.itemHeight, len(d.items)*d.itemHeight)`.
   Row-based units keep the thumb proportional regardless of item height.

### Partial last slot

If `ch` is not an exact multiple of `itemHeight`, the bottom partial slot is
not rendered. The clip in step 5 would contain any overflow, but skipping it
is cleaner.

## Hint

```go
func (d *Deck) Hint() (int, int)
```

- Width: the manually set hint width, or 0 if unset (parent decides).
- Height: `len(d.items) * d.itemHeight` — requests exactly enough rows to show
  all items without scrolling. The parent layout clips this to available space.

## Styling selectors

The render function owns all item-level visual decisions. `Deck` exposes only
container-level selectors.

| Selector | Applied to |
|----------|-----------|
| `"deck"` | Background and border of the deck container |
| `":focused"` | Container-level focused state (border highlight etc.) |
| `":disabled"` | Container when all interaction is disabled |

For render functions that want to follow the active theme, close over the
`*Deck` and call `deck.Style(selector)` to read resolved style values.

## Example: style-picker

```go
type StyleItem struct {
    label string
    color string // hex, e.g. "#1a1b2e"
}

items := []any{
    StyleItem{"Foreground", "#1a1b2e"},
    StyleItem{"Background", "#c0caf5"},
}

deck := NewDeck("style-picker", "", func(r *Renderer, x, y, w, h, index int, data any, selected bool) {
    item := data.(StyleItem)
    if selected {
        r.Set("highlight-fg", "highlight-bg", "")
    } else {
        r.Set("fg", "bg", "")
    }
    r.Text(x, y,   item.label, w)          // row 0: label
    r.Set(item.color, item.color, "")
    r.Fill(x, y+1, w-10, 1, " ")           // row 1: color bar
    r.Set("fg", "bg", "")
    r.Text(x+w-9, y+1, item.color, 9)      // row 1: hex code
    r.Text(x, y+2, " ", w)                 // row 2: spacer
}, 3)

deck.SetItems(items)
```

## Implementation Plan

1. **`deck.go`** — new file
   - Define `ItemRender` type and `Deck` struct.
   - Implement `NewDeck`, `SetItems`, `Items`, `SetDisabled`.
   - Implement `Select`, `Move`, `First`, `Last`, `PageUp`, `PageDown`
     and the internal `skip` and `adjust` helpers (mirrors `List`).
   - Implement `handleKey`, `handleMouse`, `Apply`, `Hint`, `Render`.

2. **`builder.go`** — add `Deck` method
   ```go
   func (b *Builder) Deck(id string, render ItemRender, itemHeight int) *Builder
   ```

3. **Tests** — `deck_test.go`
   - Render function receives correct `index`, `data`, and `selected=true` only
     for the highlighted slot.
   - Clip is applied per slot (verified via a render function that records the
     clip bounds it was called with).
   - `Move` skips disabled indices in both directions.
   - `PageDown` advances by the correct slot count for the given viewport height.
   - Scrollbar unit calculation produces the correct proportional thumb position.
   - `SetItems` with an empty slice sets `index = -1`.

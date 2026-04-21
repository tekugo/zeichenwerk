# Board

A scrollable columnar board widget for Kanban, sprint, and similar workflows.
Columns are named and hold an ordered list of cards; each card has a title,
optional body text, an optional colour label strip, and optional tags. The
cursor moves across columns with `←`/`→` and through cards with `↑`/`↓`.
Cards can be reordered within a column or moved between columns with keyboard
shortcuts and with mouse drag-and-drop. The Board is a single focusable widget
that owns all rendering and interaction; it does not delegate to Deck or Flex
internally.

---

## Visual layout

**Three columns, cursor on "API sketch", cardHeight = 4:**

```
┌────────────────┬────────────────┬────────────────┐
│  Backlog  (3)  │ In Progress(2) │      Done      │
├────────────────┼────────────────┼────────────────┤
│ ▌ Design doc   │ ▌ API sketch   │ ░ Auth flow    │
│   Spec out the │   Rough out    │   Completed    │
│   Board widget │   endpoints    │                │
│                │                │                │
│ ░ Unit tests   │ ▌ Write tests  │ ░ DB schema    │
│   For Board    │   Board widget │   Migrated     │
│   component    │                │                │
│                │                │                │
│ ░ Fix parser   │                │                │
│   Bug #142     │                │                │
│                │                │                │
└────────────────┴────────────────┴────────────────┘
```

- **Outer border** — box-drawing frame around the entire Board.
- **Column headers** — one row per column showing the title and card count;
  separated by `│`; a horizontal rule divides headers from cards.
- **Label strip** — 1-column coloured character (`▌`) on the left of each card.
  Empty label renders as a space in the card background colour.
- **Column separator** — `│` between columns; drawn as part of the board's
  own render, not as a border of each column.
- **Scrollbar** — drawn in the rightmost column of each column area when
  cards overflow the visible height.

**WIP limit exceeded (In Progress limit = 2, 3 cards):**

```
│ In Progress(3!)│
```

The `!` in the count — or a themed indicator — signals that the limit is
exceeded. The header switches to `"board/header:over"` style.

**During mouse drag — cursor dragging "Unit tests" into Done:**

```
│ ░ Unit tests   │                │ ▒▒▒▒▒▒▒▒▒▒▒▒▒  │
│   For Board    │                │ ▒▒▒▒▒▒▒▒▒▒▒▒▒  │
│   component    │                │ ▒▒▒▒▒▒▒▒▒▒▒▒▒  │  ← drop placeholder
│                │                │                │
```

The source card renders in `"board/card/drag"` style (dimmed). The drop
placeholder fills the target slot with the `"board.drop"` character.

---

## Data model

### Card

```go
type Card struct {
    ID       string   // unique within the board; auto-generated if empty
    Title    string   // primary line; required
    Body     string   // optional detail text; may contain \n
    Label    string   // colour token for left strip ("$red", "#ff4040", ""); empty = no colour
    Tags     []string // short labels shown as [tag] chips on the bottom card row
    Metadata any      // opaque caller data; not rendered
}
```

### Column

```go
type Column struct {
    ID    string  // unique identifier
    Title string  // header label
    cards []*Card // ordered; index 0 = top
    limit int     // WIP limit (0 = unlimited)
    width int     // explicit column width override (-1 = use Board.colWidth)
}
```

All `Column` fields except `ID` and `Title` are private; mutations go through
`Column` methods so the Board can stay consistent.

### BoardChange

Carried as data on `EvtChange`:

```go
type BoardChange struct {
    Card       *Card
    FromColumn *Column
    FromIndex  int
    ToColumn   *Column
    ToIndex    int
}
```

---

## Structure

```go
type Board struct {
    Component
    columns   []*Column
    states    []colState  // parallel to columns; scroll offsets and widths
    colIdx    int         // focused column index
    cardIdx   int         // focused card index within columns[colIdx] (-1 if empty)
    colWidth  int         // default column width in display columns (default 22)
    cardHeight int        // rows per card slot including gap row (minimum 3, default 4)
    // Drag state
    dragging      bool
    dragColIdx    int
    dragCardIdx   int
    dragTargetCol int
    dragTargetCard int
    dragMouseX    int
    dragMouseY    int
    // Characters read from theme strings in Apply.
    chSeparator string // column separator, default "│"
    chLabel     string // label strip character, default "▌"
    chDrop      string // drop placeholder fill character, default "░"
    chWipOK     string // WIP count suffix when under limit, default ""
    chWipOver   string // WIP count suffix when over limit, default "!"
}

type colState struct {
    offset int // index of first visible card
}
```

---

## Constructor

```go
func NewBoard(id, class string) *Board
```

Defaults:

- `colWidth = 22`, `cardHeight = 4`.
- `colIdx = 0`, `cardIdx = -1` (no selection until columns and cards are added).
- `chSeparator = "│"`, `chLabel = "▌"`, `chDrop = "░"`, `chWipOK = ""`,
  `chWipOver = "!"`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

---

## Column API

```go
func (b *Board) AddColumn(id, title string) *Column
```

Appends a column; appends a matching `colState{offset: 0}`; focuses
`cardIdx = 0` if the board had no columns before. Returns the `*Column`.

```go
func (b *Board) RemoveColumn(id string) bool
```

Removes the column; clamps `colIdx` to the new length; returns `true` if found.

```go
func (b *Board) MoveColumn(id string, toIndex int) bool
```

Reorders the column to `toIndex`; adjusts `colIdx` to track the same column.

```go
func (b *Board) GetColumn(id string) *Column
```

Returns the column with the matching ID, or `nil`.

```go
func (b *Board) Columns() []*Column
```

Returns the column slice (read-only snapshot).

### Column methods

```go
func (c *Column) SetTitle(s string)
func (c *Column) SetLimit(n int)  // 0 = unlimited
func (c *Column) SetWidth(n int)  // -1 = use Board.colWidth
```

```go
func (c *Column) AddCard(card *Card) *Card
func (c *Column) InsertCard(index int, card *Card) *Card
func (c *Column) RemoveCard(id string) *Card   // returns removed card or nil
func (c *Column) MoveCard(id string, toIndex int) bool
func (c *Column) Card(id string) *Card
func (c *Column) Cards() []*Card               // snapshot
func (c *Column) Len() int
func (c *Column) Over() bool                   // true when limit > 0 && Len() > limit
```

All mutating column methods call `Redraw(b)` so the Board repaints.

---

## Board display options

```go
func (b *Board) SetCardHeight(rows int)   // minimum 3
func (b *Board) SetColumnWidth(cols int)  // minimum 10; default for columns without SetWidth
```

---

## Navigation API

```go
func (b *Board) Select(colIdx, cardIdx int)
```

Focuses the card at `(colIdx, cardIdx)`. Clamps both indices to valid ranges;
adjusts `colState[colIdx].offset` to ensure the card is visible; dispatches
`EvtSelect` with `(colIdx, cardIdx)` data if the position changed.

```go
func (b *Board) SelectedColumn() int   // -1 if no columns
func (b *Board) SelectedCard() int     // -1 if column empty
func (b *Board) FocusedCard() *Card    // nil if no card focused
func (b *Board) FocusedColumn() *Column
```

---

## Keyboard interaction

### Navigation

| Key | Behaviour |
|-----|-----------|
| `↑` | Focus previous card in current column; no-op at top |
| `↓` | Focus next card in current column; no-op at bottom |
| `←` | Focus same relative card index in previous column |
| `→` | Focus same relative card index in next column |
| `Home` | Focus first card in current column |
| `End` | Focus last card in current column |
| `PgUp` | Scroll current column up one page |
| `PgDn` | Scroll current column down one page |
| `Enter` | Dispatch `EvtActivate` with `*Card` data |

`←`/`→` across columns preserves `cardIdx` clamped to the target column's
card count. If the target column is empty, `cardIdx` becomes -1.

### Card movement

| Key | Behaviour |
|-----|-----------|
| `Ctrl+↑` | Move focused card one position up within the column |
| `Ctrl+↓` | Move focused card one position down within the column |
| `Ctrl+←` | Move focused card to the bottom of the previous column |
| `Ctrl+→` | Move focused card to the bottom of the next column |

Each movement calls `Redraw(b)` and dispatches `EvtChange` with a
`BoardChange` value describing the move.

---

## Mouse interaction

### Click

A click at `(mx, my)` within the board:

1. Map `mx` to a column index using column widths and separator positions.
2. Map `my` to a card slot using `cardHeight` and the column's scroll offset.
3. If the resolved card differs from the current focus, call `Select`.
4. If the resolved card equals the current focus, dispatch `EvtActivate`.

Clicks on column headers or separators are ignored.

### Scroll wheel

`WheelUp` / `WheelDown` over a column scrolls that column's card list
independently of the focused column.

### Drag and drop

A drag is initiated when the user presses `Button1` on a card **and then moves
the cursor at least 1 cell** before releasing. A bare click (press + release
without movement) is treated as a click, not a drag.

**Drag start** — on `Button1` press on a card:

1. Record `dragColIdx`, `dragCardIdx`, `dragMouseX`, `dragMouseY`.
2. Do not yet enter drag mode (wait for movement).

**Drag enter** — on first `EvtMouse` with `Button1` held and cursor moved:

1. Set `dragging = true`.
2. Call `ui.Capture(b)` so the Board receives all subsequent mouse events
   even when the cursor leaves the board area (requires the `Capture`/`Release`
   mechanism from the Splitter spec).

**Drag move** — on `EvtMouse` with `Button1` held while `dragging`:

1. Update `dragMouseX`, `dragMouseY`.
2. Resolve `dragTargetCol`, `dragTargetCard` from the cursor position using the
   same column/row mapping as click, clamping to valid ranges.
   - Cursor left of column 0 → target is column 0.
   - Cursor right of last column → target is last column.
   - Cursor above all cards → insert at index 0.
   - Cursor below all cards → insert at last index.
3. Call `Redraw(b)`.

**Drop** — on `Button1` release while `dragging`:

1. Set `dragging = false`.
2. Call `ui.Release(b)`.
3. If `(dragTargetCol, dragTargetCard) != (dragColIdx, dragCardIdx)`:
   - Remove card from source; insert at target; adjust `colIdx` and `cardIdx`
     to follow the dropped card.
   - Dispatch `EvtChange` with `BoardChange`.
4. Call `Redraw(b)`.

**Drag abort** — `Escape` while dragging cancels and restores original
position. `ui.Release(b)` is called; `dragging = false`.

---

## Render

```go
func (b *Board) Render(r *Renderer)
```

### Layout computation

Each column `i` has a rendered width `colWidths[i]` computed from
`col.width` (if set) or `b.colWidth`. The column's screen X start is:

```
colX[0] = bx + 1            // +1 for outer left border
colX[i] = colX[i-1] + colWidths[i-1] + 1  // +1 for separator
```

### Render order

1. **Background** — `b.Component.Render(r)` fills the outer area.
2. **Outer border** — drawn using `"board"` style border characters.
3. **Column headers** — for each column `i` at row `by`:
   - Fill `colWidths[i]` cells with `"board/header"` style (or
     `"board/header:focused"` if `i == colIdx`).
   - Draw `title + " (" + count + wip_suffix + ")"`, centred, truncated.
   - Use `"board/header:over"` style when `col.Over()`.
4. **Header rule** — draw `─` across the full width at row `by + 1`,
   with `┼` at each separator position and `├`/`┤` at left/right borders.
5. **Column separators** — draw `│` at each `colX[i] - 1` for rows
   `by+2` to `by + bh - 1` (the card area).
6. **Cards** — for each column `i`, for each visible slot `s`:

```
cardIndex = colStates[i].offset + s
slotY     = by + 2 + s * cardHeight    // +2 for header + rule
slotH     = cardHeight - 1             // last row is the gap

if dragging and i == dragColIdx and cardIndex == dragCardIdx:
    render with "board/card/drag" style (dimmed)
elif dragging and i == dragTargetCol and cardIndex == dragTargetCard:
    render drop placeholder
elif i == colIdx and cardIndex == cardIdx:
    render with "board/card:focused" style
else:
    render with "board/card" style
```

7. **Scrollbars** — one per column when `len(cards) * cardHeight > available height`.

### Card render

At `(x, y, w, h)` for a single card slot (h = cardHeight - 1 — the gap
row is left blank in the card background colour):

```
Row 0:  [label 1col] [title, w-2 cols, "board/card/title" style]
Row 1:  [label 1col] [body line 0, "board/card/body" style]
…
Row h-2:[label 1col] [body line …]
Row h-1:[label 1col] [tags as "[tag]" chips, "board/card/tag" style]
```

When `cardHeight == 3` the tag row is omitted and body occupies row 1.
When `cardHeight == 2` both body and tags are omitted (title only).

The label strip at column `x` is coloured with `r.theme.Color(card.Label)` as
foreground; background comes from the card's resolved style. The label character
is `b.chLabel` (`▌`). When `card.Label == ""` the strip cell is a space.

### Drop placeholder render

Fill the entire slot (all `cardHeight - 1` rows × `colWidths[i]` columns)
with `b.chDrop` (`░`) using the `"board/drop"` style.

### Column underflow

When a column has fewer cards than the visible slot count, remaining slots are
filled with the column background `"board/column"` style.

---

## Scroll adjustment

After any navigation that changes `cardIdx` within a column, the column's
`colState.offset` is adjusted so the focused card is visible:

```
visibleSlots = (boardH - 2) / cardHeight   // -2 for header + rule
if cardIdx < offset:
    offset = cardIdx
elif cardIdx >= offset + visibleSlots:
    offset = cardIdx - visibleSlots + 1
```

---

## Events

| Event | Data | Description |
|-------|------|-------------|
| `EvtSelect` | `(int, int)` — `(colIdx, cardIdx)` | Cursor moved to a different card |
| `EvtChange` | `BoardChange` | Card moved (keyboard or drag-and-drop) |
| `EvtActivate` | `*Card` | Enter pressed or card double-clicked |

`EvtSelect` data is passed as two separate `int` arguments in the variadic
`data ...any` slice. Handlers read them as `data[0].(int)` and `data[1].(int)`.

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"board"` | Outer background and border |
| `"board/column"` | Card area background of each column |
| `"board/column:focused"` | Card area background of the focused column |
| `"board/header"` | Column header row |
| `"board/header:focused"` | Header of the focused column |
| `"board/header:over"` | Header when WIP limit is exceeded |
| `"board/card"` | Unfocused card background |
| `"board/card:focused"` | Focused card background |
| `"board/card/label"` | Label strip foreground colour override (rarely needed; label colour comes from `card.Label` directly) |
| `"board/card/title"` | Title row text |
| `"board/card/title:focused"` | Title row text on focused card |
| `"board/card/body"` | Body rows text |
| `"board/card/tag"` | Tag chip text |
| `"board/card/drag"` | Card being dragged (dimmed) |
| `"board/drop"` | Drop placeholder fill |

Example theme entries (Tokyo Night):

```go
NewStyle("board").WithColors("$fg0", "$bg0").WithBorder("thin"),
NewStyle("board/column").WithColors("$fg0", "$bg1"),
NewStyle("board/column:focused").WithColors("$fg0", "$bg2"),
NewStyle("board/header").WithColors("$fg0", "$bg1").WithFont("bold"),
NewStyle("board/header:focused").WithColors("$bg0", "$blue").WithFont("bold"),
NewStyle("board/header:over").WithColors("$bg0", "$red").WithFont("bold"),
NewStyle("board/card").WithColors("$fg0", "$bg2"),
NewStyle("board/card:focused").WithColors("$fg0", "$bg3"),
NewStyle("board/card/title").WithColors("$fg0", "$bg2").WithFont("bold"),
NewStyle("board/card/title:focused").WithColors("$fg0", "$bg3").WithFont("bold"),
NewStyle("board/card/body").WithColors("$fg1", "$bg2"),
NewStyle("board/card/tag").WithColors("$fg2", "$bg1"),
NewStyle("board/card/drag").WithColors("$fg2", "$bg1"),
NewStyle("board/drop").WithColors("$bg3", "$bg2"),
```

---

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `"board.separator"` | `"│"` | Vertical bar between columns |
| `"board.label"` | `"▌"` | Left-edge colour strip character |
| `"board.drop"` | `"░"` | Fill character for the drop placeholder slot |
| `"board.wip-ok"` | `""` | Suffix appended to count when under/at WIP limit |
| `"board.wip-over"` | `"!"` | Suffix appended to count when WIP limit exceeded |

---

## Builder usage

```go
board := NewBoard("sprint", "")
board.SetCardHeight(4).SetColumnWidth(22)

todo := board.AddColumn("todo", "To Do")
todo.AddCard(&Card{Title: "Design auth flow",  Body: "OAuth2 vs sessions", Label: "$blue"})
todo.AddCard(&Card{Title: "Write unit tests",  Body: "Filter + Board",     Label: "$orange", Tags: []string{"testing"}})
todo.AddCard(&Card{Title: "Fix parser bug",    Body: "Issue #142",         Label: "$red"})

wip := board.AddColumn("wip", "In Progress")
wip.SetLimit(2)
wip.AddCard(&Card{Title: "API sketch",    Body: "Rough endpoints",  Label: "$cyan"})
wip.AddCard(&Card{Title: "Write tests",   Body: "Board component",  Label: "$cyan"})

done := board.AddColumn("done", "Done")
done.AddCard(&Card{Title: "Auth flow",  Body: "Completed",   Label: "$green"})
done.AddCard(&Card{Title: "DB schema",  Body: "Migrated",    Label: "$green"})

board.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
    card := data[0].(*Card)
    ui.Prompt("Edit card", card.Title, func(title string) {
        card.Title = title
        Redraw(board)
    }, nil)
    return true
})

board.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
    change := data[0].(BoardChange)
    log.Printf("moved %q from %s[%d] to %s[%d]",
        change.Card.Title,
        change.FromColumn.Title, change.FromIndex,
        change.ToColumn.Title,   change.ToIndex)
    return false
})
```

---

## Implementation plan

1. **`board.go`** — new file
   - Define `Card`, `Column`, `colState`, `BoardChange`, `Board`.
   - Implement `NewBoard`.
   - Implement `Column` mutators: `SetTitle`, `SetLimit`, `SetWidth`,
     `AddCard`, `InsertCard`, `RemoveCard`, `MoveCard`, `Card`, `Cards`,
     `Len`, `Over`.
   - Implement `Board` column API: `AddColumn`, `RemoveColumn`, `MoveColumn`,
     `GetColumn`, `Columns`.
   - Implement `SetCardHeight`, `SetColumnWidth`.
   - Implement `Select`, `SelectedColumn`, `SelectedCard`, `FocusedCard`,
     `FocusedColumn`.
   - Implement `Apply(t *Theme)`: register all selectors; read theme strings.
   - Implement `Hint() (int, int)`: natural width = sum of column widths +
     separators + 2; natural height = 0 (fill parent).
   - Implement `Render(r *Renderer)`: outer border, headers, header rule,
     column separators, card slots, scrollbars.
   - Implement `renderCard(r, card, x, y, w, h, focused, dragging bool)`.
   - Implement `renderDrop(r, x, y, w, h int)`.
   - Implement `handleKey(e *tcell.EventKey) bool`: navigation and Ctrl-move.
   - Implement `handleMouse(e *tcell.EventMouse) bool`: click, wheel,
     drag start/move/drop/abort.
   - Implement `hitTest(mx, my int) (colIdx, cardIdx int)`: maps screen
     coordinates to `(column, card)`.
   - Implement `adjust(colIdx int)`: scroll offset correction.
   - Implement private card-move helper `moveCard(fromCol, fromIdx, toCol, toIdx int)`.

2. **`ui.go`** — mouse capture (if not already added for Splitter)
   - Add `capture Widget` field to `UI`.
   - Implement `Capture(w Widget)` and `Release(w Widget)`.
   - Route `EvtMouse` to `capture` when non-nil.

3. **`builder.go`** — add `Board` method
   ```go
   func (b *Builder) Board(id string) *Builder
   ```

4. **Theme** — add all `"board"` and `"board/*"` style entries and
   `"board.*"` string keys to all built-in themes.

5. **`cmd/demo/main.go`** — add a `"Board"` entry with `boardDemo`
   demonstrating a three-column sprint board with example cards, drag-and-drop,
   and an `EvtActivate` handler that opens a `Prompt` to rename the card.

6. **Tests** — `board_test.go`
   - `hitTest` maps screen coordinates to the correct `(colIdx, cardIdx)`.
   - `Select` clamps indices and updates `colState.offset`.
   - `Ctrl+↑`/`Ctrl+↓` reorders cards and dispatches `EvtChange`.
   - `Ctrl+←`/`Ctrl+→` moves a card to the adjacent column.
   - Drag: `dragging` becomes true after cursor moves; drop dispatches
     `EvtChange` with correct `FromIndex`/`ToIndex`.
   - Drag abort with Escape restores original card order.
   - `Column.Over()` returns true when `limit > 0 && Len() > limit`.
   - `RemoveColumn` clamps `colIdx`; `MoveColumn` tracks the same column.
   - `Render` does not panic for zero columns, empty columns, or
     `cardHeight < available height`.
   - Cards with `Label = ""` render the strip as a space, not `chLabel`.

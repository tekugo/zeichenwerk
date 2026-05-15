# Malwerk

A modal terminal-art editor — a vim-flavoured drawing tool for grids of styled
characters. Lives in `cmd/malwerk` and ships as a standalone binary. Box
diagrams, retro logos, status panels, ASCII headers — anything you'd build
out of glyphs and box-drawing characters.

The document is a 2D grid of `(rune, style-name)` cells plus a per-document
palette of named styles. Drawing is mode-driven (Normal / Insert / Visual /
Extended / Command), and box-drawing junctions are auto-merged so overlapping
borders and lines produce clean T- and cross-glyphs.

---

## Architecture

```
cmd/malwerk/
├── main.go        — UI shell, theme, top-level wiring
├── document.go    — Document model: palette, cells, dirty flag
├── style.go       — DocStyle (fg, bg, font, border) + palette ops
├── editor.go      — Editor widget: cursor, modes, render, key handling
├── normal.go      — Normal-mode key handler
├── insert.go      — Insert-mode key handler
├── visual.go      — Visual-mode key handler + selection rendering
├── extended.go    — Extended-mode keypad
├── junction.go    — Box-drawing junction merge tables + line drawing
├── glyphs.go      — Embedded glyph-name index for the i picker
├── history.go     — Undo / redo stack
├── register.go    — Yank / put register
├── ansi.go        — ANSI export
├── ui.go          — Status bar, dialogs (New, Resize, Style editor, Glyph picker)
└── commands.go    — Command-palette items
```

The editor widget is **not** the existing `widgets.Canvas` — Canvas's input
bindings would conflict with malwerk's. Editor is a fresh widget that
borrows Canvas's general approach (cell buffer, cursor, modes) but owns
the whole input pipeline.

---

## UI layout

```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│                    (editor canvas)                       │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
└──────────────────────────────────────────────────────────┘
 [NORMAL] 80×24 @ 12,7 (+0,0)  Style: accent  ■fg ■bg
```

- **Editor canvas** fills the full screen — no header, no border.
- **Status bar** is a single row at the bottom; can be toggled off via the
  command palette (`Document → Toggle Status Bar`).
- **Command palette** (Ctrl-K) overlays at the top.
- **Popups** (style editor, glyph picker, dialogs) are centred modals.

When the document is **smaller** than the screen, it sits in the upper-left
corner; remaining cells render with the theme's background. When it's
**larger**, the viewport scrolls — see [Scrolling](#scrolling).

---

## Status bar

```
[NORMAL] 80×24 @ 12,7 (+0,0)  Style: accent  ■fg ■bg
```

| Field | Meaning |
|-------|---------|
| `[MODE]` | Current mode in uppercase: NORMAL, INSERT, VISUAL, EXTENDED, COMMAND |
| `W×H` | Document dimensions |
| `@ x,y` | Cursor position (0-based, document-relative) |
| `(+dx,dy)` | Viewport scroll offset |
| `Style: name` | Current style name from the palette |
| `■fg ■bg` | Single-cell colour swatches showing the resolved fg / bg |

Visible by default; hidden via the command palette. When hidden, the editor
canvas reclaims the bottom row.

---

## Document model

### `DocStyle`

```go
type DocStyle struct {
    Fg     string  // theme variable ("$cyan") or literal ("#7aa2f7")
    Bg     string  // same; "" means inherit from "default"
    Font   string  // space-separated attrs: "bold", "italic", "underline", or combinations
    Border string  // named border ("thin", "double", "round", ""); used by the b / B tool
}
```

Empty fields inherit from the palette's `default` style (which itself
inherits from the active theme).

### `Cell`

```go
type Cell struct {
    Ch    string  // single rune as UTF-8
    Style string  // palette key, or "" for default
}
```

The empty cell is `{Ch: " ", Style: ""}`. The cell's effective style is
`palette[Style]` with fall-through to `palette["default"]`.

### `Document`

```go
type Document struct {
    Width   int
    Height  int
    Palette map[string]*DocStyle  // always contains a "default" entry
    Cells   [][]Cell              // [y][x]; len == Height, len([y]) == Width
    Path    string                // last save/load path; "" for unsaved
    Dirty   bool                  // true when there are unsaved changes
}
```

The `default` palette entry is created at document construction time and
cannot be deleted (only edited). Renaming a style updates every cell that
references it.

---

## JSON schema

Saved as `*.malwerk.json`:

```json
{
  "width": 80,
  "height": 24,
  "palette": {
    "default": { "fg": "$fg0", "bg": "$bg0" },
    "accent":  { "fg": "$cyan", "font": "bold", "border": "double" },
    "muted":   { "fg": "$fg2" }
  },
  "cells": [
    [
      { "ch": "┌", "style": "accent" },
      { "ch": "─", "style": "accent" },
      …
    ],
    …
  ]
}
```

- `cells` is a 2D array; outer length must equal `height`, inner length must
  equal `width`. Mismatches are rejected on load.
- Empty cells serialize as `{"ch": " ", "style": ""}` to keep the schema
  uniform (no run-length compression in v1).
- Theme variables (`$cyan`) are stored verbatim and resolved against
  whichever theme is active when the file is rendered.

### Loading

- Validate dimensions, palette, and cell-grid shape.
- A missing `default` palette entry is auto-inserted with the theme's
  default fg/bg.
- Unknown style names on cells are remapped to `default` with a warning.

---

## ANSI export

Written verbatim — no trimming, no final reset. Each row ends with `\n`.
Each cell emits `\x1b[…mC` where the SGR parameters cover the resolved fg,
bg, and font attributes. Theme variables resolve against the active theme.

```
\x1b[38;2;122;162;247m\x1b[48;2;26;27;38m┌\x1b[38;2;122;162;247m\x1b[48;2;26;27;38m─…
```

Lossy w.r.t. style names — round-tripping ANSI back into a malwerk
document is not supported.

---

## Modes

| Mode | Entry from Normal | Exit |
|------|-------------------|------|
| Normal | startup | (terminal) |
| Insert | `a` | `Esc` |
| Visual | `v` | `Esc` (also after applying any action with the action-then-stay rule) |
| Extended | `e` | `Esc` |
| Command | `Ctrl-K` | `Esc` (closes the palette) |

`Esc` is the universal "back to Normal" key — it works from any mode,
including inside dialogs and popups (where it dismisses the popup before
landing in Normal).

Visual and Extended are **sticky**:

- **Visual**: after `d`/`s`/`S`/`b`/`B`/`h`/`v`/`y` the action applies but
  the selection persists, so the user can apply more actions to the same
  region. Pressing `Esc` (or moving the cursor outside the selection with
  no modifier — *no, leave that out*) clears it.
- **Extended**: numpad inputs draw box pieces in succession; `Esc` returns
  to Normal.

---

## Key bindings

### Global (work in any mode unless overridden)

| Key | Action |
|-----|--------|
| `Esc` | Return to Normal (closes popups / cancels Visual selection / exits Insert / Extended / Command) |
| `Ctrl-K` | Open command palette |
| `Ctrl-S` | Save (or Save As if `Path == ""`) |
| `Ctrl-Q` | Quit (prompts if dirty) |
| `Ctrl-Z` | Undo |
| `Ctrl-R` | Redo |

### Normal mode

| Key | Action |
|-----|--------|
| Arrow keys | Move cursor (clamped to document bounds) |
| `Home` / `End` | Move to start / end of current row |
| `PgUp` / `PgDn` | Scroll viewport one page |
| `[` / `]` | Scroll viewport one half-page |
| `a` | Enter Insert mode |
| `e` | Enter Extended mode |
| `v` | Enter Visual mode (anchor at cursor) |
| `i` | Open glyph picker (typeahead by Unicode / nerd-font name) |
| `s` | Cycle through palette styles (current style → next entry) |
| `S` | Open style editor popup |
| `B` | Open border picker (changes the current style's `Border` field) |
| `x` | Clear cell to default style + space; advance cursor right |
| `f` | Flood-fill from cursor with current rune + style |
| `u` | Undo |
| `Ctrl-R` | Redo |
| `p` | Put yanked rectangle at cursor |

### Insert mode

| Key | Action |
|-----|--------|
| Any printable rune | Insert at cursor with current style; advance cursor right (wraps to next row at right edge) |
| Arrow keys | Move cursor (do not leave Insert mode) |
| `Backspace` | Move cursor left and clear that cell |
| `Enter` | Move to start of next row |
| `Esc` | Return to Normal |

### Visual mode

Selection is a rectangle from the anchor (where `v` was pressed) to the
cursor, inclusive. Both anchor and cursor can move; arrow keys extend the
selection.

| Key | Action |
|-----|--------|
| Arrow keys | Move cursor (selection extends) |
| `d` | Delete: fill selection with default-style spaces |
| `s` | Cycle current style and re-style the selection |
| `S` | Open style editor; chosen style applies to selection |
| `b` | Draw a border (current style's `Border`) around the selection rect |
| `B` | Open border picker, then draw the chosen border |
| `h` | Draw a horizontal line across the selection's middle row, with junction merge |
| `v` | Draw a vertical line down the selection's middle column, with junction merge |
| `y` | Yank the selection to the register |
| `Esc` | Cancel selection, return to Normal |

After `d`/`s`/`S`/`b`/`B`/`h`/`v`/`y`, the selection persists. Apply more
actions or `Esc` to leave.

### Extended mode

Numpad keys insert box-drawing glyphs at the cursor and advance right.
Layout matches the physical numpad geometry — each key inserts the piece
that lives in that position of a small frame:

```
    7 ┌    8 ┬    9 ┐
    4 ├    5 ┼    6 ┤
    1 └    2 ┴    3 ┘

    0 ─    . │    , │
```

`,` is an alias for `.` (laptop keyboards without a numpad). All glyphs use
the current style's `Border` family — see [Border families](#border-families).

| Key | Action |
|-----|--------|
| 1-9, 0, `.`, `,` | Insert the corresponding box piece, advance cursor right |
| Arrow keys | Move cursor (do not leave Extended) |
| `Esc` | Return to Normal |

---

## Drawing rules

### Border families

A "border style" is a named family of box-drawing runes:

| Name | Light | Heavy | Double | Rounded |
|------|-------|-------|--------|---------|
| `thin` | ┌┐└┘├┤┬┴┼─│ | | | |
| `heavy` | | ┏┓┗┛┣┫┳┻╋━┃ | | |
| `double` | | | ╔╗╚╝╠╣╦╩╬═║ | |
| `round` | ╭╮╰╯├┤┬┴┼─│ | | | |

`thin` is the default. The current style's `Border` field selects the
family for any drawing operation that produces a box-drawing rune (`b`,
`h`, `v`, the numpad keys).

### Junction-aware merging

When a drawing operation places a rune in a cell that already contains a
box-drawing rune of the **same family**, the cell becomes the junction
that combines the existing connections with the new ones. Each box-drawing
rune corresponds to a 4-bit connection mask (top, right, bottom, left):

```
─ = 0101 (left + right)
│ = 1010 (top + bottom)
┌ = 0110 (right + bottom)
┼ = 1111 (all four)
…
```

Drawing `─` over `│` produces `┼` (mask `0101 | 1010 = 1111`). Drawing
`─` over `┌` produces `┬` (`0110 | 0101 = 0111`). The full 16-entry mask
→ rune table lives in `junction.go` per family.

**Mixing families** (e.g. drawing `thin` over `double`) is **not**
auto-merged in v1: the new rune simply overwrites the old. Unicode does
have mixed-junction characters (`╪ ╫`) but supporting them properly
requires per-side family tracking — out of scope for v1.

### Graph-aware re-evaluation

After any line / border draw, re-evaluate every cell **adjacent** to the
modified region. A cell's rune is re-derived from the connection mask
formed by looking at its four neighbours:

```
mask = (top.connectsBottom    ? 1000 : 0)
     | (right.connectsLeft    ? 0100 : 0)
     | (bottom.connectsTop    ? 0010 : 0)
     | (left.connectsRight    ? 0001 : 0)
```

This means existing T-junctions and corners auto-correct when a new line
arrives next to them. Re-evaluation only touches cells whose rune is
already in the same family — typed text is never overwritten.

### `b` border draw (Visual mode)

Given a selection rect `(x1, y1) → (x2, y2)`:

1. Place corners: `┌` at `(x1, y1)`, `┐` at `(x2, y1)`, `└` at `(x1, y2)`,
   `┘` at `(x2, y2)`.
2. Top / bottom edges: `─` along `y1` and `y2` between corners.
3. Left / right edges: `│` along `x1` and `x2` between corners.
4. Each placement goes through junction merge.
5. Re-evaluate the rectangle's perimeter ± 1 cell.

`B` opens a picker (Pick Border), then runs the same routine.

### `h` / `v` line draw (Visual mode)

`h` draws `─` across the selection's middle row (`(y1+y2)/2`), full width;
junction-merged. `v` does the symmetric for the middle column.

### Flood fill (`f` Normal, `F` Visual)

Standard 4-connected flood-fill. From the cursor, replace all connected
cells whose `(Ch, Style)` matches the seed cell with `(currentRune,
currentStyle)`. In Visual mode (`F`), the fill is clipped to the
selection rect.

---

## Scrolling

The editor displays a viewport over the document. The viewport is the
size of the editor canvas (full screen minus the status bar).

- **Auto-scroll on cursor edge**: when the cursor moves past the viewport
  edge, the viewport advances by one row / column to keep the cursor in
  view.
- **Manual scroll**: `PgUp` / `PgDn` shift the viewport one full page;
  `[` / `]` shift one half-page.
- The viewport never scrolls past the document edges.
- When the document is smaller than the viewport, the offset is fixed at
  `(0, 0)` and the document sits in the upper-left.

The viewport offset shows in the status bar as `(+dx,dy)`.

---

## Style editor (`S`)

Modal popup. Layout:

```
┌─ Styles ──────────────────────────────────┐
│ ▶ default                                 │  ┌─ Edit ─────────────────┐
│   accent                                  │  │ Name:    accent        │
│   muted                                   │  │ Fg:      $cyan         │
│   warning                                 │  │ Bg:                    │
│                                           │  │ Font:    [bold]        │
│                                           │  │ Border:  [double]      │
│                                           │  │                        │
│                                           │  │ Sample: Hello world    │
│ [New] [Rename] [Delete] [Pick] [Close]    │  └────────────────────────┘
└───────────────────────────────────────────┘
```

- **List** on the left shows palette entries; selection cycles with
  arrows.
- **Edit form** on the right edits the highlighted style's attributes
  live; changes apply on `[Pick]` or `[Close]`.
- **New / Rename / Delete** — straightforward. The `default` entry can
  be edited but not renamed or deleted.
- **Pick** sets the highlighted style as current; in Visual mode, also
  applies it to the selection.
- **Close** exits without changing the current style.

Behind the scenes: typed values for `Fg` / `Bg` accept either theme
variables (`$cyan`) or any literal the active theme would forward (a
hex string, a tcell colour name).

---

## Glyph picker (`i`)

Modal typeahead. As the user types, the list filters glyph names
case-insensitively; arrows + Enter pick. The picker draws from an
embedded curated index (`glyphs.go`) covering:

- **Box drawing**: `┌─┐│└┘├┤┬┴┼` (light, heavy, double, round families)
- **Block elements**: `█▓▒░▌▐▀▄`
- **Geometric shapes**: `■□●○◆◇▲△▼▽◀▶`
- **Arrows**: `←→↑↓↔↕⇐⇒⇑⇓`
- **Math / misc**: `±×÷≈≠≤≥∞°§¶†‡`
- **Nerd-font icons** (a small selection — folder, file, gear, search,
  warning, …)

~3-5k entries total; baked into the binary as a `[]glyphEntry`.

After selecting a glyph, the picker inserts it at the cursor with the
current style and advances right. Press `Esc` to cancel.

---

## Command palette (`Ctrl-K`)

Modal palette with the following items:

### File
- **New** — opens a width × height dialog; replaces current document
  (prompts if dirty).
- **Open** — file chooser; loads `*.malwerk.json`.
- **Save** — writes to `Path`; falls back to Save As if unsaved.
- **Save As** — file chooser with `.malwerk.json` filter.
- **Export ANSI** — file chooser with `.ans` filter (default filename
  derived from current path); writes the ANSI render.
- **Quit** — exits malwerk; prompts if dirty.

### Document
- **Resize** — width × height dialog; keeps existing cells, clips or
  pads to the new size.
- **Toggle Status Bar** — show / hide the bottom row.
- **Clear** — fills the whole canvas with default-style spaces (undoable).

### Style
- **Pick Style** — opens a flat list of palette names; arrow + Enter
  selects. Same as `S → Pick`.
- **Edit Styles** — opens the style editor (same as `S`).
- **Pick Border** — opens a list of border family names (`thin`, `heavy`,
  `double`, `round`); the chosen value is written to the current style's
  `Border` field. Same as `B` in Normal mode.

The palette uses the existing zeichenwerk `Commands` subsystem.

---

## Undo / redo

A linear history of cell-edit batches. A "batch" is the result of a single
user action — one keystroke in Insert / Extended, one `b` border, one
flood-fill, etc. — captured as a list of `(x, y, before, after)` deltas.

```go
type Edit struct {
    X, Y       int
    Before, After Cell
}
type Batch []Edit
```

`history` is a stack of batches plus a pointer; `Undo` walks one batch
backwards, `Redo` re-applies. Maximum depth: 200 batches.

Document mutations bypass history when they're loading from disk or
constructed by the document itself (palette renames, resize). Resize
**is** undoable in v1.

---

## Yank / put

A single anonymous register holding a rectangle of cells:

```go
type Register struct {
    W, H  int
    Cells [][]Cell
}
```

- **`y` in Visual** — copies the selection into the register.
- **`p` in Normal** — pastes the register at the cursor (top-left corner
  of the rectangle aligns with the cursor). Pasting clips at the
  document edges; styles outside the document's palette are auto-added
  (or remapped to `default` if the register's palette is unavailable —
  in v1 we always paste with the document's current palette names).

Registers do not survive across runs in v1.

---

## Implementation plan

1. **`document.go` + `style.go`** — Document, DocStyle, palette ops, JSON
   round-trip with validation.
2. **`editor.go`** — Editor widget: cell access, cursor, viewport,
   render, base mode dispatch.
3. **Mode handlers** — `normal.go`, `insert.go`, `visual.go`,
   `extended.go`. Each owns its key map; the editor dispatches by mode.
4. **`junction.go`** — Border-family rune tables, mask ↔ rune lookup,
   line / border draw with junction merge + neighbour re-eval.
5. **`ui.go`** — Status bar widget, dialogs (New, Resize, Style editor,
   Glyph picker, Pick Border), prompts (dirty-quit confirmation).
6. **`commands.go`** — Palette items wired to the actions above.
7. **`history.go`** — Edit, Batch, History; mutation helpers that record
   deltas automatically.
8. **`register.go`** — Yank / put.
9. **`ansi.go`** — Export.
10. **`glyphs.go`** — Embedded glyph-name index (curated list).
11. **`main.go`** — UI shell, theme picker, top-level wiring,
   command-line flags (`--theme`, `<file>`).

Tests focus on the deterministic pieces:

- `document_test.go`: JSON round-trip, palette rename propagation,
  resize clip / pad.
- `junction_test.go`: every 16-entry mask table is consistent (mask →
  rune → mask is a fixed point); line draw produces the expected
  glyphs across a fixture grid.
- `history_test.go`: undo + redo restores byte-identical Cell state.
- `register_test.go`: yank then put round-trips a region.

UI flows are exercised by hand via the `cmd/malwerk` smoke run; no
headless UI tests in v1.

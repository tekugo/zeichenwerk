# TUI Designer Specification

## Goal

A full-screen interactive canvas occupying the entire terminal for creating
terminal user interface designs and Unicode art. VIM-style modal editing.
Fast character insertion from Unicode palettes via the NumPad. Per-cell
styling. Named hierarchical styles. Multi-page documents. Load/save as JSON.

---

## Document Model

A **Document** holds one or more **Pages**. Each page is a fixed-size grid of
**Cells** (width × height). The canvas size matches the terminal at creation
time and can be resized via command.

```
Document
  name        string
  styles      []NamedStyle     // named styles available in this document
  pages       []Page

Page
  name        string
  width       int
  height      int
  cells       [][]Cell         // [row][col]

Cell
  ch          rune             // displayed character; 0 = space
  style       string           // name of a NamedStyle, "" == "default"
```

**NamedStyle** has a name, an optional parent name, and `fg`, `bg`, `attr`
fields. Empty fields inherit from the resolved parent, enabling a
**hierarchy**: `muted` may inherit `fg` from `accent`, which inherits `bg`
from `default`.

```
NamedStyle
  name        string           // unique identifier
  parent      string           // name of parent style, "" = no parent
  fg          string           // foreground color; empty = inherit from parent
  bg          string           // background color; empty = inherit from parent
  attr        string           // "bold"|"dim"|"italic"|"underline"|""; empty = inherit
```

**Resolution** walks the parent chain from root to leaf and merges fields:
a child's non-empty field overrides the parent's resolved value. Cycle
detection caps the walk at depth 32.

The built-in `"default"` style has no parent and all-empty fields, which
maps to the terminal's own default colors with no attributes. It cannot be
deleted or renamed.

Editing a named style updates every cell that references it visually; renaming
a style updates all cell references immediately.

---

## UI Layout

```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│                    canvas (full screen)                  │
│                                                          │
│                                                          │
├──────────────────────────────────────────────────────────┤
│ MODE  │ style:name  fg:color bg:color  │ col,row / W×H  │
└──────────────────────────────────────────────────────────┘
```

The bottom line (status bar) is always visible. Everything else (palette
overlay, style editor, page list) is shown as pop-ups on the right side of
the screen, opened on demand and closed with `Esc` or their toggle key.

**Status bar fields:**
- `MODE` — current mode name (`NORMAL`, `INSERT`, `DRAW`, `VISUAL`)
- `style:name` — currently active named style (or `default`)
- `fg:color bg:color` — color preview of the active style
- `col,row` — 0-based cursor position
- `W×H` — page dimensions

---

## Modes

### NORMAL

ESC always returns to NORMAL from any other mode.

**Movement:**

| Key | Action |
|-----|--------|
| `h` / `←` | left |
| `j` / `↓` | down |
| `k` / `↑` | up |
| `l` / `→` | right |
| `0` / `Home` | start of row |
| `$` / `End` | end of row |
| `g` | first row |
| `G` | last row |
| `w` | next page |
| `W` | previous page |

**Mode and action keys:**

| Key | Action |
|-----|--------|
| `i` | INSERT mode |
| `d` | DRAW mode |
| `v` | VISUAL mode |
| `:` | COMMAND mode |
| `b` | open NumPad palette selector (0–9) |
| `s` | open style selector popup |
| `p` | paste yanked area at cursor |
| `x` | delete character at cursor (replace with space, keep style) |
| `?` | show/hide help popup |

### Help Popup (`?` in NORMAL)

A centered overlay box listing all key bindings for every mode. Implemented
as a `Text` component inside a `Box` border drawn with the thin palette
characters. Closed with `Esc` or `?`.

### DRAW

Digits `1`–`9` insert the palette character that matches the NumPad grid
position (e.g. `7` = upper-left corner, `8` = upper T, `9` = upper-right
corner, `4` = left T, `5` = inner cross, `6` = right T, `1` = lower-left
corner, `2` = lower T, `3` = lower-right corner). `-` inserts the horizontal
line character and advances the cursor right. `*` inserts the vertical line
character and advances the cursor **down**. Arrow keys and hjkl move the
cursor without inserting.

| Key | Action |
|-----|--------|
| `s` | toggle NumPad assignment overlay |
| `Space` | delete character at cursor (replace with space, advance right) |
| `Backspace` | delete character to the left |
| `i` | insert icon (Nerd Font); opens searchable icon picker popup |
| `Esc` | back to NORMAL |

### INSERT

All printable characters are inserted using the active style. Cursor advances
right after each insertion. Wraps to the next row at the right edge.

| Key | Action |
|-----|--------|
| Arrow keys | move cursor (stay in INSERT) |
| `Backspace` | delete character to the left |
| `Esc` | back to NORMAL |

### VISUAL

Rectangular selection. The anchor is set when entering VISUAL; the cursor
extends the selection. The selected area is highlighted.

| Key | Action |
|-----|--------|
| Movement keys | extend selection |
| `y` | yank (copy) the selected area to the clipboard |
| `d` | delete area (space + default style) |
| `f` | fill area with active style (preserve characters) |
| `b` | open border picker, draw a box around the selection |
| `Esc` | cancel selection, back to NORMAL |

### COMMAND

Entered with `:`. A one-line input field appears in the status bar.

| Command | Action |
|---------|--------|
| `:w` | save document |
| `:w filename` | save as filename |
| `:e filename` | load document |
| `:q` | quit (prompts if unsaved changes) |
| `:q!` | quit without saving |
| `:wq` | save and quit |
| `:page name` | rename current page |
| `:newpage name` | add a new page after current |
| `:delpage` | delete current page (prompts) |
| `:resize W H` | resize current page to W×H |

---

## Drawing Palettes

Each **palette** assigns up to 13 characters to the NumPad keys
(`0`–`9`, `,`, `+`, `-`). Palettes are identified by a short name shown in
the status bar.

**Built-in palettes (first step — just these four):**

| # | Name | Characters |
|---|------|-----------|
| 0 | thin | `─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼ ╴ ╶` |
| 1 | round | `─ │ ╭ ╮ ╰ ╯ ├ ┤ ┬ ┴ ┼ ╴ ╶` |
| 2 | double | `═ ║ ╔ ╗ ╚ ╝ ╠ ╣ ╦ ╩ ╬ ╸ ╺` |
| 3 | block | `█ ▓ ▒ ░ ▄ ▀ ▌ ▐ ▇ ▆ ▅ ▃ ▂` |

NumPad key → palette character mapping (keys match visual grid positions):

```
7 8 9      ┌ ┬ ┐   (upper-left corner, upper T, upper-right corner)
4 5 6  →   ├ ┼ ┤   (left T, inner cross, right T)
1 2 3      └ ┴ ┘   (lower-left corner, lower T, lower-right corner)
-          ─        (horizontal line)
*          │        (vertical line)
```

The layout is the same for all palettes — the corner/T/cross semantics are
preserved (e.g. `7` is always the upper-left corner character of the active
palette).

The `b` key in NORMAL mode opens a one-line popup listing palette names;
pressing the corresponding digit selects it.

---

## Style Selector Popup (`s` in NORMAL)

A top-right overlay showing all named styles. The hierarchy is visible through
indentation (two spaces per depth level) and a `↑parent` annotation. A color
preview (two `█` cells rendered with the resolved style) appears at the right
of each row.

Row markers: `>` = cursor, `*` = currently active drawing style.

**List keys:**

| Key | Action |
|-----|--------|
| `j` / `↓` | move cursor down |
| `k` / `↑` | move cursor up |
| `Enter` | set selected style as active drawing style, close popup |
| `n` | open editor to create a new style |
| `e` | open editor for the selected style |
| `d` | delete selected style; cells that referenced it revert to `"default"` |
| `Esc` | close popup without changing the active style |

The `"default"` style cannot be deleted.

### Style Editor

An inline editor (replaces the list inside the same popup frame) with five
editable fields:

| Field | Description |
|-------|-------------|
| `name` | Unique style name (required) |
| `parent` | Parent style name (must exist; leave empty for no parent) |
| `fg` | Foreground color: name, `#rrggbb`, or `color0`–`color255` |
| `bg` | Background color (same formats) |
| `attr` | `bold`, `dim`, `italic`, `underline`, or empty |

**Editor keys:**

| Key | Action |
|-----|--------|
| `Tab` / `↓` | next field |
| `Shift-Tab` / `↑` | previous field |
| Printable chars | append to current field |
| `Backspace` | delete last character |
| `Enter` | validate and save; on rename, all cell references update automatically |
| `Esc` | cancel, return to list |

**Style attributes:** `bold`, `dim`, `italic`, `underline`
**Color format:** terminal color names (`red`, `blue`, …), hex `#rrggbb`, or
terminal indices `color0`–`color255`.

---

## File Format (JSON)

```json
{
  "version": 1,
  "name": "my-design",
  "styles": {
    "header": { "fg": "#cdd6f4", "bg": "#1e1e2e", "attr": "bold" },
    "accent":  { "parent": "header", "fg": "#89b4fa" },
    "muted":   { "parent": "accent", "attr": "dim" }
  },
  "pages": [
    {
      "name": "main",
      "width": 80,
      "height": 24,
      "cells": [
        { "r": 0, "c": 5, "ch": "─", "style": "header" }
      ]
    }
  ]
}
```

Only non-empty cells are stored. Empty cells are implicit. The `cells` array
uses sparse row/column objects to keep files small.

The `"default"` style is implicit (no parent, terminal defaults) and is never
written to the `styles` map.

**Backward compatibility:** files saved by older versions stored `fg`, `bg`,
`attr` directly on each cell (no `style` name). On load, such cells are
automatically migrated to synthetic named styles so no styling is lost.

---

## First Implementation Step

The minimal viable designer covers:

1. Full-screen `Canvas` occupying `rows-1` lines; status bar on the last line.
2. NORMAL mode with movement keys and mode switching.
3. INSERT mode with character insertion and Backspace.
4. DRAW mode using one palette (thin) via NumPad.
5. VISUAL mode with yank (`y`) and delete (`d`).
6. COMMAND mode for `:w`, `:e`, `:q`, `:wq`.
7. Single-page document (multi-page is a follow-up).
8. Save/load in the JSON format above.
9. Status bar showing mode, cursor position, and page size.

Everything else (style editor, icon picker, multiple palettes, named styles,
multi-page navigation) is a follow-up once the core editing loop works.

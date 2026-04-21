# Specification: tblr

**Binary:** `cmd/tblr/main.go`  
**Package:** `main`  
**Depends on:** `zeichenwerk`, `golang.design/x/clipboard` (or `atotto/clipboard`)

---

## Purpose

`tblr` is a terminal UI for editing, converting, and pretty-printing tables.
It reads and writes CSV, TSV, Markdown, AsciiDoc, Typst, and HTML table formats.
It also functions as a headless CLI converter and a clipboard auto-formatter
that watches for table content and reformats or converts it on the fly.

---

## Layout

```
┌─ tblr ── table.csv ──────────────────────────────────────────────────────────┐
│  Name              Age    City            Score                               │
│  ──────────────────────────────────────────────                               │
│  Alice              28    Berlin           92.5                               │
│▶ Bob                34    New York         87.0                               │
│  Carol              22    Tokyo            95.1                               │
│                                                                               │
│                                                                               │
│  csv · 3 rows · 4 cols · [2,1] · ←left  ·  sorted: Age ↑                    │
│  e edit · a add · d del · s sort · / find · f format · w watch · q quit      │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Regions

| Region | Widget | Notes |
|--------|--------|-------|
| Table area | `Table` | Bound to `MutableTable` provider |
| Status bar | `Static` / `Styled` | Format · dimensions · cursor · alignment · sort state |
| Command bar | `Shortcuts` | Colour-coded key hints; context-sensitive |

---

## Internal Model — `MutableTable`

`MutableTable` implements `TableProvider` (so it binds directly to the `Table`
widget) and extends it with column alignment, cell mutation, structural
operations, sort, and dirty tracking.

```go
// Alignment is the per-column horizontal alignment.
type Alignment uint8

const (
    AlignLeft   Alignment = iota // default
    AlignCenter
    AlignRight
)

// MutableTable is the canonical in-memory table model used by tblr.
// It implements zeichenwerk.TableProvider so it can be passed directly to
// NewTable; the extra methods support editing and export.
type MutableTable struct {
    headers    []string    // column header labels
    alignments []Alignment // per-column alignment; len == len(headers)
    data       [][]string  // data rows (no header row)
    hasHeader  bool        // whether row 0 is semantically a header
    modified   bool        // true after any mutation
    delimiter  rune        // CSV field separator (default ',')
}
```

### TableProvider implementation

| Method | Behaviour |
|--------|-----------|
| `Columns() []TableColumn` | Returns the stored `[]TableColumn` as-is; `Width` is **not** recomputed — the application must call `RecalcWidths()` after mutations; `Sortable = true` |
| `Length() int` | Number of data rows |
| `Str(row, col int) string` | Cell value; `row` is 0-based data index |

### Additional methods

```go
// Metadata
func (t *MutableTable) HasHeader() bool
func (t *MutableTable) SetHasHeader(v bool)
func (t *MutableTable) ColAlignment(col int) Alignment
func (t *MutableTable) SetColAlignment(col int, a Alignment)
func (t *MutableTable) Delimiter() rune
func (t *MutableTable) SetDelimiter(r rune)
func (t *MutableTable) Modified() bool
func (t *MutableTable) ClearModified()

// RecalcWidths recomputes TableColumn.Width for every column as the maximum
// rune-count across the header and all data cells. Must be called explicitly
// by the application after any mutation that could change cell widths
// (SetCell, InsertRowAt, AppendRow, InsertColAt, MoveCol, Load, etc.).
// It does NOT set modified = true.
func (t *MutableTable) RecalcWidths()

// Cell mutation
func (t *MutableTable) SetCell(row, col int, value string)

// Structural mutation — row
func (t *MutableTable) InsertRowAt(at int)         // inserts blank row above at
func (t *MutableTable) AppendRow()                 // appends blank row at end
func (t *MutableTable) DeleteRow(at int)
func (t *MutableTable) MoveRow(from, to int)       // reorder; adjusts all rows between

// Structural mutation — column
func (t *MutableTable) InsertColAt(at int)         // inserts blank column; expands all rows
func (t *MutableTable) AppendCol(header string)
func (t *MutableTable) DeleteCol(at int)
func (t *MutableTable) MoveCol(from, to int)
func (t *MutableTable) RenameCol(col int, name string)

// Sort — sorts data rows only; never reorders the header
func (t *MutableTable) SortByCol(col int, asc bool)
// Stable sort; numeric if all non-empty cells parse as float64, else lexicographic.

// Bulk load — replaces all content; resets modified
func (t *MutableTable) Load(headers []string, data [][]string)
```

`InsertRowAt`, `DeleteRow`, `MoveRow`, `InsertColAt`, `DeleteCol`, `MoveCol`,
`SetCell`, `RenameCol`, and `SortByCol` all set `modified = true`.

---

## Format Support

| Format | Ext | Read | Write | Notes |
|--------|-----|------|-------|-------|
| CSV | `.csv` | ✓ | ✓ | Configurable delimiter; RFC 4180 quoting |
| TSV | `.tsv` | ✓ | ✓ | Tab delimiter; same parser as CSV |
| Markdown | `.md` | ✓ | ✓ | GFM pipe tables; alignment from `:---:` syntax |
| AsciiDoc | `.adoc` | ✓ | ✓ | `\|===` block; `cols` attribute for alignment |
| Typst | `.typ` | ✓ | ✓ | `#table(columns:…, …)` |
| HTML | `.html` | ✓ | ✓ | `<table>` / `<thead>` / `<tbody>` |

### Format interface

```go
// Format handles one table syntax.
type Format interface {
    Name() string                          // "csv", "markdown", etc.
    Extensions() []string                  // file extensions without dot
    Detect(data []byte) bool               // heuristic — returns true if likely this format
    Parse(data []byte, opts ParseOpts) (*MutableTable, error)
    Serialize(t *MutableTable, opts SerialOpts) ([]byte, error)
}

type ParseOpts struct {
    Delimiter rune   // CSV/TSV only; 0 = auto-detect
}

type SerialOpts struct {
    Pretty    bool   // pad columns to uniform width
    Delimiter rune   // CSV/TSV only
}
```

### Pretty printing

When `SerialOpts.Pretty = true`, each column is padded to its `Width` (from
`Columns()`). Alignment is applied: left-pad spaces for right-aligned cells,
split padding for centered cells. Markdown column separators use `:---`,
`:---:`, `---:` accordingly.

### Format detection order (for auto-detect)

1. File extension if available
2. `Detect()` called on each format in this order:
   Markdown → AsciiDoc → Typst → HTML → TSV → CSV
3. First match wins; CSV is the fallback

---

## CLI — Headless Mode

When stdin is not a TTY, or when `--to` / `--from` / `--format` flags are
present without a terminal, `tblr` runs headlessly and exits after processing.

```
tblr table.csv                         # open TUI with file
tblr                                   # open TUI, empty table

tblr --to markdown table.csv           # convert file → stdout (auto-detect input)
tblr --from csv --to markdown          # stdin → stdout converter
tblr --from csv --to markdown table.csv  # explicit input format
cat data.tsv | tblr --to html          # pipe mode
tblr --format markdown table.md        # pretty-print → stdout
tblr --format markdown -w table.md     # pretty-print in place (overwrite)
tblr --format markdown -w *.md         # pretty-print multiple files in place
tblr --delimiter ';' table.csv         # override CSV delimiter
tblr --ansi table.csv                  # render table to terminal with ANSI styles
cat data.csv | tblr --ansi             # pipe → styled terminal output
tblr --ansi --ansi-border double table.csv
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--from <fmt>` | Input format (default: auto-detect) |
| `--to <fmt>` | Output format |
| `--format <fmt>` | Pretty-print; input and output format are the same |
| `-w` | Write result back to the source file(s) instead of stdout; requires a file argument (not usable with stdin) |
| `--delimiter <char>` | CSV/TSV field separator (default `,`) |
| `--no-header` | Treat all rows as data (no header) |
| `--pretty` | Force pretty-print padding (default on for Markdown/AsciiDoc/Typst) |
| `--ansi` | Render table as ANSI-styled output for terminal display (see below) |
| `--ansi-border <style>` | Border style: `thin` (default), `double`, `rounded`, `thick`, `none` |
| `--ansi-theme <name>` | Colour theme: `auto` (default), `dark`, `light`, `16` (16-colour safe) |
| `--ansi-zebra` / `--no-ansi-zebra` | Alternating row backgrounds (default on) |
| `--width <n>` | Override output width in columns (default: terminal width or 80) |

`-w` without a file argument is an error. When multiple files are given with
`-w`, each is processed independently; a failure on one file does not abort the
others, but the exit code reflects the worst error.

`--ansi` output is not round-trippable — it is display-only and cannot be
parsed back into a table. Combining `--ansi` with `-w` is an error.

### ANSI output

`--ansi` renders the table using ANSI escape codes for direct terminal display.
It is intentionally not a serialisation format — the output is meant to be
read, not piped into another `tblr` invocation.

What it produces:

- **Border:** Unicode box-drawing characters; style controlled by
  `--ansi-border`. Maps to zeichenwerk's registered border names
  (`thin`, `double`, `rounded`, `thick`, `none`).
- **Header row:** Bold text on a distinct background (`$bg2` in dark themes);
  separated from data rows by a horizontal rule using the same box-drawing set.
- **Alignment:** Cells are padded to `Width` and aligned per column alignment
  (left/center/right).
- **Zebra striping:** Odd data rows get a subtly different background (`$bg1`
  vs default `$bg0`) when `--ansi-zebra` is on.
- **Colour theme:** `auto` uses `$COLORTERM` / `TERM` to pick truecolor, 256,
  or 16-colour output. `dark` / `light` force a built-in palette. `16` restricts
  to standard ANSI colours for maximum compatibility.
- **Width clamping:** If the table is wider than `--width`, columns are
  truncated with a `…` suffix rather than wrapping, so the grid lines stay
  intact.
- **Numeric highlighting:** Cells whose trimmed content parses as a number are
  rendered in an accent colour (`$cyan`) regardless of alignment, making numeric
  columns visually distinct at a glance.

Suggested use cases: `git diff`-style table previews in scripts, quick
inspection of CSV files (`cat data.csv | tblr --ansi`), README table
verification.

Exit code 0 on success, 1 on parse/format error (error written to stderr).

---

## TUI — Editing

### Selection model

| Selection | How to activate |
|-----------|----------------|
| Cell | Arrow keys / click |
| Range | Shift+Arrow; click-drag |
| Entire row | `Shift+Space` or click row number |
| Entire column | `Ctrl+Space` or click column header |
| All cells | `Ctrl+A` |

The `Table` widget is created with `SetCellNav(true)`. `←`/`→` move between
columns; `Tab`/`Shift+Tab` wrap across rows. `EvtSelect` carries `(row, col int)`.
The focused cell is styled with `"table/cell:focused"`; the containing row retains
`"table/highlight:focused"` so both row and cell position are visible.

Selections are highlighted with the `"table/grid:focused"` style. Row and
column selections highlight the full band.

### Keyboard bindings

| Key | Action |
|-----|--------|
| `↑↓←→` | Move cursor |
| `Shift+↑↓←→` | Extend selection |
| `Enter` / `F2` | Edit focused cell (inline input) |
| `Tab` / `Shift+Tab` | Move to next / previous cell (wraps rows) |
| `Escape` | Cancel edit / clear selection |
| `a` | Add row below selection (or at end) |
| `A` | Add column to the right of selection (prompts for header) |
| `Ctrl+Ins` | Insert row above selection |
| `Ctrl+Shift+Ins` | Insert column to the left of selection |
| `d` | Delete selected rows or columns |
| `s` | Sort ascending by focused column |
| `S` | Sort descending by focused column |
| `Alt+↑` / `Alt+↓` | Move selected row(s) up / down |
| `Alt+←` / `Alt+→` | Move selected column(s) left / right |
| `<` | Set column alignment: left |
| `>` | Set column alignment: right |
| `^` | Set column alignment: center |
| `/` | Open search bar |
| `n` / `N` | Next / previous search match |
| `f` | Open format picker (changes active format) |
| `w` | Toggle clipboard watch mode |
| `Ctrl+C` | Copy selection in active format |
| `Ctrl+V` | Paste; auto-detects format |
| `Ctrl+S` | Save to current file (or prompt if unsaved) |
| `Ctrl+Z` | Undo |
| `Ctrl+Y` / `Ctrl+Shift+Z` | Redo |
| `q` / `Ctrl+Q` | Quit (prompts if unsaved changes) |

### Cell editing

Pressing `Enter` or `F2` on a cell opens an inline `Input` widget in place.
`Enter` confirms; `Escape` cancels. Tab-confirm moves to the next cell.

### Undo / redo

Every mutation (SetCell, InsertRow, DeleteRow, MoveRow, InsertCol, DeleteCol,
MoveCol, RenameCol, SortByCol, SetColAlignment) is recorded as a command on a
stack. Undo replays the inverse. Stack depth: 100.

---

## Search and Replace

`/` opens a search bar at the bottom of the screen (replaces the command bar).
Matches are highlighted in the table. `n` / `N` navigate between matches.

The search bar has two modes toggled by a checkbox:

- **Find** — highlights matches; cursor jumps to each
- **Find & Replace** — adds a replacement input; `Enter` replaces current match,
  `Ctrl+Enter` replaces all

Patterns are Go regular expressions (`regexp` package). An invalid pattern
shows a red border on the search input; no crash.

---

## Clipboard Watch Mode

Activated by pressing `w` (or `--watch` flag in pipe mode). Two sub-modes
selectable from a status-bar toggle:

| Sub-mode | Behaviour |
|----------|-----------|
| **Format** | Watches clipboard; if table detected, reformats it in the active output format and writes back |
| **Convert** | Same, but also converts from the detected input format to the active output format |

Both modes:
- Poll the system clipboard every 200 ms (or on a clipboard-change notification
  if the platform supports it).
- Only act if the content changes and passes `Detect()` for at least one format.
- The `●` pulse indicator in the status bar animates while watch is active
  (uses `Scanner` widget).
- Pressing `w` again stops the watcher.
- The last auto-formatted content is not re-processed (deduplication by hash).

---

## File I/O

- **Open:** `tblr path` — detects format from extension, falls back to content.
- **Save:** `Ctrl+S` — overwrites original file in the same format; if the file
  was never saved (new table), prompts for a path and format.
- **Save As:** `Ctrl+Shift+S` — always prompts.
- **Export:** choosing a different format in the format picker and saving writes
  the file in the new format (no silent data loss — a confirm dialog is shown if
  the new format cannot round-trip all content, e.g. HTML losing alignment).
- The title bar shows `*` prefix when there are unsaved changes.

---

## Status Bar

```
csv · 3 rows · 4 cols · [2,1] · ← left  ·  sorted: Age ↑  ●
```

| Segment | Content |
|---------|---------|
| Format | Active format name |
| Dimensions | `N rows · M cols` |
| Cursor | `[row, col]` — 1-based |
| Alignment | Symbol + name for the focused column |
| Sort | `sorted: ColName ↑↓` when active; blank otherwise |
| Watch indicator | `●` animated when clipboard watch is active |

---

## Source File Layout

| File | Responsibility |
|------|---------------|
| `cmd/tblr/main.go` | Flag parsing, pipe/headless dispatch, TUI entry point |
| `cmd/tblr/ui.go` | `buildUI()`, Table binding, key handlers, search bar |
| `cmd/tblr/model.go` | `MutableTable`, `Alignment`, mutation methods, undo stack |
| `cmd/tblr/clipboard.go` | Clipboard watch loop, format/convert sub-modes |
| `cmd/tblr/ansi.go` | ANSI renderer; border drawing, zebra, numeric highlight, width clamping |
| `cmd/tblr/format/csv.go` | CSV + TSV parser/serializer |
| `cmd/tblr/format/markdown.go` | Markdown GFM table parser/serializer |
| `cmd/tblr/format/asciidoc.go` | AsciiDoc `\|===` parser/serializer |
| `cmd/tblr/format/typst.go` | Typst `#table(…)` parser/serializer |
| `cmd/tblr/format/html.go` | HTML `<table>` parser/serializer |
| `cmd/tblr/format/registry.go` | `Format` interface; `All()`, `ByName()`, `Detect()` |

---

## Implementation Steps

### Step 1 — Model and format package

- Implement `MutableTable` with all fields and methods, including `RecalcWidths`.
- `Columns()` returns stored widths without recomputing; callers are responsible
  for calling `RecalcWidths()` after mutations that may change cell widths.
- Implement `Format` interface and `registry.go`.
- Implement CSV/TSV format (most fundamental; shares parser via delimiter).
- Implement Markdown format.
- Tests: round-trip for CSV and Markdown; sort; MoveRow/MoveCol; alignment;
  `RecalcWidths` correctness after inserts and cell edits.

### Step 2 — CLI / pipe mode

- Flag parsing in `main.go`.
- Headless dispatch: read stdin or file, parse, serialize, write stdout or file.
- `--format` pretty-print mode.
- Tests: pipe conversion CSV→Markdown, pretty-print idempotency.

### Step 3 — TUI scaffold

- `buildUI()` with `Table` bound to an empty `MutableTable`.
- Status bar, command bar, header.
- Cell navigation and selection (cell, range, row, column, all).
- Inline cell editing.

### Step 4 — Editing operations

- Insert/delete rows and columns.
- Move rows and columns with `Alt+Arrow`.
- Sort with `s` / `S`.
- Column alignment with `<` / `>` / `^`.
- Undo/redo stack.
- Unsaved-changes indicator and quit prompt.

### Step 5 — Remaining formats

- Implement AsciiDoc, Typst, HTML formats.
- Format picker (`f` key).
- Auto-detect on paste.
- Export with format-change warning.

### Step 6 — Search and replace

- Search bar with regex; match highlighting; `n` / `N` navigation.
- Find & replace with single and replace-all.

### Step 7 — Clipboard watch

- Clipboard polling loop in `clipboard.go`.
- Format and Convert sub-modes.
- `w` key wiring; `●` scanner indicator.
- Deduplication by content hash.

---

## Non-goals (MVP)

- No formula evaluation (planned for a later step; expression language to be integrated)
- No merged cells
- No multi-table documents
- No column type inference beyond numeric-vs-string for sort
- No remote / SSH file access
- No terminal multiplexer integration

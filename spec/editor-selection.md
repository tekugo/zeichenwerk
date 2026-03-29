# Editor Selection

Adds text selection, clipboard operations, and select-all to the `Editor`
widget. Selection is the most user-visible gap in the current editor; without
it cut/copy/paste, overtype-on-type, and block deletion are all impossible.

## Selection model

Selection is defined by a **mark** (the anchor) and the **cursor** (the active
end). The region between them, in document order, is the selection.

Two new fields on `Editor`:

```go
markLine   int  // anchor line   (-1 = no selection)
markColumn int  // anchor column (-1 = no selection)
```

`markLine == -1` means no selection is active. The cursor fields `line` and
`column` are unchanged.

### Canonical start/end

Any code that needs ordered positions computes:

```go
func (e *Editor) selectionBounds() (startLine, startCol, endLine, endCol int, ok bool)
```

Returns `ok = false` when `markLine == -1` or the mark equals the cursor.
Otherwise returns the earlier position as start and the later as end, comparing
line first then column.

### Setting and clearing the mark

- **Shift+Arrow** pressed with no active mark: set mark to `(line, column)`,
  then move cursor.
- **Shift+Arrow** pressed with an active mark: keep mark, move cursor only.
- **Plain Arrow / Home / End / PgUp / PgDn / MoveTo**: clear the mark first,
  then move cursor. Exception: see *Collapsing a selection* below.
- **`ClearSelection()`**: sets `markLine = -1`.

### Collapsing a selection

When a selection is active and a plain `←`/`→` key is pressed (no Shift), most
editors collapse the selection instead of moving by one character:

| Key | Cursor lands at |
|-----|-----------------|
| `←` | selection start |
| `→` | selection end |

All other plain movement keys (`↑`, `↓`, `Home`, `End`, etc.) clear the mark
and move normally.

## New public methods

| Method | Description |
|--------|-------------|
| `HasSelection() bool` | Returns true when an active non-empty selection exists |
| `ClearSelection()` | Clears the mark; does not modify text |
| `SelectAll()` | Sets mark to `(0, 0)`, cursor to last line/last column |
| `SelectionText() string` | Returns the selected text as a string (newline-separated) |
| `DeleteSelection()` | Deletes selected text; moves cursor to start; clears mark. No-op if no selection |
| `Copy()` | Copies selection to the internal clipboard. No-op if no selection |
| `Cut()` | `Copy()` + `DeleteSelection()` |
| `Paste()` | Inserts clipboard text at cursor, replacing any active selection first |

## Changes to existing editing methods

Any editing operation that modifies text must call `DeleteSelection()` first when
a selection is active:

- `Insert(ch)` — delete selection, then insert character.
- `Delete()` (Backspace) — if selection active, `DeleteSelection()`, else existing
  logic.
- `DeleteForward()` (Delete key) — same as above.
- `Enter()` — delete selection, then split line.

This gives standard "replace-selection-on-type" behaviour.

## Clipboard

### Internal clipboard

A package-level variable holds the most recently copied text:

```go
var editorClipboard string
```

`Copy` writes to it; `Paste` reads from it. This works across multiple `Editor`
instances within the same process.

### System clipboard (opt-in)

Reading and writing the system clipboard is attempted via external commands
(`xclip -selection clipboard` on Linux, `pbcopy`/`pbpaste` on macOS) when
they are available. A helper:

```go
func systemCopy(text string) error
func systemPaste() (string, error)
```

Both fall back silently to the internal clipboard when the command is not found
or returns an error.

`Copy` and `Paste` always update the internal clipboard. System clipboard access
is attempted in addition, not instead.

## Key binding changes

| Key | Old binding | New binding |
|-----|-------------|-------------|
| `Shift+←` | — | Extend selection left |
| `Shift+→` | — | Extend selection right |
| `Shift+↑` | — | Extend selection up |
| `Shift+↓` | — | Extend selection down |
| `Shift+Home` | — | Extend selection to line start |
| `Shift+End` | — | Extend selection to line end |
| `Ctrl+A` | `DocumentHome()` | `SelectAll()` |
| `Ctrl+C` | — | `Copy()` |
| `Ctrl+X` | — | `Cut()` |
| `Ctrl+V` | — | `Paste()` |
| `Ctrl+Home` | — | `DocumentHome()` (replaces old `Ctrl+A`) |
| `Ctrl+End` | — | `DocumentEnd()` (replaces old `Ctrl+E`) |
| `Ctrl+E` | `DocumentEnd()` | unchanged (keep emacs binding) |

`tcell` exposes `KeyShiftLeft`, `KeyShiftRight`, `KeyShiftUp`, `KeyShiftDown` as
distinct key codes. `Shift+Home` and `Shift+End` arrive as `KeyShiftHome` and
`KeyShiftEnd`.

## Rendering changes

### New style selector

`Apply` registers one additional selector:

```go
theme.Apply(e, e.Selector("editor/selection"))
```

Typically a distinct background (e.g. blue) with contrasting foreground.

### Per-line rendering with selection

The existing render loop draws each visible line as a single styled string.
With selection it must render up to three segments per line:

```
[ before-selection ][ selected ][ after-selection ]
```

For each visible line `lineIdx`:

1. Compute the visible text slice (existing `getVisibleLineContent`).
2. Call `selectionBounds()`. If `ok` is false or `lineIdx` is outside
   `[startLine, endLine]`: render the line as before (no change).
3. Otherwise determine the character column range within this line that is
   selected:
   - `selStart = 0` if `lineIdx > startLine`, else `startCol`
   - `selEnd = lineLen` if `lineIdx < endLine`, else `endCol`
4. Render three segments using `r.Text`:
   - `[0, selStart)` — normal/current-line style
   - `[selStart, selEnd)` — `"editor/selection"` style
   - `[selEnd, lineLen)` — normal/current-line style
5. All three segments respect the horizontal scroll offset `offsetX` and the
   available `usableW` — only the portion visible in the viewport is drawn.

The segment coordinates are in character columns. Before rendering, map each
to visual columns via `expandTabs` (or a column-aware equivalent) to get the
correct screen x positions.

For the common case of a full-line selection (line entirely between start and
end lines), step 4 reduces to a single `r.Fill` + `r.Text` at the selection
style, which avoids re-rendering the normal segments.

## Events

No new events. The existing `EvtChange` continues to fire on text modification.
Selection state changes (mark moves) do not emit events — callers query
`HasSelection()` and `SelectionText()` on demand.

## Implementation plan

1. **`editor.go`** — data model and methods
   - Add `markLine`, `markColumn int` to `Editor` struct.
   - Add package-level `editorClipboard string`.
   - Implement `selectionBounds`, `HasSelection`, `ClearSelection`, `SelectAll`,
     `SelectionText`, `DeleteSelection`, `Copy`, `Cut`, `Paste`.
   - Update `Insert`, `Delete`, `DeleteForward`, `Enter` to call
     `DeleteSelection()` when a selection is active.
   - Update all plain movement methods (`Left`, `Right`, `Up`, `Down`, `Home`,
     `End`, `PageUp`, `PageDown`, `MoveTo`) to call `ClearSelection()` first —
     except `Left`/`Right` which collapse to selection start/end instead.
   - Add Shift+Arrow variants: `ShiftLeft`, `ShiftRight`, `ShiftUp`, `ShiftDown`,
     `ShiftHome`, `ShiftEnd` — each sets the mark if unset, then moves the cursor.

2. **`editor.go`** — key handler
   - Add cases for `KeyShiftLeft`, `KeyShiftRight`, `KeyShiftUp`, `KeyShiftDown`,
     `KeyShiftHome`, `KeyShiftEnd`.
   - Change `KeyCtrlA` from `DocumentHome()` to `SelectAll()`.
   - Add `KeyCtrlC` → `Copy()`, `KeyCtrlX` → `Cut()`, `KeyCtrlV` → `Paste()`.
   - Add `KeyCtrlHome` → `DocumentHome()`.

3. **`editor.go`** — rendering
   - Register `"editor/selection"` in `Apply`.
   - Refactor the per-line render block into a helper
     `renderLine(r, lineIdx, textX, textY, usableW)` that handles the three-segment
     logic, keeping `Render` readable.

4. **`editor.go`** — system clipboard helpers
   - Implement `systemCopy(text string) error` and `systemPaste() (string, error)`
     using `os/exec` to call `xclip`/`pbcopy` and `xclip -o`/`pbpaste`.
   - Both silently fall back to the internal clipboard on error.

5. **Theme** — add `"editor/selection"` style to built-in themes.

6. **Tests** — `editor_test.go`
   - `selectionBounds` returns correct ordered positions for mark-before-cursor
     and mark-after-cursor cases.
   - Shift+Arrow extends selection; plain Arrow collapses it.
   - `SelectAll` covers the entire document.
   - `SelectionText` returns correct multi-line text.
   - `DeleteSelection` removes the selected range and positions cursor at start.
   - `Insert` while selection active replaces the selection.
   - `Copy` + `ClearSelection` + `Paste` round-trips the clipboard text.
   - Rendering: selected range on a partially selected line produces correct
     segment boundaries (visual column mapping verified for tab-containing lines).

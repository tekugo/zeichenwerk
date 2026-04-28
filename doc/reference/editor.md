# Editor

Multi-line text editor with cursor, selection, copy/paste, optional line numbers, and auto-indent.

**Constructor:** `NewEditor(id, class string) *Editor`

## Content access

- `Text() string` — full content as a single string
- `Lines() []string` — content as a slice of lines
- `Load(text string)` — replace content from a string
- `SetContent(lines []string)` — replace content from a line slice

## Configuration

- `SetAutoIndent(auto bool)` — copy leading whitespace from the previous line on Enter
- `SetReadOnly(ro bool)` — disable editing without removing focusability
- `SetTabWidth(width int)` — visual width of a tab character
- `UseSpaces(useSpaces bool)` — insert spaces instead of tab characters
- `ShowLineNumbers(show bool)` — toggle the gutter

## Cursor / movement

- `Left() / Right() / Up() / Down()` — move one character / line
- `Home() / End()` — start / end of current line
- `DocumentHome() / DocumentEnd()` — start / end of document
- `PageUp() / PageDown()` — scroll by one page
- `MoveTo(line, column int)` — jump to position

## Editing

- `Insert(ch rune)` — insert a character at the cursor
- `Delete()` — backspace (delete before the cursor)
- `DeleteForward()` — delete at the cursor
- `Enter()` — insert a newline with optional auto-indent

## Selection

- `HasSelection() bool`
- `SelectAll()`, `ClearSelection()`
- `SelectionText() string` — selected text
- `DeleteSelection()`
- `ShiftLeft() / ShiftRight() / ShiftUp() / ShiftDown() / ShiftHome() / ShiftEnd()` — extend selection by movement

## Clipboard

- `Copy()` — copy the selection (uses the system clipboard via `atotto/clipboard`)
- `Cut()` — copy + delete
- `Paste()` — insert clipboard contents at the cursor

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | — | Content modified |

## Notes

Flags: `"focusable"`, `"readonly"`.

Backed by a gap-buffer per line for efficient editing in long files. Long lines are not wrapped; they scroll horizontally.

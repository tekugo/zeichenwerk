# Editor

Multi-line text editing widget.

**Constructor:** `NewEditor(id, class string) *Editor`

## Methods

- `Delete()` — backspace (deletes before cursor)
- `DeleteForward()` — delete (removes at cursor)
- `DocumentEnd()` — moves to end of document
- `DocumentHome()` — moves to start of document
- `Down()` — moves cursor down one line
- `End()` — moves to end of current line
- `Enter()` — inserts new line with auto-indent
- `Home()` — moves to start of current line
- `Insert(ch rune)` — inserts character at cursor
- `Left()` — moves cursor left
- `Lines() []string` — returns content as line slice
- `Load(text string)` — sets content from string
- `MoveTo(line, column int)` — moves cursor to position
- `PageDown()` — scrolls down one page
- `PageUp()` — scrolls up one page
- `Right()` — moves cursor right
- `SetAutoIndent(auto bool)` — enables auto-indentation
- `SetContent(lines []string)` — sets content from line slice
- `SetReadOnly(ro bool)` — sets read-only mode
- `SetTabWidth(width int)` — sets tab width
- `Text() string` — returns all content as string
- `Up()` — moves cursor up one line
- `UseSpaces(useSpaces bool)` — inserts spaces instead of tabs

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | — | Content modified |

# Zeichenwerk — Widget Reference

Full style keys, events, methods, and constraints for every widget.
This file is referenced by [SKILL.md](SKILL.md).

---

## Style selector syntax

| Pattern | Meaning |
|---------|---------|
| `"button"` | Base style |
| `"button.dialog"` | Class variant |
| `"button:focused"` | State variant |
| `"button.dialog:focused"` | Class + state |
| `"table/grid"` | Sub-part |
| `"table/grid:focused"` | Sub-part + state |

---

## Box

```go
NewBox(id, class, title string) *Box
```

| Style key | When |
|-----------|------|
| `"box"` | always |
| `"box:focused"` / `":disabled"` / `":hovered"` | state |
| `"box/title"` | title bar |

Events: none. `Add(Widget) error` — single child only.

---

## Button

```go
NewButton(id, class, text string) *Button
```

| Style key | When |
|-----------|------|
| `"button"` | always |
| `"button:focused"` / `":hovered"` / `":pressed"` / `":disabled"` | state |
| `"button.dialog"` / `"button.dialog:focused"` / `"button.dialog:hovered"` | inside dialog |

Events: `EvtActivate` (payload `0`). Methods: `Activate()`, `Set(any) bool`.

---

## Canvas

```go
NewCanvas(id, class string, pages, width, height int) *Canvas
```

Style key: `"canvas"`.
Events: `EvtChange`, `EvtMove(x,y int)`, `EvtMode(mode string)`.
Modes: `ModeNormal` `ModeInsert` `ModeCommand` `ModeDraw` `ModePresent` `ModeVisual`.
Methods: `Cell(x,y)`, `SetCell(x,y,ch,style)`, `SetCursor(x,y)`, `SetMode(string)`,
`SetPage(int)`, `Resize(pages,rows,cols)`, `Clear()`, `Fill(ch,style)`.

---

## Checkbox

```go
NewCheckbox(id, class, text string, checked bool) *Checkbox
```

| Style key | When |
|-----------|------|
| `"checkbox"` | always |
| `":checked"` / `":focused"` / `":hovered"` / `":disabled"` | state |

Events: `EvtChange(bool)`. Methods: `Toggle()`, `Set(any) bool`.
`FlagReadonly` prevents user toggling.

---

## Collapsible

```go
NewCollapsible(id, class, title string, expanded bool) *Collapsible
```

| Style key | When |
|-----------|------|
| `"collapsible"` | always |
| `":focused"` / `":hovered"` | state |
| `"collapsible/header"` | header row |
| `"collapsible/header:focused"` / `":hovered"` | header state |

Theme strings: `"collapsible.expanded"`, `"collapsible.collapsed"`.
Events: `EvtChange(bool)`.
Methods: `Add(Widget) error`, `Expand()`, `Collapse()`, `Toggle()`, `Expanded() bool`.
Single child only.

---

## Deck

```go
NewDeck(id, class string, render ItemRender, itemHeight int) *Deck

type ItemRender func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool)
```

| Style key | When |
|-----------|------|
| `"deck"` | always |
| `":focused"` / `":hovered"` / `":disabled"` | state |

Events: `EvtSelect(int)`, `EvtActivate(int)`.
Methods: `SetItems([]any)`, `Items() []any`, `SetDisabled([]int)`,
`Selected() int`, `Select(int)`, `Move(int)`, `First()`, `Last()`, `PageUp()`, `PageDown()`.
**itemHeight must be ≥ 1 — panics otherwise.**
`focused` in `ItemRender` is `true` when the Deck widget holds keyboard focus.

---

## Dialog

```go
NewDialog(id, class, title string) *Dialog
```

Style keys: `"dialog"`, `"dialog/title"`.
Events: none. Single child. Shown via `ui.Popup()`, not added to main tree.

---

## Editor

```go
NewEditor(id, class string) *Editor
```

| Style key | When |
|-----------|------|
| `"editor"` | always |
| `":focused"` / `":disabled"` | state |
| `"editor/current-line"` | current line |
| `"editor/current-line-number"` | current line number |
| `"editor/line-numbers"` | number column |
| `"editor/selection"` | selected text |

Events: `EvtChange`.
Methods: `SetContent([]string)`, `Load(string)`, `Lines() []string`, `Text() string`,
`SetTabWidth(int)`, `UseSpaces(bool)`, `ShowLineNumbers(bool)`, `SetAutoIndent(bool)`, `SetReadOnly(bool)`.

---

## Flex

```go
NewFlex(id, class string, horizontal bool, alignment string, spacing int) *Flex
// alignment: "start" | "center" | "end" | "stretch"
```

Style key: `"flex"`. Events: none. `Add(Widget) error`.
Child hint: positive=fixed, zero=natural, negative=fractional share.

---

## Form

```go
NewForm(id, class, title string, data any) *Form
// data must be a pointer to a struct
```

Struct tags: `label:"…"`, `control:"input|checkbox|password|select"`,
`options:"a,b,c"`, `readonly`, `width:"20"`, `line:"0"`, `group:"name"`.

Style key: `"form"`. Events: none (controls emit `EvtChange`).
Methods: `Data() any`, `Update(reflect.Value) Handler`.
Builder logs warning and skips if data is not pointer-to-struct.

---

## Grid

```go
NewGrid(id, class string, rows, columns int, lines bool) *Grid
```

Style key: `"grid"`. Events: none.
`Add(Widget, x, y, colspan, rowspan int) error`.
`Columns(sizes ...int)`, `Rows(sizes ...int)` — positive=fixed, negative=fractional, zero=auto.
`Builder.Columns()`/`Rows()` log warning and no-op outside a Grid context.

---

## Input

```go
NewInput(id, class string, params ...string) *Input
// params[0]=placeholder  params[1]=initial text  params[2]=mask char
```

| Style key | When |
|-----------|------|
| `"input"` | always |
| `":focused"` / `":disabled"` / `":hovered"` | state |

Events: `EvtChange(string)`.
Methods: `SetText(string)`, `Text() string`, `SetMask(string)`,
`Left()`, `Right()`, `Start()`, `End()`.
Single-line only. `FlagMasked` hides text.

---

## List

```go
NewList(id, class string, items []string) *List
```

| Style key | When |
|-----------|------|
| `"list"` | always |
| `":focused"` / `":disabled"` / `":hovered"` | state |
| `"list/highlight"` | highlighted item, unfocused |
| `"list/highlight:focused"` | highlighted item, focused |

Events: `EvtSelect(int)`, `EvtActivate(int)`.
Methods: `SetItems([]string)`, `Items() []string`, `Select(int)`, `Selected() int`.

---

## Progress

```go
NewProgress(id, class string, horizontal bool) *Progress
```

Style keys: `"progress"`, `"progress/bar"`. Events: none.
Methods: `SetValue(int)`, `SetTotal(int)` (0=indeterminate), `Increment(int)`, `Percentage() float64`.

Theme strings (replace `h` with `v` for vertical):
`"progress.h.prefix"` `"progress.h.suffix"`
`"progress.h.start.filled"` `"progress.h.middle.filled"` `"progress.h.end.filled"`
`"progress.h.start.empty"` `"progress.h.middle.empty"` `"progress.h.end.empty"`

---

## Rule

```go
NewHRule(class, borderStyle string) *Rule  // height=1
NewVRule(class, borderStyle string) *Rule  // width=1
```

Style key: `"rule"`. Events: none. Non-interactive.

---

## Select

```go
NewSelect(id, class string, args ...string) *Select
// args: flat pairs — value1, label1, value2, label2, ...
```

| Style key | When |
|-----------|------|
| `"select"` | always |
| `":focused"` / `":disabled"` / `":hovered"` | state |

Theme string: `"select.dropdown"` — dropdown indicator glyph.
Events: `EvtChange(string)` — selected value.
Methods: `Select(string)`, `Value() string`, `Text() string`.

---

## Static

```go
NewStatic(id, class, text string) *Static
```

Style key: `"static"`. Events: none.
Methods: `SetText(string)`, `SetAlignment(string)` (`"left"` `"center"` `"right"`), `Set(any) bool`.

---

## Styled

```go
NewStyled(id, class, text string) *Styled
```

Markdown: `# h1` `## h2` `- list` ` ``` code ``` ` `*italic*` `**bold**`
`__underline__` `~~strikethrough~~` `` `code` ``.
Style key: `"styled"`. Events: none.
Methods: `SetText(string)`, `Parse()`.

---

## Switcher

```go
NewSwitcher(id, class string) *Switcher
```

Style key: `"switcher"`. Events: `EvtShow`, `EvtHide`.
Methods: `Add(Widget) error`, `Select(any)` (int index or string id), `Children() []Widget`.
All panes fill switcher bounds; only selected pane is visible and interactive.

---

## Table

```go
NewTable(id, class string, provider TableProvider) *Table

type TableProvider interface {
    Columns() []TableColumn   // header + width per column
    Length() int              // number of data rows
    Str(row, column int) string
}
// Convenience constructor:
NewArrayTableProvider(headers []string, rows [][]string)
```

| Style key | When |
|-----------|------|
| `"table"` | always |
| `":focused"` | focused |
| `"table/grid"` | grid lines |
| `"table/grid:focused"` | grid lines, focused |
| `"table/header"` | header row |
| `"table/highlight"` | highlighted row, unfocused |
| `"table/highlight:focused"` | highlighted row, focused |
| `"table/cell"` | focused cell in cell mode, unfocused |
| `"table/cell:focused"` | focused cell in cell mode, focused |

Events: `EvtSelect(row int, col int)` · `EvtActivate(row int, rowData []string)`.  
`col` is `-1` in row mode, column index in cell mode.

**Row mode keyboard** (default): `↑`/`↓` rows · `←`/`→` scroll 1 char ·
`Ctrl+←`/`Ctrl+→` scroll by column · `PgUp`/`PgDn` · `Home`/`End` first/last row ·
`Ctrl+Home`/`Ctrl+End` first/last row + reset scroll · `Enter` activate · `Space` select.

**Cell mode keyboard** (`SetCellNav(true)`): `↑`/`↓` rows (same) · `←`/`→`
prev/next column · `Ctrl+←`/`Ctrl+→` first/last column · `Home`/`End` first/last
column in row · `Ctrl+Home`/`Ctrl+End` first/last row+column · `Tab`/`Shift+Tab`
next/prev cell wrapping across rows · `Enter` activate · `Space` select.

Methods: `Selected() (row, col int)` · `SetSelected(row, col int) bool` ·
`CellNav() bool` · `SetCellNav(bool)` · `Offset() (x, y int)` · `SetOffset(x, y int)` ·
`Set(provider)` · `Refresh()`.

---

## Tabs

```go
NewTabs(id, class string) *Tabs
```

| Style key | When |
|-----------|------|
| `"tabs"` | always |
| `"tabs/highlight"` | highlighted tab, unfocused |
| `"tabs/highlight:focused"` | highlighted tab, focused |
| `"tabs/line"` | underline bar |
| `"tabs/line:focused"` | underline bar, focused |
| `"tabs/highlight-line"` | selected tab underline, unfocused |
| `"tabs/highlight-line:focused"` | selected tab underline, focused |

Events: `EvtChange(int)`, `EvtActivate(int)`.
Methods: `Add(title string)`, `Select(int) bool`, `Selected() int`, `Count() int`.
Keyboard: Left/Right (wrap) · Home/End · letter keys (first-letter jump) · Enter activate.

---

## Terminal

```go
NewTerminal(id, class string) *Terminal
```

Renders arbitrary terminal output by interpreting ANSI/VT escape sequences.
Implements `io.Writer` — pipe a pty or subprocess directly into it.

| Style key | When |
|-----------|------|
| `"terminal"` | always |
| `"terminal:focused"` | focused |

Events: none.

Methods:

| Method | Description |
|--------|-------------|
| `Write([]byte) (int, error)` | Feed raw bytes (pty output, ANSI sequences). Thread-safe. |
| `Clear()` | Clear both buffers, reset cursor and scroll region. |
| `Resize(w, h int)` | Resize both buffers; called automatically by `Render`. |
| `Title() string` | Returns the last OSC 0/1/2 window title received. |

**Sequences handled:**

| Category | Sequences |
|----------|-----------|
| Cursor movement | CUU `A` · CUD `B` · CUF `C` · CUB `D` · CNL `E` · CPL `F` · CHA `G` · CUP `H` · VPA `d` · HVP `f` |
| Erase | ED `J` (0–3) · EL `K` (0–2) · ECH `X` · DCH `P` |
| Insert/delete | IL `L` · DL `M` |
| Scroll | SU `S` · SD `T` · RI (reverse index `ESC M`) |
| SGR | Reset · bold · dim · italic · blink · reverse · invisible · strikethrough · underline styles (single/double/curly/dotted/dashed) · 16/256/true-colour fg/bg/underline |
| Modes | `?7` auto-wrap · `?25` cursor visibility · `?1049` alternate screen |
| Scroll region | DECSTBM `r` |
| Save/restore | DECSC `ESC 7` · DECRC `ESC 8` · SCOSC `s` · SCORC `u` |
| Reset | RIS `ESC c` |
| OSC | 0/1/2 window title |

**SGR colour formats:**

```
\033[31m          # palette 0–15 (standard colours)
\033[38;5;196m    # xterm-256 palette
\033[38;2;R;G;Bm  # true colour (24-bit)
\033[4:3m          # underline style (0=off 1=single 2=double 3=curly 4=dotted 5=dashed)
\033[58;2;R;G;Bm  # underline colour (Kitty extension)
```

**Typical usage — pty subprocess:**

```go
term := NewTerminal("term", "")
term.Apply(theme)
// wire into UI …

// in a goroutine:
cmd := exec.Command("bash")
cmd.Stdout = term  // Terminal is io.Writer
cmd.Start()
```

**Typical usage — static ANSI content:**

```go
term := NewTerminal("term", "")
term.Write([]byte("\033[1;32mHello\033[0m \033[38;2;255;100;0mWorld\033[0m\n"))
term.Write([]byte("\033[4:3mCurly underline\033[0m\n"))
```

Default buffer size is 80×24. `Render` automatically resizes to the content area on
the first frame and preserves existing content. Call `SetHint(0, -1)` to let the
layout engine size the widget, or `SetHint(80, 24)` to fix it.

**v1 limitations:** no scroll-back buffer, no sixel/kitty graphics, no mouse
reporting, no DCS, no BIDI.

---

## Text

```go
NewText(id, class string, content []string, follow bool, max int) *Text
// follow: auto-scroll to bottom on new content
// max: line limit (0=unlimited)
```

Style key: `"text"`. Events: none.
Methods: `Add(lines ...string)`, `Clear()`, `Set(any) bool`.

---

## Tree

```go
NewTree(id, class string) *Tree

node := NewTreeNode("label")
node.Add(NewTreeNode("child"))
node.SetLoader(func(n *TreeNode) { n.Add(NewTreeNode("lazy")) }) // called once on first expand
tree.Add(node)
```

| Style key | When |
|-----------|------|
| `"tree"` | always |
| `":focused"` / `":disabled"` / `":hovered"` | state |
| `"tree/highlight"` | highlighted node, unfocused |
| `"tree/highlight:focused"` | highlighted node, focused |
| `"tree/indent"` | indent lines |

Theme strings: `"tree.expanded"` `"tree.collapsed"` `"tree.branch"` `"tree.last"` `"tree.trunk"`.
Events: `EvtSelect(*TreeNode)`, `EvtActivate(*TreeNode)`, `EvtChange(*TreeNode)`.
Methods: `Add(*TreeNode)`, `SetRoot(*TreeNode)`, `Selected() *TreeNode`,
`Select(*TreeNode)`, `Move(int)`, `First()`, `Last()`, `Expand(*TreeNode)`, `Collapse(*TreeNode)`.
Filesystem provider: `NewTreeFS(root string) *TreeFS`.

---

## Typeahead

```go
NewTypeahead(id, class string, params ...string) *Typeahead
// Same params as Input: placeholder, initial text, mask
```

| Style key | When |
|-----------|------|
| `"typeahead"` | always |
| `":focused"` / `":disabled"` / `":hovered"` | state |
| `"typeahead/suggestion"` | ghost text, unfocused |
| `"typeahead/suggestion:focused"` | ghost text, focused |

Events: `EvtChange(string)`, `EvtAccept(string)`.
All `Input` methods plus `SetSuggest(func(string) []string)`.
Tab or Right-arrow at end-of-text accepts suggestion.

---

## Viewport

```go
NewViewport(id, class, title string) *Viewport
```

Style key: `"viewport"`. Events: none.
`Add(Widget) error` — single child given its full preferred size.

---

## Theme strings — complete list

```
collapsible.expanded        collapsible.collapsed
progress.h.prefix           progress.h.suffix
progress.h.start.filled     progress.h.middle.filled    progress.h.end.filled
progress.h.start.empty      progress.h.middle.empty     progress.h.end.empty
progress.v.prefix           progress.v.suffix
progress.v.start.filled     progress.v.middle.filled    progress.v.end.filled
progress.v.start.empty      progress.v.middle.empty     progress.v.end.empty
select.dropdown
tree.expanded   tree.collapsed   tree.branch   tree.last   tree.trunk
```

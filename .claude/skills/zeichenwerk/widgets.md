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

## BarChart

```go
type BarSeries struct {
    Label  string
    Values []float64  // one per category, all >= 0
}

NewBarChart(id, class string) *BarChart
```

| Style key | When |
|-----------|------|
| `"bar-chart"` | always |
| `"bar-chart/s0"` … `"bar-chart/s7"` | series colours (s0 = bottom of stack) |
| `"bar-chart/axis"` | axis lines |
| `"bar-chart/grid"` | grid lines |
| `"bar-chart/label"` | category labels, unfocused |
| `"bar-chart/label:focused"` | focused category label |
| `"bar-chart/selection"` | selected category background |
| `"bar-chart/value"` | value labels above bars |
| `"bar-chart/legend"` | legend text |

Theme strings: `"bar-chart.corner"` `"bar-chart.hline"` `"bar-chart.vline"`
`"bar-chart.tick-x"` `"bar-chart.tick-y"` `"bar-chart.grid"` `"bar-chart.swatch"`.

Events: `EvtSelect(int)` — focused category changed · `EvtActivate(int)` — Enter or double-click.

Methods: `SetSeries([]BarSeries)`, `AddSeries(BarSeries)`, `SetCategories([]string)`,
`Series() []BarSeries`, `Categories() []string`,
`SetMode(ScaleMode)` (`Relative`/`Absolute`), `SetMax(float64)`,
`SetHorizontal(bool)`, `SetShowAxis(bool)`, `SetShowGrid(bool)`,
`SetShowValues(bool)`, `SetLegend(bool)`,
`SetBarWidth(int)`, `SetBarGap(int)`, `SetTicks(int)`,
`Select(int)`, `Selected() int`.

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

## Breadcrumb

```go
NewBreadcrumb(id, class string) *Breadcrumb
```

| Style key | When |
|-----------|------|
| `"breadcrumb"` | always |
| `"breadcrumb/segment"` | individual segment, unfocused |
| `"breadcrumb/segment:focused"` | focused/selected segment |
| `"breadcrumb/separator"` | separator between segments |

Theme strings: `"breadcrumb.separator"` (default `" › "`) · `"breadcrumb.overflow"` (default `"…"`).

Events: `EvtSelect(int)` — focused segment index · `EvtActivate(int)` — Enter or click.

Methods: `SetSegments([]string)`, `Select(int)`, `Selected() int`,
`SetSeparator(string)`, `SetOverflow(string)`.

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

## Combo

```go
NewCombo(id, class string, items []string) *Combo
```

Collapsed single-line display with a dropdown popup (Typeahead + List).
Opens on focus or Enter; Esc closes without confirming.

| Style key | When |
|-----------|------|
| `"combo"` | always |
| `"combo:focused"` | focused |

Events: `EvtChange(string)` — every keystroke in popup · `EvtActivate(string)` — confirmed value.

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

## CRT

```go
NewCRT(id, class string) *CRT
```

Animated root container that simulates a CRT monitor powering on/off.
During normal operation it is a zero-overhead pass-through wrapper.

`Add(Widget) error` — single child. `Start(interval)` — power-on animation.
`PowerOff(interval, onDone func())` — power-off animation, then calls `onDone`.

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

## Heatmap

```go
NewHeatmap(id, class string, rows, cols int) *Heatmap
```

| Style key | When |
|-----------|------|
| `"heatmap"` | always |
| `"heatmap/header"` | column label row |
| `"heatmap/zero"` | cells with value 0 |
| `"heatmap/mid"` | cells with intermediate values |
| `"heatmap/max"` | cells at or near maximum |

Events: none. Display only.

Methods: `SetValue(row, col int, v float64)`, `SetRow(row int, vs []float64)`,
`SetAll(vs [][]float64)`, `Value(row, col int) float64`,
`SetRowLabels([]string)`, `SetColLabels([]string)`, `SetCellWidth(int)`.

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

## Sparkline

```go
NewSparkline(id, class string) *Sparkline
```

Renders a sequence of `float64` values as a column of Unicode block characters
(`▁▂▃▄▅▆▇█`). Height is determined by the widget's content area: `h` rows
give `h×8` discrete levels per column.

| Style key | When |
|-----------|------|
| `"sparkline"` | default bar colour |
| `"sparkline/high"` | bars whose value is ≥ threshold |

`ScaleMode` constants: `Relative` (tallest visible bar = █), `Absolute` (fixed min/max).

Methods: `Append(float64)`, `SetValues([]float64)`, `Values() []float64`,
`SetMode(ScaleMode)`, `SetMin(float64)`, `SetMax(float64)`,
`SetThreshold(float64)` (0 = disabled), `SetGradient(bool)`,
`SetCapacity(int)` (0 = unlimited ring buffer).

**Gradient mode** (`SetGradient(true)`): instead of a hard colour cutoff at the
threshold, the foreground interpolates linearly from the `"sparkline"` colour
(at the threshold) to the `"sparkline/high"` colour (at the maximum value).
Has no effect when threshold is 0.

Events: none. Non-interactive (display only).

---

## Scanner

```go
NewScanner(id, class string, width int, style string) *Scanner
// style: one of "blocks", "circles", "diamonds"
```

Animated back-and-forth scanning bar with a fading trail. Embeds `Animation`;
call `Start(interval)` / `Stop()`. Style key: `"scanner"`. Events: none.

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

## Spinner

```go
NewSpinner(id, class string, sequence string) *Spinner
```

Cycling single-character animation. `sequence` is a space-separated list of
glyphs, e.g. `Spinners["braille"]`. Embeds `Animation`; call `Start(interval)` / `Stop()`.

```go
// Built-in sequences (Spinners map):
// "bar"     "dots"   "dot"   "arrow"
// "circle"  "bounce" "braille"
sp := NewSpinner("sp", "", Spinners["braille"])
```

Style key: `"spinner"`. Events: none.

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

## Marquee

```go
NewMarquee(id, class string) *Marquee
```

Continuously scrolling single-row text ticker. Text wider than the widget
scrolls left; pauses when `FlagHovered` is set. Embeds `Animation`.
Does **not** set `FlagFocusable` — display only.

Style key: `"marquee"`. Events: none.

Methods: `SetText(string)`, `Text() string`,
`SetSpeed(int)` (columns per tick, min 1), `SetGap(int)` (blank columns after text before loop).
`Start(time.Duration)`, `Stop()`, `Running() bool`.

---

## Shimmer

```go
NewShimmer(id, class string) *Shimmer
```

Sweeping highlight band animation. A band of accent colour moves left-to-right
across the text each tick, blending from base to band colour and back.
Multi-line text is supported — the band sweeps the same column on all rows.
Embeds `Animation`. Does **not** set `FlagFocusable` — display only.

Defaults: `bandWidth=6`, `edgeWidth=3`, gradient off.

| Style key | Purpose |
|-----------|---------|
| `"shimmer"` | base text colour and background |
| `"shimmer/band"` | foreground at full band intensity (bg ignored) |

Events: none.

Methods: `SetText(string)`, `Text() string`,
`SetBandWidth(int)` (core width in cols, min 1),
`SetEdgeWidth(int)` (fade cols per side; 0 = hard edge),
`SetGradient(bool)` (false = stepped linear ramp; true = smooth cosine bell).
`Start(time.Duration)`, `Stop()`, `Running() bool`.

```go
sh := NewShimmer("status", "")
sh.SetText("Analysing codebase…")
sh.SetBandWidth(10).SetEdgeWidth(5).SetGradient(true)
sh.Start(40 * time.Millisecond)
```

---

## Typewriter

```go
NewTypewriter(id, class string) *Typewriter
```

Animated character-by-character text reveal with optional blinking cursor.
Embeds `Animation`. Does **not** set `FlagFocusable` — display only.

| Style key | When |
|-----------|------|
| `"typewriter"` | text background and colour |
| `"typewriter/cursor"` | cursor character |

Theme string: `"typewriter.cursor"` (default `"▌"`).

Events: `EvtChange(bool=true)` — reveal complete · `EvtActivate(bool=true)` — animation done (`repeat=false`).

Methods: `SetText(string)`, `Text() string`,
`SetRate(int)`, `SetCursor(bool)`, `SetDwell(time.Duration)`, `SetRepeat(bool)`,
`Reset()`, `Start(time.Duration)`, `Stop()`, `Running() bool`.

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
bar-chart.corner    bar-chart.hline    bar-chart.vline
bar-chart.tick-x    bar-chart.tick-y   bar-chart.grid    bar-chart.swatch

breadcrumb.separator    breadcrumb.overflow

collapsible.expanded    collapsible.collapsed

progress.h.prefix           progress.h.suffix
progress.h.start.filled     progress.h.middle.filled    progress.h.end.filled
progress.h.start.empty      progress.h.middle.empty     progress.h.end.empty
progress.v.prefix           progress.v.suffix
progress.v.start.filled     progress.v.middle.filled    progress.v.end.filled
progress.v.start.empty      progress.v.middle.empty     progress.v.end.empty

select.dropdown

shortcuts.prefix    shortcuts.separator    shortcuts.suffix

tree.expanded   tree.collapsed   tree.branch   tree.last   tree.trunk

typewriter.cursor
```

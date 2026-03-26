# Zeichenwerk API Reference

A TUI component library based on tcell/v3.

## Core Interfaces

### Widget

The fundamental interface for all UI components.

| Method | Description |
|--------|-------------|
| `Bounds() (x, y, w, h int)` | Returns absolute screen coordinates |
| `Content() (x, y, w, h int)` | Returns inner content area coordinates |
| `Cursor() (x, y int, style string)` | Returns cursor position relative to content |
| `Dispatch(widget Widget, event string, data ...any) bool` | Dispatches event to handlers |
| `Flag(name string) bool` | Gets boolean state flag |
| `Hint() (w, h int)` | Returns preferred content size |
| `ID() string` | Returns unique identifier |
| `Info() string` | Returns human-readable description |
| `Log(source Widget, level, msg string, params ...any)` | Logs debug message |
| `On(event string, handler Handler)` | Registers event handler |
| `Parent() Container` | Returns parent container |
| `Refresh()` | Queues widget redraw |
| `Render(r *Renderer)` | Renders widget to screen |
| `SetBounds(x, y, w, h int)` | Sets absolute position and size |
| `SetFlag(name string, value bool)` | Sets boolean state flag |
| `SetHint(w, h int)` | Sets preferred content size |
| `SetParent(parent Container)` | Sets parent container |
| `SetStyle(selector string, style *Style)` | Applies style for selector |
| `State() string` | Returns widget state for rendering |
| `Style(selector ...string) *Style` | Returns style for selector |

**Common events:**

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | varies | Content or state changed |
| `"key"` | `*tcell.EventKey` | Keyboard event |
| `"mouse"` | `*tcell.EventMouse` | Mouse event |

### Container

Extends Widget with child management capabilities.

| Method | Description |
|--------|-------------|
| `Children() []Widget` | Returns all direct child widgets |
| `Layout()` | Arranges child widgets |

**Helper functions:**
- `Find(container Container, id string) Widget` - Finds widget by ID in hierarchy
- `FindAll[T any](container Container) []T` - Finds all widgets of type T
- `FindAt(container Container, x, y int) Widget` - Finds widget at coordinates
- `Layout(container Container)` - Recursively lays out all child containers
- `Traverse(Container, func(Widget) bool)` - Depth-first traversal

## Base Types

### Component

Embedded struct providing default Widget implementation.

**Public fields:** None (use constructor `NewComponent(id string)`)

**Key methods (override in embedding structs):**
- `Cursor() (int, int, string)` - Returns `(0, 0, "")` by default
- `Render(r *Renderer)` - Must be implemented by concrete widgets
- `Info() string` - Returns debug info

**Internal methods:**
- `Dispatch(widget Widget, event string, data ...any) bool`
- `Flag(name string) bool`
- `Hint() (int, int)`
- `ID() string`
- `Log(source Widget, level, msg string, params ...any)`
- `On(event string, handler Handler)`
- `Parent() Container`
- `Refresh()`
- `SetBounds(x, y, w, h int)`
- `SetFlag(name string, value bool)`
- `SetHint(w, h int)`
- `SetParent(parent Container)`
- `SetStyle(selector string, style *Style)`
- `State() string`
- `Style(selector ...string) *Style`

### Animation

Embedded struct for timed animations. Manages ticker and stop channel.

| Method | Description |
|--------|-------------|
| `Start(interval time.Duration)` | Starts animation in goroutine |
| `Stop()` | Stops animation gracefully |
| `Running() bool` | Returns true if animation active |
| `Refresh()` | Triggers widget redraw |
| `Tick()` | Called on each frame (subtypes must override) |

**Fields (embedding):**
- `Component` - Base component
- `ticker *time.Ticker` - Internal ticker (nil when stopped)
- `stop chan struct{}` - Stop signal channel (initialized in constructor)
- `fn func()` - Tick callback function

## UI

The root application class managing screen, events, and rendering.

| Method | Description |
|--------|-------------|
| `NewUI(theme *Theme, root Container, debug bool) (*UI, error)` | Constructor |
| `NewBuilder() *Builder` | Creates builder with current theme |
| `Handle(event tcell.Event) bool` | Processes tcell events |
| `Run() error` | Starts main event loop (blocks) |
| `EventLoop()` | Polls tcell events (goroutine) |
| `Draw()` | Renders entire UI |
| `DrawWidget(widget Widget)` | Renders single widget |
| `Redraw(widget Widget)` | Queues widget for individual redraw |
| `Refresh()` | Queues full screen redraw |
| `Layout()` | Recalculates layout for all layers |
| `Popup(x, y, w, h int, popup Container)` | Shows container as overlay |
| `Close()` | Removes topmost layer |
| `Focus(widget Widget)` | Sets keyboard focus |
| `SetFocus(which string)` | Navigates focus ("first", "last", "next", "previous") |
| `ShowCursor()` | Positions and shows cursor |
| `ShowDebug()` | Renders debug info bar |
| `SetTheme(theme *Theme)` | Changes active theme |
| `Theme() *Theme` | Returns current theme |
| `Log(source Widget, levelStr, msg string, params ...any)` | Adds structured log entry |
| `SetLogLevel(level slog.Level)` | Changes log level at runtime |
| `Logs() *TableLog` | Returns table log widget |

**Keyboard shortcuts (handled by UI):**

| Keys | Action |
|------|--------|
| `Tab`, `Right`, `Down` | Next focusable widget |
| `Backtab`, `Left`, `Up` | Previous focusable widget |
| `Escape` | Close topmost popup |
| `Ctrl+C`, `Ctrl+Q` | Quit application |
| `Ctrl+D` | Open inspector popup (debug mode) |
| `q`, `Q` | Quit application |

## Builder

Fluent API for constructing UIs.

**Construction:**
```go
NewBuilder(theme *Theme) *Builder
```

**Building:**
- `Build() *UI` - Returns UI instance
- `Run() *UI` - Builds and runs (blocks)
- `End() *Builder` - Pops current container from stack
- `Container() Container` - Returns top-level container
- `Find(id string) Widget` - Finds widget by ID
- `With(fn func(*Builder)) *Builder` - Composition helper

**Creating widgets (all return `*Builder`):**

| Method | Description |
|--------|-------------|
| `Box(id, title string)` | Creates a bordered box with title |
| `Button(id, text string)` | Creates a clickable button |
| `Checkbox(id, text string, checked bool)` | Creates a toggle checkbox |
| `Dialog(id, title string)` | Creates a dialog container |
| `Digits(id, text string)` | Creates large-format numeric display |
| `Editor(id string)` | Creates a multi-line text editor |
| `Flex(id string, horizontal bool, alignment string, spacing int)` | Creates a linear layout container |
| `Form(id, title string, data any)` | Creates a data-bound form |
| `Group(id, title, groupName string, horizontal bool, spacing int)` | Creates a form group |
| `Grid(id string, rows, columns int, lines bool)` | Creates a grid layout container |
| `HRule(style string)` | Creates a horizontal rule |
| `Input(id string, params ...string)` | Creates a single-line text input |
| `List(id string, values ...string)` | Creates a selectable list |
| `Progress(id string, horizontal bool)` | Creates a progress indicator |
| `Select(id string, args ...string)` | Creates a dropdown selector |
| `Spacer()` | Creates an invisible spacer |
| `Spinner(id string, sequence string)` | Creates a spinner animation |
| `Scanner(id string, width int, charStyle string)` | Creates a scanner animation |
| `Static(id, text string)` | Creates static text display |
| `Styled(id, text string)` | Creates styled text display |
| `Switcher(id string, connect bool)` | Creates a content switcher |
| `Tab(name string)` | Adds a tab (requires Tabs/Switcher) |
| `Table(id string, provider TableProvider)` | Creates a table widget |
| `Tabs(id string, names ...string)` | Creates a tab navigation widget |
| `Text(id string, content []string, follow bool, max int)` | Creates a scrollable text area |
| `Viewport(id, title string)` | Creates a scrollable viewport |
| `VRule(style string)` | Creates a vertical rule |

**Styling (all return `*Builder`):**

| Method | Description |
|--------|-------------|
| `Background(params ...string)` | Sets background color |
| `Border(params ...string)` | Sets border style |
| `Bounds(x, y, w, h int)` | Sets absolute position/size |
| `Cell(x, y, w, h int)` | Sets grid cell for next widget |
| `Class(class string)` | Sets CSS-like class |
| `Columns(columns ...int)` | Sets grid column sizes |
| `Font(params ...string)` | Sets font attributes |
| `Foreground(params ...string)` | Sets text color |
| `Flag(flag string, value bool)` | Sets widget flag |
| `Hint(width, height int)` | Sets preferred size |
| `Margin(a ...int)` | Sets margin (1-4 values) |
| `Padding(a ...int)` | Sets padding (1-4 values) |
| `Position(x, y int)` | Sets absolute position |
| `Rows(rows ...int)` | Sets grid row sizes |
| `Size(width, height int)` | Sets absolute size |

## Widgets

### Button

Clickable button with text label.

**Constructor:** `NewButton(id, text string) *Button`

| Method | Description |
|--------|-------------|
| `Click()` | Programmatically triggers click event |

**Events:** `"click"` (when activated via Enter, Space, or mouse)

**Features:**
- Keyboard: `Enter` or `Space` activates
- Mouse: Click with bounds checking
- States: `"pressed"`, `"focused"`

---

### Checkbox

Toggleable boolean input with label.

**Constructor:** `NewCheckbox(id, text string, checked bool) *Checkbox`

| Method | Description |
|--------|-------------|
| `Toggle()` | Switches checked state |

**Flags:** `"checked"`, `"readonly"`, `"focusable"`

---

### Dialog

Container with optional title and border.

**Constructor:** `NewDialog(id, title string) *Dialog`

**Container methods:**

| Method | Description |
|--------|-------------|
| `Add(widget Widget)` | Sets content widget |
| `Children() []Widget` | Returns content widget |
| `Layout()` | Positions content within dialog |

**No specific events.** Emits standard container events.

---

### Digits

Large-format numeric display using ASCII art digits.

**Constructor:** `NewDigits(id, text string) *Digits`

No public methods or specific events.

---

### Editor

Multi-line text editing widget.

**Constructor:** `NewEditor(id string) *Editor`

See Component methods. Designed for larger text buffers.

---

### Flex

Linear layout container (horizontal or vertical).

**Constructor:** `NewFlex(id string, horizontal bool, alignment string, spacing int) *Flex`

**Container methods:**

| Method | Description |
|--------|-------------|
| `Add(widget Widget)` | Appends child |
| `Children() []Widget` | Returns all children |
| `Hint() (w, h int)` | Calculates from children |
| `Layout()` | Arranges children |

**No specific events.**

**Alignment values:** `"start"`, `"center"`, `"end"`, `"stretch"`

---

### Form

Data-bound form container connected to a Go struct.

**Constructor:** `NewForm(id, title string, data any) *Form`

**Container methods:**

| Method | Description |
|--------|-------------|
| `Add(widget Widget)` | Sets single child container |
| `Children() []Widget` | Returns child (usually FormGroup) |

**Public methods:**
None (use reflection with `Group()` builder method)

**Data binding:** Uses struct tags: `group`, `label`, `control`, `options`, `width`, `line`, `readonly`

---

### FormGroup

Arranges labeled form controls within a Form.

**Constructor:** `NewFormGroup(id, title string, horizontal bool, spacing int) *FormGroup`

**Container methods:**

| Method | Description |
|--------|-------------|
| `Add(line int, label string, widget Widget)` | Adds field at grid position |
| `Children() []Widget` | Returns all form control widgets |
| `Layout()` | Arranges fields |

**No specific events.**

---

### Grid

Table-based layout container with cell spanning.

**Constructor:** `NewGrid(id string, rows, columns int, lines bool) *Grid`

**Container methods:**

| Method | Description |
|--------|-------------|
| `Add(x, y, w, h int, widget Widget)` | Places widget in cell with span |
| `Children() []Widget` | Returns all cell contents |
| `Columns(columns ...int)` | Sets column sizes |
| `Rows(rows ...int)` | Sets row sizes |
| `Layout()` | Calculates cell positions and sizes |
| `Hint() (w, h int)` | Returns grid dimensions |

**Grid constants:**
- `GridH` (1) - Horizontal separator
- `GridV` (2) - Vertical separator
- `GridB` (3) - Both separators

**Sizing:** Positive = fixed, negative = fractional

---

### Grow

Animation wrapper that grows content from 0 to full size.

**Constructor:** `NewGrow(horizontal bool) *Grow`

**Container methods:**

| Method | Description |
|--------|-------------|
| `Add(widget Widget)` | Sets child widget |
| `Children() []Widget` | Returns child |
| `Layout()` | Calculates end size and child bounds |
| `Render(r *Renderer)` | Clips rendering during animation |

**Animation methods:**

| Method | Description |
|--------|-------------|
| `Start(interval time.Duration)` | Begins grow animation |
| `Stop()` | Stops animation |
| `Running() bool` | Checks if animating |

**No events.** Uses internal `tick()` to increment `step` until reaching `end`.

---

### Input

Single-line text input field.

**Constructor:** `NewInput(id string, params ...string) *Input`

| Method | Description |
|--------|-------------|
| `Insert(ch string)` | Inserts character at cursor |
| `Delete()` | Backspace (deletes before cursor) |
| `DeleteForward()` | Delete (removes at cursor) |
| `Clear()` | Removes all text |
| `Left()` | Moves cursor left |
| `Right()` | Moves cursor right |
| `Start()` | Moves to beginning (Home) |
| `End()` | Moves to end (End) |
| `SetText(text string)` | Sets entire content |
| `Text() string` | Returns current text |
| `SetMask(mask string)` | Enables password masking |

**Events:** `"change"` (data: `string` - full text), `"enter"` (data: `string`)

**Flags:** `"focusable"`, `"masked"`, `"readonly"`

---

### List

Scrollable selectable list with optional multi-selection.

**Constructor:** `NewList(id string, items []string) *List`

| Method | Description |
|--------|-------------|
| `Items() []string` | Returns all items |
| `SetItems(items []string)` | Replaces all items |
| `Select(index int)` | Sets highlighted index (fires `"select"`) |
| `Selected() int` | Returns highlighted index |
| `Move(count int)` | Moves highlight by count (skips disabled) |
| `First()` | Jumps to first enabled item |
| `Last()` | Jumps to last enabled item |
| `PageUp()` | Moves up by viewport height |
| `PageDown()` | Moves down by viewport height |

**Events:** `"select"` (data: `int` - index), `"activate"` (data: `int` - index)

**Flags:** `"focusable"`

**Options:**
- Show line numbers: `list.numbers = true`
- Show scrollbar: `list.scrollbar = true` (default)

---

### Progress

Visual progress indicator (determinate or indeterminate).

**Constructor:** `NewProgress(id string, horizontal bool) *Progress`

| Method | Description |
|--------|-------------|
| `SetValue(value int)` | Sets current progress (clamped to 0..total) |
| `SetTotal(total int)` | Sets total work units (0 = indeterminate) |
| `Value() int` | Returns current value |
| `Total() int` | Returns total |

**No events.**

**Indeterminate mode:** total=0, displays spinning animation.

---

### Rule

Horizontal or vertical line separator.

**Constructors:** `NewHRule(style string)`, `NewVRule(style string) *Rule`

Uses border style from theme.

---

### Scanner

Back-and-forth scanning animation with fading trail.

**Constructor:** `NewScanner(id string, width int, charStyle string) *Scanner`

| Method | Description |
|--------|-------------|
| `SetSequence(sequence string)` | Updates animation frames |
| `Current() string` | Returns current frame character |

**Animation control:** Inherited from Animation (Start, Stop)

**Styles:** `"blocks"`, `"diamonds"`, `"circles"`

---

### Select

Dropdown selection widget.

**Constructor:** `NewSelect(id string, args ...string) *Select`

**Arguments:** Alternating `value1, text1, value2, text2, ...`

| Method | Description |
|--------|-------------|
| `Select(value string)` | Sets selected by value |
| `Text() string` | Returns displayed text |
| `Value() string` | Returns selected value |

**Events:** `"change"` (data: `string` - selected value)

**Flags:** `"focusable"`

**Popup:** Shows List popup on `Enter` key press; `Escape` closes.

---

### Spinner

Animated loading indicator cycling through character sequence.

**Constructor:** `NewSpinner(id string, sequence string) *Spinner`

| Method | Description |
|--------|-------------|
| `Tick()` | Advances to next frame (called by animation) |
| `SetSequence(sequence string)` | Updates character sequence |
| `Current() string` | Returns current character |

**Animation control:** Inherited from Animation (Start, Stop)

**Predefined sequences:** `Spinners` map contains `"bar"`, `"dots"`, `"dot"`, `"arrow"`, `"circle"`, `"bounce"`, `"braille"`

---

### Styled

Rich text widget with inline styling support.

**Constructor:** `NewStyled(id string, text string) *Styled`

No public methods or specific events.

Uses markup for styling (see documentation).

---

### Static

Simple text display widget.

**Constructor:** `NewStatic(id, text string) *Static`

| Method | Description |
|--------|-------------|
| `SetText(text string)` | Updates displayed text |

**No events.**

---

### Switcher

Content switcher showing one child at a time.

**Constructor:** `NewSwitcher(id string) *Switcher`

| Method | Description |
|--------|-------------|
| `Add(widget Widget)` | Adds content pane |
| `Select(index int)` | Shows pane at index |
| `Selected() int` | Returns current pane index |
| `Len() int` | Number of panes |

**No specific events.** Can auto-connect to Tabs via Builder.

---

### Table

Tabular data display with scrolling and grid lines.

**Constructor:** `NewTable(id string, provider TableProvider) *Table`

| Method | Description |
|--------|-------------|
| `Set(provider TableProvider)` | Updates data source |
| `Row() int` | Returns highlighted row index |
| `Column() int` | Returns highlighted column index |

**Events:** `"select"` (data: `int` - row index)

**Flags:** `"focusable"`

**TableProvider interface:**
```go
type TableProvider interface {
    Columns() []TableColumn  // Column definitions
    Length() int             // Number of rows
    Item(row, col int) string // Cell value
}
```

---

### Tabs

Tab navigation widget.

**Constructor:** `NewTabs(id string) *Tabs`

| Method | Description |
|--------|-------------|
| `Add(title string)` | Adds new tab |
| `Select(index int)` | Sets active tab |
| `Selected() int` | Returns active tab index |
| `Len() int` | Number of tabs |

**Events:** `"change"` (navigation), `"activate"` (Enter key, data: `int` index)

**Flags:** `"focusable"`

---

### Text

Multi-line scrollable text display.

**Constructor:** `NewText(id string, content []string, follow bool, max int) *Text`

| Method | Description |
|--------|-------------|
| `Add(lines ...string)` | Appends lines (auto-rotates if max > 0) |
| `Clear()` | Removes all content |
| `Set(content []string)` | Replaces all content |

**No specific events.**

**Options:**
- `follow=true` - Auto-scrolls to newest content
- `max=N` - Limits lines to N (rotates)

---

### Viewport

Scrollable container for oversized content.

**Constructor:** `NewViewport(id, title string) *Viewport`

**Container methods:**

| Method | Description |
|--------|-------------|
| `Add(widget Widget)` | Sets single scrollable child |
| `Children() []Widget` | Returns child |
| `Layout()` | Positions child with scroll offsets |

**Public methods:** Scroll via arrow keys, `PgUp`, `PgDn`, `Home`, `End`

**No events.**

---

### Canvas

Low-level pixel buffer for custom rendering.

**Constructor:** `NewCanvas(id string, width, height int) *Canvas`

| Method | Description |
|--------|-------------|
| `CellAt(x, y int) *Cell` | Returns cell at coordinates |
| `SetCell(x, y int, ch string, style *Style)` | Sets cell content |
| `Clear()` | Clears all cells |
| `Fill(ch string, style *Style)` | Fills entire buffer |
| `Cursor() (x, y int, style string)` | Returns cursor position |

**Modes:** `ModeNormal`, `ModeInsert`

**Flags:** `"focusable"`

---

### Custom

User-defined render function widget.

**Constructor:** `NewCustom(id string, fn func(Widget, *Renderer)) *Custom`

Call `Component.Render(renderer)` inside `fn` to draw border/background.

---

## Helper Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `FindUI` | `(widget Widget) *UI` | Finds root UI in hierarchy |
| `ID` | `(widget Widget) string` | Returns widget ID or `"<nil>"` |
| `OnKey` | `(widget Widget, handler func(Widget, *tcell.EventKey) bool)` | Registers key handler |
| `OnMouse` | `(widget Widget, handler func(Widget, *tcell.EventMouse) bool)` | Registers mouse handler |
| `Redraw` | `(widget Widget)` | Queues widget for redraw |
| `WidgetType` | `(widget Widget) string` | Returns type name without package |
| `HandleKeyEvent` | `(container Container, id string, fn func(Widget, *tcell.EventKey) bool)` | Registers key handler by ID |
| `HandleListEvent` | `(container Container, id, event string, fn func(*List, string, int) bool)` | Registers list handler by ID |
| `Update` | `(container Container, id string, value any)` | Updates widget content by type |

## Table Providers

**Prebuilt providers:**

| Function | Description |
|----------|-------------|
| `NewArrayTableProvider(headers []string, data [][]string)` | Creates provider from 2D array |
| `NewFuncTableProvider(columns []TableColumn, length func() int, item func(row, col int) string)` | Creates provider from functions |

**TableColumn struct:**

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Column header name |
| `Width` | `int` | Column width in characters |

## Styles

Style selectors: `""` (default), `":focus"`, `":hover"`, `":disabled"`, `part:state` (e.g., `"highlight:focused"`)

**Theme keys (common):**

| Key | Description |
|-----|-------------|
| `border` | Border style name |
| `foreground` | Text color |
| `background` | Background color |
| `font` | Font attributes (`bold`, `italic`, `underline`, `strikethrough`) |
| `cursor` | Cursor style |

**Style methods:**

| Method | Returns | Description |
|--------|---------|-------------|
| `WithBorder(border string) *Style` | New style | Sets border style |
| `WithForeground(color string) *Style` | New style | Sets foreground color |
| `WithBackground(color string) *Style` | New style | Sets background color |
| `WithFont(font string) *Style` | New style | Sets font attributes |
| `WithMargin(a ...int) *Style` | New style | Sets margin |
| `WithPadding(a ...int) *Style` | New style | Sets padding |
| `Fixed() bool` | `bool` | Returns true if values are explicit |
| `Border() string` | `string` | Gets border style |
| `Foreground() string` | `string` | Gets foreground color |
| `Background() string` | `string` | Gets background color |
| `Font() string` | `string` | Gets font attributes |
| `Margin() (top, right, bottom, left int)` | `(int, int, int, int)` | Gets margin values |
| `Padding() (top, right, bottom, left int)` | `(int, int, int, int)` | Gets padding values |
| `Horizontal() int` | `int` | Total horizontal margin+padding |
| `Vertical() int` | `int` | Total vertical margin+padding |
| `Cursor() string` | `string` | Gets cursor style |

## Event Types

**Common events across widgets:**

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | varies | Content modified |
| `"select"` | `int` | Item selected (List, Table) |
| `"activate"` | `int` | Item activated (Enter key) |
| `"key"` | `*tcell.EventKey` | Keyboard event |
| `"mouse"` | `*tcell.EventMouse` | Mouse event |

Event handlers return `bool` - `true` stops propagation.

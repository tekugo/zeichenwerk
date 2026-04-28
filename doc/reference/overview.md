# Zeichenwerk API Overview

A TUI component library based on tcell/v3.

## Widgets

### Containers
- [Box](box.md) — bordered box with optional title
- [Card](card.md) — bordered container with title in border, optional footer
- [Collapsible](collapsible.md) — toggleable header that hides/reveals one child
- [CRT](crt.md) — Matrix-style power-on/off animation wrapper
- [Dialog](dialog.md) — single-child container used as popup layer
- [Flex](flex.md) — linear layout (horizontal or vertical)
- [Form](form.md) — data-bound form connected to a Go struct
- [FormGroup](form-group.md) — labeled form controls within a Form
- [Grid](grid.md) — table-based layout with cell spanning
- [Grow](grow.md) — animated reveal wrapper
- [Switcher](switcher.md) — shows one child pane at a time
- [Viewport](viewport.md) — scrollable container for oversized content

### Input
- [Button](button.md) — clickable button
- [Checkbox](checkbox.md) — toggleable boolean input
- [Combo](combo.md) — text input with suggestion-list popup
- [Editor](editor.md) — multi-line text editor
- [Filter](filter.md) — search input bound to a Filterable widget
- [Input](input.md) — single-line text field
- [List](list.md) — scrollable selectable list
- [Select](select.md) — dropdown selection
- [Tree](tree.md) — expandable hierarchy of nodes
- [TreeFS](tree-fs.md) — Tree pre-wired for filesystem navigation
- [Typeahead](typeahead.md) — input with ghost-text suggestion completion

### Display
- [BarChart](bar-chart.md) — multi-series stacked bar chart
- [Breadcrumb](breadcrumb.md) — path-style segment indicator
- [Canvas](canvas.md) — low-level pixel buffer for custom rendering
- [Deck](deck.md) — fixed-height list of items rendered by a callback
- [Digits](digits.md) — large ASCII art character display
- [Heatmap](heatmap.md) — coloured cell grid for matrix data
- [Rule](rule.md) — horizontal or vertical line separator
- [Shortcuts](shortcuts.md) — single-row keyboard hint bar
- [Sparkline](sparkline.md) — inline trend chart
- [Static](static.md) — plain text display
- [Styled](styled.md) — rich text with inline markup
- [Table](table.md) — tabular data display
- [Tabs](tabs.md) — tab navigation
- [Terminal](terminal.md) — embedded terminal emulator
- [Text](text.md) — multi-line scrollable text
- [Tiles](tiles.md) — wrapping grid of fixed-size tiles

### Animated
- [Clock](clock.md) — live wall-clock display
- [Marquee](marquee.md) — scrolling text ticker
- [Progress](progress.md) — progress bar (determinate or indeterminate)
- [Scanner](scanner.md) — back-and-forth scanning animation
- [Shimmer](shimmer.md) — text with sweeping highlight band
- [Spinner](spinner.md) — animated loading indicator
- [Typewriter](typewriter.md) — character-by-character text reveal

### Custom / Extension
- [Custom](custom.md) — widget with a user-supplied render function

## Widget Interface

The fundamental interface for all UI components. All widgets share common functionality through `Component`, which provides default implementations.

- `Apply(theme *Theme)` — applies theme styles
- `Bounds() (x, y, w, h int)` — absolute screen coordinates
- `Content() (x, y, w, h int)` — inner content area coordinates
- `Cursor() (x, y int, style string)` — cursor position relative to content
- `Dispatch(source Widget, event string, data ...any) bool` — dispatches event to handlers
- `Flag(name string) bool` — gets boolean state flag
- `Hint() (w, h int)` — preferred content size
- `ID() string` — unique identifier
- `Info() string` — human-readable description
- `Log(source Widget, level, msg string, params ...any)` — logs debug message
- `On(event string, handler Handler)` — registers event handler
- `Parent() Container` — parent container
- `Refresh()` — queues widget redraw
- `Render(r *Renderer)` — renders widget to screen
- `SetBounds(x, y, w, h int)` — sets absolute position and size
- `SetFlag(name string, value bool)` — sets boolean state flag
- `SetHint(w, h int)` — sets preferred content size
- `SetParent(parent Container)` — sets parent container
- `SetStyle(selector string, style *Style)` — applies style for selector
- `State() string` — widget state for rendering
- `Style(selector ...string) *Style` — returns style for selector

## Container Interface

Extends `Widget` with child management.

- `Children() []Widget` — all direct child widgets
- `Layout()` — arranges child widgets

**Helper functions:**
- `Find(container, id string) Widget` — finds widget by ID (depth-first)
- `FindAll[T any](container) []T` — finds all widgets of type T
- `FindAt(container, x, y int) Widget` — finds widget at coordinates
- `Layout(container)` — recursively lays out all child containers
- `Traverse(container, func(Widget) bool)` — depth-first traversal

## Component

Embedded struct providing default `Widget` implementation. Embed in every custom widget.

**Constructor:** `NewComponent(id, class string) *Component`

Override in embedding structs:
- `Render(r *Renderer)` — draws margin, border, background; call `c.Component.Render(r)` first
- `Apply(theme *Theme)` — call `theme.Apply(w, w.Selector("mywidget"))`
- `Cursor() (int, int, string)` — returns `(0, 0, "")` by default

Additional public methods:
- `Class() string` — style class
- `Selector(t string) string` — builds a `type.class#id` selector string
- `Styles() []string` — all defined style selectors

## Animation

Embedded struct for timed animations. Manages ticker and goroutine.

- `Refresh()` — triggers widget redraw
- `Running() bool` — true if animation is active
- `Start(interval time.Duration)` — starts animation goroutine
- `Stop()` — stops animation gracefully
- `Tick()` — called on each frame (override in embedding struct)

## UI

Root application class managing screen, events, and rendering.

**Constructor:** `NewUI(theme *Theme, root Container, debug bool) (*UI, error)`

- `Close()` — removes topmost layer
- `Draw()` — renders entire UI
- `DrawWidget(widget Widget)` — renders single widget
- `EventLoop()` — polls tcell events (run as goroutine)
- `Focus(widget Widget)` — sets keyboard focus
- `Handle(event tcell.Event) bool` — processes tcell events
- `Layout()` — recalculates layout for all layers
- `Log(source Widget, levelStr, msg string, params ...any)` — adds structured log entry
- `Logs() *TableLog` — returns table log widget
- `NewBuilder() *Builder` — creates builder with current theme
- `Popup(x, y, w, h int, popup Container)` — shows container as overlay
- `Redraw(widget Widget)` — queues widget for individual redraw
- `Refresh()` — queues full screen redraw
- `Run() error` — starts main event loop (blocks)
- `SetFocus(which string)` — navigates focus: `"first"`, `"last"`, `"next"`, `"previous"`
- `SetLogLevel(level slog.Level)` — changes log level at runtime
- `SetTheme(theme *Theme)` — changes active theme
- `ShowCursor()` — positions and shows cursor
- `ShowDebug()` — renders debug info bar
- `Theme() *Theme` — current theme

**Keyboard shortcuts:**

| Keys | Action |
|------|--------|
| `Tab`, `Right`, `Down` | Next focusable widget |
| `Backtab`, `Left`, `Up` | Previous focusable widget |
| `Escape` | Close topmost popup |
| `Ctrl+C`, `Ctrl+Q`, `q`, `Q` | Quit application |
| `Ctrl+D` | Open inspector popup (debug mode) |

## Builder

Fluent API for constructing UIs.

**Constructor:** `NewBuilder(theme *Theme) *Builder`

**Control:**
- `Build() *UI` — returns UI instance
- `Run()` — builds and runs (blocks)

**Navigation:**
- `Container() Container` — returns top-level container
- `End() *Builder` — pops current container from stack
- `Find(id string) Widget` — finds widget by ID
- `With(fn func(*Builder)) *Builder` — inline composition helper

**Widget methods** (all return `*Builder`):
- `Box(id, title string)`
- `Button(id, text string)`
- `Checkbox(id, text string, checked bool)`
- `Dialog(id, title string)`
- `Digits(id, text string)`
- `Editor(id string)`
- `Flex(id string, horizontal bool, alignment string, spacing int)`
- `Form(id, title string, data any)`
- `Grid(id string, rows, columns int, lines bool)`
- `Group(id, title, groupName string, horizontal bool, spacing int)`
- `HRule(style string)`
- `Input(id string, params ...string)`
- `List(id string, values ...string)`
- `Progress(id string, horizontal bool)`
- `Scanner(id string, width int, charStyle string)`
- `Select(id string, args ...string)`
- `Spacer()`
- `Spinner(id string, sequence string)`
- `Static(id, text string)`
- `Styled(id, text string)`
- `Switcher(id string, connect bool)`
- `Tab(name string)`
- `Table(id string, provider TableProvider)`
- `Tabs(id string, names ...string)`
- `Text(id string, content []string, follow bool, max int)`
- `Viewport(id, title string)`
- `VRule(style string)`

**Styling methods** (all return `*Builder`):
- `Background(params ...string)`
- `Border(params ...string)`
- `Bounds(x, y, w, h int)`
- `Cell(x, y, w, h int)` — grid cell placement
- `Class(class string)`
- `Columns(columns ...int)`
- `Flag(flag string, value bool)`
- `Font(params ...string)`
- `Foreground(params ...string)`
- `Hint(width, height int)`
- `Margin(a ...int)` — 1–4 values
- `Padding(a ...int)` — 1–4 values
- `Position(x, y int)`
- `Rows(rows ...int)`
- `Size(width, height int)`

## Helper Functions

- `FindUI(widget Widget) *UI` — traverses up hierarchy to find root UI
- `HandleKeyEvent(container Container, id string, fn func(Widget, *tcell.EventKey) bool)` — registers key handler by widget ID
- `HandleListEvent(container Container, id, event string, fn func(*List, string, int) bool)` — registers list handler by widget ID
- `ID(widget Widget) string` — returns widget ID or `"<nil>"`
- `OnActivate(widget Widget, handler func(Widget, int) bool)` — registers activate handler; receives item index
- `OnChange(widget Widget, handler func(Widget, string) bool)` — registers change handler; receives new value as string
- `OnKey(widget Widget, handler func(Widget, *tcell.EventKey) bool)` — registers key handler
- `OnMouse(widget Widget, handler func(Widget, *tcell.EventMouse) bool)` — registers mouse handler
- `OnSelect(widget Widget, handler func(Widget, int) bool)` — registers select handler; receives item index
- `Redraw(widget Widget)` — queues widget for redraw
- `Update(container Container, id string, value any)` — updates widget content by type
- `WidgetType(widget Widget) string` — returns type name without package prefix

## Table Providers

```go
type TableProvider interface {
    Columns() []TableColumn
    Length() int
    Str(row, col int) string
}
```

**`TableColumn` fields:** `Name string`, `Width int`

**Built-in:** `NewArrayTableProvider(headers []string, data [][]string)`

## Styles

Selectors: `""` (default), `":focus"`, `":hover"`, `":disabled"`, `"part"`, `"part:state"`

Fallback order for `"part:state"`: exact → part → `:state` → default.

**Style methods:**
- `Background() string`
- `Border() string`
- `Cursor() string`
- `Fixed() bool` — true if values are explicit (not inherited)
- `Font() string`
- `Foreground() string`
- `Horizontal() int` — total horizontal margin + padding
- `Margin() (top, right, bottom, left int)`
- `Padding() (top, right, bottom, left int)`
- `Vertical() int` — total vertical margin + padding
- `WithBackground(color string) *Style`
- `WithBorder(border string) *Style`
- `WithFont(font string) *Style`
- `WithForeground(color string) *Style`
- `WithMargin(a ...int) *Style`
- `WithPadding(a ...int) *Style`

## Events

Event handlers have type `func(source Widget, event string, data ...any) bool`. Returning `true` stops propagation. Multiple handlers for the same event are called in reverse registration order (newest first).

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `int` | Item activated via Enter (List, Table, Tabs) |
| `"change"` | varies | Content or state modified |
| `"click"` | — | Button activated |
| `"hide"` | — | Switcher pane hidden |
| `"key"` | `*tcell.EventKey` | Keyboard event |
| `"mode"` | `string` | Canvas mode changed |
| `"mouse"` | `*tcell.EventMouse` | Mouse event |
| `"move"` | `x, y int` | Canvas cursor moved |
| `"select"` | `int` | Item highlighted (List, Table) |
| `"show"` | — | Switcher pane shown |

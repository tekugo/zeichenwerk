# Guidelines

## Normative Keywords

- **MUST**: mandatory requirement
- **SHOULD**: recommended but optional with justification
- **MAY**: optional

## Core Principles

- MUST follow idiomatic Go patterns
- MUST minimize dependencies
- MUST use explicit error handling
- SHOULD keep functions <= 50 lines
- SHOULD prefer composition over inheritance patterns

## File Formats

- MUST use Markdown for documentation files

## Libraries

- MUST use log/slog for logging

## Reading order for new contributors

- **Tutorial**: [`doc/tutorial/README.md`](doc/tutorial/README.md) — chapters from a 20-line teaser through a real SQLite query tool. Every snippet has a runnable program under `doc/tutorial/examples/`.
- **Reference**: [`doc/reference/overview.md`](doc/reference/overview.md) — one page per widget.
- **Design principles**: [`doc/principles.md`](doc/principles.md) — the small set of invariants the framework relies on.
- **Events / flags**: [`doc/events.md`](doc/events.md), [`doc/flags.md`](doc/flags.md).

## Project Overview

This is a TUI component library based on `tcell/v3`. Key types:

- **`Widget`** (`widget.go`) — interface for all UI elements; application code works against this type
- **`Component`** (`component.go`) — base struct that satisfies `Widget`; embed in every custom widget
- **`Container`** (`container.go`) — extends `Widget`; manages children and layout
- **`Style`** (`style.go`) — CSS-like styling (colors, borders, margins, padding); hierarchical/inheritable
- **`Theme`** (`theme.go`) — registry of styles indexed by CSS-like selectors (`type.class#id:state`)
- **`Renderer`** (`renderer.go`) — drawing abstraction over tcell; widgets MUST NOT import tcell directly
- **`UI`** (`ui.go`) — root container, event loop, focus management, render pipeline; takes the whole screen
- **`Builder`** (`builder.go`) — fluent API to construct and style a widget tree; preferred way to build UIs
- **`Animation`** (`animation.go`) — embed in widgets that need timed animation; manages goroutine + ticker

### Architecture rules

- All widgets embed `Component` and implement `Render(*Renderer)` and `Apply(*Theme)`
- Every new widget MUST be registered in `Builder` (add a method) and `compose/compose.go` (add a function)
- Containers call `child.SetBounds` and `child.SetParent` during `Layout()`; never touch tcell directly
- Rendering is top-down: `UI.Draw` → each layer → each container → each child
- Events bubble from focused widget up through parents; return `true` from a handler to stop propagation
- `Redraw(widget)` queues a single-widget redraw; call after any state change that affects rendering

### Selector format

`type/part.class#id:state` — all parts optional.
Examples: `"button"`, `"button.primary"`, `"button#submit:focused"`, `"input/placeholder"`.

## Project Structure

```
zeichenwerk/
├── *.go              # Core library (widgets, theme, renderer, …)
├── cmd/
│   ├── compose/      # Compose API demo
│   ├── demo/         # Builder API demo (all widgets)
│   └── showcase/     # Full interactive showcase
├── compose/          # Functional composition API (Option-based)
├── doc/              # Design proposals and reference docs
├── spec/             # Widget specifications (written before implementation)
└── .claude/
    └── skills/
        └── zeichenwerk/
            ├── SKILL.md    # Claude Code skill — auto-loaded in this project
            └── widgets.md  # Full widget reference for the skill
```

## Widget Reference

For the full style-key, event, and method reference per widget see
`.claude/skills/zeichenwerk/widgets.md`.

### Containers

| Widget        | File             | Constructor                                                    | Key methods                                                                                |
| ------------- | ---------------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| `Flex`        | `flex.go`        | `NewFlex(id, class, alignment Alignment, spacing int)`         | `Add(Widget)` — set `FlagVertical` for vertical layout                                     |
| `Grid`        | `grid.go`        | `NewGrid(id, class, rows, cols int, lines bool)`               | `Add(content Widget, params ...any)` — params are `x, y, w, h`; `Rows(…int)`, `Columns(…)` |
| `Box`         | `box.go`         | `NewBox(id, class, title)`                                     | `Add(Widget)` — single child                                                               |
| `Card`        | `card.go`        | `NewCard(id, class, title)`                                    | `Add(Widget)` — first call = content, second = footer                                      |
| `Switcher`    | `switcher.go`    | `NewSwitcher(id, class)`                                       | `Add(Widget)`, `Select(any)` — dispatches `EvtShow`/`EvtHide` to panes when `connect`      |
| `Viewport`    | `viewport.go`    | `NewViewport(id, class, title)`                                | `Add(Widget)` — scrollable single child                                                    |
| `Form`        | `form.go`        | `NewForm(id, class, title, data any)`                          | binds a Go struct via reflection                                                           |
| `FormGroup`   | `form-group.go`  | `NewFormGroup(id, class, title, horizontal bool, spacing int)` | `Add(Widget, params ...any)` — params are `line int, label string`                         |
| `Collapsible` | `collapsible.go` | `NewCollapsible(id, class, title, expanded bool)`              | `Add(Widget)`, `Expand()`, `Collapse()`, `Toggle()`                                        |
| `Dialog`      | `dialog.go`      | `NewDialog(id, class, title)`                                  | `Add(Widget)` — single child; shown as popup via `UI.Popup`                                |
| `Grow`        | `grow.go`        | `NewGrow(id, class, horizontal bool)`                          | `Add(Widget)`, `Start(interval)` — animated reveal                                         |
| `CRT`         | `crt.go`         | `NewCRT(id, class)`                                            | `Add(Widget)`, `Start(interval)`, `PowerOff(interval, onDone)` — Matrix-style boot anim    |

**Flex alignment values** (`core.Alignment`): `Start`, `Left`, `Center`, `Right`, `End`, `Stretch`, `Default`. Builder methods `HFlex` / `VFlex` create a Flex with the orientation flag pre-set.

**Grid sizing:** positive = fixed chars; negative = fractional unit; zero = auto (preferred size).
At least one row size and one column size MUST be negative for the Grid to fill available space.

### Input widgets

| Widget      | File           | Constructor                                          | Events                                  | Key public methods                                                              |
| ----------- | -------------- | ---------------------------------------------------- | --------------------------------------- | ------------------------------------------------------------------------------- |
| `Button`    | `button.go`    | `NewButton(id, class, text)`                         | `EvtActivate(int)`                      | `Activate()`, `Text() string`, `Set(string)`                                    |
| `Checkbox`  | `checkbox.go`  | `NewCheckbox(id, class, text, checked bool)`         | `EvtChange(bool)`                       | `Checked() bool`, `Toggle()`, `Set(any)`                                        |
| `Input`     | `input.go`     | `NewInput(id, class, params…)`                       | `EvtChange(string)` `EvtEnter(string)`  | `Get() string`, `Set(string)`, `SetMask(string)`, `Insert/Delete/Clear`         |
| `Typeahead` | `typeahead.go` | `NewTypeahead(id, class, params…)`                   | `EvtChange(string)` `EvtAccept(string)` | All `Input` methods + `SetSuggest(func(string) []string)`                       |
| `Combo`     | `combo.go`     | `NewCombo(id, class, items []string)`                | `EvtChange(string)` `EvtActivate(string)` | `Get() string`, `Set(string)`, `SetItems([]string)` — opens popup on focus    |
| `Filter`    | `filter.go`    | `NewFilter(id, class)`                               | inherits `Typeahead`                    | `Bind(Filterable)`, `Unbind()`, `Bound() Filterable` — drives a target widget   |
| `Select`    | `select.go`    | `NewSelect(id, class, val, label, …)`                | `EvtChange(string)`                     | `Select(string)`, `Value() string`                                              |
| `List`      | `list.go`      | `NewList(id, class, items []string)`                 | `EvtSelect(int)` `EvtActivate(int)`     | `Set([]string)`, `Items()`, `Select(int)`, `Selected() int`, `Filter(string)`   |
| `Deck`      | `deck.go`      | `NewDeck(id, class, render ItemRender, itemHeight)`  | `EvtSelect(int)` `EvtActivate(int)`     | `Get() []any`, `Set([]any)`, `Select(int)`, `Selected() int` — itemHeight ≥ 1   |
| `Tiles`     | `tiles.go`     | `NewTiles(id, class, render ItemRender, tw, th int)` | `EvtSelect(int)` `EvtActivate(int)`     | `Items()`, `SetItems([]any)`, `Move(dr, dc int)` — wrapping grid                |
| `Editor`    | `editor.go`    | `NewEditor(id, class)`                               | `EvtChange`                             | `Text()`, `Lines()`, `Load(string)`, `SetContent([]string)`, `ShowLineNumbers`  |
| `Tree`      | `tree.go`      | `NewTree(id, class)`                                 | `EvtSelect` `EvtActivate` `EvtChange`   | `Add(*TreeNode)`, `Selected() *TreeNode`, `Select(*TreeNode)`                   |
| `TreeFS`    | `tree-fs.go`   | `NewTreeFS(id, class, root string, dirsOnly bool)`   | inherits `Tree`                         | `RootPath()`, `SetRoot(string)`, `DirsOnly() bool`, `SetDirsOnly(bool)`         |

**`ItemRender`** signature: `func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool)`
`focused` is `true` when the Deck widget itself holds keyboard focus.

### Display widgets

| Widget       | File           | Constructor                                                | Key public methods                                                                                                                                                                                                                                                                |
| ------------ | -------------- | ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Static`     | `static.go`    | `NewStatic(id, class, text)`                               | `Set(any)`, `SetAlignment("left"\|"center"\|"right")`                                                                                                                                                                                                                              |
| `Text`       | `text.go`      | `NewText(id, class, lines []string, follow bool, max int)` | `Add(lines …string)`, `Clear()`, `Set([]string)`                                                                                                                                                                                                                                   |
| `Styled`     | `styled.go`    | `NewStyled(id, class, text)`                               | markup: `**bold**`, `*italic*`, `__underline__`, `` `code` ``                                                                                                                                                                                                                      |
| `Digits`     | `digits.go`    | `NewDigits(id, class, text)`                               | `Get() string`, `Set(string)` — large ASCII-art numerals                                                                                                                                                                                                                           |
| `Tabs`       | `tabs.go`      | `NewTabs(id, class)`                                       | `Add(title string)`, `Set(int) bool`, `Get() int`, `Count() int`                                                                                                                                                                                                                   |
| `Table`      | `table.go`     | `NewTable(id, class, provider TableProvider, cellNav bool)`| `Set(TableProvider)` (caller must `Refresh()` after), `Selected/SetSelected/Offset/SetOffset` — events: `EvtSelect(int,int)`, `EvtActivate(int,[]string)`                                                                                                                          |
| `Canvas`     | `canvas.go`    | `NewCanvas(id, class, pages, width, height int)`           | `SetCell(x,y,ch,style)`, `Clear()`, `SetMode(string)` — modes: `NORMAL`, `INSERT`, …                                                                                                                                                                                               |
| `Rule`       | `rule.go`      | `NewHRule(class, style)` / `NewVRule(class, style)`        | id is hard-coded to `hrule`/`vrule`; `style` is a theme border name                                                                                                                                                                                                                |
| `Breadcrumb` | `breadcrumb.go`| `NewBreadcrumb(id, class)`                                 | `Get/Set/Push/Pop/Truncate`, `Select(int)`, `Selected() int` — events: `EvtSelect(int)`, `EvtActivate(int)`                                                                                                                                                                        |
| `Shortcuts`  | `shortcuts.go` | `NewShortcuts(id, class, pairs ...string)`                 | `SetPairs(...string)` — alternating key/label                                                                                                                                                                                                                                      |
| `Terminal`   | `terminal.go`  | `NewTerminal(id, class)`                                   | `Write([]byte) (int, error)`, `Clear()`, `Resize(w, h int)`, `Title() string`                                                                                                                                                                                                      |
| `BarChart`   | `bar-chart.go` | `NewBarChart(id, class)`                                   | `SetSeries`, `AddSeries`, `SetCategories`, `SetHorizontal/SetShowAxis/SetShowGrid/SetShowValues/SetLegend`, `Select(int)` — events: `EvtSelect(int)`, `EvtActivate(int)`                                                                                                           |
| `Heatmap`    | `heatmap.go`   | `NewHeatmap(id, class, rows, cols int)`                    | `SetValue(row, col, v)`, `SetRow(row, vs)`, `SetAll(vs)`, `SetRowLabels`, `SetColLabels`, `SetCellWidth(int)` — colour-graded grid; interpolates bg between `"heatmap/zero"` and `"heatmap/max"`                                                                                  |
| `Sparkline`  | `sparkline.go` | `NewSparkline(id, class)`                                  | `Append(float64)`, `SetValues([]float64)`, `SetMode(ScaleMode)`, `SetMin/Max`, `SetThreshold(float64)`, `SetGradient(bool)`, `SetCapacity(int)` — height-adaptive: `h` rows = `h×8` levels; gradient mode blends base→high colour across `[threshold, max]`                       |

**`Terminal`** implements `io.Writer` — pipe pty/subprocess output directly into it.
`Write` is goroutine-safe. The buffer auto-resizes to the widget's content area on first render.
Handles full VT100/ANSI: cursor, erase, SGR (16/256/true-colour, underline styles), alternate screen,
scroll regions, OSC titles. See `.claude/skills/zeichenwerk/widgets.md` § Terminal for details.

**`TreeNode`**: `NewTreeNode(text string, data ...any)` or `NewLazyTreeNode(text string, loader NodeLoader, data ...any)`. Methods: `Add(*TreeNode)`, `SetLoader(NodeLoader)`, `Expand()`, `Collapse()`, `Toggle()`, `SetDisabled(bool)`. `TreeFS` wraps `Tree` for filesystem browsing.

**`TableProvider` interface**: `Columns() []TableColumn`, `Length() int`, `Str(row, col int) string`.
`NewArrayTableProvider(headers []string, rows [][]string)` — in-memory implementation.

### Animated widgets

All embed `Animation`; call `Start(interval)` after construction and `Stop()` to halt. None of them dispatch events except where noted.

| Widget       | File            | Constructor                                | Notes                                                                            |
| ------------ | --------------- | ------------------------------------------ | -------------------------------------------------------------------------------- |
| `Clock`      | `clock.go`      | `NewClock(id, class, interval, params...)` | `params[0]` = Go time-layout; `params[1]` = prefix; `Start()` (no interval arg)  |
| `Marquee`    | `marquee.go`    | `NewMarquee(id, class)`                    | `SetText`, `SetSpeed`, `SetGap`; pauses while hovered                            |
| `Progress`   | `progress.go`   | `NewProgress(id, class, horizontal bool)`  | `SetValue(int)`, `SetTotal(int)` — `total=0` = indeterminate                    |
| `Scanner`    | `scanner.go`    | `NewScanner(id, class, width, charStyle)`  | `Start(interval)`, `Stop()`                                                      |
| `Shimmer`    | `shimmer.go`    | `NewShimmer(id, class)`                    | `SetText`, `SetBandWidth`, `SetEdgeWidth`, `SetGradient(bool)`                   |
| `Spinner`    | `spinner.go`    | `NewSpinner(id, class, sequence string)`   | `Current()`, `SetSequence(string)`; built-in: `Spinners["bar"\|"dots"\|…]`     |
| `Typewriter` | `typewriter.go` | `NewTypewriter(id, class)`                 | `SetText`, `SetRate`, `SetDwell`, `SetRepeat` — events: `EvtChange`, `EvtActivate` |

### Custom / extension

| Widget            | File           | Use when                                                                  |
| ----------------- | -------------- | ------------------------------------------------------------------------- |
| `Custom`          | `custom.go`    | Simple custom rendering; pass a `func(Widget, *Renderer)` to `NewCustom`  |
| embed `Component` | `component.go` | Full custom widget; implement `Render(*Renderer)` and `Apply(*Theme)`     |
| embed `Animation` | `animation.go` | Custom widget needs a timed render loop; embed and call `Start(interval)` |

### Utility functions

From `core` package (`core/helper.go`):

```go
core.Find(container, id) Widget                        // First match by ID, depth-first; nil if not found
core.MustFind[T Widget](container, id) T               // Typed lookup; panics on miss or wrong type
core.FindAll[T any](container) []T                     // All widgets of type T in subtree
core.FindAt(container, x, y) Widget                    // Deepest widget at screen coordinates
core.Traverse(container, func(Widget) bool)            // Depth-first walk; return false to skip subtree
core.Layout(container) error                           // Recursively layout container subtree
```

From `widgets` package:

```go
widgets.OnActivate(widget, func(int) bool)             // Typed wrappers — handler signature differs
widgets.OnSelect(widget, func(int) bool)               // per event; see widgets/event-helper.go for the
widgets.OnChange(widget, func(string) bool)            // full set (OnAccept, OnEnter, OnHide, OnShow, …).
widgets.OnKey(widget, func(*tcell.EventKey) bool)      //
widgets.OnMouse(widget, func(*tcell.EventMouse) bool)  //
widgets.Redraw(widget)                                 // Queue single-widget redraw (cheap repaint)
widgets.Relayout(widget)                               // Re-run layout starting at widget; redraw
widgets.FindRoot(widget) Root                          // Walk up to the *UI / root container
widgets.ID(widget) string                              // Widget ID or "<nil>"
widgets.WidgetType(widget) string                      // Type name without package prefix
widgets.Suggest(items []string) func(string) []string  // Prefix-match suggester for Typeahead
```

From `values` package:

```go
values.Update[T any](container, id, value T)           // Push value into widget by id (uses Setter[T])
values.NewValue[T any](initial T) *Value[T]            // Reactive value; supports Bind/Subscribe
```

### AI / Agent tooling (`dump.go`)

```go
Dump(w io.Writer, root Widget)   // Stream indented widget tree to any writer
DumpToStdout(root Widget)        // Convenience: Dump to os.Stdout
ui.Dump(w io.Writer)             // Dump all layers with UI header line
```

Run `go run ./cmd/demo -dump` (or `./cmd/showcase -dump`, `./cmd/compose -dump`)
to print the fully laid-out widget hierarchy of that application to stdout and
exit — no terminal required. Useful for giving an AI agent a complete snapshot
of the UI structure and current widget state.

The output is an indented text tree: one line per widget with its type, ID,
content summary (text labels, input values, tab names, etc.), screen bounds,
and state flags (`[FOCUSED]`, `[HIDDEN]`, `[DISABLED]`). Hidden widgets are
always included with their full subtree so all content is readable. Custom
widgets may implement the `Summarizer` interface (`Summary() string`) to
provide their own single-line description.

`-dump-verbose` adds a `style:` line below each widget showing the effective
border type, non-zero padding/margin (compact CSS notation), and foreground/
background color variables for the widget's current state.

Both flags lay out the UI at a fixed 120×40 terminal size so the output is
deterministic without a live terminal session.

## Building a new widget — checklist

1. Create `mywidget.go` with `type MyWidget struct { Component; ... }`
2. Constructor: `NewMyWidget(id, class string, ...) *MyWidget` — call `SetFlag(FlagFocusable, true)` if interactive
3. Implement `Render(r *Renderer)` — call `c.Component.Render(r)` first for borders/bg
4. Implement `Apply(theme *Theme)` — call `theme.Apply(w, w.Selector("mywidget"), states...)`
5. Register in `Builder` (`builder.go`): add `func (b *Builder) MyWidget(id string, ...) *Builder`
6. Register in `compose/compose.go`: add `func MyWidget(id, class string, options ...Option) Option`
7. Add theme style keys to all five theme files (`theme-*.go`)
8. Add `doc.go` entry and export all public symbols with comments

## Events

All event constants are defined in `widgets/events.go`. Handler signature:
`func(source Widget, event Event, data ...any) bool` — return `true` to stop propagation.

Events bubble from the focused widget up through its parents, stopping at the root container; `ui` itself is currently dispatched to as well, so global handlers can be attached there. Multiple handlers on the same widget run in **reverse registration order**, so the most recently added runs first and can consume the event before earlier handlers see it.

| Constant      | Typical payload     | Notes                                                                              |
| ------------- | ------------------- | ---------------------------------------------------------------------------------- |
| `EvtActivate` | `int` or `string`   | Item activated (Enter, double-click); int for List/Tabs/etc., string for Combo     |
| `EvtSelect`   | `int` or `int, int` | Highlighted item changed; `int` for List, `(row, col)` for Table when in cell-nav  |
| `EvtChange`   | varies              | Content/state changed: `string` for Input, `bool` for Checkbox                     |
| `EvtAccept`   | `string`            | Typeahead suggestion accepted (Tab)                                                |
| `EvtEnter`    | `string`            | Enter pressed in an Input                                                          |
| `EvtFocus`    | —                   | Widget gained focus                                                                |
| `EvtBlur`     | —                   | Widget lost focus                                                                  |
| `EvtClick`    | —                   | Mouse button-1 single click                                                        |
| `EvtClose`    | —                   | Popup layer about to close                                                         |
| `EvtHide`     | —                   | Switcher pane became hidden                                                        |
| `EvtShow`     | —                   | Switcher pane became visible                                                       |
| `EvtKey`      | `*tcell.EventKey`   | Unhandled key event bubbling up                                                    |
| `EvtMode`     | `string`            | Editing mode changed (Canvas)                                                      |
| `EvtMouse`    | `*tcell.EventMouse` | Raw mouse event                                                                    |
| `EvtMove`     | `x, y int`          | Highlight/cursor position changed                                                  |
| `EvtPaste`    | `string`            | Text pasted                                                                        |
| `EvtHover`    | —                   | Mouse over widget                                                                  |

For the canonical and complete list with descriptions, see `widgets/events.go` and `doc/events.md`. Typed handlers in `widgets/event-helper.go` unwrap `data[0]` for you (`OnActivate(w, func(int) bool)`, etc.) — prefer them over raw `widget.On(...)`.

## Clean Code

- MUST NOT create getter/setter boilerplate
- MUST NOT use `Get` prefixes for property accessors

## Documentation

- MUST provide doc.go per package directory
- MUST document all exported symbols
- Inline comments MUST explain _Why_, not _What_
- Documentation should be short and concise, but describe parameters and return values
- Examples SHOULD only be part of doc.go

## Error Handling

- MUST wrap errors: `fmt.Errorf("context: %w", err)`
- MUST define sentinel errors for common error cases

## Naming

- Exported: CamelCase
- Unexported: camelCase
- Packages: short, lowercase, no underscores

## Logging

- MUST use structured logging
- MUST use log/slog

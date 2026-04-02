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

| Widget      | File            | Constructor                                                    | Key methods                                                                       |
| ----------- | --------------- | -------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| `Flex`      | `flex.go`       | `NewFlex(id, class, horizontal bool, alignment, spacing int)`  | `Add(Widget)`                                                                     |
| `Grid`      | `grid.go`       | `NewGrid(id, class, rows, cols int, lines bool)`               | `Add(x,y,w,h int, Widget)`, `Rows(…int)`, `Columns(…int)`                         |
| `Box`       | `box.go`        | `NewBox(id, class, title)`                                     | `Add(Widget)` — single child                                                      |
| `Switcher`  | `switcher.go`   | `NewSwitcher(id, class)`                                       | `Add(Widget)`, `Select(any)` — dispatches `EvtShow`/`EvtHide` to panes            |
| `Viewport`  | `viewport.go`   | `NewViewport(id, class, title)`                                | `Add(Widget)` — scrollable single child                                           |
| `Form`      | `form.go`       | `NewForm(id, class, title, data any)`                          | binds a Go struct via reflection                                                  |
| `FormGroup` | `form-group.go` | `NewFormGroup(id, class, title, horizontal bool, spacing int)` | `Add(line int, label string, Widget)`                                             |
| `Collapsible` | `collapsible.go` | `NewCollapsible(id, class, title, expanded bool)`            | `Add(Widget)`, `Expand()`, `Collapse()`, `Toggle()`                               |
| `Dialog`    | `dialog.go`     | `NewDialog(id, class, title)`                                  | `Add(Widget)` — single child; shown as popup via `UI.Popup`                       |
| `Grow`      | `grow.go`       | `NewGrow(id, class, horizontal bool)`                          | `Add(Widget)`, `Start(interval)` — animated reveal                                |

**Flex alignment values:** `"start"`, `"center"`, `"end"`, `"stretch"`

**Grid sizing:** positive = fixed chars; negative = fractional unit; zero = auto (preferred size).
At least one row size and one column size MUST be negative for the Grid to fill available space.

### Input widgets

| Widget       | File            | Constructor                                         | Events                              | Key public methods                                                               |
| ------------ | --------------- | --------------------------------------------------- | ----------------------------------- | -------------------------------------------------------------------------------- |
| `Button`     | `button.go`     | `NewButton(id, class, text)`                        | `EvtActivate`                       | `Activate()`, `SetText(string)`                                                  |
| `Checkbox`   | `checkbox.go`   | `NewCheckbox(id, class, text, checked bool)`        | `EvtChange(bool)`                   | `Checked() bool`, `Toggle()`, `Set(any) bool`                                    |
| `Input`      | `input.go`      | `NewInput(id, class, params…)`                      | `EvtChange(string)`                 | `Text() string`, `SetText(string)`, `SetMask(string)`                            |
| `Typeahead`  | `typeahead.go`  | `NewTypeahead(id, class, params…)`                  | `EvtChange(string)` `EvtAccept(string)` | All `Input` methods + `SetSuggest(func(string) []string)`                    |
| `Select`     | `select.go`     | `NewSelect(id, class, val, label, …)`               | `EvtChange(string)`                 | `Select(string)`, `Value() string`                                               |
| `List`       | `list.go`       | `NewList(id, class, items…)`                        | `EvtSelect(int)` `EvtActivate(int)` | `SetItems([]string)`, `Items() []string`, `Select(int)`, `Selected() int`        |
| `Deck`       | `deck.go`       | `NewDeck(id, class, render ItemRender, itemHeight)` | `EvtSelect(int)` `EvtActivate(int)` | `SetItems([]any)`, `Select(int)`, `Selected() int` — itemHeight ≥ 1 or panics   |
| `Editor`     | `editor.go`     | `NewEditor(id, class)`                              | `EvtChange`                         | `SetContent([]string)`, `Load(string)`, `Text() string`, `ShowLineNumbers(bool)` |

**`ItemRender`** signature: `func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool)`
`focused` is `true` when the Deck widget itself holds keyboard focus.

### Display widgets

| Widget      | File            | Constructor                                                | Key public methods                                                                   |
| ----------- | --------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| `Static`    | `static.go`     | `NewStatic(id, class, text)`                               | `SetText(string)`, `SetAlignment("left"\|"center"\|"right")`, `Set(any) bool`       |
| `Text`      | `text.go`       | `NewText(id, class, lines []string, follow bool, max int)` | `Add(lines …string)`, `Clear()`, `Set(any) bool`                                     |
| `Styled`    | `styled.go`     | `NewStyled(id, class, text)`                               | `SetText(string)` — markup: `**bold**`, `*italic*`, `__underline__`, `` `code` ``   |
| `Progress`  | `progress.go`   | `NewProgress(id, class, horizontal bool)`                  | `SetValue(int)`, `SetTotal(int)` — `total=0` = indeterminate                        |
| `Spinner`   | `spinner.go`    | `NewSpinner(id, class, sequence string)`                   | `Start(interval)`, `Stop()` — built-in: `Spinners["bar"]`, `Spinners["dots"]`, …    |
| `Scanner`   | `scanner.go`    | `NewScanner(id, class, style)`                             | `Start(interval)`, `Stop()` — styles: `"blocks"`, `"diamonds"`, `"circles"`         |
| `Digits`    | `digits.go`     | `NewDigits(id, class, text)`                               | `SetText(string)` — large ASCII-art numerals                                         |
| `Tabs`      | `tabs.go`       | `NewTabs(id, class)`                                       | `Add(title string)`, `Select(int)`, `Selected() int`                                 |
| `Table`     | `table.go`      | `NewTable(id, class, provider TableProvider)`              | `Set(TableProvider)` — events: `EvtSelect(int)`, `EvtActivate(row)`                  |
| `Tree`      | `tree.go`       | `NewTree(id, class)`                                       | `Add(*TreeNode)`, `Selected() *TreeNode` — events: `EvtSelect`, `EvtActivate`, `EvtChange` |
| `Canvas`    | `canvas.go`     | `NewCanvas(id, class, pages, width, height int)`           | `SetCell(x,y,ch,style)`, `Clear()`, `SetMode(string)` — modes: `NORMAL`, `INSERT`, … |
| `Rule`      | `rule.go`       | `NewHRule(class, style)` / `NewVRule(class, style)`        | visual divider; `style` is a theme border name                                       |
| `Terminal`  | `terminal.go`   | `NewTerminal(id, class)`                                   | `Write([]byte) (int, error)`, `Clear()`, `Resize(w, h int)`, `Title() string`        |

**`Terminal`** implements `io.Writer` — pipe pty/subprocess output directly into it.
`Write` is goroutine-safe. The buffer auto-resizes to the widget's content area on first render.
Handles full VT100/ANSI: cursor, erase, SGR (16/256/true-colour, underline styles), alternate screen,
scroll regions, OSC titles. See `.claude/skills/zeichenwerk/widgets.md` § Terminal for details.

**`TreeNode`**: `NewTreeNode(label string, data …any)`, `Add(*TreeNode)`,
`SetLoader(func(*TreeNode))` for lazy-loading children.
`TreeFS` wraps `Tree` for filesystem browsing: `NewTreeFS(id, class, root string, hidden bool)`.

**`TableProvider` interface**: `Columns() []Column`, `Length() int`, `Value(row, col int) string`.
`NewArrayTableProvider(headers []string, rows [][]string)` — in-memory implementation.

### Custom / extension

| Widget            | File           | Use when                                                                  |
| ----------------- | -------------- | ------------------------------------------------------------------------- |
| `Custom`          | `custom.go`    | Simple custom rendering; pass a `func(Widget, *Renderer)` to `NewCustom`  |
| embed `Component` | `component.go` | Full custom widget; implement `Render(*Renderer)` and `Apply(*Theme)`     |
| embed `Animation` | `animation.go` | Custom widget needs a timed render loop; embed and call `Start(interval)` |

### Utility functions (`helper.go`, `container.go`)

```go
Find(container, id)               // Widget by ID (depth-first)
FindAll[T](container)             // All widgets of type T
FindAt(container, x, y)           // Widget at screen coordinates
FindUI(widget)                    // Walk up tree to root *UI
Traverse(container, fn)           // Depth-first walk

OnKey(widget, func(*tcell.EventKey) bool)
OnMouse(widget, func(*tcell.EventMouse) bool)
Redraw(widget)                    // Queue single-widget redraw from any goroutine
Update(container, id, value any)  // Type-aware value setter (List, Static, Table, Text)
Suggest(items []string)           // Returns a suggest func for Typeahead prefix matching
```

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

All event constants are defined in `events.go`. Handler signature:
`func(source Widget, event Event, data ...any) bool` — return `true` to stop propagation.

| Constant     | Typical payload          |
| ------------ | ------------------------ |
| `EvtActivate`| `int` (index) or none    |
| `EvtChange`  | varies by widget         |
| `EvtSelect`  | `int` or `*TreeNode`     |
| `EvtAccept`  | `string`                 |
| `EvtShow`    | —                        |
| `EvtHide`    | —                        |
| `EvtMode`    | `string`                 |
| `EvtMove`    | `x, y int`               |

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

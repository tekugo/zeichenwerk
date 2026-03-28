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
- Every new widget MUST be registered in `Builder` (add a method) and in `UI.Apply` (theme hook)
- Containers call `child.SetBounds` and `child.SetParent` during `Layout()`; never touch tcell directly
- Rendering is top-down: `UI.Draw` → each layer → each container → each child
- Events bubble from focused widget up through parents; return `true` from a handler to stop propagation
- `Refresh()` queues a single-widget redraw; `UI.Refresh()` redraws everything

### Selector format

`type/part.class#id:state` — all parts optional.
Examples: `"button"`, `"button.primary"`, `"button#submit:focused"`, `"input/placeholder"`.

## Project Structure

```
zeichenwerk/
+- cmd/        # Command-line tools and demo applications
|  +- demo/    # Demo showcasing all widgets
+- doc/        # Documentation and design proposals
```

## Widget Reference

### Containers

| Widget      | File            | Constructor                                                    | Key methods                                                                      |
| ----------- | --------------- | -------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `Flex`      | `flex.go`       | `NewFlex(id, class, horizontal bool, alignment, spacing int)`  | `Add(Widget)`                                                                    |
| `Grid`      | `grid.go`       | `NewGrid(id, class, rows, cols int, lines bool)`               | `Add(x,y,w,h int, Widget)`, `Rows(…int)`, `Columns(…int)`, `Separator(x,y,type)` |
| `Box`       | `box.go`        | `NewBox(id, class, title)`                                     | `Add(Widget)` — single child                                                     |
| `Switcher`  | `switcher.go`   | `NewSwitcher(id, class)`                                       | `Add(Widget)`, `Show(index int)`                                                 |
| `Viewport`  | `viewport.go`   | `NewViewport(id, class, title)`                                | `Add(Widget)` — scrollable single child; focusable, arrow keys scroll            |
| `Form`      | `form.go`       | `NewForm(id, class, title, data any)`                          | `Add(Widget)` — binds a Go struct via reflection                                 |
| `FormGroup` | `form-group.go` | `NewFormGroup(id, class, title, horizontal bool, spacing int)` | `Add(line int, label string, Widget)`                                            |
| `Dialog`    | `dialog.go`     | `NewDialog(id, class, title)`                                  | `Add(Widget)` — single child; used as popup layer via `UI.Popup`                 |
| `Grow`      | `grow.go`       | `NewGrow(id, class, horizontal bool)`                          | `Add(Widget)`, `Start(interval)` — animated reveal; wraps a single child         |

**Flex alignment values:** `"start"`, `"center"`, `"end"`, `"stretch"`

**Grid sizing:** positive = fixed chars; negative = fractional unit; zero = auto (preferred size)

**Grid separator constants:** `GridH` (horizontal line), `GridV` (vertical line), `GridB` (both)

### Input widgets

| Widget     | File          | Constructor                                       | Events                                      | Key public methods                                                                                                  |
| ---------- | ------------- | ------------------------------------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `Button`   | `button.go`   | `NewButton(id, class, text)`                      | `"click"`                                   | `Click()`, `SetText(string)`                                                                                        |
| `Checkbox` | `checkbox.go` | `NewCheckbox(id, class, text)`                    | `"change"` → `bool`                         | `Checked() bool`, `SetChecked(bool)`                                                                                |
| `Input`    | `input.go`    | `NewInput(id, class, text?, placeholder?, mask?)` | `"change"` → `string`; `"enter"` → `string` | `Text() string`, `SetText(string)`, `SetMax(int)`, flags: `"masked"`, `"readonly"`                                  |
| `Select`   | `select.go`   | `NewSelect(id, class, val, label, …)`             | `"change"` → `string` (value)               | `Select(value string)`, `Value() string` — alternating val/label args                                               |
| `List`     | `list.go`     | `NewList(id, class, items…)`                      | `"select"` → `int`; `"activate"` → `int`    | `SetItems([]string)`, `Items() []string`, `Index() int`, `SetIndex(int)`, `Selection() []int`, `Disable(int)`       |
| `Editor`   | `editor.go`   | `NewEditor(id, class)`                            | `"change"`                                  | `Text() string`, `SetText(string)`, `Lines() []string`, `SetNumbers(width int)`, `SetTab(n int)`, `SetSpaces(bool)` |
| `Scanner`  | `scanner.go`  | `NewScanner(id, class, style)`                    | —                                           | `Start(interval)`, `Stop()` — scanning bar animation; styles: `"blocks"`, `"diamonds"`, `"circles"`                 |

### Display widgets

| Widget     | File          | Constructor                                                | Key public methods                                                                                                                             |
| ---------- | ------------- | ---------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `Static`   | `static.go`   | `NewStatic(id, class, text)`                               | `SetText(string)`, `SetAlignment("left"\|"center"\|"right")`                                                                                   |
| `Text`     | `text.go`     | `NewText(id, class, lines []string, follow bool, max int)` | `Set([]string)`, `Add(string)`, `Clear()` — scrollable; `follow` auto-scrolls to bottom                                                        |
| `Styled`   | `styled.go`   | `NewStyled(id, class, text)`                               | `SetText(string)` — inline markup: `**bold**`, `_italic_`, `__underline__`, `` `code` ``                                                       |
| `Progress` | `progress.go` | `NewProgress(id, class, horizontal bool)`                  | `SetValue(int)`, `SetTotal(int)` — `total=0` = indeterminate/animated                                                                          |
| `Spinner`  | `spinner.go`  | `NewSpinner(id, class, sequence string)`                   | `Start(interval)`, `Stop()` — use `Spinners["bar"]` etc. for built-in sequences                                                                |
| `Digits`   | `digits.go`   | `NewDigits(id, class, text)`                               | `SetText(string)` — large ASCII-art numerals (0-9, A-F, `:`, `.`, `-`)                                                                         |
| `Rule`     | `rule.go`     | `NewHRule(class, style)` / `NewVRule(class, style)`        | — visual divider line; `style` is a theme border name                                                                                          |
| `Tabs`     | `tabs.go`     | `NewTabs(id, class)`                                       | `Add(title string)`, `Select(int)`, `Selected() int` — events: `"change"` → `int`, `"activate"` → `int`                                        |
| `Table`    | `table.go`    | `NewTable(id, class, provider TableProvider)`              | `Set(TableProvider)` — events: `"select"` → `int, []string`; `"activate"` → `int, []string`                                                    |
| `Canvas`   | `canvas.go`   | `NewCanvas(id, class, pages, width, height int)`           | `Set(x,y,ch,style)`, `Clear()`, `SetMode(string)` — vim-style modal editing; modes: `NORMAL`, `INSERT`, `DRAW`, `VISUAL`, `COMMAND`, `PRESENT` |

### Custom / extension

| Widget            | File           | Use when                                                                  |
| ----------------- | -------------- | ------------------------------------------------------------------------- |
| `Custom`          | `custom.go`    | Simple custom rendering; pass a `func(Widget, *Renderer)` to `NewCustom`  |
| embed `Component` | `component.go` | Full custom widget; implement `Render(*Renderer)` and `Apply(*Theme)`     |
| embed `Animation` | `animation.go` | Custom widget needs a timed render loop; embed and call `Start(interval)` |

**TableProvider interface** (`table-provider.go`): `Columns() []TableColumn`, `Length() int`, `Str(row, col int) string`.
`ArrayTableProvider` is the built-in in-memory implementation: `NewArrayTableProvider(headers []string, data [][]string)`.

### Utility functions (`helper.go`, `container.go`)

```go
Find(container, id)         // Widget by ID (depth-first)
FindAll[T](container)       // All widgets of type T
FindAt(container, x, y)     // Widget at screen coordinates
FindUI(widget)              // Walk up tree to root *UI
Traverse(container, fn)     // Depth-first walk

OnKey(widget, func(Widget, *tcell.EventKey) bool)
OnMouse(widget, func(Widget, *tcell.EventMouse) bool)
Redraw(widget)              // Queue single-widget redraw from anywhere
Update(container, id, value any)  // Type-aware value setter (List, Static, Table, Text)
```

## Building a new widget — checklist

1. Create `mywidget.go` with `type MyWidget struct { Component; ... }`
2. Constructor: `NewMyWidget(id, class string, ...) *MyWidget` — call `SetFlag("focusable", true)` if interactive
3. Implement `Render(r *Renderer)` — call `c.Component.Render(r)` first for borders/bg
4. Implement `Apply(theme *Theme)` — call `theme.Apply(w, w.Selector("mywidget"), states...)`
5. Register in `Builder`: add `func (b *Builder) MyWidget(id string, ...) *Builder`
6. Register in `UI.Apply`: add `case *MyWidget: w.Apply(theme)` so `SetTheme` works
7. Add `doc.go` entry and export all public symbols with comments

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

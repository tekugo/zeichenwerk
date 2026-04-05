# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

- **`Styled` extended Markdown support**
  - Horizontal rules (`---` / `***` / `___`) rendered as a `─` line
  - Blockquotes (`> text`) rendered with a `│ ` left border; consecutive lines
    merged into one block
  - Nested lists — indent depth detected from leading spaces (2 per level);
    bullet cycles `•` / `◦` / `▸` across depths for unordered lists
  - Task lists (`- [ ] text` / `- [x] text`) rendered with `☐` / `☑` markers
  - Theme style keys `"styled/bq"` and `"styled/hr"` added to all built-in
    themes
  - Demo updated with sections for all new block types

### Fixed

- `Styled` block styles with no explicit background/foreground now fall back to
  the base `"styled"` style colours, preventing terminal-default background
  bleeding through on rules and headings
- `hr` and `h2` underline rules are one character shorter than the content
  width, leaving a visible space before the scrollbar

---

## v2.0.0-beta.4

### Added

- **`Value[T]` reactive binding** — generic reactive value type (`value.go`)
  - `NewValue[T](initial T)` — creates a reactive value with an initial value
  - `Get()` / `Set(T)` — thread-safe read/write with subscriber notification
  - `Bind(Setter[T])` — subscribes a widget setter; immediately syncs the
    current value
  - `Observe(Widget, ...convert)` — listens to `EvtChange` on a widget and feeds
    updates into the value; optional convert function enables e.g. `Value[int]`
    bound to an `Input`
  - `Subscribe(func(T))` — low-level callback subscription
  - `Derived[A, B](source, convert)` — derives a new `Value[B]` that mirrors a
    `Value[A]` through a conversion function
- **`Setter[T]` interface** — `type Setter[T any] interface { Set(T) }`
  (`setter.go`); replaces the old untyped `Set(value any) bool` convention
- **`Update[T]`** generic helper — finds a widget by ID and calls `Set(T)` if it
  implements `Setter[T]` (`helper.go`)
- **`Progress.Set(int)`** implements `Setter[int]`; replaces `SetValue`
- **`Styled` widget** — minimal Markdown renderer (`styled.go`)
  - Supported block types: `# h1` (box border), `## h2` (bottom rule), `### h3`
    (bold+underline), `#### h4` (bold), paragraphs, `-` unordered lists, `1.`
    ordered lists, fenced code blocks
  - Inline styles: `*italic*`, `**bold**`, `__underline__`, `~~strikethrough~~`,
    `` `code` ``
  - Word-wrapping with correct punctuation attachment (no space before `,` `.`
    etc.) and neutral space segments (spaces never inherit inline decoration)
  - Keyboard scrolling: `↑`/`↓`, `PgUp`/`PgDn`, `Home`/`End`; scrollbar when
    content overflows
  - Content padding read from the `"styled"` style; scrollbar placed at raw
    widget right edge (outside padding)
  - Theme style keys: `"styled"`, `"styled/h1"` – `"styled/h4"`, `"styled/pre"`,
    `"styled/code"` added to all built-in themes
  - `NewStyled(id, class, text)` / `SetText(text)`; `Builder.Styled(id, text)`
    wiring
  - Styled demo panel with all block types and scrolling in `cmd/demo`
- **Value demo panel** in `cmd/demo` — demonstrates three reactive binding
  groups: two `Checkbox` widgets sharing a `Value[bool]`; an `Input` and
  `Digits` sharing a `Value[string]`; an `Input` and `Progress` sharing a
  `Value[int]` with a string→int convert function

### Changed

- **`EvtChange` is no longer fired by programmatic setters** — all widget `Set`
  / `SetText` methods update state silently; `EvtChange` is dispatched only from
  user-interaction paths (key handlers, `Toggle`, `Insert`, etc.)
- **Redundant named setters removed** — `Set(T)` is now the single canonical
  setter on each widget; the following aliases were removed and all call sites
  updated:
  - `Input.SetText` → `Input.Set`
  - `Digits.SetText` → `Digits.Set`
  - `Progress.SetValue` → `Progress.Set`
  - `List.SetItems` → `List.Set` (spurious `EvtSelect` dispatch also removed)
  - `Static.SetText` → `Static.Set`
- **`Static.Set`** still accepts `any`; non-string values are formatted with
  `fmt.Sprintf("%v", value)`
- `"styled"` style gains `WithPadding(0, 1)` in all built-in themes

### Fixed

- `Progress.Set` now calls `Refresh()` so reactive bindings redraw immediately
- **`Sparkline` widget** — compact time-series display using Unicode block
  characters (`▁▂▃▄▅▆▇█`); height-adaptive with `h` content rows giving `h×8`
  discrete levels per column (`sparkline.go`)
  - `ScaleMode`: `Relative` (tallest visible bar = █, good for shape
    comparisons) or `Absolute` (fixed `[Min, Max]` bracket, good for absolute
    magnitude)
  - Ring-buffer data model via `SetCapacity(int)`; left-pads with spaces when
    series is shorter than widget width
  - Dual-colour threshold: `SetThreshold(float64)` switches bars at or above the
    threshold to the `"sparkline/high"` style
  - Gradient mode: `SetGradient(true)` blends the foreground colour smoothly
    from the base colour (at the threshold) to the high colour (at the maximum
    value) instead of a hard cutoff
  - `"sparkline"` and `"sparkline/high"` style keys added to all five built-in
    themes
  - `Builder.Sparkline(id)` wiring; demo panel in `cmd/demo`
- **`Heatmap` widget** — colour-graded activity grid (`heatmap.go`); similar to
  a GitHub contribution graph
  - `NewHeatmap(id, class string, rows, cols int)` allocates a zeroed grid
  - `SetValue(row, col int, v float64)`, `SetRow`, `SetAll` for data updates
  - Cell background interpolates linearly between `"heatmap/zero"` and
    `"heatmap/max"` background colours using `lerpColor`; foreground follows the
    same gradient
  - `SetRowLabels` / `SetColLabels` for optional axis headers
  - `SetCellWidth(int)` — cells wider than 1 give square-ish cells on most
    terminals
  - Style keys `"heatmap"`, `"heatmap/header"`, `"heatmap/zero"`,
    `"heatmap/mid"`, `"heatmap/max"` added to all five built-in themes
  - `Builder.Heatmap(id, rows, cols)` and
    `compose.Heatmap(id, class, rows, cols)` wiring
  - Demo panel in `cmd/demo`
- **`Table` cell navigation mode** — `SetCellNav(true)` switches the table from
  row-select to cell-select mode; `←`/`→` move between columns,
  `Tab`/`Shift+Tab` advance through cells wrapping across rows, `Home`/`End`
  jump to first/last column in the current row, `Ctrl+←`/`Ctrl+→` jump to
  first/last column
- New style selectors `"table/cell"` and `"table/cell:focused"` for the focused
  cell in cell navigation mode; all six built-in themes define these styles
  using a secondary accent colour distinct from the row highlight
- `"table/cell"` unfocused: `$fg0 on $bg3` (subtle elevation) across all themes
- `"table/cell:focused"` per theme: Tokyo `$cyan`, Nord `$frost1`, Midnight Neon
  `$aqua`, Lipstick `$indigo`, Gruvbox Dark `$orange`, Gruvbox Light `$yellow`

### Changed

- **`Table.Selected() (int, int)`** — replaces `GetSelectedRow() int`; returns
  `(row, col)` where col is `-1` in row mode and the focused column index in
  cell mode
- **`Table.SetSelected(row, col int) bool`** — replaces
  `SetSelectedRow(row int) bool`
- **`Table.Offset() (int, int)`** — replaces `GetScrollOffset()`
- **`Table.SetOffset(offsetX, offsetY int)`** — replaces `SetScrollOffset()`
- **`EvtSelect` payload on `Table`** — now `(row int, col int)` instead of
  `(row int, rowData []string)`; col is `-1` in row mode

---

## [2.0.0-beta.2] — 2026-04-02

### Added

- **Terminal widget** — full VT100/ANSI terminal emulator widget (`terminal.go`)
  - `CellBuffer` with four parallel `[]uint32` slices for char, fg, bg, and
    underline colour (`cell-buffer.go`)
  - `Color` type with `ColorDefault`, `PaletteColor`, `TrueColor`, and
    `colorToHex`; embedded xterm-256 palette table
  - `AnsiParser` 7-state VT/ANSI state machine with full UTF-8 support and Kitty
    sub-parameter colon encoding (`ansi.go`)
  - Handles cursor movement (CUU/CUD/CUF/CUB/CUP/HVP/CHA/VPA), erase
    (ED/EL/ECH/DCH), insert/delete lines (IL/DL), scroll (SU/SD), SGR (colours,
    bold/dim/italic/blink/reverse/invisible/strikethrough/underline), alternate
    screen (?1049), auto-wrap (?7), cursor visibility (?25), scroll regions
    (DECSTBM), save/restore cursor (DECSC/DECRC/SCOSC/SCORC), reverse index
    (RI), hard reset (RIS), OSC title
  - Implements `io.Writer` — pipe pty output directly into the widget
  - `Screen.SetUnderline(style int, color string)` added to the `Screen`
    interface for Kitty underline styles
    (none/single/double/curly/dotted/dashed) and colours; implemented in
    `TcellScreen`, stubbed in `mockScreen`
  - `Renderer.SetUnderline` pass-through method
  - `Builder.Terminal(id)` and `compose.Terminal(id, class, options...)` wiring
  - Terminal style keys (`"terminal"`, `"terminal:focused"`) added to all five
    themes
  - Terminal demo panel in `cmd/demo`
- **LipstickTheme** — Charm/Lipgloss-inspired dark theme with fuchsia, indigo,
  and cream palette
- **Deck focus visibility** — `ItemRender` gains a `focused bool` parameter so
  render functions can distinguish focused vs unfocused selected items; all
  three demo apps updated
- **`-debug` flag** — `cmd/demo` and `cmd/showcase` now accept `-debug` to start
  in debug mode

- **`UI.Confirm(title, message, onConfirm, onCancel)`** — modal confirmation
  dialog wired with OK/Cancel buttons and callbacks
- **`UI.Prompt(title, message, onAccept, onCancel)`** — modal prompt dialog with
  a text input field and OK/Cancel buttons
- **`EvtClose` event** — dispatched to a popup layer just before it is removed
  by `UI.Close`

### Changed

- `UI.Quit()` is now safe to call multiple times (protected by `sync.Once`)
- `go.mod`: `github.com/rivo/uniseg` promoted from indirect to direct dependency

### Fixed

- Demo Dialog pane: YES/NO buttons now close the dialog via `ui.Close()`

---

## [2.0.0-beta.1] — 2026-04-01

### Added

- **Composition API** — new `compose` package with a functional, option-based UI
  construction style as an alternative to the Builder API
- **Showcase app** (`cmd/showcase`) — interactive reference application
  demonstrating every widget with live code examples, theme switching, and a
  component editor
- **Tree widget** — collapsible tree view with keyboard navigation, `TreeFS`
  helper for filesystem trees
- **Deck widget** — scrollable list of fixed-height item cards with a custom
  render function
- **Form improvements** — `FormGroup` supports horizontal layout and label
  alignment; struct tags control field width, control type, and read-only state
- **Typeahead widget** — text input with inline ghost-text completions and a
  pluggable `Suggest` function
- **Theme system overhaul** — five built-in themes (Tokyo Night, Midnight Neon,
  Nord, Gruvbox Dark, Gruvbox Light) with a full colour variable system
  (`$bg0`–`$bg3`, `$fg0`–`$fg2`, named accent colours)
- **Tutorial and reference documentation** (`doc/tutorial.md`, `doc/reference/`)
- **Inspector** — debug overlay showing widget tree, layout info, and structured
  log table; activated with `Ctrl+I` in debug mode
- **`compose/compose.go`** — full widget coverage matching the Builder API
- **`cmd/compose/main.go`** — Composition API demo app mirroring `cmd/demo`
- **`FindAll[T]`** generic helper for collecting all widgets of a given type
- **`Update`** helper for setting text on a `Static` widget by ID from any
  handler
- Unicode border styles added via `AddUnicodeBorders`

### Changed

- `ItemRender` type for `Deck` extended with `selected bool` parameter
- `Builder.Add` now calls `widget.Apply(theme)` automatically
- `On`/`OnKey`/`OnMouse` handler signatures updated for consistency
- Various API cleanups and naming improvements throughout

### Fixed

- Deck scrollbar rendering when item count changes
- Collapsible height hint calculation
- Editor cursor positioning after content load
- Dialog hint computation with optional title bar

---

## [1.0.0] — 2026-03-30

### Added

- Initial public release
- Core widget set: `Box`, `Button`, `Canvas`, `Checkbox`, `Collapsible`,
  `Custom`, `Deck`, `Dialog`, `Digits`, `Editor`, `Flex`, `Form`, `FormGroup`,
  `Grid`, `Input`, `List`, `Progress`, `Rule`, `Scanner`, `Select`, `Spinner`,
  `Static`, `Styled`, `Switcher`, `Table`, `Tabs`, `Typeahead`, `Viewport`
- `Builder` fluent API for constructing UI trees
- `Component` base type with style, flag, event, and layout primitives
- `Theme` system with style keys, colour variables, border styles, and string
  tokens
- `TcellScreen` backend via `github.com/gdamore/tcell/v3`
- `Renderer` with `Set`, `Put`, `Text`, `Fill`, `Border`, `Line`, `Repeat`,
  `Scrollbar`
- Event system (`Dispatch`, `On`, `OnKey`, `OnMouse`) with bubbling
- Focus management, cursor reporting, hover tracking
- Debug mode with log panel and real-time info bar
- `Inspector` widget for live widget-tree exploration
- `Animation` overlay system for popups and dialogs
- `TreeFS` filesystem tree helper
- `AGENTS.md` and `doc/` reference documentation

[2.0.0-beta.2]:
  https://github.com/tekugo/zeichenwerk/compare/v2.0.0-beta.1...v2.0.0-beta.2
[2.0.0-beta.1]:
  https://github.com/tekugo/zeichenwerk/compare/v1.0.0...v2.0.0-beta.1
[1.0.0]: https://github.com/tekugo/zeichenwerk/releases/tag/v1.0.0

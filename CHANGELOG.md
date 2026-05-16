# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## v1.1.1 — 2026-05-16

Housekeeping release — no library code changes. Command-only apps moved
to standalone repos so the library `go.mod` stops carrying their deps.

### Changed

- **Apps moved to standalone repos.** `cmd/triebwerk`, `cmd/messwerk`,
  `cmd/tblr`, `cmd/figlet`, `cmd/malwerk`, and `cmd/hal-explorer` are
  now separate Go modules under `tekugo/`. Each consumes zeichenwerk
  as a normal dependency. See the README for the full list.
- **`go.mod` slimmed.** App-only dependencies dropped from the library:
  `fsnotify` + `doublestar` (triebwerk), `otlp` + `grpc` +
  `grpc-gateway` + `genproto` + `protobuf` (messwerk),
  `atotto/clipboard` (tblr). Zeichenwerk now requires only `tcell`,
  `uniseg`, `testify`, and `golang.org/x/tools`.
- Spec docs for the extracted apps moved to their respective repos
  (`doc/spec/{triebwerk,messwerk,tblr,malwerk}.md` removed).

---

## v1.1.0 — 2026-05-15

> **Note on the version number.** This release was being prepared as
> `v2.0.0` throughout the `v2.0.0-beta.1` … `v2.0.0-beta.6` cycle (see
> the development log below). Before tagging, it was renumbered to
> `v1.1.0` because the public API remained backward-compatible with
> `v1.0.0` — no consumer code needs to change. The `v2.0.0` slot is
> reserved for a future release that actually breaks the API.

All work captured in the `v2.0.0-beta.*` sections below is part of
this release. Highlights:

### Added

- **Composition API** — functional, option-based alternative to the
  Builder (`compose` package); full widget coverage
- **New widgets:** Combo, Filter, Heatmap, Marquee, Shimmer, Sparkline,
  Styled (Markdown renderer), Terminal (VT100/ANSI), Tree, TreeFS,
  Typeahead, Typewriter, CRT, Radio, Slider
- **Reactive bindings** — generic `Value[T]` with `Bind` / `Observe` /
  `Derived` and the `Setter[T]` interface (`value.go`, `setter.go`)
- **Showcase app** (`cmd/showcase`) — interactive widget reference with
  live code examples and theme switching
- **Tutorial and reference docs** — `doc/tutorial.md`, `doc/reference/`
- **Theme system overhaul** — five built-in themes (Tokyo Night,
  Midnight Neon, Nord, Gruvbox Dark, Gruvbox Light, Lipstick) with a
  full colour variable system
- **Table cell-navigation mode** — `SetCellNav(true)` for cell-level
  focus with dedicated `table/cell` / `table/cell:focused` styles
- **Form improvements** — `FormGroup` horizontal layout, struct-tag
  control over field width, control type, and read-only state
- **`Inspector` overlay** — live widget tree, layout info, structured
  log table (`Ctrl+I` in debug mode)
- **Designer PoC**, `zw` command, colour picker, `EvtFocus` / `EvtBlur`,
  focus stack, mouse-wheel support, unicode borders, `FindAll[T]`,
  `Update` helper

### Changed

- Major refactoring: `HFlex` / `VFlex` builder methods, `Set` / `Get`
  cleanup, `Builder.Add` auto-applies the theme, `On` / `OnKey` /
  `OnMouse` handler signatures unified
- `EvtChange` no longer fires from programmatic setters — only from
  user-interaction paths
- `EvtSelect` payload on `Table` now `(row, col)` instead of
  `(row, rowData)`

### Fixed

- Viewport tests for local-coordinate child bounds, scroll axis flags,
  Editor cursor positioning, Deck scrollbar, Collapsible height hint,
  Dialog hint with optional title bar, `Styled` block-style background
  fallback, and many more — see the beta sections for the full list.

---

## Development log — v2.0.0-beta cycle

The entries below document features as they landed during the
`v2.0.0-beta.1` … `v2.0.0-beta.6` development cycle. None of these
betas were tagged on the public repo; the work was released as
**v1.1.0** instead (see above).

---

## v2.0.0-beta.6

### Added

- **`Shimmer` widget** — sweeping highlight band animation (`shimmer.go`)
  - A band of accent colour sweeps left-to-right across the text on every tick;
    useful for loading placeholders and skeleton screens
  - Multi-line text supported — band sweeps the same column on every row
    simultaneously
  - Two intensity modes selectable via `SetGradient(bool)`:
    - *Stepped* (default): linear edge ramp + flat bright core
    - *Cosine gradient*: smooth bell curve, softer organic glow
  - `SetBandWidth(n int)` — core highlight width in columns (minimum 1)
  - `SetEdgeWidth(n int)` — fade columns on each side; 0 = hard edge
  - `"shimmer"` and `"shimmer/band"` style selectors added to all five
    built-in themes; base uses `$fg2` (dimmed) for high contrast against
    the accent band colour
  - `Builder.Shimmer(id)` method added
  - Demo panel added to `cmd/demo` (three rows: stepped, cosine gradient,
    multi-line gradient); stops/starts with switcher visibility via
    `EvtShow`/`EvtHide`

- **`Marquee` widget** — continuously scrolling text ticker (`marquee.go`)
  - Scrolls text wider than the widget from right to left; pauses on hover
  - `SetText(string)`, `SetSpeed(int)`, `SetGap(int)` — wide-char safe
  - `"marquee"` style selector added to all five built-in themes
  - `Builder.Marquee(id)` method added
  - Demo panel added to `cmd/demo`

- **`Typewriter` widget** — character-by-character animated text reveal
  (`typewriter.go`)
  - Three-phase animation: revealing → dwell (cursor blink) → done
  - `SetText(s string)` replaces content and resets the reveal state;
    multi-line text (embedded `\n`) is supported
  - `SetRate(n int)` — runes revealed per tick, clamped to minimum 1
  - `SetDwell(d time.Duration)` — how long the cursor blinks after the
    last character is shown before the animation stops or loops
  - `SetRepeat(v bool)` — continuous looping mode
  - `SetCursor(v bool)` — show or hide the blinking cursor
  - `Reset()` — rewinds to the start without changing text
  - `EvtChange` fired when reveal completes (all runes shown, before dwell)
  - `EvtActivate` fired when dwell expires and `repeat = false`
  - `"typewriter"` and `"typewriter/cursor"` style selectors added to all
    five built-in themes; `"typewriter.cursor"` string key controls the
    cursor character (default `▌`)
  - `Builder.Typewriter(id)` method added
  - Demo panel added to `cmd/demo` (stops/starts with switcher visibility
    via `EvtShow`/`EvtHide`)

- **Commands palette separator** — a `─` rule is now drawn between the
  filter input and the command list; `"commands/group"` style colours are
  used for the separator

### Fixed

- **Commands palette layout** — the filter input inherited `border="round"`
  from its parent `"commands"` style, causing two blank lines between the
  input and list and the list overflowing the bottom border; fixed by adding
  explicit `WithBorder("none")` to `"commands/input"` in all five themes

- **Demo dialog case numbers** — the `Dialog`, `Confirm`, `Prompt`,
  `File Chooser`, and `Dir Chooser` entries in `cmd/demo` were mapped to
  wrong switch-case indices after the Commands and Typewriter panels were
  added; corrected to match their actual positions in the navigation list

- **`CRT` animated root container** — simulates a CRT monitor powering on and
  off (`crt.go`)
  - Power-on: content expands symmetrically from a horizontal centre line
    outward until it fills the screen
  - Power-off: content contracts back to a line, then calls a provided callback
    (typically `ui.Quit`) — `PowerOff(interval, onDone)`
  - During normal operation the container is an invisible pass-through wrapper
    with zero rendering overhead
  - Animation areas filled with Matrix-style green character rain: eight
    brightness levels ramp from near-black far from the scan edge to `#00ff41`
    right beside it; each column scrolls at its own speed via a multiplicative
    hash, keeping the pattern irregular and non-repeating
  - Pulsating `━` scanlines flank the expanding/contracting band edges,
    alternating between `#00ff41` and `#88ffaa` each frame
  - Child content is overlaid with a pulsating green phosphor tint throughout
    the animation; the final four frames flicker between green and true colour
    before settling into `crtPhaseIdle`
  - `NewCRT(id, class string) *CRT` / `compose.CRT(id, class, options...)` API
  - `Start(interval)` begins the power-on animation; safe to call before
    `ui.Run`
  - `PowerOff` interrupts an in-progress power-on cleanly via a
    per-animation-run `done` channel handshake
  - CRT animation wired into `cmd/compose`: power-on runs at startup, `q` /
    `Q` / `Ctrl+C` / `Ctrl+Q` trigger the power-off animation before exit

---

## v2.0.0-beta.5

### Added

- **`Combo` widget** — traditional combo box (`combo.go`)
  - Collapsed single-line display of the current value with a `▼` indicator at
    the right edge; popup opens automatically when the widget gains focus or
    when `Enter` is pressed while focused
  - Popup contains a `Typeahead` input (pre-filled with the current value) and a
    filtered `List` of suggestions in a `Box > Flex` overlay
  - `↓`/`↑` copy the highlighted list item into the input (`PgDn`/`PgUp` page
    through); `Tab`/`→` accept ghost-text; `Enter` confirms; `Esc` dismisses
  - `EvtChange` — dispatched on every keystroke while the popup is open (string)
  - `EvtActivate` — dispatched with the confirmed string when `Enter` is pressed
  - `NewCombo(id, class string, items []string)` / `Builder.Combo(id, items...)`
    / `compose.Combo(id, class, items, options...)` wiring
  - `"combo"` and `"combo:focused"` style keys added to all five built-in themes
    (same palette as `"select"`)
  - Combo demo panel added to `cmd/demo`

- **`EvtFocus` and `EvtBlur` events** — dispatched by `UI.Focus` when a widget
  gains or loses keyboard focus (`events.go`)
  - `EvtFocus` is dispatched after `ui.focus` is updated so that handlers called
    from within the event (e.g. opening a popup) see the correct focus state
  - `UI.Close` restores focus to the widget that was focused before the popup
    opened (saved on a `focusStack` in `UI.Popup`) without dispatching
    `EvtFocus`, preventing widgets like `Combo` from immediately reopening their
    popup

- **`UI` focus stack** — `UI.Popup` pushes the current focus onto an internal
  `focusStack`; `UI.Close` pops it and restores focus silently, replacing the
  previous `SetFocus("first")` call that moved focus to an arbitrary widget

- **Filter demo panel** added to `cmd/demo` — shows a `Filter` input bound to a
  `List` of programming languages; typing filters in real time with ghost-text
  prefix completion

### Fixed

- **`List` stale rows after filtering** — `Render` now fills rows below the last
  visible item with the background style, clearing old content when the item
  count shrinks

- **`Viewport` scroll-axis control** — `FlagVertical` and `FlagHorizontal` added
  to restrict a viewport to a single scroll axis. When `FlagVertical` is set the
  child fills the viewport width and only a vertical scrollbar is shown; when
  `FlagHorizontal` is set the child fills the viewport height and only a
  horizontal scrollbar is shown. Arrow keys for the inactive axis are ignored.
  Default behaviour (both axes) is unchanged.

- **`Styled` extended Markdown support**
  - Horizontal rules (`---` / `***` / `___`) rendered as a `─` line
  - Blockquotes (`> text`) rendered with a `│` left border; consecutive lines
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

[1.1.1]: https://github.com/tekugo/zeichenwerk/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/tekugo/zeichenwerk/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/tekugo/zeichenwerk/releases/tag/v1.0.0

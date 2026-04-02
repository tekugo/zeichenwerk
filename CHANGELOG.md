# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [2.0.0-beta.2] — 2026-04-02

### Added
- **Terminal widget** — full VT100/ANSI terminal emulator widget (`terminal.go`)
  - `CellBuffer` with four parallel `[]uint32` slices for char, fg, bg, and underline colour (`cell-buffer.go`)
  - `Color` type with `ColorDefault`, `PaletteColor`, `TrueColor`, and `colorToHex`; embedded xterm-256 palette table
  - `AnsiParser` 7-state VT/ANSI state machine with full UTF-8 support and Kitty sub-parameter colon encoding (`ansi.go`)
  - Handles cursor movement (CUU/CUD/CUF/CUB/CUP/HVP/CHA/VPA), erase (ED/EL/ECH/DCH), insert/delete lines (IL/DL), scroll (SU/SD), SGR (colours, bold/dim/italic/blink/reverse/invisible/strikethrough/underline), alternate screen (?1049), auto-wrap (?7), cursor visibility (?25), scroll regions (DECSTBM), save/restore cursor (DECSC/DECRC/SCOSC/SCORC), reverse index (RI), hard reset (RIS), OSC title
  - Implements `io.Writer` — pipe pty output directly into the widget
  - `Screen.SetUnderline(style int, color string)` added to the `Screen` interface for Kitty underline styles (none/single/double/curly/dotted/dashed) and colours; implemented in `TcellScreen`, stubbed in `mockScreen`
  - `Renderer.SetUnderline` pass-through method
  - `Builder.Terminal(id)` and `compose.Terminal(id, class, options...)` wiring
  - Terminal style keys (`"terminal"`, `"terminal:focused"`) added to all five themes
  - Terminal demo panel in `cmd/demo`
- **LipstickTheme** — Charm/Lipgloss-inspired dark theme with fuchsia, indigo, and cream palette
- **Deck focus visibility** — `ItemRender` gains a `focused bool` parameter so render functions can distinguish focused vs unfocused selected items; all three demo apps updated
- **`-debug` flag** — `cmd/demo` and `cmd/showcase` now accept `-debug` to start in debug mode

- **`UI.Confirm(title, message, onConfirm, onCancel)`** — modal confirmation dialog wired with OK/Cancel buttons and callbacks
- **`UI.Prompt(title, message, onAccept, onCancel)`** — modal prompt dialog with a text input field and OK/Cancel buttons
- **`EvtClose` event** — dispatched to a popup layer just before it is removed by `UI.Close`

### Changed
- `UI.Quit()` is now safe to call multiple times (protected by `sync.Once`)
- `go.mod`: `github.com/rivo/uniseg` promoted from indirect to direct dependency

### Fixed
- Demo Dialog pane: YES/NO buttons now close the dialog via `ui.Close()`

---

## [2.0.0-beta.1] — 2026-04-01

### Added
- **Composition API** — new `compose` package with a functional, option-based UI construction style as an alternative to the Builder API
- **Showcase app** (`cmd/showcase`) — interactive reference application demonstrating every widget with live code examples, theme switching, and a component editor
- **Tree widget** — collapsible tree view with keyboard navigation, `TreeFS` helper for filesystem trees
- **Deck widget** — scrollable list of fixed-height item cards with a custom render function
- **Form improvements** — `FormGroup` supports horizontal layout and label alignment; struct tags control field width, control type, and read-only state
- **Typeahead widget** — text input with inline ghost-text completions and a pluggable `Suggest` function
- **Theme system overhaul** — five built-in themes (Tokyo Night, Midnight Neon, Nord, Gruvbox Dark, Gruvbox Light) with a full colour variable system (`$bg0`–`$bg3`, `$fg0`–`$fg2`, named accent colours)
- **Tutorial and reference documentation** (`doc/tutorial.md`, `doc/reference/`)
- **Inspector** — debug overlay showing widget tree, layout info, and structured log table; activated with `Ctrl+I` in debug mode
- **`compose/compose.go`** — full widget coverage matching the Builder API
- **`cmd/compose/main.go`** — Composition API demo app mirroring `cmd/demo`
- **`FindAll[T]`** generic helper for collecting all widgets of a given type
- **`Update`** helper for setting text on a `Static` widget by ID from any handler
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
- Core widget set: `Box`, `Button`, `Canvas`, `Checkbox`, `Collapsible`, `Custom`, `Deck`, `Dialog`, `Digits`, `Editor`, `Flex`, `Form`, `FormGroup`, `Grid`, `Input`, `List`, `Progress`, `Rule`, `Scanner`, `Select`, `Spinner`, `Static`, `Styled`, `Switcher`, `Table`, `Tabs`, `Typeahead`, `Viewport`
- `Builder` fluent API for constructing UI trees
- `Component` base type with style, flag, event, and layout primitives
- `Theme` system with style keys, colour variables, border styles, and string tokens
- `TcellScreen` backend via `github.com/gdamore/tcell/v3`
- `Renderer` with `Set`, `Put`, `Text`, `Fill`, `Border`, `Line`, `Repeat`, `Scrollbar`
- Event system (`Dispatch`, `On`, `OnKey`, `OnMouse`) with bubbling
- Focus management, cursor reporting, hover tracking
- Debug mode with log panel and real-time info bar
- `Inspector` widget for live widget-tree exploration
- `Animation` overlay system for popups and dialogs
- `TreeFS` filesystem tree helper
- `AGENTS.md` and `doc/` reference documentation

[Unreleased]: https://github.com/tekugo/zeichenwerk/compare/v2.0.0-beta.2...HEAD
[2.0.0-beta.2]: https://github.com/tekugo/zeichenwerk/compare/v2.0.0-beta.1...v2.0.0-beta.2
[2.0.0-beta.1]: https://github.com/tekugo/zeichenwerk/compare/v1.0.0...v2.0.0-beta.1
[1.0.0]: https://github.com/tekugo/zeichenwerk/releases/tag/v1.0.0

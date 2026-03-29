# Next Steps

Suggestions for future development, combining user proposals with additional recommendations.
Items within each section are ordered roughly by implementation effort.

---

## Widgets — Enhancements to existing widgets

### Arrow navigation and type-ahead for Select ✓ suggested

**Recommended. High value, low effort.**

The popup is already implemented. Two things are missing:

- **Arrow key opens the popup** in addition to Enter, which is the standard TUI convention.
- **Type-ahead filtering** narrows the visible list as the user types, using the same
  first-letter search already present in `List`. A separate filter input inside the popup
  box would give the most flexibility.

This is the easiest high-impact change on the list — the scaffolding is in place.

### Cell selection in Table ✓ suggested

**Recommended for data-heavy applications.**

The `column` field in `Table` is already declared but unused. Enabling it would allow:

- **Cell-level cursor** with `Left`/`Right` to move across columns within a row.
- **Range selection** (Shift+Arrow) marking a rectangular block, dispatching an
  `EvtSelect` event with row/column/span data.
- **Copy to clipboard** of the selected text.

The main design decision is whether cell selection replaces or augments the existing
row-level selection. Making it optional (a flag) is the safest path.

### Selection for Editor ✓ suggested

**Essential for a real editor. Medium effort.**

The cursor model is clean and will make this straightforward. Needed:

- A **mark position** (`markLine`, `markColumn`) set by Shift+Arrow and cleared on
  ordinary movement.
- **Visual highlight** of the selected range during rendering using a `"selected"` style
  part.
- **Cut/copy/paste** using `tcell`'s clipboard or the system clipboard via `xclip`/
  `pbcopy` (the latter as an opt-in, since it requires a subprocess).
- **Select-all** (`Ctrl+A`).

Without selection, the Editor is limited to simple scripting or log viewers. This is
the most user-visible gap in the current widget set.

---

## Widgets — New widgets

### Predefined dialogs (OK/Cancel, Yes/No, Prompt) ✓ suggested

**Recommended. Low effort, very high utility.**

Every application needs these. They are thin wrappers around `Dialog` + `Button`/`Input`
that the library should provide rather than forcing each application to rebuild:

```go
ConfirmDialog(ui, "Delete file?", func(ok bool) { ... })
PromptDialog(ui, "Enter name:", func(text string, ok bool) { ... })
```

These should be non-blocking (callback-based) and respect the existing popup/layer system.
The `Dialog` and `Builder` infrastructure makes this straightforward to implement.

### Tree widget ✓ suggested

**Recommended. Medium effort, broad applicability.**

A tree complements `List` for hierarchical data (file systems, JSON, config, dependency
graphs). Core design:

- Nodes carry a label, optional icon/prefix, and a `[]Node` children slice.
- Collapsed/expanded state per node, toggled with Space or Right Arrow.
- `EvtActivate` on Enter; `EvtSelect` on cursor movement — same contract as `List`.
- Lazy loading: children provided via a callback so large trees don't expand all at once.

The `List` rendering and key-handling code is a natural starting point. Indentation
width and the expand/collapse glyphs should be configurable.

### Notification / Toast

**Recommended. Low effort, high utility.**

A temporary overlay message that auto-dismisses after a timeout:

```go
ui.Notify("Saved", 2*time.Second)
ui.Notify("Error: connection refused", 4*time.Second, LevelError)
```

Displayed as a small floating box (bottom-right corner is conventional) using the
existing popup/layer system. The `Animation` infrastructure can drive the fade-out timer.
This is a very common need and trivially composed from existing primitives.

### Fuzzy filter input

**Recommended as a companion to List and Tree.**

A small `Input` widget that, when bound to a `List` or `Tree`, progressively filters
visible items using substring or fuzzy matching. Modelled after the pattern in `select.go`
but usable standalone:

```go
filter := NewFilter("filter", "")
filter.Bind(list)
```

Dispatches `EvtChange` with the current filter string. The bound widget re-renders only
matching items. Useful for file pickers, command palettes, log viewers.

### Splitter / resizable panes

**Medium priority. Medium effort.**

Allows the user to drag a separator between two `Flex` or `Grid` children to resize them
at runtime. Implementation: a thin `Splitter` widget sitting in a flex layout that
intercepts mouse drag events and adjusts the hints of its two neighbours. Requires no
changes to the container API — only `SetHint` calls on the adjacent children.

### Calendar / date picker ✓ suggested

**Optional. Medium effort.**

Useful for date-entry forms. A month-grid widget that:

- Renders a 7×6 cell grid of day numbers.
- Navigates months with `Left`/`Right` Page keys.
- Dispatches `EvtActivate` with a `time.Time` value on Enter.
- Can be used standalone or inside a `PromptDialog`.

Worth implementing if the library targets business/data-entry applications. Less
essential for developer tools or system utilities.

### Color picker ✓ suggested

**Low priority unless tied to the style editor.**

A terminal color picker is inherently limited by the 256-colour palette (or
24-bit where supported). The most practical design is a palette grid (16 base colours +
216 colour cube + greyscale ramp) with arrow navigation and a preview swatch.

The main use case is the Inspector style editor (see below), where it fits naturally.
As a standalone widget it has limited general utility — most TUI apps don't need
runtime colour selection.

---

## Developer tooling

### Style editor for Inspector ✓ suggested

**Recommended alongside the Inspector itself.**

The Inspector already shows widget info and style properties. Extending it to allow
live edits would make it a proper theme development tool:

- A property panel showing the selected style's fields (fg, bg, font, border, padding,
  margin) as editable `Input` and `Select` widgets.
- Changes applied immediately via `widget.SetStyle(selector, style)` so the effect
  is visible in real time.
- A "Copy as Go" button that emits the style definition as a code snippet.

The Color picker widget would integrate here as the value editor for `fg`/`bg` fields.

---

## Architecture improvements

### Fix `go.mod` CGo dependency

**High priority. See open issues.**

Move `go-sqlite3`, `figlet4go`, and `golang.org/x/tools` to a `cmd/go.mod` workspace
module so library consumers are not forced to carry a C toolchain dependency.

### Typed `Option` struct for `NewSelect`

**Low effort. See open issues.**

Replace the alternating-string variadic with `...Option{Value, Label string}`. A small
breaking change but the correct API before the library reaches wider adoption.

### Scrollable container (Viewport improvements)

The existing `Viewport` widget handles scrollable content. If it doesn't yet support
arbitrary widget trees as children (only text), extending it to do so would enable
long forms and dashboards without manual pagination.

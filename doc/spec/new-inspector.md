# Inspector

A unified developer console for any running `*UI`: widget hierarchy explorer,
style editor (widget-local **and** theme-wide), log viewer, event-handler map,
and an interactive widget creator that doubles as a UI designer with Go code
generation. Opens as a popup via `Ctrl+D` over a live application; can also be
launched stand-alone in *empty-canvas* mode (`cmd/inspector`) to design a fresh
UI from scratch.

The new Inspector replaces the existing minimal `Inspector` in `inspector.go`.
The current implementation only offers the Widgets and Debug Log tabs and is
read-only; the new one keeps that surface as a drop-in starting point and
extends it with editing, code generation, and event-handler introspection.

---

## Goals

1. **Inspect** any widget tree at runtime — drill down, see bounds, hint, state,
   flags, computed style.
2. **Edit** styles live, both widget-local overrides and the active theme.
3. **Read** the in-memory `TableLog` with filtering by level/source.
4. **Discover** which widget handles which event, and — where the runtime
   permits — show the Go function name behind each `Handler`.
5. **Build** new widgets into the tree (or into a blank document) and generate
   compilable Go code using the Builder API and the Compose API.

The Inspector is a *tool*, not a runtime concern of the application under
inspection. It must not mutate the application unless the user explicitly
performs an editing action.

---

## Visual layout

The Inspector is a `Box` with a double border, hosting a vertical `Flex` with
a `Tabs` row on top and a `Switcher` below. Each tab swaps the content pane.

```
┌─ Inspector ──────────────────────────────────────────────────────────────────┐
│ [ Widgets ] [ Styles ] [ Events ] [ Log ] [ Designer ]      Mode: ATTACHED   │
├──────────────────────────────────────────────────────────────────────────────┤
│  …tab content…                                                               │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

The mode label at the top right is one of:

| Mode | Meaning |
|------|---------|
| `ATTACHED` | Inspecting the host application (Ctrl+D popup). Edits write back to the host. |
| `DESIGN`   | Editing a stand-alone document (`cmd/inspector`). Edits build a virtual tree. |
| `READONLY` | Inspecting a snapshot (e.g. `-dump`-like JSON file). No edits possible. |

---

## Tabs

### 1. Widgets tab

The hierarchy explorer. Three columns laid out by a horizontal `Flex`:

```
┌─ Widgets ──────────────────────────────────────────────────────────────────┐
│ ui > flex#root > grid#body > list#nav                                      │  ← breadcrumb
├──────────────────┬─────────────────────────┬───────────────────────────────┤
│  Children        │  Styles (selectors)     │  Information                  │
│  ─────────       │  ─────────              │  Type      List               │
│  ▶ list#nav      │  (default)              │  ID        nav                │
│    grid          │  list:focused           │  Class                        │
│    static#title  │  list/highlight         │  Parent    grid#body          │
│                  │  list/scrollbar         │  Bounds    x=4 y=2 w=28 h=20  │
│                  │                         │  Content   x=5 y=3 w=26 h=18  │
│                  │                         │  Hint      w=28 h=-1          │
│                  │                         │  State     :focused           │
│                  │                         │  Flags     focusable, focused │
│                  │                         │  Summary   8 items, idx=2     │
│                  │  [+ Style]              │  [Edit Style]   [+ Child]     │
└──────────────────┴─────────────────────────┴───────────────────────────────┘
```

#### Children list (left)

A `List` showing direct children of the *current container* by ID
(`type#id` for unlabelled widgets, just `id` otherwise). Empty IDs are
displayed as `<type@addr>` so every row is unique. A leading `▶` marks
container children — pressing `Enter` (or double-click) descends into them.
`Backspace` ascends to the parent. The row reflects flag state via styling:
focused widgets are bold; hidden widgets are dim.

`<` / `>` in the list moves the highlight up / down; `j` / `k` likewise.

#### Styles list (centre)

The selectors registered on the *currently selected widget* (via the
`stylesProvider` interface — see §Required-additions). `(default)` stands
for the empty selector. Selecting a row populates the Information panel with
that style's resolved values and enables the `[Edit Style]` button.

#### Information panel (right)

A multi-line `Text` widget showing widget metadata when a child is highlighted,
or style metadata when a style is highlighted. Format mirrors the existing
`widgetDetails` helper (see `inspector.go:198-229`) and adds:

- **Type** with its package prefix stripped.
- **State** as a CSS pseudo-class string (`:focused`, `:hovered`, …).
- **Summary** for widgets implementing the existing `Summarizer` interface.
- **Style chain** (resolved from theme + widget-local overrides).

#### Toolbar

Below the panels:

| Button | Hotkey | Action |
|--------|--------|--------|
| `[+ Child]` | `a` | Open widget picker (containers only) — see §Designer flow |
| `[+ Sibling]` | `A` | Add sibling after current widget |
| `[Edit Style]` | `e` | Switch to **Styles** tab focused on this widget+selector |
| `[Delete]` | `d` | Remove widget from tree (ATTACHED + DESIGN modes) |
| `[Locate]` | `l` | Briefly flash the widget on screen (ATTACHED only) |
| `[Generate]` | `Ctrl+G` | Switch to **Designer** tab with code panel populated |

The `[Delete]`, `[+ Child]`, `[+ Sibling]` buttons are hidden in `READONLY`
mode. In `ATTACHED` mode they are present but show a confirmation dialog
warning the user that the host application is being mutated.

### 2. Styles tab

A two-pane layout that hosts the existing **`StyleEditor`** widget
(see `doc/spec/style-editor.md`) on the right and a *scope selector* on the
left:

```
┌─ Styles ───────────────────────────────────────────────────────────────────┐
│ Scope:                       │  ┌─ Selectors ──┐  ┌─ Properties ──────┐  │
│  ( ) Theme                   │  │ Filter [   ] │  │ Selector list:foc │  │
│  (•) Widget overrides        │  │ list         │  │ Fg     [$fg0    ] │  │
│      (list#nav)              │  │ list:focused │  │ Bg     [$bg2    ] │  │
│  ( ) Type-wide               │  │ list/highl…  │  │ Border [round  ▼] │  │
│      (List)                  │  │ list/scroll… │  │ Font   [bold    ] │  │
│                              │  └──────────────┘  └───────────────────┘  │
│  [Reset]  [Save Theme As…]   │  ┌─ Preview ────────────────────────────┐  │
│                              │  │  Sample                              │  │
│                              │  └──────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

The scope radio group controls which `*Theme` the embedded `StyleEditor`
reads/writes:

| Scope | Source | Write target |
|-------|--------|--------------|
| **Theme** | `ui.Theme()` | The live theme (every widget that resolves the selector picks the change up). |
| **Widget overrides** | `widget.Styles()` | The selected widget's own style map (stored on `Component.styles`). |
| **Type-wide** | A synthetic `*Theme` filtered to selectors prefixed `type` | Live theme with selector pinned to the type name. |

`[Save Theme As…]` opens a `FileChooser` and writes the live theme back to a
Go source file using the existing theme-codegen helper (`themes/codegen.go`).
`[Reset]` reloads the theme from disk (clears unsaved edits) — guarded by a
confirmation dialog.

### 3. Events tab

A 2-column layout: a `Table` of `(widget, event, handler)` triplets on the
left, and a "Source" panel on the right with the function's qualified name
and — when a `.go` source file is reachable — a snippet around the line.

```
┌─ Events ───────────────────────────────────────────────────────────────────┐
│ Filter: [ ]   Event: [all ▼]   Show only: [ ] Focusables                  │
├──────────────────────────────────┬─────────────────────────────────────────┤
│ Widget          Event    Handler │  Function                               │
│ list#nav        select   listSel │  cmd/demo/main.go                       │
│ list#nav        activate (lambda)│  func main.listSelected(idx int) bool   │
│ button#quit     activate quitFn  │   23 │                                  │
│ input#filter    change   onFlt   │   24 │ func listSelected(idx int) bool {│
│ ui              key      (anon)  │   25 │     ui.Log(ui, Debug, "sel %d",  │
│ …                                │   26 │         idx)                    │
│                                  │   27 │     return false                │
│                                  │   28 │ }                               │
└──────────────────────────────────┴─────────────────────────────────────────┘
```

#### Data sources

For every widget reachable from `ui` via `core.Traverse`, the inspector
reads:

1. The widget's `Handlers() map[Event][]Handler` — see §Required-additions.
2. For each handler, `runtime.FuncForPC(reflect.ValueOf(h).Pointer())`
   returns a `*runtime.Func` whose `Name()` produces e.g.
   `github.com/me/pkg.listSelected.func1` and whose `FileLine(pc)` returns
   `(file, line)`.

The `Handler` column shows a short label:

| Source | Display |
|--------|---------|
| Top-level function (`pkg.Foo`) | `Foo` |
| Method value (`pkg.(*T).M`) | `T.M` |
| Closure (`...func1`) | `(lambda)` (or, if the closure has a parent
function name, `Foo→λ`) |
| Compiler-synthesised `OnSelect` wrappers | unwrap one level so the user-
supplied function shows through |

The full qualified name is shown in the right panel along with the source
location. If the file is reachable on disk (typical during development —
`runtime.GOROOT()` and the module's source tree exist), 5 lines around
`FileLine` are read and shown in a read-only `Editor`-style snippet with line
numbers. If the function is unreachable (stripped binary, embedded source not
present), the snippet area says `<source not available>`.

#### Filters

- **Filter** — substring match against any column.
- **Event** — drop-down with all known event constants plus `all`.
- **Focusables only** — limits to widgets with `FlagFocusable`.

#### Unwrapping typed wrappers

`widgets/event-helper.go` defines wrappers like `OnSelect(w, fn)` that store
an internal closure on the widget. The closure's name resolves to e.g.
`zeichenwerk/widgets.OnSelect.func1`, which is unhelpful. The Inspector
detects names matching the regex
`^.*/widgets\.On[A-Z][a-zA-Z]+\.func\d+$`
and tries to retrieve the wrapped user function by reading the closure's
captured pointer via `reflect`. If that fails, it falls back to displaying
the wrapper's own name and tags it `(typed wrapper)`.

### 4. Log tab

The existing log table, lifted directly from `inspector.go:64-72`. The body
is a `Table` widget bound to `ui.tableLog` (`*TableLog`). Additions:

- **Level filter** (`Combo`): `all`, `error`, `warn`, `info`, `debug`.
- **Source filter** (`Input` typeahead): substring filter on the `Source`
  column.
- **Auto-follow** (`Checkbox`): when checked (default), the table jumps to the
  latest entry on every new log record. Unchecked → the table preserves the
  user's scroll position.
- **Clear** button — empties the circular buffer (calls a new
  `(*TableLog).Reset()` method, see §Required-additions).
- **Export** button — writes the buffer as JSON or NDJSON via `FileChooser`.

### 5. Designer tab

The interactive widget creator. The full data model and code generator are
specified in `doc/spec/designer.md`; this tab embeds that designer almost
verbatim, with one difference:

- In **ATTACHED** mode, the designer's *root* is the same `*UI` the Inspector
  is attached to (a live tree), and adding widgets actually mutates the
  application. The toolbar's `Save` / `Load` buttons are replaced with
  `[Apply]` / `[Revert]` that toggle a virtual overlay.
- In **DESIGN** mode (stand-alone `cmd/inspector`), the designer behaves
  exactly as in `doc/spec/designer.md`: a `*Document` JSON file is the
  source of truth.

```
┌─ Designer ─────────────────────────────────────────────────────────────────┐
│ [+Child] [+Sibling] [↑] [↓] [Indent] [Outdent] [Cut] [Copy] [Paste]       │
│ [Delete]                            [Preview ▾] [Code ▾] [Save] [Apply]   │
├──────────────┬─────────────────────────────┬─────────────────────────────┤
│  Tree        │  Properties                  │  Preview │ Code              │
│  ▼ flex root │  Kind     Flex               │   …live  │ func buildUI() { │
│    flex …    │  ID       [root          ]  │   widget │   NewBuilder(…)   │
│    grid body │  Class    [              ]  │   render │     .Flex("root", │
│      list nav│  Hint W   [-1            ]  │          │       false,…)    │
│      view m  │  Hint H   [-1            ]  │          │ }                 │
│      static c│  Padding  [0 0 0 0       ]  │          │                  │
│              │  ─── Flex ──────────────────│          │                  │
│              │  Horizontal [ ]              │          │                  │
│              │  Alignment  [stretch     ▼] │          │                  │
│              │  Spacing    [0           ]  │          │                  │
└──────────────┴─────────────────────────────┴─────────────────────────────┘
```

Design-mode internals (DesignerNode, palette, codegen, validation) are
**reused** from `doc/spec/designer.md`; the Inspector's tab is a thin
wrapper that wires the document to the same codegen pass.

#### Code panel

A read-only `Text` widget with a `Tabs` selector for `Builder | Compose`.
The generated code updates whenever the tree changes (debounced 200 ms).
Buttons:

- `[Copy]` — writes the visible code to the OS clipboard via
  `tcell/v3` clipboard support.
- `[Save…]` — writes the code to a `.go` file via `FileChooser`.

#### Apply (ATTACHED mode)

`[Apply]` rebuilds the tree from the document, calls `ui.Layout()` and
`ui.Refresh()`, and reports diagnostics to the Log tab. If the build returns
an error, the Apply is aborted and the host UI remains untouched.

`[Revert]` discards the in-memory document and reloads it from the live
tree.

---

## Required additions to the library

The Inspector intentionally introduces only small, generic accessors, so the
core library remains free of inspector-specific concerns.

### 1. `Widget.Handlers()`

Add a method on `Component`:

```go
// Handlers returns a snapshot of the registered event handlers indexed by
// event. The returned map is a copy; mutating it does not affect the widget.
// Used by the Inspector and by the dump tooling.
func (c *Component) Handlers() map[Event][]Handler
```

Slice values are copies of the internal slices. The method is exposed via a
new optional interface `HandlerProvider` so the Inspector can probe it
without forcing a hard dependency:

```go
type HandlerProvider interface {
    Handlers() map[Event][]Handler
}
```

### 2. `Widget.Styles()`

Promote the ad-hoc `stylesProvider` interface in `inspector.go:107-110` to
a named interface in the library:

```go
type StylesProvider interface {
    Styles() []string  // selector keys registered on the widget
    Style(selector string) *Style
}
```

`Component` already implements `Style(selector string)`. Add a
`Styles() []string` method that walks `c.styles` and returns the keys.

### 3. `(*TableLog).Reset()`

```go
// Reset empties the circular buffer. Used by the Inspector "Clear" action.
func (t *TableLog) Reset()
```

### 4. `(*UI).Inspector()`

```go
// Inspector returns the active Inspector instance, creating one on first
// call. The returned Inspector lives until the UI is closed.
func (ui *UI) Inspector() *Inspector
```

The keybinding in `ui.go:228-236` is rewritten to call this:

```go
case tcell.KeyCtrlD:
    ui.Inspector().Toggle()
```

`Toggle()` calls `ui.Popup` if hidden and `ui.Close` if already shown.

---

## Public API

```go
// Inspector is a developer console for a *UI.
type Inspector struct {
    Component
    ui       *UI
    target   Container        // current container being browsed
    selected Widget           // currently highlighted widget
    document *Document        // designer document (nil until first edit)
    mode     InspectorMode
    layout   Container        // internal builder-built UI
}

type InspectorMode int

const (
    ModeAttached InspectorMode = iota
    ModeDesign
    ModeReadonly
)

// NewInspector creates an Inspector attached to the given root.
// Mode is auto-detected: a live *UI → ModeAttached, a *Document → ModeDesign,
// a snapshot tree → ModeReadonly.
func NewInspector(root Container) *Inspector

// SetMode forces a specific mode (e.g. ModeReadonly) regardless of detection.
func (i *Inspector) SetMode(m InspectorMode)

// Toggle shows or hides the Inspector popup over its UI.
func (i *Inspector) Toggle()

// UI returns the inner builder container (so the popup logic can size and
// position it).
func (i *Inspector) UI() Container

// Document returns the in-memory designer document (lazily created).
func (i *Inspector) Document() *Document

// Refresh re-scans the target tree.
func (i *Inspector) Refresh()
```

### Builder integration

```go
// Inspector adds a hosted Inspector widget into the current builder chain.
// Useful when an application wants to embed the Inspector permanently
// (e.g. inside a developer side-panel) instead of opening it as a popup.
func (b *Builder) Inspector(id string) *Builder
```

### Compose integration

```go
func Inspector(id, class string, options ...Option) Option
```

---

## Events

| Event | Payload | Description |
|-------|---------|-------------|
| `EvtSelect` | `Widget` | Widget highlighted in the explorer |
| `EvtChange` | `*Style` | A style was edited |
| `EvtActivate` | `Widget` | Container was navigated into |
| `EvtClose` | — | Inspector popup is closing |
| `EvtMode` | `InspectorMode` | Mode changed |

---

## Keyboard map

Global (when Inspector has focus):

| Key | Action |
|-----|--------|
| `1`–`5` | Switch to tab 1–5 |
| `Tab` / `Shift+Tab` | Cycle focus inside current tab |
| `Esc` | Close Inspector popup |
| `Ctrl+G` | Open Designer tab with current widget pre-selected |
| `Ctrl+E` | Open Styles tab editing current widget |
| `?` | Show keybindings cheat sheet |

Widgets-tab specific:

| Key | Action |
|-----|--------|
| `↑` / `↓` / `j` / `k` | Move highlight |
| `Enter` | Descend into container |
| `Backspace` | Ascend to parent |
| `a` / `A` | Add child / sibling |
| `d` | Delete (with confirmation) |
| `e` | Edit style (jumps to Styles tab) |
| `l` | Locate (flash widget) |

Designer-tab specific: defer to `doc/spec/designer.md`.

---

## Styling selectors

Inspector-specific selectors are registered in every theme:

| Selector | Applies to |
|----------|-----------|
| `inspector` | Outer container |
| `inspector/box` | Title-bordered box |
| `inspector/tabs` | Tab strip |
| `inspector/breadcrumb` | Path display |
| `inspector/widget-list` | Widget list rows |
| `inspector/widget-list:focused` | Focused list row |
| `inspector/style-list` | Style list rows |
| `inspector/info` | Information panel `Text` |
| `inspector/locate-flash` | Bright accent border drawn during `[Locate]` |
| `inspector/event-table` | Event-handler table |
| `inspector/source` | Source-snippet panel |

The Inspector itself doesn't introduce widget *types* — every cell is an
existing primitive — so theme files only need to add the selectors above.

---

## Locate (flash) effect

`[Locate]` (Widgets tab, key `l`) overlays a 1-frame `Custom` widget over the
selected widget's bounds drawn with the `inspector/locate-flash` style, and
schedules a 200 ms timer that removes it. It does not consume input or
redraw anything else. Useful when the explorer-listed ID is hard to find on
screen.

---

## Implementation plan

The implementation lives in a new `inspector/` package so it can be excluded
from production builds with a build-tag (`!nodebug`) at the user's
discretion.

```
inspector/
├── inspector.go     — Inspector struct, mode detection, NewInspector,
│                     Toggle, public API
├── widgets-tab.go   — Hierarchy explorer panel
├── styles-tab.go    — Style editor scope wrapper around StyleEditor
├── events-tab.go    — Event-handler table + source-snippet panel
├── log-tab.go       — TableLog viewer with filters
├── designer-tab.go  — Embeds the designer document model + codegen
├── locate.go        — Flash overlay
├── handler-meta.go  — runtime.FuncForPC + closure-name unwrapping helpers
└── doc.go
```

### Step-by-step

1. **Library additions**
   - `Component.Handlers()` and the `HandlerProvider` interface.
   - `Component.Styles() []string` and the `StylesProvider` interface.
   - `(*TableLog).Reset()`.
   - `(*UI).Inspector()` returning a cached instance.
   - Replace the inline popup wiring at `ui.go:228-236` with a `Toggle()` call.

2. **Inspector skeleton (`inspector/inspector.go`)**
   - `NewInspector(root)` snapshots the tree, builds the tab shell with
     `NewBuilder`, and stores `*UI` references.
   - Each tab is a private composite widget with its own constructor.

3. **Widgets tab (`inspector/widgets-tab.go`)**
   - Lifts the existing `Inspector` body from `inspector.go:34-95`.
   - Adds toolbar buttons; Edit-Style and `[+ Child]` route to the relevant
     tabs by setting state on the parent Inspector and switching tabs.

4. **Styles tab (`inspector/styles-tab.go`)**
   - Constructs a `StyleEditor` (see `doc/spec/style-editor.md`) and a scope
     selector. The scope selector swaps the `*Theme` passed to the editor.
   - For *Widget overrides* scope: builds a synthetic `*Theme` that contains
     only the widget's own styles; on edit, copies values back via
     `widget.SetStyle(selector, style)`.

5. **Events tab (`inspector/events-tab.go`)**
   - `core.Traverse` over the target. For each widget, call `Handlers()` and
     `runtime.FuncForPC(reflect.ValueOf(h).Pointer())`.
   - Build an `[]eventRow` and feed it to a `Table` via a new
     `eventsTableProvider`.
   - On row `EvtSelect`, populate the source panel.
   - `handler-meta.go` exposes `funcName(h Handler) (qualified, short string)`
     and `funcSource(h Handler) (file string, line int, snippet []string)`.
   - Re-runs the scan whenever the parent `Inspector` fires `EvtChange`
     (debounced 250 ms).

6. **Log tab (`inspector/log-tab.go`)**
   - Reuses `ui.tableLog` directly. Wraps it in a level/source filter
     `TableProvider` that delegates `Length()` and `Str()` after filtering.
   - Auto-follow toggles a slog handler that calls `table.SetSelected(0)` on
     every new record.

7. **Designer tab (`inspector/designer-tab.go`)**
   - Imports the designer model & codegen from `cmd/designer/`. (Move them
     into a new `designer/` library package so both the stand-alone binary
     and the inspector can share the code.)
   - In ATTACHED mode, the designer is initialised by walking the live
     widget tree and emitting `DesignerNode` records — each widget kind has
     a registration entry in `palette.go` that knows how to extract its
     props (`title` from a `Box`, `text` from a `Static`, etc.).

8. **Locate (`inspector/locate.go`)**
   - `flash(w Widget)` adds a translucent `Custom` overlay onto the topmost
     UI layer with bounds equal to `w.Bounds()` and removes it after 200 ms.

9. **Stand-alone binary (`cmd/inspector/main.go`)**
   - Flags: `--file <path.json>` (open a designer document), `--snapshot
     <dump.json>` (open a read-only tree), `--theme`.
   - Without flags, starts in DESIGN mode with an empty `Flex "root"`.

10. **Theme entries**
    - Add the selectors listed in §Styling-selectors to every
      `themes/theme-*.go`.

11. **Tests**
    - `inspector/widgets-tab_test.go` — explorer navigation, breadcrumb.
    - `inspector/events-tab_test.go` — handler-name unwrapping for
      top-level fns, methods, anonymous closures, and the `OnXxx` wrappers.
    - `inspector/handler-meta_test.go` — function-name regex coverage.
    - `inspector/designer-tab_test.go` — round-trip live tree → document →
      generated code → re-parsed document.

---

## Code-generation contract

When the user clicks `[Generate]`, the Inspector emits a single Go file that
compiles against the public `zeichenwerk` API and reproduces the explored
tree. The generator is the same one specified in `doc/spec/designer.md`
(§"Code generation"); the Inspector merely seeds it from the live tree
instead of from a hand-built document.

Differences when generating from a *live* tree:

- Widget IDs that are empty are auto-named `<type><n>` so the generated code
  compiles. The Inspector keeps a name map so re-generation is stable.
- Event handlers are emitted as `// TODO: handler` placeholders, *not* as
  generated code — the Inspector cannot serialise function bodies. The
  comment includes the resolved function name so the developer knows what
  used to be wired up:

  ```go
  builder.Button("quit", "Quit").
      // TODO: On(EvtActivate, main.quitFn)
      End().
  ```

- ItemRender / TableProvider stubs are likewise emitted as `nil` with a
  `// TODO` comment.

---

## Mutation contract (ATTACHED mode)

The Inspector mutates the host UI only in response to *explicit* user
actions:

| Action | Mutation |
|--------|----------|
| Edit style in Styles tab | `widget.SetStyle(selector, *Style)` or `theme.SetStyle(selector, *Style)`; `Relayout(widget)` |
| `[+ Child]` / `[+ Sibling]` | `container.Add(child)`; `Relayout(container)` |
| `[Delete]` | `container.Remove(child)`; `Relayout(container)`; if the deleted widget held focus, refocus the parent |
| `[Apply]` (Designer) | Replace tree subtree under designer root; `ui.Layout()`; `ui.Refresh()` |

All mutations are logged at `slog.Info` level with a `source=inspector`
attribute so they show up in the Log tab. Failures (e.g. a child rejected
because the container does not accept children) are surfaced as a red
toast `Notification` for 3 seconds and recorded as `slog.Warn`.

---

## Non-goals

- **Undo/redo** for inspector edits. Mutations are immediate; users wanting
  undo should drive the application through the Designer document instead.
- **Cross-process inspection.** The Inspector lives inside the host process
  and reads in-memory state. Remote inspection (via JSON-RPC, sockets) is a
  future enhancement — the snapshot/READONLY mode is the seam.
- **Source modification.** The Generate action emits Go code to a file or
  the clipboard; the Inspector never edits the user's existing source.
- **Hot-reload.** Generated code is intentionally non-executing. Re-running
  the binary picks up changes; the Designer's ATTACHED mode is the
  short-feedback path.
- **Full Go function reconstruction.** Event handlers cannot be round-tripped
  through code generation — only their resolved name is preserved.

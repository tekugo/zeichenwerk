# Inspector

A unified developer console for a running `*UI`. Five tabs over a live widget
tree: **Widgets** (the explorer + designer), **Styles** (the context-driven
style editor), **Events** (handler map with function-name introspection),
**Theme** (the theme-style picker), **Log** (the in-memory log table).

The Inspector mutates the *live* widget tree directly. There is no separate
`Document` or `DesignerNode` model — every change touches the actual `*UI`
the Inspector is attached to. Code generation walks the live tree and asks
each widget for its configurable items via the new `Configurable` interface
described in §"Configurable model".

The new Inspector replaces the minimal one in `inspector.go`.

---

## Goals

1. **Inspect** any widget tree at runtime — drill down, see bounds, hint, state,
   flags, computed style.
2. **Edit** styles live, both widget-local overrides and the active theme.
3. **Read** the in-memory `TableLog` with filtering by level/source.
4. **Discover** which widget handles which event, and — where the runtime
   permits — show the Go function name behind each `Handler`.
5. **Build** new widgets into the tree, edit their parameters, and generate
   compilable Go code (Builder API or Compose API) from the live tree.

The Inspector is a *tool*, not a runtime concern of the application under
inspection. Mutations only happen in response to explicit user actions.

---

## Visual layout

The Inspector is a `Box` with a double border, hosting a vertical `Flex` with
a `Tabs` row at the top and a `Switcher` for the five panes underneath.

```
┌─ Inspector ──────────────────────────────────────────────────────────────────┐
│ [ Widgets ] [ Styles ] [ Events ] [ Theme ] [ Log ]      Mode: ATTACHED      │
├──────────────────────────────────────────────────────────────────────────────┤
│  …tab content…                                                               │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

The mode label at the top right is one of:

| Mode | Meaning |
|------|---------|
| `ATTACHED` | Inspecting a live `*UI` (Ctrl+D popup). Edits write back. |
| `READONLY` | Inspecting a snapshot or detached tree. No edits possible. |

The `DESIGN` mode from earlier drafts is gone: a stand-alone designer with no
host UI just attaches to a fresh `*UI` whose root is an empty `Flex`, so it
falls into the ATTACHED case.

---

## Configurable model

Every widget that wants to be editable in the Inspector implements
`Configurable`. This single interface drives the property editor *and* code
generation, so the two cannot drift apart.

```go
// PropertyKind describes the editor type used for a property.
type PropertyKind int

const (
    PropString PropertyKind = iota
    PropBool
    PropInt
    PropFloat
    PropEnum     // value comes from Property.Choices
    PropInsets   // 1–4-cell tuple; Padding / Margin
    PropBorder   // theme border name (e.g. "thin", "round")
    PropColor    // theme color reference ("$blue") or hex ("#7aa2f7")
    PropFont
    PropID       // widget id; validated against the live tree for uniqueness
)

// Property describes one editable field on a widget.
type Property struct {
    Name        string             // identifier; lower-snake-case
    Label       string             // human-readable; falls back to Name
    Kind        PropertyKind
    Get         func() any         // current value
    Set         func(v any) error  // validate + apply; returns error on bad input
    Choices     []string           // PropEnum: ordered key list
    ChoiceNames []string           // PropEnum: optional pretty labels parallel to Choices
    Constructor bool               // true: goes in the constructor argument list
    Builder     string             // Builder method name when Constructor==false (default = title-case Name)
    Compose     string             // Compose option name; same default rule
    Help        string             // tooltip / brief docstring
}

// Configurable is the optional interface a widget implements to expose its
// editable properties. The order of the slice is the order shown in the
// editor and the order arguments are emitted in the Constructor argument
// list (for Constructor==true entries).
type Configurable interface {
    Properties() []Property
}
```

### Constructor vs. Builder split

Each widget kind has a fixed constructor signature, e.g.
`NewBox(id, class, title string)`. Properties with `Constructor: true` map to
those positional args, in slice order. Everything else is emitted as a
chained method call on the Builder (or as an `Option` for Compose).

Example for `Box`:

```go
func (b *Box) Properties() []Property {
    return []Property{
        {Name: "id",    Kind: PropID,     Constructor: true, /* Get/Set */},
        {Name: "class", Kind: PropString, Constructor: true},
        {Name: "title", Kind: PropString, Constructor: true,
            Get: func() any { return b.Title },
            Set: func(v any) error { b.Title = v.(string); Redraw(b); return nil }},
        {Name: "border", Kind: PropBorder,
            Builder: "Border", Compose: "Border",
            Get: func() any { return b.Style().Border() },
            Set: func(v any) error { /* update style.border */ }},
        {Name: "padding", Kind: PropInsets,
            Builder: "Padding", Compose: "Padding",
            Get: func() any { return insetsFromStyle(b.Style()) },
            Set: func(v any) error { /* update style padding */ }},
        // …margin, font, etc.
    }
}
```

### Default property descriptors on `Component`

`Component` already owns id, class, hint, padding, margin, font, foreground,
background, border, and flags. These get a default `Properties()`
implementation on `Component` itself:

```go
func (c *Component) Properties() []Property { … }
```

Each concrete widget *appends* kind-specific properties:

```go
func (b *Box) Properties() []Property {
    return append(b.Component.Properties(), Property{Name: "title", …})
}
```

This way every widget that embeds `Component` (which is all of them) is
trivially `Configurable`, even before per-widget refinement.

### Widgets that don't implement `Configurable`

`Configurable` is technically optional. Widgets that don't return useful
properties (custom user widgets, widgets with non-serialisable state like
`Custom` with a render closure) are shown in the tree as read-only:

- The properties form shows the inherited `Component` fields only.
- Code generation emits the constructor with placeholder arguments and a
  `// TODO: <reason>` comment — no chained methods.

---

## Tabs

### 1. Widgets tab

The hierarchy explorer and the designer surface, merged into one tab.

```
┌─ Widgets ──────────────────────────────────────────────────────────────────┐
│ ui > flex#root > grid#body > list#nav                                      │
│ [+Child] [+Sibling] [↑] [↓] [Indent] [Outdent] [Cut] [Copy] [Paste]        │
│ [Delete]                                          [Edit Style]  [Generate ▾]│
├──────────────────┬─────────────────────────────────────────────────────────┤
│  Tree            │  Properties — List "nav"                                │
│  ▼ flex root     │  ──────                                                 │
│    ▼ flex header │  ID         [nav                  ]                     │
│      static …    │  Class      [                     ]                     │
│      button quit │  Hint W     [28                   ]                     │
│    ▼ grid body   │  Hint H     [-1                   ]                     │
│      ▶ list nav  │  Padding    [0 1                  ]                     │
│      viewport …  │  Border     [round              ▼]                      │
│  ─────           │  ──── Styles (widget-local) ──────                       │
│  Info            │  (default)                                              │
│  Type   List     │  list:focused                                           │
│  Bounds 4,2 28,18│  list/highlight                                         │
│  State  :focused │  list/scrollbar                                         │
│  Flags  focused  │  [+ Style]            (click a row to edit in Styles)   │
└──────────────────┴─────────────────────────────────────────────────────────┘
```

#### Tree panel (left)

A `Tree` widget. Each `TreeNode` carries the `Widget` itself as opaque data;
expansion mirrors the parent/child structure. The label is `kind#id` (or
`kind` for empty IDs, with a synthesized label like `kind@N` shown in
parentheses).

Below the tree, a small "Info" block summarises the highlighted widget's
type, bounds, state, and flags — the same info as in the current
`widgetDetails` helper (`inspector.go:198-229`), trimmed to fit.

#### Properties panel (right)

A `FormGroup` whose fields are generated dynamically from the selected
widget's `Properties()`. Each `PropertyKind` maps to a fixed editor:

| Kind        | Editor widget          |
|-------------|------------------------|
| `PropString`, `PropID`, `PropFont` | `Input` |
| `PropInt`, `PropFloat`             | `Input` (numeric validation in `Set`) |
| `PropBool`                         | `Checkbox` |
| `PropEnum`                         | `Select` (Choices/ChoiceNames) |
| `PropInsets`                       | `Input` accepting `1–4` cells (`"0"`, `"0 1"`, `"0 1 0 1"`) |
| `PropBorder`                       | `Select` populated from `theme.Borders()` |
| `PropColor`                        | `ColorPicker` (single mode) |

On every keystroke the field calls `prop.Set(value)`. If `Set` returns an
error, the editor shows an inline red `Static` underneath. Otherwise it
calls `Relayout(widget)` so the change is visible immediately.

Below the kind-specific fields, a "Styles (widget-local)" list shows the
selectors registered on this widget (via the existing `StylesProvider`).
Clicking one switches to the **Styles** tab with that selector pre-loaded.

#### Toolbar

| Button | Hotkey | Action |
|--------|--------|--------|
| `[+ Child]` | `a` | Open widget picker; new widget appended via `container.Add` |
| `[+ Sibling]` | `A` | Same picker, inserted via `container.Insert(idx+1, w)` |
| `[↑]` / `[↓]` | `K`/`J` | Reorder among siblings (`Remove` + `Insert`) |
| `[Indent]` / `[Outdent]` | `>`/`<` | Move into previous sibling / out to grandparent |
| `[Cut]` / `[Copy]` / `[Paste]` | `x`/`c`/`v` | Subtree clipboard (in-process) |
| `[Delete]` | `d` / `Delete` | `parent.Remove(widget)` |
| `[Edit Style]` | `e` | Switch to Styles tab on the widget's default selector |
| `[Generate ▾]` | `Ctrl+G` | Drop-down: Builder / Compose / Both — see §"Code generation" |

#### Add-child flow

`[+ Child]` opens a modal `Dialog` listing all widget kinds known to the
inspector's *kind registry* (see §"Kind registry" below). Selecting a kind:

1. Calls the registry's `New(theme)` factory to build a fresh widget with
   default values for all `Constructor: true` properties.
2. Appends it to the selected container via `container.Add(w)`.
3. Re-runs `Layout()` on the parent.
4. Selects the new widget in the tree and focuses the ID field in the
   properties panel.

If the selected node is not a `Container`, the dialog warns and offers
"insert as sibling" instead.

### 2. Styles tab

The pure editor. It does **not** have a primary data source of its own;
instead it shows whichever style was last selected — either a widget-local
selector picked from the Widgets tab or a theme selector picked from the
Theme tab.

```
┌─ Styles — list:focused (widget-local on list#nav) ────────────────────────┐
│ Scope:  ● Widget overrides    ○ Theme                                     │
│                                                                           │
│  Selector  list:focused                                                   │
│  Fg        [$fg0                ]                                          │
│  Bg        [$blue               ]                                          │
│  Border    [round            ▼]                                            │
│  Font      [bold                ]                                          │
│  Margin    [0                   ]                                          │
│  Padding   [0 2                 ]                                          │
│  Cursor    [—                   ]                                          │
│                                                                           │
│  ┌─ Preview ────────────────────────┐                                     │
│  │   Sample                          │                                     │
│  └───────────────────────────────────┘                                     │
└────────────────────────────────────────────────────────────────────────────┘
```

The fields and behavior are exactly those of the `StyleEditor` in
`doc/spec/style-editor.md`; the Inspector embeds it. The scope radio swaps
the editor's target between:

| Scope | Source | Write target |
|-------|--------|--------------|
| **Widget overrides** | `widget.Style(selector)` | `widget.SetStyle(selector, *Style)` |
| **Theme** | `ui.Theme().Style(selector)` | `theme.SetStyle(selector, *Style)` |

If no widget is currently selected, the **Widget overrides** option is
disabled. The header line above the scope row always shows what's being
edited (selector + scope summary).

`[Save Theme As…]` (visible only in **Theme** scope) opens a `FileChooser`
and writes the live theme back as a Go source file using the existing
theme-codegen helper (`themes/codegen.go`). `[Reset]` reloads the theme from
disk after a confirmation dialog.

### 3. Events tab

A `Table` of `(widget, event, handler)` triplets, plus a side panel showing
the resolved Go function name and (when reachable) a 5-line source snippet.
Unchanged from the previous draft — see §"Events" in the data sources
description below.

### 4. Theme tab

A picker over the active `*Theme`. Two columns:

```
┌─ Theme — Tokyo Night ─────────────────────────────────────────────────────┐
│  Filter: [             ]                                                  │
│ ┌─ Selectors ─────────────┐ ┌─ Computed ────────────────────────────────┐│
│ │ ""                      │ │ Resolved style for: list:focused          ││
│ │ box                     │ │ Foreground   $fg0          (#ffffff)      ││
│ │ box.elevated            │ │ Background   $blue         (#7aa2f7)      ││
│ │ button                  │ │ Border       round                        ││
│ │ button:focused          │ │ Font         bold                         ││
│ │ button:hovered          │ │ Margin       0                            ││
│ │ list                    │ │ Padding      0 2                          ││
│ │ ▶ list:focused          │ │                                           ││
│ │ list/highlight          │ │ [Edit in Styles]                          ││
│ │ …                       │ └───────────────────────────────────────────┘│
│ └─────────────────────────┘                                              │
└────────────────────────────────────────────────────────────────────────────┘
```

- **Selectors list** (left) — every selector registered on the active
  theme, sorted lexicographically, filterable.
- **Computed panel** (right) — the *resolved* style for the highlighted
  selector, including inherited values shown in muted colour and the
  hex-resolved values next to colour variables.
- `[Edit in Styles]` — switches to the Styles tab in **Theme** scope with
  this selector loaded.

The Theme tab is read-only by itself; all editing happens in the Styles
tab. This avoids duplicating the editor surface.

### 5. Log tab

Identical to the previous draft: the `TableLog` rendered as a `Table`,
filtered by level and source, with auto-follow, clear, and export actions.

---

## Mutating the live tree

Because the Inspector edits the host UI directly, the library needs a few
small additions so structural changes are possible without reaching into
unexported fields.

### `Container.Remove`

```go
// Remove unlinks child from this container. Returns ErrNotFound if child is
// not currently a direct child. Implementations call SetParent(nil) on the
// removed child and trigger their own Layout/Refresh.
Remove(child Widget) error
```

### `Container.Insert`

```go
// Insert puts child at index among the existing children. Index 0 prepends;
// index >= len(Children()) appends. Implementations call SetParent and
// trigger Layout/Refresh.
Insert(index int, child Widget) error
```

Both methods get default implementations on every container kind. Where the
container has structural constraints (e.g. `Box` accepts a single child,
`Card` accepts at most two), `Insert` returns `ErrFull` instead of pushing
beyond capacity.

### Property-set side effects

Every `Property.Set` is responsible for:

1. Validating the input and returning an error on failure.
2. Applying the change (e.g. `b.Title = v.(string)`).
3. Calling `Redraw(widget)` for visual-only changes or `Relayout(widget)`
   for changes that affect size (padding, margin, hint, etc.).

The inspector itself does not need to know which kind of change happened —
it just calls `Set` and trusts the widget to refresh.

---

## Kind registry

A small registry that maps widget kind names → (default factory, property
template). Used by the Add-child dialog and by code generation to find the
constructor signature.

```go
type KindEntry struct {
    Name    string                          // "Box", "Static"
    Group   string                          // "container", "leaf", "input", "display", "animated"
    New     func(theme *Theme) Widget       // fresh instance with sensible defaults
    Help    string                          // one-line description for the picker
}

// Registered kinds. Populated in inspector/kinds.go via init() functions
// that import each widget package.
var Kinds = map[string]KindEntry{}

func RegisterKind(e KindEntry) { … }
```

Each widget contributes its own entry in an `init()` block:

```go
func init() {
    inspector.RegisterKind(inspector.KindEntry{
        Name:  "Box",
        Group: "container",
        New:   func(t *Theme) Widget { return NewBox("", "", "") },
        Help:  "Bordered container with title; holds one child",
    })
}
```

---

## Code generation

Walks the live widget tree depth-first. For each widget:

1. Look up the kind entry in `Kinds`.
2. Call `widget.Properties()` (if `Configurable`); otherwise produce a
   `// TODO` placeholder.
3. Split the slice by `Constructor: true` (positional args) vs. the rest
   (chained method calls / Compose options).
4. Emit:
   - **Builder** — `Kind(arg1, arg2, …).Method1(…).Method2(…).`
     Containers recurse into children, then close with `.End()`.
   - **Compose** — `compose.Kind(arg1, arg2, …, Method1(…), Method2(…),
     children…)`.

### Argument formatting

Each `PropertyKind` has a fixed Go-literal formatter:

| Kind         | Builder snippet | Compose snippet |
|--------------|-----------------|-----------------|
| `PropString` | `"hello"`       | `"hello"`       |
| `PropInt`    | `42`            | `42`            |
| `PropBool`   | `true`          | `true`          |
| `PropEnum`   | unquoted ident if it matches a Go const (e.g. `core.Stretch`); otherwise quoted | same |
| `PropInsets` | `1, 2, 3, 4` (variadic) | `compose.Padding(1, 2, 3, 4)` |
| `PropColor`  | `"$blue"`       | `"$blue"`       |
| `PropBorder` | `"round"`       | `"round"`       |

Properties that match the kind's default value are omitted from the output
to keep the generated source compact.

### Event handlers

Handlers cannot be round-tripped (function bodies aren't serializable).
Generated code emits a comment with the resolved Go function name:

```go
builder.Button("quit", "Quit").
    // TODO: On(EvtActivate, main.quitFn)
    End().
```

The function name is resolved via the same `runtime.FuncForPC` /
`reflect.ValueOf(h).Pointer()` mechanism used by the Events tab.

### Render functions and table providers

Widgets that take a closure (`Deck`, `Tiles`) or a provider interface
(`Table`) emit the constructor with `nil` and a `// TODO` comment for the
missing argument. The Inspector cannot reproduce those.

### Output target

`[Generate ▾]` opens a small dropdown:

| Option | Action |
|--------|--------|
| Builder → file | `FileChooser` save dialog → write `.go` source |
| Builder → clipboard | OSC-52 (or fallback `$DISPLAY` clipboard tools) |
| Compose → file | as above with Compose-style output |
| Compose → clipboard | as above |
| Both → file | dual-import file with a comment separator |

The "preview / code panel" of the previous designer-tab draft is gone — the
output is delivered directly to a file or the clipboard.

---

## Public API

```go
type Inspector struct {
    Component
    ui       *UI
    target   Container
    selected Widget
    mode     InspectorMode
    layout   Container
}

type InspectorMode int

const (
    ModeAttached InspectorMode = iota
    ModeReadonly
)

// NewInspector creates an Inspector attached to the given root.
func NewInspector(root Container) *Inspector

// SetMode forces a specific mode (e.g. ModeReadonly) regardless of detection.
func (i *Inspector) SetMode(m InspectorMode)

// Toggle shows / hides the Inspector popup.
func (i *Inspector) Toggle()

// UI returns the inner builder container.
func (i *Inspector) UI() Container

// Refresh re-scans the target tree.
func (i *Inspector) Refresh()
```

### Builder integration

```go
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
| `EvtChange` | `*Style` or `Property` | A style or property was edited |
| `EvtActivate` | `Widget` | Container was navigated into |
| `EvtClose` | — | Inspector popup is closing |
| `EvtMode` | `InspectorMode` | Mode changed |

---

## Keyboard map

Global (Inspector has focus):

| Key | Action |
|-----|--------|
| `1`–`5` | Switch to tab Widgets / Styles / Events / Theme / Log |
| `Tab` / `Shift+Tab` | Cycle focus inside the active tab |
| `Esc` | Close Inspector popup |
| `Ctrl+G` | Generate from current root, save to file |
| `Ctrl+E` | Switch to Styles tab on the current widget |
| `?` | Show keybindings cheat sheet |

Widgets-tab specific:

| Key | Action |
|-----|--------|
| `↑` / `↓` / `j` / `k` | Move highlight |
| `Enter` / `Space` | Expand / collapse container |
| `a` / `A` | Add child / sibling |
| `d` / `Delete` | Delete |
| `J` / `K` | Move down / up among siblings |
| `>` / `<` | Indent / outdent |
| `x` / `c` / `v` | Cut / copy / paste subtree |
| `e` | Edit default style of selected widget |

Theme-tab specific:

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate selectors |
| `/` | Focus filter |
| `Enter` | Open in Styles tab |

---

## Required additions to the library

The Inspector intentionally introduces only small, generic extensions, keeping
inspector-specific concerns out of the core.

### 1. `Configurable`

The new interface in §"Configurable model". Default implementation lives on
`Component`; concrete widgets append kind-specific entries.

### 2. `Container.Remove(child Widget) error`

Counterpart to `Add`. Each container implementation already tracks children,
so this is a few lines per kind plus an `ErrNotFound` sentinel.

### 3. `Container.Insert(index int, child Widget) error`

Index-based child insertion. `ErrFull` for fixed-arity containers.

### 4. `Component.Handlers() map[Event][]Handler`

Already in the previous draft; still required.

### 5. `StylesProvider`

```go
type StylesProvider interface {
    Styles() []string
    Style(selector string) *Style
}
```

Already in the previous draft; still required.

### 6. `(*TableLog).Reset()`

Empties the circular buffer.

### 7. `(*UI).Inspector() *Inspector`

Cached singleton accessor; the existing `Ctrl+D` handler in `ui.go` calls
`ui.Inspector().Toggle()`.

### 8. Kind registry (`Kinds map` + `RegisterKind`)

In a new `inspector` package or in `widgets/kinds.go`. Each widget's `init`
function registers itself.

---

## Implementation plan

```
inspector/
├── inspector.go     — Inspector struct, mode detection, Toggle, public API
├── widgets-tab.go   — Tree explorer + properties form + structural actions
├── styles-tab.go    — Embeds StyleEditor; scope-driven
├── events-tab.go    — Handler table + source-snippet panel
├── theme-tab.go     — Theme selector picker + computed-style panel
├── log-tab.go       — TableLog viewer with filters
├── codegen.go       — Builder + Compose code emitters; walks live tree
├── kinds.go         — KindEntry, RegisterKind, Kinds map, defaults
├── handler-meta.go  — runtime.FuncForPC + closure-name unwrapping helpers
└── doc.go
```

### Step-by-step

1. **Library additions** — `Configurable`, `Property`, `PropertyKind`;
   `Container.Remove` / `Insert`; `Component.Handlers` and
   `Component.Properties`; `StylesProvider`; `(*TableLog).Reset`;
   `(*UI).Inspector`.

2. **Default `Component.Properties()`** — covers id, class, hint, padding,
   margin, font, fg/bg, border, focusable/visible flags. ~50 LOC.

3. **Per-widget `Properties()` overrides** — append kind-specific entries
   (`Box.Title`, `Flex.Horizontal/Alignment/Spacing`, `Grid.Rows/Cols`,
   `Button.Text`, `Static.Text`, `Input.Placeholder/Mask`, etc.). ~10–15
   LOC per widget × ~50 widgets ≈ 600 LOC total but trivially mechanical.

4. **Kind registry** — one `init()` block per widget. Mostly mechanical.

5. **Code generation (`codegen.go`)** — single tree walk, calls
   `Properties()`, splits on `Constructor`, emits with the
   formatter table from §"Argument formatting". Separate Builder and
   Compose emitters share a helper that does the property split.

6. **Widgets tab** — Tree from `core.Traverse`; properties form from
   `Properties()` with the editor-kind table; toolbar buttons drive the
   structural mutations through `Container.Add` / `Insert` / `Remove`.

7. **Styles tab** — wraps `StyleEditor` (already specified) with a scope
   selector that switches its `*Theme` between the live theme and a
   per-widget synthetic theme.

8. **Theme tab** — selector list + computed-style panel.
   `[Edit in Styles]` switches tab and pre-loads the editor.

9. **Events tab** — unchanged from previous draft.

10. **Log tab** — unchanged from previous draft.

11. **Stand-alone binary (`cmd/inspector`)** — boots a fresh `*UI` whose
    root is an empty `Flex`, opens the Inspector popup at startup. Flags:
    `--theme`, `--snapshot <dump.json>` (READONLY), `--out <file.go>`
    (write generated code on quit).

12. **Theme entries** — add the inspector-specific selectors listed below
    to every `themes/theme-*.go`.

13. **Tests** —
    - `inspector/codegen_test.go`: round-trip tree → Builder code → re-parse
      with `go/parser` for syntactic validity; spot-check generated code
      for representative trees.
    - `inspector/kinds_test.go`: every registered kind has a working `New`
      factory whose result implements `Configurable` with at least the
      `Constructor: true` properties needed to reconstruct it.
    - `inspector/properties_test.go`: per-widget `Properties()` round-trip
      (Get → format → parse → Set returns same value).
    - `inspector/handler-meta_test.go`: function-name resolution for
      top-level fns, methods, anonymous closures, typed-wrapper closures.

---

## Styling selectors

Inspector-specific selectors registered in every theme:

| Selector | Applies to |
|----------|-----------|
| `inspector` | Outer container |
| `inspector/box` | Title-bordered box |
| `inspector/tabs` | Tab strip |
| `inspector/breadcrumb` | Path display |
| `inspector/tree` | Tree explorer rows |
| `inspector/properties` | Properties form group |
| `inspector/properties:label` | Property labels |
| `inspector/info` | Info summary panel |
| `inspector/locate-flash` | Bright accent overlay drawn during `[Locate]` |
| `inspector/event-table` | Event-handler table |
| `inspector/source` | Source-snippet panel |
| `inspector/theme-list` | Theme selectors list |

The Inspector adds no new widget *types* — every cell is an existing
primitive — so themes only need the selector additions.

---

## Mutation contract (ATTACHED mode)

Mutations only happen in response to explicit user action:

| Action | Mutation |
|--------|----------|
| Property edit | `prop.Set(v)`; widget calls `Redraw` / `Relayout` |
| Style edit (widget overrides) | `widget.SetStyle(sel, *Style)`; `Relayout(widget)` |
| Style edit (theme) | `theme.SetStyle(sel, *Style)`; `Refresh()` on root |
| `[+ Child]` | `container.Add(child)`; `Relayout(container)` |
| `[+ Sibling]` | `container.Insert(idx+1, child)`; `Relayout(container)` |
| `[Delete]` | `parent.Remove(child)`; `Relayout(parent)`; refocus parent if focus was on child |
| `[↑]` / `[↓]` | `parent.Remove(child)` + `parent.Insert(newIdx, child)` |
| `[Indent]` | move child into previous sibling (must be a container) |
| `[Outdent]` | move child into grandparent at parent's position |

All mutations log at `slog.Info` with `source=inspector`. Failures
(`Container.Insert` returns `ErrFull`, e.g.) surface as a red `Notification`
toast for 3 seconds and a `slog.Warn` entry.

---

## Non-goals

- **Undo / redo** of inspector edits. The simplest path forward is to
  generate code, save to disk, and re-run; full in-memory undo is a future
  enhancement.
- **Cross-process inspection.** The Inspector lives in the host process and
  reads in-memory state. Remote inspection over a socket is a future
  enhancement; READONLY mode operating on a serialised snapshot is the seam.
- **Source modification.** The Inspector emits Go code to a file or the
  clipboard; it never edits the user's existing source.
- **Hot reload.** Generated code is non-executing. Re-running the binary
  picks up changes; the Inspector itself is the short-feedback path.
- **Round-tripping closures.** Event handlers, render functions, and table
  providers are emitted as `// TODO` stubs.

# Inspector

A unified developer console for a running `*UI`. Five tabs over a live widget
tree: **Widgets** (the explorer + designer), **Styles** (the context-driven
style editor), **Events** (handler map with function-name introspection),
**Theme** (the theme-style picker), **Log** (the in-memory log table).

The Inspector mutates the *live* widget tree directly. There is no separate
`Document` or `DesignerNode` model — every change touches the actual `*UI`
the Inspector is attached to. Editing and code generation both go through
**form structs** that live next to each widget — see §"WidgetForm model".

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

## WidgetForm model

Every editable widget has a sibling `*Form` struct that owns:

- the editing surface (a plain Go struct with reflection-friendly tags that
  the Form widget renders via the existing `BuildFormGroup`),
- the load / store between widget and form,
- the codegen `Emit` that writes the widget's call shape onto an
  in-progress chain.

Forms live in the `widgets` package, alongside the widgets they edit, so
they can read and write unexported fields directly without exposing them
through accessors. The inspector never touches widget internals; it only
talks to forms.

### Three interfaces

```go
// inspector/widget-form.go
package inspector

type WidgetForm interface {
    Name() string                  // "Static", "Grid"
    Group() string                 // "leaf", "container", "input", "display", "animated"
    Help() string                  // one-line tooltip

    New() core.Widget              // fresh instance from current form fields
    Load(core.Widget)              // copy widget state into form fields
    Store(core.Widget)             // copy form fields back into widget

    // Validate runs per-field validation. field=="" means "validate the
    // whole form including cross-field rules"; returns the first failure
    // encountered. The property panel calls per-field on change and
    // whole-form on Apply.
    Validate(field string) error

    // Emit writes the widget's call shape onto an in-progress chain. The
    // caller has already written whatever precedes this widget; an Emit
    // implementation continues with leading ".\n" so the chain stays
    // continuous. Containers emit only the constructor + chain; the
    // codegen walker writes children and the closing ".End()".
    Emit(w io.Writer, mode string) error
}

// ContainerForm extends WidgetForm for containers whose Add method takes
// per-child layout parameters. Containers that ignore Add params (Box,
// Flex, Card, …) implement only WidgetForm.
type ContainerForm interface {
    WidgetForm
    LayoutForm(parent core.Container, child core.Widget) core.LayoutForm
}
```

```go
// core/widget-form.go — LayoutForm lives in core so widget container
// forms can declare it as a return type without importing inspector.
package core

type LayoutForm interface {
    Load(parent Container, child Widget)
    Store(parent Container, child Widget)
    Validate(field string) error
    Emit(w io.Writer, mode string) error
}
```

### Codegen modes

Two string constants in inspector. Mode is passed through Emit so future
back-ends can be added without changing signatures.

```go
const (
    ModeBuilder = "builder"
    ModeCompose = "compose"
)
```

ModeCompose is reserved; current implementation produces an error if used.

### Indentation: gofmt does it

The walker writes `.\n` separators only — no tab tracking. `go/format` is
run on the final output, which normalises chained method calls to a
single tab past the leading expression. Visual nesting is conveyed by a
trailing comment on each container's closing `End()`, e.g. `End() // Grid#g1`.
The comment survives `gofmt` and gives the reader a structural anchor.

### ComponentForm — the embedded base

Every per-widget form embeds `ComponentForm`, which mirrors `Component`'s
role as the embedded base of every widget. ComponentForm covers the fields
shared across every kind:

```go
// widgets/component-form.go
type ComponentForm struct {
    // Editing surface (struct-tagged → BuildFormGroup picks them up).
    ID    string `group:"general" label:"ID" validate:"id-unique"`
    Class string `group:"general" label:"Class"`

    HintW int `group:"layout" label:"Hint W"`
    HintH int `group:"layout" label:"Hint H"`

    Skip     bool `group:"flags" label:"Skip Focus"` // FlagSkip — opts out of focus traversal
    Hidden   bool `group:"flags" label:"Hidden"`
    Disabled bool `group:"flags" label:"Disabled"`

    // Codegen snapshot (no struct tag → invisible to the property panel).
    // The Styles tab edits the live *core.Style directly via core.StyleForm;
    // ComponentForm carries this snapshot only so codegen can emit the
    // widget's style chain alongside the constructor.
    style core.StyleForm
}
```

`FlagFocusable` is *not* exposed: focusability is an inherent attribute of
each widget kind set by its constructor. Users opt out of focus traversal
by setting `FlagSkip` instead, which is what the form exposes.

### Primitives + template method

ComponentForm exposes building blocks plus a template method that
captures the standard emission shape:

```go
// Primitives — public so per-widget forms can compose them directly.
func (f *ComponentForm) CheckBuilderMode(mode string) error
func (f *ComponentForm) EmitClassPrefix(w io.Writer)        // .Class("foo") if non-empty
func (f *ComponentForm) EmitChain(w io.Writer)              // Hint, Skip, Hidden, Disabled
func (f *ComponentForm) EmitStyle(w io.Writer)              // delegates to f.style.EmitBuilderChain

// Template method — the standard order. Per-form Emit shrinks to one call
// + a body callback for the constructor.
func (f *ComponentForm) EmitFrame(w io.Writer, mode string, body func() error) error {
    if err := f.CheckBuilderMode(mode); err != nil { return err }
    f.EmitClassPrefix(w)
    if err := body(); err != nil { return err }
    f.EmitChain(w)
    f.EmitStyle(w)
    return nil
}
```

A trivial form is one expression:

```go
func (f *StaticForm) Emit(w io.Writer, mode string) error {
    return f.EmitFrame(w, mode, func() error {
        _, err := fmt.Fprintf(w, ".\nStatic(%q, %q)", f.ID, f.Text)
        return err
    })
}
```

A form with a kind-specific tail (e.g. Grid's `.Rows(...).Columns(...)`)
calls `EmitFrame` and then appends after it returns.

A form that needs to override the standard shape (e.g. Flex picking
`HFlex` vs `VFlex` from a flag) calls the primitives directly.

### StyleForm and the fixed-style rule

`core.StyleForm` is the editor + codegen surface for `*Style`. It also
holds an unexported `fixed bool` snapshot taken at Load time:

```go
type StyleForm struct {
    Selector   string `group:"selector" label:"Selector" readonly:""`

    Foreground string `group:"colors" label:"Fg" control:"color"`
    Background string `group:"colors" label:"Bg" control:"color"`

    Border  string `group:"box" label:"Border"  control:"border"`
    Padding [4]int `group:"box" label:"Padding" control:"insets"`
    Margin  [4]int `group:"box" label:"Margin"  control:"insets"`

    Font   string `group:"text" label:"Font" control:"font"`
    Cursor string `group:"text" label:"Cursor"`

    Shadow string `group:"effects" label:"Shadow"`

    fixed bool // set by Load; suppresses EmitBuilderChain when true
}
```

`EmitBuilderChain` short-circuits when `fixed` is true: themed styles are
inherited from the active theme, so emitting them in generated source
would override theme changes and bloat the output. Only widget-specific
overrides (the non-fixed leaf in the cascade) get emitted.

The Styles tab edits the *live* `*Style` of the selected widget through
its own StyleForm instance, parallel to whatever ComponentForm has loaded
internally. Edits go through `Modifiable()`, so editing a fixed style
creates a new non-fixed child that future codegen will then emit.

### Per-widget forms

```go
// widgets/static-form.go
type StaticForm struct {
    ComponentForm
    Text      string `group:"general" label:"Text"`
    Alignment string `group:"general" label:"Alignment" control:"select" options:"left,center,right"`
}

// widgets/flex-form.go
type FlexForm struct {
    ComponentForm
    Vertical  bool   `group:"layout" label:"Vertical"`
    Alignment string `group:"layout" label:"Alignment" control:"select" options:"start,center,end,stretch,left,right"`
    Spacing   int    `group:"layout" label:"Spacing"`
}

// widgets/input-form.go
type InputForm struct {
    ComponentForm
    Text        string `group:"value" label:"Text"`
    Placeholder string `group:"value" label:"Placeholder"`
    Mask        string `group:"value" label:"Mask"`
    Max         int    `group:"value" label:"Max Length"`
    Masked      bool   `group:"flags" label:"Masked"`
    Readonly    bool   `group:"flags" label:"Read-only"`
}

// widgets/grid-form.go — also implements ContainerForm via its
// LayoutForm method, returning a GridLayoutForm per child.
type GridForm struct {
    ComponentForm
    Rows    []int `group:"layout" label:"Rows"`
    Columns []int `group:"layout" label:"Columns"`
    Lines   bool  `group:"layout" label:"Lines"`
}

type GridLayoutForm struct {
    X int `group:"position" label:"Column"`
    Y int `group:"position" label:"Row"`
    W int `group:"position" label:"Col Span"`
    H int `group:"position" label:"Row Span"`
}
```

### Form (the widget) is special

`Form` (the FormGroup-backed widget) is itself a widget kind, but it
edits arbitrary user data via reflection on a struct passed at
construction. A separate "form designer" handles that case; the standard
WidgetForm shape doesn't fit.

---

## Designer

The **Designer** is the inspector's central hub. It owns the kind
registry, exposes lookup operations, drives tree edits, and runs codegen.
Widget *Form structs satisfy `WidgetForm` / `ContainerForm` purely
structurally; the Designer wires them up.

### Kind table

```go
// inspector/kinds.go
type Kind struct {
    Name  string                      // "Static", "Grid"
    Group string                      // "leaf", "container", "input", "display", "animated"
    Help  string                      // one-line description for the picker
    Type  reflect.Type                // concrete widget pointer type, e.g. (*widgets.Static)(nil)
    Make  func() WidgetForm           // factory for a fresh form instance
}
```

Registration validates the factory at Register time: the Designer calls
`Make()` once and asserts that the resulting form's Load accepts a
zero-value of `Kind.Type`. A mis-registration (wrong form for a kind)
fails immediately rather than panicking later.

### Operations

```go
type Designer struct { ... }

func NewDesigner(target core.Container) *Designer

// Registry — one Kind per widget type.
func (d *Designer) Register(k Kind)
func (d *Designer) Kinds() []Kind            // for the Add-child picker
func (d *Designer) Kind(w core.Widget) Kind  // unloaded; for capability checks
func (d *Designer) FormFor(w core.Widget) WidgetForm  // loaded; for editing

// Tree edits.
func (d *Designer) Add(parent core.Container, kind Kind) (core.Widget, error)
func (d *Designer) Remove(child core.Widget) error
func (d *Designer) Move(child core.Widget, newParent core.Container, pos int) error
func (d *Designer) SetField(widget core.Widget, fieldName string, value any) error

// Codegen.
func (d *Designer) GenerateFragment(mode string, w io.Writer) error
func (d *Designer) GenerateFile(mode string, w io.Writer, pkg, funcName string) error
```

### Concurrency

Designer is **not** safe for concurrent use; callers serialise access.
The TUI's main loop is the natural serialiser. Background codegen would
need a snapshot of the tree first.

### Why split `FormFor` from `Kind`

`FormFor(w)` allocates and Loads. `Kind(w)` returns the unloaded factory
descriptor. The codegen walker uses `Kind` for capability checks (does
this widget have a ContainerForm?) so it doesn't pay for an unused
parent.Load on every child.

### Driver-side registration

The widgets package never imports inspector. Registrations happen in the
driver:

```go
// cmd/inspector-poc/main.go
d := inspector.NewDesigner(root)
d.Register(inspector.Kind{
    Name: "Static", Group: "leaf",
    Help: "Non-interactive text label",
    Type: reflect.TypeOf((*widgets.Static)(nil)),
    Make: func() inspector.WidgetForm { return &widgets.StaticForm{} },
})
// ... one call per widget kind
```

A mechanical `init()`-based registration scheme could be layered on top
later, but the current approach keeps the dependency direction clean
(`widgets ⟂ inspector`).

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
│      ▶ list nav  │  Skip       [ ]                                         │
│      viewport …  │  Hidden     [ ]                                         │
│  ─────           │  ──── Layout (in parent grid#body) ───                  │
│  Info            │  Column     [0                    ]                     │
│  Type   List     │  Row        [1                    ]                     │
│  Bounds 4,2 28,18│  Col Span   [1                    ]                     │
│  State  :focused │  Row Span   [1                    ]                     │
│  Flags  focused  │  ──── Styles (widget-local) ──────                      │
│                  │  (default)                                              │
│                  │  list:focused                                           │
│                  │  list/highlight                                         │
│                  │  [+ Style]            (click a row to edit in Styles)   │
└──────────────────┴─────────────────────────────────────────────────────────┘
```

#### Tree panel (left)

A `Tree` widget. Each `TreeNode` carries the `Widget` itself as opaque
data; expansion mirrors the parent/child structure. The label is
`kind#id` (or `kind` for empty IDs, with a synthesized label like
`kind@N` shown in parentheses).

Below the tree, a small "Info" block summarises the highlighted widget's
type, bounds, state, and flags.

#### Properties panel (right)

Built dynamically from the selected widget's `WidgetForm` via the
existing `BuildFormGroup` helper. Each tagged form field becomes an
editor:

| Tag (`control:"…"`) | Editor widget |
|---|---|
| (default for `string`) | `Input` |
| (default for `int`/`float`) | `Input` with numeric validation |
| (default for `bool`) | `Checkbox` |
| `select` (with `options:"a,b,c"`) | `Select` |
| `color` | `ColorPicker` (single mode) |
| `border` | `Select` populated from `theme.Borders()` |
| `insets` | `Input` accepting 1–4 cells (`"0"`, `"0 1"`, `"0 1 0 1"`) |
| `font` | `Select` populated from theme's font registry |

`group:"…"` clusters fields under a section heading; `label:"…"` sets the
display label; `readonly:""` makes the field read-only.

On every keystroke the editor calls `form.Validate(field)`. If validation
fails, the editor shows an inline red `Static` underneath; the widget is
not yet updated. On focus-out (or Apply), `form.Store(widget)` writes
back, and the inspector calls `Relayout(widget)`.

Below the kind-specific fields, **Layout (in parent X)** appears when
the parent container has a `ContainerForm`. The fields come from
`parent.LayoutForm(parent, widget)` — e.g. Grid's cell coordinates.

Below that, a **Styles (widget-local)** list shows the selectors
registered on this widget (via `StylesProvider`). Clicking one switches
to the **Styles** tab with that selector pre-loaded.

#### Toolbar

| Button | Hotkey | Action (via Designer) |
|--------|--------|------------------------|
| `[+ Child]` | `a` | `d.Add(selected, kind)` |
| `[+ Sibling]` | `A` | `d.Add(selected.Parent, kind)` then `d.Move(child, ..., idx+1)` |
| `[↑]` / `[↓]` | `K`/`J` | `d.Move(child, parent, newIdx)` |
| `[Indent]` / `[Outdent]` | `>`/`<` | `d.Move(child, prevSibling, end)` / `d.Move(child, grandparent, parentIdx)` |
| `[Cut]` / `[Copy]` / `[Paste]` | `x`/`c`/`v` | Subtree clipboard (in-process) |
| `[Delete]` | `d` / `Delete` | `d.Remove(child)` |
| `[Edit Style]` | `e` | Switch to Styles tab on the widget's default selector |
| `[Generate ▾]` | `Ctrl+G` | Drop-down: Builder fragment / Builder file / Compose / Both — see §"Code generation" |

#### Add-child flow

`[+ Child]` opens a modal `Dialog` listing all kinds from
`d.Kinds()`, grouped by `Kind.Group`. Selecting a kind:

1. Calls `d.Add(selectedContainer, kind)` which:
   - calls `kind.Make()` for a fresh form,
   - calls `form.New()` for a default-valued widget,
   - calls `parent.Add(child)` (with whatever extra params the
     ContainerForm requires for the default child position).
2. Re-runs `Layout()` on the parent.
3. Selects the new widget in the tree and focuses the ID field in the
   properties panel.

If the selected node is not a `Container`, the dialog warns and offers
"insert as sibling" instead.

### 2. Styles tab

The pure editor. It does **not** have a primary data source of its own;
instead it shows whichever style was last selected — either a widget-local
selector picked from the Widgets tab or a theme selector picked from the
Theme tab.

The fields are exactly those of `core.StyleForm`. The scope radio swaps
the editor's target between:

| Scope | Source | Write target |
|-------|--------|--------------|
| **Widget overrides** | `widget.Style(selector)` | `widget.SetStyle(selector, *Style)` |
| **Theme** | `ui.Theme().Style(selector)` | `theme.SetStyle(selector, *Style)` |

Edits go through `*Style.Modifiable()` so editing a fixed style produces
a new non-fixed child rather than mutating the theme by accident.

If no widget is currently selected, the **Widget overrides** option is
disabled. The header line above the scope row always shows what's being
edited (selector + scope summary).

`[Save Theme As…]` (visible only in **Theme** scope) opens a `FileChooser`
and writes the live theme back as a Go source file using the existing
theme-codegen helper (`themes/codegen.go`). `[Reset]` reloads the theme
from disk after a confirmation dialog.

### 3. Events tab

A `Table` of `(widget, event, handler)` triplets, plus a side panel
showing the resolved Go function name and (when reachable) a 5-line
source snippet.

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

The `TableLog` rendered as a `Table`, filtered by level and source, with
auto-follow, clear, and export actions.

---

## Code generation

The Designer's `GenerateFragment` and `GenerateFile` walk the live tree
depth-first and assemble a chained Builder (or Compose) expression.

### Walker

For each widget the walker:

1. **Layout prefix.** If the parent has a `ContainerForm`, calls
   `parent.LayoutForm(parent, widget).Emit(w, mode)`. This writes things
   like `.Cell(0, 1, 1, 1)` for Grid.
2. **Constructor + chain.** Calls `widget's WidgetForm.Emit(w, mode)`.
   For a form using `EmitFrame`, this writes:
   - class prefix (`.Class("foo")` if non-empty)
   - the kind's constructor (`Static("id", "Hello")`)
   - the standard ComponentForm chain (`Hint`, `Skip`, `Hidden`,
     `Disabled`)
   - the style chain (only widget-specific overrides; themed styles
     suppressed by `StyleForm.fixed`)
3. **Children.** If the widget is a `Container`, recurse into each
   child.
4. **Close.** Emit `.End() // Kind#ID` where Kind comes from
   `WidgetForm.Name()` and ID from the widget's id.

The walker does no indent tracking. Output is `.\n`-separated; trailing
End-comments preserve nesting visually after gofmt flattens the chain.

### Output shapes

```go
func (d *Designer) GenerateFragment(mode string, w io.Writer) error
func (d *Designer) GenerateFile(mode string, w io.Writer, pkg, funcName string) error
```

`GenerateFragment` writes a chained expression starting with
`NewBuilder(theme)` and ending without a trailing newline. Useful for
tests and for piping into bigger generators.

`GenerateFile` wraps the fragment with `package`, imports (computed from
the kinds touched during the walk), and a `func funcName(theme *Theme)
*UI { return … .Build() }` body. The output is a complete, compilable
Go file.

Both run `go/format.Source` on the result before returning, so callers
always see canonical Go.

### Compose mode

`mode == ModeCompose` returns a "not implemented" error today; the
walker's structure is designed so a compose back-end can be added by
implementing `WidgetForm.Emit(..., ModeCompose)` per kind without
walker changes.

### Event handlers

Handlers cannot be round-tripped (function bodies aren't serializable).
Generated code emits a comment with the resolved Go function name:

```go
builder.Button("quit", "Quit").
    // TODO: On(EvtActivate, main.quitFn)
    End()
```

The function name is resolved via the same `runtime.FuncForPC` /
`reflect.ValueOf(h).Pointer()` mechanism used by the Events tab.

### Render functions and table providers

Widgets that take a closure (`Deck`, `Tiles`) or a provider interface
(`Table`) emit the constructor with `nil` and a `// TODO` comment for
the missing argument.

### Output target

`[Generate ▾]` opens a small dropdown:

| Option | Method |
|--------|--------|
| Builder fragment → clipboard | `GenerateFragment(ModeBuilder, clipboard)` |
| Builder → file | `GenerateFile(ModeBuilder, file, pkg, fn)` |
| Compose → file | `GenerateFile(ModeCompose, file, pkg, fn)` (when implemented) |
| Both → file | dual files with a comment separator |

---

## Public API

```go
type Inspector struct {
    Component
    ui       *UI
    target   Container
    selected Widget
    mode     InspectorMode
    designer *Designer
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

The Inspector intentionally introduces only small, generic extensions,
keeping inspector-specific concerns out of the core.

### 1. `core.LayoutForm` interface

Defined in core because widget container forms declare it as a return
type. See §"WidgetForm model".

### 2. `core.StyleForm`

Already implemented. The `fixed` snapshot field is added so
`EmitBuilderChain` short-circuits on themed styles.

### 3. ComponentForm + per-widget *Form structs

Live in the widgets package. ComponentForm exposes primitives plus the
`EmitFrame` template method. Each concrete widget gets a sibling
`*Form` struct (~50 forms total, mostly mechanical copies of the
established pattern).

### 4. `Container.Remove(child Widget) error`

Counterpart to `Add`. Each container implementation already tracks
children, so this is a few lines per kind plus an `ErrNotFound`
sentinel.

### 5. `Container.Insert(index int, child Widget, params ...any) error`

Index-based child insertion. `ErrFull` for fixed-arity containers.
Variadic `params` mirrors `Add` for containers (Grid) that need
per-child layout arguments.

### 6. `Component.Handlers() map[Event][]Handler`

Already in the previous draft; still required for the Events tab.

### 7. `StylesProvider`

```go
type StylesProvider interface {
    Styles() []string
    Style(selector string) *Style
}
```

### 8. `(*TableLog).Reset()`

Empties the circular buffer.

### 9. `(*UI).Inspector() *Inspector`

Cached singleton accessor; the existing `Ctrl+D` handler in `ui.go`
calls `ui.Inspector().Toggle()`.

### 10. Form-widget control types

The Form widget's `BuildFormGroup` already handles `string`, `int`,
`bool`, and `select`. The inspector needs `color`, `border`, `insets`,
and `font` controls — small additions to the form-control switch.

---

## Implementation plan

```
inspector/
├── inspector.go       — Inspector struct, mode detection, Toggle, public API
├── designer.go        — Designer struct, kind table, tree edits, codegen
├── widget-form.go     — WidgetForm, ContainerForm, ModeBuilder/Compose
├── kinds.go           — Kind struct + Designer.Register validation helpers
├── widgets-tab.go     — Tree explorer + properties form + structural actions
├── styles-tab.go      — Embeds StyleEditor over core.StyleForm; scope-driven
├── events-tab.go      — Handler table + source-snippet panel
├── theme-tab.go       — Theme selector picker + computed-style panel
├── log-tab.go         — TableLog viewer with filters
├── handler-meta.go    — runtime.FuncForPC + closure-name unwrapping helpers
└── doc.go

widgets/
├── component-form.go  — ComponentForm + primitives + EmitFrame
├── static-form.go
├── flex-form.go
├── grid-form.go       — GridForm + GridLayoutForm
├── input-form.go
├── … per-widget forms
└── form-designer.go   — Special handling for the Form widget

core/
├── style-form.go      — StyleForm with fixed snapshot
└── widget-form.go     — LayoutForm interface
```

### Step-by-step

1. **Library prerequisites** — `Container.Remove`, `Container.Insert`,
   `Component.Handlers`, `StylesProvider`, `(*TableLog).Reset`,
   `(*UI).Inspector`. ~1 day.

2. **Form-widget control types** — extend `BuildFormGroup` /
   `buildFormControl` to support `color`, `border`, `insets`, `font`.

3. **WidgetForm interfaces + Designer skeleton** — `inspector/widget-form.go`,
   `inspector/designer.go`, `inspector/kinds.go`. Implement Register
   validation, Kind/FormFor lookup.

4. **ComponentForm** — primitives + `EmitFrame`. Add `style` snapshot field;
   wire Load/Store through `*Style.Modifiable()`.

5. **core.StyleForm.fixed** — add the snapshot field; short-circuit
   `EmitBuilderChain`.

6. **Per-widget Form structs** — ~50 forms, mostly mechanical copies of
   the StaticForm pattern.

7. **Designer tree-edit operations** — `Add`, `Remove`, `Move`,
   `SetField`. Each calls into the form, mutates the tree, relayouts,
   and logs.

8. **Codegen walker + GenerateFragment/GenerateFile** — including
   `go/format` post-processing and trailing `End() // Kind#ID` comments.

9. **Round-trip test** — table-driven: canned trees → `GenerateFragment`
   → `go/parser` → walk AST and rebuild a `*Form` tree → compare to
   the original via Load. Catches field-level codegen bugs.

10. **Inspector tabs** — Widgets, Styles, Events, Theme, Log.

11. **Stand-alone binary (`cmd/inspector`)** — boots a fresh `*UI` whose
    root is an empty `Flex`; opens the Inspector popup at startup.
    Flags: `--theme`, `--snapshot <dump.json>` (READONLY), `--out
    <file.go>` (write generated code on quit).

12. **Theme entries** — add the inspector-specific selectors listed
    below to every `themes/theme-*.go`.

13. **Form designer** — separate handling for the `Form` widget itself,
    which edits user data via reflection on a struct field.

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

All mutations go through the Designer; the Inspector never touches
container methods directly.

| Action | Designer call | Side effects |
|--------|---------------|--------------|
| Property edit | `d.SetField(widget, name, value)` | validate, store, `Relayout(widget)` |
| Style edit (widget overrides) | `widget.SetStyle(sel, *Style)` (via Styles tab) | `Relayout(widget)` |
| Style edit (theme) | `theme.SetStyle(sel, *Style)` (via Styles tab) | `Refresh()` on root |
| `[+ Child]` | `d.Add(parent, kind)` | `Relayout(parent)` |
| `[+ Sibling]` | `d.Add(parent, kind)` + `d.Move(child, parent, idx+1)` | `Relayout(parent)` |
| `[Delete]` | `d.Remove(child)` | `Relayout(parent)`; refocus parent if focus was on child |
| `[↑]` / `[↓]` | `d.Move(child, parent, newIdx)` | `Relayout(parent)` |
| `[Indent]` | `d.Move(child, prevSibling, end)` | `Relayout(prevSibling)` and old parent |
| `[Outdent]` | `d.Move(child, grandparent, parentIdx)` | `Relayout` both |

All mutations log at `slog.Info` with `source=inspector`. Failures
(`Container.Insert` returns `ErrFull`, e.g.) surface as a red
`Notification` toast for 3 seconds and a `slog.Warn` entry.

---

## Non-goals

- **Undo / redo** of inspector edits. The simplest path forward is to
  generate code, save to disk, and re-run; full in-memory undo is a
  future enhancement.
- **Cross-process inspection.** The Inspector lives in the host process
  and reads in-memory state. Remote inspection over a socket is a
  future enhancement; READONLY mode operating on a serialised snapshot
  is the seam.
- **Source modification.** The Inspector emits Go code to a file or the
  clipboard; it never edits the user's existing source.
- **Hot reload.** Generated code is non-executing. Re-running the
  binary picks up changes; the Inspector itself is the short-feedback
  path.
- **Round-tripping closures.** Event handlers, render functions, and
  table providers are emitted as `// TODO` stubs.
- **Concurrent Designer use.** The Designer is single-threaded; the
  TUI's main loop is the natural serialiser.

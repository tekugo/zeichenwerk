# Designer Forms

Every editable widget kind has a `*-form.go` file in package
`widgets`. The form is a plain struct that implements
`inspector.WidgetForm` (and optionally `inspector.ContainerForm`).
This document explains the conventions, the standard helpers, and
how to add a new form for a custom widget.

## File layout

```
widgets/
├── component-form.go     ← ComponentForm: id/class/hint/flags/style
├── style-form.go         ← StyleForm:     fg/bg/border/padding/margin/font  (in core)
├── layout-form.go        ← LayoutForm:    Add-params per cell              (in core)
│
├── static-form.go        ← StaticForm   for *Static
├── button-form.go        ← ButtonForm   for *Button
├── flex-form.go          ← FlexForm     for *Flex
├── grid-form.go          ← GridForm     for *Grid (also: GridLayoutForm)
├── breadcrumb-form.go    ← BreadcrumbForm
├── …                       (one *-form.go per widget kind)
```

The form lives next to the widget it edits because it must read and
write the widget's unexported fields. The form file does **not**
import `inspector` — only `core` and the surrounding `widgets`
package.

## Skeleton

A typical form looks like this:

```go
package widgets

import (
    "fmt"
    "io"

    "github.com/tekugo/zeichenwerk/core"
)

// MyWidgetForm is the WidgetForm for *MyWidget.
type MyWidgetForm struct {
    ComponentForm

    Title  string `group:"general" label:"Title"`
    Active bool   `group:"general" label:"Active"`
    Count  int    `group:"value"   label:"Count"`
}

func (f *MyWidgetForm) Name() string  { return "MyWidget" }
func (f *MyWidgetForm) Group() string { return "leaf" }
func (f *MyWidgetForm) Help() string  { return "One-line help text" }

func (f *MyWidgetForm) Load(w core.Widget) {
    m := w.(*MyWidget)
    f.ComponentForm.Load(&m.Component)
    f.Title = m.title
    f.Active = m.active
    f.Count = m.count
}

func (f *MyWidgetForm) Store(w core.Widget) {
    m := w.(*MyWidget)
    f.ComponentForm.Store(&m.Component)
    m.title = f.Title
    m.active = f.Active
    m.count = f.Count
}

func (f *MyWidgetForm) New() core.Widget {
    m := NewMyWidget("", "", f.Title)
    f.Store(m)
    return m
}

func (f *MyWidgetForm) Validate(field string) error { return nil }

func (f *MyWidgetForm) Emit(w io.Writer, mode string) error {
    return f.EmitFrame(w, mode, func() error {
        _, err := fmt.Fprintf(w, "MyWidget(%q, %q, %t, %d).\n",
            f.ID, f.Title, f.Active, f.Count)
        return err
    })
}
```

Every concrete form follows this shape: embed `ComponentForm`,
declare kind-specific fields with struct tags, implement the seven
required methods.

## Identity methods

```go
func (f *MyWidgetForm) Name() string  // Display name in the toolbox / palette
func (f *MyWidgetForm) Group() string // "container", "input", "leaf", "animation", …
func (f *MyWidgetForm) Help() string  // One-line tooltip / status help
```

`Name` should match the widget type's Go name. `Group` is a freeform
string used by the toolbox to bin similar widgets together; the
groups currently in use are:

- `container` — Box, Card, Flex, Grid, Collapsible, Dialog,
  Switcher, Viewport, Form
- `input` — Input, Checkbox, Combo, Filter, List, Editor, Typeahead
- `leaf` — Static, Button, Breadcrumb, Clock, Digits, Indicator,
  Marquee, Progress, Rule, Scanner, Select, Shortcuts, Spinner,
  Styled, Tabs, Table, Terminal, Text, Tiles, Deck, Tree,
  Typewriter

## Lifecycle methods

```go
func (f *MyWidgetForm) New() core.Widget         // construct a fresh widget
func (f *MyWidgetForm) Load(w core.Widget)       // form ← widget
func (f *MyWidgetForm) Store(w core.Widget)      // form → widget
func (f *MyWidgetForm) Validate(field string) error
```

`Load` copies state from the widget into the form. The caller has
just constructed the form (via `Make()` from the registry); after
`Load` the form is ready to render with current values pre-filled.

`Store` is the inverse: copy form fields back into the widget. Apply
calls `Store`, then `Relayout(widget)`, so any size-affecting field
edit propagates immediately.

`New` returns a freshly constructed widget initialised from the
form's current values. The toolbox uses this when the user picks a
kind from the palette. For widgets whose constructor takes a
non-trivial argument (`Table` needs a `TableProvider`, `Tiles` and
`Deck` need an `ItemRender`, `TreeWidgets` needs a root widget), pass
a placeholder and emit a `// TODO` comment in `Emit`.

`Validate(field)` is called by the form host before applying a
single edit. Return an error to reject the change. Most forms return
`nil` — only forms that parse durations or numeric ranges (e.g.
`ClockForm`, `TypewriterForm`) implement real validation.

## ComponentForm and StyleForm

`ComponentForm` carries six standard fields:

| Field      | Tag                              | What it edits                    |
|------------|----------------------------------|----------------------------------|
| `ID`       | `group:"general"`                | widget id                        |
| `Class`    | `group:"general"`                | style class                      |
| `HintW/H`  | `group:"layout"`                 | preferred size hints             |
| `Skip`     | `group:"flags"`                  | `FlagSkip`                       |
| `Hidden`   | `group:"flags"`                  | `FlagHidden`                     |
| `Disabled` | `group:"flags"`                  | `FlagDisabled`                   |

It also holds an unexported `style StyleForm` snapshot — a per-load
copy of the widget's `*Style`. The embedded `Style()` method returns
a pointer to that snapshot, which the popup's Style tab edits
through a separate `Form`. Because the snapshot and the form bind to
the same struct, edits made via the Style tab flush back through the
next `Store`.

`ComponentForm.Load` and `Store` cover all six fields plus the style
snapshot. Concrete forms call `f.ComponentForm.Load(&w.Component)`
first, then handle their own kind-specific state.

## Emit

`Emit(out, mode)` writes the widget's call shape onto an
in-progress chain. The framework guarantees:

- `mode == "builder"` for now (compose returns an error).
- The caller has already written whatever precedes this widget.
- Each chain element ends with `".\n"`. The last `"."` is stripped
  by the codegen walker.
- `gofmt` runs over the entire emitted output, so indentation and
  line breaks are not the form's concern.

### EmitFrame

The standard wrapper is:

```go
return f.EmitFrame(w, mode, func() error {
    _, err := fmt.Fprintf(w, "Widget(%q, %q).\n", f.ID, f.Title)
    return err
})
```

`EmitFrame` does three things:

1. Calls `CheckBuilderMode` to reject unknown modes.
2. Emits the leading `Class("…")` prefix when `Class` is set.
3. After `body()`, emits the trailing chain (`Hint`, `Flag(…)` for
   skip / hidden / disabled, then `EmitStyle` for the per-widget
   styling chain).

Forms with extra trailing chain elements (Grid emits `.Rows(...)` /
`.Columns(...)` after the standard chain) call `EmitFrame` and then
write their tail directly. Forms with a non-standard prefix shape
(Flex picks `HFlex` or `VFlex` based on the `Vertical` flag) bypass
`EmitFrame` and call the primitives — `EmitClassPrefix`, `EmitChain`,
`EmitStyle` — manually.

### TODO comments

When the Builder has no chained setter for a property (`Set`, `SetText`,
`SetSpeed`, `Select`, …), emit a `// TODO: …` comment after the
frame. The user replaces the comment with the real call after
generation. The convention is:

```go
if f.Speed > 1 {
    fmt.Fprintf(w, "// TODO: SetSpeed(%d) — no Builder setter\n", f.Speed)
}
```

Keep the comment short, mention the proposed setter and the value,
and add a "no Builder setter" reason so a reader can see why it is
not a normal chain element.

## ContainerForm and LayoutForm

If the widget's `Add` consumes per-child parameters, the form
satisfies `ContainerForm` and returns a `LayoutForm`:

```go
type ContainerForm interface {
    WidgetForm
    LayoutForm(parent core.Container, child core.Widget) core.LayoutForm
}
```

The canonical example is `GridForm`. `Grid.Add(child, row, col,
rowspan, colspan)` takes per-cell coordinates, so `GridForm.LayoutForm`
returns a `GridLayoutForm` carrying those four ints. The popup's
Layout tab renders that form into a `FormGroup` so users can edit
the cell while the rest of the property panel keeps showing the
child widget's own properties.

If the widget's `Add` ignores its variadic — Box, Card, Flex,
Collapsible, Dialog, Switcher, Viewport, Form — implement only
`WidgetForm`. The Layout tab will fall back to a placeholder.

## Comma-separated list fields

Form controls only support primitive Go kinds (string, bool,
int/uint, float). For widgets with slice fields (List items, Tabs
names, Combo items, Tree segments, …), the convention is to expose
the slice as a comma-separated string and convert in `Load` /
`Store`:

```go
type ListForm struct {
    ComponentForm
    ItemsRaw string `group:"value" label:"Items (comma-separated)"`
}

func (f *ListForm) Load(w core.Widget) {
    l := w.(*List)
    f.ItemsRaw = strings.Join(l.items, ", ")
    …
}

func (f *ListForm) Store(w core.Widget) {
    l := w.(*List)
    l.items = parseItems(f.ItemsRaw)
    …
}
```

The package-level `parseItems` helper drops blank entries so an
accidental trailing comma does not produce a phantom item. For
key:value pairs (Select options, Shortcuts pairs), parse with
`strings.IndexByte(t, ':')` and split on the colon —
`parseSelectOptions` / `parseShortcutPairs` are the canonical
helpers.

## Duration, level, alignment fields

Widgets that store typed values without a primitive form-control
backing surface them as strings:

- **time.Duration** — `time.ParseDuration` round-trip in Load / Store
  (`ClockForm.Interval`, `TypewriterForm.Dwell`).
- **core.Level** — string-typed already; the form uses
  `control:"select" options:"debug,info,success,warning,error,fatal"`
  and the `levelConst` helper picks the constant or falls back to a
  `Level("…")` cast in codegen.
- **core.Alignment** — `parseAlignment` / `alignmentConst` translate
  between the string surface and the constants
  (`Start, Center, End, Stretch, Left, Right, Default`).

When you add a form for a widget that uses a typed enum, follow the
same pattern: select control with options, a private `parseXxx`
helper for `Store`, and a `xxxConst` mapping for codegen.

## Registration

Forms are not active until they are registered with an
`inspector.Designer`. The driver does this once at startup:

```go
func registerKinds(d *inspector.Designer) {
    register := func(t reflect.Type, mk func() inspector.WidgetForm) {
        if err := d.Register(inspector.Kind{Type: t, Make: mk}); err != nil {
            panic(err)
        }
    }
    register(reflect.TypeOf((*Static)(nil)),
        func() inspector.WidgetForm { return &StaticForm{} })
    register(reflect.TypeOf((*Flex)(nil)),
        func() inspector.WidgetForm { return &FlexForm{} })
    // …
}
```

`cmd/designer-poc/main.go` and `cmd/inspector-poc/main.go` both ship
a `registerKinds` table. Custom drivers do the same — there is no
auto-registration on package import. This is deliberate: a host can
register a subset (a domain-specific palette) or include
out-of-tree widgets without touching the framework.

## Adding a form for a new widget

1. **Create `widgets/myname-form.go`** next to `widgets/myname.go`.
2. **Embed ComponentForm**, declare your editable fields with
   `group` / `label` / `control` / `options` tags.
3. **Implement `Name`, `Group`, `Help`**.
4. **Implement `Load` and `Store`**, calling
   `f.ComponentForm.Load(&w.Component)` /
   `f.ComponentForm.Store(&w.Component)` first.
5. **Implement `New`**, constructing a fresh widget and calling
   `f.Store(w)` to apply the form's current state.
6. **Implement `Validate`** — usually `return nil`.
7. **Implement `Emit`**, normally via `EmitFrame`. Add `// TODO`
   comments for properties without chained Builder setters.
8. **Register the form** in your driver's `registerKinds` table.
9. **Build and test** — the popup will pick up the new kind on next run.

The form is purely additive — adding it does not change any existing
widget behaviour. Removing it removes only the editing surface, not
the widget itself.

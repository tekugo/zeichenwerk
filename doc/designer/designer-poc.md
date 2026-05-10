# designer-poc

`cmd/designer-poc/main.go` is the reference driver around the
inspector framework. It is a working demo, not a finished product —
the goal is to show how the pieces fit together.

```bash
go run ./cmd/designer-poc
```

```text
┌──────────────────── designer popup ───────────────────────┐
│ Header  file • • TokyoNight    [Save] [Generate] [Run]   │
├──────────────────────┬────────────────────────────────────┤
│ Tree                 │ Tabs: General | Layout | Style | …│
│  ▼ VFlex (#root)     │  ┌──────────────────────────────┐ │
│    ▶ HFlex (#header) │  │ ID            [main       ]  │ │
│    ▶ Grid (#g1)      │  │ Class         [           ]  │ │
│                      │  │ Hint W / H    [0 ] [0 ]      │ │
│                      │  │ Vertical      [x]            │ │
│                      │  │ Alignment     [stretch ▼]    │ │
│                      │  │ Spacing       [0 ]           │ │
│                      │  └──────────────────────────────┘ │
│ [Add] [Del] [Up] …   │  [Apply]  [Reset]                 │
├──────────────────────┴────────────────────────────────────┤
│ status: applied → Flex#root                              │
└────────────────────────────────────────────────────────────┘
```

## Keyboard

| Key            | Action                                          |
|----------------|-------------------------------------------------|
| `Ctrl+Space`   | Open / close the designer popup                 |
| `Alt+1`–`Alt+4`| Jump to General / Layout / Style / Info tab     |
| `Tab`          | Cycle focus through form fields                 |
| `Apply`        | Commit form edits to the live widget            |
| `Reset`        | Reload the form from the live widget            |
| `Generate`     | Write Builder-mode source to `proj.OutPath`     |
| `Save`         | Same as Generate, with the Save label           |
| `Esc`          | Close the popup (no commit)                     |
| `Ctrl+Q`       | Quit                                            |

## What the driver builds

`main.go` constructs a four-section UI:

1. **Live target** — a `Preview` containing a `VFlex` with a header
   `HFlex` and a `Grid`. This is the tree the designer edits. The
   `Preview` wrapper hides the subtree from focus / hit-testing, so
   the framework treats the popup as the only interactive surface
   while the target keeps rendering as a live preview.
2. **Status bar** — a one-line `Static` at the bottom of the main
   screen showing keyboard hints (`Ctrl+Space → designer`).
3. **Designer popup** — a free-standing container built by
   `buildDesignerPopup`. Hidden until `Ctrl+Space` opens it via
   `ui.Popup`.
4. **Add-child dialog** — a separate popup launched from the tree
   pane's `[Add]` toolbar button.

## Popup anatomy

The popup is a vertical flex of three rows:

- **Header** — file label + dirty dot + theme label + Save /
  Generate / Run / Settings buttons. The dirty dot is a `Static`
  toggled via `FlagHidden`; its visibility *is* the dirty state.
- **Body** — horizontal flex with the tree on the left and the
  detail tabs on the right.
- **Status** — a one-line `Static` at the bottom. `setStatus`
  helper writes a message and triggers a redraw.

### Tree pane

A `TreeWidgets` rooted at the live target. Selecting a node fires
`EvtSelect`, which calls `rebuildPane(w)` — that is where most of
the UI work happens.

Below the tree is a toolbar of structural-action buttons: Add
(`add-child`), Delete (`delete-child`), Move Up / Down
(`move-up`, `move-down`), Wrap, Unwrap. Each button operates on the
currently selected widget; the handlers update the tree, set the
dirty dot, and rebuild the detail pane.

### Detail tabs

Four tabs share a horizontal flex laid out by `Tabs` + `Switcher`:

- **General** — the per-widget form. `rebuildPane` reflects over
  the form struct, walks anonymous embedded structs (e.g.
  `ComponentForm`), and adds one `FormGroup` section per level so
  every embedded struct gets its own header. Sections are separated
  by thin rules.
- **Layout** — when the parent's form implements `ContainerForm`
  and returns a non-nil `LayoutForm`, the per-child Add-params form
  renders here. Otherwise a "parent has no per-child layout
  parameters" placeholder appears. Below either branch is a
  read-only Computed block showing the current bounds and hint.
- **Style** — `form.Style()` returns a pointer to the style
  snapshot that `ComponentForm.Load` populated. The Style tab
  binds it as its own `Form` so foreground / background / border /
  padding / margin / font edits flush back through the same Apply.
- **Info** — read-only kind summary: type, id, parent description,
  child count, flag summary.

The popup precomputes the four `*Box` panes (`tab-general`,
`tab-layout`, `tab-style`, `tab-info`) and reuses them across
selections; `rebuildPane` clears each pane and adds the new content.

## Apply, Reset, Generate

`apply()` is the heart of the editing loop:

```go
apply := func() {
    currentForm.Store(currentWidget)
    if currentLayout != nil && currentParent != nil {
        currentLayout.Store(currentParent, currentWidget)
    }
    Relayout(currentWidget)
    currentNode.SetText(treeLabel(currentWidget))
    rebuildPane(currentWidget)
    setDirty(true)
    setStatus(status, "applied → …")
}
```

It writes the form back into the widget, refreshes the per-child
layout if a `LayoutForm` is active, relays the widget so size
changes propagate, updates the tree label (the id may have
changed), then rebuilds the detail pane to pick up cascaded changes.

`reset()` reloads the form from the live widget, undoing any
in-progress edits without applying them.

`writeFile()` calls `Designer.GenerateFile(ModeBuilder, &buf,
proj.Package, proj.FuncName)` and writes the result to
`proj.OutPath`. Save and Generate share this implementation.

## Form registration

The `registerKinds(d)` table is the only piece of glue between the
inspector framework and the widgets package. Each entry pairs a
widget type with a factory that returns a fresh form:

```go
register(reflect.TypeOf((*Static)(nil)),
    func() inspector.WidgetForm { return &StaticForm{} })
register(reflect.TypeOf((*Flex)(nil)),
    func() inspector.WidgetForm { return &FlexForm{} })
register(reflect.TypeOf((*Breadcrumb)(nil)),
    func() inspector.WidgetForm { return &BreadcrumbForm{} })
// … one entry per supported kind
```

Adding support for a new widget is a two-step process:

1. Write the `*-form.go` file in the `widgets` package (see
   [forms.md](forms.md)).
2. Add a `register(...)` line to `registerKinds` in
   `cmd/designer-poc/main.go` (and, optionally,
   `cmd/inspector-poc/main.go`).

If the new kind is not registered, `FormFor(w)` returns nil and the
detail pane shows "no form registered for this widget".

## Currently registered kinds

The PoC currently registers the following kinds (alphabetical):

- Containers — `Box`, `Card`, `Collapsible`, `Dialog`, `Flex`,
  `Form`, `Grid`, `Switcher`, `Viewport`
- Inputs — `Button`, `Checkbox`, `Combo`, `Editor`, `Filter`,
  `Input`, `List`, `Typeahead`
- Display / leaf — `Breadcrumb`, `Clock`, `Deck`, `Digits`,
  `Indicator`, `Marquee`, `Progress`, `Rule`, `Scanner`,
  `Select`, `Shortcuts`, `Spinner`, `Static`, `Styled`, `Table`,
  `Tabs`, `Terminal`, `Text`, `Tiles`, `Tree`, `Typewriter`

`TreeFS` and `TreeWidgets` register their inner `*Tree` to the
parent at build time, so a single `*Tree` registration covers the
common case. Specialised editing for those wrappers happens through
dedicated dialogs rather than the General tab.

## Project metadata

```go
type project struct {
    Name     string  // file label in the header
    OutPath  string  // file written by Save / Generate
    Package  string  // emitted package declaration
    FuncName string  // emitted func wrapper name
    Theme    string  // theme label shown in the header
}
```

The defaults match the PoC's working assumptions
(`/tmp/designer-poc-out.go`, `package main`, `func BuildUI`,
`TokyoNight`). The Settings dialog (header `[Settings]` button)
edits these in place; they round-trip through codegen so a
generated file picks up the current values.

## Extending the driver

Common customisations:

- **New tab** — add a `*Box` pane to `buildDesignerPopup`, register
  it as a `Switcher` pane, and have `rebuildPane` populate it.
  Useful for an Events tab (`form.Help()` already returns a
  one-line description; an Events tab could enumerate registered
  handlers).
- **Custom toolbox** — the structural-action toolbar
  (`add-child`, `delete-child`, ...) is a horizontal flex of
  buttons; adding more is a matter of `_ = toolbar.Add(NewButton(…))`
  and wiring an `EvtActivate` handler.
- **Different popup shape** — `buildDesignerPopup` is a single
  function. Swap it for a side-panel / split-screen version
  without touching the inspector framework.
- **Theme switcher** — the project metadata stores a theme name;
  a Settings dialog could expose a dropdown of registered themes
  and rebuild the UI on change.

The framework's contract is `register kinds → ask for forms → call
Emit`. Anything else is the driver's responsibility, including
which forms to expose, how to lay them out, and what to do with
generated code.

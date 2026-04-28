# 7. Containers in depth

Beyond Flex and Grid, four containers come up often enough to deserve
their own chapter: **Switcher**, **Tabs**, **Form/FormGroup**, and
**Dialog** (for popups).

## Switcher — one child at a time

`Switcher(id, connect)` holds many children but renders only one. Calling
`.Select(i)` flips to a different child. Pair it with a `List`, `Tabs`, or
sidebar of buttons to build a multi-screen shell:

```go
ui := NewBuilder(theme).
    Grid("shell", 1, 2, false).Columns(20, -1).
        Cell(0, 0, 1, 1).
            List("nav", "Dashboard", "Settings", "About").
        Cell(1, 0, 1, 1).
            Switcher("content", false).
                With(dashboard).        // each is a func(*Builder)
                With(settings).
                With(about).
            End().
    End().
    Build()

switcher := core.MustFind[*widgets.Switcher](ui, "content")
widgets.OnSelect(core.MustFind[*widgets.List](ui, "nav"), func(i int) bool {
    switcher.Select(i)
    return true
})
```

Each pane is a regular function that adds children to the current
container:

```go
func dashboard(b *Builder) {
    b.VFlex("dash", core.Stretch, 1).Padding(1, 2).
        Static("dash-title", "Dashboard").Font("bold").Foreground("$cyan").
        Static("dash-body", "Metrics, charts, KPIs go here.").
    End()
}
```

`With(fn)` is the builder's general-purpose composition helper — it just
calls `fn(builder)`. Same trick works in any container, not only
Switcher.

The `connect` flag (second arg) controls how children are deactivated:

- `connect: false` — switching just hides the inactive child. Cheap.
  State (scroll position, focus, animation tickers) is preserved.
- `connect: true` — fires `EvtShow` / `EvtHide` on each transition. Use
  this when a pane needs to start/stop work when it becomes visible
  (load data on show, pause an animation on hide).

Full example: [`examples/07-switcher-nav`](examples/07-switcher-nav/main.go).

## Tabs

`Tabs(id, names...)` renders a tab strip. Wire its `EvtActivate` to a
sibling `Switcher`:

```go
Tabs("strip", "Inbox", "Sent", "Drafts").Hint(0, 1).
Switcher("pane", false).
    With(inbox).
    With(sent).
    With(drafts).
End()
…
widgets.OnActivate(core.MustFind[*widgets.Tabs](ui, "strip"), func(i int) bool {
    core.MustFind[*widgets.Switcher](ui, "pane").Select(i)
    return true
})
```

The Tabs widget does its own keyboard handling (←/→ to move, Enter to
activate, letter shortcuts). You only wire the activate handler.

## Forms — bind a struct to widgets

`Form(id, title, data)` accepts a **pointer to a struct** and reflects on
it. Combined with `Group(...)` it auto-generates a Input/Checkbox/Select
for each field:

```go
type Profile struct {
    Name       string `label:"Full name"        width:"30"`
    Email      string `label:"Email address"    width:"30"`
    Password   string `label:"Password"         control:"password" width:"20"`
    Theme      string `label:"Preferred theme"  control:"select"   options:"tokyo,nord,gruvbox,lipstick"`
    Newsletter bool   `label:"Subscribe to newsletter"`
}

profile := &Profile{Name: "Ada", Theme: "tokyo"}

NewBuilder(theme).
    Form("profile", "", profile).
        Group("fields", "", "", false, 1).
        End(). // Group
    End().     // Form
    Run()
```

> **Common mistake.** `Form` is a *single-child* container — calling `Add`
> on it a second time replaces the first child. So if you forget the
> `End()` for `Form` and chain another widget at that depth, that widget
> silently displaces the `Group` and you see the new widget where the
> form should be. Always close `Form` explicitly.

Tags the form recognises:

| Tag | Default | Effect |
|-----|---------|--------|
| `label:"…"` | field name | Visible label. `label:"-"` skips the field. |
| `control:"…"` | `input` for strings, `checkbox` for bools | One of `input`, `password`, `checkbox`, `select`. |
| `options:"a,b,c"` | — | Choices for `select`. |
| `width:"N"` | `10` | Width of the control in cells. |
| `readonly:""` | — | Marks the field as readonly. |
| `group:"name"` | — | Group filter — only fields with this tag appear in `Group(…, "name", …)`. |
| `line:"N"` | auto | Pin the field to a specific line in the group. |

Updates flow back automatically: every change event writes the new value
into the bound struct field. Read the struct in your save handler and
you've got the form data.

Full runnable: [`examples/07-form`](examples/07-form/main.go).

## Dialog — popup overlay

`Dialog(id, title)` is a single-child container made for popup layers.
Build a dialog tree separately, then push it on top of the UI with
`ui.Popup(x, y, w, h, dialog)`:

```go
b := ui.NewBuilder()
dialog := b.Dialog("confirm", "Are you sure?").
    VFlex("body", core.Stretch, 1).Padding(1, 2).
        Static("msg", "This will delete the file.").
        HFlex("actions", core.End, 2).
            Button("yes", "Delete").
            Button("no", "Cancel").
        End().
    End().
    Container() // returns the Dialog as a Container

// Coordinates: -1 in x or y means "center on that axis".
ui.Popup(-1, -1, 50, 7, dialog.(core.Container))
```

`ui.NewBuilder()` returns a builder seeded with the UI's theme — handy
when you need to build extra trees after the main one.

The framework also ships two helper modal dialogs you'd otherwise build
yourself:

```go
ui.Confirm("Delete?", "Are you sure?",
    func() { /* confirmed */ },
    func() { /* cancelled */ })

ui.Prompt("Enter name", "Your name:",
    func(name string) { /* accepted */ },
    func() { /* cancelled */ })
```

Both push a dialog on a new layer. Press `Esc` to close, or call
`ui.Close()` programmatically.

## Viewport — scroll oversize content

When a child is too big for its slot, wrap it in a `Viewport`:

```go
Viewport("scroll", "Logs").Border("thin").Hint(0, -1).
    Text("body", lines, false, 1000).
End()
```

Page Up / Page Down, arrow keys, and the mouse wheel scroll. By default
both axes are enabled; restrict to one with the
`FlagVertical` / `FlagHorizontal` flags via `.Flag(core.FlagVertical, true)`.

→ [Next: Custom widgets](08-custom-widgets.md)

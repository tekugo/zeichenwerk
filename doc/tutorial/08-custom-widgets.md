# 8. Custom widgets

When the built-in widgets don't fit — and they often won't for
application-specific visuals — you write your own. This chapter walks
through a small status-indicator widget that draws "● Online" with a
colour-coded dot. It shows every piece you need: struct, constructor,
`Render`, `Apply`, `Hint`, and how to put it into a Builder tree.

For one-off renders that don't deserve a full type, skip ahead to the
[`Custom`](#one-off-the-custom-widget) section at the end of the chapter.

## The shape of a widget

Every concrete widget in zeichenwerk is a Go struct that embeds
`*widgets.Component` (or a sibling like `widgets.Animation`) and overrides
the methods it cares about:

```go
type MyWidget struct {
    *widgets.Component
    // … your own fields …
}
```

`Component` provides default implementations of every method on the
`core.Widget` interface — bounds, focus flag, parent reference, event
dispatch, theme application, the lot. You only override what you want to
specialise:

| Override | Why |
|----------|-----|
| `Render(r *core.Renderer)` | Always — that's where you draw. |
| `Apply(theme *core.Theme)` | Almost always — point the theme at your widget type. |
| `Hint() (w, h int)` | If your natural size depends on content (text width, item count, …). |
| `Cursor() (x, y int, style string)` | Only for input widgets that show a cursor. |

## A worked example: `Status`

[`examples/08-custom-widget/main.go`](examples/08-custom-widget/main.go)
defines a status indicator:

```go
type Status struct {
    *widgets.Component
    Stat  string  // "ok", "warn", "err", or anything else
    Label string
}

func NewStatus(id, class, stat, label string) *Status {
    return &Status{
        Component: widgets.NewComponent(id, class),
        Stat:      stat,
        Label:     label,
    }
}
```

`widgets.NewComponent(id, class)` returns a fresh `*Component` with the ID
and class wired in. We embed it as a pointer because `Component`'s state
fields (bounds, parent, handlers) are unexported; we couldn't construct
one by value from outside the `widgets` package anyway.

> **Naming tip.** `Component` already has a method called `State()` (it
> returns "focused", "disabled", etc., used by the style system). The
> example uses `Stat` for the widget's own state field to avoid
> shadowing.

### Apply

`Apply` is how the framework asks your widget to read its style from a
theme. The convention is to pass `Selector("yourTypeName")` — that's the
hook themes use to target your widget specifically:

```go
func (s *Status) Apply(theme *core.Theme) {
    theme.Apply(s, s.Selector("status"))
}
```

`s.Selector("status")` returns `"status"`, `"status.myclass"`, or
`"status#myid"` depending on what was set. Themes can register styles for
any of those.

### Hint

`Hint` reports the widget's natural content size — what it'd like if it
got to ask. The `Status` widget needs two cells for "● " and one per rune
of the label:

```go
func (s *Status) Hint() (int, int) {
    return 2 + utf8.RuneCountInString(s.Label), 1
}
```

The Builder's `.Hint(w, h)` overrides this. Containers consult `Hint`
during layout when `.Hint(0, 0)` (or no `.Hint(...)`) is in effect — the
"auto" case.

### Render

`Render` does the actual drawing. The contract:

1. Call `s.Component.Render(r)` first. That paints margin, border, and
   background defined by the theme — everything outside your content.
2. Use `s.Content()` to get the **inner** drawable rectangle (after
   margin / border / padding). Never draw past `w` and `h`.
3. The renderer's style state is **sticky** — `r.Set(fg, bg, font)` lasts
   until the next `Set`. Call it before each batch of draws that need a
   different look.

```go
func (s *Status) Render(r *core.Renderer) {
    s.Component.Render(r)

    x, y, w, _ := s.Content()
    if w <= 0 {
        return
    }

    style := s.Style()
    bg := style.Background()

    dot := "$gray"
    switch s.Stat {
    case "ok":   dot = "$green"
    case "warn": dot = "$yellow"
    case "err":  dot = "$red"
    }
    r.Set(dot, bg, "")
    r.Text(x, y, "●", 1)

    r.Set(style.Foreground(), bg, style.Font())
    r.Text(x+2, y, s.Label, w-2)
}
```

Theme variables (`$green`, etc.) are resolved by the renderer that wraps
your widget — `core.Renderer` looks up `$green` in the theme and forwards
the resolved colour to tcell. So you can write theme variables directly
in `r.Set` without doing your own lookup.

### Mutators

Setters that change state should:

- update the field,
- queue a redraw via `widgets.Redraw(self)`,
- **not** fire `EvtChange` (that's reserved for user-driven changes; see
  [`doc/principles.md`](../principles.md)).

```go
func (s *Status) Set(stat, label string) {
    s.Stat = stat
    s.Label = label
    widgets.Redraw(s)
}
```

If your change *also* alters the natural size (e.g. the label grew),
`widgets.Relayout(s)` is the correct helper instead of `Redraw`.

## Adding to the tree

There's no built-in Builder method for your widget — you'd have to add one
to `builder.go` for that. Use the generic `Builder.Add` instead:

```go
b := zw.NewBuilder(themes.TokyoNight()).
    VFlex("root", core.Stretch, 1).Padding(1, 2).
    Static("title", "Status panel").Font("bold")

for _, s := range statuses {
    b.Add(s)
}

b.End().Run()
```

`Add` calls `widget.Apply(theme)` for you, so the theme is in effect
before the widget ever renders.

## Animated widgets

If your widget needs to tick on a timer (a clock, a spinner, a live
chart), embed `widgets.Animation` instead of `*widgets.Component`:

```go
type Counter struct {
    widgets.Animation     // value embed, contains *Component already
    n int
}

func NewCounter(id, class string) *Counter {
    c := &Counter{}
    // tickFn is the entry point Animation calls each frame.
    c.Animation = widgets.Animation{ /* ... see widgets/spinner.go ... */ }
    return c
}

func (c *Counter) Tick() {
    c.n++
    c.Refresh() // queue a redraw
}
```

Then call `counter.Start(time.Second)` after construction. `Stop()` halts
the goroutine cleanly.

The full pattern is a bit denser than `Status` — read
[`widgets/spinner.go`](../../widgets/spinner.go) as a complete reference.

## One-off: the `Custom` widget

When you only need a small splash of custom drawing — a divider, a logo,
an unusual cell pattern — promoting it to a full type is overkill. Use
`widgets.Custom`:

```go
ascii := widgets.NewCustom("logo", "", func(w core.Widget, r *core.Renderer) {
    x, y, _, _ := w.(*widgets.Custom).Content()
    r.Set("$cyan", "", "bold")
    r.Text(x, y,   "  ███╗   ██╗ ", 0)
    r.Text(x, y+1, "  ████╗  ██║ ", 0)
    r.Text(x, y+2, "  ██╔██╗ ██║ ", 0)
    r.Text(x, y+3, "  ██║╚██╗██║ ", 0)
    r.Text(x, y+4, "  ██║ ╚████║ ", 0)
    r.Text(x, y+5, "  ╚═╝  ╚═══╝ ", 0)
})
b.Add(ascii)
```

`Custom` gives you the `Component` defaults (focus, events, theme) and
just delegates the draw call to your closure. Promote to a real type when
you find yourself copying the closure into multiple files.

→ [Next: Building a SQLite query tool](09-sqlite-query-tool.md)

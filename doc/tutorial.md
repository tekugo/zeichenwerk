# Zeichenwerk Tutorial

This tutorial walks you through building terminal UIs with Zeichenwerk, from a
minimal "hello world" to a small interactive application. It uses the Builder API
throughout; the [Composition API](#composition-api) section at the end shows the
same patterns with the functional alternative.

## Prerequisites

```bash
go get github.com/tekugo/zeichenwerk
```

## 1. Hello World

The minimal program puts text on screen and waits for the user to press `q` or
`Ctrl+C`:

```go
package main

import . "github.com/tekugo/zeichenwerk"

func main() {
    NewBuilder(TokyoNightTheme()).
        Static("greeting", "Hello, Zeichenwerk!").
        Run()
}
```

`NewBuilder` takes a theme and returns a builder. Every widget method adds a widget
and returns the same builder so calls can be chained. `Run()` builds the UI and
starts the event loop — it blocks until the user quits.

## 2. Layouts

Widgets are arranged with **Flex** (linear) and **Grid** (table) containers.

### Flex

`Flex(id, horizontal, alignment, spacing)` stacks children in a row or column.

```go
NewBuilder(TokyoNightTheme()).
    Flex("root", false, "stretch", 0).   // vertical column
        Static("line1", "First line").
        Static("line2", "Second line").
    End().
    Run()
```

`End()` closes the current container and returns to its parent. Forgetting `End()`
is the most common mistake — the builder will happily accept more children into the
wrong parent.

**Alignment** controls the cross-axis:
- `"stretch"` — children fill the full width/height
- `"start"` / `"end"` — align to the near/far edge
- `"center"` — centre children

**Size hints** control how space is distributed. Pass `Hint(width, height)` after
a widget method. `-1` means "fill remaining space"; `0` means "auto-size to
content"; positive values are fixed cell counts:

```go
Flex("root", true, "stretch", 0).  // horizontal row
    Static("label", "Name:").Hint(8, 1).
    Input("name", "Your name").Hint(-1, 1).  // fills remaining width
End()
```

### Grid

`Grid(id, rows, columns, lines)` places children in named cells. Call `Columns`
and `Rows` to set sizes, then `Cell(x, y, w, h)` before each child to choose its
position and span:

```go
NewBuilder(TokyoNightTheme()).
    Grid("layout", 1, 2, false).Columns(24, -1).
        Cell(0, 0, 1, 1).Static("sidebar", "Sidebar").
        Cell(1, 0, 1, 1).Static("content", "Content area").
    End().
    Run()
```

## 3. Styling

### Themes

Four built-in themes are provided: `TokyoNightTheme()`, `MidnightNeonTheme()`,
`GruvboxDarkTheme()`, and `NordTheme()`. Pass one to `NewBuilder`.

### Per-widget colours and fonts

Styling methods are chained after the widget they apply to:

```go
Static("title", "Dashboard").
    Foreground("$cyan").
    Font("bold").
    Padding(0, 2)
```

Colour values are either theme variables (`"$cyan"`, `"$fg0"`, `"$bg1"`) or
literal colour names supported by tcell (`"red"`, `"#ff6347"`). Theme variables
are preferred — they automatically adapt when the user switches themes.

Common theme variables (defined in every built-in theme):

| Variable | Meaning |
|----------|---------|
| `$bg0`, `$bg1`, `$bg2` | Background shades (darkest → lightest) |
| `$fg0`, `$fg1` | Foreground shades (brightest → dimmed) |
| `$gray` | Muted text |
| `$blue`, `$cyan`, `$green`, `$red`, `$yellow`, `$magenta`, `$orange` | Accent colours |

### Borders and padding

```go
Box("card", "User Info").
    Border("round").
    Padding(1, 2)
```

`Border` accepts the style name defined in the active theme (`"none"`, `"thin"`,
`"thick"`, `"round"`, `"double"`, `"lines"`, …). `Padding(v, h)` sets vertical
and horizontal padding; four values follow the CSS order (top, right, bottom, left).

## 4. Events

Register handlers with `.On(event, handler)` after a widget. The handler signature
is `func(source Widget, event Event, data ...any) bool`. Returning `true` stops
the event from bubbling to parent widgets.

```go
Button("ok", "Confirm").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
    // do something
    return true
})
```

Typed helpers avoid the manual `data` type assertion for common events:

```go
OnActivate(button, func(index int) bool { … })
OnChange(input, func(value string) bool { … })
OnSelect(list, func(index int) bool { … })
```

Common events:

| Constant | Fires when |
|----------|-----------|
| `EvtActivate` | Enter pressed on a Button, List row, or Tab |
| `EvtChange` | Input text changed, Checkbox toggled, Select value changed |
| `EvtSelect` | Highlighted item changed in a List or Table |
| `EvtShow` / `EvtHide` | A Switcher pane becomes visible or hidden |
| `EvtKey` | Any key event (data is `*tcell.EventKey`) |

## 5. A Small Interactive Application

Here is a task list — a List showing items and a Button to mark the selected item
done:

```go
package main

import (
    "fmt"

    . "github.com/tekugo/zeichenwerk"
)

func main() {
    tasks := []string{
        "Write tests",
        "Fix bug #42",
        "Update docs",
    }

    NewBuilder(TokyoNightTheme()).
        Flex("root", false, "stretch", 0).Padding(1, 2).
            Static("title", "Task List").Font("bold").Foreground("$cyan").
            HRule("thin").
            List("tasks", tasks...).Hint(0, -1).
            HRule("thin").
            Flex("actions", true, "end", 2).
                Button("done", "Mark Done").
                Button("quit", "Quit").
            End().
        End().
        Run()
}
```

Now wire the buttons. Because the builder returns `*UI` from `Build()`, retrieve
widgets with `Find` and attach handlers after construction:

```go
func main() {
    tasks := []string{"Write tests", "Fix bug #42", "Update docs"}

    ui := NewBuilder(TokyoNightTheme()).
        Flex("root", false, "stretch", 0).Padding(1, 2).
            Static("title", "Task List").Font("bold").Foreground("$cyan").
            HRule("thin").
            List("tasks", tasks...).Hint(0, -1).
            HRule("thin").
            Flex("actions", true, "end", 2).
                Button("done", "Mark Done").
                Button("quit", "Quit").
            End().
        End().
        Build()

    list := Find(ui, "tasks").(*List)

    Find(ui, "done").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
        i := list.Index()
        if i >= 0 {
            items := list.Items()
            items[i] = "✓ " + items[i]
            list.SetItems(items)
        }
        return true
    })

    Find(ui, "quit").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
        ui.Quit()
        return true
    })

    ui.Run()
}
```

The pattern is: **build the tree declaratively**, then **retrieve widgets by ID and
wire them imperatively**. `Find` does a depth-first search and returns `nil` if the
ID is not found, so the type assertion will panic on a typo — keep IDs short and
consistent.

## 6. Switcher: Multiple Screens

`Switcher` shows one child at a time. Pair it with a `List` or `Deck` in a sidebar
to build a classic navigation shell:

```go
ui := NewBuilder(TokyoNightTheme()).
    Grid("shell", 1, 2, false).Columns(20, -1).
        Cell(0, 0, 1, 1).
            List("nav", "Dashboard", "Settings", "About").
        Cell(1, 0, 1, 1).
            Switcher("content", false).
                With(dashboardScreen).
                With(settingsScreen).
                With(aboutScreen).
            End().
    End().
    Build()

switcher := Find(ui, "content").(*Switcher)
nav := Find(ui, "nav").(*List)
nav.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
    if i, ok := data[0].(int); ok {
        switcher.Select(i)
    }
    return true
})
```

Each screen is a plain function that receives the builder in scope and adds
children to the current container:

```go
func dashboardScreen(b *Builder) {
    b.Static("dash-title", "Dashboard").Font("bold")
    b.Static("dash-body", "Metrics go here.")
}
```

## 7. Composition API

The `compose` sub-package offers a functional alternative to the builder. Every
widget is an `Option` — a plain function — and options are nested directly. The
theme flows through automatically:

```go
import (
    z "github.com/tekugo/zeichenwerk"
    . "github.com/tekugo/zeichenwerk/compose"
)

func main() {
    UI(z.TokyoNightTheme(),
        Flex("root", "", false, "stretch", 0,
            Padding(1, 2),
            Static("title", "", "Task List", Font("bold"), Fg("$cyan")),
            List("tasks", "", []string{"Write tests", "Fix bug #42", "Update docs"},
                Hint(0, -1),
            ),
        ),
    ).Run()
}
```

Split screens into separate functions with `Include`:

```go
UI(z.TokyoNightTheme(),
    Flex("shell", "", false, "stretch", 0,
        Include(header),
        Include(body),
        Include(footer),
    ),
).Run()

func header(theme *z.Theme) z.Widget {
    return Build(theme,
        Flex("header", "", true, "center", 0,
            Bg("$bg1"), Padding(0, 1),
            Static("title", "", "My App", Font("bold"), Fg("$fg0")),
        ),
    )
}
```

Where direct widget access is needed after construction, use `z.Find` exactly as
with the builder:

```go
ui := UI(z.TokyoNightTheme(), …)
list := z.Find(ui, "tasks").(*z.List)
list.On(z.EvtActivate, …)
ui.Run()
```

## 8. Debugging

Pass `.Debug()` to the UI before calling `Run()` to enable the debug bar and the
built-in inspector. While the app is running, press `Ctrl+D` to open the inspector
and explore the live widget tree, bounds, and styles:

```go
NewBuilder(TokyoNightTheme()).
    …
    Build().
    Debug().
    Run()
```

## Next Steps

- **Widget reference** — [`doc/reference/overview.md`](reference/overview.md)
- **Builder API** — [`builder.go`](../builder.go)
- **Composition API** — [`compose/compose.go`](../compose/compose.go)
- **Full example** — [`cmd/showcase/main.go`](../cmd/showcase/main.go) (Builder)
  and [`cmd/compose/main.go`](../cmd/compose/main.go) (Composition)

```bash
go run ./cmd/showcase   # run the showcase with the Builder API
go run ./cmd/compose    # same showcase with the Composition API
```

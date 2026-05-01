# 1. A TUI in 20 lines

Before we explain anything, let's just look at what a small zeichenwerk app
looks like. You'll meet most of the framework's load-bearing ideas in one go;
the rest of the tutorial unpacks them.

## What we're building

A task list with three widgets and one wired-up button:

- a **title** in cyan bold,
- a scrollable **list** of tasks,
- two **buttons** at the bottom — one marks the highlighted task done, the
  other quits.

It's about 35 lines including imports.

## The code

[`examples/01-teaser/main.go`](examples/01-teaser/main.go):

```go
package main

import (
    . "github.com/tekugo/zeichenwerk/v2"
    "github.com/tekugo/zeichenwerk/v2/core"
    "github.com/tekugo/zeichenwerk/v2/themes"
    "github.com/tekugo/zeichenwerk/v2/widgets"
)

func main() {
    tasks := []string{
        "Read the zeichenwerk teaser",
        "Build a tiny TUI",
        "Wire up an event handler",
        "Try a different theme",
    }

    ui := NewBuilder(themes.TokyoNight()).
        VFlex("root", core.Stretch, 0).Padding(1, 2).
            Static("title", "My First TUI").Font("bold").Foreground("$cyan").
            HRule("thin").
            List("tasks", tasks...).Hint(0, -1).
            HRule("thin").
            HFlex("actions", core.End, 2).
                Button("done", "Mark Done").
                Button("quit", "Quit").
            End().
        End().
        Build()

    list := core.MustFind[*widgets.List](ui, "tasks")

    widgets.OnActivate(core.MustFind[*widgets.Button](ui, "done"), func(_ int) bool {
        i := list.Selected()
        if i >= 0 {
            items := list.Items()
            items[i] = "✓ " + items[i]
            list.Set(items)
        }
        return true
    })

    widgets.OnActivate(core.MustFind[*widgets.Button](ui, "quit"), func(_ int) bool {
        ui.Quit()
        return true
    })

    ui.Run()
}
```

Run it:

```bash
go run ./doc/tutorial/examples/01-teaser
```

Use ↑/↓ to move through the list, Enter to "click" a button, Tab to switch
between the list and the buttons, and `q` or `Ctrl-Q` to quit.

## What just happened

Reading top-to-bottom:

**Theme.** `themes.TokyoNight()` returns a fully-populated `*Theme` —
colours, borders, default styling for every widget. Other themes:
`themes.Nord()`, `themes.MidnightNeon()`, `themes.GruvboxDark()`,
`themes.GruvboxLight()`, `themes.Lipstick()`. Try swapping it.

**Builder.** `NewBuilder(theme)` returns a fluent builder. Every widget
method (`VFlex`, `Static`, `List`, `Button`, …) adds a widget and returns
the same builder, so calls chain. Container widgets (`VFlex`, `HFlex`,
`Grid`, `Box`, …) push onto a stack; `End()` pops back to the parent.

**Layout.** `VFlex` stacks children vertically; `HFlex` stacks them
horizontally. `core.Stretch` is a cross-axis alignment — children fill the
full width here. `Hint(0, -1)` says "auto-width, take all remaining
height" — that's how the list grows to fill the middle while the title and
buttons stay at their natural size.

**Styling per widget.** `.Font("bold")` and `.Foreground("$cyan")` apply
inline. The `$cyan` is a *theme variable* — switch themes and the colour
follows. Literal colours (`"red"`, `"#ff6347"`) work too but lock you to one
palette.

**Build, then wire.** `Build()` returns the `*UI`. After construction we
fish widgets out with `core.MustFind[T]` (a generic typed lookup that panics
if the ID is missing or the type is wrong) and attach handlers with
`widgets.OnActivate`. Buttons fire `EvtActivate` when Enter is pressed; the
typed helper unpacks the event for you.

**Quitting.** `ui.Quit()` ends the event loop cleanly. The framework also
binds `q`, `Q`, `Ctrl-Q`, and `Ctrl-C` to quit globally, so you don't have
to wire those yourself.

## What you didn't have to write

Compare against a hand-rolled tcell program — these are all free:

- the tcell screen lifecycle and event poller,
- a layout engine (Flex, Grid, padding, hints),
- focus traversal with Tab / Shift-Tab,
- mouse hit-testing and click routing,
- a redraw / refresh pipeline that only repaints what changed,
- theme-aware styling for every widget,
- a built-in inspector you can pop open with `Ctrl-D` (see
  [Debugging](A-debugging.md)).

That's the entire teaser. The next chapter explains the mental model behind
what you just wrote — the four layers, the frame lifecycle, and why we
build the tree first and wire events afterwards.

→ [Next: Architecture & lifecycle](02-architecture.md)

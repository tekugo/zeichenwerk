# 2. Architecture & lifecycle

This chapter is conceptual — no new code. It gives you the mental model so
the rest of the tutorial reads as "here's the next layer," not "here's
another magic incantation."

## The four layers

Zeichenwerk is built as four packages, each with a clear role:

```
┌──────────────────────────────────────────────────────────────┐
│  builder / compose      ← DSLs you call from your app code    │
├──────────────────────────────────────────────────────────────┤
│  widgets                ← concrete widgets: List, Button, …   │
├──────────────────────────────────────────────────────────────┤
│  core                   ← Widget / Container / Theme contract │
├──────────────────────────────────────────────────────────────┤
│  renderer               ← thin tcell wrapper, drawing primitives │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                       github.com/gdamore/tcell
```

| Layer | What lives there | When you touch it |
|-------|------------------|-------------------|
| **`renderer`** | `Screen`, `Renderer`, drawing primitives over tcell | Almost never — only when writing very low-level custom widgets. |
| **`core`** | `Widget` and `Container` interfaces, `Theme`, `Style`, `Event`, alignment / flag constants, geometry helpers (`Find`, `Layout`, `Traverse`, `MustFind`) | Often — `core.Stretch`, `core.MustFind`, `core.Debug`, etc. |
| **`widgets`** | Concrete widgets and event helpers (`OnActivate`, `OnChange`, …) | Often — both for widget types in `MustFind[*widgets.List]` and for typed event helpers. |
| **`builder`** (root package) and **`compose`** | The two DSLs for assembling a tree | Almost every line of UI code. |

The `themes` and `values` sub-packages are smaller add-ons:

- **`themes`** — pre-built `*Theme` values (`themes.TokyoNight()`, etc.) plus
  helpers to register Unicode borders and default styles.
- **`values`** — a tiny reactive layer for binding form fields and pushing
  data into widgets (`values.Update(ui, "tables", names)`).

## A typical import block

You'll see this shape over and over:

```go
import (
    . "github.com/tekugo/zeichenwerk"     // NewBuilder, Builder, UI
    "github.com/tekugo/zeichenwerk/core"  // Stretch, MustFind, Debug, …
    "github.com/tekugo/zeichenwerk/themes"
    "github.com/tekugo/zeichenwerk/widgets"
)
```

The dot-import of the root package gives you `NewBuilder`, `Builder`, and
`UI` directly. Everything else is qualified — that keeps it obvious whether
a name comes from `core`, `widgets`, or your own code.

## The frame lifecycle

A zeichenwerk app passes through these phases:

```
                ┌──────────────────────────┐
                │  NewBuilder(theme).…       │   you build the tree
                │  …Build()                 │
                └──────────────┬────────────┘
                               │
                               ▼
                ┌──────────────────────────┐
                │  ui.Run()                 │
                │   ├── tcell screen init  │
                │   ├── Layout()           │   ⇦ first layout pass
                │   ├── Draw()             │   ⇦ first paint
                │   └── event loop         │
                └──────────────┬────────────┘
                               │
                  ┌────────────┴───────────┐
                  ▼                        ▼
          tcell.EventKey           tcell.EventResize
          tcell.EventMouse                 │
                  │                        ▼
                  ▼                  Layout() + Draw()
          dispatch to widget
                  │
                  ▼
          handler returns
                  │
       ┌──────────┴───────────┐
       ▼                      ▼
   Redraw(w)              Relayout(w)
   (cheap repaint        (size hint changed,
    of one widget)         re-layout subtree)
```

Two principles fall out of this:

**`Layout()` never paints; `Render()` never mutates.** The layout pass
computes bounds (`SetBounds(x, y, w, h)`) for every widget. The render pass
reads those bounds and pushes cells to the screen. They're separate phases
on purpose: rendering is allowed to skip cleanly when nothing changed.

**Use the cheapest update.** When you change a widget's *contents* but not
its *size*, call `widgets.Redraw(w)` — only that widget repaints. When the
change affects size (a list grew taller, a label's text widened), call
`widgets.Relayout(w)` — the framework re-runs layout starting at the
nearest container that cares. `ui.Refresh()` repaints the whole screen
and is always correct but rarely necessary.

## Build declaratively, wire imperatively

This is the most important pattern in zeichenwerk. The teaser already used
it; now it has a name.

```go
// 1) Build the whole tree in one expression.
ui := NewBuilder(themes.TokyoNight()).
    VFlex("root", core.Stretch, 0).
        List("tasks", tasks...).
        Button("done", "Mark Done").
    End().
    Build()

// 2) Find widgets by ID and attach behaviour.
list := core.MustFind[*widgets.List](ui, "tasks")
widgets.OnActivate(core.MustFind[*widgets.Button](ui, "done"), func(_ int) bool {
    // …mutate `list`…
    return true
})

// 3) Run.
ui.Run()
```

Why this split?

- The tree describes structure and looks like the screen does. Mixing event
  closures into it makes a 30-line builder block into 200 lines of
  spaghetti.
- Handlers often need references to *other* widgets (the button's handler
  reaches the list). Building the tree first means every widget exists by
  the time you wire callbacks.
- Tests can construct the tree without running it.

`MustFind[T]` is a generic typed lookup. `core.MustFind[*widgets.List](ui,
"tasks")` returns a `*widgets.List` directly, panicking with a clear error
if the ID is wrong or the widget isn't a `*widgets.List`. There's also
`core.Find(container, id) Widget` (untyped, returns `nil` when not found)
if you want to handle absence gracefully.

## IDs are like DOM IDs

Widget IDs are how `MustFind` finds things. The framework does **not** check
uniqueness — duplicates silently let `Find` return whichever appears first
in depth-first order. Treat them like HTML element IDs: short, stable, and
unique within the UI. Empty IDs are fine for cosmetic widgets you'll never
look up (separators, spacers, decorative labels).

## Ahead

With the model in mind, the next three chapters cover the three things
every TUI developer fights with: **layout** (where things go), **styling**
(how they look), and **events** (what they do).

→ [Next: Layout](03-layout.md)

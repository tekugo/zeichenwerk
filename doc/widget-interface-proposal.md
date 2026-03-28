# Proposal: Narrowing the Widget Interface

**Status:** Proposal **Relates to:** review.md §1.1

---

## Problem

`Widget` has 22 methods. Many of them are framework-internals that application
code (event handlers, UI setup code, custom widget authors) never calls. Having
them on the public interface creates three concrete problems:

1. **Noise for consumers.** Anyone holding a `Widget` value sees `Render`,
   `Apply`, `Cursor`, `Class` in their IDE — methods they should never call.
2. **Higher barrier for custom widgets.** Implementing `Widget` requires
   satisfying all 22 methods, including several that are only meaningful to the
   render pipeline.
3. **No signal about intent.** There is no way to tell from the interface which
   methods are "use this freely" vs "only the framework touches this".

---

## Evidence: who calls what

The following table was derived by grepping for every `Widget` method called on
an _interface value_ (not `self` or a concrete type pointer).

| Method                      | Callers via `Widget` interface value                                                                                                                                          |
| --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Render(*Renderer)`         | `ui.go:329,353` draw loop; every container's `Render`                                                                                                                         |
| `Apply(*Theme)`             | `ui.go:726` `SetTheme`; `builder.go:121,288,294,300,305`                                                                                                                      |
| `Cursor() (int,int,string)` | `ui.go:494` `ShowCursor` only                                                                                                                                                 |
| `Class() string`            | **nowhere** — only `self.Class()` inside `Component.Selector`                                                                                                                 |
| `Info() string`             | **nowhere** — `inspector.go` calls a private `widgetInfo(w Widget)` helper that works only with `w.ID()`, `w.Bounds()`, etc.; `Info()` is never invoked through the interface |
| `State() string`            | `inspector.go:208`                                                                                                                                                            |
| `Style(...string)`          | `flex`, `box`, `grid`, `form`, `grow`, `dialog`, `builder`, `ui.go`, `inspector.go`                                                                                           |
| Everything else             | application code + framework (both)                                                                                                                                           |

---

## Proposed split

### `Widget` — the public interface

What application code, event handlers, and `Find`/`Traverse` callers actually
need:

```go
type Widget interface {
    // Identity & position
    ID() string
    Bounds() (int, int, int, int)
    Content() (int, int, int, int)

    // State
    Flag(name string) bool
    SetFlag(name string, value bool)
    State() string          // kept: inspector uses it; also useful for custom widgets

    // Sizing
    Hint() (int, int)
    SetHint(width, height int)

    // Layout management (needed by custom container authors)
    SetBounds(x, y, width, height int)
    SetParent(parent Container)
    Parent() Container

    // Events
    On(event string, handler Handler)
    Dispatch(source Widget, event string, data ...any) bool

    // Styling
    Style(selector ...string) *Style
    SetStyle(selector string, style *Style)

    // Runtime
    Refresh()
    Log(source Widget, level string, message string, data ...any)
}
```

That is **19 → 18 methods** after removing `Render`, `Apply`, `Cursor`, `Class`,
`Info` and keeping `State` and `Style` (both are called through the interface
today).

### `Renderable` — implemented by all widgets, used only by the framework

```go
// Renderable is satisfied by all concrete widgets.
// The rendering pipeline (UI and containers) uses this type;
// application code never needs to call these methods.
type Renderable interface {
    Render(r *Renderer)
    Apply(theme *Theme)
}
```

`ui.go` and every container's `Render` method hold and call children as
`Renderable`, not `Widget`. The two types are always both satisfied by concrete
widgets — the split is purely about which type is exposed at each call site.

### `CursorProvider` — optional, implemented only by interactive widgets

```go
// CursorProvider is an optional interface for widgets that expose a text
// cursor. Only Input, Editor, and Canvas implement a non-trivial Cursor().
// UI.ShowCursor does a type assertion: if cp, ok := ui.focus.(CursorProvider); ok { ... }
type CursorProvider interface {
    Cursor() (int, int, string)
}
```

Today every widget implements `Cursor()` but 95% return `(0, 0, "")`. Making it
optional documents the intent and removes the empty method from the base
`Component`.

### `Debuggable` — optional, for inspector and dev tooling

```go
// Debuggable is an optional interface for inspector / debug tools.
// Inspector does: if d, ok := w.(Debuggable); ok { text = d.Info() }
type Debuggable interface {
    Info() string
}
```

### `Class()` — remove from interfaces entirely

`Class()` is never called through a `Widget` (or any other) interface value. It
is only used inside `Component.Selector()` as `c.Class()` — a self-call on the
concrete type. It should remain as an unexported method or a plain method on
`Component`, but not appear on any exported interface.

---

## What changes at each call site

### `Container.Children()` and internal child storage

Containers currently store `[]Widget` and expose them via `Children() []Widget`.
After the split, the _internal_ slice type changes to `[]Renderable` so
containers can call `child.Render(r)` directly. `Children()` still returns
`[]Widget`:

```go
// Inside Flex (example):
type Flex struct {
    Component
    children []Renderable   // was []Widget
    ...
}

func (f *Flex) Children() []Widget {
    result := make([]Widget, len(f.children))
    for i, c := range f.children {
        result[i] = c.(Widget)   // safe: all Renderables are Widgets
    }
    return result
}

func (f *Flex) Render(r *Renderer) {
    f.Component.Render(r)
    for _, child := range f.children {
        child.Render(r)   // no assertion needed
    }
}
```

If the conversion in `Children()` feels awkward, an alternative is to define a
combined internal interface and store that:

```go
type widgetNode interface {
    Widget
    Renderable
}
// children []widgetNode — satisfies both, Children() returns []Widget via slice copy
```

### `UI.SetTheme` (the `Traverse` loop)

```go
// Before
Traverse(ui, func(widget Widget) bool {
    widget.Apply(theme)
    return true
})

// After — type assert is safe because all widgets are Renderable
Traverse(ui, func(widget Widget) bool {
    if r, ok := widget.(Renderable); ok {
        r.Apply(theme)
    }
    return true
})
```

Or, change `ui.layers` from `[]Container` to `[]widgetNode` (the combined
interface), then `SetTheme` can traverse without any assertion.

### `UI.DrawWidget`

```go
// Before
widget.Render(ui.renderer)

// After
if r, ok := widget.(Renderable); ok {
    r.Render(ui.renderer)
}
// or panic-safe: widget.(Renderable).Render(ui.renderer)
// (safe because every widget will satisfy Renderable)
```

### `UI.ShowCursor`

```go
// Before
cx, cy, cursor := ui.focus.Cursor()

// After — expresses intent: only some widgets provide a cursor
if cp, ok := ui.focus.(CursorProvider); ok {
    cx, cy, cursor := cp.Cursor()
    ...
}
```

This also allows removing the no-op `Cursor() (0, 0, "")` implementation from
`Component`, making it only appear on `Input`, `Editor`, `Canvas`.

### Inspector

```go
// Before
result += "State: " + w.State() + "\n"   // Widget.State() — stays as-is

// Info() if moved to Debuggable:
if d, ok := w.(Debuggable); ok {
    result += d.Info()
}
```

### Builder

`b.current` is stored as `Widget`. Builder calls `b.current.Style(...)` and
`b.current.SetStyle(...)` — both remain on `Widget`, no change. Builder calls
`widget.Apply(b.theme)` — this becomes `widget.(Renderable).Apply(b.theme)`.

---

## Migration plan

The changes can be made incrementally in three steps, each independently
compilable and testable:

**Step 1 — Remove `Class()` from Widget (zero call-site changes)**

`Class()` is never called via a `Widget` interface value. Removing it from the
`Widget` interface is a no-op at all call sites. It remains a method on
`Component`.

**Step 2 — Extract `CursorProvider` and `Debuggable`**

- Remove `Cursor()` from `Widget`. Add `CursorProvider` interface.
- Remove `Info()` from `Widget`. Add `Debuggable` interface.
- Update `UI.ShowCursor` to type-assert `CursorProvider`.
- Update `Inspector` to type-assert `Debuggable` (or use the `widgetInfo`
  helper).
- Remove the stub `Cursor() (0, 0, "")` from `Component`; keep only in `Input`,
  `Editor`, `Canvas`.

**Step 3 — Extract `Renderable`**

- Remove `Render()` and `Apply()` from `Widget`. Add `Renderable` interface.
- Change container internal child slices from `[]Widget` to `[]Renderable` (or
  to a combined `widgetNode` interface).
- Update `Children()` on all containers to return `[]Widget` by conversion.
- Update `UI.Draw`, `UI.DrawWidget`, `UI.SetTheme` to use `Renderable`.
- Update `Builder` to assert `Renderable` for `Apply` calls.

---

## Summary of interface sizes

| Interface        | Before       | After                    |
| ---------------- | ------------ | ------------------------ |
| `Widget`         | 22 methods   | 18 methods               |
| `Container`      | `Widget` + 2 | `Widget` + 2 (unchanged) |
| `Renderable`     | —            | 2 methods (new)          |
| `CursorProvider` | —            | 1 method (new, optional) |
| `Debuggable`     | —            | 1 method (new, optional) |

The 4 removed methods (`Render`, `Apply`, `Cursor`, `Class`) + 1 moved to
optional (`Info`) reduce what every custom widget author must understand as the
"contract", and remove methods that have no business being called by application
code.

---
name: zeichenwerk
description: Build terminal UIs with the zeichenwerk Go library — widgets, themes, events, and both Builder and Compose APIs.
user-invocable: false
---

# Zeichenwerk — terminal UI library for Go

Module: `github.com/tekugo/zeichenwerk`

Sub-packages:

- `core` — `Widget`, `Container`, `Theme`, `Style`, `Event`, `Handler`, alignment / flag constants, `Find`, `MustFind`, `FindAll`, `FindAt`, `Layout`, `Traverse`
- `widgets` — concrete widget types (`List`, `Button`, …), event helpers (`OnActivate`, `OnChange`, `OnKey`, `OnMouse`, …), `Redraw`, `Relayout`
- `themes` — pre-built themes: `TokyoNight()`, `Nord()`, `MidnightNeon()`, `GruvboxDark()`, `GruvboxLight()`, `Lipstick()`
- `compose` — functional alternative to the Builder
- `values` — reactive `Value[T]`, `Setter[T]` interface, `Update[T](container, id, value)`

Tutorial: [`doc/tutorial/README.md`](../../../doc/tutorial/README.md)
Full widget reference: [widgets.md](widgets.md)

---

## Two APIs — choose one per file

**Builder API** — fluent method chain:

```go
import (
    . "github.com/tekugo/zeichenwerk"
    "github.com/tekugo/zeichenwerk/core"
    "github.com/tekugo/zeichenwerk/themes"
)

ui := NewBuilder(themes.TokyoNight()).
    VFlex("root", core.Stretch, 0).
        Static("hello", "Hello, World!").
    End().
    Build()
ui.Run()
```

The Builder offers `HFlex(id, alignment, spacing)` and `VFlex(id, alignment, spacing)` shortcuts — they wrap `NewFlex(id, class, alignment, spacing)` and pre-set `FlagVertical`. There is **no** `Flex` builder method that takes a `horizontal` bool.

**Compose API** — `Option` functions:

```go
import (
    "github.com/tekugo/zeichenwerk/core"
    . "github.com/tekugo/zeichenwerk/compose"
    "github.com/tekugo/zeichenwerk/themes"
)

ui := UI(themes.TokyoNight(),
    VFlex("root", "", core.Stretch, 0,
        Static("hello", "", "Hello, World!"),
    ),
)
ui.Run()
```

Both APIs produce identical widget trees. Don't mix them in the same file.

---

## Theme

```go
import (
    . "github.com/tekugo/zeichenwerk/core"   // for NewTheme, NewStyle
    . "github.com/tekugo/zeichenwerk/themes" // for AddUnicodeBorders, AddDefaultStyles
)

t := NewTheme()
AddUnicodeBorders(t) // registers: thin, double, double-thin, round,
                     // thick, thick-thin, thick-slashed, lines, lines2
AddNerdStrings(t)    // glyphs for arrows, marks, list bullets …
AddDefaultStyles(t)  // baseline styles for every built-in widget

t.SetColors(map[string]string{
    "$bg0":  "#1a1b26",
    "$fg0":  "#c0caf5",
    "$blue": "#7aa2f7",
    // …
})
t.AddStyles(
    NewStyle("").WithColors("$fg0", "$bg0").WithMargin(0).WithPadding(0),
    NewStyle("button").WithColors("$bg0", "$blue").WithBorder("none").WithPadding(0, 2),
    NewStyle("button:focused").WithColors("$fg0", "$blue"),
)
t.SetStrings(map[string]string{
    "collapsible.expanded":  "▼ ",
    "collapsible.collapsed": "▶ ",
    // see widgets.md § Theme strings for the full registry
})
```

Style builder: `WithColors(fg, bg)`, `WithForeground`, `WithBackground`,
`WithBorder(name)`, `WithMargin(…)`, `WithPadding(…)`, `WithFont(font)`, `WithCursor(style)`.

**Selector syntax:** `"button"` · `"button.dialog"` · `"button:focused"` ·
`"button.dialog:focused"` · `"table/grid"` · `"table/grid:focused"`

Sub-part comes before class/state with a `/`. Component selectors resolve in
fallback order: exact → part → state → default. `Theme.Get` never returns nil.

---

## Events

| Constant      | Payload                       |
| ------------- | ----------------------------- |
| `EvtActivate` | `int` (index) or `string` (Combo, …) — Button always sends `0` |
| `EvtChange`   | varies — see widget entry     |
| `EvtSelect`   | `int` or `int, int` (Table cellNav) or `*TreeNode` |
| `EvtAccept`   | `string` (Typeahead suggestion accepted) |
| `EvtEnter`    | `string` (Enter pressed in Input) |
| `EvtFocus` / `EvtBlur` | —                    |
| `EvtShow` / `EvtHide` | —                     |
| `EvtClose`    | — (popup about to close)      |
| `EvtKey`      | `*tcell.EventKey`             |
| `EvtMouse`    | `*tcell.EventMouse`           |
| `EvtClick`    | — (mouse button-1 single click) |
| `EvtMode`     | `string` (Canvas mode change) |
| `EvtMove`     | `int, int` (cursor / highlight position) |
| `EvtPaste`    | `string`                      |
| `EvtHover`    | `*tcell.EventMouse`           |

Handler signature: `func(source Widget, event Event, data ...any) bool` —
return `true` to stop propagation. Handlers run in **reverse registration
order** (newest first).

```go
widget.On(widgets.EvtActivate, handler)

// Typed helpers in the widgets package — unwrap data[0] for you:
widgets.OnActivate(button, func(idx int) bool { … })
widgets.OnSelect  (list,   func(idx int) bool { … })
widgets.OnChange  (input,  func(value string) bool { … })  // string-payload only
widgets.OnEnter   (input,  func(value string) bool { … })
widgets.OnKey     (widget, func(*tcell.EventKey)  bool { … })
widgets.OnMouse   (widget, func(*tcell.EventMouse) bool { … })
```

`OnChange` only fires for `string`-valued change events (Input, Select,
Typeahead). For `Checkbox` (`bool`) or `Tabs` (`int`), use the raw form and
type-assert `data[0]`.

---

## Flags

`core.Flag` constants — pass to `widget.Flag(name)` / `widget.SetFlag(name, value)`:

| Constant          | Effect |
|-------------------|--------|
| `FlagChecked`     | Checkbox checked state |
| `FlagDisabled`    | Non-interactive; renders `:disabled` style |
| `FlagFocusable`   | Eligible for keyboard focus |
| `FlagFocused`     | Currently holds keyboard focus |
| `FlagGrid`        | Table: render inner grid lines |
| `FlagHidden`      | Invisible to renderer / hit-test / focus traversal; layout slot preserved |
| `FlagHorizontal`  | Viewport: restrict to horizontal scrolling only |
| `FlagHovered`     | Mouse cursor over widget |
| `FlagMasked`      | Input: hide text behind a mask character |
| `FlagPressed`     | Mouse button held on widget |
| `FlagReadonly`    | Prevents user edits; focus & selection still work |
| `FlagRight`       | Right-align content (Digits, Static) |
| `FlagSearch`      | List: enable incremental search-as-you-type |
| `FlagSkip`        | Excluded from Tab/Shift-Tab traversal even if focusable |
| `FlagVertical`    | Flex: vertical orientation. Viewport: restrict to vertical scrolling only |

```go
widget.SetFlag(core.FlagDisabled, true)
widget.Flag(core.FlagFocused) // → bool
```

State priority for `:state` selectors: `disabled > pressed > focused > hovered`.

---

## Widget quick-reference

| Widget       | Constructor                                                  | Key events                                  |
| ------------ | ------------------------------------------------------------ | ------------------------------------------- |
| BarChart     | `NewBarChart(id, class)`                                     | `EvtSelect(int)` `EvtActivate(int)`         |
| Box          | `NewBox(id, class, title)`                                   | —                                           |
| Breadcrumb   | `NewBreadcrumb(id, class)`                                   | `EvtSelect(int)` `EvtActivate(int)`         |
| Button       | `NewButton(id, class, text)`                                 | `EvtActivate(int)` (always 0)               |
| Canvas       | `NewCanvas(id, class, pages, w, h)`                          | `EvtChange` `EvtMove(int,int)` `EvtMode(string)` |
| Card         | `NewCard(id, class, title)`                                  | —                                           |
| Checkbox     | `NewCheckbox(id, class, text, checked)`                      | `EvtChange(bool)`                           |
| Clock        | `NewClock(id, class, interval, params...)`                   | —                                           |
| Collapsible  | `NewCollapsible(id, class, title, expanded)`                 | `EvtChange(bool)`                           |
| Combo        | `NewCombo(id, class, items)`                                 | `EvtChange(string)` `EvtActivate(string)`   |
| CRT          | `NewCRT(id, class)`                                          | —                                           |
| Deck         | `NewDeck(id, class, render, itemHeight)`                     | `EvtSelect(int)` `EvtActivate(int)`         |
| Dialog       | `NewDialog(id, class, title)`                                | `EvtClose`                                  |
| Editor       | `NewEditor(id, class)`                                       | `EvtChange`                                 |
| Filter       | `NewFilter(id, class)`                                       | `EvtChange(string)` `EvtAccept(string)`     |
| Flex         | `NewFlex(id, class, alignment Alignment, spacing int)`       | —                                           |
| Form         | `NewForm(id, class, title, data any)`                        | —                                           |
| FormGroup    | `NewFormGroup(id, class, title, horizontal, spacing)`        | —                                           |
| Grid         | `NewGrid(id, class, rows, cols, lines)`                      | —                                           |
| Grow         | `NewGrow(id, class, horizontal bool)`                        | —                                           |
| Heatmap      | `NewHeatmap(id, class, rows, cols)`                          | —                                           |
| Input        | `NewInput(id, class, params...)`                             | `EvtChange(string)` `EvtEnter(string)`      |
| List         | `NewList(id, class, items []string)`                         | `EvtSelect(int)` `EvtActivate(int)`         |
| Marquee      | `NewMarquee(id, class)`                                      | —                                           |
| Progress     | `NewProgress(id, class, horizontal)`                         | —                                           |
| Rule         | `NewHRule(class, style)` / `NewVRule(class, style)`          | —                                           |
| Scanner      | `NewScanner(id, class, width, charStyle)`                    | —                                           |
| Select       | `NewSelect(id, class, val, lbl, …)`                          | `EvtChange(string)`                         |
| Shimmer      | `NewShimmer(id, class)`                                      | —                                           |
| Shortcuts    | `NewShortcuts(id, class, pairs ...string)`                   | —                                           |
| Sparkline    | `NewSparkline(id, class)`                                    | —                                           |
| Spinner      | `NewSpinner(id, class, sequence)`                            | —                                           |
| Static       | `NewStatic(id, class, text)`                                 | —                                           |
| Styled       | `NewStyled(id, class, text)`                                 | —                                           |
| Switcher     | `NewSwitcher(id, class)`                                     | `EvtShow` `EvtHide` (when `connect=true`)   |
| Table        | `NewTable(id, class, provider, cellNav bool)`                | `EvtSelect(int,int)` `EvtActivate(int,[]string)` |
| Tabs         | `NewTabs(id, class)`                                         | `EvtChange(int)` `EvtActivate(int)`         |
| Terminal     | `NewTerminal(id, class)`                                     | —                                           |
| Text         | `NewText(id, class, lines, follow, max)`                     | —                                           |
| Tiles        | `NewTiles(id, class, render, tileW, tileH)`                  | `EvtSelect(int)` `EvtActivate(int)`         |
| Tree         | `NewTree(id, class)`                                         | `EvtSelect(*TreeNode)` `EvtActivate(*TreeNode)` `EvtChange(*TreeNode)` |
| TreeFS       | `NewTreeFS(id, class, root, dirsOnly)`                       | inherits Tree                               |
| Typeahead    | `NewTypeahead(id, class, params...)`                         | `EvtChange(string)` `EvtAccept(string)`     |
| Typewriter   | `NewTypewriter(id, class)`                                   | `EvtChange(bool)` `EvtActivate(bool)`       |
| Viewport     | `NewViewport(id, class, title)`                              | —                                           |

---

## Flex sizing rules

`Hint(w, h)` values per child:

- **Positive** → fixed cells
- **Zero** → natural `Hint()` size
- **Negative** → fractional share of remaining space (`-1` = 1 share, `-2` = 2 shares)

**To fill the available area**, at least one child must have a negative hint on
the axis of the Flex (height for vertical, width for horizontal). Without it the
Flex shrinks to the sum of its children.

**To right-align trailing content** in a horizontal Flex, place a `Spacer` with
`Hint(-1, 0)` before the right-hand widgets — it expands to absorb all remaining space:

```go
HFlex("header", core.Center, 0).
    Static("title", "App").
    Spacer().Hint(-1, 0).   // pushes everything after it to the right
    Button("quit", "Quit").
End()
```

---

## Grid sizing rules

Same hint semantics as Flex apply to both `Rows` and `Columns`:

- **Positive** → fixed cells
- **Zero** → auto-size from content
- **Negative** → fractional share of remaining space

**To fill the available area**, at least one row size and at least one column
size must be negative. A Grid where all sizes are fixed or zero will not expand.

`Cell(x, y, w, h)` (Builder) sets the position for the **next** widget. Skipping
a `Cell` call defaults to `(0, 0, 1, 1)`.

---

## Rendering is not clipped

Widget `Render` methods draw directly to the screen and are **not automatically
clipped** to the widget's bounds. A widget can draw outside its allocated area.
This is intentional for overlapping effects (borders, dropdowns, popups) but
means that oversized content from one widget can visually overwrite a neighbour.
Design widget sizes and content lengths defensively.

---

## Finding widgets after build

```go
import "github.com/tekugo/zeichenwerk/core"

ui := builder.Build()

// Untyped — returns Widget; nil if not found
btn := core.Find(ui, "my-button").(*widgets.Button)

// Typed — panics with a clear message if missing or wrong type
btn := core.MustFind[*widgets.Button](ui, "my-button")

widgets.OnActivate(btn, func(idx int) bool { /* … */ ; return true })
```

Other helpers:

- `core.FindAll[T](container) []T` — every descendant of type `T`
- `core.FindAt(container, x, y)` — deepest widget at screen coordinates
- `core.Traverse(container, fn)` — depth-first walk
- `widgets.FindRoot(widget) Root` — walk up to the `*UI` (or root container)

---

## Pushing data into widgets

`values.Update[T](container, id, value)` looks up a widget by id and calls
`Set(value)` on it if it implements `values.Setter[T]`. No type assertion
needed:

```go
values.Update(ui, "tables",  []string{"authors", "books", "loans"}) // List
values.Update(ui, "result",  widgets.NewArrayTableProvider(cols, rows)) // Table
values.Update(ui, "status",  "Ready")                                // Static (Setter[any])
```

> **Important:** `Table.Set(provider)` does **not** trigger a redraw on its
> own (unlike `List.Set`). Always call `core.Find(ui, id).Refresh()` after
> updating a Table's provider.

---

## Commands palette

```go
// Accessed via the UI, not constructed directly:
cmds := ui.Commands()
cmds.Register("File", "Save", "ctrl+s", func() { /* … */ })
cmds.Register("File", "Open", "ctrl+o", func() { /* … */ })
cmds.SetMaxItems(8)   // visible rows before scrolling
cmds.SetWidth(60)     // palette width in columns (0 = auto)
cmds.Open()           // show palette; Esc or action closes it
```

Style keys: `"commands"` `"commands/input"` `"commands/item"` `"commands/item:focused"`
`"commands/shortcut"` `"commands/shortcut:focused"` `"commands/group"`.

---

## Animation widgets

`Clock`, `Marquee`, `Progress`, `Scanner`, `Shimmer`, `Spinner`, `Typewriter`
all embed `Animation`. Start and stop them explicitly; pair with
`EvtShow`/`EvtHide` on the enclosing `Switcher` (with `connect=true`) so they
only run while visible:

```go
container.On(widgets.EvtShow, func(_ Widget, _ Event, _ ...any) bool {
    tw.Start(30 * time.Millisecond)
    return true
})
container.On(widgets.EvtHide, func(_ Widget, _ Event, _ ...any) bool {
    tw.Stop()
    return true
})
```

`widgets.Spinners` is a `map[string]string` of built-in sequences:
`"bar"` `"dots"` `"dot"` `"arrow"` `"circle"` `"bounce"` `"braille"`.

`Clock.Start()` takes no argument — the interval is fixed at construction time.

---

## CRT container

`CRT` wraps any single child and plays a power-on / power-off animation.
During normal operation it is a zero-overhead pass-through container.

```go
crt := widgets.NewCRT("root", "")
crt.Add(myRootWidget)
ui := zeichenwerk.NewUI(theme, crt)
crt.Start(16 * time.Millisecond)               // power-on animation
// later, to shut down:
crt.PowerOff(16*time.Millisecond, ui.Quit)
```

---

## Updating widget state

| Helper                     | Use when                                                  |
| -------------------------- | --------------------------------------------------------- |
| `widgets.Redraw(w)`        | Visual change only; widget's `Hint()` didn't change       |
| `widgets.Relayout(w)`      | Change affects size — re-runs layout starting at nearest container |
| `ui.Refresh()`             | Full screen repaint — always correct, rarely needed       |

Many widgets call the right helper internally (`List.Set`, `Static.Set`,
`Editor` keystrokes, …). You only need these helpers when you mutate widget
state directly.

---

## Error sentinels

```go
core.ErrScreenInit  // tcell init failure from Run(); supports errors.Is()
core.ErrChildIsNil  // nil passed to Add()
```

---

## Invariants

| Rule | Detail |
|------|--------|
| `Deck` itemHeight | ≥ 1 — panics otherwise |
| `Tiles` tile dims | Both ≥ 1 — panics otherwise |
| `Form` data | Must be pointer-to-struct |
| Single-child containers | `Box`, `Card`, `Collapsible`, `Viewport`, `Dialog`, `CRT`, `Form` — second `Add()` replaces the previous child |
| `compose.Build()` | Panics with clear message if no widget added |
| `Builder.Columns()` / `Rows()` | No-op with warning outside Grid context |
| `NewUI` | Returns `*UI` — no error return; signature is `NewUI(theme *Theme, root Container) *UI` |
| `Layout` MUST not paint | Compute bounds via `SetBounds`; no draw calls allowed |
| `Render` MUST not mutate | Geometry is already resolved; render is read-only |
| `Set(...)` on a widget | MUST never fire `EvtChange` (reserved for user-driven changes) |
| `ItemRender` `focused` param | `true` when the host widget itself has focus (use to distinguish focus from selection) |
| `Terminal` default size | 80×24 — call `SetHint` or let `Render` resize to content area |
| `Terminal.Write` | Thread-safe; call from any goroutine (pty, subprocess, test) |
| Single-child container `Add` followed by sibling | Replaces the child silently — close with `End()` before adding siblings (bug magnet for `Form`/`Box`) |
| `OnKey(ui, …)` | Currently dispatched (root-container exclusion was lifted); attach global shortcuts there |

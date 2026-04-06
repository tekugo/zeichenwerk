---
name: zeichenwerk
description: Build terminal UIs with the zeichenwerk Go library — widgets, themes, events, and both Builder and Compose APIs.
user-invocable: false
---

# Zeichenwerk — terminal UI library for Go

Module: `github.com/tekugo/zeichenwerk`
Compose sub-package: `github.com/tekugo/zeichenwerk/compose`

Full widget reference: [widgets.md](widgets.md)

---

## Two APIs — choose one per file

**Builder API** — fluent method chain:

```go
ui := NewBuilder(theme).
    Flex("root", false, "stretch", 0).
    Static("hello", "Hello, World!").
    End().
    Build()
ui.Run()
```

**Compose API** — `Option` functions:

```go
ui := compose.UI(theme,
    compose.Flex("root", "", false, "stretch", 0,
        compose.Static("hello", "", "Hello, World!"),
    ),
)
ui.Run()
```

Do not mix APIs in the same file. Both produce identical widget trees.

---

## Theme

```go
t := NewTheme()
AddUnicodeBorders(t) // registers: thin, double, double-thin, round,
                     // thick, thick-thin, thick-slashed, lines, lines2
t.SetColors(map[string]string{"$bg0": "#1a1b26", "$fg0": "#c0caf5"})
t.SetStyles(
    NewStyle("").WithColors("$fg0", "$bg0").WithMargin(0).WithPadding(0),
    NewStyle("button").WithColors("$bg0", "$blue").WithBorder("none").WithPadding(0, 2),
    NewStyle("button:focused").WithColors("$fg0", "$blue"),
)
t.SetStrings(map[string]string{
    "collapsible.expanded": "▼ ", "collapsible.collapsed": "▶ ",
    // ... see widgets.md § Theme strings
})
```

Style builder: `WithColors(fg,bg)`, `WithForeground`, `WithBackground`,
`WithBorder(name)`, `WithMargin(…)`, `WithPadding(…)`, `WithFont(font)`, `WithCursor(style)`.

**Selector syntax:** `"button"` · `"button.dialog"` · `"button:focused"` ·
`"button.dialog:focused"` · `"table/grid"` · `"table/grid:focused"`

---

## Events

| Constant | Payload |
|----------|---------|
| `EvtActivate` | `int` (index) or `0` (button) |
| `EvtChange` | varies — see widget entry |
| `EvtSelect` | `int` or `*TreeNode` |
| `EvtAccept` | `string` (typeahead suggestion) |
| `EvtShow` / `EvtHide` | — |
| `EvtMode` | `string` |
| `EvtMove` | `x, y int` |

Handler signature: `func(source Widget, event Event, data ...any) bool`
Return `true` = handled.

```go
widget.On(EvtActivate, handler)
OnKey(widget, func(e *tcell.EventKey) bool { … })   // typed helper
OnMouse(widget, func(e *tcell.EventMouse) bool { … })
// Compose equivalents: compose.On, compose.OnKey, compose.OnMouse
```

---

## Flags

| Constant | Effect |
|----------|--------|
| `FlagDisabled` | Non-interactive; renders `:disabled` style |
| `FlagFocusable` | Eligible for keyboard focus |
| `FlagFocused` | Currently holds keyboard focus |
| `FlagHidden` | Invisible and excluded from layout |
| `FlagMasked` | Hides input text (password) |
| `FlagReadonly` | Prevents user edits |
| `FlagSkip` | Excluded from Tab/Shift-Tab traversal |
| `FlagChecked` | Checkbox checked state |
| `FlagVertical` | Viewport: restrict to vertical scrolling only (child fills width) |
| `FlagHorizontal` | Viewport: restrict to horizontal scrolling only (child fills height) |

```go
widget.SetFlag(FlagDisabled, true)
widget.Flag(FlagFocused) // → bool
```

---

## Widget quick-reference

| Widget | Constructor | Key events |
|--------|-------------|------------|
| Box | `NewBox(id, class, title)` | — |
| Button | `NewButton(id, class, text)` | `EvtActivate` |
| Canvas | `NewCanvas(id, class, pages, w, h)` | `EvtChange` `EvtMove` `EvtMode` |
| Checkbox | `NewCheckbox(id, class, text, checked)` | `EvtChange(bool)` |
| Collapsible | `NewCollapsible(id, class, title, expanded)` | `EvtChange(bool)` |
| Deck | `NewDeck(id, class, render, itemHeight)` | `EvtSelect` `EvtActivate` |
| Dialog | `NewDialog(id, class, title)` | — |
| Editor | `NewEditor(id, class)` | `EvtChange` |
| Flex | `NewFlex(id, class, horizontal, alignment, spacing)` | — |
| Form | `NewForm(id, class, title, data)` | — |
| Grid | `NewGrid(id, class, rows, cols, lines)` | — |
| Input | `NewInput(id, class, params…)` | `EvtChange(string)` |
| List | `NewList(id, class, items)` | `EvtSelect` `EvtActivate` |
| Progress | `NewProgress(id, class, horizontal)` | — |
| Rule | `NewHRule(class, style)` / `NewVRule(class, style)` | — |
| Select | `NewSelect(id, class, val,lbl,…)` | `EvtChange(string)` |
| Sparkline | `NewSparkline(id, class)` | — |
| Static | `NewStatic(id, class, text)` | — |
| Styled | `NewStyled(id, class, text)` | — |
| Switcher | `NewSwitcher(id, class)` | `EvtShow` `EvtHide` |
| Table | `NewTable(id, class, provider)` | `EvtSelect` `EvtActivate` |
| Tabs | `NewTabs(id, class)` | `EvtChange` `EvtActivate` |
| Terminal | `NewTerminal(id, class)` | — |
| Text | `NewText(id, class, lines, follow, max)` | — |
| Tree | `NewTree(id, class)` | `EvtSelect` `EvtActivate` `EvtChange` |
| Typeahead | `NewTypeahead(id, class, params…)` | `EvtChange` `EvtAccept` |
| Viewport | `NewViewport(id, class, title)` | — |

---

## Flex sizing rules

Children hints in Flex layouts:
- **Positive** → fixed size
- **Zero** → natural `Hint()` size
- **Negative** → fractional share of remaining space (`-1` = 1 share, `-2` = 2 shares)

**To fill the available area**, at least one child must have a negative hint on
the axis of the Flex (height for vertical, width for horizontal). Without it the
Flex shrinks to the sum of its children.

**To right-align trailing content** in a horizontal Flex, place a `Spacer` with
`Hint(-1, 0)` before the right-hand widgets — it expands to absorb all remaining space:

```go
// Builder
Flex("header", true, "center", 0).
    Static("title", "App").
    Spacer().Hint(-1, 0).   // pushes everything after it to the right
    Button("quit", "Quit").
End()
```

---

## Grid sizing rules

Same hint semantics as Flex apply to both `Rows` and `Columns`:
- **Positive** → fixed size in characters
- **Zero** → auto-size from content
- **Negative** → fractional share of remaining space

**To fill the available area**, at least one row size and at least one column
size must be negative. A Grid where all sizes are fixed or zero will not expand.

---

## Rendering is not clipped

Widget `Render` methods draw directly to the screen and are **not automatically
clipped** to the widget's bounds. A widget can draw outside its allocated area.
This is intentional for overlapping effects (e.g. dropdowns, popups) but means
that oversized content from one widget can visually overwrite a neighbour.
Design widget sizes and content lengths defensively.

---

## Finding widgets after build

```go
ui := builder.Build()
btn := Find(ui, "my-button").(*Button)
btn.On(EvtActivate, handler)
```

---

## Error sentinels

```go
ErrScreenInit  // tcell init failure from Run(); supports errors.Is()
ErrChildIsNil  // nil passed to Add()
```

---

## Invariants

| Rule | Detail |
|------|--------|
| `Deck` itemHeight | ≥ 1 — panics otherwise |
| `Form` data | Must be pointer-to-struct |
| Single-child containers | `Box`, `Collapsible`, `Viewport`, `Dialog` — second `Add()` replaces |
| `compose.Build()` | Panics with clear message if no widget added |
| `Builder.Columns()`/`Rows()` | No-op with warning outside Grid context |
| `NewUI` | Returns `*UI` only — no error return |
| `ItemRender` `focused` param | `true` when Deck itself has focus — use to distinguish focus from selection |
| `Terminal` default size | 80×24 — call `SetHint` or let `Render` resize to content area |
| `Terminal.Write` | Thread-safe; call from any goroutine (pty, subprocess, test) |

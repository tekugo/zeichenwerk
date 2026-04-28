# Widget Flags

Flags are simple boolean state markers on every `Widget`. They drive
behaviour (focusable? hidden? read-only?) and visual state (focused?
hovered? pressed?). They're queried and mutated via:

```go
widget.Flag(core.FlagFocused) bool
widget.SetFlag(core.FlagHidden, true)
```

Flag constants live in [`core/flags.go`](../core/flags.go) as typed
`core.Flag` values — using the named type instead of bare strings
prevents typos at call sites.

## Standard flags

Listed alphabetically (matches `core/flags.go`).

| Flag | Constant | Effect |
|------|----------|--------|
| `"checked"` | `FlagChecked` | Marks a widget (e.g. `Checkbox`) as checked. |
| `"disabled"` | `FlagDisabled` | Non-interactive: skipped by focus / input; rendered with the `:disabled` style. |
| `"focusable"` | `FlagFocusable` | Eligible for keyboard focus. Widgets without this are skipped by tab-order traversal. Set in the constructor of every interactive widget (`Button`, `Input`, `List`, `Editor`, `Combo`, `Tree`, …). |
| `"focused"` | `FlagFocused` | Currently holds keyboard focus. Managed by the UI's focus system. |
| `"grid"` | `FlagGrid` | `Table`: render the inner grid lines. |
| `"hidden"` | `FlagHidden` | Hidden from rendering, mouse hit-testing (`FindAt`), and tab-order traversal — invisible from the user's perspective. **Layout bounds are preserved**, so revealing the widget later is cheap and doesn't reshuffle siblings. The idiomatic way to toggle visibility without re-parenting. |
| `"horizontal"` | `FlagHorizontal` | `Viewport`: restrict scrolling to horizontal only. |
| `"hovered"` | `FlagHovered` | Mouse cursor is over the widget. Used for hover styling. |
| `"masked"` | `FlagMasked` | `Input`: display the mask character instead of real characters (password fields). |
| `"pressed"` | `FlagPressed` | Widget is being activated (mouse button held down). |
| `"readonly"` | `FlagReadonly` | Value cannot be modified. The widget can still receive focus, scroll, and let the user select text. |
| `"right"` | `FlagRight` | Right-align content within the content area (currently `Digits`, `Static`). |
| `"search"` | `FlagSearch` | `List`: enable incremental search-as-you-type. |
| `"skip"` | `FlagSkip` | Exclude from Tab/Shift-Tab traversal even though the widget is focusable. Useful for cosmetic widgets that respond to focus but shouldn't take focus during keyboard navigation. |
| `"vertical"` | `FlagVertical` | `Flex`: lay out children top-to-bottom. `Viewport`: restrict scrolling to vertical only. |

## State priority

`Component.State()` returns the highest-priority active state for
selector resolution. The order is:

```
disabled  >  pressed  >  focused  >  hovered  >  ""  (default)
```

A `Style(":focus")` lookup falls back to the bare style if no
`:focus`-specific style is registered (see [`core/style.go`](../core/style.go)).

## Setting flags from the Builder

The Builder exposes a generic `Flag` method that takes the typed constant:

```go
.Flag(core.FlagHidden, true)
.Flag(core.FlagVertical, true)   // makes a Flex vertical
```

You can also call `widget.SetFlag(...)` directly after construction.

## Theme flags

`Theme.Flag(name string)` is a separate generic flag registry for
theme-level configuration toggles (e.g. `"debug"`, `"production"`,
`"feature_x"`). These are unrelated to widget state — they're stored on
the theme and used by widgets that want to read theme-wide settings.
Set them with `Theme.SetFlags(map[string]bool)`.

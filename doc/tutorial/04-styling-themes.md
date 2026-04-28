# 4. Styling & themes

Three layers of styling, in order of increasing scope:

1. **Inline overrides** — `.Foreground("$cyan")` on a specific widget.
2. **Classes** — like CSS classes; styled by the theme, applied to many
   widgets at once.
3. **Themes** — the palette and default styles for every widget type.

Most apps live happily in layers 1 and 3 and only reach for layer 2 when
several widgets need to look the same.

## Built-in themes

```go
themes.TokyoNight()
themes.MidnightNeon()
themes.Nord()
themes.GruvboxDark()
themes.GruvboxLight()
themes.Lipstick()
```

Each returns a fully-populated `*Theme`. Pass it to `NewBuilder(theme)` and
every widget renders with theme-appropriate colours, borders, and string
glyphs (Nerd Font icons by default; the theme also registers ASCII
fallbacks).

## Theme variables

Themes expose a small palette of named colour variables, prefixed with `$`.
Tokyo Night for example registers:

| Variable | Role |
|----------|------|
| `$bg0`, `$bg1`, `$bg2`, `$bg3` | Background shades, darkest → lightest (`$bg3` is the highlight tone) |
| `$fg0`, `$fg1`, `$fg2` | Foreground shades, brightest → muted |
| `$gray` | Lines, line numbers, decorations |
| `$blue`, `$cyan`, `$aqua`, `$magenta`, `$red`, `$orange`, `$yellow`, `$green` | Accent colours |

All built-in themes define the same variable names with different hex
values, so a UI that says `.Foreground("$cyan")` keeps working when you
swap themes. Literal colours (`"red"`, `"#ff6347"`) work too but won't
follow theme changes.

## Inline overrides

Chained after the widget they apply to:

```go
Static("title", "Dashboard").
    Foreground("$cyan").
    Background("$bg1").
    Font("bold").
    Padding(0, 2)
```

`Font(...)` accepts: `"bold"`, `"italic"`, `"underline"`, `"strikethrough"`,
or combinations like `"bold italic"`.

Most styling methods take an optional **selector** as their first argument
when you pass two strings:

```go
Button("ok", "Confirm").
    Background("$bg2").                  // default
    Background(":focus", "$blue").       // when focused
    Foreground(":focus", "$bg0")
```

The framework picks the matching style by walking selectors in this order:

```
exact match (e.g. ":focus")  →  bare match (no selector)  →  default
```

Useful selector states:

- `:focus` — widget has keyboard focus
- `:hover` — mouse is over the widget
- `:disabled` — `FlagDisabled` is set
- `:checked` — for `Checkbox`, when checked
- `:focused` — alias used in some themes; check the source if a
  highlight isn't kicking in

## Classes

Class lets you tag the next widget(s) with a CSS-like name; the theme is
free to style it.

```go
Class("header").
    Static("title", "App").
    Static("subtitle", "v1.0").
Class("").                       // reset
```

Anything you set with `.Class("name")` sticks until you set a different
class — including the empty string `""` to go back to default. Pre-styled
classes shipped by the built-in themes include `header`, `footer`,
`shortcut`, `error`, `info`, `warning`. Search the theme source if you're
unsure (`grep -r '"header"' themes/`).

You can register your own class styles on the `*Theme` after construction,
but most apps don't bother — inline overrides are usually enough.

## Borders

`.Border(style)` looks up a border style by name in the active theme.
Default theme registers:

```
"none"      "thin"      "thick"      "round"
"double"    "lines"     "halfblock"  …
```

```go
Box("card", "User Info").
    Border("round").                     // default state
    Border(":focus", "double").          // when focused
    Padding(1, 2)
```

`Box` is the simplest container that draws a border + title; any container
will draw a border if you give it one.

## Putting it together

[`examples/04-styling/main.go`](examples/04-styling/main.go) shows inline
overrides, classes, border styles, and `:focus` states. Tab between the
buttons to see the focus state in action:

```bash
go run ./doc/tutorial/examples/04-styling
```

Then change `themes.TokyoNight()` to `themes.GruvboxLight()` (or any other
built-in) and run again — every `$cyan`, `$magenta`, `$bg0` updates with
no other code change.

## When inline gets old

Two heuristics for when to climb the layer ladder:

- **3+ widgets need the same look:** make a class.
- **The whole UI's mood is wrong:** make / pick a different theme.

Custom themes are out of scope for this tutorial — start by reading
[`themes/tokyo-night.go`](../../themes/tokyo-night.go) and copy. The shape
is small: register colours, register a few class styles, register border
styles if you want non-default ones.

→ [Next: Events & focus](05-events-focus.md)

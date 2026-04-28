# 6. Widget tour

A reference table of every widget shipped in `widgets/`, with a one-liner
on what it does, the Builder method that creates it, and a pointer to the
detailed reference page.

This chapter is a catalog — skim it, mark what you need. The next chapter
takes containers seriously; the chapter after looks at writing your own
widget.

## Containers — they hold other widgets

| Builder method | Widget | One-liner | Reference |
|---|---|---|---|
| `Box(id, title)` | `Box` | Bordered container with optional title. | [box](../reference/box.md) |
| `Card(...)` | `Card` | Box with extra header style for grouped content. | — |
| `Collapsible(id, title, expanded)` | `Collapsible` | Header you can toggle to reveal/hide one child. | [collapsible](../reference/collapsible.md) |
| `Dialog(id, title)` | `Dialog` | Single-child container intended for popup layers. | [dialog](../reference/dialog.md) |
| `HFlex(id, alignment, spacing)` | `Flex` | Linear layout, horizontal. | [flex](../reference/flex.md) |
| `VFlex(id, alignment, spacing)` | `Flex` | Linear layout, vertical. | [flex](../reference/flex.md) |
| `Grid(id, rows, cols, lines)` | `Grid` | Cell-spanning table layout. | [grid](../reference/grid.md) |
| `Form(id, title, data)` | `Form` | Auto-generated form bound to a struct. | [form](../reference/form.md) |
| `Group(id, title, name, horizontal, spacing)` | `FormGroup` | Labeled cluster of form controls. | [form-group](../reference/form-group.md) |
| `Switcher(id, connect)` | `Switcher` | Shows one child at a time; `Select(i)` swaps. | [switcher](../reference/switcher.md) |
| `Tabs(id, names...)` | `Tabs` | Tab strip; usually paired with a Switcher. | [tabs](../reference/tabs.md) |
| `Viewport(id, title)` | `Viewport` | Scrollable wrapper for oversized content. | [viewport](../reference/viewport.md) |

## Input — they accept user data

| Builder method | Widget | One-liner | Reference |
|---|---|---|---|
| `Button(id, text)` | `Button` | Clickable button; fires `EvtActivate`. | [button](../reference/button.md) |
| `Checkbox(id, text, checked)` | `Checkbox` | Toggleable boolean; fires `EvtChange` (bool). | [checkbox](../reference/checkbox.md) |
| `Combo(id, items...)` | `Combo` | Free-text input with a suggestion list. | — |
| `Editor(id)` | `Editor` | Multi-line text editor (gap-buffer based). | [editor](../reference/editor.md) |
| `Filter(id)` | `Filter` | Generic filter input wired to a list/table. | — |
| `Input(id, params...)` | `Input` | Single-line text field. | [input](../reference/input.md) |
| `List(id, items...)` | `List` | Scrollable selectable list. | [list](../reference/list.md) |
| `Select(id, args...)` | `Select` | Dropdown selection. | [select](../reference/select.md) |
| `Tree(id)` | `Tree` | Expandable hierarchy. | [tree](../reference/tree.md) |
| `TreeFS(id, root, dirsOnly)` | `Tree` | Tree pre-bound to a filesystem path. | — |
| `Typeahead(id, params...)` | `Typeahead` | Input + filtered suggestions. | [typeahead](../reference/typeahead.md) |

## Display — they show data

| Builder method | Widget | One-liner | Reference |
|---|---|---|---|
| `Static(id, text)` | `Static` | Plain text label. | [static](../reference/static.md) |
| `Styled(id, text)` | `Styled` | Rich text with inline markup. | [styled](../reference/styled.md) |
| `Text(id, content, follow, max)` | `Text` | Multi-line scrollable text. | [text](../reference/text.md) |
| `Digits(id, text)` | `Digits` | Big ASCII-art digits, e.g. for clocks. | [digits](../reference/digits.md) |
| `Breadcrumb(id)` | `Breadcrumb` | Path-style segment indicator. | — |
| `Table(id, provider, cellNav)` | `Table` | Tabular data. Drives off a `TableProvider`. | [table](../reference/table.md) |
| `HRule(style)` / `VRule(style)` | `Rule` | Single-line separator. | [rule](../reference/rule.md) |
| `Spacer()` | — | Invisible flex child that swallows leftover space. | — |

## Animated — they tick

| Builder method | Widget | One-liner | Reference |
|---|---|---|---|
| `Clock(id, interval, params...)` | `Clock` | Live wall-clock display. | — |
| `Marquee(id)` | `Marquee` | Scrolling text. | — |
| `Progress(id, horizontal)` | `Progress` | Bar (determinate or indeterminate). | [progress](../reference/progress.md) |
| `Scanner(id, width, charStyle)` | `Scanner` | Back-and-forth scanning indicator. | [scanner](../reference/scanner.md) |
| `Shimmer(id)` | `Shimmer` | Skeleton-style loading shimmer. | — |
| `Sparkline(id)` | `Sparkline` | Tiny inline trend chart. | [sparkline](../reference/sparkline.md) |
| `Spinner(id, sequence)` | `Spinner` | Animated loading glyph. | [spinner](../reference/spinner.md) |
| `Typewriter(id)` | `Typewriter` | Reveals text one character at a time. | — |

## Specialised

| Builder method | Widget | One-liner | Reference |
|---|---|---|---|
| `BarChart(id)` | `BarChart` | Multi-series stacked bars. | — |
| `Canvas(id, pages, w, h)` | `Canvas` | Low-level pixel buffer for custom drawings. | [canvas](../reference/canvas.md) |
| `Deck(id, render, itemHeight)` | `Deck` | Stack of items rendered by a callback. | [deck](../reference/deck.md) |
| `Heatmap(id, rows, cols)` | `Heatmap` | Coloured cell grid for matrix data. | [heatmap](../reference/heatmap.md) |
| `Terminal(id)` | `Terminal` | Embedded terminal emulator. | — |
| `Tiles(id, render, tileW, tileH)` | `Tiles` | Wrapping grid of fixed-size tiles. | — |

## Reading the reference pages

Each reference page in `doc/reference/` follows the same shape:

- One-paragraph description.
- Constructor signature.
- Public methods you'd call after `Find` / `MustFind`.
- Events the widget dispatches, with the `data` types.
- Style selectors and parts the theme can target.

Start there once you know which widget you want.

## Patterns worth noting

A few cross-cutting patterns the table doesn't surface:

**Setter widgets accept `values.Update`.** Many widgets implement
`values.Setter[T]` for some `T`. `values.Update(ui, "tables", names)` calls
`Set([]string)` on a `*List`; `values.Update(ui, "result", provider)` calls
`Set(TableProvider)` on a `*Table`. The values package is how you get data
into a widget without first calling `MustFind` and casting yourself.

**Tables eat anything.** `widgets.NewArrayTableProvider(cols, rows)`
covers the common case; for anything else, implement the three-method
`TableProvider` interface (`Columns`, `Length`, `Str(row, col)`).

**Tabs and Switcher are usually paired.** `Tabs` renders the strip and
fires `EvtActivate` with the tab index; `Switcher` flips children when you
call `.Select(i)`. The `connect` flag on `Switcher` shows/hides children
via `EvtShow`/`EvtHide` rather than reparenting them, which keeps state
intact across switches.

**`Custom`** (in `widgets/custom.go`) is the escape hatch for one-off
visuals — you give it a render callback and skip writing a full widget
type. Good for prototypes; promote to a real widget once you start
copy-pasting it.

→ [Next: Containers in depth](07-containers.md)

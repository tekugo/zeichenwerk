# Slider

Horizontal integer range input — value clamped to `[min, max]` with arrow-key
and mouse interaction.

**Constructor:** `NewSlider(id, class string) *Slider`

Defaults: `min=0`, `max=100`, `value=0`, `step=1`. The widget is focusable.

The renderer picks the visual style from the available content height:

- **Height 1** — compact one-row style: a heavy horizontal track (`━`) with a
  heavy vertical thumb (`┃`).
- **Height ≥ 2** — centred two-row rounded box (`╭─╮`/`╰─╯`) with a double-stem
  thumb (`╥`/`╨`) piercing both rows. Extra height becomes padding above and
  below.

## Methods

- `Set(value int)` — sets the current value (clamped, dispatches `EvtChange`
  when it actually changes)
- `Value() int` — returns the current value
- `Min() int`, `Max() int`, `Step() int`
- `SetMin(int)`, `SetMax(int)` — bounds setters; current value is reclamped
- `SetStep(int)` — coarse step used by arrow keys (clamped to ≥ 1)
- `Summary() string` — `value=N [min..max]` for Dump output

## Keyboard

| Key            | Action                              |
| -------------- | ----------------------------------- |
| `←` / `h`      | Decrease by `step` (clamps at min)  |
| `→` / `l`      | Increase by `step` (clamps at max)  |
| `Home`         | Jump to `min`                       |
| `End`          | Jump to `max`                       |

## Mouse

Left-click on the slider maps the click column to a value. In box style the
inner track (excluding the rounded corners) is the mappable area, so corner
clicks resolve to the nearest bound.

## Events

| Event      | Data | Description                                    |
| ---------- | ---- | ---------------------------------------------- |
| `"change"` | `int`| Value changed (only when the value differs)    |

`Set()` is *not* silent — every value change fires `EvtChange`, whether it
came from a keystroke, mouse click, or programmatic setter.

## Theme strings

The glyphs for both styles are configurable per theme.

### Compact (height 1)

| Key                     | Default | Description                       |
| ----------------------- | ------- | --------------------------------- |
| `slider.compact.track`  | `━`     | Horizontal track glyph            |
| `slider.compact.thumb`  | `┃`     | Vertical thumb glyph              |

### Box (height ≥ 2)

| Key                         | Default | Description                       |
| --------------------------- | ------- | --------------------------------- |
| `slider.box.top-left`       | `╭`     | Top-left corner of the box        |
| `slider.box.top-right`      | `╮`     | Top-right corner of the box       |
| `slider.box.bottom-left`    | `╰`     | Bottom-left corner of the box     |
| `slider.box.bottom-right`   | `╯`     | Bottom-right corner of the box    |
| `slider.box.horizontal`     | `─`     | Horizontal border line            |
| `slider.box.thumb-top`      | `╥`     | Thumb top half (joins top border) |
| `slider.box.thumb-bottom`   | `╨`     | Thumb bottom half (joins bottom)  |

Override via `theme.SetStrings(...)` or by registering an alternate string set
(`AddUnicodeStrings`, `AddNerdStrings`).

## Style selectors

The slider has no part selectors — the entire widget (track, box border, and
thumb) renders in a single state-resolved style. To make the slider visually
"come alive" on focus, change the colours on the `:focused` selector, not on
a thumb sub-part.

| Selector                | Purpose                                            |
| ----------------------- | -------------------------------------------------- |
| `slider`                | Whole widget when not focused                      |
| `slider:disabled`       | Whole widget when disabled                         |
| `slider:focused`        | Whole widget when the widget has focus             |
| `slider:hovered`        | Hover state                                        |

## Notes

- Flags: `"focusable"`
- Default hint = `(0, 1)` — fills horizontally, one row by default. Place in a
  layout that grants height ≥ 2 to opt into the rounded box style.
- `Readonly` or `Disabled` widgets ignore keyboard and mouse input.

## Example

Builder API:

```go
NewBuilder(themes.TokyoNight()).
    Slider("volume").
    On("change", func(_ core.Widget, _ core.Event, data ...any) bool {
        fmt.Println("volume:", data[0].(int))
        return true
    })
```

Composition API:

```go
Slider("volume", "",
    Range(0, 11),
    Step(1),
    Value(7),
    On(z.EvtChange, func(_ c.Widget, _ z.Event, data ...any) bool {
        fmt.Println("volume:", data[0].(int))
        return true
    }),
)
```

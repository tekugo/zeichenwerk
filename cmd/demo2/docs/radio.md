# Radio

Vertical radio-button group — mutually-exclusive selection where every option
is shown on its own row.

**Constructor:** `NewRadio(id, class string, args ...string) *Radio`

`args` are alternating value/label pairs: `value1, label1, value2, label2, ...`
The interface mirrors [Select](select.md); the difference is purely visual:
Radio always shows the full list rather than hiding it behind a dropdown.

## Methods

- `Select(value string)` — selects option by value (silent — no `EvtChange`)
- `Text() string` — returns display label of selected option
- `Value() string` — returns value of selected option
- `Summary() string` — `selected="..."` for Dump output

## Keyboard

| Key            | Action                              |
| -------------- | ----------------------------------- |
| `Up` / `k`     | Select previous option (clamps)     |
| `Down` / `j`   | Select next option (clamps)         |
| `Home`         | Select first option                 |
| `End`          | Select last option                  |

There is no separate cursor — every navigation key changes the selection
immediately and dispatches `EvtChange`.

## Mouse

Left-click on a row selects that option.

## Events

| Event      | Data     | Description                |
| ---------- | -------- | -------------------------- |
| `"change"` | `string` | Selected value changed     |

`Select()` is silent so programmatic initialisation does not fire handlers;
keyboard and mouse interaction always fire `EvtChange`.

## Theme strings

The glyphs are configurable per theme — both can be any rune width (3 cells
for `(•)`/`( )`, 1 cell for `◉`/`○`, or Nerd Font codepoints). The widget
pads the narrower glyph so labels stay column-aligned.

| Key         | Default | Notes                         |
| ----------- | ------- | ----------------------------- |
| `radio.on`  | `(•)`   | Glyph for the selected row    |
| `radio.off` | `( )`   | Glyph for unselected rows     |

Override via `theme.SetStrings(...)` or by registering the alternate string
set (`AddUnicodeStrings` ships `◉`/`○`; `AddNerdStrings` uses Nerd Font
circle glyphs).

## Style selectors

| Selector                 | Purpose                                           |
| ------------------------ | ------------------------------------------------- |
| `radio`                  | Base row style for unselected options             |
| `radio:disabled`         | Whole widget when disabled                        |
| `radio:focused`          | Unselected rows when the widget has focus         |
| `radio:hovered`          | Hover state                                       |
| `radio/selected`         | Selected row (widget unfocused)                   |
| `radio/selected:focused` | Selected row when the widget has focus            |

## Notes

- Flags: `"focusable"`
- Hint height = number of options; width = longest label + 4 (prefix budget)
- `Readonly` or `Disabled` widgets ignore keyboard and mouse input

## Example

```go
NewBuilder(themes.TokyoNight()).
    Radio("size", "s", "Small", "m", "Medium", "l", "Large").
    On("change", func(w core.Widget, _ core.Event, data ...any) bool {
        fmt.Println("size:", data[0].(string))
        return true
    })
```

Composition API:

```go
Radio("size", "", []string{"s", "Small", "m", "Medium", "l", "Large"},
    On(z.EvtChange, func(_ c.Widget, _ z.Event, data ...any) bool {
        fmt.Println("size:", data[0].(string))
        return true
    }),
)
```

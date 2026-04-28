# Deck

Scrollable list where every item occupies a fixed number of rows. All visual content is delegated to a caller-supplied render function — no per-item widget allocations. Suited for rich multi-line cards such as colour pickers, contact cards, or navigation menus.

**Constructor:** `NewDeck(id, class string, render ItemRender, itemHeight int) *Deck`

`itemHeight` must be >= 1 (panics otherwise). `index` starts at -1.

### Render function

```go
type ItemRender func(r *Renderer, x, y, w, h, index int, data any, selected bool)
```

Called once per visible slot. `x/y/w/h` are the slot's absolute content-area coordinates. The function is responsible for all drawing within those bounds.

## Methods

- `Get() []any` — returns the current items slice
- `Set(items []any)` — replaces all items; resets index to 0 (or -1 if empty)
- `SetDisabled(indices []int)` — replaces the non-selectable index list
- `Select(index int)` — highlights item at index; adjusts scroll; dispatches `"select"`
- `Selected() int` — returns the highlighted index (-1 if none)
- `Move(count int)` — moves highlight by count (skips disabled items, clamps at bounds)
- `First()` — highlights first enabled item
- `Last()` — highlights last enabled item
- `PageUp()` — moves up by the number of fully visible slots
- `PageDown()` — moves down by the number of fully visible slots

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `int` | Enter pressed, or an already-selected item clicked again |
| `"select"` | `int` | Highlighted item changed |

## Notes

Flags: `"focusable"`

Keyboard: `↑`/`↓` move by one; `PgUp`/`PgDn` move by visible slot count; `Home`/`End` jump to first/last; `Enter` activates.

Mouse: single click selects; clicking the already-selected item activates.

Scrollbar uses row-based units (`offset × itemHeight`) so the thumb remains proportional regardless of item height.

Style selector: `"deck"` with `:focused`, `:hovered`, `:disabled` states. Item-level styling is the render function's responsibility.

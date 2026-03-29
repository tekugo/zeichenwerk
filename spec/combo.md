# Combo

A filterable list widget that combines a text input with a list. The user types to
narrow down items; navigation and activation work without leaving the input.
Intended as the popup content for the `Select` widget.

## Structure

```go
type Combo struct {
    Component
    input   *Typeahead // Filter text input with ghost-text completion (top row)
    list    *List      // Filtered results (remaining rows)
    all     []string   // Unfiltered source items
    indices []int      // Maps filtered index → original index
}
```

`indices` lets callers translate an activated filtered-list position back to the
original item index (needed by `Select` to map into its `options` slice).

The `Typeahead`'s suggest function is wired in the constructor to return the
first item in `c.all` that has the query as a case-insensitive prefix, giving
inline ghost-text completion as the user types.

## Constructor

```go
func NewCombo(id, class string, items []string) *Combo
```

- Creates a `Typeahead` (id `id+"-input"`) and a `List` (id `id+"-list"`) as
  child widgets. Both are stored as fields; they are not added to a Container —
  `Combo` renders them directly.
- Wires the `Typeahead`'s suggest function to return the first item in `c.all`
  that has the query as a case-insensitive prefix.
- Registers key handlers on both sub-widgets (see Interaction below).
- Sets `FlagFocusable` on `Combo`; the sub-widgets do **not** participate in
  normal focus cycling — `Combo` manages them internally.
- Calls `filter("")` to initialise `indices`.

## Layout & Rendering

`Hint()` returns `(maxItemWidth+2, len(items)+1)` — one extra row for the input.

`Render(r)` splits the content area:
- Row 0: calls `input.SetBounds` and `input.Render(r)`
- Rows 1…h: calls `list.SetBounds` and `list.Render(r)`

`Apply(theme)` delegates to both sub-widgets and applies the `"combo"` selector to
itself for container-level styling.

`Cursor()` delegates to `input.Cursor()` so the terminal cursor sits in the input
field while `Combo` is focused.

## Interaction

`Combo` owns a single `OnKey` handler that runs before the sub-widgets:

| Key | Behaviour |
|-----|-----------|
| Printable rune / Backspace / Delete | Forward to `input`; after input processes it call `filter(input.Value())` |
| `↓` | `list.Move(+1)` |
| `↑` | `list.Move(-1)` |
| `PgDn` | `list.PageDown()` |
| `PgUp` | `list.PageUp()` |
| `Home` | `list.First()` |
| `End` | `list.Last()` |
| `Enter` | If list is non-empty: dispatch `EvtActivate` with `indices[list.Selected()]` |
| `Esc` | Return `false` — propagate to parent (popup close) |

The `input` widget's own key handler is not registered; `Combo` forwards only the
editing keys to keep full control.

## Filtering

```go
func (c *Combo) filter(query string)
```

- Case-insensitive substring match against each item in `c.all`.
- Rebuilds `c.indices` and calls `list.SetItems(filtered)`.
- Resets list highlight to 0 after each filter update.
- Empty query shows all items.

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"activate"` | `int` | Original index of the activated item |

No `"select"` event — `Select` only needs to know when an item is confirmed.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"combo"` | `Combo` container background/border |
| `"combo/input"` (via Input's own `"input"` selector) | Filter input |
| `"combo/list"` (via List's own `"list"` selector) | Result list |

Theme entries follow existing conventions; no new theme keys required.

## Integration with `Select`

`Select.popup()` replaces its `Box + List` construction with a `Combo`:

```go
combo := NewCombo("select-combo", "popup", items)
combo.Select(s.index)          // pre-highlight current value
combo.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
    s.index = data[0].(int)    // original index, already resolved by Combo
    s.Dispatch(s, EvtChange, s.Value())
    ui.Close()
    ui.Focus(s)
    return true
})
ui.Popup(s.x, s.y+s.height, s.width, 10, combo)
```

A `Select(index int)` method on `Combo` pre-selects the highlighted item without
filtering, mirroring `List.Select`.

## Implementation Plan

1. **`combo.go`** — new file
   - Define `Combo` struct and `NewCombo`.
   - Implement `filter`, `Select`, `Apply`, `Hint`, `Render`, `Cursor`, `handleKey`.
   - Wire `Typeahead` suggest function and key handler in constructor.

2. **`builder.go`** — add `Combo` method
   ```go
   func (b *Builder) Combo(id string, items ...string) *Builder
   ```

3. **`select.go`** — update `popup()`
   - Replace `Box + List` popup with `Combo`.
   - Remove the now-redundant Escape key handler (Combo propagates Esc itself).

4. **Theme** — verify `"combo"` selector resolves gracefully (inherits from base
   style if not explicitly defined; no hard requirement on theme changes).

5. **Tests** — `combo_test.go`
   - `filter` narrows items and rebuilds `indices` correctly.
   - Arrow keys delegate to list.
   - `EvtActivate` carries the original index, not the filtered index.
   - Empty query restores full list.

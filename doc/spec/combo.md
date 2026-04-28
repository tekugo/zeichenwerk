# Combo

A traditional combo box: a single-line display of the current value that opens a popup on Enter. The popup contains a free-text [Typeahead] input and a filtered suggestion [List]. The user can type anything or pick an item from the list; either way the confirmed value is what gets submitted.

Typical use case: search fields with a history or a set of common values.

`Combo` is a standalone widget unrelated to `Select`. `Select` retains its current `Box + List` popup.

## Structure

```go
type Combo struct {
    Component
    value string   // Last confirmed value (shown in collapsed state)
    items []string // Full (unfiltered) suggestion set
}
```

The collapsed widget is a single row that renders the current value. All input and list interaction lives in the popup.

## Constructor

```go
func NewCombo(id, class string, items []string) *Combo
```

- Sets `FlagFocusable` on the widget.
- Registers a key handler: `Enter` opens the popup.

## Layout & Rendering

`Hint()` returns `(maxItemWidth+2, 1)` — always a single row.

`Render(r)` draws `c.value` in the content area.

`Apply(theme)` applies the `"combo"` selector with `"disabled"`, `"focused"`, and `"hovered"` modifiers.

## Popup

Pressing `Enter` calls `popup()`, which builds the following widget tree and passes it to `ui.Popup`:

```
Box("combo-popup", "")
  Flex("combo-popup-body", vertical=false, align="stretch", gap=0)
    Typeahead("combo-popup-input", initialValue=c.value)  [Hint(0,1)]
    List("combo-popup-list", items...)                    [Hint(0,-1)]
```

`ui.Popup` automatically focuses the first focusable widget (`combo-popup-input`).

The popup is positioned below the Combo widget (`y = c.y + c.height`). If it would overflow the terminal height, it flips above (`y = c.y - popupH`). Width matches `c.width`; height is `min(len(items), 8) + 1`, minimum 3.

After the popup is built:
- `input.SetSuggest(list.Suggest)` — ghost-text from list prefix matching.
- `OnChange(input, ...)` — calls `list.Filter(value)` and dispatches `EvtChange` on the Combo.
- `OnKey(input, ...)` — handles navigation and confirmation (see Interaction).

## Interaction

| Key | Behaviour |
|-----|-----------|
| `Enter` (collapsed) | Opens the popup |
| Printable rune / Backspace / Delete | Handled by the `Typeahead` inside the popup |
| `Tab` / `→` at end of text | Accept ghost-text suggestion (handled by `Typeahead`) |
| `↓` | `list.Move(+1)` and copy highlighted item into input |
| `↑` | `list.Move(-1)` and copy highlighted item into input |
| `PgDn` | `list.PageDown()` and copy highlighted item into input |
| `PgUp` | `list.PageUp()` and copy highlighted item into input |
| `Enter` (in popup) | Set `c.value = input.Text()`, dispatch `EvtActivate`, close popup, restore focus to Combo |
| `Esc` (in popup) | Close popup, restore focus to Combo |

Moving through the list with `↓`/`↑` copies the highlighted string into the input field via `comboPopupCopy`, matching the classic combo-box feel.

## Events

| Event | Data | Description |
|-------|------|-------------|
| `EvtChange` | `string` | Current input text, fired on every keystroke while popup is open |
| `EvtActivate` | `string` | Confirmed value when `Enter` is pressed in the popup |

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"combo"` | Collapsed widget background / border |
| `"popup"` | Popup container (applied via `ui.NewBuilder().Class("popup")`) |
| `"typeahead"` / `"typeahead/suggestion"` | Filter input inside popup (via `Typeahead.Apply`) |
| `"list"` / `"list/highlight"` | Suggestion list inside popup (via `List.Apply`) |

No new theme keys required.

## Builder

```go
func (b *Builder) Combo(id string, items ...string) *Builder
```

Creates a `NewCombo(id, b.class, items)`, calls `Apply`, and adds it to the current container.

## compose/ API

```go
func Combo(id, class string, items []string, options ...Option) Option
```

Use `On(z.EvtActivate, ...)` to receive the confirmed string value (payload is `string`, not `int`).

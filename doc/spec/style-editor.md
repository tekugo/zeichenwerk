# Style Editor

A developer panel for inspecting and live-editing the `Style` entries in a
`*Theme`. It renders as a three-column layout: a filterable list of selectors
on the left, an editable property form in the centre, and a live preview swatch
on the right. Edits are applied to the theme immediately, making any open widget
that uses the changed selector re-render with the new values.

This is a **composite widget** — it builds on existing primitives (List,
Typeahead, Input, Select, Static, Flex, Grid) rather than defining a new
rendering primitive.

---

## Visual layout

```
┌─ Style Editor ────────────────────────────────────────────────────────────┐
│ ┌─ Selectors ──────┐  ┌─ Properties ─────────────────┐  ┌─ Preview ────┐ │
│ │ Filter: [      ] │  │ Selector  button:focused      │  │             │ │
│ │ ""               │  │ Fg        [$fg0            ]  │  │  Sample     │ │
│ │ box              │  │ Bg        [$blue           ]  │  │             │ │
│ │ button           │  │ Border    [none ▼]            │  └─────────────┘ │
│ │ ▶ button:focused │  │ Font      [bold            ]  │                  │
│ │   button:hovered │  │ Margin    [0               ]  │                  │
│ │ checkbox         │  │ Padding   [0 2             ]  │                  │
│ │ checkbox:focused │  │ Cursor    [—               ]  │                  │
│ │ …                │  └──────────────────────────────┘                  │
│ └──────────────────┘                                                     │
└───────────────────────────────────────────────────────────────────────────┘
```

**Selectors column** (left, ~28 cols):
- A `Typeahead`-style filter input at the top.
- A scrollable `List` below it showing all matching selectors in
  lexicographic order, with the currently selected item highlighted.

**Properties column** (centre, ~36 cols):
- A fixed-label `FormGroup` with one `Input` per editable property.
- `Border` uses a `Select` widget with the border names available in the
  current theme.
- Fields are pre-filled from the style's own values (not inherited). Empty
  means "inherit from parent".

**Preview column** (right, remaining width):
- A `Static` widget styled on-the-fly to reflect the current property values,
  showing the word `"Sample"` centred.
- Redraws after every property change.

---

## Structure

```go
type StyleEditor struct {
    Component
    theme    *Theme
    styles   []*Style   // sorted snapshot, refreshed when theme changes
    filter   string     // current filter text
    selected *Style     // currently selected style (nil = none)
}
```

The inner layout is built with `NewBuilder` and stored in the struct. All
sub-widget pointers are captured at construction time via `Find`.

---

## Constructor

```go
func NewStyleEditor(id, class string, theme *Theme) *StyleEditor
```

1. Snapshots `theme.Styles()`, sorts by selector, stores in `se.styles`.
2. Builds the inner layout (see §Layout below).
3. Wires the filter input's `EvtChange` → `se.applyFilter`.
4. Wires the list's `EvtSelect` → `se.selectStyle`.
5. Wires each property input's `EvtChange` → `se.applyProperty`.
6. Selects the first item in the list (the `""` base style if present).
7. Sets `FlagFocusable = false` on the editor itself; focus lives inside.

---

## Internal layout (built once in constructor)

```go
NewBuilder(theme).
    Flex("se-root", true, "stretch", 0).   // horizontal
    Flex("se-left", false, "stretch", 0).Hint(28, 0).
        Input("se-filter", "…filter styles").Hint(-1, 1).
        List("se-list").Hint(-1, -1).
    End().
    FormGroup("se-props", "Properties").Hint(36, 0).
        // rows added dynamically per property
    End().
    Flex("se-preview", false, "center", 0).Hint(-1, 0).
        Static("se-preview-label", "Sample").Hint(0, 3).
    End().
    End()
```

Property rows in `se-props` are pre-created (not dynamic) with fixed IDs:

| ID | Widget | Property |
|----|--------|----------|
| `se-fg` | `Input` | Foreground (`OwnForeground()`) |
| `se-bg` | `Input` | Background (`OwnBackground()`) |
| `se-border` | `Select` | Border (options from `theme.Borders()`) |
| `se-font` | `Input` | Font string |
| `se-margin` | `Input` | Margin (space-separated: `top right bottom left`) |
| `se-padding` | `Input` | Padding (same format) |
| `se-cursor` | `Input` | Cursor style string |

---

## Methods

### Theme management

| Method | Description |
|--------|-------------|
| `SetTheme(t *Theme)` | Replaces the active theme, re-snapshots styles, resets filter and selection |
| `Refresh()` | Re-snapshots `theme.Styles()` and repopulates the list (call after programmatic style changes) |

### Filter

```go
func (se *StyleEditor) applyFilter(text string)
```

Rebuilds the list items to include only selectors where
`strings.Contains(selector, text)`. Preserves the current selection if it
still appears; otherwise selects the first visible item.

### Selection

```go
func (se *StyleEditor) selectStyle(index int)
```

1. Sets `se.selected` from the filtered list.
2. Fills property inputs from `se.selected.Own*()` values (empty string if
   unset, so the placeholder shows the inherited value).
3. Updates the preview swatch.

### Property edit

```go
func (se *StyleEditor) applyProperty(field string, value string)
```

Called when any property input fires `EvtChange`. `field` is the property
name (`"fg"`, `"bg"`, `"border"`, `"font"`, `"margin"`, `"padding"`,
`"cursor"`). Parses `value`, mutates the selected `*Style` using the
appropriate `With*` method, then:

1. Updates the preview swatch.
2. Dispatches `EvtChange` on the `StyleEditor` with payload `se.selected`.
3. Calls `Relayout(se)` so the rest of the UI picks up the change.

---

## Events

| Event | Payload | Description |
|-------|---------|-------------|
| `EvtSelect` | `*Style` | A different selector was chosen in the list |
| `EvtChange` | `*Style` | A property of the selected style was edited |

---

## Preview swatch

The preview `Static` is restyled inline on every selection or property change:

```go
preview.SetStyle("", NewStyle().
    WithColors(se.resolved("fg"), se.resolved("bg")).
    WithFont(se.selected.OwnFont()).
    WithBorder(se.selected.OwnBorder()).
    WithPadding(1, 2))
preview.Set("Sample")
Redraw(preview)
```

`se.resolved(field)` returns `theme.Color(own)` if `own` is non-empty, else
the inherited value from the parent chain. This ensures the swatch always
shows a visible colour rather than blank.

---

## Styling selectors

The `StyleEditor` itself uses only standard selectors from the theme; it
introduces no new selectors.

| Selector used | Applied to |
|---------------|-----------|
| `"flex"` | Outer container and column containers |
| `"formgroup"` | Properties column |
| `"formgroup:title"` | "Properties" label |
| `"formgroup:label"` | Property field labels |
| `"input"` / `"input:focused"` | Property value inputs |
| `"select"` / `"select:focused"` | Border drop-down |
| `"list/highlight"` / `"list/highlight:focused"` | Selected selector row |
| `"static"` | Preview label |

---

## Keyboard interaction

Focus traversal within the panel follows the standard Tab/Shift-Tab order:
filter input → list → property inputs (top to bottom) → back to filter.

| Widget | Key | Action |
|--------|-----|--------|
| Filter input | any printable | Updates filter, list scrolls to top |
| List | `↑` / `↓` | Navigate selectors |
| List | `Enter` | Confirm selection (already live) |
| Property inputs | `Enter` | Commit and move focus to next field |
| Any | `Esc` | Clear filter (if filter is focused) |

---

## Hint

```go
func (se *StyleEditor) Hint() (int, int)
```

Delegates to the inner root flex's `Hint()`. Typical minimum useful size is
**80 × 20**.

---

## Apply

```go
func (se *StyleEditor) Apply(theme *Theme)
```

Calls `Apply` recursively on all inner widgets by delegating to
`theme.Apply(se, se.Selector("style-editor"))`. Because the inner builder
used the same theme, no additional `Apply` call is needed at construction
time.

---

## Builder method

```go
func (b *Builder) StyleEditor(id string, theme *Theme) *Builder
```

```go
// Usage:
builder.StyleEditor("editor", theme).Hint(-1, -1)
editor := Find(ui, "editor").(*StyleEditor)
editor.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
    style := data[0].(*Style)
    log.Printf("changed: %s", style.Selector())
    return true
})
```

---

## Compose option

```go
func StyleEditor(id, class string, theme *Theme, options ...Option) Option
```

---

## Implementation plan

1. **`style-editor.go`** — new file
   - `StyleEditor` struct and `NewStyleEditor`.
   - `SetTheme`, `Refresh`, `applyFilter`, `selectStyle`, `applyProperty`.
   - `resolved(field)` helper for preview colour resolution.
   - `Hint`, `Apply`, `Layout` (delegates to inner root), `Render` (delegates
     to inner root).
   - `Children()` returns the inner root flex's children (makes `Find`
     traverse into it).

2. **`builder.go`** — add `StyleEditor` method.

3. **`compose/compose.go`** — add `StyleEditor` option function.

4. **`cmd/demo/main.go`** — add `"Style Editor"` item to the navigation list
   and a `styleEditorDemo` pane to the switcher. The pane passes the current
   live theme (`ui.Theme()`) to `NewStyleEditor` and re-passes on theme change
   events.

5. **Tests** — `style-editor_test.go`
   - `NewStyleEditor` populates list with all selectors.
   - `applyFilter("")` shows all selectors; `applyFilter("button")` shows only
     button variants.
   - `selectStyle` fills inputs from style's own values.
   - `applyProperty("fg", "#ff0000")` mutates the style and dispatches
     `EvtChange`.
   - `SetTheme` resets the panel state.

# Indicator

A compact display widget that pairs a coloured status glyph with a static
label — "● Online", "● Failed", "● Pending". The glyph colour is derived from
a `core.Level` value, so the six levels used by the logging and message-code
systems (Debug, Info, Success, Warning, Error, Fatal) drive the visual
treatment. Display-only: not focusable, no keyboard or mouse interaction.
Useful for dashboard rows, connection state, build status, anywhere a state
badge needs to sit inline with descriptive text.

The implementation closely mirrors the `Status` walkthrough in
[tutorial chapter 8](../tutorial/08-custom-widgets.md) — `Indicator` is the
production-grade version of that example.

---

## Visual layout

**level = `Info`, label = `"Connected"`:**

```
● Connected
```

**level = `Error`, label = `"Connection refused"`:**

```
● Connection refused
```

The glyph is one cell wide, followed by a single space, then the label. The
glyph character is themed (default `●`) and shared across all levels; the
colour changes per level.

---

## Structure

```go
type Indicator struct {
    Component
    level core.Level
    label string
}
```

`Level` is the existing `core.Level` string type — `Debug`, `Info`, `Success`,
`Warning`, `Error`, `Fatal`. No new type is introduced.

---

## Constructor

```go
func NewIndicator(id, class string, level core.Level, label string) *Indicator
```

- Stores the supplied `level` and `label`.
- Does **not** set `FlagFocusable` — display-only.
- No event handlers are registered.

A nil/empty `level` falls back to `core.Info` at render time.

---

## Methods

| Method | Description |
|--------|-------------|
| `Level() core.Level` | Returns the current level |
| `SetLevel(l core.Level)` | Updates the level; calls `Refresh()`. Does **not** dispatch `EvtChange` (no user-driven change — see [`doc/principles.md`](../principles.md)) |
| `Label() string` | Returns the current label |
| `SetLabel(s string)` | Updates the label; calls `Relayout(self)` because the natural width changes |

The hot path is `SetLevel` — labels are typically set once at construction
and left alone.

---

## State

`Indicator` overrides `State()` to return the current level as a string —
`"debug"`, `"info"`, `"success"`, `"warning"`, `"error"`, or `"fatal"`. A
zero / unknown `Level` falls back to `"info"` so an unstyled state never
exists.

```go
func (i *Indicator) State() string {
    if i.level == "" {
        return string(core.Info)
    }
    return string(i.level)
}
```

There are no other states — `Indicator` is not focusable, not interactive,
and does not honour `FlagDisabled`.

> **Note on dispatch.** Go's embedded-method binding means `Component.Render`
> sees `Component.State()` (always `""`) rather than the override, so
> `Component.Render` paints background and border using the **base**
> `indicator` style. The override is reached only by direct calls
> (`i.State()`), which is what `Render` uses to pick the level variant for
> the glyph. In practice this is what you want anyway: uniform panel chrome,
> coloured glyph.

---

## Rendering

```go
func (i *Indicator) Render(r *core.Renderer)
```

1. `i.Component.Render(r)` — paints margin, border, and background using
   the **base** `indicator` style (see the dispatch note above).
2. `x, y, w, _ := i.Content()`; bail out if `w <= 0`.
3. Look up the base style with `i.Style("")` and the level-specific style
   with `i.Style(":" + i.State())`.
4. Draw the glyph (theme string `indicator.dot`, default `●`) at `(x, y)`
   using the glyph style's foreground over the base style's background.
5. If `w >= 2`, draw the label starting at `(x+2, y)`, clipped to `w-2`
   columns, using the base style's foreground / background / font.

When `w < 2` the label is omitted; when `w == 0` nothing renders. Only the
glyph is tinted by the level — the label always uses the base `indicator`
style so prose stays readable against the surrounding panel.

---

## Hint

```go
func (i *Indicator) Hint() (int, int)
```

- Width: `2 + utf8.RuneCountInString(label)` (glyph + space + label runes).
- Height: `1`.

A Builder `.Hint(w, h)` override wins, as usual.

---

## Apply

```go
func (i *Indicator) Apply(theme *core.Theme) {
    theme.Apply(i, i.Selector("indicator"),
        "debug", "info", "success", "warning", "error", "fatal")
    if s := theme.String("indicator.dot"); s != "" {
        i.dot = s
    }
}
```

`theme.Apply` installs the base `indicator` style under `""` and one
variant per level under `":<level>"`. `Render` looks them up directly via
`i.Style(":" + i.State())`. The `indicator.dot` theme string overrides the
default `●` glyph when set.

---

## Events

None. `Indicator` is purely a display widget.

---

## Theme strings

| Key | Default | Description |
|-----|---------|-------------|
| `indicator.dot` | `●` | Glyph drawn before the label |

Alternative glyphs an application might install: `■` (square), `▲`
(triangle), `◆` (diamond), `⬤` (heavy round), `○` (hollow round).

---

## Styling selectors

The level is exposed as a state selector — `:debug`, `:info`, `:success`,
`:warning`, `:error`, `:fatal`. Same mechanism as `:focused` / `:disabled`
on interactive widgets, just driven by level instead of input state.

| Selector | Applied to |
|----------|-----------|
| `"indicator"` | Background, border, **and the label** (always — no level variant on the label) |
| `"indicator:debug"` | Glyph foreground when `level == core.Debug` |
| `"indicator:info"` | Glyph foreground when `level == core.Info` (also the unset / unknown fallback) |
| `"indicator:success"` | Glyph foreground when `level == core.Success` |
| `"indicator:warning"` | Glyph foreground when `level == core.Warning` |
| `"indicator:error"` | Glyph foreground when `level == core.Error` |
| `"indicator:fatal"` | Glyph foreground when `level == core.Fatal` |

Example theme entries (Tokyo Night):

```go
NewStyle("indicator").WithBorder("none"),
NewStyle("indicator:debug").WithForeground("$gray"),
NewStyle("indicator:info").WithForeground("$blue"),
NewStyle("indicator:success").WithForeground("$green"),
NewStyle("indicator:warning").WithForeground("$yellow"),
NewStyle("indicator:error").WithForeground("$red"),
NewStyle("indicator:fatal").WithForeground("$magenta"),
```

If a theme omits a level-specific variant, the renderer falls back to the
base `"indicator"` style — the indicator still draws, just without a
coloured glyph.

---

## Implementation plan

1. **`indicator.go`** — new file in the `widgets` package
   - Define `Indicator` struct and `NewIndicator`.
   - Implement `Apply`, `State`, `Hint`, `Render`.
   - Implement `Level`, `SetLevel`, `Label`, `SetLabel`.
   - `State()` returns `string(i.level)`, falling back to `"info"` when the
     level is empty.

2. **`builder.go`** — add an `Indicator` method
   ```go
   func (b *Builder) Indicator(id string, level core.Level, label string) *Builder
   ```

3. **Compose API** — add a matching `Indicator(id, class, level, label, opts...)`
   constructor alongside the other compose helpers.

4. **Themes** — add the `indicator.*` string keys and `"indicator"` plus the
   six `"indicator:<level>"` selectors to every built-in theme.

5. **Tests** — `indicator_test.go`
   - `Hint` width equals `2 + runeCount(label)`; height is `1`.
   - `State()` returns the level string; zero value returns `"info"`.
   - `SetLevel` updates the field and queues a redraw; no `EvtChange` fired.
   - `SetLabel` updates the field and triggers a relayout.
   - Each level renders the glyph with the foreground from
     `indicator:<level>` and the label with the base `indicator` foreground.
   - Unknown / zero-value level uses the `indicator:info` glyph style.
   - Label is clipped when content width is less than the natural width.
   - Content width `0` produces no draw calls past the component background.

6. **Demo** — add a small `Indicator` panel to `cmd/demo` so the widget is
   exercised in the live demo, with one row per level.

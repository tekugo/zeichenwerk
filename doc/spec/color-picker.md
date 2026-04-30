# Color Picker

A panel for choosing one colour, or a foreground/background pair, by editing
any of three synchronised representations (Hex, RGB, HSL). Whenever the user
changes one representation, the other two are recalculated and the live
preview swatch redraws. In foreground/background mode an additional preview
swatch shows the two colours together with the WCAG contrast ratio.

The widget intentionally has **no buttons** — confirm and cancel belong to
the container that hosts it (a `Dialog`, a side panel, a settings page, etc.).
That way the same picker can serve both modal and inline workflows.

It is a composite built from three reusable components:

| Component | Purpose |
|-----------|---------|
| `ColorPanel`   | One colour editor (3-row swatch + RGB / HSL / Hex inputs) |
| `PreviewPanel` | Foreground-on-background swatch + WCAG contrast ratio |
| `ColorPicker`  | Outer composite that arranges one or two `ColorPanel`s and an optional `PreviewPanel` |

Each is also useful on its own — a settings dialog can drop in a single
`ColorPanel`, and a theme inspector can place a bare `PreviewPanel` next to
existing fg/bg pickers.

---

## Visual layout

### `ColorPanel` (single colour editor)

```
┌─ Foreground ─────────┐
│ ████████████████████ │
│ ████████████████████ │
│ ████████████████████ │
│ ──────────────────── │
│ R 255  G 128  B  64  │
│ H  20  S 100  L  63  │
│ Hex #ff8040          │
└──────────────────────┘
```

Outer **24 × 9**, inner **22 × 7**.

| Row | Content |
|-----|---------|
| 1–3 | Preview swatch (3 rows of solid colour) |
| 4   | Horizontal rule |
| 5   | RGB row — `R nnn  G nnn  B nnn` |
| 6   | HSL row — `H nnn  S nnn  L nnn` |
| 7   | Hex row — `Hex #rrggbb` |

The numeric rows are laid out as

```
L NNNN  L NNNN  L NNNN
```

with 1-char label + 1 space + 4-col input + 2 spaces between groups
= `(1 + 1 + 4) + 2 + (1 + 1 + 4) + 2 + (1 + 1 + 4) = 22`.

The Hex row uses `Hex ` (4 chars) + 7-char `#RRGGBB` input = 11 chars; the
remaining 11 cols are left blank.

There is **no horizontal rule between RGB and HSL** — they read as one
two-row numeric block.

### `PreviewPanel` (fg / bg preview + contrast)

```
┌─ Preview ─────┐
│ ████████████  │
│ ██ Sample ██  │
│ ████████████  │
│ ───────────── │
│               │
│ Contrast 4.5  │
│               │
└───────────────┘
```

Outer **15 × 9**, inner **13 × 7**.

| Row | Content |
|-----|---------|
| 1–3 | Preview swatch — fg-on-bg with `Sample` centred |
| 4   | Horizontal rule |
| 5   | Empty (vertical centring) |
| 6   | `Contrast` + space + value (`%4.1f`, e.g. ` 4.5` or `21.0`) |
| 7   | Empty (vertical centring) |

Width: `Contrast` (8) + ` ` (1) + 4-char value = **13**.

The empty rows above and below the contrast label match the panel's height
to that of a `ColorPanel`, so the picker is rectangular without any vertical
juggling.

### `ColorPicker` (composite)

**Single-colour mode** — one panel, **24 × 9**:

```
┌─ Color ──────────────┐
│ ████████████████████ │
│ ████████████████████ │
│ ████████████████████ │
│ ──────────────────── │
│ R 255  G 128  B  64  │
│ H  20  S 100  L  63  │
│ Hex #ff8040          │
└──────────────────────┘
```

**Foreground/background mode** — two `ColorPanel`s and a `PreviewPanel`,
separated by a single empty column. Total **65 × 9**:

```
┌─ Foreground ─────────┐ ┌─ Background ─────────┐ ┌─ Preview ─────┐
│ ████████████████████ │ │ ████████████████████ │ │ ████████████  │
│ ████████████████████ │ │ ████████████████████ │ │ ██ Sample ██  │
│ ████████████████████ │ │ ████████████████████ │ │ ████████████  │
│ ──────────────────── │ │ ──────────────────── │ │ ───────────── │
│ R 255  G 128  B  64  │ │ R  16  G  16  B  32  │ │               │
│ H  20  S 100  L  63  │ │ H 240  S  33  L   9  │ │ Contrast 6.2  │
│ Hex #ff8040          │ │ Hex #101020          │ │               │
└──────────────────────┘ └──────────────────────┘ └───────────────┘
```

`24 + 1 + 24 + 1 + 15 = 65`. The gutter between panels is one empty column
(handled by the outer `Flex`'s gap).

---

## ColorPanel

### Structure

```go
type ColorPanel struct {
    Component
    title   string
    color   RGB
    suspend bool   // re-entrancy guard during cross-channel updates

    // captured at construction time via Find
    swatch *Static
    inR, inG, inB *Input
    inH, inS, inL *Input
    inHex         *Input
}

type RGB struct{ R, G, B uint8 }
```

The picker stores RGB as the source of truth. HSL is recomputed on demand
from RGB on each refresh, so round-tripping does not drift after repeated
edits.

`suspend` is `true` while the panel programmatically rewrites inputs in
response to a user edit, so the resulting `EvtChange` events on those inputs
are ignored and conversions do not loop.

### Constructor

```go
func NewColorPanel(id, class, title string) *ColorPanel
```

1. Stores `title` and defaults `color = RGB{0, 0, 0}`.
2. Builds the inner layout (see below).
3. Wires every input's `EvtChange`:
   - R/G/B → `applyRGB`
   - H/S/L → `applyHSL`
   - Hex   → `applyHex`
4. Populates inputs and swatch from `color`.
5. `FlagFocusable = false` on the panel itself; focus lives in the inputs.

### Internal layout

```go
NewBuilder(theme).
    Box("cp-{id}", title).Hint(24, 9).
        Flex("cp-{id}-body", false, "stretch", 0).
            Static("cp-{id}-swatch", "").Hint(-1, 3).
            Hrule("cp-{id}-r1").
            Flex("cp-{id}-rgb", true, "start", 2).Hint(-1, 1).
                Static("", "R").Hint(1, 1).
                Input("cp-{id}-r", "0").Hint(4, 1).
                Static("", "G").Hint(1, 1).
                Input("cp-{id}-g", "0").Hint(4, 1).
                Static("", "B").Hint(1, 1).
                Input("cp-{id}-b", "0").Hint(4, 1).
            End().
            Flex("cp-{id}-hsl", true, "start", 2).Hint(-1, 1).
                Static("", "H").Hint(1, 1).
                Input("cp-{id}-h", "0").Hint(4, 1).
                Static("", "S").Hint(1, 1).
                Input("cp-{id}-s", "0").Hint(4, 1).
                Static("", "L").Hint(1, 1).
                Input("cp-{id}-l", "0").Hint(4, 1).
            End().
            Flex("cp-{id}-hex", true, "start", 1).Hint(-1, 1).
                Static("", "Hex").Hint(3, 1).
                Input("cp-{id}-hex", "#000000").Hint(7, 1).
            End().
        End().
    End()
```

The label `Static`s have no id; only the inputs and the swatch are looked
up via `Find`.

### Methods

| Method | Description |
|--------|-------------|
| `SetColor(c string)` | Sets colour from `#RGB` or `#RRGGBB`; refreshes inputs and swatch; dispatches `EvtChange` |
| `Color() string` | Returns the current colour as `#RRGGBB` |
| `SetTitle(t string)` | Updates the box title; calls `Refresh()` |

Internal handlers:

```go
func (cp *ColorPanel) applyRGB()  // R/G/B → HSL, Hex, swatch
func (cp *ColorPanel) applyHSL()  // H/S/L → RGB, Hex, swatch
func (cp *ColorPanel) applyHex()  // Hex   → RGB, HSL, swatch
func (cp *ColorPanel) refresh()   // rewrites all inputs + swatch from cp.color
```

Each handler parses on a best-effort basis. If parsing fails, the offending
input is rendered with the `"colorpicker/input.error"` style and the rest of
the panel is left untouched.

### Hint

```go
func (cp *ColorPanel) Hint() (int, int)
```

Returns **(24, 9)** unless overridden via `SetHint`.

### Events

| Event | Payload | When |
|-------|---------|------|
| `EvtChange` | `*ColorPanel` | The panel's colour changed via input edit or `SetColor` |

---

## PreviewPanel

### Structure

```go
type PreviewPanel struct {
    Component
    fg, bg RGB

    swatch        *Static  // "Sample"
    contrastLabel *Static  // "Contrast 4.5"
}
```

### Constructor

```go
func NewPreviewPanel(id, class string) *PreviewPanel
```

Builds a 15 × 9 box with the inner layout shown below. Defaults to
`fg = RGB{0, 0, 0}`, `bg = RGB{255, 255, 255}` (contrast ≈ 21.0).

### Internal layout

```go
Box("cp-preview", "Preview").Hint(15, 9).
    Flex("cp-preview-body", false, "stretch", 0).
        Static("cp-preview-swatch", "Sample").Hint(-1, 3).
        Hrule("cp-preview-r").
        Static("", "").Hint(-1, 1).                  // empty
        Static("cp-preview-contrast", "").Hint(-1, 1).
        Static("", "").Hint(-1, 1).                  // empty
    End().
End()
```

### Methods

| Method | Description |
|--------|-------------|
| `SetColors(fg, bg RGB)` | Updates both colours, redraws the swatch, recomputes the label |
| `SetForeground(fg RGB)` | Updates fg only |
| `SetBackground(bg RGB)` | Updates bg only |
| `Contrast() float64` | Returns the current WCAG contrast ratio between fg and bg |

### Hint

Returns **(15, 9)**.

### Events

The panel does **not** emit events. It is a pure display component driven
by the parent `ColorPicker`.

---

## ColorPicker

### Structure

```go
type ColorPicker struct {
    Component
    mode    ColorPickerMode
    fg      *ColorPanel
    bg      *ColorPanel    // nil in ColorSingle
    preview *PreviewPanel  // nil in ColorSingle
}

type ColorPickerMode int

const (
    ColorSingle ColorPickerMode = iota
    ColorFgBg
)
```

### Constructor

```go
func NewColorPicker(id, class string, mode ColorPickerMode) *ColorPicker
```

1. Sets `mode`.
2. Constructs `fg = NewColorPanel("{id}-fg", class, "Foreground")`.
3. In `ColorFgBg`:
   - Constructs `bg = NewColorPanel("{id}-bg", class, "Background")`.
   - Constructs `preview = NewPreviewPanel("{id}-preview", class)`.
4. Builds the outer layout (see below).
5. Subscribes to each `ColorPanel`'s `EvtChange` to keep the preview in sync
   and to re-emit a single `EvtChange` from the picker.

### Internal layout

```go
NewBuilder(theme).
    Flex("cp-{id}", true, "stretch", 1).   // horizontal, gap = 1
        Add(cp.fg).                        // ColorPanel is itself a widget
        Add(cp.bg).                        // only in ColorFgBg
        Add(cp.preview).                   // only in ColorFgBg
    End()
```

The `ColorPanel` and `PreviewPanel` widgets are added directly — the outer
flex does not need to know about their internal structure.

### Methods

| Method | Description |
|--------|-------------|
| `SetForeground(c string)` | Delegates to `fg.SetColor(c)` |
| `SetBackground(c string)` | Delegates to `bg.SetColor(c)` (no-op in `ColorSingle`) |
| `Foreground() string` | Returns `fg.Color()` |
| `Background() string` | Returns `bg.Color()` (empty in `ColorSingle`) |
| `Mode() ColorPickerMode` | Returns the current mode |
| `SetMode(m ColorPickerMode)` | Switches modes; (re)builds children; preserves the current colours |
| `Contrast() float64` | Returns `preview.Contrast()` (or `1.0` in `ColorSingle`) |

### Hint

| Mode | Hint |
|------|------|
| `ColorSingle` | (24, 9) |
| `ColorFgBg`   | (65, 9) |

### Events

| Event | Payload | When |
|-------|---------|------|
| `EvtChange` | `*ColorPicker` | Either panel's colour changed |

The payload is the picker itself; consumers call `Foreground()`,
`Background()`, and `Contrast()` to read the new values. There is no
separate event for fg vs bg — handlers that need to distinguish should
compare against their own remembered state.

---

## Colour conversions

All conversions operate on integer RGB ∈ `[0, 255]³` and float HSL with
`H ∈ [0, 360)`, `S ∈ [0, 100]`, `L ∈ [0, 100]`. The picker exposes integer
HSL to the user (rounded for display).

### Hex → RGB

| Form | Parse rule |
|------|-----------|
| `#RGB` | Each digit is doubled (`#abc` → `#aabbcc`) |
| `#RRGGBB` | Two hex digits per channel |

Anything else is a parse error.

### RGB → HSL

```
r' = R/255    g' = G/255    b' = B/255
M  = max(r', g', b')
m  = min(r', g', b')
L  = (M + m) / 2
if  M == m            : H = 0,           S = 0
elif L < 0.5          : S = (M - m) / (M + m)
else                  : S = (M - m) / (2 - M - m)

if  M == r'           : H = (g' - b') / (M - m)
elif M == g'          : H = 2 + (b' - r') / (M - m)
else                  : H = 4 + (r' - g') / (M - m)

H = (H * 60) mod 360
```

`S` and `L` are then scaled to percentages.

### HSL → RGB

Standard `hueToRgb` formulation (see CSS Color Module 4 §6). Result is
rounded to the nearest integer per channel.

### Contrast ratio (WCAG 2.1)

```
relLuminance(R, G, B):
    for each channel c in [R/255, G/255, B/255]:
        c = c/12.92                if c <= 0.03928
        c = ((c + 0.055)/1.055)^2.4 otherwise
    return 0.2126*r + 0.7152*g + 0.0722*b

L1 = max(relLuminance(fg), relLuminance(bg))
L2 = min(relLuminance(fg), relLuminance(bg))
contrast = (L1 + 0.05) / (L2 + 0.05)
```

Formatted with `%4.1f`. Values ≥ 4.5 are styled with
`"colorpicker/contrast.ok"`, values below with `"colorpicker/contrast.warn"`,
giving the user an at-a-glance accessibility hint.

---

## Swatch rendering

Each colour swatch is a `Static` whose background is updated on every
refresh:

```go
swatch.SetStyle("", NewStyle().WithBackground(rgbToHex(panel.color)))
swatch.Set("")
Redraw(swatch)
```

The `PreviewPanel`'s swatch additionally sets the foreground:

```go
preview.swatch.SetStyle("", NewStyle().
    WithColors(rgbToHex(preview.fg), rgbToHex(preview.bg)).
    WithFont("bold"))
preview.swatch.Set("Sample")
Redraw(preview.swatch)
```

---

## Styling selectors

The components reuse standard primitives where possible and add four new
selectors for the colour-specific parts:

| Selector | Applied to |
|----------|-----------|
| `"colorpanel"` | `ColorPanel` outer box |
| `"colorpanel/swatch"` | The 3-row colour swatch in a `ColorPanel` |
| `"previewpanel"` | `PreviewPanel` outer box |
| `"previewpanel/swatch"` | The fg/bg swatch in a `PreviewPanel` |
| `"colorpicker/contrast.ok"` | Contrast label when ratio ≥ 4.5 |
| `"colorpicker/contrast.warn"` | Contrast label when ratio < 4.5 |
| `"colorpicker/input.error"` | Any RGB/HSL/Hex input whose value cannot be parsed |
| `"box"`, `"hrule"`, `"input"`, `"static"` | Standard primitives — no override |

All built-in themes (`tokyo-night`, `midnight-neon`, `nord`, `gruvbox-dark`,
`gruvbox-light`, `lipstick`) must define the new keys.

---

## Builder methods

```go
func (b *Builder) ColorPanel(id, title string) *Builder
func (b *Builder) PreviewPanel(id string) *Builder
func (b *Builder) ColorPicker(id string, mode ColorPickerMode) *Builder
```

```go
// Usage:
builder.ColorPicker("picker", ColorFgBg)
picker := Find(ui, "picker").(*ColorPicker)
picker.SetForeground("#ff8040")
picker.SetBackground("#101020")
picker.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
    cp := data[0].(*ColorPicker)
    log.Printf("fg=%s bg=%s contrast=%.1f",
        cp.Foreground(), cp.Background(), cp.Contrast())
    return true
})
```

---

## Compose options

```go
func ColorPanel(id, class, title string, options ...Option) Option
func PreviewPanel(id, class string, options ...Option) Option
func ColorPicker(id, class string, mode ColorPickerMode, options ...Option) Option
```

---

## Implementation plan

1. **`color-panel.go`** — new file
   - `ColorPanel` struct + `RGB` helper.
   - `NewColorPanel`, `SetColor`, `Color`, `SetTitle`.
   - `applyRGB`, `applyHSL`, `applyHex`, `refresh`.
   - `Hint`, `Apply`, `Layout`, `Render` delegate to the inner box.
   - `Children()` returns the inner box's children.

2. **`preview-panel.go`** — new file
   - `PreviewPanel` struct.
   - `NewPreviewPanel`, `SetColors`, `SetForeground`, `SetBackground`,
     `Contrast`.
   - `Hint`, `Apply`, `Layout`, `Render` delegate to the inner box.

3. **`color-picker.go`** — new file
   - `ColorPicker` struct + `ColorPickerMode` constants.
   - `NewColorPicker`, public delegators (`SetForeground`, `Foreground`, …),
     `SetMode`.
   - Subscribes to `fg`/`bg` `EvtChange` to keep `preview` in sync and
     re-emits a single `EvtChange`.

4. **`core/color.go`** — extend
   - Export `RGBToHSL`, `HSLToRGB`, `ParseHexColor`, `ContrastRatio`.
   - `parseHex` already covers `#RRGGBB`; add 3-digit support.

5. **`builder.go`** — add `ColorPanel`, `PreviewPanel`, `ColorPicker`
   methods.

6. **`compose/compose.go`** — add the three option functions.

7. **Themes** — add `colorpanel.*`, `previewpanel.*`, and
   `colorpicker.*` keys to all six built-in themes.

8. **`cmd/demo/main.go`** — add a `"Color Picker"` item to the navigation
   list with a pane that hosts both modes side by side and logs `EvtChange`.

9. **Tests** — `color-panel_test.go`, `preview-panel_test.go`,
   `color-picker_test.go`
   - `RGBToHSL` / `HSLToRGB` round-trip preserves the input within ±1 on
     each channel for 100 random colours.
   - `ParseHexColor("#abc")` equals `ParseHexColor("#aabbcc")`.
   - `ContrastRatio(black, white)` ≈ 21.0; `ContrastRatio(c, c)` = 1.0.
   - `ColorPanel.applyRGB` updates HSL and Hex inputs to consistent values.
   - `ColorPanel.applyHex("#ff8040")` updates R=255, G=128, B=64.
   - Editing R does not re-fire `EvtChange` from the H/S/L inputs (no loop).
   - Invalid hex sets `"colorpicker/input.error"` style on the Hex input
     and leaves the swatch unchanged.
   - `ColorPicker.SetMode(ColorFgBg)` rebuilds the layout and adds the bg +
     preview panels; previously set colours are preserved.
   - `ColorPicker` re-emits a single `EvtChange` with itself as payload
     when either child panel changes.
   - `PreviewPanel.Contrast()` matches the WCAG formula for known pairs.

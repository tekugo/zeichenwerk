# Slider

A range input widget that lets the user select a `float64` value between a
configurable minimum and maximum. The thumb `●` is positioned on a track
proportional to the current value. Works in horizontal (default) or vertical
orientation. Implements `Setter` and supports `Value[float64]` data binding.

---

## Visual layout

**Horizontal:**

```
0 ──────────●───────── 100
            42
```

**Vertical:**

```
100
 │
 │
 ●  42
 │
 │
 0
```

The track spans the available content width (horizontal) or content height
(vertical), minus the space consumed by the min/max bound labels when
`showBounds` is `true`. The value label is drawn on the row immediately below
the thumb in horizontal mode, and to the right of the thumb in vertical mode.

---

## Structure

```go
type Slider struct {
    Component
    value      float64
    min        float64
    max        float64
    step       float64  // coarse step; Arrow keys
    fineStep   float64  // fine step; Shift+Arrow
    horizontal bool
    showValue  bool
    showBounds bool
    format     string   // fmt.Sprintf verb for value and bound labels
}
```

---

## Constructor

```go
func NewSlider(id, class string) *Slider
```

- `min = 0`, `max = 100`, `value = 0`.
- `step = 10`, `fineStep = 1`.
- `horizontal = true`, `showValue = true`, `showBounds = true`.
- `format = "%.0f"`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

---

## Methods

| Method | Description |
|--------|-------------|
| `SetValue(v float64)` | Sets the current value (clamped to `[min, max]`); dispatches `EvtChange(float64)`; calls `Refresh()` |
| `Value() float64` | Returns the current value |
| `SetMin(v float64)` | Sets the minimum bound; clamps `value` if needed; calls `Refresh()` |
| `SetMax(v float64)` | Sets the maximum bound; clamps `value` if needed; calls `Refresh()` |
| `SetStep(v float64)` | Sets the coarse step used by Arrow keys |
| `SetFineStep(v float64)` | Sets the fine step used by Shift+Arrow |
| `SetHorizontal(v bool)` | Switches orientation; calls `Refresh()` |
| `SetShowValue(v bool)` | Shows or hides the value label near the thumb |
| `SetShowBounds(v bool)` | Shows or hides the min/max labels at the track ends |
| `SetFormat(f string)` | `fmt.Sprintf` format verb for value and bound labels (default `"%.0f"`) |

---

## Rendering

### Track geometry (horizontal mode)

The content area is `(cx, cy, cw, ch)`. The track occupies a single row
centred vertically:

```
trackRow = cy + ch/2
```

When `showBounds` is `true`, the formatted min and max labels flank the track
with one space of padding each:

```
minLabel + " " + track + " " + maxLabel
```

Track width and origin:

```
trackW = cw - len(minLabel) - len(maxLabel) - 2   // -2 for flanking spaces
trackX = cx + len(minLabel) + 1
```

Thumb position within the track:

```
fraction = clamp((value - min) / (max - min), 0, 1)
thumbX   = trackX + int(fraction * float64(trackW-1) + 0.5)
```

Each column in `[trackX, trackX+trackW)` is drawn with the `slider.track`
theme string using the `"slider/track"` style. The thumb overwrites its column
with the `slider.thumb` theme string, styled with `"slider/thumb:focused"` when
the widget has focus and `"slider/thumb"` otherwise.

When `showValue` is `true` and `ch > 1`, the formatted value is centred on
`thumbX` on the row immediately below `trackRow`.

### Track geometry (vertical mode)

The track occupies a single column centred horizontally:

```
trackCol = cx + cw/2
```

When `showBounds` is `true`, the max label is drawn at `cy` and the min label
at the bottom of the content area.

Track height and origin:

```
boundsRows = 2 if showBounds else 0
valueRows  = 1 if showValue  else 0
trackH     = ch - boundsRows - valueRows
trackY     = cy + (1 if showBounds else 0)
```

Thumb row (high value = near top):

```
thumbRow = trackY + int((1.0-fraction) * float64(trackH-1) + 0.5)
```

When `showValue` is `true`, the formatted value is drawn one column to the
right of the thumb on `thumbRow`.

### Render steps

```go
func (s *Slider) Render(r *Renderer)
```

1. `s.Component.Render(r)` — draws background and border.
2. Compute `fraction` and `thumbStyle` based on focus state.
3. Draw min/max labels if `showBounds`.
4. Draw all track characters with `"slider/track"` style.
5. Overwrite the thumb position with the thumb character and `thumbStyle`.
6. Draw the value label if `showValue` and space permits.

---

## Hint

```go
func (s *Slider) Hint() (int, int)
```

**Horizontal mode:**
- Width: manually set hint, or `0` (fills parent).
- Height: manually set hint, or `2` when `showValue` is `true`, `1` otherwise.

**Vertical mode:**
- Width: manually set hint, or the width of the longest bound/value label + 2
  (thumb column + space + label).
- Height: manually set hint, or `0` (fills parent).

---

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `←` / `↓` | Decrease value by `step` |
| `→` / `↑` | Increase value by `step` |
| `Shift+←` / `Shift+↓` | Decrease value by `fineStep` |
| `Shift+→` / `Shift+↑` | Increase value by `fineStep` |
| `Home` | Jump to `min` |
| `End` | Jump to `max` |

In horizontal mode `←` / `→` are the primary axis; in vertical mode `↑` / `↓`
are. The cross-axis arrows mirror their equivalents so that either pair works
in both orientations. All changes clamp to `[min, max]` and dispatch
`EvtChange(float64)`.

---

## Mouse interaction

**Click on track:** Map the click coordinate to a fraction and call `SetValue`.

```
// horizontal
fraction = float64(mouseX - trackX) / float64(trackW - 1)

// vertical
fraction = 1.0 - float64(mouseY - trackY) / float64(trackH - 1)

value = clamp(min + fraction*(max-min), min, max)
```

**Drag:** A `ButtonDown` anywhere on the track or thumb begins a drag.
Subsequent `MouseMotion` events update the value with the same formula.
The drag ends on `ButtonUp`. Each update dispatches `EvtChange(float64)`.

---

## Events

| Event | Data | Description |
|-------|------|-------------|
| `EvtChange` | `float64` | Value changed by keyboard or mouse |

---

## Value binding

`Slider` implements `Setter` and works with `Value[float64]` data bindings.
When the bound value changes externally, `SetValue` is called without
re-dispatching `EvtChange`.

---

## Theme strings

| Key | Default | Description |
|-----|---------|-------------|
| `slider.track`   | `─` | Track character (horizontal) |
| `slider.track-v` | `│` | Track character (vertical) |
| `slider.thumb`   | `●` | Thumb character |

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"slider"` | Background and border |
| `"slider/track"` | Track line |
| `"slider/thumb"` | Thumb character (unfocused) |
| `"slider/thumb:focused"` | Thumb character when the slider has focus |

Example theme entries (Tokyo Night):

```go
NewStyle("slider").WithBorder("none").WithPadding(0, 1),
NewStyle("slider/track").WithForeground("$bg3"),
NewStyle("slider/thumb").WithForeground("$blue"),
NewStyle("slider/thumb:focused").WithForeground("$cyan"),
```

---

## Implementation plan

1. **`slider.go`** — new file
   - Define `Slider` struct and `NewSlider`.
   - Implement setters: `SetValue`, `SetMin`, `SetMax`, `SetStep`, `SetFineStep`,
     `SetHorizontal`, `SetShowValue`, `SetShowBounds`, `SetFormat`.
   - Implement helpers: `fraction()`, `trackBounds()` (returns `trackX`,
     `trackW` / `trackY`, `trackH` for the current orientation and content area).
   - Implement `Hint`, `Render`, `handleKey`, `handleMouse`.
   - Implement `Setter` interface for `Value[float64]` binding.

2. **`builder.go`** — add `Slider` method
   ```go
   func (b *Builder) Slider(id string) *Builder
   ```

3. **Theme** — add `"slider"` family and `slider.*` string keys to all
   built-in themes.

4. **Tests** — `slider_test.go`
   - `SetValue` clamps to `[min, max]` at both bounds.
   - `fraction()` returns `0` at min, `1` at max, and `0.5` at the midpoint.
   - Thumb renders at the leftmost track column when `value = min`.
   - Thumb renders at the rightmost track column when `value = max`.
   - Arrow key increases / decreases value by `step`; clamps at bounds without
     dispatching additional events.
   - `Shift+Arrow` changes value by `fineStep`.
   - `Home` / `End` jump to `min` / `max`.
   - Mouse click at track start sets value to `min`; at track end to `max`.
   - `EvtChange` is dispatched with the correct `float64` on each change.
   - `showBounds = false` widens the track to the full content width.
   - Vertical mode: thumb renders at `trackY` when `value = max` and at
     `trackY + trackH - 1` when `value = min`.
   - `SetValue` via binding does not dispatch `EvtChange`.

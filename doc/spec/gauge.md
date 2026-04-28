# Gauge

A semicircular arc widget that displays a percentage value — typically used for
ratios such as cache-hit rate or budget utilisation. The arc is drawn with
Braille Unicode characters for a smooth curve. The numeric value is rendered as
text centred beneath the arc. An optional animated needle sweeps from the
previous value to the new value when the data changes, using zeichenwerk's
existing animation system.

## Visual layout

```
  ⣠⣴⣶⣿⣿⣶⣦⣄
 ⣼⡿       ⢿⣧
⣸⡟           ⢻⣇
⣿⡇   72 %   ⢸⣿
⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛
```

Fixed dimensions for a standard gauge:
- Width: 13 characters
- Height: 5 rows

The arc spans 180° from the left end of the baseline to the right end. The
filled portion represents the current value; the unfilled portion uses a dimmed
style. The value label is centred in the interior on the bottom row of content.

## Structure

```go
type Gauge struct {
    Component
    value     float64   // current value in [0, 1]
    prevValue float64   // value at start of current animation
    label     string    // optional override text (default: formatted percentage)
    animated  bool      // whether needle animation is enabled
    anim      *Animation
}
```

## Constructor

```go
func NewGauge(id, class string) *Gauge
```

- `value = 0`, `prevValue = 0`, `animated = true`.
- Not focusable (display-only widget).

## Methods

| Method | Description |
|--------|-------------|
| `SetValue(v float64)` | Sets the target value (clamped to [0, 1]); starts sweep animation when `animated` is true; calls `Refresh()` |
| `Value() float64` | Returns the current target value |
| `SetLabel(s string)` | Overrides the centre label text; empty string restores the default percentage |
| `SetAnimated(b bool)` | Enables or disables the sweep animation |

## Animation

When `animated` is true and `SetValue` changes the value, a sweep animation
runs from `prevValue` to the new `value` using the existing `Animation` type.
The animation duration is `200 ms` with an ease-out curve. While the animation
is running, `Render` uses the interpolated position to determine how much of
the arc to fill. The label always shows the final target value, not the
interpolated position.

When `animated` is false, `SetValue` updates immediately with no interpolation.

## Rendering

### Arc geometry

The gauge occupies a bounding box of width `w` and height `h = ceil(w/2) + 1`
(the +1 accounts for the flat baseline and label row). The semicircle arc is
mapped onto a Braille character grid where each cell is 2×4 Braille dots
(`2 dots wide × 4 dots tall`), giving `w` Braille columns and `h` Braille rows.

For each Braille dot position `(dotX, dotY)`:
1. Compute the angle `θ` of the dot's position relative to the arc centre.
2. The dot is **inside the arc** if its distance from the centre is within the
   stroke width and its angle is in `[0°, 180°]`.
3. The dot is **filled** if it is inside the arc and `θ ≤ filledAngle`, where
   `filledAngle = animatedValue * 180°`.
4. Assemble Braille characters cell by cell; apply `"gauge/filled"` or
   `"gauge/empty"` style per character.

### Label

The centre text is rendered at the vertical midpoint of the interior. Default
format: `"%d %%"` of `int(value * 100 + 0.5)`. A custom label set via
`SetLabel` replaces this text entirely. The label is centred horizontally
within the arc's interior width and uses the `"gauge/label"` style.

```go
func (g *Gauge) Render(r *Renderer)
```

1. `g.Component.Render(r)` — draws background and border.
2. Compute content area `(cx, cy, cw, ch)`.
3. Determine interpolated arc fill fraction from the running animation (or
   `g.value` directly when not animating).
4. Render the Braille arc, filled and unfilled portions with their respective
   styles.
5. Render the label centred on row `cy + ch - 2`.

## Hint

```go
func (g *Gauge) Hint() (int, int)
```

- Width: manually set hint, or `13` (default standard size).
- Height: `ceil(hintWidth/2) + 1`.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"gauge"` | Background and border |
| `"gauge/filled"` | Arc segment from 0 to the current value |
| `"gauge/empty"` | Arc segment from the current value to 100 % |
| `"gauge/label"` | Centre value text |

## Implementation plan

1. **`gauge.go`** — new file
   - Define `Gauge` struct and `NewGauge`.
   - Implement `SetValue`, `Value`, `SetLabel`, `SetAnimated`.
   - Implement Braille arc geometry helpers: dot-to-angle mapping, arc
     membership test, Braille cell assembly.
   - Implement `Apply`, `Hint`, `Render` with animation interpolation and
     label centering.

2. **`builder.go`** — add `Gauge` method
   ```go
   func (b *Builder) Gauge(id string) *Builder
   ```

3. **Theme** — add `"gauge"`, `"gauge/filled"`, `"gauge/empty"`, and
   `"gauge/label"` style entries to built-in themes.

4. **Tests** — `gauge_test.go`
   - `SetValue(0)` produces an entirely empty arc.
   - `SetValue(1)` produces an entirely filled arc.
   - `SetValue(0.5)` fills exactly the left half of the arc.
   - Label defaults to the correct percentage string.
   - Custom label overrides the default.
   - With `animated = false`, the rendered fraction equals `value` immediately.
   - Hint height is `ceil(w/2) + 1` for various widths.

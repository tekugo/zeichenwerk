# Shimmer

An `Animation`-driven widget that displays text with a highlight band sweeping
continuously from left to right — the "working" effect used by tools like Claude
Code to signal ongoing activity. Characters inside the band are rendered in a
bright accent colour; characters outside it use the base style. The band
advances one column per tick and wraps back to the left edge seamlessly. Calling
`Stop()` freezes the text in its base style. Suitable for progress labels,
loading placeholders, and skeleton screens.

---

## Visual layout

**Single-line, band at column 10, bandWidth = 8 (accent region marked with `█`):**

```
Analysing co█████████ase…
            ^         ^
            bandPos   bandPos+bandWidth
```

**With soft edges — brightness falls off at the band boundaries:**

```
Analysing codebase…
          ░▒▓████▓▒░
```

The soft edge spans `edgeWidth` columns on each side of the hard band,
blending from the base foreground to the accent foreground using `lerpColor`.

**Multi-line — band sweeps the same column on every row simultaneously:**

```
Searching for references…
Processing matched file▓█▓░
Updating cross-reference░
```

**Stopped (base style only):**

```
Analysing codebase…
```

---

## Structure

```go
type Shimmer struct {
    Animation
    text      string     // displayed text; may contain \n
    lines     []string   // text split at \n (updated in SetText)
    maxWidth  int        // display-column width of the longest line
    bandPos   int        // leftmost column of the band [0, maxWidth)
    bandWidth int        // number of columns in the bright core of the band
    edgeWidth int        // columns of gradient fade on each side of the band
}
```

---

## Constructor

```go
func NewShimmer(id, class string) *Shimmer
```

Defaults:

- `bandWidth = 6`, `edgeWidth = 3`.
- `FlagFocusable` is **not** set.
- Sets `sh.fn = sh.Tick`.
- Does not start the animation.

---

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetText(s string)` | Replaces text; splits into `lines`; recomputes `maxWidth`; resets `bandPos = 0`; calls `Redraw(sh)` |
| `Text() string` | Returns the current text |

### Display

| Method | Description |
|--------|-------------|
| `SetBandWidth(n int)` | Core highlight width in columns; clamped to minimum 1 |
| `SetEdgeWidth(n int)` | Gradient columns on each side; 0 = hard edge |

### Animation control

| Method | Description |
|--------|-------------|
| `Start(interval time.Duration)` | Inherited from `Animation` |
| `Stop()` | Inherited; `bandPos` is preserved so the band resumes from where it stopped |
| `Running() bool` | Inherited |

---

## Tick

```go
func (sh *Shimmer) Tick()
```

1. If `maxWidth == 0`, return.
2. `sh.bandPos = (sh.bandPos + 1) % sh.maxWidth`.
3. `Redraw(sh)`.

The full sweep width — including both edge regions — is
`bandWidth + 2 * edgeWidth`. Because `bandPos` wraps on `maxWidth` alone,
the leading and trailing edges can simultaneously be visible on opposite
sides of the text at the wrap boundary, which gives a seamless loop.

---

## Hint

```go
func (sh *Shimmer) Hint() (int, int)
```

- **Width**: manual override if set; otherwise `maxWidth` plus style
  horizontal overhead.
- **Height**: manual override if set; otherwise line count plus style
  vertical overhead.

`maxWidth` is computed from the full text at `SetText` time.

---

## Render

```go
func (sh *Shimmer) Render(r *Renderer)
```

1. `sh.Component.Render(r)` — background and border.
2. Obtain `(cx, cy, cw, ch)` from `sh.Content()`.
3. Resolve `baseStyle = sh.Style()`, `bandStyle = sh.Style("band")`.
4. Pre-compute resolved foreground colours:
   - `baseFg = r.theme.Color(baseStyle.Foreground())`
   - `bandFg = r.theme.Color(bandStyle.Foreground())`
5. For each line `i` (up to `ch`):
   For each display column `c` in `[0, cw)`:
   - Compute the character at column `c` in `lines[i]` (or `" "` if past
     line end).
   - Compute `intensity` — how much of the band colour to apply at column `c`:

```
// Columns relative to band position, with wrap-around:
dist = min((c - bandPos + maxWidth) % maxWidth,
           (bandPos - c + maxWidth) % maxWidth)

if dist <= edgeWidth:
    // in leading or trailing gradient
    intensity = 1.0 - float64(dist) / float64(edgeWidth + 1)
elif dist <= edgeWidth + bandWidth / 2:
    intensity = 1.0   // inside core band
else:
    intensity = 0.0   // outside band
```

   - Resolved foreground: `lerpColor(baseFg, bandFg, intensity)`.
   - `r.Set(fg, baseStyle.Background(), baseStyle.Font())`.
   - `r.Put(cx+c, cy+i, ch)`.

The `lerpColor` utility already exists in `color.go`. When `edgeWidth == 0`,
`intensity` is either 0 or 1 (hard edge, no blending needed).

---

## Events

The Shimmer dispatches no events. It is a pure display widget.

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"shimmer"` | Base text colour and widget background |
| `"shimmer/band"` | Foreground colour at full band intensity |

Only the **foreground** of `"shimmer/band"` is used; background and font are
taken from `"shimmer"` for all cells so the background never flashes.

Example theme entries (Tokyo Night):

```go
NewStyle("shimmer").WithColors("$fg1", "$bg0"),
NewStyle("shimmer/band").WithForeground("$fg0"),
```

For a more dramatic glow effect, use a saturated accent colour for the band:

```go
NewStyle("shimmer/band").WithForeground("$cyan"),
```

---

## Builder usage

```go
builder.Shimmer("status").Hint(-1, 1)

sh := builder.Find("status").(*Shimmer)
sh.SetText("Analysing codebase…")
sh.SetBandWidth(8).SetEdgeWidth(4)
sh.Start(30 * time.Millisecond)

// Stop and update when work is done:
sh.Stop()
sh.SetText("Done.")
```

`Hint(-1, 1)` fills the parent width and takes exactly one row.

---

## Implementation plan

1. **`shimmer.go`** — new file
   - Define `Shimmer` struct embedding `Animation`.
   - Implement `NewShimmer(id, class string) *Shimmer`.
   - Implement `SetText`, `Text`, `SetBandWidth`, `SetEdgeWidth`.
   - Implement `Tick()`.
   - Implement `Apply(t *Theme)`: register `"shimmer"` and `"shimmer/band"`
     selectors.
   - Implement `Hint() (int, int)`.
   - Implement `Render(r *Renderer)` with per-column intensity computation
     and `lerpColor` blending.

2. **`builder.go`** — add `Shimmer` method
   ```go
   func (b *Builder) Shimmer(id string) *Builder
   ```

3. **Theme** — add `"shimmer"` and `"shimmer/band"` style entries to all
   built-in themes.

4. **`cmd/demo/main.go`** — add `"Shimmer"` entry with `shimmerDemo`, showing
   single- and multi-line variants with controls for band width and edge
   softness, plus a toggle to start/stop the animation.

5. **Tests** — `shimmer_test.go`
   - `Tick` advances `bandPos` by 1 and wraps at `maxWidth`.
   - `intensity` is 1.0 inside the core band, 0.0 beyond edge region.
   - `lerpColor` is called with the correct `t` value at each edge column.
   - `SetText` recomputes `maxWidth` from the longest line.
   - Multi-line text: each line is rendered independently; band position is
     the same column index on all rows.
   - `edgeWidth = 0` produces a hard cut with no `lerpColor` calls.

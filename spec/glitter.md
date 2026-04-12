# Glitter

An `Animation`-driven widget where individual characters twinkle independently
at random phase offsets rather than following a single sweeping band. Each
character cycles through a small palette of accent colours, producing a
sparkling effect. Density controls what fraction of characters are lit at any
moment; palette controls the colour sequence each character cycles through.
Suitable for celebratory moments, splash screens, and ambient decoration.
Can serve as a direct drop-in replacement for `Static` when animation is wanted.

---

## Visual layout

**`text = "  zeichenwerk  "`, density = 0.3, 4-colour palette, tick N:**

```
  z e i c h e n w e r k  
    ^   ^         ^
    │   │         └── palette[2] (cyan)
    │   └──────────── palette[0] (blue)
    └──────────────── palette[3] (fg0 — dim sparkle)
```

At each tick a different subset of characters glows; the others render in the
base `"glitter"` style. Characters transition smoothly through the palette
before returning to the base style.

**Multi-line:**

```
★  W e l c o m e  ★
   t o  z e i c h e n w e r k
```

Each character has an independent phase, so multi-line text sparkles uniformly
across all rows.

---

## Structure

```go
type Glitter struct {
    Animation
    text        string    // displayed text; may contain \n
    runes       []rune    // flat rune slice (no \n); updated in SetText
    lineBreaks  []int     // indices in runes where a new line begins
    phases      []float64 // per-rune phase offset in [0, 1); randomised at SetText
    palette     []string  // resolved colour strings ($-refs allowed)
    density     float64   // fraction of runes active at any tick [0, 1]
    globalPhase float64   // advances by speed each tick; wraps at 1.0
    speed       float64   // phase advancement per tick (default 0.05)
}
```

`phases` is allocated and randomised once in `SetText` using `math/rand`.
`palette` stores raw colour strings as supplied; theme `$`-refs are resolved
via `r.theme.Color(s)` at render time.

---

## Constructor

```go
func NewGlitter(id, class string) *Glitter
```

Defaults:

- `density = 0.25`, `speed = 0.05`.
- `palette = []string{}` — empty; the widget renders in base style only until
  `SetPalette` is called.
- `FlagFocusable` is **not** set.
- Sets `gl.fn = gl.Tick`.
- Does not start the animation.

---

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetText(s string)` | Replaces text; builds `runes` and `lineBreaks`; allocates and randomises `phases`; resets `globalPhase = 0`; calls `Redraw(gl)` |
| `Text() string` | Returns the current text |

### Display

| Method | Description |
|--------|-------------|
| `SetPalette(colors ...string)` | Replaces the palette; `$`-prefixed theme refs are accepted; calls `Redraw(gl)` |
| `SetDensity(d float64)` | Fraction of runes active per tick; clamped to `[0, 1]` |
| `SetSpeed(s float64)` | Phase advancement per tick; clamped to `(0, 1]`; higher = faster cycling |

### Animation control

| Method | Description |
|--------|-------------|
| `Start(interval time.Duration)` | Inherited from `Animation` |
| `Stop()` | Inherited; `globalPhase` and `phases` are preserved |
| `Running() bool` | Inherited |

---

## Phase model

Each rune `i` has a fixed random phase offset `phases[i] ∈ [0, 1)` assigned
once at `SetText`. The global phase `globalPhase` advances by `speed` on every
tick and wraps at 1.0.

The **local phase** for rune `i` at a given tick is:

```
localPhase = fmod(globalPhase + phases[i], 1.0)
```

The rune is **active** (twinkling) when `localPhase < density`. While active,
its palette colour index is:

```
colorIdx = int(localPhase / density * float64(len(palette)))
           // clamped to [0, len(palette)-1]
```

This maps the active portion of the cycle (`[0, density)`) uniformly across
the palette, so each character smoothly cycles through all palette colours
during its active window before returning to the base style.

**Effect of parameters:**

| Parameter | Effect |
|-----------|--------|
| Low `density` (0.1) | Few characters lit at once; sparse sparkle |
| High `density` (0.6) | Most characters lit; heavy shimmer |
| Low `speed` (0.02) | Slow, lazy twinkle |
| High `speed` (0.15) | Rapid, energetic sparkle |
| Large palette | Characters cycle through more colours per active window |
| Small palette (1 colour) | Binary on/off flash |

Because each rune has a unique `phases[i]`, characters with similar phase
values activate close together in time, producing emergent "waves" that
vary across runs due to the random initialisation.

---

## Tick

```go
func (gl *Glitter) Tick()
```

1. `gl.globalPhase = math.Mod(gl.globalPhase + gl.speed, 1.0)`.
2. `Redraw(gl)`.

All per-character colour decisions are deferred to `Render`. `Tick` has no
per-rune work and no allocations.

---

## Hint

```go
func (gl *Glitter) Hint() (int, int)
```

- **Width**: manual override if set; otherwise display-column width of the
  longest line, plus style horizontal overhead.
- **Height**: manual override if set; otherwise line count, plus style
  vertical overhead.

Computed from `runes` and `lineBreaks` at `SetText` time and cached.

---

## Render

```go
func (gl *Glitter) Render(r *Renderer)
```

1. `gl.Component.Render(r)` — background and border.
2. Obtain `(cx, cy, cw, ch)` from `gl.Content()`.
3. Resolve `baseStyle = gl.Style()`.
4. Take a **snapshot** of `gl.globalPhase` (single read — no lock needed; a
   torn read costs at most one frame of visual inconsistency).
5. Walk `gl.runes` line by line using `gl.lineBreaks`, up to `ch` rows and
   `cw` columns per row:

```
row, col = 0, 0
for i, ru in gl.runes:
    if lineBreak at i:
        row++; col = 0; continue
    if row >= ch or col >= cw: break

    localPhase = fmod(phase + phases[i], 1.0)
    if len(palette) > 0 and localPhase < density:
        cidx  = clamp(int(localPhase / density * len(palette)), 0, len(palette)-1)
        fg    = r.theme.Color(palette[cidx])
    else:
        fg    = baseStyle.Foreground (resolved)

    r.Set(fg, baseStyle.Background(), baseStyle.Font())
    r.Put(cx+col, cy+row, string(ru))
    col++
```

`r.theme.Color` resolves `$`-prefixed palette entries against the current
theme's colour registry at render time, so palette changes respect live theme
switching.

---

## Events

The Glitter dispatches no events. It is a pure display widget.

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"glitter"` | Background, base foreground, and font for all characters |

Per-character colour comes from the palette slice, not from sub-part selectors.
The background and font are always taken from `"glitter"` regardless of whether
a character is currently active.

Example theme entries (Tokyo Night):

```go
NewStyle("glitter").WithColors("$fg0", "$bg0"),
```

Palette colours are set programmatically and are intentionally not in the theme,
since they are data rather than style:

```go
gl.SetPalette("$blue", "$cyan", "$magenta", "$fg0")
```

---

## Builder usage

```go
builder.Glitter("title")

gl := builder.Find("title").(*Glitter)
gl.SetText("  zeichenwerk  ").
    SetDensity(0.3).
    SetPalette("$blue", "$cyan", "$purple", "$fg0")
gl.Start(80 * time.Millisecond)
```

For a subtle ambient effect on a label:

```go
gl.SetDensity(0.1).SetSpeed(0.02).SetPalette("$fg0", "$fg1")
gl.Start(120 * time.Millisecond)
```

For an energetic celebratory burst, stop the animation after a fixed time:

```go
gl.Start(50 * time.Millisecond)
time.AfterFunc(3*time.Second, gl.Stop)
```

---

## Implementation plan

1. **`glitter.go`** — new file
   - Define `Glitter` struct embedding `Animation`.
   - Implement `NewGlitter(id, class string) *Glitter`.
   - Implement `SetText`: build `runes`, `lineBreaks`, randomise `phases`
     via `math/rand.Float64()`, compute `maxWidth` and line count.
   - Implement `SetPalette`, `SetDensity`, `SetSpeed`.
   - Implement `Tick()`: advance `globalPhase`, call `Redraw(gl)`.
   - Implement `Apply(t *Theme)`: register `"glitter"` selector.
   - Implement `Hint() (int, int)`.
   - Implement `Render(r *Renderer)` with per-rune phase evaluation and
     palette lookup using `r.theme.Color`.

2. **`builder.go`** — add `Glitter` method
   ```go
   func (b *Builder) Glitter(id string) *Builder
   ```

3. **Theme** — add `"glitter"` style entry to all built-in themes.

4. **`cmd/demo/main.go`** — add `"Glitter"` entry with `glitterDemo`, showing
   a title string with configurable density and speed (via `Value` widgets),
   a palette selector (preset combinations), and start/stop controls.

5. **Tests** — `glitter_test.go`
   - `SetText` allocates one phase per rune and randomises them in `[0, 1)`.
   - `Tick` advances `globalPhase` by `speed` and wraps at 1.0.
   - `localPhase` for rune `i` equals `fmod(globalPhase + phases[i], 1.0)`.
   - `colorIdx` maps `localPhase / density` to `[0, len(palette)-1]`.
   - Rune is inactive (base style) when `localPhase >= density`.
   - Rune is inactive when `palette` is empty, regardless of `localPhase`.
   - Line breaks in text produce correct `row` / `col` advancement.
   - `Hint` returns dimensions based on the full text, not the active subset.

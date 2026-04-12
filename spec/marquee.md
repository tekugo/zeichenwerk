# Marquee

An `Animation`-driven single-row scrolling text ticker. Text wider than the
widget scrolls continuously to the left at a configurable speed. A configurable
gap of spaces separates the end of the text from the looping start, making the
repeat boundary visually clear. Scrolling pauses automatically while the mouse
cursor is over the widget and resumes when it leaves. Useful for status feeds,
event logs, and live-data dashboards where a line of information must always
be visible without occupying more than one row.

---

## Visual layout

**text = `"Status: System running normally."` (32 cols), gap = 4, widget width = 48**

At offset 0 — the text starts at the left edge:

```
Status: System running normally.    Status: Syst
```

At offset 12 — twelve columns have scrolled off the left:

```
m running normally.    Status: System running nor
```

At offset 32 — inside the gap region; spaces visible at the left:

```
    Status: System running normally.    Status: S
```

The widget loops seamlessly: the cycle length is `textWidth + gap` (36 here).

**Text shorter than widget width — no scrolling:**

```
Hello!
```

When `textWidth <= availableWidth` the text is pinned at the left edge and the
animation does not advance `offset`.

---

## Structure

```go
type Marquee struct {
    Animation
    text      string
    textWidth int    // cached display-column width of text; updated in SetText
    offset    int    // current scroll position in display columns, in [0, cycle)
    speed     int    // display columns advanced per tick (minimum 1)
    gap       int    // spaces appended after text before the loop restarts
}
```

`Animation` is embedded (not a pointer field), exactly as in `Scanner`. This
gives `Marquee` the `Start(interval)`, `Stop()`, `Running()`, and `Tick()`
machinery.

---

## Constructor

```go
func NewMarquee(id, class string) *Marquee
```

Defaults:

- `speed = 1`, `gap = 4`, `offset = 0`.
- `FlagFocusable` is **not** set — the marquee is a display-only widget.
- Sets `m.fn = m.Tick` so the embedded `Animation` calls the right method.
- Does **not** start the animation; the caller calls `m.Start(interval)`.

---

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetText(s string)` | Replaces the scrolling text; recomputes `textWidth`; resets `offset = 0`; calls `Redraw(m)` |
| `Text() string` | Returns the current text |

### Animation control

| Method | Description |
|--------|-------------|
| `Start(interval time.Duration)` | Begins scrolling at the given tick interval (inherited from `Animation`) |
| `Stop()` | Stops the ticker (inherited); `offset` is preserved |
| `Running() bool` | Returns whether the ticker is active (inherited) |
| `SetSpeed(n int)` | Columns advanced per tick; clamped to minimum 1 |
| `SetGap(n int)` | Spaces between end of text and loop start; clamped to minimum 0; calls `Redraw(m)` |

---

## Tick

```go
func (m *Marquee) Tick()
```

Called by `Animation` on each ticker fire:

1. If `m.Flag(FlagHovered)` is set, return immediately — hover pause, no
   state change, no redraw.
2. If `textWidth <= 0` or `textWidth <= availableWidth` (cached at last
   render), return immediately — text fits; nothing to scroll.
3. Advance: `m.offset = (m.offset + m.speed) % m.cycle()`.
4. Call `Redraw(m)`.

`cycle()` is a private helper: `m.textWidth + m.gap`.

`availableWidth` is stored as a field `renderWidth int` and updated at the
start of each `Render` call so that `Tick` has a stable snapshot without
needing to call `Content()` from the animation goroutine.

---

## Hint

```go
func (m *Marquee) Hint() (int, int)
```

- **Width**: 0 — fill the parent. The marquee is meaningless at a fixed width
  smaller than the text, and it adapts to whatever space is given.
- **Height**: 1 (plus style vertical overhead).

Manual overrides via `SetHint` are honoured in the usual way.

---

## Render

```go
func (m *Marquee) Render(r *Renderer)
```

1. `m.Component.Render(r)` — draws background and border.
2. Obtain `(cx, cy, cw, _)` from `m.Content()`. Store `cw` in `m.renderWidth`
   for use by `Tick`.
3. If `m.textWidth == 0` or `cw == 0`, return.
4. Resolve style from `m.Style()`.
5. If `m.textWidth <= cw` — text fits; render it left-aligned at `(cx, cy)`,
   padded with spaces to fill the row. Return.
6. Otherwise, render `cw` display columns starting at virtual position
   `m.offset`:

```
col = 0
for col < cw:
    vpos = (m.offset + col) % m.cycle()
    if vpos < m.textWidth:
        ch = rune at display column vpos in m.text
        r.Put(cx+col, cy, string(ch))
        col += displayWidth(ch)   // 1 for ASCII, 2 for wide CJK
    else:
        r.Put(cx+col, cy, " ")
        col++
```

The rune-at-column lookup requires walking `m.text` from the start on each
render. For performance with very long strings, a pre-built `[]rune` column
index may be maintained alongside `textWidth` in `SetText`.

All cells are drawn with the single `"marquee"` style; the marquee has no
sub-part selectors.

---

## Hover pause

The UI automatically sets and clears `FlagHovered` on the widget as the mouse
enters and leaves it (see `ui.go` mouse dispatch). `Tick` reads this flag
without acquiring any lock — the flag mutation and the Tick goroutine are both
on the same logical "UI state" path and the cost of a torn read is at most one
extra tick of advancement.

No explicit `EvtHover` handler is registered by the marquee itself.

---

## Events

The Marquee dispatches no events of its own. It is a pure display widget.

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"marquee"` | The entire widget: background, foreground, and font |

A single selector is sufficient because all scrolling text is rendered
uniformly. Callers who need highlighted segments (e.g. coloured status words)
should pre-format the text using ANSI escape sequences and rely on the existing
ANSI parser, or use a `Styled` widget instead.

Example theme entries (Tokyo Night):

```go
NewStyle("marquee").WithColors("$cyan", "$bg0"),
```

---

## Builder usage

```go
builder.Marquee("ticker").Hint(-1, 1)

m := builder.Find("ticker").(*Marquee)
m.SetText("Status: All systems operational.  CPU 12%  MEM 4.1 GB  NET ↑ 1.2 MB/s")
m.Start(80 * time.Millisecond)
```

`Hint(-1, 1)` requests full parent width and exactly one row height. The
`80 ms` interval at `speed = 1` gives a smooth 12.5 fps scroll advancing one
display column per tick.

---

## Implementation plan

1. **`marquee.go`** — new file
   - Define `Marquee` struct embedding `Animation`.
   - Implement `NewMarquee(id, class string) *Marquee` with defaults and
     `m.fn = m.Tick`.
   - Implement `SetText`, `Text`, `SetSpeed`, `SetGap`.
   - Implement `cycle() int` private helper.
   - Implement `Apply(t *Theme)`: register `"marquee"` selector.
   - Implement `Hint() (int, int)`.
   - Implement `Tick()`: hover check → fit check → advance offset → `Redraw(m)`.
   - Implement `Render(r *Renderer)`: update `renderWidth`, static path,
     scrolling path with rune-column loop.

2. **`builder.go`** — add `Marquee` method
   ```go
   func (b *Builder) Marquee(id string) *Builder
   ```

3. **Theme** — add `"marquee"` style entry to all built-in themes.

4. **`cmd/demo/main.go`** — add a `"Marquee"` entry with `marqueeDemo`,
   showing a live-updating scrolling status line and a checkbox to toggle
   the animation on/off.

5. **Tests** — `marquee_test.go`
   - `cycle()` returns `textWidth + gap` for ASCII and multi-byte text.
   - `Tick` does not advance `offset` when `FlagHovered` is set.
   - `Tick` does not advance `offset` when `textWidth <= renderWidth`.
   - `Tick` wraps `offset` correctly at the cycle boundary.
   - `SetText` resets `offset` to 0 and updates `textWidth`.
   - `Render` produces the correct column sequence at a given offset,
     including wrap-around across the gap region.
   - `Render` falls back to static left-aligned rendering when text fits.
   - Wide (2-column) runes advance `col` by 2 in the render loop.

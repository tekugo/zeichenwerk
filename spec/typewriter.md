# Typewriter

An `Animation`-driven widget that reveals its text content character by
character, optionally followed by a blinking cursor. When the full text has
been shown the cursor blinks for a configurable dwell period, then the
animation stops — or restarts from the beginning if `repeat` is set. Suitable
for onboarding flows, help text, and terminal-style prompts.

---

## Visual layout

**Revealing — 18 of 26 characters shown, cursor visible:**

```
Initialising subsys▌
```

**Complete — cursor blinking during dwell:**

```
Initialising subsystems…▌
```

**Complete — cursor off (blink phase):**

```
Initialising subsystems…
```

**Multi-line text (two lines revealed, third in progress):**

```
Step 1: Load configuration.
Step 2: Connect to database.
Step 3: Apply migra▌
```

The cursor character `▌` is drawn immediately after the last revealed
character on the current line in the `"typewriter/cursor"` style.

---

## Structure

```go
type Typewriter struct {
    Animation
    text      string   // full text; may contain \n for multiple lines
    runes     []rune   // pre-split runes of text (set in SetText)
    shown     int      // number of runes currently visible [0, len(runes)]
    rate      int      // runes revealed per tick (minimum 1)
    showCursor bool    // whether a cursor is rendered at all
    cursorOn   bool    // current blink state; toggled during dwell phase
    dwell     time.Duration // how long the cursor blinks after completion
    dwellTicks int    // dwell duration converted to tick count in Start()
    dwellTick  int    // ticks elapsed in dwell phase
    repeat    bool    // restart from shown=0 after dwell expires
    interval  time.Duration // stored from Start() for dwellTicks calculation
    // Character read from theme strings in Apply.
    chCursor  string  // default "▌"
}
```

---

## Constructor

```go
func NewTypewriter(id, class string) *Typewriter
```

Defaults:

- `rate = 1`, `showCursor = true`, `dwell = 500ms`, `repeat = false`.
- `chCursor = "▌"`.
- `FlagFocusable` is **not** set — purely a display widget.
- Sets `tw.fn = tw.Tick`.
- Does not start the animation; the caller calls `tw.Start(interval)`.

---

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetText(s string)` | Replaces text; resets `shown = 0`, `dwellTick = 0`, `cursorOn = showCursor`; pre-splits into `runes`; calls `Redraw(tw)` |
| `Text() string` | Returns the full text |

### Display

| Method | Description |
|--------|-------------|
| `SetRate(n int)` | Runes revealed per tick; clamped to minimum 1 |
| `SetCursor(v bool)` | Shows or hides the cursor; calls `Redraw(tw)` |
| `SetDwell(d time.Duration)` | Blink duration after reveal completes; recomputes `dwellTicks` if running |
| `SetRepeat(v bool)` | Enables continuous looping |

### Animation control

| Method | Description |
|--------|-------------|
| `Start(interval time.Duration)` | Stores `interval`; computes `dwellTicks = int(dwell/interval)`; calls `Animation.Start(interval)` |
| `Stop()` | Inherited from `Animation` |
| `Running() bool` | Inherited from `Animation` |
| `Reset()` | Sets `shown = 0`, `dwellTick = 0`, `cursorOn = showCursor`; calls `Redraw(tw)` |

---

## Animation states

The reveal cycle has three phases determined by `shown` and `dwellTick`:

| Phase | Condition | Description |
|-------|-----------|-------------|
| **Revealing** | `shown < len(runes)` | Advance `shown` by `rate` each tick |
| **Dwell** | `shown == len(runes)` and `dwellTick < dwellTicks` | Count `dwellTick` up; blink cursor |
| **Done** | `dwellTick >= dwellTicks` | Stop or loop |

---

## Tick

```go
func (tw *Typewriter) Tick()
```

```
switch:
case shown < len(runes):                   // Revealing
    shown = min(shown + rate, len(runes))
    cursorOn = showCursor                   // cursor always on during reveal
    Redraw(tw)

case dwellTick < dwellTicks:               // Dwell — blink cursor
    dwellTick++
    cursorOn = showCursor && (dwellTick % 2 == 0)
    Redraw(tw)

default:                                   // Done
    if repeat:
        shown = 0
        dwellTick = 0
        cursorOn = showCursor
        Redraw(tw)
    else:
        cursorOn = false
        Redraw(tw)
        tw.Stop()
```

`tw.Stop()` is safe to call from within `Tick` — it sends to the stop channel
which the animation goroutine drains on its next select iteration.

When `dwell == 0`, `dwellTicks == 0` and the dwell phase is skipped
entirely: the animation stops (or loops) immediately after the last character
is revealed.

---

## Hint

```go
func (tw *Typewriter) Hint() (int, int)
```

- **Width**: manual override if set; otherwise the display-column width of
  the longest line in `text`, plus 1 for the cursor if `showCursor`, plus
  style horizontal overhead.
- **Height**: manual override if set; otherwise the line count of `text`
  (number of `\n` + 1), plus style vertical overhead.

Lines are counted from the **full** text (not the revealed portion) so that
the surrounding layout does not shift as text is revealed.

---

## Render

```go
func (tw *Typewriter) Render(r *Renderer)
```

1. `tw.Component.Render(r)` — background and border.
2. Obtain `(cx, cy, cw, ch)` from `tw.Content()`.
3. Resolve `baseStyle = tw.Style()` and `cursorStyle = tw.Style("cursor")`.
4. Split `tw.runes[:tw.shown]` into lines at `\n`.
5. For each line `i` (up to `ch` rows):
   - `r.Set(baseStyle…)`
   - `r.Text(cx, cy+i, string(line), cw)`
6. Draw cursor if `tw.cursorOn` and `tw.showCursor`:
   - Position: `(cx + displayWidth(lastLine), cy + lastLineIndex)`.
   - `r.Set(cursorStyle…)`, `r.Put(cursorX, cursorY, tw.chCursor)`.

---

## Events

| Event | Data | Description |
|-------|------|-------------|
| `EvtChange` | `bool` (= `true`) | Dispatched when the reveal completes (all runes shown), before dwell starts |
| `EvtActivate` | `bool` (= `true`) | Dispatched when dwell expires and `repeat = false` (animation done) |

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"typewriter"` | Text and widget background |
| `"typewriter/cursor"` | The cursor character `▌` |

Example theme entries (Tokyo Night):

```go
NewStyle("typewriter").WithColors("$fg0", "$bg0"),
NewStyle("typewriter/cursor").WithColors("$blue", "$bg0"),
```

---

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `"typewriter.cursor"` | `"▌"` | Character drawn after the last revealed rune |

---

## Builder usage

```go
builder.Typewriter("intro")

tw := builder.Find("intro").(*Typewriter)
tw.SetText("Initialising subsystems…").
    SetRate(2).
    SetCursor(true).
    SetDwell(1500 * time.Millisecond)
tw.Start(20 * time.Millisecond)
tw.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
    tw.SetText("Ready.")
    tw.Start(30 * time.Millisecond)
    return true
})
```

---

## Implementation plan

1. **`typewriter.go`** — new file
   - Define `Typewriter` struct embedding `Animation`.
   - Implement `NewTypewriter(id, class string) *Typewriter`.
   - Implement `SetText`, `Text`, `SetRate`, `SetCursor`, `SetDwell`,
     `SetRepeat`, `Reset`.
   - Override `Start(interval time.Duration)` to compute `dwellTicks`.
   - Implement `Tick()` with the three-phase state machine.
   - Implement `Apply(t *Theme)`: register `"typewriter"` and
     `"typewriter/cursor"` selectors; read `"typewriter.cursor"` string.
   - Implement `Hint() (int, int)`.
   - Implement `Render(r *Renderer)`.
   - Dispatch `EvtChange` when `shown` first reaches `len(runes)`.
   - Dispatch `EvtActivate` when done and `repeat = false`.

2. **`builder.go`** — add `Typewriter` method
   ```go
   func (b *Builder) Typewriter(id string) *Builder
   ```

3. **Theme** — add `"typewriter"` and `"typewriter/cursor"` style entries and
   the `"typewriter.cursor"` string key to all built-in themes.

4. **`cmd/demo/main.go`** — add `"Typewriter"` entry with `typewriterDemo`,
   showing a multi-phrase sequence (each phrase starts on `EvtActivate`),
   and controls for rate, dwell, and repeat via checkboxes.

5. **Tests** — `typewriter_test.go`
   - `Tick` advances `shown` by `rate` each call during revealing.
   - `shown` does not exceed `len(runes)`.
   - `dwellTick` increments during dwell; `cursorOn` toggles every tick.
   - `Stop()` is called after dwell expires when `repeat = false`.
   - `repeat = true` resets `shown` and `dwellTick` after dwell.
   - `EvtChange` fires exactly once per reveal cycle.
   - `EvtActivate` fires on completion when `repeat = false`.
   - `Hint` returns full-text dimensions (not revealed portion).
   - Multi-line text splits correctly at `\n`.

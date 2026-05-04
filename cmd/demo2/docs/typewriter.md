# Typewriter

Animated widget that reveals text character by character. Supports a blinking cursor during and after the reveal, an optional dwell phase, and unlimited repeats.

**Constructor:** `NewTypewriter(id, class string) *Typewriter`

Defaults: rate = 1 char/tick, cursor visible (`▌`), 500 ms dwell after reveal, no repeat.

## Methods

- `SetText(s string) *Typewriter` — replace text, reset reveal state
- `Text() string` — full text (not just revealed portion)
- `SetRate(n int) *Typewriter` — characters revealed per tick (clamped to ≥ 1)
- `SetCursor(v bool) *Typewriter` — toggle the trailing cursor
- `SetDwell(d time.Duration) *Typewriter` — how long the cursor blinks after reveal completes
- `SetRepeat(v bool) *Typewriter` — restart automatically after completion
- `Reset()` — reset reveal state without changing the text
- `Start(interval time.Duration)` — start animation, stores interval to compute dwell ticks
- `Stop()` — halt (inherited from `Animation`)

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `bool` (always `true`) | Reveal phase reached the end |
| `"activate"` | `bool` (always `true`) | Animation completed (only when `repeat = false`) |

## Notes

Embeds `Animation`. Three-phase state machine: revealing → dwell → done. The cursor blinks during dwell (every other tick) and disappears at done; if `repeat` is true the cycle restarts instead.

Style selectors: `typewriter` for the text, `typewriter/cursor` for the cursor glyph. Override the cursor character via theme string `typewriter.cursor` (default `▌`).

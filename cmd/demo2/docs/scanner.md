# Scanner

Back-and-forth scanning animation with fading trail.

**Constructor:** `NewScanner(id, class string, width int, style string) *Scanner`

## Methods

- `Running() bool` — true if animation is active
- `Start(interval time.Duration)` — starts animation
- `Stop()` — stops animation

## Notes

**Style values:** `"blocks"`, `"diamonds"`, `"circles"`

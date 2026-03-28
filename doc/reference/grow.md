# Grow

Animation wrapper that reveals content by growing from 0 to full size.

**Constructor:** `NewGrow(id, class string, horizontal bool) *Grow`

## Methods

- `Add(widget Widget)` — sets child widget
- `Children() []Widget` — returns child widget
- `Layout()` — calculates end size and child bounds
- `Running() bool` — true if animation is active
- `Start(interval time.Duration)` — begins grow animation
- `Stop()` — stops animation

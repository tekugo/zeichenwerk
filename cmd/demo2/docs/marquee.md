# Marquee

Single-row scrolling text ticker. Text wider than the widget scrolls continuously to the left; scrolling pauses while the mouse cursor is over the widget (`FlagHovered`).

**Constructor:** `NewMarquee(id, class string) *Marquee`

Defaults: speed = 1 column / tick, gap = 4 spaces between cycles. The animation is **not** started automatically — call `Start(interval)` to begin scrolling.

## Methods

- `SetText(s string) *Marquee` — replace scrolling text, reset offset, redraw
- `Text() string` — current scrolling text
- `SetSpeed(n int) *Marquee` — display columns advanced per tick (clamped to ≥ 1)
- `SetGap(n int) *Marquee` — blank columns inserted between end of text and looping start (clamped to ≥ 0)
- `Tick()` — animation frame; advances the scroll offset
- `Start(interval time.Duration)` — begin scrolling (inherited from `Animation`)
- `Stop()` — halt scrolling

## Notes

Embeds `Animation`. Dispatches no events — pure display widget.

Wide characters (CJK, emoji) are handled correctly: each rune's display-column width is tracked so half-wide-character clipping at the edges renders properly.

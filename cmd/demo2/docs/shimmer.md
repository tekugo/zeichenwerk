# Shimmer

Skeleton-style loading shimmer: text with a highlight band sweeping continuously left-to-right. Characters inside the band blend toward an accent colour; characters outside use the base style. Multi-line text is supported.

**Constructor:** `NewShimmer(id, class string) *Shimmer`

Defaults: bandWidth = 6, edgeWidth = 3, gradient = off (linear falloff). The animation is **not** started automatically — call `Start(interval)` to begin.

## Methods

- `SetText(s string) *Shimmer` — replace text (split at `\n`); resets band position
- `Text() string` — current text
- `SetBandWidth(n int) *Shimmer` — core highlight width in columns (clamped to ≥ 1)
- `SetEdgeWidth(n int) *Shimmer` — gradient columns on each side of the core (0 = hard edge)
- `SetGradient(on bool) *Shimmer` — smooth cosine blending vs. stepped linear falloff
- `Tick()` — animation frame; advances the band by one column
- `Start(interval time.Duration)` — start animation (inherited from `Animation`)
- `Stop()` — stop and freeze in base style

## Notes

Embeds `Animation`. Dispatches no events — pure display widget.

Style selectors: `shimmer` for the base text, `shimmer/band` for the highlight colour. The renderer interpolates between the two as the band passes.

Calling `Stop()` freezes the text mid-sweep. To restart, call `Start(interval)` again.

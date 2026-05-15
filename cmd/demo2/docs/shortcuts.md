# Shortcuts

Single-line row of keyboard hint pairs — a highlighted key followed by a dimmed label, repeated and separated by theme-defined glyphs. Useful for footer help bars.

**Constructor:** `NewShortcuts(id, class string, pairs ...string) *Shortcuts`

`pairs` is alternating key/label strings (same convention as `NewSelect`):

```go
zw.NewShortcuts("help", "", "r", "run", "w", "watch", "q", "quit")
//  → renders:    r run   w watch   q quit
```

## Methods

- `SetPairs(pairs ...string)` — replace all pairs (same alternating-string convention) and redraw

## Notes

Style selectors:

- `shortcuts` — base style (background, prefix/suffix colour, optional border)
- `shortcuts/key` — colour/font for each key portion
- `shortcuts/label` — colour/font for each label portion

Theme string tokens (override decoration without touching code):

| Token | Default | Effect |
|-------|---------|--------|
| `shortcuts.prefix` | `""` | Rendered once before the first pair |
| `shortcuts.separator` | `"   "` | Rendered between consecutive pairs |
| `shortcuts.suffix` | `""` | Rendered after the last pair |

Width hint is the natural sum of all rendered runes; height is always 1. Dispatches no events.

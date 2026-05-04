# Indicator

Status glyph followed by a label. The glyph and its colour follow the indicator level (Debug, Info, Success, Warning, Error, Fatal).

**Constructor:** `NewIndicator(id, class string, level Level, label string) *Indicator`

## Methods

- `Level() Level` — current level
- `SetLevel(l Level)` — change level (updates glyph & colour via the `indicator:<level>` selector)
- `Label() string` / `SetLabel(s string)` — change the label text

## Notes

Levels: `Debug`, `Info`, `Success`, `Warning`, `Error`, `Fatal`. Each level has its own selector — e.g. `indicator:warning` — so themes can pick distinct glyphs and colours per level.

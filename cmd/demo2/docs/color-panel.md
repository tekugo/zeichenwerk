# ColorPanel

Theme-colour preview palette. Renders every named color from the active theme as a coloured swatch with its name and hex value.

**Constructor:** `NewColorPanel(id, class, title string) *ColorPanel`

## Methods

- `Refresh()` — re-read the theme palette and redraw

## Notes

Useful for theme designers and as a debugging aid. The panel re-reads the theme on `Apply`, so swapping themes updates it automatically.

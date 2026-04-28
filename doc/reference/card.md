# Card

Bordered container with a title in the top border line, a main content area, and an optional fixed-height footer pinned to the bottom.

**Constructor:** `NewCard(id, class, title string) *Card`

The first widget added via `Add` becomes the content; the second becomes the footer. Further calls replace the footer.

## Methods

- `Add(widget Widget) error` — assigns content (first call) or footer (second call); subsequent calls replace the footer
- `Children() []Widget` — returns content and footer (in that order, omitting nil)
- `Set(value string)` — updates the title and refreshes
- `Layout() error` — positions content and footer
- `Hint() (w, h int)` — combined size of content + footer

## Notes

Style selectors: `card`, `card/title` (theme can target the title strip independently from the body).

Layout details:
- With a footer present, content fills available height minus the footer's hint height; footer is pinned to the bottom.
- Without a footer, content fills the entire content area.

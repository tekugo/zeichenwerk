# Text

Multi-line scrollable text display.

**Constructor:** `NewText(id, class string, content []string, follow bool, max int) *Text`

## Methods

- `Add(lines ...string)` — appends lines; rotates oldest when max is set
- `Clear()` — removes all content
- `Set(content []string)` — replaces all content

## Notes

- `follow=true` — auto-scrolls to newest content
- `max=N` — limits total lines to N (oldest are discarded)

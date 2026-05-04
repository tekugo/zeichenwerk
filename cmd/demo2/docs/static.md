# Static

Plain non-interactive text label.

**Constructor:** `NewStatic(id, class, text string) *Static`

## Methods

- `Set(value any)` — replace text (non-string values are formatted with `fmt.Sprintf("%v", v)`); queues a refresh
- `SetAlignment(align string)` — set text alignment: `"left"`, `"center"`, `"right"`

The text and alignment are also exposed as public fields (`Text`, `Alignment`) for direct access if needed.

## Notes

Hint width is `utf8.RuneCountInString(Text)`; height is 1. Override with `.Hint(w, h)` for explicit sizing.

Set `FlagRight` to right-align the text within the content area; otherwise the `Alignment` field controls placement (default `"left"`).

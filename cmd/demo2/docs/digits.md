# Digits

Large-format display using ASCII-art characters — typically used for clocks and counters.

**Constructor:** `NewDigits(id, class, text string) *Digits`

## Methods

- `Get() string` — current displayed text
- `Set(value string)` — replace displayed text (queues a redraw)

## Notes

Supported characters: `0–9`, `A–F`, `:`, `.`, `-`. Other characters render as blanks.

The `FlagRight` flag right-aligns the digits within the content area instead of left-aligning.

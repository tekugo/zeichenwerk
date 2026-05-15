# Spinner

Animated loading indicator cycling through a character sequence.

**Constructor:** `NewSpinner(id, class, sequence string) *Spinner`

## Methods

- `Current() string` — returns currently displayed character
- `Running() bool` — true if animation is active
- `SetSequence(sequence string)` — updates animation sequence (space-separated characters)
- `Start(interval time.Duration)` — starts animation
- `Stop()` — stops animation
- `Tick()` — advances to next frame

## Notes

**Predefined sequences** (use `Spinners["key"]`):

| Key | Characters |
|-----|-----------|
| `"bar"` | `\| / - \` |
| `"bounce"` | `⠁ ⠂ ⠄ ⠂` |
| `"braille"` | `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏` |
| `"circle"` | `◐ ◓ ◑ ◒` |
| `"dot"` | `⠁ ⠂ ⠄ ⡀ ⢀ ⠠ ⠐ ⠈` |
| `"dots"` | `. o O o` |
| `"arrow"` | `← ↖ ↑ ↗ → ↘ ↓ ↙` |

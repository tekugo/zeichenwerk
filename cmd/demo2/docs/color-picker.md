# ColorPicker

Interactive RGB / HSL / Hex colour selector with optional foreground & background pickers and a contrast-ratio readout.

**Constructor:** `NewColorPicker(id, class string, mode ColorPickerMode) *ColorPicker`

`mode` is one of:

| Mode | Behaviour |
|------|-----------|
| `ColorSingle` | One color (`Foreground()`) |
| `ColorFgBg`   | Foreground + background; shows WCAG contrast ratio |

## Methods

- `Foreground() string` / `SetForeground(hex string)`
- `Background() string` / `SetBackground(hex string)` (FgBg mode only)
- `Contrast() float64` — current foreground/background contrast ratio
- `SetMode(mode ColorPickerMode)` — switch between modes

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `*ColorPicker` | Any colour component changed |

## Notes

Editing any of R/G/B, H/S/L, or the hex field updates all other representations. Tab cycles between fields; arrow keys nudge the selected component.

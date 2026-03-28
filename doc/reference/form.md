# Form

Data-bound form container connected to a Go struct via reflection.

**Constructor:** `NewForm(id, class, title string, data any) *Form`

## Methods

- `Add(widget Widget)` — sets single child container
- `Children() []Widget` — returns child (usually FormGroup)
- `Data() any` — returns data struct pointer
- `Title() string` — returns form title

## Notes

Struct field tags: `group`, `label`, `control`, `options`, `width`, `line`, `readonly`

# FormGroup

Arranges labeled form controls within a Form.

**Constructor:** `NewFormGroup(id, class, title string, horizontal bool, spacing int) *FormGroup`

## Methods

- `Add(line int, label string, widget Widget)` — adds field at grid position
- `Children() []Widget` — returns all form control widgets
- `Layout()` — arranges fields

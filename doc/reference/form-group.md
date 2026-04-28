# FormGroup

Container for labeled form controls inside a `Form`. Widgets are organised into "lines" (rows), each line carrying one or more labelled controls. Layout can be horizontal (label and control side-by-side) or vertical (label above control).

**Constructor:** `NewFormGroup(id, class, title string, horizontal bool, spacing int) *FormGroup`

## Methods

- `Add(widget Widget, params ...any) error` — append a control. `params[0]` (int) is the line index; `params[1]` (string) is the label. Multiple controls on the same line render in a row. Returns `ErrChildIsNil` if `widget` is nil.
- `Children() []Widget` — all controls across all lines, flattened
- `Layout() error` — position labels and controls
- `Hint() (w, h int)` — preferred size needed to display every line at its natural size

## Notes

When used inside a Form, the Builder's `Group(id, title, name, horizontal, spacing)` method auto-generates one labelled control per matching struct field — you usually don't call `Add` yourself.

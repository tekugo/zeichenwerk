# Dialog

Container with optional title and border. Used as popup layer via `UI.Popup`.

**Constructor:** `NewDialog(id, class, title string) *Dialog`

## Methods

- `Add(widget Widget)` — sets content widget
- `Children() []Widget` — returns content widget
- `Layout()` — positions content within dialog

# Select

Dropdown selection widget.

**Constructor:** `NewSelect(id, class string, args ...string) *Select`

`args` are alternating value/label pairs: `value1, label1, value2, label2, ...`

## Methods

- `Select(value string)` — selects option by value
- `Text() string` — returns display label of selected option
- `Value() string` — returns value of selected option

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `string` | Selected value changed |

## Notes

Flags: `"focusable"`

Opens a List popup on Enter; Escape closes it.

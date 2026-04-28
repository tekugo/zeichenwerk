# Combo

Single-line input that opens a popup with a Typeahead and a List of suggestions on focus or Enter. The user can type freely or pick from the list — the value is whatever was confirmed.

**Constructor:** `NewCombo(id, class string, items []string) *Combo`

## Methods

- `Get() string` — current confirmed value
- `Set(value string)` — set the value (does not fire `EvtChange`)
- `SetItems(items []string)` — replace the suggestion list

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `string` | Current input text while the popup is open |
| `"activate"` | `string` | Confirmed value when Enter is pressed |

## Notes

Flags: `"focusable"`.

The popup is opened automatically on `EvtFocus` and on Enter. Inside the popup, the embedded Typeahead handles search and arrow-key navigation through the List.

Useful for search fields with a small set of common candidates or a history.

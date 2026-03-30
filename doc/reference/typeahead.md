# Typeahead

Text input with inline ghost-text suggestion. Extends `Input` — all Input methods and events apply.

**Constructor:** `NewTypeahead(id, class string, params ...string) *Typeahead`

`params` follow the same convention as `NewInput`: optional initial value and placeholder.

## Methods

- `SetSuggest(fn func(string) []string)` — sets the suggestion provider; called on every keystroke with the current text; the first returned string becomes the ghost-text suggestion

All other methods are inherited from `Input`.

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"accept"` | `string` | Suggestion accepted; data is the completed text |
| `"change"` | `string` | Text changed (inherited from Input) |

## Notes

Flags: `"focusable"`, `"masked"`, `"readonly"`

`Tab` or `→` (when cursor is at end) accepts the active suggestion and dispatches `"accept"`. `Esc` clears the suggestion without accepting.

Style selectors: `"typeahead"`, `"typeahead/suggestion"` — with `:focused` state.

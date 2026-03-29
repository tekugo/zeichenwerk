# Typeahead

A single-line text input that shows an inline ghost-text suggestion after the
cursor. The suggestion is provided by a caller-supplied callback and accepted
with Tab. All standard `Input` editing behaviour is preserved unchanged.

## Structure

```go
type Typeahead struct {
    Input                          // All input editing functionality
    suggest func(string) []string  // Suggestion provider; nil disables typeahead
    hint    string                 // Current suggestion (full string, not just suffix)
}
```

`hint` holds the full suggested string. Only the suffix — `hint[len(text):]` — is
rendered as ghost text. When the input text no longer matches `hint` as a prefix,
`hint` is cleared.

## Constructor

```go
func NewTypeahead(id, class string, params ...string) *Typeahead
```

- Identical params to `NewInput` (`[0]` initial text, `[1]` placeholder,
  `[2]` mask character).
- Initialises the embedded `Input` with the same field setup as `NewInput`.
- Registers `Typeahead`'s own key handler via `OnKey`. Because `On` prepends
  handlers, this handler runs **before** `Input`'s handler and can intercept
  Tab and `→` at end-of-text.
- `suggest` is `nil` until set via `SetSuggest`.

## Setting the suggestion provider

```go
func (t *Typeahead) SetSuggest(fn func(string) []string)
```

Called by the application to supply completions. The function receives the
current input text and returns candidate strings. Returning `nil` or an empty
slice clears the ghost text. The function must be cheap to call synchronously;
asynchronous updating is out of scope.

## Updating the hint

`updateHint(text string)` is called internally after every text change:

1. If `suggest` is nil, set `hint = ""` and return.
2. Call `suggest(text)` and take the first result that has `text` as a
   case-sensitive prefix. Store it as `hint`.
3. If no match, set `hint = ""`.
4. Call `Refresh()`.

`updateHint` is wired to `EvtChange` in the constructor so it runs after every
edit, including programmatic `SetText` calls.

## Interaction

Typeahead's key handler runs first, then `Input`'s handler processes everything
it was not intercepted.

| Key | Condition | Behaviour |
|-----|-----------|-----------|
| `Tab` | `hint != ""` | Accept: call `Input.SetText(hint)`, move cursor to end, clear `hint`; return `true` |
| `Tab` | `hint == ""` | Return `false` (propagate) |
| `→` | cursor at end of text and `hint != ""` | Accept same as Tab; return `true` |
| `→` | cursor not at end | Return `false` — `Input` moves cursor normally |
| `Esc` | any | Clear `hint`, `Refresh()`, return `false` (propagate) |
| All others | — | Return `false` — `Input` handles them; `EvtChange` triggers `updateHint` |

Accepting dispatches `EvtChange` (via `Input.SetText`) and `EvtAccept` (see
Events). No additional `EvtChange` is fired by Typeahead itself.

## Rendering

```go
func (t *Typeahead) Render(r *Renderer)
```

1. Call `t.Input.Render(r)` — draws the typed text exactly as a plain Input.
2. If `hint == ""` or input is masked, return.
3. Compute ghost suffix: `suffix = hint[len(text):]`. If empty, return.
4. Compute render position: `ghostX = cursorX` (from `t.Cursor()`), `ghostY = 0`.
5. Compute available width: `availW = contentW - ghostX`.
6. If `availW <= 0`, return.
7. Render `suffix` at `(cx+ghostX, cy+ghostY)` with width `availW` using the
   `"typeahead/hint"` style (dimmed by default).

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `string` | Text changed (inherited from `Input`) |
| `"enter"` | `string` | Enter key pressed (inherited from `Input`) |
| `"accept"` | `string` | Suggestion accepted; data is the accepted value |

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"typeahead"` | Widget-level style (inherits from `"input"`) |
| `"typeahead/hint"` | Ghost text — typically dimmed foreground, same background |

`Apply` calls `theme.Apply(t, t.Selector("typeahead"), "disabled", "focused", "hovered")`
in addition to the embedded `Input.Apply` call, so both selectors are populated.

## Notes

- `hint` is always cleared when `suggest` returns no match, when the text is
  cleared, and on Esc.
- Masking (`FlagMasked`) suppresses ghost text rendering entirely — guessing a
  masked value is not useful.
- `Typeahead` does not open a popup. For a dropdown of suggestions, pair a
  `Typeahead` with a `Combo` widget, or wire `EvtChange` to update a separate
  `List`.
- The `suggest` callback is called on every keystroke. Callers are responsible
  for caching or debouncing if the candidate set is expensive to compute.

## Implementation Plan

1. **`typeahead.go`** — new file
   - Define `Typeahead` struct embedding `Input`.
   - Implement `NewTypeahead`: initialise embedded `Input` fields, set
     `FlagFocusable`, register `Input`'s key handler, then register
     `Typeahead`'s own key handler (prepended, runs first).
   - Implement `SetSuggest`, `updateHint`, `handleKey`, `Apply`, `Render`.
   - Wire `updateHint` to `EvtChange` in the constructor.

2. **`builder.go`** — add `Typeahead` method
   ```go
   func (b *Builder) Typeahead(id string, params ...string) *Builder
   ```

3. **Theme** — add `"typeahead/hint"` style entry to built-in themes with a
   dimmed foreground. No breaking changes; falls back gracefully if absent.

4. **Tests** — `typeahead_test.go`
   - `updateHint` sets ghost text when suggest returns a match.
   - `updateHint` clears ghost text when no match or `suggest` is nil.
   - Tab with active hint dispatches `EvtAccept` and sets input text.
   - Tab without hint propagates (returns `false`).
   - `→` at end with hint accepts; `→` mid-text does not.
   - Esc clears hint without consuming the event.
   - Masking suppresses ghost text rendering.

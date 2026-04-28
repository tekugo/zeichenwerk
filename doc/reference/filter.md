# Filter

Search input that progressively filters another widget (List, Tree, …) as the user types. Embeds `Typeahead`, so it also shows ghost-text prefix completion. Default placeholder is `"Filter…"`.

**Constructor:** `NewFilter(id, class string) *Filter`

## Methods

- `Bind(w Filterable)` — wire to a target widget. If `w` also implements `Suggester`, its `Suggest` method becomes the ghost-text provider.
- `Unbind()` — detach; resets the bound widget's filter to empty
- `Bound() Filterable` — current binding (or nil)
- `Clear()` — clear the input text (bound widget's filter is reset via the EvtChange chain)

Plus everything inherited from `Typeahead` / `Input`: `Get()`, `Set(text string)`, `Text()`, `SetSuggest(fn)`, etc.

## Interfaces

```go
type Filterable interface {
    Filter(filter string)        // empty string clears the filter
}
```

`List` and `Tree` (among others) implement `Filterable`. A widget that *also* implements `Suggester` (`Suggest(query) []string`) gives Filter the data it needs to draw ghost-text completions; without it, ghost text is disabled but filtering still works.

## Notes

Flags: `"focusable"`.

Ghost-text and list filtering intentionally use different matching: ghost text shows the first item whose text *starts with* the typed text; the bound widget shows all items that *contain* it as a substring.

Style selectors fall through `filter` → `typeahead` → `typeahead/suggestion` so themes that don't define `filter` keep working.

# Filter

A standalone input widget that progressively filters a bound `List` or `Tree`
as the user types, and optionally shows the first prefix match as inline ghost
text. `Filter` embeds `Typeahead` — it is a full text input with ghost-text
completion; the only addition is the `Bind` / `Unbind` mechanism and the
filtering side-effect on `EvtChange`.

## Filterable interface

```go
type Filterable interface {
    Filter(filter string)
}
```

`Filter` applies a case-insensitive substring match and updates the widget's
visible content. An empty string resets the filter and restores the full,
unfiltered content — the same behavior triggered when the user deletes the
filter input.

Both `List` and `Tree` implement this interface (see *Changes to existing
widgets* below).

## Suggester interface (optional)

```go
type Suggester interface {
    Suggest(query string) []string
}
```

An optional extension of `Filterable`. When the bound widget also implements
`Suggester`, `Filter` wires it as the typeahead suggestion provider so ghost
text appears as the user types.

`Suggest` should return items (or node labels) that have `query` as a
case-insensitive prefix. Because `Typeahead.updateHint` picks the first result
that has the typed text as a case-sensitive prefix, returning only prefix
matches avoids false negatives from case mismatches.

Both `List` and `Tree` implement this interface (see *Changes to existing
widgets* below).

## Structure

```go
type Filter struct {
    Typeahead                  // Ghost-text input; suggest wired on Bind
    bound    Filterable        // Currently bound widget (nil = no binding)
    placeholder string         // Shown when empty (default "Filter…")
}
```

## Constructor

```go
func NewFilter(id, class string) *Filter
```

- Creates the embedded `Typeahead` with placeholder `"Filter…"` and no initial
  suggest function.
- Wires an `EvtChange` handler (prepended, runs first) that calls
  `applyFilter()` on every text change.
- Sets `FlagFocusable`.

## Methods

| Method | Description |
|--------|-------------|
| `Bind(w Filterable)` | Sets the bound widget; if `w` implements `Suggester` calls `SetSuggest(w.Suggest)`, otherwise calls `SetSuggest(nil)`; immediately calls `applyFilter()` |
| `Unbind()` | Calls `bound.Filter("")`, `SetSuggest(nil)`, then sets `bound = nil` |
| `Bound() Filterable` | Returns the currently bound widget |
| `Clear()` | Clears the input text and calls `bound.Filter("")` if bound (overrides `Typeahead.Clear`) |

`applyFilter()` is unexported:

```go
func (f *Filter) applyFilter() {
    if f.bound == nil {
        return
    }
    f.bound.Filter(f.Text())
}
```

Note: ghost text (prefix completion) and list filtering use different matching
semantics intentionally. Ghost text shows the first item that starts with the
typed text; the bound widget shows all items that contain it as a substring.

## Events

Inherits `Typeahead` events unchanged.

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `string` | Text changed (inherited — fired after `applyFilter`) |
| `"enter"` | `string` | Enter key pressed (inherited) |
| `"accept"` | `string` | Suggestion accepted via Tab or → (inherited from `Typeahead`) |

## Styling

Inherits `"typeahead"` and `"typeahead/hint"` selectors fully. `Apply`
additionally registers `"filter"` for callers that want to style the filter
field separately from plain inputs. Falls back to `"typeahead"` styles if
`"filter"` is not defined.

## Changes to existing widgets

### `List`

New fields:

```go
original []string  // unfiltered items (nil = no active filter)
```

New methods:

```go
func (l *List) Filter(filter string)
func (l *List) Suggest(query string) []string
```

`Filter`:
- If `filter == ""`: if `original != nil`, call `l.SetItems(original)` and set `original = nil`; return.
- Otherwise: if `original == nil`, save `l.items` as `original`. Build a filtered slice: items from `original` where `strings.Contains(strings.ToLower(item), strings.ToLower(filter))`. Call `l.SetItems(filtered)`.

`Suggest`:
1. Use `original` as the source if non-nil (unfiltered), otherwise `l.items`.
2. Return all items where `strings.HasPrefix(strings.ToLower(item), strings.ToLower(query))`.
3. Return nil if no matches.

`SetItems` is unchanged — it always replaces `l.items` and resets the index,
which is correct for both filtered and unfiltered updates.

### `Tree`

New fields:

```go
filterQuery string  // active query ("" = no filter)
```

New methods:

```go
func (t *Tree) Filter(filter string)
func (t *Tree) Suggest(query string) []string
```

`Filter`:
1. Store `filterQuery = filter`.
2. Call `rebuild()` — the flattening pass already walks all nodes.
   When `filterQuery == ""`, rebuild proceeds normally (no filtering).
   Otherwise, during flattening, skip nodes whose text does not match
   `filterQuery` (and have no matching descendants). A node is included if
   it matches OR any of its descendants match (so the path to matching nodes
   stays visible).
3. When non-empty, matching nodes have their parents auto-expanded for this
   render; the underlying `expanded` state is not mutated.

`Suggest`:
1. Walk all nodes (depth-first).
2. Return labels of nodes where `strings.HasPrefix(strings.ToLower(label), strings.ToLower(query))`.
3. Return nil if no matches.

## Builder

```go
func (b *Builder) Filter(id string) *Builder
```

## Implementation plan

1. **`filterable.go`** — new file: define `Filterable` and `Suggester`
   interfaces.

2. **`filter.go`** — new file: define `Filter` struct, `NewFilter`, `Bind`,
   `Unbind`, `Bound`, `Clear`, `applyFilter`, `Apply`.

3. **`list.go`** — add `original []string` field; implement `Filter` and
   `Suggest`.

4. **`tree.go`** — add `filterQuery string` field; extend `rebuild` to skip
   non-matching subtrees when `filterQuery != ""`; implement `Filter` and
   `Suggest`.

5. **`builder.go`** — add `Filter` method.

6. **Tests** — `filter_test.go`
   - Typing in a bound Filter calls `Filter` on the List with the correct query.
   - Clearing the Filter calls `Filter("")` and restores original List items.
   - `Unbind` restores the List and detaches.
   - Tree filter exposes parent nodes of matching descendants.
   - `Filter("")` on List and Tree restores the full unfiltered content.
   - Ghost text appears when bound widget implements `Suggester` and query is a prefix of an item.
   - No ghost text when bound widget does not implement `Suggester`.
   - Accepting ghost text via Tab dispatches `EvtAccept` and updates the filter.

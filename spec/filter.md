# Filter

A standalone input widget that progressively filters a bound `List` or `Tree`
as the user types. The Filter widget owns the filtering logic; bound widgets
implement a small `Filterable` interface that Filter calls on every change.

## Filterable interface

```go
type Filterable interface {
    Filter(query string)
    ResetFilter()
}
```

`Filter` applies a case-insensitive substring match and updates the widget's
visible content. `ResetFilter` restores the full, unfiltered content.

Both `List` and `Tree` implement this interface (see *Changes to existing
widgets* below).

## Structure

```go
type Filter struct {
    Input                   // All single-line input behaviour
    bound    Filterable     // Currently bound widget (nil = no binding)
    placeholder string      // Shown when empty (default "Filter…")
}
```

`Filter` embeds `Input` — it is a full text input with the same editing keys,
cursor, placeholder, and events. The only addition is the `Bind` / `Unbind`
mechanism and the filtering side-effect on `EvtChange`.

## Constructor

```go
func NewFilter(id, class string) *Filter
```

- Creates the embedded `Input` with placeholder `"Filter…"`.
- Wires an `EvtChange` handler (prepended, runs first) that calls
  `applyFilter()` on every text change.
- Sets `FlagFocusable`.

## Methods

| Method | Description |
|--------|-------------|
| `Bind(w Filterable)` | Sets the bound widget; immediately calls `applyFilter()` with the current text |
| `Unbind()` | Calls `bound.ResetFilter()` then sets `bound = nil` |
| `Bound() Filterable` | Returns the currently bound widget |
| `Clear()` | Clears the input text and calls `bound.ResetFilter()` if bound (overrides `Input.Clear`) |

`applyFilter()` is unexported:

```go
func (f *Filter) applyFilter() {
    if f.bound == nil {
        return
    }
    q := f.Text()
    if q == "" {
        f.bound.ResetFilter()
    } else {
        f.bound.Filter(q)
    }
}
```

## Events

Inherits `Input` events unchanged.

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `string` | Text changed (inherited — fired after `applyFilter`) |
| `"enter"` | `string` | Enter key pressed (inherited) |

## Styling

Inherits `"input"` selector fully. `Apply` additionally registers
`"filter"` for callers that want to style the filter field separately from
plain inputs. Falls back to `"input"` styles if `"filter"` is not defined.

## Changes to existing widgets

### `List`

New fields:

```go
original []string  // unfiltered items (nil = no active filter)
```

New methods:

```go
func (l *List) Filter(query string)
func (l *List) ResetFilter()
```

`Filter`:
1. If `original == nil`, save `l.items` as `original`.
2. Build a filtered slice: items from `original` where `strings.Contains(strings.ToLower(item), strings.ToLower(query))`.
3. Call `l.SetItems(filtered)`.

`ResetFilter`:
1. If `original == nil`, return.
2. Call `l.SetItems(original)`.
3. Set `original = nil`.

`SetItems` is unchanged — it always replaces `l.items` and resets the index,
which is correct for both filtered and unfiltered updates.

### `Tree`

New fields:

```go
filterQuery string  // active query ("" = no filter)
```

New methods:

```go
func (t *Tree) Filter(query string)
func (t *Tree) ResetFilter()
```

`Filter`:
1. Store `filterQuery = query`.
2. Call `rebuild()` — the flattening pass already walks all nodes.
   During flattening, skip nodes whose text does not match `filterQuery`
   (and have no matching descendants). A node is included if it matches OR
   any of its descendants match (so the path to matching nodes stays visible).
3. Matching nodes have their parents auto-expanded for this render; the
   underlying `expanded` state is not mutated.

`ResetFilter`:
1. Set `filterQuery = ""`.
2. Call `rebuild()`.

## Builder

```go
func (b *Builder) Filter(id string) *Builder
```

## Implementation plan

1. **`filterable.go`** — new file: define `Filterable` interface.

2. **`filter.go`** — new file: define `Filter` struct, `NewFilter`, `Bind`,
   `Unbind`, `Bound`, `Clear`, `applyFilter`, `Apply`.

3. **`list.go`** — add `original []string` field; implement `Filter` and
   `ResetFilter`.

4. **`tree.go`** — add `filterQuery string` field; extend `rebuild` to skip
   non-matching subtrees; implement `Filter` and `ResetFilter`.

5. **`builder.go`** — add `Filter` method.

6. **Tests** — `filter_test.go`
   - Typing in a bound Filter calls `Filter` on the List with the correct query.
   - Clearing the Filter calls `ResetFilter` and restores original List items.
   - `Unbind` restores the List and detaches.
   - Tree filter exposes parent nodes of matching descendants.
   - Empty query always calls `ResetFilter`, never `Filter("")`.

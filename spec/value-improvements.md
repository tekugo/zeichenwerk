Now I have the full picture. Here is a comprehensive review with
recommendations:

---

## Analysis of `value.go` and Two-Way Bindings

### Current Architecture

`Value[T]` provides:

- **Reactive storage** with subscribers
- **Bind**: value → widget (requires `Setter[T]` with `Set(T)` method)
- **Observe**: widget → value (listens to `EvtChange`, extracts `params[0]` as
  `T` via type assertion or custom converter)
- **Derived**: computed values from source

This gives **one-way** binding in both directions, but not a unified two-way
primitive.

### Gap Analysis

| Widget                   | One-way ✓ | Two-way ✗ | Missing pieces                                                               |
| ------------------------ | --------- | --------- | ---------------------------------------------------------------------------- |
| Input                    | ✓         | ✓ (works) | - already implements Setter[string] + EvtChange→string; bidirectional works  |
| Checkbox                 | ✓         | ✓ (works) | -                                                                            |
| List                     | ✓         | ✗         | - No `SetSelected(int)` method; selection only via `Select(int)`             |
| Table                    | ✓         | ✗         | - `Set(TableProvider)` not `Set(...T)`; selection lacks setter               |
| Tree                     | ✗         | ✗         | - No `Set` for data; selection lacks setter; node expansion state            |
| Deck                     | ✓         | ✗         | - No `SetSelected` method                                                    |
| Select                   | ✗         | ✗         | - No `Set(value string)` method; uses `Select(value)` instead                |
| Text                     | ✓         | ✗         | - `Set([]string)` not `Set(value T)` for a single string; no line-level sync |
| Static/Digits            | only push | ✗         | - Display-only; never two-way                                                |
| Editor                   | ✗         | ✗         | - `Load(string)` not `Set(string)`; no selection/cursor binding              |
| Progress/Spinner/Scanner | ✗         | ✗         | - Not value-editable                                                         |

**Current two-way success cases**: Input, Checkbox work because they have:

- `Set(value)` that updates display
- `EvtChange` that emits the new value type

For those, you can do:

```go
v := NewValue("hello")
v.Bind(input)     // value → input
v.Observe(input)  // input → value
```

This is effectively two-way.

---

## Recommendations for `value.go` (Core Enhancements)

### 1. Add `BindBidirectional` Convenience

While `Bind` + `Observe` composes, a helper prevents feedback loops and
documents intent:

```go
// BindBidirectional connects a Value[T] to a widget both ways.
//   - value → widget: uses the Setter[T] interface (widget must implement Set(T))
//   - widget → value: listens to the provided event (default EvtChange)
// The function returns an unbinder that disconnects both directions.
func (v *Value[T]) BindBidirectional(
    w Setter[T],
    widget Widget,
    event Event,
    opts ...Option,
) *Binding {
    // Implementation detail: guard against feedback loops
    // (if the widget's Set method would emit the event we listen to)
    syncing := false
    unbindWidget := widget.On(event, func(_ Widget, _ Event, d ...any) bool {
        if syncing {
            return false
        }
        if val, ok := d[0].(T); ok && !reflect.DeepEqual(v.Get(), val) {
            // value ← widget
            v.Set(val)
        }
        return false
    })
    // value → widget
    v.Bind(w)
    // Return composite unbinder
    return &Binding{
        unbind1: func() { /* remove widget event handler */ },
        unbind2: func() { /* remove v subscriber */ },
    }
}
```

### 2. `Observe` Conversion Improvements

Current `Observe` uses `params[0].(T)` which panics on mismatch. Safer:

```go
func (v *Value[T]) Observe(widget Widget, convert ...func(any) (T, bool)) *Value[T] {
    widget.On(EvtChange, func(_ Widget, _ Event, params ...any) bool {
        if len(params) == 0 {
            return false
        }
        var val T
        var ok bool
        if len(convert) > 0 {
            val, ok = convert[0](params[0])
        } else {
            // Use a safe conversion helper that returns false on type mismatch
            val, ok = safeConvert[T](params[0])
        }
        if ok {
            v.Set(val)
        }
        return false
    })
    return v
}
```

Add `safeConvert[T](any) (T, bool)` using `reflect` or a type-switch cache for
built-in types.

### 3. Derive vs Map

`Derived` creates a new `Value[U]` from source. Rename to `Map` for familiarity
or keep both:

```go
func Map[A, B any](src *Value[A], fn func(A) B) *Value[B] { /* same */ }
```

### 4. Add `SubscribeOnce` for init patterns

Useful for "initial value only" handlers:

```go
func (v *Value[T]) SubscribeOnce(fn func(T)) {
    once := sync.Once{}
    v.Subscribe(func(t T) { once.Do(func() { fn(t) }) })
}
```

### 5. Channel-based observation

```go
func (v *Value[T]) Channel() <-chan T {
    ch := make(chan T, 1)
    var unsub func()
    unsub = v.Subscribe(func(t T) {
        select {
        case ch <- t:
        default:
        }
    })
    // Note: caller should retain unsub to close channel when done?
    return ch
}
```

---

## New `binding.go` File: Widget-Specific Adapters

### Selection Binding

Create helpers for widgets exposing selection index:

```go
// BindListSelection binds a Value[int] to a List's highlighted item.
// The Value is updated on EvtSelect; setting the Value calls list.Select().
func BindListSelection(v *Value[int], list *List) *Binding {
    // widget → value
    unbind1 := list.On(EvtSelect, func(_ Widget, _ Event, d ...any) bool {
        if idx, ok := d[0].(int); ok && v.Get() != idx {
            v.Set(idx)
        }
        return false
    })
    // value → widget
    unbind2 := v.Subscribe(func(idx int) {
        if list.Selected() != idx {
            list.Select(idx)
        }
    })
    return newBinding(unbind1, unbind2)
}

// BindTreeSelection binds a Value[*TreeNode] to a Tree's selected node.
// Node identity is used; nil means no selection.
func BindTreeSelection(v *Value[*TreeNode], tree *Tree) *Binding {
    unbind1 := tree.On(EvtSelect, func(_ Widget, _ Event, d ...any) bool {
        if node, ok := d[0].(*TreeNode); ok && v.Get() != node {
            v.Set(node)
        }
        return false
    })
    unbind2 := v.Subscribe(func(node *TreeNode) {
        if tree.Selected() != node {
            tree.Select(node)
        }
    })
    return newBinding(unbind1, unbind2)
}

// BindTableSelection binds a Value[TablePos] (row/col) to Table selection.
type TablePos struct{ Row, Col int }

func BindTableSelection(v *Value[TablePos], table *Table) *Binding {
    unbind1 := table.On(EvtSelect, func(_ Widget, _ Event, d ...any) bool {
        // Table EvtSelect data is (int, int) row,col
        if len(d) >= 2 {
            r, rok := d[0].(int)
            c, cok := d[1].(int)
            if rok && cok {
                pos := TablePos{r, c}
                if v.Get() != pos {
                    v.Set(pos)
                }
            }
        }
        return false
    })
    unbind2 := v.Subscribe(func(pos TablePos) {
        if table.Selected() != (pos) {
            table.SetSelected(pos.Row, pos.Col)
        }
    })
    return newBinding(unbind1, unbind2)
}
```

### Items/Data Binding

For List and Deck (items as `[]T`):

```go
// BindListItems binds a Value[[]string] to a List.
// Changes to the slice replace the entire list.
func BindListItems(v *Value[[]string], list *List) *Binding {
    // initial
    list.Set(v.Get())
    // value → list
    unbind1 := v.Subscribe(func(items []string) {
        list.Set(items)
    })
    // list → value: List doesn't modify items internally; no reverse needed
    return newBinding(unbind1, nil)
}

// BindDeckItems binds a Value[[]T] to a Deck with a custom render function.
// The Deck's items are replaced when value changes. Selection is separate.
func BindDeckItems[T any](
    v *Value[[]T],
    deck *Deck,
    render ItemRender,
) *Binding {
    // initial: deck.Set expects []any, need conversion
    anyItems := make([]any, len(v.Get()))
    for i, item := range v.Get() {
        anyItems[i] = item
    }
    deck.Set(anyItems)
    // value → deck
    unbind1 := v.Subscribe(func(items []T) {
        anyItems := make([]any, len(items))
        for i, item := range items {
            anyItems[i] = item
        }
        deck.Set(anyItems)
    })
    return newBinding(unbind1, nil)}
```

### Table Data Binding with Model

Create a `TableModel[T]` that adapts a `Value[[]T]` to `TableProvider`:

```go
// TableModel adapts a Value[[]T] and an extractor to a TableProvider.
type TableModel[T any] struct {
    data     *Value[[]T]
    columns  []Column
    extract  func(T) []string  // converts row T to []string cell values
    onChange func()
}

func NewTableModel[T any](
    data *Value[[]T],
    columns []Column,
    extract func(T) []string,
) *TableModel[T] {
    m := &TableModel[T]{data: data, columns: columns, extract: extract}
    data.Subscribe(func([]T) {
        if m.onChange != nil {
            m.onChange()
        }
    })
    return m
}

func (m *TableModel[T]) Columns() []TableColumn {
    cols := make([]TableColumn, len(m.columns))
    copy(cols, m.columns)
    return cols
}
func (m *TableModel[T]) Length() int { return len(m.data.Get()) }
func (m *TableModel[T]) Str(row, col int) string {
    rows := m.data.Get()
    if row < 0 || row >= len(rows) {
        return ""
    }
    cells := m.extract(rows[row])
    if col < 0 || col >= len(cells) {
        return ""
    }
    return cells[col]
}

// SetOnChange registers a callback when underlying data changes.
func (m *TableModel[T]) SetOnChange(fn func()) { m.onChange = fn }
```

Usage:

```go
type Person struct{ Name, Age string }
people := NewValue([]Person{{"Alice","30"}, {"Bob","25"}})
model := NewTableModel(people,
    []TableColumn{{Header:"Name", Width:10}, {Header:"Age", Width:5}},
    func(p Person) []string { return []string{p.Name, p.Age} })
table.Set(model)
people.OnChange(func() { table.Refresh() }) // or use model.SetOnChange
```

Add a helper to wire refresh automatically:

```go
func BindTableModel[T any](table *Table, model *TableModel[T]) *Binding {
    // When model data changes, refresh table
    unbind := model.data.Subscribe(func([]T) { table.Refresh() })
    return &Binding{unbind: unbind}
}
```

### Tree Model

For Tree, nodes themselves hold state (expanded, disabled). Can have a
`TreeModel[T]` where each node wraps a `T`:

```go
type TreeModel[T any] struct {
    root     *TreeNode
    getText  func(T) string
    getData  func(T) any
    getChildren func(T) []T
}

// Build tree from a value slice and rebuilds when slice changes.
func NewTreeModel[T any](
    roots []T,
    text func(T) string,
    children func(T) []T,
) *TreeModel[T] {
    // Build initial TreeNode tree
    // Subscribe to updates in the underlying data slice to rebuild
}
```

This gets more involved; perhaps keep Tree node mutation manual.

---

## Data Conversion Enhancements

The existing `Observe(..., convert func(any) (T, bool))` is good. Add a library
of converters:

```go
// Converters
func ConvertToInt(src any) (int, bool) {
    switch v := src.(type) {
    case int:
        return v, true
    case string:
        i, err := strconv.Atoi(v)
        return i, err == nil
    case float64:
        return int(v), true
    default:
        return 0, false
    }
}
func ConvertToString(src any) (string, bool) {
    if s, ok := src.(string); ok {
        return s, true
    }
    return fmt.Sprintf("%v", src), true // fallback
}
func ConvertToBool(src any) (bool, bool) {
    if b, ok := src.(bool); ok {
        return b, true
    }
    return false, false
}
```

Add `ObserveInt`, `ObserveString`, `ObserveBool` methods that use these
internally to avoid type assertions in user code:

```go
func (v *Value[int]) ObserveInt(widget Widget) *Value[int] {
    return v.Observe(widget, ConvertToInt)
}
func (v *Value[string]) ObserveString(widget Widget) *Value[string] {
    return v.Observe(widget, ConvertToString)
}
// etc.
```

---

## Validation

```go
type Validator[T any] func(T) error

func (v *Value[T]) SetValidated(val T, validate Validator[T]) error {
    if err := validate(val); err != nil {
        return err
    }
    v.Set(val)
    return nil
}

// Subscribe with validation filter
func (v *Value[T]) SubscribeValidated(fn func(T), validate Validator[T]) *Value[T] {
    return v.Subscribe(func(t T) {
        if err := validate(t); err == nil {
            fn(t)
        }
    })
}
```

Could integrate into `BindBidirectional`:

```go
err := v.BindBidirectional(..., WithValidator(myValidator))
```

---

## Performance: Debounce Throttle

For high-frequency events (typing, scrolling), use `Derived` with a time.After:

```go
func (v *Value[T]) Debounced(d time.Duration) *Value[T] {
    out := NewValue(v.Get())
    var timer *time.Timer
    v.Subscribe(func(val T) {
        if timer == nil {
            timer = time.NewTimer(d)
        } else {
            timer.Reset(d)
        }
        go func() {
            <-timer.C
            out.Set(val)
        }()
    })
    return out
}
```

But that's more advanced; could live in a separate `reactive` package.

---

## Widgets That Need Adapters for Setter[T]

Widgets with a `Set` method but different signature:

- **List**: `Set([]string)` - already matches `Setter[[]string]`
- **Text**: `Set([]string)` - matches `Setter[[]string]`
- **Static**: `Set(any)` - does NOT satisfy `Setter[string]` (type mismatch)
- **Button**: `Set(string)` - matches `Setter[string]`
- **Digits**: inherits Static's Set(any)
- **Styled**: `SetText(string)` not `Set`

Ways to handle:

1. Add `Set(string)` to Static (already exists as `Set(any)`). Could add
   `SetValue(string)` that's typed and have `Set(any)` call it. But better:
   change `Static.Set` signature to `Set(value any)` is intentionally flexible;
   to satisfy Setter[string], we'd need a wrapper:

```go
type StaticStringSetter struct{ *Static }
func (s *StaticStringSetter) Set(v string) { s.Static.Set(v) }
```

2. **Select**: has `Select(string)` not `Set`. Provide adapter:

```go
type SelectSetter struct{ *Select }
func (s *SelectSetter) Set(v string) { s.Select(v) }
```

3. **Editor**: Could add `SetText(string)` that calls `Load(string)`. That would
   satisfy Setter[string].

4. **Table**: `Set(TableProvider)` - can't be Setter of a simple type. Need a
   custom model adapter as described.

5. **Tree**: No Setter at all.

6. **Checkbox**: already has `Set(bool)`, good.

7. **Input**: already has `Set(string)`, good.

---

## Proposed Changes Summary

### value.go – Small API additions

1. Add `BindBidirectional(w Setter[T], widget Widget, event Event) *Binding`  
   (with loop prevention and unbind support)

2. Add typed Observe helpers:

   ```go
   func (v *Value[int]) ObserveInt(widget Widget) *Value[int]
   func (v *Value[string]) ObserveString(widget Widget) *Value[string]
   func (v *Value[bool]) ObserveBool(widget Widget) *Value[T]
   ```

   (implement via generic helpers to avoid duplication)

3. Add `Channel() <-chan T`

4. Add `SetIfChanged(val T) bool` (uses `cmp.Equal` for slices? Or deep equality
   for simple types only; maybe not)

5. Improve `Observe` conversion safety:
   - Use a safeCast helper internally that returns false on mismatch instead of
     panic.
   - Provide `ObserveWith(conv func(any) (T, error))` variant for error-aware
     conversion.

6. Add unsubscribe capability: Currently `Subscribe` returns `*Value[T]` for
   chaining, but not an unsubscribe function. Change to:

   ```go
   func (v *Value[T]) Subscribe(fn func(T)) Subscription {
       v.mu.Lock()
       v.subscribers = append(v.subscribers, fn)
       v.mu.Unlock()
       fn(v.value)
       return &sub{parent: v, fn: fn}
   }
   type sub struct { parent *Value[T]; fn func(T) }
   func (s *sub) Unsubscribe() { /* remove from subscribers */ }
   ```

   This enables manual unbinding for `BindBidirectional` cleanup.

### New `binding.go` – Widget Adapters

Helper constructors for common widget bindings:

#### Simple value widgets

```go
func BindInput(v *Value[string], i *Input) *Binding
func BindCheckbox(v *Value[bool], cb *Checkbox) *Binding
func BindStatic(v *Value[string], s *Static) *Binding
func BindButtonLabel(v *Value[string], b *Button) *Binding
```

#### Selection widgets

```go
func BindListSelection(v *Value[int], l *List) *Binding
func BindDeckSelected(v *Value[int], d *Deck) *Binding
func BindTabsSelected(v *Value[int], t *Tabs) *Binding
func BindTreeSelection(v *Value[*TreeNode], t *Tree) *Binding
func BindTableSelection(v *Value[TablePos], t *Table) *Binding
```

#### Multi-value widgets

```go
func BindListItems(v *Value[[]string], l *List) *Binding
func BindTextLines(v *Value[[]string], t *Text) *Binding  // Text is read-only for now? Could add editing later
```

#### Complex models

```go
func NewTableProviderFromValue[T any](
    data *Value[[]T],
    columns []Column,
    extract func(T) []string,
) *tableProviderFromValue[T]

func NewTreeModel[T any](
    roots []T,
    text func(T) string,
    children func(T) []T,
    id func(T) any, // optional identity for selection persistence
) *TreeModel[T]
```

---

## Two-Way Binding for Complex Widgets

### List: Items + Selection

Items binding is one-way (value → list) because List doesn't edit its own items
(user only selects). Selection is two-way:

```go
items := NewValue([]string{"A","B","C"})
BindListItems(items, list)

selected := NewValue(-1)
BindListSelection(selected, list)
```

### Table: Data + Selection

Two-way data binding would require Table to notify when its data changes
(editing). Table currently doesn't have built-in editing. To add editing:

- Implement `EditableTableProvider` interface:

```go
type EditableTableProvider interface {
    TableProvider
    SetCell(row, col int, value string)
}
```

- Then `BindEditableTable` allows both directions. But this is a new feature.

### Tree: Node Expansion + Selection

TreeNode already has `expanded` and `disabled` as fields. Could wrap with a
reactive node:

```go
type ReactiveTreeNode struct {
    *TreeNode
    expanded *Value[bool]  // when set, calls Expand/Collapse
}
```

Or add `OnChange` to TreeNode to notify listeners.

---

## Data Conversion and `any` Handling

The `Setter[T]` interface is strongly typed. For widgets that currently accept
`any` (Static, Text.Set), consider adding typed convenience wrappers:

```go
// Static implements Setter[string] via typed wrapper:
func (s *Static) SetString(text string) { s.Set(text) }
```

But you can't have two Set methods with different type parameters in the same
struct; they'd conflict in the interface implementation because both are `Set`
but different signatures? No, Go method signatures include parameter types;
`Set(any)` and `Set(string)` are different methods, so both can exist. However,
only one will satisfy `Setter[string]` (the `Set(string)` one). So we could add:

```go
func (s *Static) SetString(v string) { s.Text = v; s.Refresh() }
```

and have a wrapper type:

```go
type StaticStringSetter struct{ *Static }
func (s *StaticStringSetter) Set(v string) { s.Static.SetString(v) }
```

Better: change Static's Set to be generic? Not possible in Go (no generic method
on non-generic type). So adapter is fine.

---

## Suggestions for Tables & Trees (complex data)

### Table Model Approach

Instead of expecting Table to implement Setter[T], create a `TableModel[T]` that
acts as both `TableProvider` and `Value[[]T]` subscriber:

```go
type TableModel[T any] struct {
    data     *Value[[]T]
    columns  []TableColumn
    extract  func(T) []string
    // Optional: bijective mapping T ↔ cell editing?
}

func (m *TableModel[T]) sync() {
    // rebuild provider from m.data.Get()
    // notify Table via callback or implement TableProvider on the model itself
}
```

Make `TableModel[T]` implement `TableProvider`. The Table widget stores the
provider; when `data` changes, call `table.Set(model)` again (or have model
notify). Simpler: `BindTableModel` as earlier.

### Tree Node Identity

When binding Tree selection to a `Value[*TreeNode]`, be careful: node pointers
may become invalid if tree is rebuilt. Better to bind by path or ID:

```go
type TreeNodeID string // or int
func (n *TreeNode) ID() TreeNodeID { /* field */ }

func BindTreeSelectionByID(v *Value[TreeNodeID], tree *Tree) {
    // translate ID ↔ node lookup
}
```

---

## Specific Refactorings

### 1. Make `Observe` safe

Replace direct `params[0].(T)` with a helper:

```go
func cast[T any](x any) (T, bool) {
    v, ok := x.(T)
    return v, ok
}
```

Already essentially that; OK.

### 2. Add `TryConvert` utility

```go
func TryConvert[T any](src any, fn ...func(any) (T, bool)) (T, bool) {
    if len(fn) > 0 {
        return fn[0](src)
    }
    var zero T
    return zero, false
}
```

### 3. Add `SubscribeTransformer`

Allow transformations in subscription chain:

```go
func (v *Value[T]) SubscribeMap[U any](fn func(T) U) *Value[U] {
    out := NewValue[U](fn(v.Get()))
    v.Subscribe(func(t T) { out.Set(fn(t)) })
    return out
}
```

This is essentially `Derived` but with a different name; keep both.

### 4. Two-Way Binding for Editable Widgets

For widgets that need edit notifications (e.g., Table cell edit), the current
pattern uses `Observe` on the Value. But what if the widget modifies its own
data internally (like user editing a cell)? Currently Table has no editing. If
editing were added, you'd dispatch `EvtChange` with cell coordinates and new
value. Then `Observe` with custom converter could update the model.

---

## Concrete Suggestions for `value.go`

Given the scope, keep `value.go` minimal and create `binding.go` for
widget-specific glue. Here's the diff-style suggestion for `value.go`:

```diff
@@
 func (v *Value[T]) Observe(widget Widget, convert ...func(any) (T, bool)) *Value[T] {
     widget.On(EvtChange, func(_ Widget, _ Event, params ...any) bool {
         if len(params) == 0 {
             return false
         }
         var val T
         var ok bool
         if len(convert) > 0 {
-            val, ok = convert[0](params[0])
+            val, ok = convert[0](params[0])
         } else {
-            val, ok = params[0].(T)
+            val, ok = cast[T](params[0])
         }
         if ok {
             v.Set(val)
         }
         // We return false, so other EvtChange handlers still get called
         return false
     })
     return v
 }

+// ObserveWith converts event data using a converter that returns an error.
+// If the converter returns an error, the update is ignored (could log).
+func (v *Value[T]) ObserveWith(widget Widget, conv func(any) (T, error)) *Value[T] {
+    widget.On(EvtChange, func(_ Widget, _ Event, params ...any) bool {
+        if len(params) == 0 {
+            return false
+        }
+        if val, err := conv(params[0]); err == nil {
+            v.Set(val)
+        }
+        return false
+    })
+    return v
+}
+
 // Subscribe adds a new callback function for receiving updates.
 func (v *Value[T]) Subscribe(fn func(T)) *Value[T] {
     v.mu.Lock()
     v.subscribers = append(v.subscribers, fn)
@@
 func Derived[A, B any](source *Value[A], convert func(A) B) *Value[B] {
     derived := NewValue(convert(source.Get()))
     source.Subscribe(func(a A) {
         derived.Set(convert(a))
     })
     return derived
 }
+
+// Map is an alias for Derived, emphasizing transformation.
+func Map[A, B any](source *Value[A], convert func(A) B) *Value[B] {
+    return Derived(source, convert)
+}
+
+// Channel returns a read-only channel that emits every new value.
+// The channel has buffer size 1; if the consumer is slow, the latest value
+// overwrites pending ones.
+func (v *Value[T]) Channel() <-chan T {
+    ch := make(chan T, 1)
+    v.Subscribe(func(t T) {
+        select {
+        case ch <- t:
+        default:
+            // overwrite pending
+            select {
+            case ch <- t:
+            default:
+            }
+        }
+    })
+    return ch
+}
```

Add unsubscribe support:

```go
type Subscription interface {
    Unsubscribe()
}

type subscriberFunc[T any] struct {
    fn func(T)
}

func (s *subscriberFunc[T]) Unsubscribe() {
    // need parent reference; not trivial. Could return a closure:
    // Subscribe returns func() that unsubscribes.
}
```

Simpler: change `Subscribe` to return an unsubscribe function:

```go
func (v *Value[T]) Subscribe(fn func(T)) func() {
    v.mu.Lock()
    v.subscribers = append(v.subscribers, fn)
    v.mu.Unlock()
    fn(v.value)
    return func() {
        v.removeSubscriber(fn)
    }
}
```

That would require a `removeSubscriber` method. That's a useful addition.

```go
func (v *Value[T]) Subscribe(fn func(T)) func() {
    v.mu.Lock()
    defer v.mu.Unlock()
    v.subscribers = append(v.subscribers, fn)
    fn(v.value)
    return func() {
        v.mu.Lock()
        defer v.mu.Unlock()
        for i, f := range v.subscribers {
            if f == fn {
                v.subscribers = append(v.subscribers[:i], v.subscribers[i+1:]...)
                break
            }
        }
    }
}
```

Now `BindBidirectional` can return these unsubscribers.

---

## binding.go – New File

Create a new file `binding.go` with all widget adapters. Example items:

```go
package zeichenwerk

// Binding represents an active two-way binding that can be terminated.
type Binding struct {
    unbind1 func()
    unbind2 func()
}

func (b *Binding) Unbind() {
    if b.unbind1 != nil { b.unbind1() }
    if b.unbind2 != nil { b.unbind2() }
}

// Input ↔ string
func BindInput(v *Value[string], i *Input) *Binding {
    v.Bind(i)
    unbind := v.Observe(i) // typed Observe? We can add ObserveString
    return &Binding{unbind1: func(){}, unbind2: unbind}
}

// Checkbox ↔ bool
func BindCheckbox(v *Value[bool], cb *Checkbox) *Binding {
    v.Bind(cb)
    unbind := v.Observe(cb)
    return &Binding{unbind1: func(){}, unbind2: unbind}
}

// List selection
func BindListSelection(v *Value[int], l *List) *Binding {
    // value → widget
    unbindValue := v.Subscribe(func(idx int) {
        if l.Selected() != idx {
            l.Select(idx)
        }
    })
    // widget → value
    unbindWidget := l.On(EvtSelect, func(_ Widget, _ Event, d ...any) bool {
        if idx, ok := d[0].(int); ok && v.Get() != idx {
            v.Set(idx)
        }
        return false
    })
    return &Binding{unbind1: unbindValue, unbind2: unbindWidget}
}

// Tabs selected
func BindTabsSelected(v *Value[int], t *Tabs) *Binding {
    unbindValue := v.Subscribe(func(idx int) {
        if t.Get() != idx {
            t.Set(idx)
        }
    })
    unbindWidget := t.On(EvtChange, func(_ Widget, _ Event, d ...any) bool {
        if idx, ok := d[0].(int); ok && v.Get() != idx {
            v.Set(idx)
        }
        return false
    })
    return &Binding{unbind1: unbindValue, unbind2: unbindWidget}
}

// Select (dropdown) selected value
type SelectSetter struct{ *Select }
func (s *SelectSetter) Set(v string) { s.Select(v) }

func BindSelect(v *Value[string], s *Select) *Binding {
    v.Bind(&SelectSetter{s})
    unbindWidget := s.On(EvtChange, func(_ Widget, _ Event, d ...any) bool {
        if val, ok := d[0].(string); ok && v.Get() != val {
            v.Set(val)
        }
        return false
    })
    return &Binding{unbind1: func(){}, unbind2: unbindWidget}
}

// Deck selected
func BindDeckSelected(v *Value[int], d *Deck) *Binding {
    unbindValue := v.Subscribe(func(idx int) {
        if d.Selected() != idx {
            d.Select(idx)
        }
    })
    unbindWidget := d.On(EvtSelect, func(_ Widget, _ Event, d ...any) bool {
        if idx, ok := d[0].(int); ok && v.Get() != idx {
            v.Set(idx)
        }
        return false
    })
    return &Binding{unbind1: unbindValue, unbind2: unbindWidget}
}

// Table selection binding
type TablePos struct{ Row, Col int }

func BindTableSelection(v *Value[TablePos], t *Table) *Binding {
    unbindValue := v.Subscribe(func(pos TablePos) {
        if t.Selected() != (pos) {
            t.SetSelected(pos.Row, pos.Col)
        }
    })
    unbindWidget := t.On(EvtSelect, func(_ Widget, _ Event, d ...any) bool {
        if len(d) >= 2 {
            if r, ok1 := d[0].(int); ok1 {
                if c, ok2 := d[1].(int); ok2 {
                    pos := TablePos{r, c}
                    if v.Get() != pos {
                        v.Set(pos)
                    }
                }
            }
        }
        return false
    })
    return &Binding{unbind1: unbindValue, unbind2: unbindWidget}
}

// Table data binding using a model
func BindTableModel[T any](t *Table, model *TableModel[T]) *Binding {
    // Initial set
    t.Set(model)
    // Refresh table when data changes
    unbind := model.data.Subscribe(func([]T) { t.Refresh() })
    return &Binding{unbind1: unbind, unbind2: func(){}} // maybe also watch selection?
}
```

---

## Tree Two-Way Binding (Advanced)

Tree is the hardest because nodes are not indexed; selection is a `*TreeNode`.
If you rebuild the tree, old pointers become stale. Need either:

- Stable node identity (each TreeNode has a unique ID)
- Or selection by path (string like "0/2/1")

**Option A**: Bind by pointer (simple, fragile):

```go
func BindTreeSelection(v *Value[*TreeNode], t *Tree) *Binding {
    unbindValue := v.Subscribe(func(node *TreeNode) {
        t.Select(node)
    })
    unbindWidget := t.On(EvtSelect, func(_ Widget, _ Event, d ...any) bool {
        if node, ok := d[0].(*TreeNode); ok && v.Get() != node {
            v.Set(node)
        }
        return false
    })
    return &Binding{unbind1: unbindValue, unbind2: unbindWidget}
}
```

Works as long as tree structure doesn't get rebuilt fully (e.g., via `SetRoot`).
If user rebuilds tree, need to re-bind or ensure node identity preserved.

**Option B**: Bind by ID:

```go
type NodeID string
func (n *TreeNode) ID() NodeID { return NodeID(fmt.Sprintf("%p", n)) } // or user-provided

func BindTreeSelectionByID(v *Value[NodeID], t *Tree) *Binding {
    // value → node
    unbindValue := v.Subscribe(func(id NodeID) {
        node := findNodeByID(t.Root(), id)
        if node != nil {
            t.Select(node)
        }
    })
    // node → id
    unbindWidget := t.On(EvtSelect, func(_ Widget, _ Event, d ...any) bool {
        if node, ok := d[0].(*TreeNode); ok {
            id := NodeID(fmt.Sprintf("%p", node))
            if v.Get() != id {
                v.Set(id)
            }
        }
        return false
    })
    return &Binding{unbind1: unbindValue, unbind2: unbindWidget}
}
```

But this uses pointer address as ID; not stable across tree rebuilds. Better to
have user-provided ID.

So suggest: extend `TreeNode` with an optional `id string` field.

---

## Recommendations Summary

**value.go (core)**:

1. Add `BindBidirectional` helper with loop detection (optional but nice)
2. Add `Unsubscribe()` return from `Subscribe` (instead of just chaining)
3. Add typed observe shortcuts: `ObserveInt`, `ObserveString`, `ObserveBool` (or
   keep generic but provide safe converters)
4. Add `Map` alias for `Derived` and `Channel()`

**binding.go (new)**: 5. Provide adapter functions for all widgets with `Set`
and `EvtChange`:

- Input, Checkbox, Button, Static (via wrapper), Text? (lines vs string?), maybe
  not Text

6. Provide selection binding for:
   - List, Deck, Tabs, Tree (pointer or ID-based), Table (TablePos)
7. Provide items binding for:
   - List (items slice), Deck (items slice)
8. Provide table model adapters:
   - `TableModel[T]` that implements TableProvider and syncs with Value[[]T]
9. Provide tree model adapter (if feasible)
10. Provide `BindEditableTable` if/when Table gains editing capability

**Widget method additions** (to satisfy Setter[T]): 11. Add `SetText(string)` to
Editor (calls `Load(string)`), making it `Setter[string]` 12. Add
`SetItems([]string)` to List is already `Set([]string)`, so `Setter[[]string]`
works 13. Add `SetSelected(int)` to List, Deck, Tabs to bind selection directly
without wrapper:

```go
func (l *List) SetSelected(idx int) { l.Select(idx) }
```

Then BindListSelection can simply `v.Bind(list)` plus observe. 14. Consider
adding `SetProvider(TableProvider)` already exists; but binding to a
`Value[[]T]` would need adapter.

**Event data standardization**: 15. Consider defining specific event payload
types via generics: `EvtChange[T]`? Not feasible because Event is string
constant. But we can document payload types per widget and provide typed helper
OnChange[T].

Add:

```go
func OnChange[T any](w Widget, fn func(T)) {
    w.On(EvtChange, func(_ Widget, _ Event, d ...any) bool {
        if len(d) > 0 {
            if v, ok := d[0].(T); ok {
                fn(v)
            }
        }
        return false
    })
}
```

Similar to `OnChange` helper already in helper.go but that only takes
`func(string)`. So we could generalize:

```go
// OnChangeTyped registers a change handler with type-safe payload.
func OnChange[T any](w Widget, fn func(T)) {
    w.On(EvtChange, func(_ Widget, _ Event, d ...any) bool {
        if v, ok := d[0].(T); ok {
            fn(v)
        }
        return false
    })
}
```

But the existing `OnChange(widget, handler func(string) bool)` is already
type-asserted to string. That's fine for Input/Select. For Checkbox, EvtChange
payload is bool. So we need both. Could rename existing to `OnChangeString` and
add generic `OnChange[T]`. Yet that would be breaking change. Keep both; binding
code uses whichever.

**Form integration**: 16. Form already uses reflection to bind struct fields to
widgets. Could extend to bind to `Value[T]` fields automatically:

```go
form.BindValues(&myViewModel) // where myViewModel has fields of type *Value[T]
```

That would be a higher-level feature.

---

## Specific Code Suggestions

### value.go: Add `SetIfChanged`

```go
import "reflect"

func (v *Value[T]) SetIfChanged(new T) bool {
    if reflect.DeepEqual(v.Get(), new) {
        return false
    }
    v.Set(new)
    return true
}
```

But note: `reflect.DeepEqual` is slow for large slices; optional optimization:
user can compare themselves. Could also do pointer equality for slices.

### value.go: Add `CompareAndSwap` (CAS)

```go
func (v *Value[T]) CompareAndSwap(old, new T) bool {
    v.mu.Lock()
    defer v.mu.Unlock()
    if reflect.DeepEqual(v.value, old) {
        v.value = new
        // notify subscribers async?
        return true
    }
    return false
}
```

Rarely needed.

### value.go: Make `Derived` automatically unsubscribe from source on no subscribers?

Not needed; keep simple.

### binding.go: Use a struct to group unbinds

```go
type Binding struct {
    unbinds []func()
}

func (b *Binding) Unbind() {
    for _, u := range b.unbinds {
        u()
    }
    b.unbinds = nil
}
```

Then each binding function returns `*Binding`.

---

## Prioritization

**High Value, Low Effort**:

1. Add `Subscribe` returning unsubscribe function
2. Add `BindBidirectional` (even if just composes Bind+Observe with loop guard)
3. Add `OnChangeTyped[T]` helper (generic) to `helper.go`
4. Add typed Observe shortcuts (`ObserveInt`, `ObserveString`, `ObserveBool`)
   via generics:

   ```go
   func (v *Value[int]) ObserveInt(w Widget) *Value[int] { return v.Observe(w) }
   ```

   (Already works but type inference may need explicit type; not necessary)

**Medium Value**: 5. TableModel[T] for reactive table data 6. BindListSelection,
BindTabsSelected, BindDeckSelected, BindTreeSelection, BindTableSelection
adapters 7. Add `SetSelected(int)` to List/Deck/Tabs to make them `Setter[int]`
for selection 8. Add `SetText(string)` to Editor to make it `Setter[string]` 9.
Add `Channel()` for streaming

**Lower Value / More Complex**: 10. TreeModel with node identity 11. Editable
TableProvider 12. Debounced derived values 13. Validation framework

---

## Conclusion

The `value.go` reactive core is well-designed for its scope. The main
improvement area is **binding ergonomics** for complex widgets (selection,
items, table rows). I recommend:

1. **Enhance `Value` with unsubscribe returns** (so bindings can be torn down).
2. **Add `BindBidirectional`** as a convenience.
3. **Add a new `binding.go`** with per-widget binding helpers (selection, items,
   table model).
4. **Add `SetSelected` methods** to selection widgets to unify the API.
5. **Add typed `OnChange[T]`** helper in `helper.go`.
6. **Consider `TableModel[T]`** to bridge Go structs to Table easily.

The existing `Observe` conversion pattern is good; just add typed convenience
wrappers and maybe a safer default converter.

Would you like me to implement any of these specific suggestions?

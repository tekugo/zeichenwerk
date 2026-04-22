I've reviewed `value.go`. Here are my suggestions for improvements:

## 1. **Missing `setter` interface definition**

The `Setter[T]` interface is used in `Bind()` but not defined in this file.
Check if it's defined elsewhere, but consider documenting or making it explicit:

```go
// Setter is an interface for objects that can receive value updates.
type Setter[T any] interface {
    Set(T)
}
```

## 2. **`Observe` method naming and directionality**

`Observe` suggests reading from the widget into the `Value`, but the naming is
ambiguous. Consider:

- Renaming to `Watch` or `SyncFrom` for clarity
- Adding a comment explaining the data flow direction

## 3. **No way to unsubscribe**

There's no `Unsubscribe` or `Unbind` method. Long-lived applications could leak
memory if callbacks accumulate:

```go
// Unsubscribe removes a previously registered callback.
func (v *Value[T]) Unsubscribe(fn func(T)) {
    v.mu.Lock()
    defer v.mu.Unlock()
    // find and remove fn from subscribers
}
```

## 4. **`Derived` should be exported consistently**

`Derived` is exported but doesn't follow the `New...` naming convention.
Consider:

- `NewDerived[A, B any](...)` for consistency with `NewValue`
- Or document why it breaks the pattern

## 5. **No `CompareAndSet` or change detection**

`Set` always notifies subscribers even if the value hasn't changed. For
expensive operations or avoiding unnecessary re-renders:

```go
// SetIfChanged sets the value only if it differs from the current one.
func (v *Value[T]) SetIfChanged(val T) bool
```

## 6. **Missing `Reset` or `Swap` operations**

Common patterns that could be useful:

- `Reset()` to zero value
- `Swap(new T) (old T)` for atomic exchange

## 7. **`Bind` returns `*Value[T]` for chaining, but `Subscribe` does too**

This is inconsistent with typical Go patterns. Consider:

- Returning `func()` (unsubscribe handle) from `Subscribe`/`Bind`
- Or document the chaining intent clearly

---

Would you like me to implement any of these suggestions?

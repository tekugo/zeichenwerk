# Zeichenwerk – Code Review

**Date:** 2026-03-26
**Last updated:** 2026-03-26
**Scope:** Full codebase review — bugs, design issues, performance, and improvement suggestions.

---

## Summary

Zeichenwerk is a well-structured Go TUI framework with a clean Widget/Container interface hierarchy, a CSS-like theming system, and a fluent builder API. The overall architecture is sound, but several concrete bugs and design issues exist that should be addressed before wider use.

---

## 1. Bugs and Correctness Errors

### ✅ 1.1 Unicode Bug in `Input` — `len(i.text)` vs. rune count

**Fixed.** The `Input` widget was rewritten to use `GapBuffer` as its backing store (see 2.2). All boundary checks now use `buf.Length()`, which operates on rune counts. The `unicode/utf8` import was removed.

---

### ✅ 1.2 Off-by-one in `Table.Hint()` — separator count

**Fixed.** `table.go:90`: condition is now `if i > 0`, counting separators between every pair of adjacent columns.

---

### ✅ 1.3 `Table.Set()` accumulates `tableWidth` on repeated calls

**Fixed.** `table.go:64`: `t.tableWidth = 0` is now the first statement in `Set()`.

---

### ✅ 1.4 Data race in `Animation`

**Fixed.** `animation.go`: A `sync.Mutex` field (`mu`) was added to `Animation`. The running check and `ticker` assignment now happen under the mutex on the caller side (before the goroutine is launched), eliminating the TOCTOU race where two concurrent `Start()` calls could both pass the nil check. `ticker.Stop()` and the nil assignment on teardown are also performed under the mutex inside a deferred function in the goroutine. `Stop()` no longer writes `a.ticker = nil` directly; `Running()` acquires the mutex before reading `ticker`.

---

### 1.5 `builder.go:Build()` always enables debug mode

**File:** `builder.go:58`

```go
func (b *Builder) Build() *UI {
    ui, _ := NewUI(b.theme, b.stack.Peek(), true)  // debug always true
    return ui
}
```

The `debug: true` flag activates the debug status bar and verbose slog output. Every application built with the builder runs in debug mode with no way to disable it.

**Fix:** Accept a `debug bool` parameter in `Build()`, or default to `false`.

---

### ✅ 1.6 `WidgetType()` strips wrong package prefix

**Fixed.** `helper.go:86`: now uses `"*zeichenwerk."` as the trim prefix.

---

### 1.7 `Table.column` field is declared but never read

**File:** `table.go:12`

```go
row, column int  // column is never used anywhere in table.go
```

The `column` field was presumably intended for column-focused navigation (Tab/Shift+Tab keys also call `moveToColumn`/`scrollByColumn`, not `column` tracking). Dead state adds confusion about the widget's data model.

**Fix:** Remove the `column` field or implement column-tracking properly.

---

### ✅ 1.8 `NewUI()` log widget lookup happens before parent is set

**Fixed.** `ui.go:144`: `root.SetParent(ui)` now occurs before the `Find(ui, "debug-log")` call at line 160.

---

## 2. Design Issues

### 2.1 `Builder.Apply()` panics on unknown widget types

**File:** `builder.go:202`

```go
default:
    panic(fmt.Errorf("no style for widget type %T", widget))
```

Any custom widget passed to the builder causes a panic. This forces every third-party widget author to fork the builder or avoid the fluent API entirely.

**Suggestion:** Log a warning and apply no style rather than panicking. Alternatively, add an `Styleable` interface that widgets can implement to self-register their styling.

---

### ✅ 2.2 `Input` doesn't use `GapBuffer`

**Fixed.** `Input` now uses `*GapBuffer` as its backing store instead of a plain `string`. The `text` field was replaced with `buf *GapBuffer`, initialized via `NewGapBufferFromString`. All editing operations (`Insert`, `Delete`, `DeleteForward`) call `buf.Move(pos)` then the appropriate gap buffer method, giving O(1) insertions and deletions at the cursor. Bulk operations (`CtrlK`, `CtrlU`, `Clear`, `SetText`) rebuild or drain the buffer. The `[]rune` conversions on every keystroke were eliminated.

---

### 2.3 `Refresh()` vs `Redraw()` inconsistency

`Widget.Refresh()` chains up the parent hierarchy to `UI.Refresh()`, which triggers a *full screen redraw*. The optimized single-widget path (`UI.Redraw(widget)`) is only reachable via the package-level `Redraw(widget)` helper. The two code paths are easy to confuse:

- `input.Refresh()` → `UI.Refresh()` → full screen
- `Redraw(input)` → `UI.Redraw(input)` → single widget

The naming suggests `Redraw` is more expensive than `Refresh`, but it's the opposite. Widgets like `Input` and `Table` already override `Refresh()` to call `Redraw(t)` — this pattern should be documented as the idiomatic way to do partial redraws.

---

### 2.4 Side effect in `Component.Style()` getter

**File:** `component.go:416`

```go
if c.styles[""] == nil {
    c.styles[""] = NewStyle("")
}
return c.styles[""]
```

A getter allocates and stores a new `Style` object as a side effect. This means that *calling `Style()`* modifies the widget's state, which is unexpected. It can also allocate a map entry for every un-styled component encountered during a render pass.

**Suggestion:** Return `&DefaultStyle` (or a package-level empty style) instead of creating and storing a new one.

---

### 2.5 `Builder.Spacer()` produces duplicate IDs

**File:** `builder.go:471`

```go
func (b *Builder) Spacer() *Builder {
    spacer := NewComponent("spacer")
    ...
}
```

Every spacer gets the same ID `"spacer"`. Since IDs are the primary lookup key (`Find`, theme selectors), multiple spacers in one layout break ID-based operations.

**Fix:** Use a counter or a UUID-like approach: `fmt.Sprintf("spacer-%d", b.spacerCount)`.

---

### 2.6 `Theme.Add()` discards `WithParent` return value

**File:** `theme.go:146`

```go
style.WithParent(parent)   // return value discarded
```

`WithParent` is a functional method — on a fixed style it returns a *new* child style. Here the return value is discarded. In practice this currently works only because `style` is not yet fixed when passed to `Add()`, making `Modifiable()` return `self`. But it is fragile: if a fixed style is added, the parent assignment silently has no effect.

**Fix:** `style = style.WithParent(parent)` (use the returned value).

---

### 2.7 `NewUI()` always returns `nil` error

**File:** `ui.go:130`

The function signature is `func NewUI(...) (*UI, error)` but the error is always `nil`. Callers must write `ui, _ := NewUI(...)` and silently ignore it. Either return a meaningful error (e.g. from logger setup), or simplify the signature to return just `*UI`.

---

### 2.8 Global key bindings conflict with app content

**File:** `ui.go:219-223`

```go
case tcell.KeyRune:
    switch event.Str() {
    case "q", "Q":
        close(ui.quit)
    }
```

When no widget handles the 'q' key (e.g. when focus is on a non-input widget), the application quits. This can be surprising. Additionally, pressing Escape closes the topmost layer, which could accidentally close modal dialogs the user did not intend to dismiss.

**Suggestion:** Make the quit key(s) configurable in `NewUI()`, and require explicit opt-in (or at least document this behavior prominently).

---

### 2.9 `HandleListEvent` is the only widget-specific event helper

**File:** `helper.go:100`

`HandleListEvent` is a convenience wrapper that exists only for `*List`. No equivalent helpers exist for `*Table`, `*Input`, `*Select`, etc. Either generalize the pattern (e.g., via generic event helpers) or remove it in favour of `.On("event", ...)` directly.

---

## 3. Performance

### 3.1 `Renderer.Text()` puts characters one-by-one

**File:** `renderer.go:226`

Every call to `Text()` iterates rune-by-rune and calls `Put()` (→ `TcellScreen.Put()` → `tcell.Screen.Put()`) for each character. For wide tables or long text widgets this is the hot path in the render loop. Batching multiple cells before flushing to tcell would reduce the overhead.

---

### 3.2 `GapBuffer.Runes()` spawns a goroutine per iteration

**File:** `gap-buffer.go:172`

```go
func (gb *GapBuffer) Runes(start int) <-chan rune {
    ch := make(chan rune)
    go func() { ... }()
    return ch
}
```

Goroutine + channel creation for simple sequential iteration is unnecessarily heavyweight. A closure-based iterator or a direct `[]rune` slice return would be far cheaper for small-to-medium buffers.

---

### 3.3 Style lookup uses regex on every call

**File:** `component.go:399`

`stylePartRegExp.FindStringSubmatch(actual)` runs a regex on every style lookup, which happens once per widget part per render frame. For UIs with many widgets this is measurable overhead. Since the selector format is simple (`part:state`), a `strings.Cut(actual, ":")` would be sufficient and much faster.

---

### 3.4 `Renderer.Fill()` and `Renderer.Colorize()` call `Put()` per cell

**File:** `renderer.go:91, 74`

Both methods loop over every cell individually. For large background fills (e.g. clearing a full-screen dialog), this causes many redundant style applications. Consider bulk-writing via tcell's `Fill` where applicable.

---

### 3.5 `Input` converts string to rune slice on every operation

**Partially addressed by 2.2.** `Insert`, `Delete`, `DeleteForward`, and cursor movement no longer perform `[]rune(i.text)` on the hot path. `visible()` and the bulk-delete shortcuts (`CtrlK`, `CtrlU`) still convert via `buf.String()` + `[]rune(...)`, but these are either render-time or infrequent user actions.

---

## 4. Suggestions for Improvement

### 4.1 Package-level documentation

`doc.go` exists but only contains the package declaration. A proper package comment describing the widget model, the builder pattern, the styling system, and a minimal example would significantly improve discoverability.

---

### 4.2 Thread safety documentation

Several components interact across goroutines (Animation ticks call `Redraw()`, which enqueues into `UI.redraw`). The threading model (what is safe to call from goroutines vs. what must run on the event loop) is not documented. A brief doc comment on `UI.Redraw()` and `UI.Refresh()` explaining goroutine safety would reduce bugs for users.

---

### 4.3 `Table` should expose total row count and scroll state

`Table` has `GetScrollOffset()` and `SetScrollOffset()` but doesn't expose the total content size. A `RowCount()` method would allow callers to implement custom paginators or accessibility overlays without needing the provider reference.

---

### 4.4 Consider making `Container.Layout()` return an error

Layout can fail if a widget receives dimensions of 0 (e.g. a flex child with zero remaining space). Currently negative or zero dimensions are passed silently to children, which may render garbage or panic in edge cases (e.g. `Grow.Hint()` panics if `g.child` is nil).

---

### 4.5 `Styles()` method returns non-deterministic order

**File:** `component.go:429`

`slices.Collect(maps.Keys(c.styles))` iterates a map, so the order of returned selectors is random. While this is currently only used for debugging, callers relying on stable output (tests, introspection) will be surprised.

---

### 4.6 `Update()` helper is loosely typed

**File:** `helper.go:126`

`Update(container, id, value any)` silently does nothing if the widget type or value type doesn't match. There is no way for the caller to know whether the update succeeded. Return a `bool` or an error to surface failures.

---

### 4.7 `Grow.Hint()` panics if child is nil

**File:** `grow.go:43`

```go
func (g *Grow) Hint() (int, int) {
    w, h := g.child.Hint()  // panics if g.child is nil
```

A nil child is a valid intermediate state (e.g. after `SetParent(nil)` is called on the previous child). Guard with `if g.child == nil { return 0, 0 }`.

---

### 4.8 `Progress` animation bar characters in `NewTheme()` use ASCII defaults

**File:** `theme.go:113-131`

The default progress bar characters (`#` and `.`) are plain ASCII. The rest of the framework uses Unicode box-drawing characters extensively. Consistent Unicode defaults (e.g. `█`, `░`) would look more polished out of the box and align with the scrollbar rendering.

---

## 5. Positive Observations

- **Gap buffer with KMP search** is an excellent implementation choice for a text editor widget.
- **Selector specificity cascade** in `theme.go` is well-thought-out and follows CSS conventions closely.
- **Layer-based popup system** is simple and handles z-ordering correctly.
- **Dirty-flag + channel-based rendering** (`redraw` + `refresh` channels) is a clean separation between single-widget and full-screen invalidation.
- **Generic `FindAll[T]`** is idiomatic Go 1.18+ usage.
- **The builder pattern** is ergonomic and makes UI construction code very readable.
- **Comprehensive keyboard navigation** in `Table` and `List` matches professional TUI application expectations.

---

## Priority Summary

| # | Severity | Location | Issue | Status |
|---|----------|----------|-------|--------|
| 1.1 | **High** | `input.go` | Unicode bug in `Input` — byte vs. rune length | ✅ Fixed (via 2.2) |
| 1.3 | **High** | `table.go:64` | `tableWidth` accumulates on repeated `Set()` calls | ✅ Fixed |
| 1.4 | **High** | `animation.go` | Data race on `ticker` field | ✅ Fixed |
| 1.2 | **Medium** | `table.go:90` | Off-by-one separator in `Hint()` | ✅ Fixed |
| 1.5 | **Medium** | `builder.go:58` | `Build()` always enables debug mode | ❌ Open |
| 2.1 | **Medium** | `builder.go:202` | `Apply()` panics on unknown widget types | ❌ Open |
| 2.2 | **Medium** | `input.go` | `Input` doesn't use `GapBuffer` | ✅ Fixed |
| 2.5 | **Medium** | `builder.go:471` | Duplicate spacer IDs | ❌ Open |
| 1.6 | **Low** | `helper.go:86` | Wrong package prefix in `WidgetType()` | ✅ Fixed |
| 1.7 | **Low** | `table.go:12` | Unused `column` field | ❌ Open |
| 1.8 | **Low** | `ui.go:144` | Parent set after log widget lookup | ✅ Fixed |
| 2.4 | **Low** | `component.go:416` | Side-effecting `Style()` getter | ❌ Open |
| 2.6 | **Low** | `theme.go:146` | `WithParent` return value discarded | ❌ Open |
| 3.2 | **Low** | `gap-buffer.go:172` | Goroutine per iteration in `Runes()` | ❌ Open |
| 3.3 | **Low** | `component.go:399` | Regex in hot render path | ❌ Open |
| 4.7 | **Low** | `grow.go:43` | Nil child panic in `Grow.Hint()` | ❌ Open |

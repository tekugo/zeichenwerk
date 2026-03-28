# Zeichenwerk — Project Review

**Date:** 2026-03-28
**Scope:** Design, interface, usability, extensibility, documentation, tests

---

## Executive Summary

Zeichenwerk is a well-architected, idiomatic Go TUI library with solid foundations. The core
abstractions (Widget interface, Component base, Builder, Theme) are clean and coherent. The
data-structure layer (GapBuffer, Stack) is particularly strong. The main areas requiring
attention are: test coverage for UI widgets, a few interface usability rough edges, two
open layout TODOs in Grid, and a CGo dependency leaking into the library module.

---

## 1. Architecture & Design

### Strengths

- **Layered architecture** is clear: data structures → styling → rendering → widgets →
  containers → UI/event loop. Each layer has a single responsibility.
- **Composition over inheritance** is consistently applied. Widgets embed `Component`
  rather than extending a class hierarchy, which keeps the Go idiomatic.
- **CSS-like theme system** with specificity-ordered resolution
  (`type → class → state → id → combined`) is expressive and familiar to web developers.
- **Builder pattern** provides a fluent, readable declarative API that mirrors the visual
  hierarchy of the resulting layout.
- **Screen interface** decouples the rendering pipeline from tcell, enabling testing and
  future backend swaps.

### Issues

#### 1.1 `Widget` interface is too wide

`Widget` has ~22 methods, including several that are infrastructure concerns
(`Log`, `Apply`, `SetParent`, `SetBounds`). In Go, small interfaces are preferred.
Consider splitting into focused interfaces that containers and the event loop use
internally, while the public API exposes only what application code needs.

```go
// Suggestion: smaller public surface
type Widget interface {
    ID() string
    Bounds() (int, int, int, int)
    On(event string, handler Handler)
    Dispatch(widget Widget, event string, data ...any) bool
    Flag(name string) bool
    SetFlag(name string, value bool)
    // ...
}
// Infrastructure-only (not exported or in a separate interface)
type renderable interface { Render(r *Renderer) }
type themeable  interface { Apply(theme *Theme) }
```

#### 1.2 `Update()` helper is a type-switch anti-pattern

`helper.go:Update()` performs a type switch over concrete widget types to call
`SetItems`, `SetText`, `Set`, etc. This is a smell: it means the polymorphism is
incomplete. A `SetValue(any)` method on an opt-in interface (or individual typed
setter helpers) would be more extensible and avoid breaking whenever a new widget
is added.

#### 1.3 Widget flags use untyped magic strings

Flags like `"focusable"`, `"focused"`, `"hidden"`, `"pressed"`, `"checked"`,
`"masked"`, `"readonly"`, `"hovered"` are raw strings scattered across the codebase.
A single typo silently produces incorrect behaviour. Define exported constants:

```go
const (
    FlagFocusable = "focusable"
    FlagFocused   = "focused"
    FlagHidden    = "hidden"
    // ...
)
```

#### 1.4 `go.mod` specifies a non-existent Go version

`go 1.25.5` does not exist. The latest stable release as of this review is Go 1.24.
This should be corrected to the actual minimum required version.

---

## 2. Interface & Usability

### Strengths

- `NewBuilder(theme).Flex(...).Grid(...).End().Run()` is ergonomic and readable.
- `With(func(*Builder))` enables composable, reusable UI fragments without breaking the
  fluent chain.
- `Find`, `FindAll[T]`, `FindAt`, `Traverse` give flexible tree traversal.
- `OnKey` and `OnMouse` typed helpers reduce boilerplate over raw `On("key", ...)`.

### Issues

#### 2.1 `Dispatch` first parameter is confusing

`Dispatch(widget Widget, event string, data ...any)` is a method on `Widget`, but
its first argument is also a `Widget`. From the implementation, this parameter is the
"source" widget (i.e. where the event originated). The name `widget` conflicts with the
receiver and the concept is not obvious from the signature alone. Rename to `source` and
document the distinction clearly:

```go
Dispatch(source Widget, event string, data ...any) bool
```

#### 2.2 `NewSelect` uses awkward alternating string pairs

```go
NewSelect("sel", "", "val1", "Label 1", "val2", "Label 2")
```

Odd-length args silently drop the last item. A slice of a small struct is safer and
more readable:

```go
type Option struct { Value, Label string }
NewSelect("sel", "", Option{"val1", "Label 1"}, Option{"val2", "Label 2"})
```

(The same issue applies to `NewList` if it follows the same convention.)

#### 2.3 `Select` is not a real dropdown

The `Select` widget cycles through options in-place with arrow keys; it does not open a
popup list. This is a significant usability gap — users expect a dropdown to present
all options at once. The TODO in `select.go:54` acknowledges the width issue, but the
lack of popup rendering is a deeper design limitation. A layer-based popup (like
`Dialog`) could be used.

#### 2.4 `OnMouse` registers the wrong event

`helper.go:OnMouse` registers for event name `"mouse"`, but `doc/EVENTS.md` states
the framework dispatches `"hover"` instead of `"mouse"` for mouse movement. This
makes `OnMouse` a no-op in practice for movement events. Either the helper should
register `"hover"`, or the framework should be made consistent.

#### 2.5 Event data requires type assertions

All event data flows as `...any`, so handlers must do:

```go
text := data[0].(string)  // panic if wrong type
```

For the documented events with fixed data types, typed handler wrappers (like `OnKey`)
would improve safety and IDE discoverability:

```go
OnChange(widget, func(w Widget, text string) bool)
OnActivate(widget, func(w Widget, index int) bool)
```

#### 2.6 Stale reference in `widget.go` comment

`widget.go:7` reads: *"All widgets share common functionality through BaseWidget"*.
The actual type is `Component`. This stale comment will mislead readers searching
for `BaseWidget`.

---

## 3. Extensibility

### Strengths

- `TableProvider` interface cleanly decouples data from display.
- `Screen` interface abstracts tcell, so alternative backends are possible.
- `Custom` base type provides a starting point for user-defined widgets.
- Builder's `Add(widget)` escape hatch allows adding arbitrary widgets.

### Issues

#### 3.1 CGo dependency in the library module

`go.mod` lists `github.com/mattn/go-sqlite3` as a direct dependency. This is a CGo
library used only in `cmd/` tools. Embedding it in the library module forces **all**
library consumers to have a C toolchain. Either move it to its own `cmd/` module, or
use a pure-Go SQLite driver.

Similarly, `github.com/mbndr/figlet4go` and `golang.org/x/tools` are cmd-only
dependencies. Consider a `cmd/go.mod` workspace or separate module for the demo tools.

#### 3.2 No extension points for focus traversal

Focus order is determined by the tree traversal order. There is no mechanism for an
application to customise focus order (e.g. skip a widget, wrap focus to a specific
widget, or implement roving tabindex). Adding a `FocusOrder()` or `NextFocus()`
override point would improve extensibility.

#### 3.3 `Apply(theme)` is easy to implement incorrectly

Each widget must override `Apply` and call `theme.Apply` with the correct selector and
state names. If a new widget forgets a state (e.g. `"disabled"`) its styling will be
silently wrong. Consider a registration mechanism or a builder helper so the theme
system discovers which states each widget supports.

---

## 4. Documentation

### Strengths

- `AGENTS.md` provides clear, normative guidelines (MUST/SHOULD/MAY) for contributors.
- `doc/EVENTS.md` is comprehensive and includes per-widget event data types.
- `doc/flags.md` covers all standard flags with descriptions.
- `doc/reference.md` is a thorough 775-line API reference.
- Inline comments consistently explain "why" rather than "what", as required.

### Issues

#### 4.1 README title inconsistency

`README.md` heading is `# zeichenwerk/next`. If this has graduated from a "next"
branch, the title should be updated to simply `# zeichenwerk`.

#### 4.2 `AGENTS.md` references deleted `archive/` directory

The project structure in `AGENTS.md` still lists `+- archive/ # Old version`, but git
history shows it was deleted. Remove the stale entry.

#### 4.3 Several widgets not covered in `doc/reference.md`

The following widgets have no entry or only passing mention in the reference:
`Canvas`, `Dialog`, `Digits`, `FormGroup`, `Grow`, `Inspector`, `Rule`, `Spinner`,
`Switcher`, `Tabs`, `Viewport`. Add sections documenting their constructor,
key methods, and events.

#### 4.4 No usage examples for the Builder

`doc.go` is documented as the home for examples, but contains no `Example*` functions.
Adding even two or three `Example` functions would appear in `go doc` output and help
new users onboard faster.

#### 4.5 `AGENTS.md` sentinel-error requirement is unmet in library code

`AGENTS.md` mandates: *"MUST define sentinel errors for common error cases"*, but the
library's `.go` files (outside `cmd/`) contain no `var Err... = errors.New(...)`.
`NewUI` returns an error from tcell but there are no sentinel errors consumers could
check with `errors.Is`. Define at minimum `ErrScreenInit` or similar.

---

## 5. Tests

### Strengths

- `gap-buffer_test.go` is exemplary: 18+ test functions, table-driven cases, Unicode
  coverage, panic recovery tests, and 5 benchmarks. This should be the model for all
  future test files.
- `stack_test.go` has 20+ tests and benchmarks covering all edge cases.
- `insets_test.go` and `style_test.go` cover the styling primitives well.

### Issues

#### 5.1 No widget-level tests

None of the 20+ widgets (Button, Input, Checkbox, Select, List, Table, Editor, Text,
Styled, Progress, Spinner, Canvas, etc.) have test files. Widget rendering and event
handling are the most user-visible parts of the library and are completely untested.

Recommended minimum per widget:
- `Render` does not panic for empty/zero-size bounds.
- Key event handlers fire the correct events with correct data.
- State flags (`focused`, `disabled`, `hidden`) affect rendering correctly.

#### 5.2 No container layout tests

`Flex`, `Grid`, `Box`, `Switcher`, `Viewport`, `Form` have no tests. Layout bugs
(off-by-one in column widths, incorrect nesting) are the hardest to debug visually.

At minimum, test that:
- `Layout()` assigns non-overlapping, non-negative bounds to children.
- Fractional sizing adds up to the available space.
- The two open Grid TODOs produce correct results once resolved.

#### 5.3 No Builder tests

The Builder is the primary user-facing API. Test that:
- Constructed widget trees have the expected parent-child relationships.
- `Find()` locates widgets created with the builder.
- `Build()` vs `Run()` produce the same UI instance.

#### 5.4 No Theme/style resolution tests

The cascading specificity logic (`type → class → state → id → combined`) is complex
and has no dedicated test coverage. A unit test that verifies resolution order and
inheritance would prevent regressions.

#### 5.5 `canvas_test.go` appears minimal

Check that the Canvas tests exercise actual drawing operations and not just
construction.

---

## 6. Known Technical Debt

| Location | Issue | Severity |
|---|---|---|
| `grid.go:283`, `grid.go:333` | `TODO: Distribute remaining space evenly` — last fractional column/row absorbs all rounding error | Medium |
| `select.go:54` | `TODO: Get real dropdown width?` — width estimate is wrong because renderer is unavailable in `Hint()` | Low |
| `gap-buffer.go` | `Move()` panic message is in German (`"Cursor außerhalb des gültigen Bereichs"`) — should be English | Low |
| `helper.go:Update()` | Type-switch dispatch — breaks if new widgets are added without updating the switch | Low |
| `go.mod` | `go 1.25.5` version does not exist; CGo `go-sqlite3` included at library level | High |

---

## 7. Prioritised Recommendations

**High priority (correctness / library health):**

1. Fix `go.mod` Go version and move `go-sqlite3`, `figlet4go`, `golang.org/x/tools` to
   a separate `cmd/` module or workspace to eliminate the CGo requirement for library
   users.
2. Resolve the two `grid.go` TODOs for even fractional space distribution.
3. Fix `OnMouse` to register `"hover"` or align the framework event name with the
   helper.

**Medium priority (usability):**

4. Define flag name constants to replace magic strings.
5. Replace alternating-string `NewSelect` / `NewList` constructor with typed
   `Option{Value, Label}` slices.
6. Add `Example` functions to `doc.go` and expand `doc/reference.md` to cover all
   widgets.
7. Add sentinel errors for `NewUI` and other failure paths.

**Medium priority (test coverage):**

8. Add render/event tests for at least the five most-used widgets: Button, Input,
   Checkbox, List, Table.
9. Add layout tests for Flex and Grid covering fractional sizing and nesting.
10. Add a Theme specificity resolution test.

**Low priority (design / ergonomics):**

11. Narrow the `Widget` interface; move infrastructure methods to internal interfaces.
12. Add typed event handler helpers (`OnChange`, `OnActivate`, `OnSelect`).
13. Fix stale `BaseWidget` comment in `widget.go`.
14. Remove stale `archive/` entry from `AGENTS.md` and update README title.
15. Translate the German panic message in `gap-buffer.go` to English.

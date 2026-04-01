# Zeichenwerk — Project Review (Open Issues)

**Updated:** 2026-04-01
**Scope:** Design, interface, usability, extensibility, documentation, tests

This document contains only unresolved issues. Resolved items have been removed.

---

## 1. Bugs

#### 1.7 `builder.go` unrecoverable panic in `buildGroup` (`builder.go:223`)

```go
panic("buildGroup called outside a Form context")
```

User-facing method; should return an error or at minimum recover gracefully rather
than crashing the process.


---

## 2. Interface & Usability


---

## 3. Extensibility

#### 3.1 CGo dependencies in the library module

`go.mod` lists `github.com/mattn/go-sqlite3` (CGo), `github.com/mbndr/figlet4go`, and
`golang.org/x/tools` as direct dependencies. All three are used only in `cmd/` tools.
This forces library consumers to have a C toolchain. Move them to a separate `cmd/`
module or workspace.

#### 3.2 No extension points for focus traversal

`UI.SetFocus` (`ui.go:453`) does a flat `Traverse` of the active layer, collecting
every widget that has `FlagFocusable` set and `FlagHidden` unset, in tree order.
There is no mechanism to skip a widget, reorder traversal, or confine Tab within a
group. Three targeted options are proposed below; they can be adopted independently.

**Option B — `FocusOrder []string` on containers (explicit ordering)**

Add an optional `SetFocusOrder(ids ...string)` method to `Component` (or to
specific containers). When non-empty the slice lists child IDs in the desired
tab order; unmentioned IDs are appended after in tree order.

```go
flex.SetFocusOrder("username", "password", "submit")
```

`SetFocus` walks each container: if it implements `FocusOrder() []string`, the
returned IDs replace that subtree's natural traversal order. No new interface
type needed; the method can live on `Component` with a nil/empty default.

**Option C — `FocusScope` interface (roving tabindex / trapped focus)**

Define a small interface that containers may implement to intercept traversal:

```go
type FocusScope interface {
    // NextFocus returns the widget that should receive focus after current
    // within this scope. Returning nil falls through to normal traversal.
    NextFocus(current Widget, forward bool) Widget
}
```

`SetFocus` checks each container as the traversal enters it; if it implements
`FocusScope`, control is delegated. This enables dialog/modal focus trapping
(always return a widget within the scope), roving tabindex in toolbars, and
ordered groups — without changing `Component` or any existing widget.

**Recommendation**

Defer B and C until a concrete use-case arises — premature focus API is hard to
remove once published.

#### 3.3 `Apply(theme)` is easy to implement incorrectly

Every widget overrides `Apply` and hand-writes `theme.Apply(w, selector, states...)` calls
with plain string literals for both the selector and the state list. Two failure modes:

- **Selector typo** — the `"tabs/hightlight"` bug (fixed in 1.2) is the canonical example.
- **Forgotten state** — omitting `"disabled"` from a focusable widget means the widget
  never renders its disabled style; the compiler cannot catch this.

There are 33 `Apply` overrides across the codebase. Two approaches:

**Option A — State constants (low effort, prevents typos)**

Promote the state strings to typed constants alongside the existing `Flag` constants:

```go
const (
    StateChecked  = "checked"
    StateDisabled = "disabled"
    StateFocused  = "focused"
    StateHovered  = "hovered"
    StatePressed  = "pressed"
)
```

Each `Apply` override then becomes:

```go
func (b *Button) Apply(theme *Theme) {
    theme.Apply(b, b.Selector("button"), StateDisabled, StateFocused, StateHovered, StatePressed)
}
```

Typos in state names become compile errors. Forgotten states are still possible but
the `StateXxx` names serve as a checklist. Low churn — purely mechanical substitution.

**Option B — `StyleSpec` registration (high effort, eliminates the entire class)**

Add a slice of specs to `Component`. Widgets register their selectors and states once
at construction time; `Component.Apply` iterates over them automatically, removing the
need for per-widget `Apply` overrides entirely.

```go
type StyleSpec struct {
    Selector string
    States   []string
}

// In NewButton:
b.AddStyleSpec(StyleSpec{"button", []string{StateDisabled, StateFocused, StateHovered, StatePressed}})

// Component.Apply becomes the only implementation needed:
func (c *Component) Apply(theme *Theme) {
    for _, spec := range c.styleSpecs {
        theme.Apply(c, c.Selector(spec.Selector), spec.States...)
    }
}
```

Widgets that need multiple sub-selectors (e.g. `list/highlight`) add a second spec.
This eliminates all 32 non-base `Apply` overrides and makes the set of styled parts
introspectable at runtime (useful for the Inspector).

**Recommendation**

Option A is a safe, low-risk improvement that can be done incrementally.
Option B is a larger refactor but produces a strictly better design — worth doing
when the widget set is stable enough to absorb the churn.

---

## 4. Documentation

#### 4.4 No `Example` functions in `doc.go`

`doc.go` contains no `func Example*` functions. Adding two or three runnable examples
would appear in `go doc` output and help new users onboard faster.


---

## 5. Tests

#### 5.1 Minimal widget-level tests

`component_test.go` covers `Component` (6 tests). `canvas_test.go` has a single
construction test. All other widgets — Button, Input, Checkbox, Select, List, Table,
Editor, Text, Styled, Progress, Spinner — have no tests.

Recommended minimum per widget:
- `Render` does not panic for empty/zero-size bounds.
- Key event handlers fire the correct events with correct data.
- State flags (`focused`, `disabled`, `hidden`) affect rendering correctly.

#### 5.2 Incomplete container layout tests

`flex_test.go` (21 tests) and `grid_test.go` (26 tests) are now present. Box,
Switcher, Viewport, and Form still have no layout tests.

#### 5.3 No Builder tests

The Builder is the primary user-facing API and is untested. Test that:
- Constructed widget trees have the expected parent-child relationships.
- `Find()` locates widgets created with the builder.
- Nested `With()` calls produce the correct hierarchy.

#### 5.4 No `Component.Style()` specificity resolution tests

`style_test.go` tests the `Style` struct's properties. The cascading fallback logic
in `component.go:Style()` — `"part:state"` → `"part"` → `":state"` → `""` — has no
dedicated test coverage and could silently regress.

#### 5.5 `canvas_test.go` is minimal

A single `TestCanvas` construction test. Drawing operations, cursor movement,
page switching, and event dispatch are not covered.

---

## 6. Remaining Technical Debt

| Location | Issue | Severity |
|---|---|---|
| `select.go:54` | `TODO: Get real dropdown width?` — width estimate is wrong because renderer is unavailable in `Hint()` | Low |
| `gap-buffer.go:96` | Panic message is in German (`"Cursor außerhalb des gültigen Bereichs"`) — should be English | Low |
| `go.mod` | CGo `go-sqlite3`, `figlet4go`, and `golang.org/x/tools` included at library level | High |
| `go.mod` | `go 1.26.1` declared — this is a future (unreleased) version; should be `go 1.23` or whatever is actually installed | Medium |
| `compose/compose.go` | Zero tests — all compose constructors, styling helpers, and event options are untested | High |

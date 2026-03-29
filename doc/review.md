# Zeichenwerk — Project Review (Open Issues)

**Updated:** 2026-03-29
**Scope:** Design, interface, usability, extensibility, documentation, tests

This document contains only unresolved issues. Resolved items have been removed.

---

## 2. Interface & Usability

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

#### 2.5 Typed event helpers incomplete

`OnKey`, `OnMouse`, `OnActivate`, `OnChange`, and `OnSelect` exist. Typed helpers for
`EvtClick`, `EvtEnter`, `EvtShow`, and `EvtHide` are still missing, requiring callers to
write raw `On(EvtClick, func(_ Widget, _ Event, data ...any) bool { ... })` with manual
type assertions.

---

## 3. Extensibility

#### 3.1 CGo dependencies in the library module

`go.mod` lists `github.com/mattn/go-sqlite3` (CGo), `github.com/mbndr/figlet4go`, and
`golang.org/x/tools` as direct dependencies. All three are used only in `cmd/` tools.
This forces library consumers to have a C toolchain. Move them to a separate `cmd/`
module or workspace.

#### 3.2 No extension points for focus traversal

Focus order is determined by tree traversal order. There is no mechanism to customise
it (skip a widget, force focus to a specific target, implement roving tabindex). A
`FocusOrder()` or `NextFocus()` override point would improve extensibility.

#### 3.3 `Apply(theme)` is easy to implement incorrectly

Each widget must override `Apply` and call `theme.Apply` with the correct selector and
state names. A forgotten state (e.g. `"disabled"`) produces silently wrong styling.
Consider a registration mechanism or a builder helper so the theme system discovers
which states each widget supports.

---

## 4. Documentation

#### 4.3 `Inspector` not covered in `doc/reference/`

All other widgets (Canvas, Dialog, Digits, FormGroup, Grow, Rule, Spinner, Switcher,
Tabs, Viewport) now have reference pages. `Inspector` is the only one still missing.

#### 4.4 No `Example` functions in `doc.go`

`doc.go` contains no `func Example*` functions. Adding two or three runnable examples
would appear in `go doc` output and help new users onboard faster.

#### 4.5 Sentinel error coverage is minimal

`errors.go` defines `ErrChildIsNil`. `AGENTS.md` mandates sentinel errors for all
common error cases, but failure paths in `NewUI`, layout, and rendering have none.
Add at minimum `ErrScreenInit` (wrapping tcell init failures) so callers can use
`errors.Is`.

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

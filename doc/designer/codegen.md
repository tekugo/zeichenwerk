# Designer Codegen

The designer can dump the live widget tree as a Go expression. Two
entry points cover the common cases:

```go
// Emit just the chain expression — useful when the caller already
// has a builder set up and wants to splice the tree into existing
// scaffolding.
func (d *Designer) GenerateFragment(mode string, out io.Writer) error

// Emit a complete, gofmt-clean Go file with the package declaration,
// the function wrapper, and the chain expression as the function
// body.
func (d *Designer) GenerateFile(mode, out io.Writer, pkg, funcName string) error
```

`mode` is one of `inspector.ModeBuilder` or `inspector.ModeCompose`.
Today only `ModeBuilder` is implemented; `ModeCompose` is reserved
for a future compose-style backend.

## Output shape

Builder mode produces a single chained expression:

```go
NewBuilder(theme).
    VFlex("ui-root", Stretch, 0).
    Class("highlight").
    HFlex("header", Center, 2).
    Static("title", "Inspector PoC").Padding(0, 1).
    Input("search", "", "", "type to filter…").
    End(). // HFlex#header
    Grid("g1", 2, 2, false).
    Cell(0, 0, 1, 1).
    Static("s1", "Hello").
    Cell(1, 0, 1, 1).
    Class("highlight").
    Static("s2", "World").
    End(). // Grid#g1
    End(). // VFlex#ui-root
    Build()
```

Containers emit only their constructor, prefix, and chain — children
and the closing `End()` are written by the codegen walker. Each
container's `End()` is followed by `// Kind#id` so a reader can
match the close to its open. Widgets without an id render as
`// Kind`.

## Chain-element convention

Every emitted call ends with `".\n"`. The trailing dot is the chain
separator; the newline lets `gofmt` decide whether to keep the chain
on one line or break it across several. The very last call in the
expression has its trailing `"."` stripped by the walker before the
formatter runs. Forms therefore never have to worry about whether
they are emitting the last element — they always emit `".\n"` and
let the framework strip the dot.

This is why a typical form body looks like:

```go
func (f *XForm) Emit(w io.Writer, mode string) error {
    return f.EmitFrame(w, mode, func() error {
        _, err := fmt.Fprintf(w, "X(%q, %q).\n", f.ID, f.Title)
        return err
    })
}
```

`EmitFrame` writes the body, then continues the chain with the
standard `Hint(…)`, `Flag(…)`, and styling calls — each of them also
ending with `".\n"`. The walker handles the rest.

## Class prefix

A widget with a non-empty `Class` field emits `Class("…").\n` before
its constructor:

```go
Class("highlight").
Static("s2", "World").
```

`Class` is a chained Builder method that sets the class register for
the *next* call. The form does not include the class as an argument
to the constructor itself; instead, `ComponentForm.EmitClassPrefix`
emits a separate chain element. The Builder consumes the class and
applies it to the widget it constructs next.

## Hint, flags, and style chain

After the constructor (and any kind-specific tail), `EmitFrame`
appends the standard chain:

```go
.Hint(0, -1).
.Flag(FlagSkip, true).
.Flag(FlagHidden, true).
.Flag(FlagDisabled, true).
.Background("$bg-2").
.Foreground("$fg").
.Border("thin").
.Padding(0, 1).
.Margin(1, 0).
.Font("bold")
```

Only entries that differ from the widget's default state are
emitted. Hint defaults to `(0, 0)` (no override), the three flags
default to `false`, and styling chain elements only fire for
non-themed (non-fixed) styles. A widget that inherits everything
from the theme emits nothing after the constructor — exactly what
you would write by hand.

## Styling: themed vs explicit

Styles registered through a `*Theme` are marked **fixed** at
registration time. Codegen treats fixed styles as "this came from
the theme; the user did not edit it" and skips them — emitting them
in source would override theme changes at the call site.

Edits made through the Style tab produce a non-fixed child style
(via `Style.Modifiable`) that cascades from the original. Codegen
sees the non-fixed status and emits the styling chain
(`Background`, `Foreground`, `Border`, `Padding`, `Margin`, `Font`)
for each property the user actually changed.

The result: a freshly generated file matches the theme it was
designed under. Switching themes after generation works correctly,
because the source file does not contain the theme's defaults.

## TODO comments

When a form has a non-trivial property that the Builder does not
expose as a chained setter, the form emits a `// TODO` comment so
the user can fill in the call manually after generation. The
canonical examples:

```go
// TODO: SetSpeed(2) — no Builder setter
// TODO: Set("hello world") — no Builder setter
// TODO: SetItems([]string{"one", "two", "three"}) — no Builder setter
```

The line is short, names the proposed setter, includes the value,
and adds a one-phrase reason ("no Builder setter") so a reader sees
the gap in the chain rather than a mysterious blank line.

A generated file may contain several TODOs. They are not blockers —
the file compiles and runs without them — but they flag deferred
configuration the user needs to apply before the output matches the
designed widget exactly.

## Container-only state

Some widgets carry application-level data that codegen cannot
serialise:

- `Table` needs a `TableProvider` (interface)
- `Tiles` and `Deck` need an `ItemRender` (function)
- `TreeWidgets` needs a root widget reference

For these, the form emits a placeholder identifier:

```go
Table("rows", tableProvider /* TODO */, false).
Tiles("grid", tileRender /* TODO */, 12, 4).
TreeWidgets("inspector-tree", rootWidget /* TODO */).
```

The user replaces the identifier with the actual variable when
integrating the generated code.

## ModeBuilder vs ModeCompose

`ModeBuilder = "builder"` produces the chain expression shown above.

`ModeCompose = "compose"` is reserved for a future compose-style
backend that emits explicit constructor calls instead of a fluent
chain:

```go
// future shape — not yet implemented
flex := NewFlex("ui-root", "", Stretch, 0)
flex.SetFlag(FlagVertical, true)
header := NewFlex("header", "", Center, 2)
…
flex.Add(header)
```

Adding compose support is mostly a matter of writing a parallel
`EmitFrame` helper for the compose shape. The form interface is
deliberately mode-agnostic — `Emit(w, mode)` lets a single form
support multiple backends without changing its signature.

## Headless usage

Codegen does not require a UI. The Designer is a pure model — a
host can register kinds, build a tree programmatically, and call
`GenerateFile` without ever opening the popup. This is how
`cmd/inspector-poc` works: it builds a tree by hand, registers a
small set of kinds, and writes the generated chain to stdout.

The same machinery powers `Save` and `Generate` in the
`designer-poc` popup — the buttons are just one path into the same
`GenerateFile` call.

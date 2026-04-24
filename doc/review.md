# Project Review

Date: 2026-04-24
Reviewer: Claude (Opus 4.7, 1M context)
Scope: full project sweep — structure, correctness, error handling,
security. The project is ~41k lines of Go across `core/`, `renderer/`,
`themes/`, `values/`, `widgets/`, `compose/`, and the root package.

Priorities:

- **Critical** — actual bug or design flaw that can produce wrong behaviour
  or a crash. Fix soon.
- **Important** — design rough edge, easy pitfall, or inconsistency.
  Worth addressing but not urgent.
- **Minor** — style, future-proofing, or an opinion.

Findings are grouped by theme, then prioritised within each theme.

---

## 1. Correctness bugs

### Critical

**1.1  `Renderer.ScrollbarV` can divide by zero** —
`renderer/renderer.go:182`

```go
pos := min(max(offset*(height-thumb)/(total-height), 0), height-thumb)
```

When `total == height` (content exactly fills the visible area), the
denominator `total - height` is 0 and the call panics. The sibling
function `ScrollbarH` at `renderer.go:219-224` guards this case
explicitly:

```go
if total > width {
    pos = min(max(offset*(width-thumb)/(total-width), 0), width-thumb)
} else {
    pos = 0
}
```

The asymmetry is almost certainly an oversight — both should guard the
same way. One-line fix.

**1.2  German-language panic message** —
`core/gap-buffer.go:197`

```go
panic("Cursor außerhalb des gültigen Bereichs")
```

Every other panic and log message in the codebase is English; this one
is German. Likely a leftover from early development. Replace with
`panic("GapBuffer.Move: position out of range")` or similar.

### Important

**1.3  `Find(...).(\*T)` panics silently if a widget is missing** —
`ui.go:763`, `file-chooser.go:81,128-131`, `inspector.go:76-77`,
`commands.go:148`.

Pattern:

```go
input := Find(dialog, "prompt-input").(*Input)
```

`Find` returns `nil` on a miss; a non-ok type assertion then panics.
These call sites are all internal (we built the dialog ourselves so the
lookup always succeeds today), but any ID-rename, widget-swap, or copy-
paste into user code turns them into silent panics. Either switch to
the comma-ok form and log a descriptive error, or introduce a
`MustFind[T Widget](c Container, id string) T` helper that centralises
the behaviour and the diagnostic.

**1.4  Integer overflow risk in scrollbar math** —
`renderer/renderer.go:181,216`

`height*height/total` and `width*width/total` overflow `int` on 32-bit
platforms for sizes above ~46k, and on 64-bit platforms for sizes
above ~3G. Not reachable in practice in a TUI (terminals are not that
big), but also not robust — switching to `int64` for the intermediate
product would cost nothing.

**1.5  `Layout` helper hides all but the last error** —
`core/helper.go:136-146`

```go
for _, child := range container.Children() {
    if inner, ok := child.(Container); ok {
        if err = inner.Layout(); err != nil {
            container.Log(inner, Error, "Layout failed", "error", err)
        }
    }
}
return err
```

`err` is overwritten every iteration, so the caller sees only the last
failing child (or nil if the last child succeeded after earlier
failures). The log has the full story but programmatic callers do not.
Either use `errors.Join` (Go 1.20+) or track a boolean "any failed" and
return a wrapped sentinel — documented behaviour is still surprising.

**1.6  `Input` / `Editor` accept unbounded paste** —
`widgets/input.go:127-130`, `widgets/editor.go` (no max enforced).

`Input.Set` honours `max` only when `max > 0`, and `max` defaults to 0.
A paste of 1 GB into an Input allocates 1 GB in the GapBuffer. Not
remotely exploitable — a TUI is local — but a malicious clipboard or
errant script can OOM the process. Consider a default soft cap (e.g.
64 KiB) with an opt-out for Editor-like widgets.

**1.7  `GapBuffer.Runes` can leak a goroutine** —
`core/gap-buffer.go:235-260`

A goroutine is spawned to feed the returned channel. If the caller
breaks out of the `for r := range runes` loop before consuming every
rune, the producer blocks on `ch <-` forever. Documented, but easy to
misuse. A context-cancellable variant, or a `Runes(start int) iter.Seq[rune]`
using the `iter` package (already used by `TimeSeries.All`), would
eliminate the hazard entirely.

### Minor

**1.8  `TimeSeries.Min` / `Max` panic on zero-sized series** —
`core/time-series.go:106-113,117-124`

`NewTimeSeries(..., size=0, ...)` produces a zero-length buffer; `Min`
and `Max` index `ts.buf[0]` unconditionally. Documented as invariant
violation, matches the project panic-for-invariants convention.

**1.9  `TODO` in `widgets/select.go:58`** — dropdown width hint is
approximate. Cosmetic.

**1.10  Unchecked `regexp.Compile` in `core/theme.go:31`** — the regex
literal is constant and cannot fail, but suppressing the error reads as
sloppy. Prefer `regexp.MustCompile`.

**1.11  Unchecked `os.Getwd` in `file-chooser.go:54`** and a few other
spots. The fallback path is safe but the `_, _ = ` pattern looks
deliberately lazy.

---

## 2. Package organization

### Critical

**2.1  Flag constants split arbitrarily between `core` and `widgets`** —
`core/flags.go:21` defines only `FlagHidden`; `widgets/flags.go` defines
the other fourteen (`FlagChecked`, `FlagDisabled`, `FlagFocusable`,
`FlagFocused`, `FlagGrid`, `FlagHorizontal`, `FlagHovered`,
`FlagMasked`, `FlagPressed`, `FlagReadonly`, `FlagRight`, `FlagSearch`,
`FlagSkip`, `FlagVertical`). `widgets/flags.go:26` openly acknowledges
the split with a comment.

`FlagHidden` is not more "fundamental" than `FlagFocusable` or
`FlagDisabled` — all three affect cross-cutting concerns (rendering,
input, layout). The split forces every widget constructor that wants
`FlagFocusable` to import both packages and makes the flag set hard to
enumerate.

Two options:

- Move every flag to `core/flags.go`. Keeps flags uniform. Drawback:
  `core` grows slightly and references implementation details
  (`FlagPressed` only makes sense with mouse interaction implemented by
  widgets).
- Move `FlagHidden` to `widgets/flags.go`. Keeps `core` minimal but
  breaks the existing core-level semantics — `core.FindAt` currently
  references `FlagHidden`.

First option is cleaner given the current usage.

### Important

**2.2  `widgets/` is a flat 103-file package with ~72 widget types** —
discoverability will degrade as the catalogue grows. Reasonable
splits: `widgets/input` (Input, Checkbox, Select, Combo, Typeahead),
`widgets/layout` (Flex, Grid, Box, Card, Tiles), `widgets/data`
(Sparkline, Heatmap, BarChart, Progress, TimeSeries bindings),
`widgets/text` (Text, Marquee, Typewriter, Editor), `widgets/tree`
(Tree, TreeFS, List, Table). Blocker: it would force a dot-import
rename across every `cmd/` example and would move `Component` out from
under every widget. Worth considering for a future major version, not
now.

### Minor

**2.3  `file-chooser.go`, `inspector.go`, `commands-panel.go` at
project root** — these are UI dialogs tied to the root `UI` type and
their current placement is defensible. An alternative `ui/dialogs/`
sub-directory (using the root package name) would reduce noise in the
project root but is not a correctness concern.

**2.4  File naming is mostly consistent kebab-case** — occasional
compound-noun cases (`file-chooser.go`, `commands-panel.go`) match the
convention. No changes needed.

**2.5  `compose/` and `themes/` have zero test coverage** — the
packages are small (1.3k and 1.4k LOC respectively) and largely data
definitions (themes) or facade code (compose), but a handful of
smoke-tests exercising common compose flows and theme Build calls
would catch regressions across the Theme.Add / Style.Fix surface.

---

## 3. Error handling

### Important

**3.1  `MessageCode` sentinel coverage is thin** — only three exist
(`ErrChildIsNil`, `ErrNoContainer`, `ErrScreenInit`). Recurring
failures that would benefit from sentinels include:

- widget lookup failure (`Find` returning nil),
- invalid selector (unparseable by `styleRegExp`),
- theme asset not found (border or string by name),
- I/O errors in `tree-fs` and `file-chooser` that currently surface
  as plain `fmt.Errorf` wrappers.

Sentinels let callers branch on failure class without parsing message
text. The current pattern of "log and continue" is fine for rendering-
path failures but deserves a second look for builder and theme
operations where recovery is possible.

**3.2  `Handler` has no error channel** — `Handler func(source, event,
data ...any) bool` only signals consumed / not-consumed. A handler
that hits a runtime error can only log-and-swallow. This is a
deliberate design choice (events are fire-and-forget) but an
application that wants to react to handler errors (for example to pop
up a dialog) has no hook. Worth mentioning so future design changes do
not break assumed behaviour.

### Minor

**3.3  Widget constructors panic instead of returning errors** —
`widgets/cell-buffer.go` panics if dimensions are < 1, etc. Matches
the "invariant violation" convention in `core/doc.go`. Acceptable.

**3.4  Error message capitalisation / punctuation** — mostly follows
Go idiom (lowercase, no trailing period). A handful of places in
`ui.go` and `file-chooser.go` use capitalised messages; worth a pass
for consistency.

---

## 4. Security

The attack surface is intentionally narrow (a local TUI, no network
listener, no deserialisation of untrusted data). There are no critical
security findings. Items below are worth noting but not exploitable.

### Important

**4.1  `FileChooser` permits traversal outside its initial root** —
`file-chooser.go:87-124`. Tilde expansion and `filepath.Clean` work
as expected, but there is no "confined to subtree" option. Callers
who expect sandboxed browsing must filter selections themselves. Add
a `Root string` / `Restrict bool` option on `FileChooser` if that use
case is real.

**4.2  `Input` / `Editor` unbounded memory** — see 1.6 above. Bumped
here because it is also the closest thing to a local DoS in the
framework.

### Minor

**4.3  Regex in `core/theme.go:31` is not a ReDoS risk** — the
character class `[0-9A-Za-z_\-]*` is simple and anchored; no nested
quantifiers. Safe even with 100 KB inputs.

**4.4  Terminal widget (`widgets/ansi.go`) interprets ANSI
intentionally** — escape-sequence injection is by design there.
Regular text widgets rasterise input as cells via tcell and are safe.

**4.5  Logs may contain filesystem paths** — unavoidable in a file
browser. No secret-bearing data appears to be logged.

**4.6  `Builder` / `Form` use `reflect`** — struct-pointer validation
only, no `unsafe.Pointer`, no type confusion.

**4.7  No `os/exec` usage anywhere in the framework packages** —
confirmed via search. The `cmd/` examples are unaffected because they
are user programs.

---

## 5. Suggested action list (short)

If only a few things get done, pick from the top:

1. Fix 1.1 (ScrollbarV division-by-zero). Trivial.
2. Fix 1.2 (German panic). Trivial.
3. Consolidate 2.1 (flag constants). One move, update imports.
4. Address 1.3 (unchecked Find+cast) with a `MustFind[T]` helper;
   audit call sites.
5. Decide on a default `Input` length cap (1.6 / 4.2).
6. Replace `GapBuffer.Runes` channel with an `iter.Seq[rune]` (1.7).
7. Broaden `MessageCode` sentinels (3.1).

Everything else is style, future-proofing, or opinion and can wait.

# Specification: Terminal Widget

**File:** `terminal.go`  
**Package:** `zeichenwerk`

---

## Purpose

`Terminal` is a zeichenwerk widget that renders arbitrary terminal output.
It holds two `CellBuffer`s (main and alternate screen), processes byte streams
via an embedded `AnsiParser`, and renders the active buffer to the zeichenwerk
screen via the standard `Renderer` pipeline.

It implements `io.Writer` so callers can pipe pty output directly into it
without an intermediate channel.

---

## Struct definitions

### termCursor

```go
type termCursor struct {
    x, y   int    // 0-based column and row
    fg     Color  // current foreground colour
    bg     Color  // current background colour
    ul     Color  // current underline colour
    attrs  uint32 // packed char word bits 21-31 (attrs + ul_style)
    wrap   bool   // pending wrap: next Print will move to next line first
    saved  *termCursor  // single DECSC/SCOSC save slot; nil if nothing saved
}
```

`wrap` is the deferred-wrap flag: when the cursor reaches the last column,
`wrap` is set to `true` rather than moving to the next line immediately. The
next `Print` call checks `wrap` first, advances the line, and then writes.
This matches VT100 behaviour where the cursor *stays* on the last column
visually until the next printable character arrives.

### Terminal

```go
type Terminal struct {
    Component
    main    *CellBuffer
    alt     *CellBuffer
    active  *CellBuffer   // pointer to main or alt
    parser  *AnsiParser
    handler *termHandler
    cur     termCursor
    scroll  struct {
        top int  // first row of scroll region (0-based, default 0)
        bot int  // last row of scroll region (0-based, default height-1)
    }
    autoWrap   bool  // mode ?7, default true
    showCursor bool  // mode ?25, default true
    mu         sync.Mutex
}
```

`mu` protects `Write` calls arriving from a goroutine other than the UI thread.
The `Render` method acquires `mu` before reading the active buffer.

### termHandler (private)

```go
type termHandler struct {
    t *Terminal
}
```

Implements `AnsiHandler`. All handler methods acquire no locks — they are
always called from within `Terminal.Write`, which holds `mu`.

---

## Constructor

```go
func NewTerminal(id, class string) *Terminal
```

1. Creates `main` and `alt` as `NewCellBuffer(80, 24)` (default VT100 size).
2. Sets `active = main`.
3. Initialises `cur` with `ColorDefault` colours, zero position, `autoWrap=true`,
   `showCursor=true`.
4. Sets scroll region to `{0, 23}`.
5. Creates `parser = NewAnsiParser(handler)`.
6. Sets `FlagFocusable = true`.

---

## Public API

### Write

```go
func (t *Terminal) Write(data []byte) (int, error)
```

1. Acquires `t.mu`.
2. Calls `t.parser.Feed(data)`.
3. Releases `t.mu`.
4. Calls `Redraw(t)` to schedule a repaint.
5. Returns `len(data), nil`.

### Apply

```go
func (t *Terminal) Apply(theme *Theme)
```

Applies `"terminal"` and `"terminal:focused"` styles. No sub-parts.

### Hint

```go
func (t *Terminal) Hint() (int, int)
```

Returns `(t.main.Width(), t.main.Height())`. Callers who want a flexible-size
terminal should set a custom hint via `SetHint`.

### Render

```go
func (t *Terminal) Render(r *Renderer)
```

See § Render below.

### Clear

```go
func (t *Terminal) Clear()
```

Acquires `mu`, clears `main` and `alt`, resets cursor to `{0,0}`, resets
scroll region, releases `mu`, calls `Redraw(t)`.

### Resize

```go
func (t *Terminal) Resize(w, h int)
```

Acquires `mu`, calls `Resize(w, h)` on both buffers, clamps cursor and scroll
region to new dimensions, releases `mu`.

Called automatically by `Render` when content dimensions change.

---

## Render method

```go
func (t *Terminal) Render(r *Renderer) {
    t.Component.Render(r)    // background fill, border

    x0, y0, cw, ch := t.Content()
    if cw <= 0 || ch <= 0 { return }

    t.mu.Lock()

    // Resize buffers if the widget dimensions changed.
    if t.active.Width() != cw || t.active.Height() != ch {
        t.main.Resize(cw, ch)
        t.alt.Resize(cw, ch)
        // Clamp cursor and scroll region.
        t.clampCursor()
        t.clampScroll()
    }

    // Render each cell.
    for y := 0; y < ch; y++ {
        for x := 0; x < cw; x++ {
            ch, fg, bg, ul, attrs := t.active.Get(x, y)

            // Apply reverse-video: swap fg and bg.
            if attrs&charReverse != 0 {
                fg, bg = bg, fg
            }

            fgStr := colorToHex(fg)
            bgStr := colorToHex(bg)
            font  := attrsToFont(attrs)

            r.Set(fgStr, bgStr, font)

            // Underline colour and style (requires Screen.SetUnderline).
            if attrs&charULMask != 0 {
                ulStyle := int((attrs & charULMask) >> charULShift)
                r.SetUnderline(ulStyle, colorToHex(ul))
            }

            glyph := string(ch)
            if ch == 0 || attrs&charInvis != 0 {
                glyph = " "
            }
            r.Put(x0+x, y0+y, glyph)
        }
    }

    t.mu.Unlock()
}
```

`r.SetUnderline` is a new method on `Renderer` (see § Screen extension).

---

## termHandler — AnsiHandler implementation

### Print

```go
func (h *termHandler) Print(r rune)
```

1. If `t.cur.wrap` is true: call `lineFeed()` then set `t.cur.x = 0`,
   `t.cur.wrap = false`.
2. Write the rune into the active buffer at `(t.cur.x, t.cur.y)` with current
   cursor colours and attrs.
3. Determine cell width: `rw = runewidth.RuneWidth(r)` (using
   `github.com/mattn/go-runewidth` — already in `go.mod`; verify before
   assuming). If `rw == 2`, set the wide flag and fill the next cell with a
   continuation placeholder (`rune=0, charWide=0`).
4. Advance `t.cur.x` by `rw`.
5. If `t.cur.x >= t.active.Width()`:
   - If `t.autoWrap`: set `t.cur.wrap = true`, set `t.cur.x = t.active.Width()-1`.
   - Else: clamp `t.cur.x = t.active.Width()-1`.

### Execute

```go
func (h *termHandler) Execute(code byte)
```

| Code | Action |
|------|--------|
| `0x07` | BEL: no-op |
| `0x08` | BS: `t.cur.x = max(0, t.cur.x-1)`; `t.cur.wrap = false` |
| `0x09` | HT: advance `x` to next multiple of 8, clamped to `width-1` |
| `0x0A`, `0x0B`, `0x0C` | LF/VT/FF: `lineFeed()` |
| `0x0D` | CR: `t.cur.x = 0`; `t.cur.wrap = false` |
| `0x7F` | DEL: no-op |

### CsiDispatch

Routes by `final` byte. Default parameter values are applied before routing
(missing params become the default listed in the ANSI parser spec § CSI table).

**Cursor movement** (`A`–`D`, `E`–`G`, `H`, `d`, `f`):
- Move cursor by n rows/columns.
- `H`/`f`: set absolute position (1-based → 0-based by subtracting 1).
- All movements clamp to `[0, width-1]` × `[scroll.top, scroll.bot]` for
  vertical, `[0, width-1]` for horizontal (except `H`/`f` which clamp to
  the full buffer).
- Clear `wrap` flag after any movement.

**Erase** (`J`, `K`, `X`, `P`):
- `J 0`: erase from cursor to end of screen.
- `J 1`: erase from start of screen to cursor.
- `J 2`/`J 3`: erase entire screen; `J 3` also resets scroll-back (no-op
  since no scroll-back is stored).
- `K 0`: erase cursor to end of line (`ClearLine(y, x, width)`).
- `K 1`: erase start of line to cursor (`ClearLine(y, 0, x+1)`).
- `K 2`: erase entire line (`ClearLine(y, 0, width)`).
- `X n`: erase n characters from cursor (erase, not delete — no shift).
- `P n`: delete n characters at cursor (shift remaining chars left, fill end
  with spaces).

Erased cells are filled with space (`rune=0`), `ColorDefault` colours, and
zero attrs — *except* the background colour, which is filled with `t.cur.bg`
(ANSI erase uses the current background).

**Insert/Delete lines** (`L`, `M`):
- `L n`: insert n blank lines at cursor row; lines below scroll.bot are lost.
- `M n`: delete n lines at cursor row; blank lines appear at scroll.bot.

**Scroll** (`S`, `T`):
- `S n`: scroll up n lines within scroll region.
- `T n`: scroll down n lines within scroll region.

**Mode** (`h`/`l` with inter `?`):
- `?7 h/l`: set/clear `autoWrap`.
- `?25 h/l`: set/clear `showCursor`.
- `?1049 h`: switch to alt screen, clear it, reset cursor.
- `?1049 l`: switch back to main screen, restore cursor.

**SGR** (`m`): delegated to `applySGR(params)`.

### OscDispatch

```go
func (h *termHandler) OscDispatch(cmd int, data string)
```

| cmd | Action |
|-----|--------|
| 0, 1, 2 | Window/icon title — store in `t.title string` field; emit no event |
| others | Ignore |

### EscDispatch

```go
func (h *termHandler) EscDispatch(inter, final byte)
```

| inter | final | Action |
|-------|-------|--------|
| 0 | `7` | DECSC: save cursor (`t.cur.saved = copy of t.cur`) |
| 0 | `8` | DECRC: restore cursor (`t.cur = *t.cur.saved` if saved != nil) |
| 0 | `M` | RI: reverse index (`reverseIndex()`) |
| 0 | `c` | RIS: hard reset |

---

## Internal helpers

### lineFeed

```go
func (t *Terminal) lineFeed()
```

1. If `t.cur.y < t.scroll.bot`: `t.cur.y++`.
2. Else (cursor is at bottom of scroll region): `scrollUp(1)`.
3. `t.cur.wrap = false`.

### reverseIndex

```go
func (t *Terminal) reverseIndex()
```

1. If `t.cur.y > t.scroll.top`: `t.cur.y--`.
2. Else: `scrollDown(1)`.

### scrollUp

```go
func (t *Terminal) scrollUp(n int)
```

Scrolls lines `scroll.top..scroll.bot` up by `n`. Lines that scroll off the
top are lost. New blank lines appear at the bottom filled with space +
`t.cur.bg`.

Implementation: for each of the `n` steps, copy rows
`scroll.top+1..scroll.bot` upward by one, then clear `scroll.bot`.

### scrollDown

```go
func (t *Terminal) scrollDown(n int)
```

Inverse of `scrollUp`: lines scroll.top..scroll.bot−n shift down; new blank
lines appear at scroll.top.

### applySGR

```go
func (t *Terminal) applySGR(params []int)
```

Processes the params slice left-to-right according to the SGR table in
`spec/ansi-parser.md`. Updates `t.cur.fg`, `t.cur.bg`, `t.cur.ul`,
and the attr bits in `t.cur.attrs`.

### clampCursor

```go
func (t *Terminal) clampCursor()
```

Ensures `t.cur.x` and `t.cur.y` are within `[0, width-1]` and `[0, height-1]`.

### clampScroll

```go
func (t *Terminal) clampScroll()
```

Ensures `t.scroll.top` and `t.scroll.bot` are within `[0, height-1]` and
`top < bot`.

---

## Screen interface extension

`Renderer` gains one new method:

```go
// SetUnderline sets the underline style and colour for subsequent Put calls.
// style: 0=none, 1=single, 2=double, 3=curly, 4=dotted, 5=dashed.
// color: empty string = terminal default.
func (r *Renderer) SetUnderline(style int, color string)
```

`Renderer.SetUnderline` calls `r.screen.SetUnderline(style, r.theme.Color(color))`.

### Screen interface

```go
type Screen interface {
    // existing methods …
    SetUnderline(style int, color string)
}
```

### TcellScreen.SetUnderline

```go
func (s *TcellScreen) SetUnderline(style int, color string) {
    us := tcell.UnderlineStyleNone
    switch style {
    case 1: us = tcell.UnderlineStyleSolid
    case 2: us = tcell.UnderlineStyleDouble
    case 3: us = tcell.UnderlineStyleCurly
    case 4: us = tcell.UnderlineStyleDotted
    case 5: us = tcell.UnderlineStyleDashed
    }
    s.style = s.style.UnderlineStyle(us)
    if color != "" {
        s.style = s.style.UnderlineColor(parseColor(color))
    }
}
```

`parseColor` is the existing color-string-to-tcell helper already used by
`TcellScreen.Set`.

### mock.go stub

```go
func (m *mockScreen) SetUnderline(style int, color string) {}
```

---

## Style keys

```
"terminal"           base style (fg/bg colours)
"terminal:focused"   when widget holds keyboard focus (border colour change)
```

All five existing themes add:

```go
NewStyle("terminal").WithColors("$fg0", "$bg0"),
NewStyle("terminal:focused").WithColors("$fg0", "$bg0"),
```

No border is set by default; callers add `.WithBorder(…)` via the builder
if they want one.

---

## Builder integration

**`builder.go`**:

```go
// Terminal adds a Terminal widget to the current container.
func (b *Builder) Terminal(id string) *Builder {
    w := NewTerminal(id, b.class)
    w.Apply(b.theme)
    b.add(w)
    b.current = w
    return b
}
```

No child widgets, so no matching `End()` call is needed (Terminal is a leaf).

**`compose/compose.go`**:

```go
func Terminal(id, class string, options ...Option) Option {
    return func(theme *z.Theme, parent z.Widget) {
        w := z.NewTerminal(id, class)
        w.Apply(theme)
        for _, opt := range options {
            opt(theme, w)
        }
        parent.(z.Container).Add(w)
    }
}
```

---

## Keyboard input

When `Terminal` receives a key event, it should translate it to the appropriate
byte sequence and surface it via a `EvtKey` re-dispatch or a dedicated
`EvtInput([]byte)` event so the caller can write bytes to the pty. The exact
mechanism (event vs callback) is deferred to implementation.

---

## Dependencies

- `sync` — `Mutex` for `mu`.
- `unicode/utf8` — for rune width checking in the parser.
- `github.com/mattn/go-runewidth` — `RuneWidth(r)` for wide character detection
  in `Print`. Verify this module is in `go.mod` before use; add if absent.

---

## Limitations (explicit non-goals for v1)

- No scroll-back buffer beyond the visible screen.
- No sixel or kitty graphics protocol.
- No mouse reporting (CSI M escape sequences ignored).
- No DCS sequences (device control strings).
- No bidirectional (BIDI) text.
- No font size / line spacing changes (Kitty font-sizing noted for future).

---

## Verification

```bash
go build ./...
go test ./...
```

**Manual smoke test** in `cmd/demo` or a dedicated `cmd/terminal`:

```go
term := NewTerminal("t", "")
term.Apply(theme)
term.SetHint(80, 24)
// …add to UI…

// Feed test sequences:
term.Write([]byte("\033[1;32mHello\033[0m \033[38;2;255;100;0mWorld\033[0m\n"))
term.Write([]byte("\033[4:3mCurly underline\033[0m\n"))
term.Write([]byte("\033[1mBold\033[0m \033[3mItalic\033[0m \033[9mStrike\033[0m\n"))
```

Expected: green bold "Hello", true-colour orange "World", curly underline,
bold/italic/strikethrough on subsequent lines.

# Terminal

Embedded terminal emulator. Holds two cell buffers (main + alternate screen), processes byte streams via an ANSI parser, and renders the active buffer. Implements `io.Writer` so callers can pipe pty output directly into it.

**Constructor:** `NewTerminal(id, class string) *Terminal`

Default buffer size is 80×24 (VT100). Call `SetHint(w, h)` for a flexible-size terminal.

## Methods

- `Write(data []byte) (int, error)` — feed bytes to the ANSI parser and queue a redraw (safe from any goroutine)
- `Clear()` — clear both buffers, reset cursor and scroll region
- `Resize(w, h int)` — change the buffer dimensions
- `Title() string` — current title set by an OSC `1`/`2` escape sequence
- `SetBounds(x, y, w, h int)` — also resizes the buffer to fit the new content area

## Notes

Flags: `"focusable"`.

Supports a substantial subset of VT100/xterm: SGR colour and attributes (bold, underline with style, italic, strike), cursor movement, scroll regions, alternate screen, line feed / reverse index, DECSC / SCOSC save/restore, OSC titles, 256-colour and true-colour escapes.

Pair with a pty package (`github.com/creack/pty`) to run shells or other tty programs:

```go
term := zw.NewTerminal("term", "")
cmd := exec.Command("/bin/bash")
ptmx, _ := pty.Start(cmd)
go io.Copy(term, ptmx)             // pty output → widget
go io.Copy(ptmx, keyForwarder(term)) // keypresses → pty
```

# Specification: ANSI / VT Parser

**File:** `ansi.go`  
**Package:** `zeichenwerk`

---

## Purpose

A pure-Go, zero-allocation VT/ANSI escape sequence parser driven by a state
machine. It accepts arbitrary byte streams (from a pty, pipe, or test fixture)
and dispatches parsed sequences to an `AnsiHandler`. It has no dependency on
any terminal library and contains no I/O of its own.

The state machine follows the model described in Paul Flo Williams'
"A parser for DEC's ANSI-compatible video terminals" (vt500-parser.md /
state-machine.c) with minor simplifications for the sequences we care about.

---

## Handler interface

```go
// AnsiHandler receives parsed terminal sequences from AnsiParser.
// All methods are called synchronously from within Feed().
type AnsiHandler interface {
    // Print is called for every printable rune (including multi-byte Unicode).
    Print(r rune)

    // Execute is called for C0 control codes (0x00–0x1F excluding ESC)
    // and DEL (0x7F).
    Execute(code byte)

    // CsiDispatch is called when a complete CSI sequence has been parsed.
    //   params  — semicolon-separated numeric parameters; empty params default to 0.
    //             Sub-parameters (colon-separated, e.g. "4:2") are passed as
    //             negative values: param[i] = -(major*100 + minor).
    //   inter   — intermediate byte (0x20–0x2F), or 0 if absent.
    //   final   — final byte (0x40–0x7E).
    CsiDispatch(params []int, inter, final byte)

    // OscDispatch is called when an OSC string is complete (ST or BEL terminator).
    //   cmd    — the leading numeric command (before the first ';'), or 0.
    //   data   — everything after the first ';', or the whole string if no ';'.
    OscDispatch(cmd int, data string)

    // EscDispatch is called for ESC sequences that are not CSI or OSC.
    //   inter — intermediate byte (0x20–0x2F), or 0 if absent.
    //   final — final byte (0x30–0x7E).
    EscDispatch(inter, final byte)
}
```

---

## Parser struct

```go
type AnsiParser struct {
    state   parserState
    params  []int          // parameter accumulator
    cur     int            // digit accumulator for current param
    hasDigit bool          // whether cur contains a digit
    inter   byte           // intermediate byte, 0 if none
    oscBuf  strings.Builder
    utf8buf [4]byte        // partial UTF-8 byte accumulator
    utf8len int            // bytes accumulated so far
    handler AnsiHandler
}

type parserState uint8

const (
    stGround parserState = iota
    stEscape
    stEscInter
    stCsiEntry
    stCsiParam
    stCsiIgnore
    stOscString
)
```

---

## Constructor

```go
func NewAnsiParser(h AnsiHandler) *AnsiParser
```

Creates a parser in the `stGround` state with an empty parameter list.
`h` must not be nil.

---

## Feed

```go
func (p *AnsiParser) Feed(data []byte)
```

Processes all bytes in `data`. May be called repeatedly with partial writes
(e.g. one byte at a time). No internal buffering beyond the state fields.

The method loops over each byte and dispatches to one of the transition
functions below. UTF-8 multi-byte sequences are assembled in `p.utf8buf`
before being emitted as a single `rune` via `handler.Print`.

---

## State machine transitions

### UTF-8 handling (all states except oscString)

Before the state machine processes a byte, UTF-8 continuation bytes
(`0x80`–`0xBF`) are accumulated into `p.utf8buf`. When the sequence is
complete (`utf8.DecodeRune` returns a valid rune and `rune != RuneError`),
`handler.Print(r)` is called and `p.utf8len` resets to 0.

Multi-byte lead bytes (`0xC0`–`0xFF`) start a new accumulation. Single-byte
printable ASCII (`0x20`–`0x7E`) bypasses the accumulator entirely in
`stGround`.

### Ground state

| Byte range | Action |
|------------|--------|
| `0x00`–`0x17`, `0x19`, `0x1C`–`0x1F` | `handler.Execute(b)` |
| `0x1B` | → `stEscape`, reset inter, params |
| `0x20`–`0x7E` | `handler.Print(rune(b))` |
| `0x7F` | `handler.Execute(0x7F)` |
| `0x80`–`0xFF` | UTF-8 accumulator |

### Escape state

| Byte range | Action |
|------------|--------|
| `0x00`–`0x17`, `0x19`, `0x1C`–`0x1F` | `handler.Execute(b)` (stay) |
| `0x1B` | reset; stay in escape |
| `0x20`–`0x2F` | `p.inter = b`; → `stEscInter` |
| `0x30`–`0x4F`, `0x51`–`0x57`, `0x59`–`0x5A`, `0x5C`, `0x60`–`0x7E` | `handler.EscDispatch(0, b)`; → `stGround` |
| `0x5B` (`[`) | → `stCsiEntry`, reset params |
| `0x5D` (`]`) | → `stOscString`, reset oscBuf |
| `0x7F` | ignore (stay) |

### EscInter state

| Byte range | Action |
|------------|--------|
| `0x20`–`0x2F` | `p.inter = b` (overwrite; stay) |
| `0x30`–`0x7E` | `handler.EscDispatch(p.inter, b)`; → `stGround` |
| `0x7F` | ignore (stay) |

### CsiEntry state

| Byte range | Action |
|------------|--------|
| `0x00`–`0x17`, `0x19`, `0x1C`–`0x1F` | `handler.Execute(b)` |
| `0x20`–`0x2F` | `p.inter = b`; → `stCsiParam` |
| `0x30`–`0x38`, `0x3B` | digit/separator; accumulate; → `stCsiParam` |
| `0x3A` (`:`) | sub-parameter separator; → `stCsiParam` |
| `0x3C`–`0x3F` | private marker; `p.inter = b`; → `stCsiParam` |
| `0x40`–`0x7E` | dispatch: `handler.CsiDispatch(params, 0, b)`; → `stGround` |
| `0x7F` | ignore |

### CsiParam state

| Byte range | Action |
|------------|--------|
| `0x30`–`0x39` (`0`–`9`) | digit: `p.cur = p.cur*10 + int(b-'0')`; `p.hasDigit=true` |
| `0x3B` (`;`) | push `p.cur` onto `p.params` (or 0 if !hasDigit); reset cur |
| `0x3A` (`:`) | sub-param separator: push `-(p.cur*100 + next_digit_value)` — see below |
| `0x20`–`0x2F` | `p.inter = b` |
| `0x40`–`0x7E` | finalize: push last cur; dispatch; → `stGround` |
| `0x3C`–`0x3F` | → `stCsiIgnore` (private sequences we don't handle) |
| `0x7F` | ignore |

**Sub-parameter encoding** (`:`): when a colon appears between digits, the parser
encodes the pair `major:minor` as the negative value `-(major*100 + minor)` and
pushes it onto `params`. This allows `CsiDispatch` to detect `4:2` (curly
underline) without a separate parameter slice. Only one level of sub-parameters
is supported.

### CsiIgnore state

All bytes are discarded until a final byte (`0x40`–`0x7E`) is seen, then → `stGround`.

### OscString state

| Byte | Action |
|------|--------|
| `0x07` (BEL) | terminate: parse and dispatch; → `stGround` |
| `0x1B` | → `stEscape` (ST = `ESC \` will trigger dispatch via EscDispatch handler checking for `\\`) |
| `0x20`–`0xFF` | append to `p.oscBuf` |
| `0x00`–`0x06`, `0x08`–`0x1A` | ignore |

**OSC termination and dispatch:**
When the OSC string is complete, split `oscBuf.String()` on the first `";"`:
- Left side → parse as integer → `cmd` (0 if empty or non-numeric).
- Right side → `data`.
- Call `handler.OscDispatch(cmd, data)`.

---

## Parameter finalisation

Before calling `CsiDispatch`, push the final accumulated value:

```
if p.hasDigit {
    p.params = append(p.params, p.cur)
} else if len(p.params) > 0 || current byte is ';' {
    p.params = append(p.params, 0)
}
```

An empty parameter list (no digits, no semicolons) results in `params = nil`,
which `CsiDispatch` implementations treat as `[]int{0}` where a default of 0
is semantically meaningful (e.g. `CSI m` = reset SGR).

The maximum number of parameters is capped at **16** to prevent unbounded
allocation; additional parameters are silently discarded.

---

## SGR dispatch (reference for Terminal implementation)

`CsiDispatch` implementations must handle `final == 'm'`. The params slice
is processed left-to-right as a state machine consuming 1, 3, or 5 elements
per step:

```
consume(params):
  p = next param (or 0 if none)
  switch p:
    0           → reset all attrs and colours
    1           → set bold
    2           → set dim
    3           → set italic
    4           → set underline-single; if negative sub-param → decode style
    5           → set blink
    7           → set reverse
    8           → set invisible
    9           → set strikethrough
    21          → set underline-double (alternative interpretation of 4:2)
    22          → unset bold+dim
    23          → unset italic
    24          → unset underline
    25          → unset blink
    27          → unset reverse
    28          → unset invisible
    29          → unset strikethrough
    30..37      → fg = PaletteColor(p - 30)
    38          → extended FG: consume 38;5;n or 38;2;r;g;b
    39          → fg = ColorDefault
    40..47      → bg = PaletteColor(p - 40)
    48          → extended BG: consume 48;5;n or 48;2;r;g;b
    49          → bg = ColorDefault
    58          → extended UL: consume 58;5;n or 58;2;r;g;b  (Kitty)
    59          → ul = ColorDefault
    90..97      → fg = PaletteColor(p - 90 + 8)   (bright FG)
    100..107    → bg = PaletteColor(p - 100 + 8)  (bright BG)

extended colour (38/48/58 n) consume:
    next = params[i+1]
    if next == 5:     colour = PaletteColor(params[i+2])  ; advance by 2
    if next == 2:     colour = TrueColor(params[i+2], params[i+3], params[i+4]) ; advance by 4
    otherwise: ignore
```

**Sub-parameter form** (e.g. `38:2:255:0:128`): when `params[i]` is negative,
it encodes `38:2` as `-(3802)`. Extract major = abs/100, minor = abs%100.
The following parameters carry r, g, b as normal positive values.

---

## Execute codes handled by Terminal

| Code | Name | Action |
|------|------|--------|
| `0x07` | BEL | bell (ignored or audible, implementation choice) |
| `0x08` | BS | cursor left 1 (clamp at column 0) |
| `0x09` | HT | advance cursor to next tab stop (every 8 columns) |
| `0x0A` | LF | cursor down 1; scroll if at bottom |
| `0x0B` | VT | same as LF |
| `0x0C` | FF | same as LF |
| `0x0D` | CR | cursor to column 0 |
| `0x7F` | DEL | ignore |

---

## CSI sequences handled by Terminal

| Final | Params | Name | Action |
|-------|--------|------|--------|
| `A` | n=1 | CUU | cursor up n |
| `B` | n=1 | CUD | cursor down n |
| `C` | n=1 | CUF | cursor forward n |
| `D` | n=1 | CUB | cursor back n |
| `E` | n=1 | CNL | cursor next line n |
| `F` | n=1 | CPL | cursor previous line n |
| `G` | n=1 | CHA | cursor horizontal absolute (1-based) |
| `H` | r=1,c=1 | CUP | cursor position (1-based row, col) |
| `J` | n=0 | ED | erase display: 0=cursor→end, 1=start→cursor, 2=all, 3=all+scrollback |
| `K` | n=0 | EL | erase line: 0=cursor→end, 1=start→cursor, 2=all |
| `L` | n=1 | IL | insert n blank lines |
| `M` | n=1 | DL | delete n lines |
| `P` | n=1 | DCH | delete n characters |
| `S` | n=1 | SU | scroll up n lines |
| `T` | n=1 | SD | scroll down n lines |
| `X` | n=1 | ECH | erase n characters |
| `d` | n=1 | VPA | cursor vertical absolute (1-based) |
| `f` | r=1,c=1 | HVP | same as CUP |
| `h` | — | SM | mode set — see modes table |
| `l` | — | RM | mode reset — see modes table |
| `m` | — | SGR | select graphic rendition — see SGR table |
| `r` | t=1,b=rows | DECSTBM | set scrolling region (top, bottom, 1-based) |
| `s` | — | SCOSC | save cursor position |
| `u` | — | SCORC | restore cursor position |

**Missing parameter defaults:** if a parameter is absent, use the default shown
in the Params column. For `H`/`f`, both params default to 1 independently.

### Modes (`h`/`l` with `?` private marker in inter)

| Mode | `?` | Effect |
|------|-----|--------|
| 7 | yes | Auto-wrap (set on, reset off; on by default) |
| 25 | yes | Cursor visibility (set=visible, reset=hidden) |
| 1049 | yes | Alternate screen (set=enter, reset=leave) |

Unrecognised modes are silently ignored.

---

## ESC sequences handled by Terminal

| Final | Inter | Name | Action |
|-------|-------|------|--------|
| `7` | 0 | DECSC | save cursor (position + attrs + colours) |
| `8` | 0 | DECRC | restore cursor |
| `M` | 0 | RI | reverse index: cursor up 1, scroll down if at top of scroll region |
| `c` | 0 | RIS | hard reset (clear screen, reset cursor, attrs, scroll region) |

---

## Error handling

The parser never returns errors. Unrecognised sequences are consumed and
discarded. This matches real terminal emulator behaviour — garbage in, silence
out, no exception.

---

## Testing

Each state transition should have a unit test. Key test cases:

- Plain ASCII text → series of `Print` calls
- `ESC [ 1 m` → `CsiDispatch([1], 0, 'm')`
- `ESC [ 38 ; 2 ; 255 ; 128 ; 0 m` → true-colour FG SGR
- `ESC [ 4 : 3 m` → curly underline sub-param
- `ESC ] 0 ; title BEL` → `OscDispatch(0, "title")`
- `ESC 7` → `EscDispatch(0, '7')`
- Multi-byte UTF-8 `é` (0xC3 0xA9) → single `Print('é')`
- Partial write: feed one byte at a time, verify same result as single feed
- Parameter capping: 17 semicolons → only 16 params dispatched

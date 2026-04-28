# Specification: CellBuffer

**File:** `cell-buffer.go`  
**Package:** `zeichenwerk`

---

## Purpose

`CellBuffer` is a compact, flat 2D grid of terminal cells used by the `Terminal`
widget. It stores character, colour, and attribute data in four parallel `[]uint32`
slices rather than a slice of structs, so that bulk operations (clear, copy,
resize) touch only the relevant data array.

---

## Color type

```go
type Color uint32
```

### Encoding

| Value range | Meaning |
|-------------|---------|
| `0x00000000` | `ColorDefault` — use the terminal's own default colour |
| `0x00000001`–`0x00000100` | Palette colour: `value - 1` = xterm-256 index (0–255) |
| `0x80000000`–`0x80FFFFFF` | True colour: bits 0–23 = `0xRRGGBB` |
| All other values | Reserved / invalid |

The bit-31 flag unambiguously separates true colour from palette/default:

```
bit 31 = 0, value = 0          → default
bit 31 = 0, value = 1..256     → palette index value-1
bit 31 = 1, bits 0-23 = RGB    → true colour
```

### Constants and constructors

```go
const ColorDefault Color = 0

func PaletteColor(index int) Color {
    // precondition: 0 <= index <= 255
    return Color(index + 1)
}

func TrueColor(r, g, b uint8) Color {
    return Color(0x80000000 | uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}
```

### colorToHex

```go
func colorToHex(c Color) string
```

Converts a `Color` to a string accepted by `renderer.Set(fg, bg, font)`:

- `ColorDefault` → `""` (empty — renderer inherits terminal/widget default)
- Palette 0–255 → look up `xterm256[index]` (embedded `[256][3]uint8` table),
  return `"#RRGGBB"`
- True colour → `fmt.Sprintf("#%06X", c & 0x00FFFFFF)`

The xterm-256 RGB table must be embedded in the file as a package-level
`var xterm256 = [256][3]uint8{…}`. Values are the canonical xterm palette
as standardised by XFree86 / xterm source.

---

## char uint32 — bit layout

Each cell's character word packs rune, cell-width flag, text attributes, and
underline style into one `uint32`:

```
 31  30  29  28  27  26  25  24  23  22  21  20 ........... 0
  └──┬──┘   │   │   │   │   │   │   │   │   └──────────────┘
 ul_style  inv stk rev blk itl dim bld wid      rune (21 bits)
 (3 bits)
```

| Bits | Name | Description |
|------|------|-------------|
| 0–20 | `rune` | Unicode code point (0–0x10FFFF). 0 = empty cell (render as space) |
| 21 | `wide` | 1 = double-width cell (Kitty/East Asian wide chars). The cell to the right is a continuation placeholder (`rune=0, wide=0`) |
| 22 | `bold` | Bold text |
| 23 | `dim` | Faint / dim text |
| 24 | `italic` | Italic text |
| 25 | `blink` | Blinking text |
| 26 | `reverse` | Reverse video (swap fg/bg at render time) |
| 27 | `strikethrough` | Strikethrough |
| 28 | `invisible` | Invisible text (render as space) |
| 29–31 | `ul_style` | Underline style: 0=none, 1=single, 2=double, 3=curly, 4=dotted, 5=dashed |

### Bit constants

```go
const (
    charRuneMask  uint32 = 0x001FFFFF // bits 0-20
    charWide      uint32 = 1 << 21
    charBold      uint32 = 1 << 22
    charDim       uint32 = 1 << 23
    charItalic    uint32 = 1 << 24
    charBlink     uint32 = 1 << 25
    charReverse   uint32 = 1 << 26
    charStrike    uint32 = 1 << 27
    charInvis     uint32 = 1 << 28
    charULMask    uint32 = 0xE0000000 // bits 29-31
    charULShift          = 29

    // Underline style values (stored in bits 29-31)
    ULNone   = 0
    ULSingle = 1
    ULDouble = 2
    ULCurly  = 3
    ULDotted = 4
    ULDashed = 5
)
```

### Attribute helpers

```go
// PackAttrs builds the attribute portion of a char word (bits 21-31).
// attrs is a bitmask of charBold | charDim | … constants.
// ulStyle is ULNone..ULDashed.
func PackAttrs(attrs uint32, ulStyle int) uint32

// UnpackAttrs extracts attrs bits and underline style from a char word.
func UnpackAttrs(char uint32) (attrs uint32, ulStyle int)

// attrsToFont converts the attr bits to a space-separated font string
// compatible with renderer.Set: "bold", "italic", "underline", etc.
// Underline is emitted as "underline" when ulStyle != ULNone.
func attrsToFont(char uint32) string
```

---

## CellBuffer struct

```go
type CellBuffer struct {
    char []uint32  // rune + attrs + ul-style (one per cell)
    fg   []uint32  // foreground Color (one per cell)
    bg   []uint32  // background Color (one per cell)
    ul   []uint32  // underline Color (one per cell, Kitty extension)
    w, h int
}
```

Cell at `(x, y)` is at index `i = y*w + x`. All four slices have length `w*h`.

---

## Constructor

```go
func NewCellBuffer(w, h int) *CellBuffer
```

- Allocates all four slices of length `w*h`.
- Fills with default values: `char[i]=0`, `fg[i]=0`, `bg[i]=0`, `ul[i]=0`.
- Panics if `w < 1` or `h < 1`.

---

## Methods

### Dimensions

```go
func (b *CellBuffer) Width() int
func (b *CellBuffer) Height() int
```

### Get

```go
func (b *CellBuffer) Get(x, y int) (ch rune, fg, bg, ul Color, attrs uint32)
```

- Returns the rune, three colours, and the raw `char` word for attribute
  inspection (bits 21–31 only; the rune bits are already returned as `ch`).
- Returns zero values for all fields if `x` or `y` is out of bounds — does
  not panic.
- `ch == 0` means empty cell; callers should render as space.

### Set

```go
func (b *CellBuffer) Set(x, y int, ch rune, fg, bg, ul Color, attrs uint32)
```

- `attrs` is the packed char word (bits 21–31); callers use `PackAttrs`.
- Silently ignores out-of-bounds coordinates.
- Stores `(uint32(ch) & charRuneMask) | (attrs & ^charRuneMask)` in `char[i]`.
- Stores `fg`, `bg`, `ul` as raw `uint32` in the corresponding slices.

### SetChar

```go
func (b *CellBuffer) SetChar(x, y int, ch rune)
```

Updates only the rune in `char[i]`, leaving attribute and colour data unchanged.
Out-of-bounds is silently ignored.

### Clear

```go
func (b *CellBuffer) Clear()
```

Sets all cells to the default state: `char[i]=0`, `fg[i]=0`, `bg[i]=0`,
`ul[i]=0`. Equivalent to filling with space and default colours.

Uses `clear(b.char)` / `clear(b.fg)` / `clear(b.bg)` / `clear(b.ul)` (Go 1.21
built-in) for performance.

### ClearLine

```go
func (b *CellBuffer) ClearLine(y, x1, x2 int)
```

Clears cells `x1..x2-1` on row `y` to default. Clamps to buffer bounds.
Used by ANSI erase-in-line (`CSI K`) operations.

### Resize

```go
func (b *CellBuffer) Resize(w, h int)
```

Resizes the buffer to `w × h`, preserving as much existing content as possible:

1. If the new size equals the current size, return immediately.
2. Allocate new slices of length `w*h`, zero-initialised.
3. Copy rows from the old buffer: for each row `y` in `0..min(old_h, h)-1`,
   copy columns `0..min(old_w, w)-1` from old to new.
4. Replace `b.char`, `b.fg`, `b.bg`, `b.ul` with the new slices.
5. Update `b.w`, `b.h`.
6. Panics if `w < 1` or `h < 1`.

---

## Thread safety

`CellBuffer` is **not** thread-safe. The `Terminal` widget owns its buffers and
must not be written to from a goroutine other than the UI event loop without
external synchronisation.

---

## xterm-256 colour table

The table maps palette indices 0–255 to RGB triples. The canonical values are:

- **0–7**: Standard colours (black, red, green, yellow, blue, magenta, cyan, white)
- **8–15**: Bright/high-intensity variants
- **16–231**: 6×6×6 colour cube: `index = 16 + 36*r + 6*g + b` where r,g,b ∈ 0–5;
  RGB component = `if v == 0 { 0 } else { 55 + 40*v }`
- **232–255**: Greyscale ramp: `value = 8 + 10*(index-232)` for all three channels

The table must be defined in `cell-buffer.go` as a package-level variable so
that `colorToHex` can reference it without allocation.

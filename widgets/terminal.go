package widgets

import (
	"sync"

	"github.com/rivo/uniseg"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ==== AI ===================================================================

// termCursor holds the current cursor state.
type termCursor struct {
	x, y  int         // 0-based column and row
	fg    Color       // current foreground colour
	bg    Color       // current background colour
	ul    Color       // current underline colour
	attrs uint32      // packed char word bits 21-31 (attrs + ul_style)
	wrap  bool        // deferred-wrap: next Print moves to next line first
	saved *termCursor // single DECSC/SCOSC save slot; nil if nothing saved
}

// Terminal is a zeichenwerk widget that renders arbitrary terminal output.
// It holds two CellBuffers (main and alternate screen), processes byte streams
// via an embedded AnsiParser, and renders the active buffer via the Renderer.
// It implements io.Writer so callers can pipe pty output directly into it.
type Terminal struct {
	Component
	main    *CellBuffer
	alt     *CellBuffer
	active  *CellBuffer // points to main or alt
	parser  *AnsiParser
	handler *termHandler
	cur     termCursor
	scroll  struct {
		top int // first row of scroll region (0-based, default 0)
		bot int // last row of scroll region (0-based, default height-1)
	}
	title      string
	autoWrap   bool
	showCursor bool
	mu         sync.Mutex
}

// termHandler implements AnsiHandler. Methods are always called from within
// Terminal.Write which holds mu, so no additional locking is needed.
type termHandler struct {
	t *Terminal
}

// NewTerminal creates a Terminal widget with the default VT100 size (80×24).
func NewTerminal(id, class string) *Terminal {
	t := &Terminal{
		Component:  Component{id: id, class: class},
		main:       NewCellBuffer(80, 24),
		alt:        NewCellBuffer(80, 24),
		autoWrap:   true,
		showCursor: true,
	}
	t.active = t.main
	t.cur = termCursor{
		fg: ColorDefault,
		bg: ColorDefault,
		ul: ColorDefault,
	}
	t.scroll.top = 0
	t.scroll.bot = 23
	t.handler = &termHandler{t: t}
	t.parser = NewAnsiParser(t.handler)
	t.SetFlag(FlagFocusable, true)
	return t
}

// Write implements io.Writer. It feeds data to the ANSI parser and schedules
// a repaint. Safe to call from any goroutine.
func (t *Terminal) Write(data []byte) (int, error) {
	t.mu.Lock()
	t.parser.Feed(data)
	t.mu.Unlock()
	Redraw(t)
	return len(data), nil
}

// Apply applies "terminal" and "terminal:focused" styles from the theme.
func (t *Terminal) Apply(theme *Theme) {
	theme.Apply(t, t.Selector("terminal"))
}

// Hint returns the buffer dimensions (80×24 by default).
// Callers wanting a flexible-size terminal should call SetHint.
func (t *Terminal) Hint() (int, int) {
	if t.hwidth != 0 || t.hheight != 0 {
		return t.hwidth, t.hheight
	}
	return t.main.Width(), t.main.Height()
}

// Clear clears both buffers, resets cursor and scroll region.
func (t *Terminal) Clear() {
	t.mu.Lock()
	t.main.Clear()
	t.alt.Clear()
	t.cur.x, t.cur.y = 0, 0
	t.cur.wrap = false
	t.scroll.top = 0
	t.scroll.bot = t.main.Height() - 1
	t.mu.Unlock()
	Redraw(t)
}

// Resize resizes both buffers and clamps cursor/scroll region.
// If the scroll region covered the full old buffer (i.e. it was the default,
// never explicitly set via DECSTBM), it is expanded to cover the full new
// buffer so that the terminal does not start scrolling at the old height.
func (t *Terminal) Resize(w, h int) {
	t.mu.Lock()
	oldH := t.main.Height()
	wasFullScreen := t.scroll.top == 0 && t.scroll.bot == oldH-1
	t.main.Resize(w, h)
	t.alt.Resize(w, h)
	if wasFullScreen {
		t.scroll.bot = h - 1
	}
	t.clampCursor()
	t.clampScroll()
	t.mu.Unlock()
}

// SetBounds overrides Component.SetBounds to immediately resize the cell
// buffers to match the new content area. This ensures that Write() always
// writes into a buffer that is sized to the visible widget area, regardless
// of whether Render() has been called yet.
func (t *Terminal) SetBounds(x, y, w, h int) {
	t.Component.SetBounds(x, y, w, h)
	style := t.Style()
	cw := w - style.Horizontal()
	ch := h - style.Vertical()
	if cw > 0 && ch > 0 {
		t.Resize(cw, ch)
	}
}

// Render draws the active buffer to the screen.
func (t *Terminal) Render(r *Renderer) {
	t.Component.Render(r)

	x0, y0, cw, ch := t.Content()
	if cw <= 0 || ch <= 0 {
		return
	}

	t.mu.Lock()

	// Resize buffers if widget dimensions changed.
	if t.active.Width() != cw || t.active.Height() != ch {
		t.main.Resize(cw, ch)
		t.alt.Resize(cw, ch)
		t.clampCursor()
		t.clampScroll()
	}

	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			glyph, fg, bg, ul, attrs := t.active.Get(x, y)

			// Apply reverse-video: swap fg and bg at render time.
			if attrs&charReverse != 0 {
				fg, bg = bg, fg
			}

			fgStr := colorToHex(fg)
			bgStr := colorToHex(bg)
			font := attrsToFont(attrs)

			r.Set(fgStr, bgStr, font)

			// Underline colour and style.
			if attrs&charULMask != 0 {
				ulStyle := int((attrs & charULMask) >> charULShift)
				r.SetUnderline(ulStyle, colorToHex(ul))
			}

			s := string(glyph)
			if glyph == 0 || attrs&charInvis != 0 {
				s = " "
			}
			r.Put(x0+x, y0+y, s)
		}
	}

	t.mu.Unlock()
}

// Title returns the OSC window title, if set.
func (t *Terminal) Title() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.title
}

// ---- Internal helpers -------------------------------------------------------

func (t *Terminal) lineFeed() {
	if t.cur.y < t.scroll.bot {
		t.cur.y++
	} else {
		t.scrollUp(1)
	}
	t.cur.wrap = false
}

func (t *Terminal) reverseIndex() {
	if t.cur.y > t.scroll.top {
		t.cur.y--
	} else {
		t.scrollDown(1)
	}
}

func (t *Terminal) scrollUp(n int) {
	for i := 0; i < n; i++ {
		for row := t.scroll.top + 1; row <= t.scroll.bot; row++ {
			t.active.copyRow(row-1, row)
		}
		t.active.ClearLineColor(t.scroll.bot, 0, t.active.Width(), t.cur.bg)
	}
}

func (t *Terminal) scrollDown(n int) {
	for i := 0; i < n; i++ {
		for row := t.scroll.bot - 1; row >= t.scroll.top; row-- {
			t.active.copyRow(row+1, row)
		}
		t.active.ClearLineColor(t.scroll.top, 0, t.active.Width(), t.cur.bg)
	}
}

func (t *Terminal) clampCursor() {
	w := t.active.Width()
	h := t.active.Height()
	if t.cur.x >= w {
		t.cur.x = w - 1
	}
	if t.cur.x < 0 {
		t.cur.x = 0
	}
	if t.cur.y >= h {
		t.cur.y = h - 1
	}
	if t.cur.y < 0 {
		t.cur.y = 0
	}
}

func (t *Terminal) clampScroll() {
	h := t.active.Height()
	if t.scroll.top < 0 {
		t.scroll.top = 0
	}
	if t.scroll.top >= h {
		t.scroll.top = h - 1
	}
	if t.scroll.bot < 0 {
		t.scroll.bot = 0
	}
	if t.scroll.bot >= h {
		t.scroll.bot = h - 1
	}
	if t.scroll.top >= t.scroll.bot {
		t.scroll.bot = t.scroll.top + 1
		if t.scroll.bot >= h {
			t.scroll.bot = h - 1
			t.scroll.top = t.scroll.bot - 1
			if t.scroll.top < 0 {
				t.scroll.top = 0
			}
		}
	}
}

func (t *Terminal) hardReset() {
	t.main.Clear()
	t.alt.Clear()
	t.active = t.main
	t.cur = termCursor{fg: ColorDefault, bg: ColorDefault, ul: ColorDefault}
	t.scroll.top = 0
	t.scroll.bot = t.main.Height() - 1
	t.autoWrap = true
	t.showCursor = true
}

// applySGR processes SGR (Select Graphic Rendition) parameters.
func (t *Terminal) applySGR(params []int) {
	if len(params) == 0 {
		params = []int{0}
	}
	i := 0
	for i < len(params) {
		p := params[i]
		i++

		// Handle sub-parameter form (negative value = major:minor encoded)
		if p < 0 {
			major := (-p) / 100
			minor := (-p) % 100
			switch major {
			case 4:
				// Underline style: 4:0=off, 4:1=single, 4:2=double, 4:3=curly, 4:4=dotted, 4:5=dashed
				t.cur.attrs &^= charULMask
				if minor >= ULSingle && minor <= ULDashed {
					t.cur.attrs |= uint32(minor) << charULShift
				}
			case 38:
				// Extended FG colour via sub-params not commonly used this way; skip
			case 48:
				// Extended BG colour; skip
			case 58:
				// Extended UL colour; skip
			}
			continue
		}

		switch {
		case p == 0:
			// Reset all
			t.cur.fg = ColorDefault
			t.cur.bg = ColorDefault
			t.cur.ul = ColorDefault
			t.cur.attrs = 0
		case p == 1:
			t.cur.attrs |= charBold
		case p == 2:
			t.cur.attrs |= charDim
		case p == 3:
			t.cur.attrs |= charItalic
		case p == 4:
			// Underline single (or check next for style sub-param)
			t.cur.attrs &^= charULMask
			t.cur.attrs |= uint32(ULSingle) << charULShift
		case p == 5:
			t.cur.attrs |= charBlink
		case p == 7:
			t.cur.attrs |= charReverse
		case p == 8:
			t.cur.attrs |= charInvis
		case p == 9:
			t.cur.attrs |= charStrike
		case p == 21:
			// Double underline (some terminals use 4:2)
			t.cur.attrs &^= charULMask
			t.cur.attrs |= uint32(ULDouble) << charULShift
		case p == 22:
			t.cur.attrs &^= charBold | charDim
		case p == 23:
			t.cur.attrs &^= charItalic
		case p == 24:
			t.cur.attrs &^= charULMask
		case p == 25:
			t.cur.attrs &^= charBlink
		case p == 27:
			t.cur.attrs &^= charReverse
		case p == 28:
			t.cur.attrs &^= charInvis
		case p == 29:
			t.cur.attrs &^= charStrike
		case p >= 30 && p <= 37:
			t.cur.fg = PaletteColor(p - 30)
		case p == 38:
			c, consumed := t.parseExtendedColor(params, i)
			i += consumed
			if consumed > 0 {
				t.cur.fg = c
			}
		case p == 39:
			t.cur.fg = ColorDefault
		case p >= 40 && p <= 47:
			t.cur.bg = PaletteColor(p - 40)
		case p == 48:
			c, consumed := t.parseExtendedColor(params, i)
			i += consumed
			if consumed > 0 {
				t.cur.bg = c
			}
		case p == 49:
			t.cur.bg = ColorDefault
		case p == 58:
			c, consumed := t.parseExtendedColor(params, i)
			i += consumed
			if consumed > 0 {
				t.cur.ul = c
			}
		case p == 59:
			t.cur.ul = ColorDefault
		case p >= 90 && p <= 97:
			t.cur.fg = PaletteColor(p - 90 + 8)
		case p >= 100 && p <= 107:
			t.cur.bg = PaletteColor(p - 100 + 8)
		}
	}
}

// parseExtendedColor reads an extended color (38;5;n or 38;2;r;g;b) starting
// at params[start]. Returns the Color and how many params were consumed.
func (t *Terminal) parseExtendedColor(params []int, start int) (Color, int) {
	if start >= len(params) {
		return ColorDefault, 0
	}
	switch params[start] {
	case 5:
		// Palette: 38;5;n
		if start+1 < len(params) {
			return PaletteColor(params[start+1]), 2
		}
	case 2:
		// True colour: 38;2;r;g;b
		if start+3 < len(params) {
			r := uint8(params[start+1])
			g := uint8(params[start+2])
			b := uint8(params[start+3])
			return TrueColor(r, g, b), 4
		}
	}
	return ColorDefault, 0
}

// ---- termHandler: AnsiHandler implementation --------------------------------

func (h *termHandler) Print(r rune) {
	t := h.t

	// Deferred wrap: move to next line before printing.
	if t.cur.wrap {
		t.lineFeed()
		t.cur.x = 0
		t.cur.wrap = false
	}

	// Determine display width of rune.
	rw := uniseg.StringWidth(string(r))
	if rw < 1 {
		rw = 1
	}

	// Write the rune into the active buffer.
	t.active.Set(t.cur.x, t.cur.y, r, t.cur.fg, t.cur.bg, t.cur.ul, t.cur.attrs)

	// For wide characters, fill the continuation cell.
	if rw == 2 && t.cur.x+1 < t.active.Width() {
		wideAttrs := t.cur.attrs | charWide
		t.active.Set(t.cur.x, t.cur.y, r, t.cur.fg, t.cur.bg, t.cur.ul, wideAttrs)
		t.active.Set(t.cur.x+1, t.cur.y, 0, t.cur.fg, t.cur.bg, t.cur.ul, 0)
	}

	// Advance cursor.
	t.cur.x += rw
	if t.cur.x >= t.active.Width() {
		if t.autoWrap {
			t.cur.wrap = true
			t.cur.x = t.active.Width() - 1
		} else {
			t.cur.x = t.active.Width() - 1
		}
	}
}

func (h *termHandler) Execute(code byte) {
	t := h.t
	switch code {
	case 0x07: // BEL
		// no-op
	case 0x08: // BS
		if t.cur.x > 0 {
			t.cur.x--
		}
		t.cur.wrap = false
	case 0x09: // HT (tab)
		next := (t.cur.x/8 + 1) * 8
		if next >= t.active.Width() {
			next = t.active.Width() - 1
		}
		t.cur.x = next
	case 0x0A, 0x0B, 0x0C: // LF, VT, FF
		t.lineFeed()
	case 0x0D: // CR
		t.cur.x = 0
		t.cur.wrap = false
	case 0x7F: // DEL
		// no-op
	}
}

func (h *termHandler) CsiDispatch(params []int, inter, final byte) {
	t := h.t

	// Helper to get a parameter with a default value.
	param := func(idx, def int) int {
		if idx < len(params) && params[idx] != 0 {
			return params[idx]
		}
		return def
	}
	// Helper for params that default to 1 even when explicitly 0
	p1 := func(idx int) int { return param(idx, 1) }

	w := t.active.Width()
	h2 := t.active.Height()

	switch final {
	case 'A': // CUU - cursor up
		n := p1(0)
		t.cur.y -= n
		if t.cur.y < t.scroll.top {
			t.cur.y = t.scroll.top
		}
		t.cur.wrap = false

	case 'B': // CUD - cursor down
		n := p1(0)
		t.cur.y += n
		if t.cur.y > t.scroll.bot {
			t.cur.y = t.scroll.bot
		}
		t.cur.wrap = false

	case 'C': // CUF - cursor forward
		n := p1(0)
		t.cur.x += n
		if t.cur.x >= w {
			t.cur.x = w - 1
		}
		t.cur.wrap = false

	case 'D': // CUB - cursor back
		n := p1(0)
		t.cur.x -= n
		if t.cur.x < 0 {
			t.cur.x = 0
		}
		t.cur.wrap = false

	case 'E': // CNL - cursor next line
		n := p1(0)
		t.cur.y += n
		if t.cur.y > t.scroll.bot {
			t.cur.y = t.scroll.bot
		}
		t.cur.x = 0
		t.cur.wrap = false

	case 'F': // CPL - cursor previous line
		n := p1(0)
		t.cur.y -= n
		if t.cur.y < t.scroll.top {
			t.cur.y = t.scroll.top
		}
		t.cur.x = 0
		t.cur.wrap = false

	case 'G': // CHA - cursor horizontal absolute (1-based)
		col := p1(0) - 1
		if col < 0 {
			col = 0
		}
		if col >= w {
			col = w - 1
		}
		t.cur.x = col
		t.cur.wrap = false

	case 'H', 'f': // CUP/HVP - cursor position (1-based row, col)
		row := p1(0) - 1
		col := p1(1) - 1
		if row < 0 {
			row = 0
		}
		if row >= h2 {
			row = h2 - 1
		}
		if col < 0 {
			col = 0
		}
		if col >= w {
			col = w - 1
		}
		t.cur.y = row
		t.cur.x = col
		t.cur.wrap = false

	case 'J': // ED - erase display
		n := param(0, 0)
		switch n {
		case 0: // cursor to end of screen
			t.active.ClearLineColor(t.cur.y, t.cur.x, w, t.cur.bg)
			for y := t.cur.y + 1; y < h2; y++ {
				t.active.ClearLineColor(y, 0, w, t.cur.bg)
			}
		case 1: // start to cursor
			for y := 0; y < t.cur.y; y++ {
				t.active.ClearLineColor(y, 0, w, t.cur.bg)
			}
			t.active.ClearLineColor(t.cur.y, 0, t.cur.x+1, t.cur.bg)
		case 2, 3: // entire screen
			for y := 0; y < h2; y++ {
				t.active.ClearLineColor(y, 0, w, t.cur.bg)
			}
		}

	case 'K': // EL - erase line
		n := param(0, 0)
		switch n {
		case 0: // cursor to end of line
			t.active.ClearLineColor(t.cur.y, t.cur.x, w, t.cur.bg)
		case 1: // start to cursor
			t.active.ClearLineColor(t.cur.y, 0, t.cur.x+1, t.cur.bg)
		case 2: // entire line
			t.active.ClearLineColor(t.cur.y, 0, w, t.cur.bg)
		}

	case 'L': // IL - insert n blank lines
		n := p1(0)
		for i := 0; i < n; i++ {
			for row := t.scroll.bot; row > t.cur.y; row-- {
				t.active.copyRow(row, row-1)
			}
			t.active.ClearLineColor(t.cur.y, 0, w, t.cur.bg)
		}

	case 'M': // DL - delete n lines
		n := p1(0)
		for i := 0; i < n; i++ {
			for row := t.cur.y; row < t.scroll.bot; row++ {
				t.active.copyRow(row, row+1)
			}
			t.active.ClearLineColor(t.scroll.bot, 0, w, t.cur.bg)
		}

	case 'P': // DCH - delete n characters (shift left)
		n := p1(0)
		y := t.cur.y
		for x := t.cur.x; x < w-n; x++ {
			glyph, fg, bg, ul, attrs := t.active.Get(x+n, y)
			t.active.Set(x, y, glyph, fg, bg, ul, attrs)
		}
		end := w - n
		if end < t.cur.x {
			end = t.cur.x
		}
		t.active.ClearLineColor(y, end, w, t.cur.bg)

	case 'S': // SU - scroll up n lines
		n := p1(0)
		t.scrollUp(n)

	case 'T': // SD - scroll down n lines
		n := p1(0)
		t.scrollDown(n)

	case 'X': // ECH - erase n characters (no shift)
		n := p1(0)
		t.active.ClearLineColor(t.cur.y, t.cur.x, t.cur.x+n, t.cur.bg)

	case 'd': // VPA - vertical position absolute (1-based)
		row := p1(0) - 1
		if row < 0 {
			row = 0
		}
		if row >= h2 {
			row = h2 - 1
		}
		t.cur.y = row
		t.cur.wrap = false

	case 'h', 'l': // SM/RM - mode set/reset
		set := final == 'h'
		isPrivate := inter == '?'
		for _, mode := range params {
			if isPrivate {
				switch mode {
				case 7: // Auto-wrap
					t.autoWrap = set
				case 25: // Cursor visibility
					t.showCursor = set
				case 1049: // Alternate screen
					if set {
						// Enter alt screen: clear it, reset cursor
						t.alt.Clear()
						t.active = t.alt
						t.cur.x, t.cur.y = 0, 0
						t.cur.wrap = false
					} else {
						// Return to main screen
						t.active = t.main
					}
					t.clampScroll()
				}
			}
		}

	case 'm': // SGR
		t.applySGR(params)

	case 'r': // DECSTBM - set scrolling region (1-based top, bottom)
		top := p1(0) - 1
		bot := p1(1) - 1
		if bot == 0 {
			bot = h2 - 1
		}
		if top < 0 {
			top = 0
		}
		if bot >= h2 {
			bot = h2 - 1
		}
		if top < bot {
			t.scroll.top = top
			t.scroll.bot = bot
		}
		t.cur.x, t.cur.y = 0, 0
		t.cur.wrap = false

	case 's': // SCOSC - save cursor position
		saved := t.cur
		t.cur.saved = &saved

	case 'u': // SCORC - restore cursor position
		if t.cur.saved != nil {
			saved := *t.cur.saved
			t.cur = saved
		}
	}
}

func (h *termHandler) OscDispatch(cmd int, data string) {
	t := h.t
	switch cmd {
	case 0, 1, 2:
		t.title = data
	}
}

func (h *termHandler) EscDispatch(inter, final byte) {
	t := h.t
	if inter != 0 {
		return
	}
	switch final {
	case '7': // DECSC - save cursor
		saved := t.cur
		t.cur.saved = &saved
	case '8': // DECRC - restore cursor
		if t.cur.saved != nil {
			saved := *t.cur.saved
			t.cur = saved
		}
	case 'M': // RI - reverse index
		t.reverseIndex()
	case 'c': // RIS - hard reset
		t.hardReset()
	}
}

package zeichenwerk

import (
	"strings"
	"unicode/utf8"
)

// ==== AI ===================================================================

// AnsiHandler receives parsed terminal sequences from AnsiParser.
// All methods are called synchronously from within Feed().
type AnsiHandler interface {
	// Print is called for every printable rune (including multi-byte Unicode).
	Print(r rune)

	// Execute is called for C0 control codes (0x00–0x1F excluding ESC) and DEL (0x7F).
	Execute(code byte)

	// CsiDispatch is called when a complete CSI sequence has been parsed.
	// params are semicolon-separated numeric parameters; sub-parameters (colon-
	// separated, e.g. "4:2") are encoded as negative values: -(major*100 + minor).
	// inter is the intermediate byte (0x20–0x2F), or 0 if absent.
	// final is the final byte (0x40–0x7E).
	CsiDispatch(params []int, inter, final byte)

	// OscDispatch is called when an OSC string is complete.
	// cmd is the leading numeric command (before the first ';'), or 0.
	// data is everything after the first ';'.
	OscDispatch(cmd int, data string)

	// EscDispatch is called for ESC sequences that are not CSI or OSC.
	// inter is the intermediate byte (0x20–0x2F), or 0 if absent.
	// final is the final byte (0x30–0x7E).
	EscDispatch(inter, final byte)
}

// ---- Parser state machine --------------------------------------------------

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

const maxParams = 16

// AnsiParser is a pure-Go VT/ANSI escape sequence parser.
// It is driven by Feed() and dispatches parsed sequences to an AnsiHandler.
// It contains no I/O and has no external dependencies beyond the standard library.
type AnsiParser struct {
	state    parserState
	params   []int // parameter accumulator
	cur      int   // digit accumulator for current parameter
	hasDigit bool  // whether cur contains a digit
	inSub    bool  // true when accumulating a sub-parameter (after ':')
	subMajor int   // major value saved when ':' was seen
	inter    byte  // intermediate byte, 0 if none
	oscBuf   strings.Builder
	utf8buf  [4]byte // partial UTF-8 byte accumulator
	utf8len  int     // bytes accumulated so far
	handler  AnsiHandler
}

// NewAnsiParser creates a parser in the ground state with an empty parameter list.
// h must not be nil.
func NewAnsiParser(h AnsiHandler) *AnsiParser {
	return &AnsiParser{
		state:   stGround,
		params:  make([]int, 0, 8),
		handler: h,
	}
}

// Feed processes all bytes in data. May be called repeatedly with partial writes.
func (p *AnsiParser) Feed(data []byte) {
	for _, b := range data {
		// UTF-8 multi-byte accumulation (all states except oscString)
		if p.state != stOscString {
			if p.utf8len > 0 {
				// Continuation byte
				if b >= 0x80 && b <= 0xBF {
					p.utf8buf[p.utf8len] = b
					p.utf8len++
					if r, size := utf8.DecodeRune(p.utf8buf[:p.utf8len]); r != utf8.RuneError && size > 0 {
						p.handler.Print(r)
						p.utf8len = 0
					} else if p.utf8len == 4 {
						// Max length reached, discard
						p.utf8len = 0
					}
					continue
				}
				// Non-continuation byte: discard partial and process normally
				p.utf8len = 0
			}
			// Multi-byte lead byte
			if b >= 0xC0 {
				p.utf8buf[0] = b
				p.utf8len = 1
				continue
			}
		}

		switch p.state {
		case stGround:
			p.processGround(b)
		case stEscape:
			p.processEscape(b)
		case stEscInter:
			p.processEscInter(b)
		case stCsiEntry:
			p.processCsiEntry(b)
		case stCsiParam:
			p.processCsiParam(b)
		case stCsiIgnore:
			p.processCsiIgnore(b)
		case stOscString:
			p.processOsc(b)
		}
	}
}

func (p *AnsiParser) resetParams() {
	p.params = p.params[:0]
	p.cur = 0
	p.hasDigit = false
	p.inSub = false
	p.subMajor = 0
	p.inter = 0
}

// pushParam finalises the current accumulator and appends to params.
func (p *AnsiParser) pushParam() {
	if len(p.params) >= maxParams {
		// Silently discard
		p.cur = 0
		p.hasDigit = false
		p.inSub = false
		return
	}
	if p.inSub {
		p.params = append(p.params, -(p.subMajor*100 + p.cur))
		p.inSub = false
	} else if p.hasDigit {
		p.params = append(p.params, p.cur)
	} else {
		p.params = append(p.params, 0)
	}
	p.cur = 0
	p.hasDigit = false
}

func (p *AnsiParser) dispatchCsi(final byte) {
	// Finalise last parameter
	if p.hasDigit || p.inSub || len(p.params) > 0 {
		p.pushParam()
	}
	var params []int
	if len(p.params) > 0 {
		params = p.params
	}
	p.handler.CsiDispatch(params, p.inter, final)
	p.state = stGround
	p.resetParams()
}

// ---- State processors -------------------------------------------------------

func (p *AnsiParser) processGround(b byte) {
	switch {
	case b <= 0x17 || b == 0x19 || (b >= 0x1C && b <= 0x1F):
		p.handler.Execute(b)
	case b == 0x1B:
		p.resetParams()
		p.state = stEscape
	case b >= 0x20 && b <= 0x7E:
		p.handler.Print(rune(b))
	case b == 0x7F:
		p.handler.Execute(b)
	}
}

func (p *AnsiParser) processEscape(b byte) {
	switch {
	case b <= 0x17 || b == 0x19 || (b >= 0x1C && b <= 0x1F):
		p.handler.Execute(b)
	case b == 0x1B:
		// Reset and stay in escape
		p.resetParams()
	case b >= 0x20 && b <= 0x2F:
		p.inter = b
		p.state = stEscInter
	case (b >= 0x30 && b <= 0x4F) ||
		(b >= 0x51 && b <= 0x57) ||
		(b >= 0x59 && b <= 0x5A) ||
		b == 0x5C ||
		(b >= 0x60 && b <= 0x7E):
		p.handler.EscDispatch(0, b)
		p.state = stGround
		p.resetParams()
	case b == 0x5B: // '['
		p.resetParams()
		p.state = stCsiEntry
	case b == 0x5D: // ']'
		p.oscBuf.Reset()
		p.state = stOscString
	case b == 0x7F:
		// Ignore
	}
}

func (p *AnsiParser) processEscInter(b byte) {
	switch {
	case b >= 0x20 && b <= 0x2F:
		p.inter = b
	case b >= 0x30 && b <= 0x7E:
		p.handler.EscDispatch(p.inter, b)
		p.state = stGround
		p.resetParams()
	case b == 0x7F:
		// Ignore
	}
}

func (p *AnsiParser) processCsiEntry(b byte) {
	switch {
	case b <= 0x17 || b == 0x19 || (b >= 0x1C && b <= 0x1F):
		p.handler.Execute(b)
	case b >= 0x20 && b <= 0x2F:
		p.inter = b
		p.state = stCsiParam
	case (b >= 0x30 && b <= 0x38) || b == 0x3B:
		// Digit or semicolon: transition to CsiParam and process there
		p.state = stCsiParam
		p.processCsiParam(b)
	case b == 0x3A:
		// Sub-parameter separator at entry — unusual but handle it
		p.state = stCsiParam
	case b >= 0x3C && b <= 0x3F:
		// Private marker
		p.inter = b
		p.state = stCsiParam
	case b >= 0x40 && b <= 0x7E:
		p.handler.CsiDispatch(nil, 0, b)
		p.state = stGround
		p.resetParams()
	case b == 0x7F:
		// Ignore
	}
}

func (p *AnsiParser) processCsiParam(b byte) {
	switch {
	case b >= 0x30 && b <= 0x39: // '0'-'9'
		p.cur = p.cur*10 + int(b-'0')
		p.hasDigit = true
	case b == 0x3B: // ';'
		p.pushParam()
	case b == 0x3A: // ':'
		// Sub-parameter separator: save major, reset for minor
		p.subMajor = p.cur
		p.cur = 0
		p.hasDigit = false
		p.inSub = true
	case b >= 0x20 && b <= 0x2F:
		p.inter = b
	case b >= 0x40 && b <= 0x7E:
		p.dispatchCsi(b)
	case b >= 0x3C && b <= 0x3F:
		// Second private marker → ignore sequence
		p.state = stCsiIgnore
	case b == 0x7F:
		// Ignore
	}
}

func (p *AnsiParser) processCsiIgnore(b byte) {
	if b >= 0x40 && b <= 0x7E {
		p.state = stGround
		p.resetParams()
	}
	// All other bytes discarded
}

func (p *AnsiParser) processOsc(b byte) {
	switch b {
	case 0x07: // BEL terminates OSC
		p.dispatchOsc()
		p.state = stGround
	case 0x1B: // ESC — ST (ESC \) will follow; just transition
		p.dispatchOsc()
		p.state = stEscape
		p.resetParams()
	default:
		if b >= 0x20 {
			p.oscBuf.WriteByte(b)
		}
		// 0x00-0x06, 0x08-0x1A: ignore
	}
}

func (p *AnsiParser) dispatchOsc() {
	s := p.oscBuf.String()
	p.oscBuf.Reset()
	cmd := 0
	data := s
	if idx := strings.IndexByte(s, ';'); idx >= 0 {
		left := s[:idx]
		data = s[idx+1:]
		for _, ch := range left {
			if ch >= '0' && ch <= '9' {
				cmd = cmd*10 + int(ch-'0')
			}
		}
	}
	p.handler.OscDispatch(cmd, data)
}

package widgets

import (
	"testing"
)

// ---- test helper ------------------------------------------------------------

type ansiEvent struct {
	kind   string
	r      rune
	code   byte
	params []int
	inter  byte
	final  byte
	cmd    int
	data   string
}

type ansiRecorder struct {
	events []ansiEvent
}

func (h *ansiRecorder) Print(r rune) {
	h.events = append(h.events, ansiEvent{kind: "print", r: r})
}

func (h *ansiRecorder) Execute(code byte) {
	h.events = append(h.events, ansiEvent{kind: "execute", code: code})
}

func (h *ansiRecorder) CsiDispatch(params []int, inter, final byte) {
	var cp []int
	if params != nil {
		cp = make([]int, len(params))
		copy(cp, params)
	}
	h.events = append(h.events, ansiEvent{kind: "csi", params: cp, inter: inter, final: final})
}

func (h *ansiRecorder) OscDispatch(cmd int, data string) {
	h.events = append(h.events, ansiEvent{kind: "osc", cmd: cmd, data: data})
}

func (h *ansiRecorder) EscDispatch(inter, final byte) {
	h.events = append(h.events, ansiEvent{kind: "esc", inter: inter, final: final})
}

func newParser() (*AnsiParser, *ansiRecorder) {
	h := &ansiRecorder{}
	return NewAnsiParser(h), h
}

func feed(p *AnsiParser, s string) { p.Feed([]byte(s)) }

// ---- ground state -----------------------------------------------------------

func TestAnsi_PrintableASCII(t *testing.T) {
	p, h := newParser()
	feed(p, "Hello")
	if len(h.events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(h.events))
	}
	for i, ch := range "Hello" {
		if h.events[i].kind != "print" || h.events[i].r != ch {
			t.Errorf("[%d] expected print(%q), got %+v", i, ch, h.events[i])
		}
	}
}

func TestAnsi_C0ControlCodes(t *testing.T) {
	cases := []byte{0x00, 0x08, 0x09, 0x0A, 0x0D, 0x17, 0x19, 0x1C, 0x1F}
	for _, b := range cases {
		p, h := newParser()
		p.Feed([]byte{b})
		if len(h.events) != 1 || h.events[0].kind != "execute" || h.events[0].code != b {
			t.Errorf("byte 0x%02X: expected execute, got %+v", b, h.events)
		}
	}
}

func TestAnsi_DEL(t *testing.T) {
	p, h := newParser()
	p.Feed([]byte{0x7F})
	if len(h.events) != 1 || h.events[0].kind != "execute" || h.events[0].code != 0x7F {
		t.Errorf("expected execute(DEL), got %+v", h.events)
	}
}

func TestAnsi_0x18_0x1A_NotExecuted(t *testing.T) {
	// 0x18 and 0x1A are not in the C0 execute set for the ground state
	for _, b := range []byte{0x18, 0x1A} {
		p, h := newParser()
		p.Feed([]byte{b})
		if len(h.events) != 0 {
			t.Errorf("byte 0x%02X: expected no event, got %+v", b, h.events)
		}
	}
}

// ---- UTF-8 ------------------------------------------------------------------

func TestAnsi_UTF8_TwoByteRune(t *testing.T) {
	p, h := newParser()
	feed(p, "é") // U+00E9: 0xC3 0xA9
	if len(h.events) != 1 || h.events[0].kind != "print" || h.events[0].r != 'é' {
		t.Errorf("expected print(é), got %+v", h.events)
	}
}

func TestAnsi_UTF8_ThreeByteRune(t *testing.T) {
	p, h := newParser()
	feed(p, "中") // U+4E2D
	if len(h.events) != 1 || h.events[0].kind != "print" || h.events[0].r != '中' {
		t.Errorf("expected print(中), got %+v", h.events)
	}
}

func TestAnsi_UTF8_FourByteRune(t *testing.T) {
	p, h := newParser()
	feed(p, "😀") // U+1F600: 0xF0 0x9F 0x98 0x80
	if len(h.events) != 1 || h.events[0].kind != "print" || h.events[0].r != '😀' {
		t.Errorf("expected print(😀), got %+v", h.events)
	}
}

func TestAnsi_UTF8_SplitAcrossFeeds(t *testing.T) {
	p, h := newParser()
	b := []byte("é") // 0xC3 0xA9
	p.Feed(b[:1])
	if len(h.events) != 0 {
		t.Fatalf("expected no event after lead byte, got %+v", h.events)
	}
	p.Feed(b[1:])
	if len(h.events) != 1 || h.events[0].kind != "print" || h.events[0].r != 'é' {
		t.Errorf("expected print(é) after continuation byte, got %+v", h.events)
	}
}

func TestAnsi_UTF8_InvalidContinuation_Discarded(t *testing.T) {
	p, h := newParser()
	// 0xC3 (lead) followed by 0x41 ('A', not a continuation) → discard partial, print 'A'
	p.Feed([]byte{0xC3, 0x41})
	if len(h.events) != 1 || h.events[0].kind != "print" || h.events[0].r != 'A' {
		t.Errorf("expected print('A') after discarding bad lead, got %+v", h.events)
	}
}

// ---- ESC sequences ----------------------------------------------------------

func TestAnsi_EscSimple(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1BA") // ESC A (0x41 is in 0x30-0x4F)
	if len(h.events) != 1 || h.events[0].kind != "esc" || h.events[0].inter != 0 || h.events[0].final != 'A' {
		t.Errorf("expected EscDispatch(0, 'A'), got %+v", h.events)
	}
}

func TestAnsi_EscWithIntermediate(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B(B") // ESC 0x28 0x42
	if len(h.events) != 1 || h.events[0].kind != "esc" || h.events[0].inter != '(' || h.events[0].final != 'B' {
		t.Errorf("expected EscDispatch('(', 'B'), got %+v", h.events)
	}
}

func TestAnsi_DoubleEsc_Resets(t *testing.T) {
	// ESC ESC A → the second ESC resets; then 'A' dispatches
	p, h := newParser()
	feed(p, "\x1B\x1BA")
	if len(h.events) != 1 || h.events[0].kind != "esc" || h.events[0].final != 'A' {
		t.Errorf("expected single EscDispatch after double ESC, got %+v", h.events)
	}
}

// ---- CSI sequences ----------------------------------------------------------

func TestAnsi_CsiNoParams(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B[A") // cursor up, no params
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(h.events), h.events)
	}
	e := h.events[0]
	if e.kind != "csi" || e.params != nil || e.inter != 0 || e.final != 'A' {
		t.Errorf("expected CsiDispatch(nil, 0, 'A'), got %+v", e)
	}
}

func TestAnsi_CsiSingleParam(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B[5A") // cursor up 5
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	e := h.events[0]
	if e.kind != "csi" || len(e.params) != 1 || e.params[0] != 5 || e.final != 'A' {
		t.Errorf("expected CsiDispatch([5], 0, 'A'), got %+v", e)
	}
}

func TestAnsi_CsiMultipleParams(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B[1;32m") // SGR bold green
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	e := h.events[0]
	if e.kind != "csi" || len(e.params) != 2 || e.params[0] != 1 || e.params[1] != 32 || e.final != 'm' {
		t.Errorf("expected CsiDispatch([1,32], 0, 'm'), got %+v", e)
	}
}

func TestAnsi_CsiSubParameter(t *testing.T) {
	// ESC [ 4:2 m — underline style; sub-param encodes as -(4*100+2) = -402
	p, h := newParser()
	feed(p, "\x1B[4:2m")
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	e := h.events[0]
	if e.kind != "csi" || len(e.params) != 1 || e.params[0] != -402 || e.final != 'm' {
		t.Errorf("expected CsiDispatch([-402], 0, 'm'), got %+v", e)
	}
}

func TestAnsi_CsiEmptyParams_DefaultsToZero(t *testing.T) {
	// ESC [ ; m → two params, both default to 0
	p, h := newParser()
	feed(p, "\x1B[;m")
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	e := h.events[0]
	if e.kind != "csi" || len(e.params) != 2 || e.params[0] != 0 || e.params[1] != 0 || e.final != 'm' {
		t.Errorf("expected CsiDispatch([0,0], 0, 'm'), got %+v", e)
	}
}

func TestAnsi_CsiPrivateMarker(t *testing.T) {
	// ESC [ ? 25 h — show cursor
	p, h := newParser()
	feed(p, "\x1B[?25h")
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	e := h.events[0]
	if e.kind != "csi" || len(e.params) != 1 || e.params[0] != 25 || e.inter != '?' || e.final != 'h' {
		t.Errorf("expected CsiDispatch([25], '?', 'h'), got %+v", e)
	}
}

func TestAnsi_CsiIgnore_DoublePrivateMarker(t *testing.T) {
	// A second private-marker byte in CsiParam puts the parser into CsiIgnore;
	// the final byte resets to ground without dispatching.
	p, h := newParser()
	feed(p, "\x1B[?<25h")
	if len(h.events) != 0 {
		t.Errorf("expected no event (sequence ignored), got %+v", h.events)
	}
	// Parser should be back in ground: plain text works again
	feed(p, "X")
	if len(h.events) != 1 || h.events[0].kind != "print" || h.events[0].r != 'X' {
		t.Errorf("expected print('X') after ignored CSI, got %+v", h.events)
	}
}

func TestAnsi_CsiSplitAcrossFeeds(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B[1;")
	if len(h.events) != 0 {
		t.Fatalf("expected no event mid-sequence, got %+v", h.events)
	}
	feed(p, "32m")
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	e := h.events[0]
	if e.kind != "csi" || len(e.params) != 2 || e.params[0] != 1 || e.params[1] != 32 || e.final != 'm' {
		t.Errorf("expected CsiDispatch([1,32], 0, 'm'), got %+v", e)
	}
}

func TestAnsi_CsiMaxParams_Overflow(t *testing.T) {
	// maxParams = 16; feeding 20 semicolons should not panic and should clamp at 16
	p, h := newParser()
	seq := "\x1B["
	for i := 0; i < 20; i++ {
		seq += "1;"
	}
	seq += "m"
	feed(p, seq)
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	if len(h.events[0].params) > maxParams {
		t.Errorf("params overflow: got %d, max %d", len(h.events[0].params), maxParams)
	}
}

// ---- OSC sequences ----------------------------------------------------------

func TestAnsi_OscBelTerminator(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B]0;window title\x07")
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(h.events), h.events)
	}
	e := h.events[0]
	if e.kind != "osc" || e.cmd != 0 || e.data != "window title" {
		t.Errorf("expected OscDispatch(0, 'window title'), got %+v", e)
	}
}

func TestAnsi_OscStTerminator(t *testing.T) {
	// ST = ESC '\' — the ESC triggers OscDispatch; the trailing '\' (0x5C) is
	// then dispatched as a standalone EscDispatch(0, 0x5C), so two events total.
	p, h := newParser()
	feed(p, "\x1B]2;my title\x1B\\")
	if len(h.events) != 2 {
		t.Fatalf("expected 2 events (osc + esc), got %d: %+v", len(h.events), h.events)
	}
	osc := h.events[0]
	if osc.kind != "osc" || osc.cmd != 2 || osc.data != "my title" {
		t.Errorf("expected OscDispatch(2, 'my title'), got %+v", osc)
	}
	esc := h.events[1]
	if esc.kind != "esc" || esc.inter != 0 || esc.final != 0x5C {
		t.Errorf("expected EscDispatch(0, 0x5C), got %+v", esc)
	}
}

func TestAnsi_OscNoSemicolon(t *testing.T) {
	// No ';' in the OSC string → cmd=0, data=full string
	p, h := newParser()
	feed(p, "\x1B]hello\x07")
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(h.events), h.events)
	}
	e := h.events[0]
	if e.kind != "osc" || e.cmd != 0 || e.data != "hello" {
		t.Errorf("expected OscDispatch(0, 'hello'), got %+v", e)
	}
}

func TestAnsi_OscMultiDigitCmd(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B]133;A\x07")
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	e := h.events[0]
	if e.kind != "osc" || e.cmd != 133 || e.data != "A" {
		t.Errorf("expected OscDispatch(133, 'A'), got %+v", e)
	}
}

func TestAnsi_OscEmptyData(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B]0;\x07")
	if len(h.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(h.events))
	}
	e := h.events[0]
	if e.kind != "osc" || e.cmd != 0 || e.data != "" {
		t.Errorf("expected OscDispatch(0, ''), got %+v", e)
	}
}

// ---- mixed sequences --------------------------------------------------------

func TestAnsi_TextAroundSequences(t *testing.T) {
	p, h := newParser()
	feed(p, "Hi\x1B[32mThere")
	// H, i, CsiDispatch, T, h, e, r, e
	if len(h.events) != 8 {
		t.Fatalf("expected 8 events, got %d: %+v", len(h.events), h.events)
	}
	if h.events[0].kind != "print" || h.events[0].r != 'H' {
		t.Errorf("[0] expected print('H'), got %+v", h.events[0])
	}
	if h.events[2].kind != "csi" || h.events[2].final != 'm' {
		t.Errorf("[2] expected csi 'm', got %+v", h.events[2])
	}
	if h.events[3].kind != "print" || h.events[3].r != 'T' {
		t.Errorf("[3] expected print('T'), got %+v", h.events[3])
	}
}

func TestAnsi_SequentialSequences(t *testing.T) {
	p, h := newParser()
	feed(p, "\x1B[1m\x1B[0m")
	if len(h.events) != 2 {
		t.Fatalf("expected 2 events, got %d: %+v", len(h.events), h.events)
	}
	if h.events[0].kind != "csi" || h.events[0].params[0] != 1 {
		t.Errorf("[0] expected CsiDispatch([1], …), got %+v", h.events[0])
	}
	if h.events[1].kind != "csi" || h.events[1].params[0] != 0 {
		t.Errorf("[1] expected CsiDispatch([0], …), got %+v", h.events[1])
	}
}

func TestAnsi_EmptyFeed(t *testing.T) {
	p, h := newParser()
	p.Feed([]byte{})
	if len(h.events) != 0 {
		t.Errorf("expected no events for empty feed, got %+v", h.events)
	}
}

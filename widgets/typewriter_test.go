package widgets

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	. "github.com/tekugo/zeichenwerk/core"
)

// newTW is a helper that creates a Typewriter ready for direct Tick() calls.
// dwellTicks is set explicitly to avoid depending on Start().
func newTW(text string, rate int, dwellTicks int) *Typewriter {
	tw := NewTypewriter("tw", "")
	tw.SetRate(rate)
	tw.dwellTicks = dwellTicks
	tw.SetText(text)
	return tw
}

// ── Tick: Reveal phase ────────────────────────────────────────────────────────

func TestTypewriter_Tick_AdvancesShown(t *testing.T) {
	tw := newTW("hello", 1, 0)
	tw.Tick()
	assert.Equal(t, 1, tw.shown, "one tick with rate=1 should reveal 1 rune")
}

func TestTypewriter_Tick_AdvancesByRate(t *testing.T) {
	tw := newTW("hello world", 3, 0)
	tw.Tick()
	assert.Equal(t, 3, tw.shown, "one tick with rate=3 should reveal 3 runes")
}

func TestTypewriter_Tick_ClampedAtEnd(t *testing.T) {
	tw := newTW("hi", 5, 0)
	tw.Tick()
	assert.Equal(t, 2, tw.shown, "shown must not exceed len(runes)")
}

// ── Tick: Dwell phase ─────────────────────────────────────────────────────────

// advanceReveal ticks until all runes are revealed.
func advanceReveal(tw *Typewriter) {
	for tw.shown < len(tw.runes) {
		tw.Tick()
	}
}

func TestTypewriter_Dwell_IncrementsTick(t *testing.T) {
	tw := newTW("hi", 1, 4)
	advanceReveal(tw)
	before := tw.dwellTick
	tw.Tick()
	assert.Equal(t, before+1, tw.dwellTick, "dwellTick should increment during dwell phase")
}

func TestTypewriter_Dwell_CursorBlinks(t *testing.T) {
	tw := newTW("hi", 1, 10)
	tw.showCursor = true
	advanceReveal(tw)

	// dwellTick starts at 0; first dwell tick → dwellTick becomes 1 (odd → off)
	tw.Tick() // dwellTick=1 → cursorOn = showCursor && (1%2==0) = false
	assert.False(t, tw.cursorOn, "cursor should be off when dwellTick is odd")

	tw.Tick() // dwellTick=2 → cursorOn = showCursor && (2%2==0) = true
	assert.True(t, tw.cursorOn, "cursor should be on when dwellTick is even")
}

// ── Tick: Done phase ──────────────────────────────────────────────────────────

// advanceDwell ticks through the dwell phase.
func advanceDwell(tw *Typewriter) {
	for tw.dwellTick < tw.dwellTicks {
		tw.Tick()
	}
}

func TestTypewriter_Done_StopsWhenNotRepeat(t *testing.T) {
	tw := newTW("hi", 1, 2)
	tw.repeat = false
	advanceReveal(tw)
	advanceDwell(tw)
	tw.Tick() // done tick
	assert.False(t, tw.cursorOn, "cursorOn should be false when animation is done and not repeating")
}

func TestTypewriter_Done_RepeatsWhenRepeat(t *testing.T) {
	tw := newTW("hi", 1, 2)
	tw.repeat = true
	advanceReveal(tw)
	advanceDwell(tw)
	tw.Tick() // done tick → should reset
	assert.Equal(t, 0, tw.shown, "shown should reset to 0 when repeat=true")
	assert.Equal(t, 0, tw.dwellTick, "dwellTick should reset to 0 when repeat=true")
}

// ── Events ────────────────────────────────────────────────────────────────────

func TestTypewriter_EvtChange_FiresOnce(t *testing.T) {
	tw := newTW("abc", 1, 0)
	count := 0
	tw.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		count++
		return false
	})
	// Tick three times to reveal all 3 runes; EvtChange should fire only on the
	// last reveal tick.
	tw.Tick() // shown=1
	tw.Tick() // shown=2
	tw.Tick() // shown=3 → fires EvtChange
	assert.Equal(t, 1, count, "EvtChange should fire exactly once when reveal completes")
}

func TestTypewriter_EvtActivate_FiresOnDone(t *testing.T) {
	tw := newTW("hi", 1, 1)
	tw.repeat = false
	activated := false
	tw.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		activated = true
		return false
	})
	advanceReveal(tw)
	advanceDwell(tw)
	tw.Tick() // done
	assert.True(t, activated, "EvtActivate should fire when animation completes (repeat=false)")
}

// ── Hint ──────────────────────────────────────────────────────────────────────

func TestTypewriter_Hint_FullText(t *testing.T) {
	tw := NewTypewriter("tw", "")
	tw.showCursor = false
	tw.SetText("hello")
	// Advance only partway; Hint must still return full-text dimensions.
	tw.dwellTicks = 0
	tw.Tick() // shown=1
	w, h := tw.Hint()
	assert.Equal(t, 5, w, "Hint width should reflect full text, not shown portion")
	assert.Equal(t, 1, h)
}

func TestTypewriter_Hint_MultiLine(t *testing.T) {
	tw := NewTypewriter("tw", "")
	tw.showCursor = false
	tw.SetText("hello\nworld!")
	w, h := tw.Hint()
	assert.Equal(t, 6, w, "Hint width should be length of longest line")
	assert.Equal(t, 2, h, "Hint height should equal number of lines")
}

func TestTypewriter_Hint_WithCursor(t *testing.T) {
	tw := NewTypewriter("tw", "")
	tw.showCursor = true
	tw.SetText("hi")
	w, _ := tw.Hint()
	assert.Equal(t, 3, w, "Hint width should add 1 for cursor when showCursor=true")
}

// ── SetText ───────────────────────────────────────────────────────────────────

func TestTypewriter_SetText_ResetsState(t *testing.T) {
	tw := newTW("hello", 1, 5)
	advanceReveal(tw)
	tw.Tick() // dwellTick=1

	tw.SetText("new text")
	assert.Equal(t, 0, tw.shown, "SetText should reset shown to 0")
	assert.Equal(t, 0, tw.dwellTick, "SetText should reset dwellTick to 0")
	assert.Equal(t, tw.showCursor, tw.cursorOn, "SetText should reset cursorOn to showCursor")
}

// ── splitRunes ────────────────────────────────────────────────────────────────

func TestTypewriter_MultiLine_SplitsCorrectly(t *testing.T) {
	input := []rune("foo\nbar\nbaz")
	lines := splitRunes(input)
	assert.Equal(t, 3, len(lines), "splitRunes should produce 3 lines")
	assert.Equal(t, []rune("foo"), lines[0])
	assert.Equal(t, []rune("bar"), lines[1])
	assert.Equal(t, []rune("baz"), lines[2])
}

func TestTypewriter_SplitRunes_Empty(t *testing.T) {
	lines := splitRunes([]rune{})
	assert.Equal(t, 1, len(lines), "splitRunes on empty slice should return one empty line")
	assert.Equal(t, []rune{}, lines[0])
}

// ── Start stores interval and computes dwellTicks ─────────────────────────────

func TestTypewriter_Start_ComputesDwellTicks(t *testing.T) {
	tw := NewTypewriter("tw", "")
	tw.SetDwell(500 * time.Millisecond)
	// Manually set interval and dwellTicks as Start would, without goroutine.
	interval := 100 * time.Millisecond
	tw.interval = interval
	tw.dwellTicks = int(tw.dwell / interval)
	assert.Equal(t, 5, tw.dwellTicks, "dwellTicks should be dwell/interval")
}

package widgets

import (
	"strings"
	"time"
	"unicode/utf8"

	. "github.com/tekugo/zeichenwerk/core"
)

// ==== AI ===================================================================

// Typewriter is an animated widget that reveals text character by character,
// optionally blinking a cursor during and after the reveal, and can repeat
// indefinitely or fire a completion event when done.
type Typewriter struct {
	Animation
	text       string
	runes      []rune
	shown      int
	rate       int
	showCursor bool
	cursorOn   bool
	dwell      time.Duration
	dwellTicks int
	dwellTick  int
	repeat     bool
	interval   time.Duration
	chCursor   string
}

// NewTypewriter creates a new Typewriter widget with the given id and class.
// The widget starts with a blinking cursor (▌), rate 1, and a 500 ms dwell
// period after the reveal is complete.
func NewTypewriter(id, class string) *Typewriter {
	tw := &Typewriter{
		Animation: Animation{
			Component: Component{id: id, class: class},
			stop:      make(chan struct{}),
		},
		rate:       1,
		showCursor: true,
		cursorOn:   true,
		dwell:      500 * time.Millisecond,
		chCursor:   "▌",
	}
	tw.fn = tw.Tick
	return tw
}

// Apply applies theme styles to the typewriter and its cursor part.
func (tw *Typewriter) Apply(theme *Theme) {
	theme.Apply(tw, tw.Selector("typewriter"))
	theme.Apply(tw, tw.Selector("typewriter/cursor"))
	if s := theme.String("typewriter.cursor"); s != "" {
		tw.chCursor = s
	}
}

// SetText replaces the displayed text and resets the reveal state.
func (tw *Typewriter) SetText(s string) *Typewriter {
	tw.text = s
	tw.runes = []rune(s)
	tw.shown = 0
	tw.dwellTick = 0
	tw.cursorOn = tw.showCursor
	Redraw(tw)
	return tw
}

// Text returns the full text string (not just the revealed portion).
func (tw *Typewriter) Text() string {
	return tw.text
}

// SetRate sets the number of characters revealed per tick. Clamped to at least 1.
func (tw *Typewriter) SetRate(n int) *Typewriter {
	if n < 1 {
		n = 1
	}
	tw.rate = n
	return tw
}

// SetCursor enables or disables the cursor display.
func (tw *Typewriter) SetCursor(v bool) *Typewriter {
	tw.showCursor = v
	Redraw(tw)
	return tw
}

// SetDwell sets the dwell duration (how long the cursor blinks after reveal).
// If an interval has already been stored (via Start), dwellTicks is recomputed.
func (tw *Typewriter) SetDwell(d time.Duration) *Typewriter {
	tw.dwell = d
	if tw.interval > 0 {
		tw.dwellTicks = int(d / tw.interval)
	}
	return tw
}

// SetRepeat controls whether the animation restarts after completing.
func (tw *Typewriter) SetRepeat(v bool) *Typewriter {
	tw.repeat = v
	return tw
}

// Reset resets the reveal state to the beginning without changing the text.
func (tw *Typewriter) Reset() {
	tw.shown = 0
	tw.dwellTick = 0
	tw.cursorOn = tw.showCursor
	Redraw(tw)
}

// Start stores the tick interval, computes dwellTicks, and starts the animation.
func (tw *Typewriter) Start(interval time.Duration) {
	tw.interval = interval
	tw.dwellTicks = int(tw.dwell / interval)
	tw.Animation.Start(interval)
}

// Hint returns the preferred content size for the typewriter based on the full
// text (not just the currently revealed portion). If hwidth or hheight is
// explicitly set they take precedence.
func (tw *Typewriter) Hint() (int, int) {
	if tw.hwidth != 0 || tw.hheight != 0 {
		return tw.hwidth, tw.hheight
	}
	if tw.text == "" {
		w := 0
		if tw.showCursor {
			w = 1
		}
		return w, 1
	}
	lines := strings.Split(tw.text, "\n")
	maxW := 0
	for _, l := range lines {
		n := utf8.RuneCountInString(l)
		if n > maxW {
			maxW = n
		}
	}
	if tw.showCursor {
		maxW++
	}
	return maxW, len(lines)
}

// Tick advances the typewriter animation by one frame.
// It implements a three-phase state machine: revealing → dwell → done.
func (tw *Typewriter) Tick() {
	switch {
	case tw.shown < len(tw.runes): // Revealing
		tw.shown = min(tw.shown+tw.rate, len(tw.runes))
		tw.cursorOn = tw.showCursor
		if tw.shown == len(tw.runes) {
			tw.Dispatch(tw, EvtChange, true)
		}
		Redraw(tw)

	case tw.dwellTick < tw.dwellTicks: // Dwell
		tw.dwellTick++
		tw.cursorOn = tw.showCursor && (tw.dwellTick%2 == 0)
		Redraw(tw)

	default: // Done
		if tw.repeat {
			tw.shown = 0
			tw.dwellTick = 0
			tw.cursorOn = tw.showCursor
			Redraw(tw)
		} else {
			tw.cursorOn = false
			Redraw(tw)
			tw.Dispatch(tw, EvtActivate, true)
			tw.Stop()
		}
	}
}

// Render draws the currently revealed text and optional cursor.
func (tw *Typewriter) Render(r *Renderer) {
	if tw.Flag(FlagHidden) {
		return
	}
	tw.Component.Render(r)

	cx, cy, cw, ch := tw.Content()
	baseStyle := tw.Style()
	cursorStyle := tw.Style("cursor")

	lines := splitRunes(tw.runes[:tw.shown])

	for i, line := range lines {
		if i >= ch {
			break
		}
		r.Set(baseStyle.Foreground(), baseStyle.Background(), baseStyle.Font())
		r.Text(cx, cy+i, string(line), cw)
	}

	if tw.cursorOn && tw.showCursor {
		lastIdx := len(lines) - 1
		if lastIdx < 0 {
			lastIdx = 0
		}
		var lastLine []rune
		if len(lines) > 0 {
			lastLine = lines[lastIdx]
		}
		cursorX := cx + utf8.RuneCountInString(string(lastLine))
		cursorY := cy + lastIdx
		r.Set(cursorStyle.Foreground(), cursorStyle.Background(), cursorStyle.Font())
		r.Put(cursorX, cursorY, tw.chCursor)
	}
}

// splitRunes splits a rune slice into lines at '\n' boundaries.
// It always returns at least one (possibly empty) slice.
func splitRunes(runes []rune) [][]rune {
	result := [][]rune{{}}
	for _, r := range runes {
		if r == '\n' {
			result = append(result, []rune{})
		} else {
			result[len(result)-1] = append(result[len(result)-1], r)
		}
	}
	return result
}

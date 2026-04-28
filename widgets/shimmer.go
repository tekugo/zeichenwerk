package widgets

import (
	"math"
	"strings"

	"github.com/rivo/uniseg"
	. "github.com/tekugo/zeichenwerk/core"
)

// ==== AI ===================================================================

// Shimmer is an Animation-driven widget that displays text with a highlight
// band sweeping continuously from left to right. Characters inside the band
// are blended toward an accent colour; characters outside use the base style.
// The band advances one column per tick and wraps seamlessly. Calling Stop()
// freezes the text in its base style.
//
// With SetGradient(true), the band edges follow a smooth cosine curve instead
// of the stepped linear falloff, producing a softer, more organic glow.
//
// The Shimmer dispatches no events — it is a pure display widget.
type Shimmer struct {
	Animation
	text      string   // displayed text; may contain \n
	lines     []string // text split at \n (updated in SetText)
	maxWidth  int      // display-column width of the longest line
	bandPos   int      // leftmost column of the band [0, maxWidth)
	bandWidth int      // number of columns in the bright core of the band
	edgeWidth int      // columns of gradient fade on each side of the band
	gradient  bool     // when true, use a smooth cosine intensity curve
}

// NewShimmer creates a Shimmer widget with the given id and class.
// Defaults: bandWidth=6, edgeWidth=3. The animation is not started automatically.
func NewShimmer(id, class string) *Shimmer {
	sh := &Shimmer{
		Animation: Animation{
			Component: Component{id: id, class: class},
			stop:      make(chan struct{}),
		},
		bandWidth: 6,
		edgeWidth: 3,
	}
	sh.fn = sh.Tick
	return sh
}

// Apply applies theme styles to the shimmer widget.
func (sh *Shimmer) Apply(theme *Theme) {
	theme.Apply(sh, sh.Selector("shimmer"))
	theme.Apply(sh, sh.Selector("shimmer/band"))
}

// SetText replaces the displayed text, splits it into lines, recomputes
// maxWidth, resets bandPos to 0, and triggers a redraw.
func (sh *Shimmer) SetText(s string) *Shimmer {
	sh.text = s
	sh.lines = strings.Split(s, "\n")
	sh.maxWidth = 0
	for _, line := range sh.lines {
		w := uniseg.StringWidth(line)
		if w > sh.maxWidth {
			sh.maxWidth = w
		}
	}
	sh.bandPos = 0
	Redraw(sh)
	return sh
}

// Text returns the current text.
func (sh *Shimmer) Text() string { return sh.text }

// SetBandWidth sets the core highlight width in columns. Clamped to minimum 1.
func (sh *Shimmer) SetBandWidth(n int) *Shimmer {
	if n < 1 {
		n = 1
	}
	sh.bandWidth = n
	return sh
}

// SetEdgeWidth sets the gradient columns on each side of the core band.
// 0 produces a hard edge with no blending.
func (sh *Shimmer) SetEdgeWidth(n int) *Shimmer {
	if n < 0 {
		n = 0
	}
	sh.edgeWidth = n
	return sh
}

// SetGradient enables or disables smooth cosine blending. When true, the
// intensity follows a cosine curve across the full band+edge span, giving a
// soft, organic glow. When false (the default), the stepped linear falloff
// is used, which gives a crisper highlight.
func (sh *Shimmer) SetGradient(on bool) *Shimmer {
	sh.gradient = on
	return sh
}

// Tick advances the band position by one column and triggers a redraw.
// It is a no-op when maxWidth is 0.
func (sh *Shimmer) Tick() {
	if sh.maxWidth == 0 {
		return
	}
	sh.bandPos = (sh.bandPos + 1) % sh.maxWidth
	Redraw(sh)
}

// Hint returns the preferred size: width = maxWidth + horizontal style overhead,
// height = line count + vertical style overhead.
func (sh *Shimmer) Hint() (int, int) {
	if sh.hwidth != 0 || sh.hheight != 0 {
		return sh.hwidth, sh.hheight
	}
	s := sh.Style()
	lineCount := len(sh.lines)
	if lineCount == 0 {
		lineCount = 1
	}
	return sh.maxWidth + s.Horizontal(), lineCount + s.Vertical()
}

// Render draws the shimmer widget to the screen.
func (sh *Shimmer) Render(r *Renderer) {
	if sh.Flag(FlagHidden) {
		return
	}
	sh.Component.Render(r)

	cx, cy, cw, ch := sh.Content()
	if cw == 0 || ch == 0 {
		return
	}

	baseStyle := sh.Style("")
	bandStyle := sh.Style("band")

	baseFg := r.Theme.Color(baseStyle.Foreground())
	bandFg := r.Theme.Color(bandStyle.Foreground())

	for i, line := range sh.lines {
		if i >= ch {
			break
		}

		// Build a per-column rune slice for this line.
		type cell struct {
			ch    string
			width int
		}
		cells := make([]cell, 0, len(line))
		for _, ru := range line {
			w := uniseg.StringWidth(string(ru))
			if w < 1 {
				w = 1
			}
			cells = append(cells, cell{ch: string(ru), width: w})
			if w == 2 {
				cells = append(cells, cell{ch: "", width: 0}) // wide sentinel
			}
		}

		col := 0
		for col < cw {
			// Determine the character to draw.
			var char string
			if col < len(cells) {
				c := cells[col]
				if c.width == 0 {
					// Second cell of a wide rune — draw a space.
					char = " "
				} else {
					char = c.ch
				}
			} else {
				char = " "
			}

			intensity := sh.intensity(col)
			fg := LerpColor(baseFg, bandFg, intensity)
			r.Set(fg, baseStyle.Background(), baseStyle.Font())
			r.Put(cx+col, cy+i, char)
			col++
		}
	}
}

// intensity returns how much of the band colour to apply at display column c.
func (sh *Shimmer) intensity(c int) float64 {
	if sh.maxWidth == 0 {
		return 0
	}
	// Wrap-aware distance from c to bandPos.
	mw := sh.maxWidth
	d1 := (c - sh.bandPos + mw) % mw
	d2 := (sh.bandPos - c + mw) % mw
	dist := d1
	if d2 < d1 {
		dist = d2
	}

	hw := sh.bandWidth / 2
	ew := sh.edgeWidth
	totalHalf := ew + hw

	if sh.gradient {
		// Smooth cosine bell: 1.0 at the band centre, 0.0 at dist == totalHalf.
		if totalHalf == 0 || dist >= totalHalf {
			return 0
		}
		return (1 + math.Cos(math.Pi*float64(dist)/float64(totalHalf))) / 2
	}

	// Stepped mode: linear edge ramp + flat core.
	switch {
	case ew > 0 && dist <= ew:
		return 1.0 - float64(dist)/float64(ew+1)
	case dist <= totalHalf:
		return 1.0
	default:
		return 0.0
	}
}

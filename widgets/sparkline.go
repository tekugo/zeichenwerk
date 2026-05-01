package widgets

import (
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// sparklineBlocks is the ordered set of Unicode block-fill characters used to
// draw bar columns, from one-eighth height (▁) to full height (█).
var sparklineBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// FloatSlice adapts a []float64 to DataProvider.
// The slice must be ordered oldest-first (index 0 = oldest, last = newest),
// matching the natural append order. Get(0) returns the newest element.
type FloatSlice []float64

func (f FloatSlice) Size() int         { return len(f) }
func (f FloatSlice) Get(i int) float64 { return f[len(f)-1-i] }

// Sparkline renders a sequence of float64 values as a column of Unicode block
// characters. The widget adapts to any content height: with h rows each column
// has h×8 discrete levels — from a single ▁ at the bottom up to all rows
// filled with █.
type Sparkline struct {
	Component
	provider  DataProvider
	absolute  bool    // false = relative (auto-scale), true = fixed [min, max]
	min       float64 // lower bound for absolute mode
	max       float64 // upper bound for absolute mode
	threshold float64 // dual-colour split point; 0 = disabled
	gradient  bool    // smooth colour interpolation across the threshold range
}

// NewSparkline creates a Sparkline with relative scaling and no initial data.
func NewSparkline(id, class string) *Sparkline {
	return &Sparkline{
		Component: Component{id: id, class: class},
		max:       1,
	}
}

// ---- Widget interface -------------------------------------------------------

// Apply registers the "sparkline" base style and the "sparkline/high" style
// (used for bars whose value is at or above the threshold).
func (s *Sparkline) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("sparkline"))
	theme.Apply(s, s.Selector("sparkline/high"))
}

// Hint returns the preferred (width, height). Width defaults to provider.Size()
// (minimum 1) when not set explicitly; height defaults to 1.
func (s *Sparkline) Hint() (int, int) {
	w, h := s.hwidth, s.hheight
	if w == 0 {
		if s.provider != nil {
			w = s.provider.Size()
		}
		if w == 0 {
			w = 1
		}
	}
	if h == 0 {
		h = 1
	}
	return w, h
}

// Render draws the sparkline into the widget's content area.
func (s *Sparkline) Render(r *Renderer) {
	s.Component.Render(r)

	cx, cy, cw, ch := s.Content()
	if cw <= 0 || ch <= 0 || s.provider == nil {
		return
	}

	lo, hi := s.rangeFor(cw)

	baseStyle := s.Style("")
	highStyle := s.Style("high")

	// Pre-resolve colours and threshold position once — only pay the cost when needed.
	var baseFg, highFg string
	var threshLevel float64
	if s.threshold > 0 && s.gradient {
		baseFg = r.Theme.Color(baseStyle.Foreground())
		highFg = r.Theme.Color(highStyle.Foreground())
		if hi != lo {
			threshLevel = (s.threshold - lo) / (hi - lo)
		}
	}

	n := s.provider.Size()
	for col := 0; col < cw; col++ {
		// provIdx 0 = rightmost (newest); provIdx cw-1 = leftmost (oldest shown).
		provIdx := cw - 1 - col
		if provIdx >= n {
			// Left-pad: columns with no corresponding data point are blank.
			r.Set(baseStyle.Foreground(), baseStyle.Background(), baseStyle.Font())
			for row := 0; row < ch; row++ {
				r.Put(cx+col, cy+row, " ")
			}
			continue
		}

		v := s.provider.Get(provIdx)

		var level float64
		if hi != lo {
			level = (v - lo) / (hi - lo)
		} else {
			level = 0.5
		}
		if level < 0 {
			level = 0
		} else if level > 1 {
			level = 1
		}

		// Map level to a step in [0, h*8-1].
		totalSteps := ch * 8
		step := int(level*float64(totalSteps-1) + 0.5)
		fullRows := step / 8
		partial := step % 8

		// Determine foreground colour for this column.
		var fg string
		switch {
		case s.threshold > 0 && s.gradient:
			// Interpolate from base to high across [threshLevel, 1].
			var t float64
			if remaining := 1 - threshLevel; remaining > 0 && level > threshLevel {
				t = (level - threshLevel) / remaining
				if t > 1 {
					t = 1
				}
			}
			fg = LerpColor(baseFg, highFg, t)
		case s.threshold > 0 && v >= s.threshold:
			fg = r.Theme.Color(highStyle.Foreground())
		default:
			fg = r.Theme.Color(baseStyle.Foreground())
		}
		r.Set(fg, baseStyle.Background(), baseStyle.Font())

		for row := 0; row < ch; row++ {
			rowFromBottom := ch - 1 - row
			var block rune
			switch {
			case rowFromBottom < fullRows:
				block = '█'
			case rowFromBottom == fullRows:
				block = sparklineBlocks[partial]
			default:
				block = ' '
			}
			r.Put(cx+col, cy+row, string(block))
		}
	}
}

// ---- Data methods ----------------------------------------------------------

// SetProvider sets the data source and calls Refresh.
func (s *Sparkline) SetProvider(p DataProvider) {
	s.provider = p
	s.Refresh()
}

// Provider returns the current data source, or nil if none has been set.
func (s *Sparkline) Provider() DataProvider {
	return s.provider
}

// Count returns the number of data points in the current provider, or 0 if
// no provider has been set.
func (s *Sparkline) Count() int {
	if s.provider == nil {
		return 0
	}
	return s.provider.Size()
}

// SetAbsolute switches between relative (false, default) and absolute (true)
// scaling. In absolute mode the [Min, Max] bracket is used directly. In
// relative mode the range is computed from the visible values each render pass.
func (s *Sparkline) SetAbsolute(v bool) {
	s.absolute = v
	s.Refresh()
}

// SetMin sets the lower bound for absolute scaling and calls Refresh.
func (s *Sparkline) SetMin(v float64) {
	s.min = v
	s.Refresh()
}

// SetMax sets the upper bound for absolute scaling and calls Refresh.
func (s *Sparkline) SetMax(v float64) {
	s.max = v
	s.Refresh()
}

// SetThreshold sets the value at or above which bars are drawn with the
// "sparkline/high" style. Pass 0 to disable dual-colour rendering.
func (s *Sparkline) SetThreshold(v float64) {
	s.threshold = v
	s.Refresh()
}

// SetGradient enables or disables smooth colour interpolation across the
// threshold range. When true, bars below the threshold use the base colour and
// bars above it blend linearly from the base colour toward the "sparkline/high"
// colour, reaching the full high colour at the maximum value. Has no effect
// when threshold is 0.
func (s *Sparkline) SetGradient(v bool) {
	s.gradient = v
	s.Refresh()
}

// ---- internal helpers -------------------------------------------------------

// rangeFor returns the [lo, hi] scaling bounds for the visible window.
func (s *Sparkline) rangeFor(cw int) (lo, hi float64) {
	if s.absolute {
		return s.min, s.max
	}
	n := s.provider.Size()
	visible := cw
	if n < visible {
		visible = n
	}
	if visible == 0 {
		return 0, 1
	}
	lo = s.provider.Get(0)
	hi = lo
	for i := 1; i < visible; i++ {
		v := s.provider.Get(i)
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	// All-equal: expand symmetrically so the bars land at mid-height (level=0.5).
	if lo == hi {
		lo -= 1
		hi += 1
	}
	return lo, hi
}

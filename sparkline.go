package zeichenwerk

// sparklineBlocks is the ordered set of Unicode block-fill characters used to
// draw bar columns, from one-eighth height (▁) to full height (█).
var sparklineBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// ScaleMode controls how Sparkline values are mapped to bar heights.
type ScaleMode int

const (
	// Relative rescales every render pass so the tallest visible bar fills
	// the full height. Good for shape comparisons across multiple sparklines.
	Relative ScaleMode = iota
	// Absolute maps values to a fixed [Min, Max] bracket.
	// Good for showing absolute magnitude over time.
	Absolute
)

// Sparkline renders a sequence of float64 values as a column of Unicode block
// characters. The widget adapts to any content height: with h rows each column
// has h×8 discrete levels — from a single ▁ at the bottom up to all rows
// filled with █.
type Sparkline struct {
	Component
	values    []float64
	mode      ScaleMode
	min       float64 // lower bound for Absolute mode
	max       float64 // upper bound for Absolute mode
	threshold float64 // dual-colour split point; 0 = disabled
	gradient  bool    // smooth colour interpolation across the threshold range
	capacity  int     // ring-buffer cap; 0 = unlimited
}

// NewSparkline creates a Sparkline with Relative scaling and no initial data.
func NewSparkline(id, class string) *Sparkline {
	return &Sparkline{
		Component: Component{id: id, class: class},
		mode:      Relative,
		min:       0,
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

// Hint returns the preferred (width, height). Width defaults to len(values)
// (minimum 1) when not set explicitly; height defaults to 1.
func (s *Sparkline) Hint() (int, int) {
	w, h := s.hwidth, s.hheight
	if w == 0 {
		w = len(s.values)
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
	if cw <= 0 || ch <= 0 {
		return
	}

	// Select the rightmost cw values (most recent data).
	vs := s.values
	if len(vs) > cw {
		vs = vs[len(vs)-cw:]
	}

	lo, hi := s.rangeFor(vs)

	baseStyle := s.Style("")
	highStyle := s.Style("high")

	// Pre-resolve colours and threshold position once — only pay the cost when needed.
	var baseFg, highFg string
	var threshLevel float64
	if s.threshold > 0 && s.gradient {
		baseFg = r.theme.Color(baseStyle.Foreground())
		highFg = r.theme.Color(highStyle.Foreground())
		if hi != lo {
			threshLevel = (s.threshold - lo) / (hi - lo)
		}
	}

	for col := 0; col < cw; col++ {
		// Left-pad: columns with no corresponding data point are blank.
		dataIdx := col - (cw - len(vs))
		if dataIdx < 0 {
			r.Set(baseStyle.Foreground(), baseStyle.Background(), baseStyle.Font())
			for row := 0; row < ch; row++ {
				r.Put(cx+col, cy+row, " ")
			}
			continue
		}

		v := vs[dataIdx]

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
			fg = lerpColor(baseFg, highFg, t)
		case s.threshold > 0 && v >= s.threshold:
			fg = r.theme.Color(highStyle.Foreground())
		default:
			fg = r.theme.Color(baseStyle.Foreground())
		}
		r.Set(fg, baseStyle.Background(), baseStyle.Font())

		for row := 0; row < ch; row++ {
			rowFromBottom := ch - 1 - row
			var ch rune
			switch {
			case rowFromBottom < fullRows:
				ch = '█'
			case rowFromBottom == fullRows:
				ch = sparklineBlocks[partial]
			default:
				ch = ' '
			}
			r.Put(cx+col, cy+row, string(ch))
		}
	}
}

// ---- Data methods ----------------------------------------------------------

// Append adds a data point to the end of the series. When a capacity is set
// and the buffer is full, the oldest point is dropped first. Calls Refresh.
func (s *Sparkline) Append(v float64) {
	s.values = append(s.values, v)
	if s.capacity > 0 && len(s.values) > s.capacity {
		s.values = s.values[len(s.values)-s.capacity:]
	}
	s.Refresh()
}

// SetValues replaces the entire data series and calls Refresh.
func (s *Sparkline) SetValues(vs []float64) {
	cp := make([]float64, len(vs))
	copy(cp, vs)
	s.values = cp
	s.Refresh()
}

// Values returns the current data series.
func (s *Sparkline) Values() []float64 {
	return s.values
}

// SetMode sets the scale mode and calls Refresh.
func (s *Sparkline) SetMode(m ScaleMode) {
	s.mode = m
	s.Refresh()
}

// SetMin sets the lower bound for Absolute mode and calls Refresh.
func (s *Sparkline) SetMin(v float64) {
	s.min = v
	s.Refresh()
}

// SetMax sets the upper bound for Absolute mode and calls Refresh.
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

// SetCapacity sets the maximum number of data points retained. Existing data
// beyond the new capacity is trimmed immediately. Calls Refresh.
func (s *Sparkline) SetCapacity(n int) {
	s.capacity = n
	if n > 0 && len(s.values) > n {
		s.values = s.values[len(s.values)-n:]
	}
	s.Refresh()
}

// ---- internal helpers -------------------------------------------------------

// rangeFor returns the [lo, hi] scaling bounds for the given visible slice.
func (s *Sparkline) rangeFor(vs []float64) (lo, hi float64) {
	if s.mode == Absolute {
		return s.min, s.max
	}
	if len(vs) == 0 {
		return 0, 1
	}
	lo, hi = vs[0], vs[0]
	for _, v := range vs[1:] {
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

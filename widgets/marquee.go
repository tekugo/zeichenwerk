package widgets

import (
	"github.com/rivo/uniseg"
	. "github.com/tekugo/zeichenwerk/core"
)

// ==== AI ===================================================================

// runeCol holds a rune and its display-column width, indexed by display column.
// For a wide (2-column) rune the first cell stores the rune; the second cell
// has width == 0 and r == 0 (a sentinel meaning "second half of wide char").
type runeCol struct {
	r     rune
	width int
}

// Marquee is an Animation-driven single-row scrolling text ticker. Text wider
// than the widget scrolls continuously to the left. Scrolling pauses while
// [FlagHovered] is set (the mouse cursor is over the widget).
//
// The Marquee dispatches no events — it is a pure display widget.
type Marquee struct {
	Animation
	text        string
	textWidth   int       // cached display-column width of text
	cols        []runeCol // per-display-column rune index, len == textWidth
	offset      int       // current scroll position in [0, cycle)
	speed       int       // display columns advanced per tick (minimum 1)
	gap         int       // space columns appended after text before loop
	renderWidth int       // content width from the last Render; read by Tick
}

// NewMarquee creates a Marquee widget with the given id and class.
// Defaults: speed=1, gap=4. The animation is not started automatically.
func NewMarquee(id, class string) *Marquee {
	m := &Marquee{
		Animation: Animation{
			Component: Component{id: id, class: class},
			stop:      make(chan struct{}),
		},
		speed: 1,
		gap:   4,
	}
	m.fn = m.Tick
	return m
}

// Apply applies theme styles to the marquee.
func (m *Marquee) Apply(theme *Theme) {
	theme.Apply(m, m.Selector("marquee"))
}

// SetText replaces the scrolling text, resets the scroll offset to 0, and
// redraws the widget. It pre-builds the per-column rune index used by Render.
func (m *Marquee) SetText(s string) *Marquee {
	m.text = s
	m.rebuildCols()
	m.offset = 0
	Redraw(m)
	return m
}

// Text returns the current scrolling text.
func (m *Marquee) Text() string { return m.text }

// SetSpeed sets the number of display columns advanced per tick. Clamped to
// a minimum of 1.
func (m *Marquee) SetSpeed(n int) *Marquee {
	if n < 1 {
		n = 1
	}
	m.speed = n
	return m
}

// SetGap sets the number of blank columns inserted between the end of the text
// and the looping start. Clamped to a minimum of 0. Calls Redraw.
func (m *Marquee) SetGap(n int) *Marquee {
	if n < 0 {
		n = 0
	}
	m.gap = n
	Redraw(m)
	return m
}

// cycle returns the total virtual width of one scroll loop.
func (m *Marquee) cycle() int { return m.textWidth + m.gap }

// rebuildCols rebuilds the per-display-column rune index from m.text.
func (m *Marquee) rebuildCols() {
	m.cols = m.cols[:0]
	col := 0
	for _, r := range m.text {
		w := uniseg.StringWidth(string(r))
		if w < 1 {
			w = 1
		}
		m.cols = append(m.cols, runeCol{r: r, width: w})
		col += w
		// For wide runes, append a sentinel for the second display column.
		if w == 2 {
			m.cols = append(m.cols, runeCol{r: 0, width: 0})
		}
	}
	m.textWidth = col
}

// Hint returns the preferred content size for the marquee.
// Width 0 means "fill the parent"; height is always 1.
func (m *Marquee) Hint() (int, int) {
	if m.hwidth != 0 || m.hheight != 0 {
		return m.hwidth, m.hheight
	}
	s := m.Style()
	return 0, 1 + s.Vertical()
}

// Tick advances the scroll offset by m.speed and triggers a redraw.
// It is a no-op when the widget is hovered or the text fits without scrolling.
func (m *Marquee) Tick() {
	if m.Flag(FlagHovered) {
		return
	}
	if m.textWidth <= 0 || m.textWidth <= m.renderWidth {
		return
	}
	m.offset = (m.offset + m.speed) % m.cycle()
	Redraw(m)
}

// Render draws the marquee to the screen.
func (m *Marquee) Render(r *Renderer) {
	if m.Flag(FlagHidden) {
		return
	}
	m.Component.Render(r)

	cx, cy, cw, _ := m.Content()
	m.renderWidth = cw

	if m.textWidth == 0 || cw == 0 {
		return
	}

	style := m.Style()
	r.Set(style.Foreground(), style.Background(), style.Font())

	if m.textWidth <= cw {
		// Text fits — render left-aligned, padded with spaces.
		r.Text(cx, cy, m.text, cw)
		return
	}

	// Scrolling path: render cw display columns starting at virtual position m.offset.
	cycle := m.cycle()
	col := 0
	for col < cw {
		vpos := (m.offset + col) % cycle
		if vpos < m.textWidth {
			rc := m.cols[vpos]
			if rc.width == 0 {
				// Second cell of a wide character — emit a space so we don't
				// draw a half-character at the scroll boundary.
				r.Put(cx+col, cy, " ")
				col++
			} else {
				r.Put(cx+col, cy, string(rc.r))
				col += rc.width
			}
		} else {
			// Inside the gap region.
			r.Put(cx+col, cy, " ")
			col++
		}
	}
}

package zeichenwerk

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

// Breadcrumb renders a single-row path indicator showing an ordered list of
// string segments separated by a configurable separator. Each segment is
// individually focusable and clickable. When the total rendered width exceeds
// the available space, leading segments are collapsed to an overflow marker,
// always keeping the focused segment visible.
//
// Events:
//   - [EvtSelect]   – int: focused segment index changed
//   - [EvtActivate] – int: Enter pressed, or segment clicked while already focused
type Breadcrumb struct {
	Component
	segments  []string
	selected  int
	first     int
	separator string
	overflow  string
}

// NewBreadcrumb creates a new Breadcrumb with sensible defaults: no segments,
// selected = -1, first = 0, separator = " › ", overflow = "…".
func NewBreadcrumb(id, class string) *Breadcrumb {
	c := &Breadcrumb{
		Component: Component{id: id, class: class},
		selected:  -1,
		first:     0,
		separator: " › ",
		overflow:  "…",
	}
	c.SetFlag(FlagFocusable, true)
	c.On(EvtFocus, func(_ Widget, _ Event, _ ...any) bool {
		if c.selected == -1 && len(c.segments) > 0 {
			c.Select(len(c.segments) - 1)
		}
		return false
	})
	OnKey(c, c.handleKey)
	OnMouse(c, c.handleMouse)
	return c
}

// ── Data ──────────────────────────────────────────────────────────────────────

// SetSegments replaces all segments, resets first to 0, clamps selected, and redraws.
func (c *Breadcrumb) SetSegments(segs []string) {
	c.segments = segs
	c.first = 0
	if c.selected >= len(segs) {
		c.selected = len(segs) - 1
	}
	c.Refresh()
}

// Push appends one segment and redraws.
func (c *Breadcrumb) Push(seg string) {
	c.segments = append(c.segments, seg)
	c.Refresh()
}

// Pop removes and returns the last segment, clamps selected, and redraws.
func (c *Breadcrumb) Pop() string {
	n := len(c.segments)
	if n == 0 {
		return ""
	}
	seg := c.segments[n-1]
	c.segments = c.segments[:n-1]
	if c.selected >= len(c.segments) {
		c.selected = len(c.segments) - 1
	}
	c.Refresh()
	return seg
}

// Truncate removes all segments after index (keeps segments[0..index] inclusive).
func (c *Breadcrumb) Truncate(index int) {
	if index < 0 {
		index = 0
	}
	if index+1 < len(c.segments) {
		c.segments = c.segments[:index+1]
	}
	if c.selected >= len(c.segments) {
		c.selected = len(c.segments) - 1
	}
	c.Refresh()
}

// Segments returns the current segments slice.
func (c *Breadcrumb) Segments() []string { return c.segments }

// ── Navigation ────────────────────────────────────────────────────────────────

// Select focuses the segment at index, clamps to valid range, ensures it is
// visible, and dispatches EvtSelect. No-op when already selected.
func (c *Breadcrumb) Select(index int) {
	n := len(c.segments)
	if n == 0 {
		c.selected = -1
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= n {
		index = n - 1
	}
	if index == c.selected {
		return
	}
	c.selected = index
	if index < c.first {
		c.first = index
	}
	c.Dispatch(c, EvtSelect, index)
	Redraw(c)
}

// Selected returns the currently focused segment index, or -1 if none.
func (c *Breadcrumb) Selected() int { return c.selected }

// ── Display setters ───────────────────────────────────────────────────────────

// SetSeparator overrides the separator string and redraws.
func (c *Breadcrumb) SetSeparator(sep string) { c.separator = sep; c.Refresh() }

// SetOverflow overrides the overflow marker and redraws.
func (c *Breadcrumb) SetOverflow(marker string) { c.overflow = marker; c.Refresh() }

// ── Theme / Apply ─────────────────────────────────────────────────────────────

// Apply registers all breadcrumb style selectors and reads theme strings.
func (c *Breadcrumb) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("breadcrumb"), "focused", "hovered", "disabled")
	theme.Apply(c, c.Selector("breadcrumb/segment"), "focused")
	theme.Apply(c, c.Selector("breadcrumb/separator"))
	str := func(key, def string) string {
		if s := theme.String(key); s != "" {
			return s
		}
		return def
	}
	c.separator = str("breadcrumb.separator", " › ")
	c.overflow = str("breadcrumb.overflow", "…")
}

// ── Hint ──────────────────────────────────────────────────────────────────────

// Hint returns the natural width of all segments joined by the separator, plus
// style overhead. Returns 0 (fill parent) width when no segments are set.
// Height is always 1 plus style vertical overhead.
func (c *Breadcrumb) Hint() (int, int) {
	if c.hwidth != 0 || c.hheight != 0 {
		return c.hwidth, c.hheight
	}
	n := len(c.segments)
	if n == 0 {
		return 0, 1 + c.Style().Vertical()
	}
	w := 0
	sepW := utf8.RuneCountInString(c.separator)
	for _, seg := range c.segments {
		w += utf8.RuneCountInString(seg)
	}
	w += sepW * (n - 1)
	w += c.Style().Horizontal()
	return w, 1 + c.Style().Vertical()
}

// ── Overflow helpers ──────────────────────────────────────────────────────────

// renderWidth returns the display width of segments[start:] with an optional
// overflow prefix when start > 0.
func (c *Breadcrumb) renderWidth(start int) int {
	sepW := utf8.RuneCountInString(c.separator)
	w := 0
	if start > 0 {
		w = utf8.RuneCountInString(c.overflow) + sepW
	}
	segs := c.segments[start:]
	for i, seg := range segs {
		w += utf8.RuneCountInString(seg)
		if i < len(segs)-1 {
			w += sepW
		}
	}
	return w
}

// computeFirstVis returns the smallest start index such that segments[start:]
// (with overflow prefix if start > 0) fits within availW. At least one segment
// is always shown.
func (c *Breadcrumb) computeFirstVis(availW int) int {
	start := c.first
	if start < 0 {
		start = 0
	}
	for {
		if c.renderWidth(start) <= availW {
			return start
		}
		start++
		if start >= len(c.segments)-1 {
			return start
		}
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

// Render draws the breadcrumb.
func (c *Breadcrumb) Render(r *Renderer) {
	if c.Flag(FlagHidden) {
		return
	}
	c.Component.Render(r)
	cx, cy, cw, _ := c.Content()
	if cw < 1 || len(c.segments) == 0 {
		return
	}

	sepS := c.Style("separator")
	segS := c.Style("segment")
	segFocS := c.Style("segment:focused")

	sepW := utf8.RuneCountInString(c.separator)
	start := c.computeFirstVis(cw)
	x := cx

	// Overflow prefix.
	if start > 0 {
		ovfW := utf8.RuneCountInString(c.overflow)
		r.Set(sepS.Foreground(), sepS.Background(), sepS.Font())
		r.Text(x, cy, c.overflow, min(ovfW, cx+cw-x))
		x += ovfW
		if x < cx+cw {
			r.Text(x, cy, c.separator, min(sepW, cx+cw-x))
			x += sepW
		}
	}

	for i := start; i < len(c.segments); i++ {
		if x >= cx+cw {
			break
		}
		seg := c.segments[i]
		segW := utf8.RuneCountInString(seg)
		remaining := cx + cw - x
		if c.Flag(FlagFocused) && i == c.selected {
			r.Set(segFocS.Foreground(), segFocS.Background(), segFocS.Font())
		} else {
			r.Set(segS.Foreground(), segS.Background(), segS.Font())
		}
		r.Text(x, cy, seg, min(segW, remaining))
		x += segW
		if i < len(c.segments)-1 && x < cx+cw {
			r.Set(sepS.Foreground(), sepS.Background(), sepS.Font())
			r.Text(x, cy, c.separator, min(sepW, cx+cw-x))
			x += sepW
		}
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func (c *Breadcrumb) handleKey(evt *tcell.EventKey) bool {
	n := len(c.segments)
	if n == 0 {
		return false
	}
	switch evt.Key() {
	case tcell.KeyLeft:
		if c.selected <= 0 {
			c.Select(n - 1)
		} else {
			c.Select(c.selected - 1)
		}
		return true
	case tcell.KeyRight:
		if c.selected >= n-1 {
			c.Select(0)
		} else {
			c.Select(c.selected + 1)
		}
		return true
	case tcell.KeyHome:
		c.Select(0)
		return true
	case tcell.KeyEnd:
		c.Select(n - 1)
		return true
	case tcell.KeyEnter:
		if c.selected >= 0 {
			c.Dispatch(c, EvtActivate, c.selected)
		}
		return true
	}
	return false
}

// ── Mouse ─────────────────────────────────────────────────────────────────────

func (c *Breadcrumb) handleMouse(evt *tcell.EventMouse) bool {
	if evt.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := evt.Position()
	cx, cy, cw, _ := c.Content()
	if mx < cx || mx >= cx+cw || my != cy {
		return false
	}

	start := c.computeFirstVis(cw)
	x := cx

	// Overflow prefix is not clickable — skip past it.
	if start > 0 {
		x += utf8.RuneCountInString(c.overflow) + utf8.RuneCountInString(c.separator)
		if mx < x {
			return false
		}
	}

	sepW := utf8.RuneCountInString(c.separator)
	for i := start; i < len(c.segments); i++ {
		segW := utf8.RuneCountInString(c.segments[i])
		if mx >= x && mx < x+segW {
			prev := c.selected
			c.Select(i)
			if i == prev {
				c.Dispatch(c, EvtActivate, i)
			}
			return true
		}
		x += segW + sepW
		if x >= cx+cw {
			break
		}
	}
	return false
}

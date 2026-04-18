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
	bc := &Breadcrumb{
		Component: Component{id: id, class: class},
		selected:  -1,
		first:     0,
		separator: " › ",
		overflow:  "…",
	}
	bc.SetFlag(FlagFocusable, true)
	bc.On(EvtFocus, func(_ Widget, _ Event, _ ...any) bool {
		if bc.selected == -1 && len(bc.segments) > 0 {
			bc.Select(len(bc.segments) - 1)
		}
		return false
	})
	OnKey(bc, bc.handleKey)
	OnMouse(bc, bc.handleMouse)
	return bc
}

// ---- Widget Methods -------------------------------------------------------

// Apply registers all breadcrumb style selectors and reads theme strings.
func (bc *Breadcrumb) Apply(theme *Theme) {
	theme.Apply(bc, bc.Selector("breadcrumb"), "focused", "hovered", "disabled")
	theme.Apply(bc, bc.Selector("breadcrumb/segment"), "focused")
	theme.Apply(bc, bc.Selector("breadcrumb/separator"))
	str := func(key, def string) string {
		if s := theme.String(key); s != "" {
			return s
		}
		return def
	}
	bc.separator = str("breadcrumb.separator", " › ")
	bc.overflow = str("breadcrumb.overflow", "…")
}

// Hint returns the natural width of all segments joined by the separator, plus
// style overhead. Returns 0 (fill parent) width when no segments are set.
// Height is always 1 plus style vertical overhead.
func (bc *Breadcrumb) Hint() (int, int) {
	if bc.hwidth != 0 || bc.hheight != 0 {
		return bc.hwidth, bc.hheight
	}
	n := len(bc.segments)
	if n == 0 {
		return 0, 1 + bc.Style().Vertical()
	}
	w := 0
	sepW := utf8.RuneCountInString(bc.separator)
	for _, seg := range bc.segments {
		w += utf8.RuneCountInString(seg)
	}
	w += sepW * (n - 1)
	w += bc.Style().Horizontal()
	return w, 1 + bc.Style().Vertical()
}

// ---- Getter and Setter ----------------------------------------------------

// Get returns the current segments.
func (bc *Breadcrumb) Get() []string {
	return bc.segments
}

// Sets replaces all segments, resets first to 0, clamps selected, and redraws.
func (bc *Breadcrumb) Set(segs []string) {
	bc.segments = segs
	bc.first = 0
	if bc.selected >= len(segs) {
		bc.selected = len(segs) - 1
	}
	bc.Refresh()
}

// ---- Breadcrumb Methods ---------------------------------------------------

// Pop removes and returns the last segment, clamps selected, and redraws.
func (bc *Breadcrumb) Pop() string {
	n := len(bc.segments)
	if n == 0 {
		return ""
	}
	seg := bc.segments[n-1]
	bc.segments = bc.segments[:n-1]
	if bc.selected >= len(bc.segments) {
		bc.selected = len(bc.segments) - 1
	}
	bc.Refresh()
	return seg
}

// Push appends one segment and redraws.
func (bc *Breadcrumb) Push(seg string) {
	bc.segments = append(bc.segments, seg)
	bc.Refresh()
}

// Truncate removes all segments after index (keeps segments[0..index] inclusive).
func (bc *Breadcrumb) Truncate(index int) {
	if index < 0 {
		index = 0
	}
	if index+1 < len(bc.segments) {
		bc.segments = bc.segments[:index+1]
	}
	if bc.selected >= len(bc.segments) {
		bc.selected = len(bc.segments) - 1
	}
	bc.Refresh()
}

// Segments returns the current segments slice.
func (bc *Breadcrumb) Segments() []string { return bc.segments }

// ---- Navigation -----------------------------------------------------------

// Select focuses the segment at index, clamps to valid range, ensures it is
// visible, and dispatches EvtSelect. No-op when already selected.
func (bc *Breadcrumb) Select(index int) {
	n := len(bc.segments)
	if n == 0 {
		bc.selected = -1
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= n {
		index = n - 1
	}
	if index == bc.selected {
		return
	}
	bc.selected = index
	if index < bc.first {
		bc.first = index
	}
	bc.Dispatch(bc, EvtSelect, index)
	Redraw(bc)
}

// Selected returns the currently focused segment index, or -1 if none.
func (bc *Breadcrumb) Selected() int { return bc.selected }

// ---- Display Options ------------------------------------------------------

// SetOverflow overrides the overflow marker and redraws.
func (bc *Breadcrumb) SetOverflow(marker string) {
	bc.overflow = marker
	bc.Refresh()
}

// SetSeparator overrides the separator string and redraws.
func (bc *Breadcrumb) SetSeparator(sep string) {
	bc.separator = sep
	bc.Refresh()
}

// ---- Overflow Helpers -----------------------------------------------------

// computeFirst returns the smallest start index such that segments[start:]
// (with overflow prefix if start > 0) fits within availW. At least one segment
// is always shown.
func (bc *Breadcrumb) computeFirst(availW int) int {
	start := bc.first
	if start < 0 {
		start = 0
	}
	for {
		if bc.renderWidth(start) <= availW {
			return start
		}
		start++
		if start >= len(bc.segments)-1 {
			return start
		}
	}
}

// renderWidth returns the display width of segments[start:] with an optional
// overflow prefix when start > 0.
func (bc *Breadcrumb) renderWidth(start int) int {
	sepW := utf8.RuneCountInString(bc.separator)
	w := 0
	if start > 0 {
		w = utf8.RuneCountInString(bc.overflow) + sepW
	}
	segs := bc.segments[start:]
	for i, seg := range segs {
		w += utf8.RuneCountInString(seg)
		if i < len(segs)-1 {
			w += sepW
		}
	}
	return w
}

// ---- Rendering ------------------------------------------------------------

// Render draws the breadcrumb.
func (bc *Breadcrumb) Render(r *Renderer) {
	if bc.Flag(FlagHidden) {
		return
	}
	bc.Component.Render(r)
	cx, cy, cw, _ := bc.Content()
	if cw < 1 || len(bc.segments) == 0 {
		return
	}

	sepS := bc.Style("separator")
	segS := bc.Style("segment")
	segFocS := bc.Style("segment:focused")

	sepW := utf8.RuneCountInString(bc.separator)
	start := bc.computeFirst(cw)
	x := cx

	// Overflow prefix.
	if start > 0 {
		ovfW := utf8.RuneCountInString(bc.overflow)
		r.Set(sepS.Foreground(), sepS.Background(), sepS.Font())
		r.Text(x, cy, bc.overflow, min(ovfW, cx+cw-x))
		x += ovfW
		if x < cx+cw {
			r.Text(x, cy, bc.separator, min(sepW, cx+cw-x))
			x += sepW
		}
	}

	for i := start; i < len(bc.segments); i++ {
		if x >= cx+cw {
			break
		}
		seg := bc.segments[i]
		segW := utf8.RuneCountInString(seg)
		remaining := cx + cw - x
		if bc.Flag(FlagFocused) && i == bc.selected {
			r.Set(segFocS.Foreground(), segFocS.Background(), segFocS.Font())
		} else {
			r.Set(segS.Foreground(), segS.Background(), segS.Font())
		}
		r.Text(x, cy, seg, min(segW, remaining))
		x += segW
		if i < len(bc.segments)-1 && x < cx+cw {
			r.Set(sepS.Foreground(), sepS.Background(), sepS.Font())
			r.Text(x, cy, bc.separator, min(sepW, cx+cw-x))
			x += sepW
		}
	}
}

// ---- Event Handling -------------------------------------------------------

func (bc *Breadcrumb) handleKey(evt *tcell.EventKey) bool {
	n := len(bc.segments)
	if n == 0 {
		return false
	}
	switch evt.Key() {
	case tcell.KeyLeft:
		if bc.selected <= 0 {
			bc.Select(n - 1)
		} else {
			bc.Select(bc.selected - 1)
		}
		return true
	case tcell.KeyRight:
		if bc.selected >= n-1 {
			bc.Select(0)
		} else {
			bc.Select(bc.selected + 1)
		}
		return true
	case tcell.KeyHome:
		bc.Select(0)
		return true
	case tcell.KeyEnd:
		bc.Select(n - 1)
		return true
	case tcell.KeyEnter:
		if bc.selected >= 0 {
			bc.Dispatch(bc, EvtActivate, bc.selected)
		}
		return true
	}
	return false
}

func (bc *Breadcrumb) handleMouse(evt *tcell.EventMouse) bool {
	if evt.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := evt.Position()
	cx, cy, cw, _ := bc.Content()
	if mx < cx || mx >= cx+cw || my != cy {
		return false
	}

	start := bc.computeFirst(cw)
	x := cx

	// Overflow prefix is not clickable — skip past it.
	if start > 0 {
		x += utf8.RuneCountInString(bc.overflow) + utf8.RuneCountInString(bc.separator)
		if mx < x {
			return false
		}
	}

	sepW := utf8.RuneCountInString(bc.separator)
	for i := start; i < len(bc.segments); i++ {
		segW := utf8.RuneCountInString(bc.segments[i])
		if mx >= x && mx < x+segW {
			prev := bc.selected
			bc.Select(i)
			if i == prev {
				bc.Dispatch(bc, EvtActivate, i)
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

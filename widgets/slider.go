package widgets

import (
	"fmt"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// Slider represents a horizontal range input that lets the user pick an
// integer value between a configurable minimum and maximum. The widget
// renders in one of two styles depending on the content height the layout
// gives it:
//
//   - Height 1: a one-line compact bar — a horizontal track glyph with a
//     vertical thumb glyph overlaid at the current value's position.
//   - Height ≥ 2: a two-row rounded box (top/bottom border) with the thumb
//     piercing both rows. The box is vertically centred within the content
//     area, so extra height becomes padding above and below.
//
// All glyphs come from theme strings under the "slider.*" keys; the renderer
// only decides which set to use based on the available height. Value semantics
// mirror [Progress] — integer-valued, clamped to [min, max], programmatic
// setters dispatch [EvtChange].
type Slider struct {
	Component
	value int
	min   int
	max   int
	step  int
}

// NewSlider creates a new horizontal slider with sensible defaults
// (min=0, max=100, value=0, step=1). The widget is focusable and registers
// keyboard and mouse handlers.
//
// The size hint defaults to (0, 1) — full width, one row — so the compact
// style renders by default. Place the slider in a layout that gives it two
// or more rows to get the rounded two-row style.
func NewSlider(id, class string) *Slider {
	s := &Slider{
		Component: Component{id: id, class: class},
		value:     0,
		min:       0,
		max:       100,
		step:      1,
	}
	s.SetHint(0, 1)
	s.SetFlag(FlagFocusable, true)
	OnKey(s, s.handleKey)
	OnMouse(s, s.handleMouse)
	return s
}

// Apply installs the theme styles. The slider has no part selectors — the
// entire widget (track and thumb) renders in a single state-resolved style.
func (s *Slider) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("slider"), "disabled", "focused", "hovered")
}

// Refresh requests a redraw scoped to the slider's own bounds.
func (s *Slider) Refresh() {
	Redraw(s)
}

// ---- Value accessors -----------------------------------------------------

// Value returns the current slider value.
func (s *Slider) Value() int { return s.value }

// Min returns the minimum allowed value.
func (s *Slider) Min() int { return s.min }

// Max returns the maximum allowed value.
func (s *Slider) Max() int { return s.max }

// Step returns the step size used by arrow keys.
func (s *Slider) Step() int { return s.step }

// Set sets the current value, clamped to [min, max]. Dispatches [EvtChange]
// when the value actually changes and requests a redraw.
func (s *Slider) Set(value int) {
	if value < s.min {
		value = s.min
	}
	if value > s.max {
		value = s.max
	}
	if value == s.value {
		return
	}
	s.value = value
	s.Dispatch(s, EvtChange, s.value)
	s.Refresh()
}

// SetMin sets the minimum bound. If max < min, max is raised to min. The
// current value is reclamped to the new range.
func (s *Slider) SetMin(v int) {
	s.min = v
	if s.max < s.min {
		s.max = s.min
	}
	if s.value < s.min {
		s.value = s.min
	}
	s.Refresh()
}

// SetMax sets the maximum bound. If max < min, min is lowered to max. The
// current value is reclamped to the new range.
func (s *Slider) SetMax(v int) {
	s.max = v
	if s.min > s.max {
		s.min = s.max
	}
	if s.value > s.max {
		s.value = s.max
	}
	s.Refresh()
}

// SetStep sets the step size used by arrow-key navigation. Values < 1 are
// clamped to 1.
func (s *Slider) SetStep(v int) {
	if v < 1 {
		v = 1
	}
	s.step = v
}

// Info returns a human-readable description of the slider configuration.
func (s *Slider) Info() string {
	return fmt.Sprintf("Slider(value=%d, min=%d, max=%d, step=%d)", s.value, s.min, s.max, s.step)
}

// Summary returns a one-line summary used by the inspector dump.
func (s *Slider) Summary() string {
	return fmt.Sprintf("value=%d [%d..%d]", s.value, s.min, s.max)
}

// ---- Rendering -----------------------------------------------------------

// thumbColumn returns the offset (0..w-1) of the thumb within a track of
// width w. When max == min the thumb sits at the leftmost column.
func (s *Slider) thumbColumn(w int) int {
	if w <= 1 {
		return 0
	}
	span := s.max - s.min
	if span <= 0 {
		return 0
	}
	rel := s.value - s.min
	// Round to nearest column to keep the thumb centred on its slot.
	return (rel*(w-1)*2 + span) / (2 * span)
}

// Render draws the slider. The visual style is picked from the content
// height: 1 row uses the compact track-and-thumb glyphs; 2+ rows use the
// rounded two-row box, vertically centred. The whole widget paints in a
// single state-resolved style — there is no separate thumb style, and
// Render does not change the renderer's style mid-draw.
func (s *Slider) Render(r *Renderer) {
	if s.Flag(FlagHidden) {
		return
	}
	s.Component.Render(r)

	x, y, w, h := s.Content()
	if w <= 0 || h <= 0 {
		return
	}

	style := s.Style()
	r.Set(style.Foreground(), style.Background(), style.Font())

	if h == 1 {
		s.renderCompact(r, x, y, w)
		return
	}
	// Two-row style centred vertically in the content area.
	offset := (h - 2) / 2
	s.renderBox(r, x, y+offset, w)
}

// renderCompact draws the one-row style: a horizontal track filling the row,
// with the thumb glyph overwriting the column at the current value. The
// renderer style must be set by the caller before this is invoked.
func (s *Slider) renderCompact(r *Renderer, x, y, w int) {
	track := stringOr(r.Theme.String("slider.compact.track"), "━")
	thumb := stringOr(r.Theme.String("slider.compact.thumb"), "┃")

	r.Fill(x, y, w, 1, track)
	r.Put(x+s.thumbColumn(w), y, thumb)
}

// renderBox draws the two-row style: a rounded box (top row + bottom row)
// with the thumb glyphs piercing both rows at the current value column. The
// renderer style must be set by the caller before this is invoked.
func (s *Slider) renderBox(r *Renderer, x, y, w int) {
	if w < 2 {
		// Not enough room for a box; degrade to compact.
		s.renderCompact(r, x, y, w)
		return
	}

	tl := stringOr(r.Theme.String("slider.box.top-left"), "╭")
	tr := stringOr(r.Theme.String("slider.box.top-right"), "╮")
	bl := stringOr(r.Theme.String("slider.box.bottom-left"), "╰")
	br := stringOr(r.Theme.String("slider.box.bottom-right"), "╯")
	hr := stringOr(r.Theme.String("slider.box.horizontal"), "─")
	tT := stringOr(r.Theme.String("slider.box.thumb-top"), "╥")
	bT := stringOr(r.Theme.String("slider.box.thumb-bottom"), "╨")

	// Top row: tl, hr×(w-2), tr
	r.Put(x, y, tl)
	for i := 1; i < w-1; i++ {
		r.Put(x+i, y, hr)
	}
	r.Put(x+w-1, y, tr)

	// Bottom row: bl, hr×(w-2), br
	r.Put(x, y+1, bl)
	for i := 1; i < w-1; i++ {
		r.Put(x+i, y+1, hr)
	}
	r.Put(x+w-1, y+1, br)

	// Thumb pierces both rows. The thumb is positioned over the inner
	// width [1..w-2] so it never replaces the corner glyphs.
	innerW := w - 2
	col := s.thumbColumn(innerW)
	if col >= innerW {
		col = innerW - 1
	}
	thumbX := x + 1 + col

	r.Put(thumbX, y, tT)
	r.Put(thumbX, y+1, bT)
}

// stringOr returns s if non-empty, otherwise fallback.
func stringOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// ---- Input handling ------------------------------------------------------

// handleKey processes keyboard navigation. ←/h decreases by one step,
// →/l increases by one step, Home jumps to min, End jumps to max.
func (s *Slider) handleKey(evt *tcell.EventKey) bool {
	if s.Flag(FlagReadonly) || s.Flag(FlagDisabled) {
		return false
	}
	switch evt.Key() {
	case tcell.KeyLeft:
		s.Set(s.value - s.step)
		return true
	case tcell.KeyRight:
		s.Set(s.value + s.step)
		return true
	case tcell.KeyHome:
		s.Set(s.min)
		return true
	case tcell.KeyEnd:
		s.Set(s.max)
		return true
	case tcell.KeyRune:
		switch evt.Str() {
		case "h":
			s.Set(s.value - s.step)
			return true
		case "l":
			s.Set(s.value + s.step)
			return true
		}
	}
	return false
}

// handleMouse maps a left-button click on the slider to a value. The mouse
// column is mapped to the same track geometry the renderer uses, so clicks
// land where they look like they should regardless of the render style.
func (s *Slider) handleMouse(evt *tcell.EventMouse) bool {
	if s.Flag(FlagReadonly) || s.Flag(FlagDisabled) {
		return false
	}
	if evt.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := evt.Position()
	cx, cy, cw, ch := s.Content()
	if mx < cx || mx >= cx+cw || my < cy || my >= cy+ch {
		return false
	}

	// Compute the column range the track occupies. For the compact style
	// the track spans the full content width; for the box style it spans
	// the inner width (excluding the rounded corners).
	var trackX, trackW int
	if ch == 1 {
		trackX, trackW = cx, cw
	} else {
		if cw < 2 {
			trackX, trackW = cx, cw
		} else {
			trackX, trackW = cx+1, cw-2
		}
	}
	rel := mx - trackX
	if rel < 0 {
		rel = 0
	}
	if rel >= trackW {
		rel = trackW - 1
	}
	if trackW <= 1 {
		s.Set(s.min)
		return true
	}
	span := s.max - s.min
	// Round-half-up mapping mouse column → value.
	v := s.min + (rel*span*2+(trackW-1))/(2*(trackW-1))
	s.Set(v)
	return true
}


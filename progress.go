package next

import (
	"fmt"
	"unicode/utf8"
)

// Progress represents a visual progress indicator widget that displays the completion
// status of a task or operation. It provides visual feedback through a filled bar that
// represents progress percentage. The progress bar supports both determinate and
// indeterminate (spinning/loading) modes.
//
// The widget is fully themeable via theme strings with keys like "progress.h.*" for
// horizontal orientation and "progress.v.*" for vertical orientation.
type Progress struct {
	Component
	value      int  // Current progress value (0 to total)
	total      int  // Total work units (0 = indeterminate mode)
	horizontal bool // horizontal orientation flag
}

// NewProgress creates a new progress widget with default settings.
// The progress indicator is initialized with value=0 and total=0 (indeterminate mode).
// It defaults to horizontal orientation.
//
// Parameters:
//   - id: Unique identifier for the progress widget
//
// Returns:
//   - *Progress: A new progress widget instance
//
// Example usage:
//
//	progress := NewProgress("download-progress")
//	progress.SetTotal(100)   // Set determinate mode
//	progress.SetValue(50)    // 50% complete
func NewProgress(id string, horizontal bool) *Progress {
	progress := &Progress{
		Component:  Component{id: id},
		value:      0,
		total:      0,
		horizontal: horizontal,
	}
	if horizontal {
		progress.SetHint(0, 1)
	} else {
		progress.SetHint(1, 0)
	}
	return progress
}

// SetValue sets the current progress value. The value is automatically clamped
// to the range 0..total. In indeterminate mode (total==0), the value is ignored
// but can be stored for internal use.
//
// Parameters:
//   - value: The new progress value (will be clamped to 0-total)
func (p *Progress) SetValue(value int) {
	if p.total == 0 {
		// Indeterminate mode - store but don't clamp
		p.value = value
		return
	}
	if value < 0 {
		value = 0
	}
	if value > p.total {
		value = p.total
	}
	p.value = value
}

// SetTotal sets the total amount of work. If total is 0, the progress enters
// indeterminate mode (often shown as a spinning or pulsing indicator). If total
// is greater than 0, the progress enters determinate mode where SetValue controls
// the completion percentage.
//
// Parameters:
//   - total: The total work units (0 for indeterminate)
func (p *Progress) SetTotal(total int) {
	if total < 0 {
		total = 0
	}
	p.total = total
	// Reclamp current value to new total
	p.SetValue(p.value)
}

// Percentage returns the current progress as a percentage (0.0 to 100.0).
// In indeterminate mode (total==0), returns 0.0.
func (p *Progress) Percentage() float64 {
	if p.total == 0 {
		return 0.0
	}
	return float64(p.value) / float64(p.total) * 100.0
}

// Increment increases the progress value by the specified amount.
// The result is automatically clamped to the valid range.
//
// Parameters:
//   - amount: The amount to add to the current value
func (p *Progress) Increment(amount int) {
	p.SetValue(p.value + amount)
}

// Info returns a human-readable description of the progress widget's configuration.
func (p *Progress) Info() string {
	return fmt.Sprintf("Progress(value=%d, total=%d)", p.value, p.total)
}

// Render implements the Widget interface. It delegates to the appropriate
// orientation-specific renderer.
func (p *Progress) Render(r *Renderer) {
	// Check if the widget is visible
	if p.Flag("hidden") {
		return
	}

	// Render component styling
	p.Component.Render(r)

	// Determine the style based on state
	state := p.State()
	if state != "" {
		state = ":" + state
	}
	baseStyle := p.Style(state)

	// Get content area
	x, y, w, h := p.Content()
	if w <= 0 || h <= 0 {
		return
	}

	// Switch by orientation
	if p.horizontal {
		p.renderHorizontal(r, x, y, w, h, baseStyle)
	} else {
		p.renderVertical(r, x, y, w, h, baseStyle)
	}
}

// renderHorizontal renders a horizontal progress bar using theme strings.
// The bar is composed of: prefix + [start][middle*][end] + suffix.
// Filled and empty portions use different glyphs from the theme.
func (p *Progress) renderHorizontal(r *Renderer, x, y, w, h int, baseStyle *Style) {
	// Fetch theme strings for horizontal orientation
	prefix := r.theme.String("progress.h.prefix")
	suffix := r.theme.String("progress.h.suffix")
	startFilled := r.theme.String("progress.h.start.filled")
	startEmpty := r.theme.String("progress.h.start.empty")
	middleFilled := r.theme.String("progress.h.middle.filled")
	middleEmpty := r.theme.String("progress.h.middle.empty")
	endFilled := r.theme.String("progress.h.end.filled")
	endEmpty := r.theme.String("progress.h.end.empty")

	// Fallbacks if theme strings are missing
	if startFilled == "" {
		startFilled = "#"
	}
	if startEmpty == "" {
		startEmpty = "."
	}
	if middleFilled == "" {
		middleFilled = "#"
	}
	if middleEmpty == "" {
		middleEmpty = "."
	}
	if endFilled == "" {
		endFilled = "#"
	}
	if endEmpty == "" {
		endEmpty = "."
	}

	// Compute available track width accounting for prefix/suffix (in cells)
	prefixWidth := utf8.RuneCountInString(prefix)
	suffixWidth := utf8.RuneCountInString(suffix)
	trackW := w - prefixWidth - suffixWidth
	if trackW < 0 {
		trackW = 0
		// Not enough space - render prefix only if space available
		if w > 0 {
			baseFg := r.theme.Color(baseStyle.Foreground())
			baseBg := r.theme.Color(baseStyle.Background())
			baseFont := baseStyle.Font()
			r.Set(baseFg, baseBg, baseFont)
			r.Text(x, y, prefix, w)
		}
		return
	}

	// Compute number of filled cells
	fill := 0
	if p.total > 0 {
		fill = (p.value * trackW) / p.total
	}

	// Render prefix (left side) using base style
	curX := x
	if prefixWidth > 0 {
		baseFg := r.theme.Color(baseStyle.Foreground())
		baseBg := r.theme.Color(baseStyle.Background())
		baseFont := baseStyle.Font()
		r.Set(baseFg, baseBg, baseFont)
		r.Text(curX, y, prefix, prefixWidth)
		curX += prefixWidth
	}

	// Get bar style for filled portion, fallback to base
	barStyle := p.Style("bar")
	if barStyle == nil || barStyle == &DefaultStyle {
		barStyle = baseStyle
	}
	barFg := r.theme.Color(barStyle.Foreground())
	barBg := r.theme.Color(barStyle.Background())
	barFont := barStyle.Font()

	// Render track cells directly with Put for proper Unicode width handling
	// Filled portion first
	cell := 0
	for ; cell < fill; cell++ {
		var ch string
		switch {
		case cell == 0:
			ch = startFilled
		case cell == trackW-1:
			ch = endFilled
		default:
			ch = middleFilled
		}
		r.Set(barFg, barBg, barFont)
		r.Put(curX, y, ch)
		curX += utf8.RuneCountInString(ch)
	}

	// Empty portion
	for ; cell < trackW; cell++ {
		var ch string
		switch {
		case cell == 0:
			ch = startEmpty
		case cell == trackW-1:
			ch = endEmpty
		default:
			ch = middleEmpty
		}
		r.Set(r.theme.Color(baseStyle.Foreground()), r.theme.Color(baseStyle.Background()), baseStyle.Font())
		r.Put(curX, y, ch)
		curX += utf8.RuneCountInString(ch)
	}

	// Render suffix (right side)
	if suffixWidth > 0 {
		baseFg := r.theme.Color(baseStyle.Foreground())
		baseBg := r.theme.Color(baseStyle.Background())
		baseFont := baseStyle.Font()
		r.Set(baseFg, baseBg, baseFont)
		r.Text(x+w-suffixWidth, y, suffix, suffixWidth)
	}
}

// renderVertical renders a vertical progress bar (bottom-up filling).
// It uses theme keys "progress.v.*" analogous to horizontal but transposed.
func (p *Progress) renderVertical(r *Renderer, x, y, w, h int, baseStyle *Style) {
	prefix := r.theme.String("progress.v.prefix")
	suffix := r.theme.String("progress.v.suffix")
	startFilled := r.theme.String("progress.v.start.filled") // top of filled (since we fill upward, top cap of filled region)
	startEmpty := r.theme.String("progress.v.start.empty")   // top of empty region
	middleFilled := r.theme.String("progress.v.middle.filled")
	middleEmpty := r.theme.String("progress.v.middle.empty")
	endFilled := r.theme.String("progress.v.end.filled") // bottom of filled region
	endEmpty := r.theme.String("progress.v.end.empty")   // bottom of empty region

	// Fallbacks
	if startFilled == "" {
		startFilled = "#"
	}
	if startEmpty == "" {
		startEmpty = "."
	}
	if middleFilled == "" {
		middleFilled = "#"
	}
	if middleEmpty == "" {
		middleEmpty = "."
	}
	if endFilled == "" {
		endFilled = "#"
	}
	if endEmpty == "" {
		endEmpty = "."
	}

	// For vertical, prefix and suffix are single-line strings placed above and below the track.
	prefixRows := 0
	if prefix != "" {
		prefixRows = 1
	}
	suffixRows := 0
	if suffix != "" {
		suffixRows = 1
	}
	trackH := h - prefixRows - suffixRows
	if trackH < 0 {
		trackH = 0
		if h > 0 {
			// Not enough space; just render prefix maybe
			if prefix != "" {
				baseFg := r.theme.Color(baseStyle.Foreground())
				baseBg := r.theme.Color(baseStyle.Background())
				baseFont := baseStyle.Font()
				r.Set(baseFg, baseBg, baseFont)
				r.Text(x, y, prefix, w)
			}
		}
		return
	}

	fill := 0
	if p.total > 0 {
		fill = (p.value * trackH) / p.total
	}

	curY := y
	// Render prefix (top)
	if prefixRows > 0 {
		baseFg := r.theme.Color(baseStyle.Foreground())
		baseBg := r.theme.Color(baseStyle.Background())
		baseFont := baseStyle.Font()
		r.Set(baseFg, baseBg, baseFont)
		r.Text(x, curY, prefix, w)
		curY += prefixRows
	}

	// Render track rows: from top (curY) to curY+trackH-1.
	// Filled cells are the bottom-most `fill` rows.
	barStyle := p.Style("bar")
	if barStyle == nil || barStyle == &DefaultStyle {
		barStyle = baseStyle
	}
	barFg := r.theme.Color(barStyle.Foreground())
	barBg := r.theme.Color(barStyle.Background())
	barFont := barStyle.Font()

	// Iterate rows in the track
	for i := 0; i < trackH; i++ {
		// Row index relative to track: 0=top, trackH-1=bottom
		isFilled := i >= trackH-fill
		var ch string
		switch {
		case i == 0:
			if isFilled {
				ch = startFilled
			} else {
				ch = startEmpty
			}
		case i == trackH-1:
			if isFilled {
				ch = endFilled
			} else {
				ch = endEmpty
			}
		default:
			if isFilled {
				ch = middleFilled
			} else {
				ch = middleEmpty
			}
		}

		// Set style
		if isFilled {
			r.Set(barFg, barBg, barFont)
		} else {
			r.Set(r.theme.Color(baseStyle.Foreground()), r.theme.Color(baseStyle.Background()), baseStyle.Font())
		}
		r.Text(x, curY+i, ch, 1)
	}

	// Render suffix (bottom)
	if suffixRows > 0 {
		baseFg := r.theme.Color(baseStyle.Foreground())
		baseBg := r.theme.Color(baseStyle.Background())
		baseFont := baseStyle.Font()
		r.Set(baseFg, baseBg, baseFont)
		r.Text(x, y+h-suffixRows, suffix, w)
	}
}

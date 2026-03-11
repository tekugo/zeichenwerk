package zeichenwerk

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type scannerChar struct {
	ch     string  // Render character
	fading float64 // color fading factor (dimming)
	color  string  // Hex colors string, calculated automatically, if fading is > 0
}

// scannerStyle defines character sets for each scanner style.
type scannerStyle struct {
	active   scannerChar   // character and color for the moving scanner
	inactive scannerChar   // character and color for inactive (outside trail)
	trail    []scannerChar // character and color for the scanner trail
}

var scannerConfigs = map[string]scannerStyle{
	"blocks": {
		active:   scannerChar{ch: "■", fading: 1},
		inactive: scannerChar{ch: "⬝", fading: 0.3},
		trail:    []scannerChar{{ch: "■", fading: 0.8}, {ch: "■", fading: 0.6}, {ch: "■", fading: 0.5}},
	},
	"diamonds": {
		active:   scannerChar{ch: "⬥", fading: 1},
		inactive: scannerChar{ch: "·", fading: 0.3},
		trail:    []scannerChar{{ch: "◆", fading: 0.8}, {ch: "⬩", fading: 0.6}, {ch: "⬪", fading: 0.5}},
	},
	"circles": {
		active:   scannerChar{ch: "●", fading: 1},
		inactive: scannerChar{ch: "⬝", fading: 0.3},
		trail:    []scannerChar{{ch: "●", fading: 0.8}, {ch: "●", fading: 0.6}, {ch: "●", fading: 0.5}},
	},
}

// Scanner represents a back-and-forth scanning animation with a fading trail.
type Scanner struct {
	Animation

	width int          // display width in characters
	style scannerStyle // Scanner style

	// Animation state
	pos  int // current position (-len(trail)+1 to width+len(trail))
	dir  int // +1 (forward/right) or -1 (backward/left)
	hold int // frames remaining in hold state
}

// NewScanner creates a new Scanner widget with the specified ID, width, and character style.
// The scanner is initialized in a stopped state and must be explicitly started.
//
//		 Parameters:
//	  - id: Unique identifier for the scanner widget
//	  - width: Display width in characters (must be >= 1)
//	  - style: Character set style, key of scannerConfigs
//
// Returns:
//   - *Scanner: Configured scanner widget ready for use
//
// Note: If an invalid charStyle is provided, "blocks" is used as default.
func NewScanner(id string, width int, style string) *Scanner {
	if width < 1 {
		width = 1
	}
	if _, ok := scannerConfigs[style]; !ok {
		style = "blocks"
	}

	s := &Scanner{
		Animation: Animation{
			Component: Component{id: id},
			stop:      make(chan struct{}),
		},
		width: width,
		style: scannerConfigs[style],
		pos:   0,
		dir:   1,
		hold:  10, // initial hold at start
	}
	s.fn = s.Tick
	return s
}

// Hint returns the preferred size for the scanner widget.
// The scanner prefers a width equal to its configured width and height of 1.
func (s *Scanner) Hint() (int, int) {
	return s.width, 1
}

// Refresh triggers a redraw of the scanner widget.
func (s *Scanner) Refresh() {
	Redraw(s)
}

// Tick updates the animation state on each frame.
const (
	holdStartFrames = 10
	holdEndFrames   = 5
)

func (s *Scanner) Tick() {
	if s.hold > 0 {
		s.hold--
		s.Refresh()
		return
	}

	// Move position
	s.pos += s.dir

	// Check boundaries
	if s.pos >= s.width+len(s.style.trail) {
		s.pos = s.width
		s.dir = -1
		s.hold = holdEndFrames
	} else if s.pos < -len(s.style.trail) {
		s.pos = -1
		s.dir = 1
		s.hold = holdStartFrames
	}

	s.Refresh()
}

// Render draws the scanner widget with its current animation frame.
func (s *Scanner) Render(r *Renderer) {
	s.Component.Render(r)

	x, y, w, h := s.Content()
	if w < 1 || h < 1 {
		return
	}

	// Get styling
	style := s.Style()
	baseColor := r.theme.Color(style.Foreground())
	bgColor := style.Background()
	font := style.Font()

	// Calculate colors for scanner parts based on fading values
	activeColor := dimColor(baseColor, s.style.active.fading)
	inactiveColor := dimColor(baseColor, s.style.inactive.fading)
	trailColors := make([]string, len(s.style.trail))
	for i, tc := range s.style.trail {
		trailColors[i] = dimColor(baseColor, tc.fading)
	}

	// Determine the actual number of columns to draw (limited by content width)
	limit := s.width
	if limit > w {
		limit = w
	}

	// Draw each column
	for col := 0; col < limit; col++ {
		var ch string
		var fg string

		if col == s.pos {
			ch = s.style.active.ch
			fg = activeColor
		} else {
			found := false
			for i, tc := range s.style.trail {
				trailPos := s.pos - (i+1)*s.dir
				if col == trailPos {
					ch = tc.ch
					fg = trailColors[i]
					found = true
					break
				}
			}
			if !found {
				ch = s.style.inactive.ch
				fg = inactiveColor
			}
		}

		r.Set(fg, bgColor, font)
		r.Put(x+col, y, ch)
	}
}

// dimColor takes a color string (hex #RRGGBB) and a factor (0-1)
// and returns a dimmed version by reducing RGB brightness.
func dimColor(hex string, factor float64) string {
	if !strings.HasPrefix(hex, "#") || len(hex) != 7 {
		return hex
	}
	r, g, b := parseHex(hex)
	dr := uint8(math.Round(float64(r) * factor))
	dg := uint8(math.Round(float64(g) * factor))
	db := uint8(math.Round(float64(b) * factor))
	return fmtHex(dr, dg, db)
}

// parseHex converts #RRGGBB to (r,g,b) uint8 values.
func parseHex(hex string) (uint8, uint8, uint8) {
	var r, g, b uint64
	if len(hex) == 7 {
		r, _ = strconv.ParseUint(hex[1:3], 16, 8)
		g, _ = strconv.ParseUint(hex[3:5], 16, 8)
		b, _ = strconv.ParseUint(hex[5:7], 16, 8)
	}
	return uint8(r), uint8(g), uint8(b)
}

// fmtHex converts (r,g,b) to #RRGGBB string.
func fmtHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

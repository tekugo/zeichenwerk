package next

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// scannerChars maps style names to the character used for scanner display.
var scannerChars = map[string]string{
	"blocks":   "█",
	"diamonds": "⬥",
}

// Scanner represents a back-and-forth scanning animation with a fading trail.
type Scanner struct {
	Component

	width     int    // display width in characters
	charStyle string // "blocks" or "diamonds"

	// Animation state
	pos    int // current position (0 to width-1)
	dir    int // +1 (forward/right) or -1 (backward/left)
	hold   int // frames remaining in hold state
	ticker *time.Ticker
	stop   chan struct{}
}

// NewScanner creates a new Scanner widget with the specified ID, width, and character style.
// The scanner is initialized in a stopped state and must be explicitly started.
//
// Parameters:
//   - id: Unique identifier for the scanner widget
//   - width: Display width in characters (must be >= 1)
//   - charStyle: Character set style, either "blocks" or "diamonds"
//
// Returns:
//   - *Scanner: Configured scanner widget ready for use
//
// Note: If an invalid charStyle is provided, "blocks" is used as default.
func NewScanner(id string, width int, charStyle string) *Scanner {
	if width < 1 {
		width = 1
	}
	if _, ok := scannerChars[charStyle]; !ok {
		charStyle = "blocks"
	}

	s := &Scanner{
		Component: Component{id: id},
		width:     width,
		charStyle: charStyle,
		pos:       0,
		dir:       1,
		hold:      10, // initial hold at start
		stop:      make(chan struct{}),
	}
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

// Start begins the scanner animation with the specified time interval.
// The animation runs in a separate goroutine. If the scanner is already running,
// this method does nothing.
func (s *Scanner) Start(interval time.Duration) {
	go func() {
		if s.ticker != nil {
			return // already running
		}

		s.ticker = time.NewTicker(interval)
		defer s.ticker.Stop()

		for {
			select {
			case <-s.stop:
				s.ticker = nil
				return
			case <-s.ticker.C:
				s.tick()
			}
		}
	}()
}

// Stop gracefully halts the scanner animation.
func (s *Scanner) Stop() {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	s.ticker = nil
}

// Running returns whether the scanner animation is currently active.
func (s *Scanner) Running() bool {
	return s.ticker != nil
}

// tick updates the animation state on each frame.
const (
	holdStartFrames = 10
	holdEndFrames   = 5
)

func (s *Scanner) tick() {
	if s.hold > 0 {
		s.hold--
		s.Refresh()
		return
	}

	// Move position
	s.pos += s.dir

	// Check boundaries
	if s.pos >= s.width-1 {
		s.pos = s.width - 1
		s.dir = -1
		s.hold = holdEndFrames
	} else if s.pos <= 0 {
		s.pos = 0
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

	// Character to use for active and trail
	ch := scannerChars[s.charStyle]

	// Draw the scanner width, left-aligned within content area
	for i := 0; i < s.width; i++ {
		var fg string

		if i == s.pos {
			// Active position: full brightness
			fg = baseColor
		} else if (s.dir == 1 && i < s.pos) || (s.dir == -1 && i > s.pos) {
			// Position is behind the active scanner (in the trail)
			dist := s.pos - i
			if dist < 0 {
				dist = -dist
			}
			// Compute fade factor based on distance (1.0 at active, down to 0.2 at far end)
			maxDist := s.width - 1
			if maxDist == 0 {
				maxDist = 1
			}
			factor := 1.0 - float64(dist)/float64(maxDist)
			if factor < 0.2 {
				factor = 0.2
			}
			fg = dimColor(baseColor, factor)
		} else {
			// Empty space (ahead of active direction)
			fg = baseColor
		}

		// Determine character to draw: block for active/trail, space otherwise
		drawCh := " "
		if i == s.pos || ((s.dir == 1 && i < s.pos) || (s.dir == -1 && i > s.pos)) {
			drawCh = ch
		}

		r.Set(fg, bgColor, font)
		r.Put(x+i, y, drawCh)
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
	return "#" + strconv.FormatUint(uint64(r), 16) + strconv.FormatUint(uint64(g), 16) + strconv.FormatUint(uint64(b), 16)
}

package next

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// scannerConfig defines character sets for each scanner style.
type scannerConfig struct {
	active   string   // character for active position (index 0)
	trail    []string // characters for trail positions (index 1,2,...)
	inactive string   // character for inactive (outside trail)
}

var scannerConfigs = map[string]scannerConfig{
	"blocks": {
		active:   "■",
		trail:    []string{"■"}, // all trail positions use same block
		inactive: "⬝",
	},
	"diamonds": {
		active:   "⬥",
		trail:    []string{"◆", "⬩", "⬪"}, // trail steps 1,2,3
		inactive: "·",
	},
}

// Scanner represents a back-and-forth scanning animation with a fading trail.
type Scanner struct {
	Animation

	width     int    // display width in characters
	charStyle string // "blocks" or "diamonds"

	// Animation state
	pos  int // current position (0 to width-1)
	dir  int // +1 (forward/right) or -1 (backward/left)
	hold int // frames remaining in hold state
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
	if _, ok := scannerConfigs[charStyle]; !ok {
		charStyle = "blocks"
	}

	s := &Scanner{
		Animation: Animation{
			Component: Component{id: id},
			stop:      make(chan struct{}),
		},
		width:     width,
		charStyle: charStyle,
		pos:       0,
		dir:       1,
		hold:      10, // initial hold at start
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
	s.Animation.Refresh()
}

// Start begins the scanner animation with the specified time interval.
func (s *Scanner) Start(interval time.Duration) {
	s.Animation.Start(interval)
}

// Stop gracefully halts the scanner animation.
func (s *Scanner) Stop() {
	s.Animation.Stop()
}

// Running returns whether the scanner animation is currently active.
func (s *Scanner) Running() bool {
	return s.Animation.Running()
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

	// Get character set
	cfg := scannerConfigs[s.charStyle]

	// Precomputed color factors for trail gradient (6 steps total, index 0..5)
	colorFactors := []float64{1.0, 0.9, 0.7, 0.5, 0.3, 0.2}
	inactiveFactor := 0.15

	// Draw the scanner width, left-aligned within content area
	for i := 0; i < s.width; i++ {
		var fg string
		var ch string

		// Compute directional distance: positive if position is behind active in movement direction
		dist := 0
		if i == s.pos {
			dist = 0
		} else if (s.dir == 1 && i < s.pos) || (s.dir == -1 && i > s.pos) {
			// Behind active: calculate absolute distance
			dist = s.pos - i
			if dist < 0 {
				dist = -dist
			}
		} else {
			// Ahead of active or outside trail -> inactive
			dist = -1
		}

		switch {
		case dist == 0:
			// Active position
			fg = baseColor
			ch = cfg.active
		case dist > 0 && dist <= len(colorFactors):
			// Trail position: pick color based on distance index
			idx := dist
			if idx > len(colorFactors)-1 {
				idx = len(colorFactors) - 1
			}
			factor := colorFactors[idx]
			fg = dimColor(baseColor, factor)
			// Choose character
			if s.charStyle == "diamonds" {
				// For diamonds, use trail characters array; index into trail (dist-1) but clamp to array length-1
				trailIdx := dist - 1
				if trailIdx >= len(cfg.trail) {
					trailIdx = len(cfg.trail) - 1
				}
				if trailIdx < 0 {
					trailIdx = 0
				}
				ch = cfg.trail[trailIdx]
			} else {
				// blocks: use same active character for all trail positions
				ch = cfg.active
			}
		default:
			// Inactive positions (outside trail)
			fg = dimColor(baseColor, inactiveFactor)
			ch = cfg.inactive
		}

		r.Set(fg, bgColor, font)
		r.Put(x+i, y, ch)
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

package zeichenwerk


type scannerChar struct {
	ch     string  // Render character
	fading float64 // color fading factor (dimming)
	color  string  // Hex colors string, calculated automatically, if fading is > 0
}

// scannerConfig defines character sets for each scanner style.
type scannerConfig struct {
	active   scannerChar   // character and color for the moving scanner
	inactive scannerChar   // character and color for inactive (outside trail)
	trail    []scannerChar // character and color for the scanner trail
}

// Tick updates the animation state on each frame.
const (
	holdStartFrames = 10
	holdEndFrames   = 5
)

var scannerConfigs = map[string]scannerConfig{
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

	width  int           // display width in characters
	config scannerConfig // Scanner configuration/style

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
//	  - class: Style class
//	  - width: Display width in characters (must be >= 1)
//	  - style: Character set style, key of scannerConfigs
//
// Returns:
//   - *Scanner: Configured scanner widget ready for use
//
// Note: If an invalid charStyle is provided, "blocks" is used as default.
func NewScanner(id, class string, width int, style string) *Scanner {
	if width < 1 {
		width = 1
	}
	if _, ok := scannerConfigs[style]; !ok {
		style = "blocks"
	}

	s := &Scanner{
		Animation: Animation{
			Component: Component{id: id, class: class},
			stop:      make(chan struct{}),
		},
		width:  width,
		config: scannerConfigs[style],
		pos:    0,
		dir:    1,
		hold:   10, // initial hold at start
	}
	s.fn = s.Tick
	return s
}

// Apply applies a theme's styles to the component.
func (s *Scanner) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("scanner"))
}

// Hint returns the preferred size for the scanner widget.
// The scanner prefers a width equal to its configured width and height of 1.
func (s *Scanner) Hint() (int, int) {
	if s.hwidth != 0 || s.hheight != 0 {
		return s.hwidth, s.hheight
	}
	return s.width, 1
}

// Refresh triggers a redraw of the scanner widget.
func (s *Scanner) Refresh() {
	Redraw(s)
}

func (s *Scanner) Tick() {
	if s.hold > 0 {
		s.hold--
		s.Refresh()
		return
	}

	// Move position
	s.pos += s.dir

	// Check boundaries
	if s.pos >= s.width+len(s.config.trail) {
		s.pos = s.width
		s.dir = -1
		s.hold = holdEndFrames
	} else if s.pos < -len(s.config.trail) {
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
	activeColor := dimColor(baseColor, s.config.active.fading)
	inactiveColor := dimColor(baseColor, s.config.inactive.fading)
	trailColors := make([]string, len(s.config.trail))
	for i, tc := range s.config.trail {
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
			ch = s.config.active.ch
			fg = activeColor
		} else {
			found := false
			for i, tc := range s.config.trail {
				trailPos := s.pos - (i+1)*s.dir
				if col == trailPos {
					ch = tc.ch
					fg = trailColors[i]
					found = true
					break
				}
			}
			if !found {
				ch = s.config.inactive.ch
				fg = inactiveColor
			}
		}

		r.Set(fg, bgColor, font)
		r.Put(x+col, y, ch)
	}
}


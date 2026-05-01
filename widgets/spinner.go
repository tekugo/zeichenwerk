package widgets

import (
	"strings"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// Spinners provides a collection of predefined spinner animations as strings.
// Each string contains Unicode characters that create different visual effects
// when cycled through rapidly.
//
// Available spinner styles:
//   - "bar": Classic rotating bar animation (|/-\)
//   - "dots": Growing and shrinking dots (.oOo)
//   - "dot": Braille-based single dot moving in a circle
//   - "arrow": Arrow rotating through 8 directions
//   - "circle": Quarter-circle rotation animation
//   - "bounce": Simple braille bouncing effect
//   - "braille": Complex braille pattern creating a spinning effect
var Spinners = map[string]string{
	"bar":     "| / - \\",
	"dots":    ". o O o",
	"dot":     "⠁ ⠂ ⠄ ⡀ ⢀ ⠠ ⠐ ⠈",
	"arrow":   "← ↖ ↑ ↗ → ↘ ↓ ↙",
	"circle":  "◐ ◓ ◑ ◒",
	"bounce":  "⠁ ⠂ ⠄ ⠂",
	"braille": "⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏",
}

// Spinner represents an animated loading indicator widget that cycles through
// a sequence of Unicode strings to create visual motion effects. Spinners
// are commonly used to indicate background processing or loading states.
type Spinner struct {
	Animation
	sequence []string // Sequence of Unicode strings to cycle through
	index    int      // Current position in the runes array
}

// NewSpinner creates a new spinner widget with the specified character sequence.
// The spinner is initialized in a stopped state and must be explicitly started.
//
// Parameters:
//   - id: Unique identifier for the spinner widget
//   - class: Style class
//   - sequence: Sequence of Unicode characters to cycle through for animation
//
// Returns:
//   - *Spinner: Configured spinner widget ready for use
//
// Note: The spinner is not focusable and starts with the first character
// in the sequence. An empty runes slice will cause runtime panics.
func NewSpinner(id, class string, sequence string) *Spinner {
	spinner := &Spinner{
		Animation: Animation{
			Component: Component{id: id, class: class},
			stop:      make(chan struct{}),
		},
		sequence: strings.Split(sequence, " "),
	}
	spinner.fn = spinner.Tick
	return spinner
}

// Apply applies a theme's styles to the component.
func (s *Spinner) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("spinner"))
}

// Hint returns the preferred size for the spinner widget.
// Spinners always prefer a 1x1 character size since they display
// a single animated character at a time.
//
// Returns:
//   - int: Preferred width (always 1)
//   - int: Preferred height (always 1)
func (s *Spinner) Hint() (int, int) {
	if s.hwidth != 0 || s.hheight != 0 {
		return s.hwidth, s.hheight
	}
	max := 1
	for _, s := range s.sequence {
		if len(s) > max {
			max = len(s)
		}
	}
	return max, 1
}

// Refresh and thus redraw the spinner widget.
func (s *Spinner) Refresh() {
	Redraw(s)
}

// Tick updates the animation state on each frame.
func (s *Spinner) Tick() {
	s.index++
	if s.index >= len(s.sequence) {
		s.index = 0
	}
	s.Refresh()
}

// SetSequence updates the character sequence used for animation.
// If the spinner is currently running, it continues with the new sequence.
// The current index is reset to 0 to avoid potential out-of-bounds issues.
func (s *Spinner) SetSequence(sequence string) {
	s.sequence = strings.Split(sequence, " ")
	s.index = 0 // Reset to avoid out-of-bounds
	if s.Running() {
		s.Refresh() // Update display immediately if running
	}
}

// Current returns the currently displayed string from the animation sequence.
// This method is called by the renderer to get the character to display.
func (s *Spinner) Current() string {
	return s.sequence[s.index]
}

// Render draws the spinner widget.
func (s *Spinner) Render(r *Renderer) {
	s.Component.Render(r)

	x, y, w, h := s.Content()
	if w < 1 || h < 1 {
		return
	}
	style := s.Style()
	r.Set(style.Foreground(), style.Background(), style.Font())
	r.Text(x, y, s.Current(), 0)
}

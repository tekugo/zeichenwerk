package next

import (
	"strings"
	"time"
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
	Component
	sequence []string      // Sequence of Unicode strings to cycle through
	index    int           // Current position in the runes array
	ticker   *time.Ticker  // Timer for controlling animation speed
	stop     chan struct{} // Channel for signaling animation stop
}

// NewSpinner creates a new spinner widget with the specified character sequence.
// The spinner is initialized in a stopped state and must be explicitly started.
//
// Parameters:
//   - id: Unique identifier for the spinner widget
//   - sequence: Sequence of Unicode characters to cycle through for animation
//
// Returns:
//   - *Spinner: Configured spinner widget ready for use
//
// Note: The spinner is not focusable and starts with the first character
// in the sequence. An empty runes slice will cause runtime panics.
func NewSpinner(id string, sequence string) *Spinner {
	spinner := &Spinner{
		Component: Component{id: id},
		sequence:  strings.Split(sequence, " "),
		stop:      make(chan struct{}),
	}
	return spinner
}

// Hint returns the preferred size for the spinner widget.
// Spinners always prefer a 1x1 character size since they display
// a single animated character at a time.
//
// Returns:
//   - int: Preferred width (always 1)
//   - int: Preferred height (always 1)
func (s *Spinner) Hint() (int, int) {
	max := 1
	for _, s := range s.sequence {
		if len(s) > max {
			max = len(s)
		}
	}
	return max, 1
}

// Refresh triggers a redraw of the spinner widget.
// This method is called automatically during animation cycles
// but can be called manually if needed.
func (s *Spinner) Refresh() {
	Redraw(s)
}

// Start begins the spinner animation with the specified time interval.
// The animation runs in a separate goroutine to avoid blocking the main thread.
// Each tick advances to the next character in the sequence and triggers a redraw.
//
// Note: The spinner must be stopped before starting again. Multiple concurrent
// animations on the same spinner are not supported and will cause a panic.
func (s *Spinner) Start(interval time.Duration) {
	// Starting the spinner will block, so we start it as a separate go routine
	go func() {
		if s.ticker != nil {
			panic("ticker for spinner already started")
		}

		s.ticker = time.NewTicker(interval)
		defer s.ticker.Stop()

		for {
			select {
			case <-s.stop:
				s.Log(s, "debug", "Spinner stopped")
				s.ticker = nil
				return
			case <-s.ticker.C:
				s.index++
				if s.index >= len(s.sequence) {
					s.index = 0
				}
				s.Refresh()
			}
		}
	}()
}

// Current returns the currently displayed string from the animation sequence.
// This method is called by the renderer to get the character to display.
func (s *Spinner) Current() string {
	return s.sequence[s.index]
}

// Stop gracefully halts the spinner animation and cleans up resources.
// The method is thread-safe and can be called multiple times without issues.
// After stopping, the spinner can be restarted with Start().
func (s *Spinner) Stop() {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	s.ticker = nil
}

// Running returns whether the spinner animation is currently active.
// This is useful for checking spinner state before starting or stopping.
func (s *Spinner) Running() bool {
	return s.ticker != nil
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

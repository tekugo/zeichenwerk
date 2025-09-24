package zeichenwerk

import "time"

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
//
// Usage:
//
//	spinner := NewSpinner("loading", []rune(Spinners["braille"]))
//	spinner.Start(100 * time.Millisecond)
var Spinners = map[string]string{
	"bar":     "|/-\\",
	"dots":    ".oOo",
	"dot":     "⠁⠂⠄⡀⢀⠠⠐⠈",
	"arrow":   "←↖↑↗→↘↓↙",
	"circle":  "◐◓◑◒",
	"bounce":  "⠁⠂⠄⠂",
	"braille": "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏",
}

// Spinner represents an animated loading indicator widget that cycles through
// a sequence of Unicode characters to create visual motion effects. Spinners
// are commonly used to indicate background processing or loading states.
//
// The spinner widget features:
//   - Customizable character sequences for different visual effects
//   - Configurable animation speed via time intervals
//   - Non-blocking operation using goroutines
//   - Automatic cleanup and resource management
//   - Thread-safe start/stop controls
//
// Spinners are not focusable and typically have a fixed 1x1 character size.
// They can be styled like other widgets and support borders and backgrounds.
//
// Usage Example:
//
//	// Create a spinner with braille animation
//	spinner := NewSpinner("loader", []rune(Spinners["braille"]))
//
//	// Start animation at 100ms intervals
//	spinner.Start(100 * time.Millisecond)
//
//	// Stop animation when done
//	defer spinner.Stop()
//
//	// Use with builder pattern
//	ui := NewBuilder(DefaultTheme()).
//		Flex("container", "horizontal", "center", 0).
//		Label("status", "Loading...", 0).
//		Add(spinner).
//		Build()
type Spinner struct {
	BaseWidget
	runes  []rune        // Sequence of Unicode characters to cycle through
	index  int           // Current position in the runes array
	ticker *time.Ticker  // Timer for controlling animation speed
	stop   chan struct{} // Channel for signaling animation stop
}

// NewSpinner creates a new spinner widget with the specified character sequence.
// The spinner is initialized in a stopped state and must be explicitly started.
//
// Parameters:
//   - id: Unique identifier for the spinner widget
//   - runes: Slice of Unicode characters to cycle through for animation
//
// Returns:
//   - *Spinner: Configured spinner widget ready for use
//
// Example:
//
//	// Using predefined spinner styles
//	spinner1 := NewSpinner("loader1", []rune(Spinners["braille"]))
//	spinner2 := NewSpinner("loader2", []rune(Spinners["circle"]))
//
//	// Using custom character sequence
//	custom := []rune("◢◣◤◥")
//	spinner3 := NewSpinner("custom", custom)
//
// Note: The spinner is not focusable and starts with the first character
// in the sequence. An empty runes slice will cause runtime panics.
func NewSpinner(id string, runes []rune) *Spinner {
	spinner := &Spinner{
		BaseWidget: BaseWidget{id: id, focusable: false},
		runes:      runes,
		stop:       make(chan struct{}),
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
	return 1, 1
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
// Parameters:
//   - interval: Time duration between animation frames (e.g., 100*time.Millisecond)
//
// Behavior:
//   - Starts a new goroutine for non-blocking animation
//   - Cycles through the character sequence indefinitely
//   - Automatically wraps to the first character after the last
//   - Can be stopped at any time using Stop()
//   - Panics if called when already running (use Stop() first)
//
// Example:
//
//	spinner := NewSpinner("loader", []rune(Spinners["braille"]))
//	spinner.Start(100 * time.Millisecond)  // 10 FPS animation
//
//	// For faster animation
//	spinner.Start(50 * time.Millisecond)   // 20 FPS animation
//
//	// For slower animation
//	spinner.Start(200 * time.Millisecond)  // 5 FPS animation
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
				if s.index >= len(s.runes) {
					s.index = 0
				}
				s.Refresh()
			}
		}
	}()
}

// Rune returns the currently displayed character from the animation sequence.
// This method is called by the renderer to get the character to display.
//
// Returns:
//   - rune: The Unicode character at the current animation position
//
// Note: This method will panic if called with an empty runes slice or
// if the internal index becomes invalid. This should not happen under
// normal usage.
func (s *Spinner) Rune() rune {
	return s.runes[s.index]
}

// Stop gracefully halts the spinner animation and cleans up resources.
// The method is thread-safe and can be called multiple times without issues.
// After stopping, the spinner can be restarted with Start().
//
// Behavior:
//   - Sends a stop signal to the animation goroutine
//   - Non-blocking: returns immediately even if animation is still stopping
//   - Safe to call multiple times or when already stopped
//   - Automatically cleans up the internal ticker
//   - Preserves the current character position for potential restart
//
// Example:
//
//	spinner := NewSpinner("loader", []rune(Spinners["circle"]))
//	spinner.Start(100 * time.Millisecond)
//
//	// Stop after some work
//	time.Sleep(2 * time.Second)
//	spinner.Stop()
//
//	// Can restart later
//	spinner.Start(200 * time.Millisecond)
//
// Note: It's good practice to defer Stop() when creating spinners to ensure
// cleanup even if the program exits unexpectedly.
func (s *Spinner) Stop() {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	s.ticker = nil
}

// IsRunning returns whether the spinner animation is currently active.
// This is useful for checking spinner state before starting or stopping.
//
// Returns:
//   - bool: true if animation is running, false if stopped
//
// Example:
//
//	if !spinner.IsRunning() {
//	    spinner.Start(100 * time.Millisecond)
//	}
func (s *Spinner) IsRunning() bool {
	return s.ticker != nil
}

// SetRunes updates the character sequence used for animation.
// If the spinner is currently running, it continues with the new sequence.
// The current index is reset to 0 to avoid potential out-of-bounds issues.
//
// Parameters:
//   - runes: New sequence of Unicode characters for animation
//
// Example:
//
//	spinner := NewSpinner("loader", []rune(Spinners["bar"]))
//	spinner.Start(100 * time.Millisecond)
//
//	// Switch to a different animation style
//	spinner.SetRunes([]rune(Spinners["braille"]))
//
// Note: An empty runes slice will cause runtime panics when Rune() is called.
func (s *Spinner) SetRunes(runes []rune) {
	s.runes = runes
	s.index = 0 // Reset to avoid out-of-bounds
	if s.IsRunning() {
		s.Refresh() // Update display immediately if running
	}
}

// GetCurrentIndex returns the current position in the animation sequence.
// This can be useful for debugging or for implementing custom animation logic.
//
// Returns:
//   - int: Zero-based index of the current character in the sequence
func (s *Spinner) GetCurrentIndex() int {
	return s.index
}

// Reset sets the animation back to the first character in the sequence.
// This does not affect whether the spinner is running or stopped.
//
// Example:
//
//	spinner.Reset() // Back to first character
//	if !spinner.IsRunning() {
//	    spinner.Start(100 * time.Millisecond)
//	}
func (s *Spinner) Reset() {
	s.index = 0
	if s.IsRunning() {
		s.Refresh()
	}
}

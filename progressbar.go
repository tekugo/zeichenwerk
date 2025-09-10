package zeichenwerk

import (
	"fmt"
)

// ProgressBar represents a progress indicator widget that displays the completion
// status of a task or operation. It provides visual feedback to users about
// ongoing processes through a filled bar that represents progress percentage.
//
// Features:
//   - Configurable value range (min/max)
//   - Horizontal and vertical orientations
//   - Determinate mode (shows specific progress percentage)
//   - Indeterminate mode (shows activity without specific progress)
//   - Automatic value clamping to valid range
//   - Customizable styling through the theme system
//
// The progress bar supports both determinate progress (when you know the total
// amount of work) and indeterminate progress (when you only know that work is
// happening but not how much remains).
type ProgressBar struct {
	BaseWidget
	Value         int    // Current progress value within the Min-Max range
	Min           int    // Minimum value of the progress range (typically 0)
	Max           int    // Maximum value of the progress range (typically 100)
	Orientation   string // Display orientation: "horizontal" or "vertical"
	Indeterminate bool   // Enable indeterminate/pulsing mode for unknown progress
}

// NewProgressBar creates a new progress bar widget with default settings.
// The progress bar is initialized with a 0-100 range, starting at 0,
// and configured for horizontal display in determinate mode.
//
// Parameters:
//   - id: Unique identifier for the progress bar widget
//
// Returns:
//   - *ProgressBar: A new progress bar widget instance
//
// Default configuration:
//   - Value: 0 (no progress)
//   - Range: 0 to 100
//   - Orientation: horizontal
//   - Mode: determinate
//
// Example usage:
//
//	progress := NewProgressBar("download-progress")
//	progress.SetValue(50)  // 50% complete
func NewProgressBar(id string) *ProgressBar {
	return &ProgressBar{
		BaseWidget:  BaseWidget{id: id},
		Value:       0,
		Min:         0,
		Max:         100,
		Orientation: "horizontal",
	}
}

// SetValue sets the current progress value with automatic range clamping.
// The value will be automatically constrained to the valid Min-Max range
// to prevent invalid progress states.
//
// Parameters:
//   - value: The new progress value (will be clamped to Min-Max range)
//
// Behavior:
//   - Values below Min are set to Min
//   - Values above Max are set to Max
//   - Valid values are set as-is
//
// Example:
//
//	progress.SetRange(0, 100)
//	progress.SetValue(150)  // Will be clamped to 100
//	progress.SetValue(-10)  // Will be clamped to 0
func (p *ProgressBar) SetValue(value int) {
	if value < p.Min {
		value = p.Min
	}
	if value > p.Max {
		value = p.Max
	}
	p.Value = value
}

// SetRange sets the minimum and maximum values for the progress bar.
// This defines the valid range for progress values and automatically
// adjusts the current value if it falls outside the new range.
//
// Parameters:
//   - min: The minimum value for the progress range
//   - max: The maximum value for the progress range
//
// Behavior:
//   - If min > max, the values are automatically swapped
//   - The current Value is re-clamped to the new range
//   - Progress percentage is recalculated based on the new range
//
// Example:
//
//	progress.SetRange(0, 1000)    // Range: 0-1000
//	progress.SetRange(50, 10)     // Automatically becomes 10-50
func (p *ProgressBar) SetRange(min, max int) {
	if min > max {
		min, max = max, min
	}
	p.Min = min
	p.Max = max
	p.SetValue(p.Value) // Clamp current value to new range
}

// SetOrientation sets the display orientation of the progress bar.
// This controls whether the progress bar is displayed horizontally or vertically.
//
// Parameters:
//   - orientation: "horizontal" or "vertical"
//
// Note: Invalid orientation values will be ignored, keeping the current setting.
func (p *ProgressBar) SetOrientation(orientation string) {
	if orientation == "horizontal" || orientation == "vertical" {
		p.Orientation = orientation
	}
}

// SetIndeterminate sets the indeterminate mode for the progress bar.
// In indeterminate mode, the progress bar shows activity without indicating
// specific progress, useful when the total work amount is unknown.
//
// Parameters:
//   - indeterminate: true for indeterminate mode, false for determinate mode
func (p *ProgressBar) SetIndeterminate(indeterminate bool) {
	p.Indeterminate = indeterminate
}

// Percentage returns the current progress as a percentage (0-100).
// This calculates the completion percentage based on the current value
// and the Min-Max range.
//
// Returns:
//   - float64: Progress percentage (0.0 to 100.0)
//
// Example:
//
//	progress.SetRange(0, 200)
//	progress.SetValue(50)
//	pct := progress.Percentage()  // Returns 25.0
func (p *ProgressBar) Percentage() float64 {
	if p.Max == p.Min {
		return 0.0
	}
	return float64(p.Value-p.Min) / float64(p.Max-p.Min) * 100.0
}

// IsComplete returns true if the progress bar has reached its maximum value.
// This is useful for checking if a task or operation has finished.
//
// Returns:
//   - bool: true if Value equals Max, false otherwise
func (p *ProgressBar) IsComplete() bool {
	return p.Value == p.Max
}

// Reset resets the progress bar to its minimum value.
// This is useful for restarting progress tracking for a new operation.
func (p *ProgressBar) Reset() {
	p.Value = p.Min
}

// Increment increases the progress value by the specified amount.
// The result is automatically clamped to the valid range.
//
// Parameters:
//   - amount: The amount to add to the current value
//
// Example:
//
//	progress.SetValue(10)
//	progress.Increment(5)  // Value becomes 15
func (p *ProgressBar) Increment(amount int) {
	p.SetValue(p.Value + amount)
}

// Info returns a human-readable description of the progress bar's configuration.
// This includes the orientation and current state information, useful for
// debugging and development purposes.
//
// Returns:
//   - string: Formatted string with progress bar information
func (p *ProgressBar) Info() string {
	return fmt.Sprintf("ProgressBar(%s)", p.Orientation)
}

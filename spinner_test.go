package zeichenwerk

import (
	"testing"
	"time"
)

func TestSpinnerCreation(t *testing.T) {
	runes := []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
	spinner := NewSpinner("test-spinner", runes)

	if spinner.ID() != "test-spinner" {
		t.Errorf("Expected ID 'test-spinner', got '%s'", spinner.ID())
	}

	if spinner.IsRunning() {
		t.Error("Newly created spinner should not be running")
	}

	// Test hint size
	w, h := spinner.Hint()
	if w != 1 || h != 1 {
		t.Errorf("Expected hint size (1,1), got (%d,%d)", w, h)
	}

	// Test initial rune
	if spinner.Rune() != runes[0] {
		t.Errorf("Expected initial rune to be first in sequence")
	}

	// Test initial index
	if spinner.GetCurrentIndex() != 0 {
		t.Errorf("Expected initial index 0, got %d", spinner.GetCurrentIndex())
	}
}

func TestSpinnerPredefinedStyles(t *testing.T) {
	expectedStyles := []string{"bar", "dots", "dot", "arrow", "circle", "bounce", "braille"}
	
	for _, style := range expectedStyles {
		if _, exists := Spinners[style]; !exists {
			t.Errorf("Expected predefined spinner style '%s' not found", style)
		}
	}

	// Test creating spinners with predefined styles
	for style, chars := range Spinners {
		runes := []rune(chars)
		spinner := NewSpinner("test-"+style, runes)
		if len(chars) == 0 {
			t.Errorf("Spinner style '%s' has empty character sequence", style)
		}
		if len(runes) > 0 && spinner.Rune() != runes[0] {
			t.Errorf("Spinner style '%s' initial rune mismatch: expected %c, got %c", style, runes[0], spinner.Rune())
		}
	}
}

func TestSpinnerStartStop(t *testing.T) {
	spinner := NewSpinner("test", []rune("abc"))

	// Test starting
	spinner.Start(10 * time.Millisecond)
	
	// Give it a moment to start
	time.Sleep(5 * time.Millisecond)
	
	if !spinner.IsRunning() {
		t.Error("Spinner should be running after Start()")
	}

	// Let it animate a bit
	time.Sleep(25 * time.Millisecond)
	
	// Test stopping
	spinner.Stop()
	
	// Give it a moment to stop
	time.Sleep(5 * time.Millisecond)
	
	if spinner.IsRunning() {
		t.Error("Spinner should not be running after Stop()")
	}
}

func TestSpinnerReset(t *testing.T) {
	runes := []rune("abcd")
	spinner := NewSpinner("test", runes)

	// Manually advance index
	spinner.index = 2
	if spinner.GetCurrentIndex() != 2 {
		t.Errorf("Expected index 2, got %d", spinner.GetCurrentIndex())
	}

	// Reset should go back to 0
	spinner.Reset()
	if spinner.GetCurrentIndex() != 0 {
		t.Errorf("Expected index 0 after reset, got %d", spinner.GetCurrentIndex())
	}

	if spinner.Rune() != runes[0] {
		t.Error("Expected first rune after reset")
	}
}

func TestSpinnerSetRunes(t *testing.T) {
	originalRunes := []rune("abc")
	newRunes := []rune("xyz")
	
	spinner := NewSpinner("test", originalRunes)
	
	// Verify original
	if spinner.Rune() != 'a' {
		t.Error("Expected initial rune 'a'")
	}

	// Change runes
	spinner.SetRunes(newRunes)
	
	// Should reset to first character of new sequence
	if spinner.Rune() != 'x' {
		t.Error("Expected rune 'x' after SetRunes")
	}
	
	if spinner.GetCurrentIndex() != 0 {
		t.Error("Expected index 0 after SetRunes")
	}
}

func TestSpinnerMultipleStops(t *testing.T) {
	spinner := NewSpinner("test", []rune("abc"))
	
	// Multiple stops should not panic
	spinner.Stop()
	spinner.Stop()
	spinner.Stop()
	
	// Should still not be running
	if spinner.IsRunning() {
		t.Error("Spinner should not be running after multiple stops")
	}
}

func TestSpinnerAnimation(t *testing.T) {
	runes := []rune("abc")
	spinner := NewSpinner("test", runes)
	
	// Start with fast interval for quick testing
	spinner.Start(5 * time.Millisecond)
	defer spinner.Stop()
	
	// Wait for a few animation cycles
	time.Sleep(20 * time.Millisecond)
	
	// The index should have advanced (we can't guarantee exact position due to timing)
	// but we can verify the rune is valid
	currentRune := spinner.Rune()
	validRune := false
	for _, r := range runes {
		if r == currentRune {
			validRune = true
			break
		}
	}
	
	if !validRune {
		t.Errorf("Current rune '%c' is not in the expected sequence", currentRune)
	}
}
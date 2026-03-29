package zeichenwerk

import (
	"testing"
)

func TestCanvas(t *testing.T) {
	// Create a small canvas
	c := NewCanvas("test-canvas", "", 1, 10, 5)

	// Check initial state (not focused - cursor hidden)
	x, y, cs := c.Cursor()
	if x != -1 || y != -1 || cs != "" {
		t.Errorf("Expected hidden cursor (-1,-1,'') when not focused, got (%d,%d,%s)", x, y, cs)
	}

	// Focus the canvas
	c.SetFlag(FlagFocused, true)
	x, y, cs = c.Cursor()
	if x != 0 || y != 0 {
		t.Errorf("Expected cursor at (0,0) when focused, got (%d,%d)", x, y)
	}
	// In normal mode, default cursor should be "block"
	if cs != "block" {
		t.Errorf("Expected block cursor in normal mode, got %s", cs)
	}

	// Check initial mode is normal
	if c.Mode() != ModeNormal {
		t.Errorf("Expected initial mode %s, got %s", ModeNormal, c.Mode())
	}

	// Set a cell
	style := NewStyle("").WithColors("white", "blue")
	c.SetCell(5, 2, "A", style)
	cell := c.Cell(5, 2)
	if cell == nil || cell.ch != "A" {
		t.Error("Failed to set cell at (5,2)")
	}

	// Move cursor
	c.move(2, 1)
	x, y, _ = c.Cursor()
	if x != 2 || y != 1 {
		t.Errorf("Expected cursor at (2,1), got (%d,%d)", x, y)
	}

	// Test boundary clamping
	c.move(-10, -10)
	x, y, _ = c.Cursor()
	if x != 0 || y != 0 {
		t.Errorf("Expected cursor clamped to (0,0), got (%d,%d)", x, y)
	}

	// Clear canvas
	c.Clear()
	cell = c.Cell(5, 2)
	if cell.ch != "" {
		t.Error("Clear did not remove cell content")
	}

	// Fill canvas
	c.Fill("#", NewStyle(""))
	cell = c.Cell(0, 0)
	if cell.ch != "#" {
		t.Error("Fill did not populate cells")
	}

	// Check size
	w, h := c.Size()
	if w != 10 || h != 5 {
		t.Errorf("Expected size 10x5, got %dx%d", w, h)
	}

	// Test mode switching to insert
	c.SetMode(ModeInsert)
	if c.Mode() != ModeInsert {
		t.Errorf("Expected mode %s, got %s", ModeInsert, c.Mode())
	}
	// In insert mode, cursor should be bar
	x, y, cs = c.Cursor()
	if cs != "bar" {
		t.Errorf("Expected bar cursor in insert mode, got %s", cs)
	}

	// Switch back to normal mode
	c.SetMode(ModeNormal)
	if c.Mode() != ModeNormal {
		t.Errorf("Expected mode %s, got %s", ModeNormal, c.Mode())
	}
	x, y, cs = c.Cursor()
	if cs != "block" {
		t.Errorf("Expected block cursor back in normal mode, got %s", cs)
	}
}

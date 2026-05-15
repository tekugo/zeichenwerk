package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Cell is a single character cell — one rune (as a UTF-8 string) plus the
// palette key naming its style. An empty Style means "use the default
// palette entry".
type Cell struct {
	Ch    string `json:"ch"`
	Style string `json:"style,omitempty"`
}

// EmptyCell is the representation of a blank cell.
var EmptyCell = Cell{Ch: " ", Style: ""}

// Document is the in-memory model of a malwerk drawing. The cell grid is
// indexed [y][x] with len(Cells) == Height and len(Cells[y]) == Width.
type Document struct {
	Width   int                  `json:"width"`
	Height  int                  `json:"height"`
	Palette map[string]*DocStyle `json:"palette"`
	Cells   [][]Cell             `json:"cells"`

	Path  string `json:"-"`
	Dirty bool   `json:"-"`
}

// NewDocument builds an empty document of the given size with a single
// "default" palette entry.
func NewDocument(width, height int) *Document {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	d := &Document{
		Width:   width,
		Height:  height,
		Palette: map[string]*DocStyle{"default": {Fg: "$fg0", Bg: "$bg0"}},
		Cells:   make([][]Cell, height),
	}
	for y := range d.Cells {
		row := make([]Cell, width)
		for x := range row {
			row[x] = EmptyCell
		}
		d.Cells[y] = row
	}
	return d
}

// LoadDocument reads a *.malwerk.json file and validates its shape.
func LoadDocument(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	d := &Document{}
	if err := json.Unmarshal(data, d); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if d.Width < 1 || d.Height < 1 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", d.Width, d.Height)
	}
	if len(d.Cells) != d.Height {
		return nil, fmt.Errorf("cells height = %d, want %d", len(d.Cells), d.Height)
	}
	for y, row := range d.Cells {
		if len(row) != d.Width {
			return nil, fmt.Errorf("row %d width = %d, want %d", y, len(row), d.Width)
		}
		for x, c := range row {
			if c.Ch == "" {
				d.Cells[y][x] = EmptyCell
				d.Cells[y][x].Style = c.Style
			}
		}
	}
	if d.Palette == nil {
		d.Palette = map[string]*DocStyle{}
	}
	if _, ok := d.Palette["default"]; !ok {
		d.Palette["default"] = &DocStyle{Fg: "$fg0", Bg: "$bg0"}
	}
	for y, row := range d.Cells {
		for x, c := range row {
			if c.Style == "" {
				continue
			}
			if _, ok := d.Palette[c.Style]; !ok {
				d.Cells[y][x].Style = ""
			}
		}
	}
	d.Path = path
	d.Dirty = false
	return d, nil
}

// Save writes the document to its current Path. Returns an error if Path
// is empty — caller should prompt for Save As first.
func (d *Document) Save() error {
	if d.Path == "" {
		return fmt.Errorf("document has no path; use SaveAs")
	}
	return d.SaveAs(d.Path)
}

// SaveAs writes the document to the given path and updates Path / Dirty.
func (d *Document) SaveAs(path string) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	d.Path = path
	d.Dirty = false
	return nil
}

// At returns the cell at (x, y), or EmptyCell if out of bounds.
func (d *Document) At(x, y int) Cell {
	if x < 0 || y < 0 || y >= d.Height || x >= d.Width {
		return EmptyCell
	}
	return d.Cells[y][x]
}

// Set writes a cell at (x, y). Out-of-bounds writes are silently dropped.
// Marks the document dirty when the cell actually changes.
func (d *Document) Set(x, y int, c Cell) {
	if x < 0 || y < 0 || y >= d.Height || x >= d.Width {
		return
	}
	if d.Cells[y][x] == c {
		return
	}
	d.Cells[y][x] = c
	d.Dirty = true
}

// StyleFor returns the resolved DocStyle for a cell's style name. Falls
// back to the "default" palette entry, then to a zero-value DocStyle. The
// returned pointer must not be mutated.
func (d *Document) StyleFor(name string) *DocStyle {
	if s, ok := d.Palette[name]; ok && s != nil {
		return s
	}
	if s, ok := d.Palette["default"]; ok && s != nil {
		return s
	}
	return &DocStyle{}
}

// RenameStyle renames a palette entry and updates every cell that
// references the old name. The "default" entry cannot be renamed.
func (d *Document) RenameStyle(from, to string) error {
	if from == "default" {
		return fmt.Errorf("cannot rename the default style")
	}
	if from == to {
		return nil
	}
	if _, ok := d.Palette[from]; !ok {
		return fmt.Errorf("style %q not found", from)
	}
	if _, ok := d.Palette[to]; ok {
		return fmt.Errorf("style %q already exists", to)
	}
	d.Palette[to] = d.Palette[from]
	delete(d.Palette, from)
	for y, row := range d.Cells {
		for x, c := range row {
			if c.Style == from {
				d.Cells[y][x].Style = to
			}
		}
	}
	d.Dirty = true
	return nil
}

// DeleteStyle removes a palette entry. Cells using the removed name are
// remapped to the empty (default) style. The "default" entry cannot be
// deleted.
func (d *Document) DeleteStyle(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete the default style")
	}
	if _, ok := d.Palette[name]; !ok {
		return fmt.Errorf("style %q not found", name)
	}
	delete(d.Palette, name)
	for y, row := range d.Cells {
		for x, c := range row {
			if c.Style == name {
				d.Cells[y][x].Style = ""
			}
		}
	}
	d.Dirty = true
	return nil
}

// Resize changes the document size, clipping cells that fall outside the
// new bounds and padding with empty cells where it grows.
func (d *Document) Resize(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if width == d.Width && height == d.Height {
		return
	}
	cells := make([][]Cell, height)
	for y := range height {
		row := make([]Cell, width)
		for x := range width {
			if y < d.Height && x < d.Width {
				row[x] = d.Cells[y][x]
			} else {
				row[x] = EmptyCell
			}
		}
		cells[y] = row
	}
	d.Width = width
	d.Height = height
	d.Cells = cells
	d.Dirty = true
}

// Clear fills the document with empty cells and marks it dirty.
func (d *Document) Clear() {
	for y, row := range d.Cells {
		for x := range row {
			d.Cells[y][x] = EmptyCell
		}
	}
	d.Dirty = true
}

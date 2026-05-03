package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDocument_NewHasDefaultStyle(t *testing.T) {
	d := NewDocument(10, 5)
	if _, ok := d.Palette["default"]; !ok {
		t.Error("new document missing default palette entry")
	}
	if d.Width != 10 || d.Height != 5 {
		t.Errorf("dims = %dx%d; want 10x5", d.Width, d.Height)
	}
	if len(d.Cells) != 5 || len(d.Cells[0]) != 10 {
		t.Errorf("cells shape = %dx%d; want 5x10", len(d.Cells), len(d.Cells[0]))
	}
	if d.Cells[0][0] != EmptyCell {
		t.Errorf("first cell = %+v; want EmptyCell", d.Cells[0][0])
	}
}

func TestDocument_SaveLoadRoundTrip(t *testing.T) {
	d := NewDocument(4, 3)
	d.Palette["accent"] = &DocStyle{Fg: "$cyan", Font: "bold"}
	d.Set(1, 1, Cell{Ch: "┌", Style: "accent"})
	d.Set(2, 1, Cell{Ch: "─", Style: "accent"})

	path := filepath.Join(t.TempDir(), "doc.malwerk.json")
	if err := d.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	loaded, err := LoadDocument(path)
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if loaded.Width != d.Width || loaded.Height != d.Height {
		t.Errorf("dims after round-trip = %dx%d; want %dx%d",
			loaded.Width, loaded.Height, d.Width, d.Height)
	}
	if !reflect.DeepEqual(loaded.Cells, d.Cells) {
		t.Errorf("cells differ after round-trip\n got %+v\nwant %+v", loaded.Cells, d.Cells)
	}
	if loaded.Palette["accent"] == nil || loaded.Palette["accent"].Font != "bold" {
		t.Error("palette accent entry lost during round-trip")
	}
	if loaded.Dirty {
		t.Error("freshly loaded document should not be dirty")
	}
}

func TestDocument_LoadRejectsBadShape(t *testing.T) {
	bad := `{"width":3,"height":2,"cells":[[{"ch":" "},{"ch":" "}]]}`
	tmp := filepath.Join(t.TempDir(), "bad.json")
	if err := writeFile(tmp, bad); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	if _, err := LoadDocument(tmp); err == nil {
		t.Error("expected error for mismatched cell-grid shape")
	}
}

func TestDocument_RenameStylePropagates(t *testing.T) {
	d := NewDocument(3, 1)
	d.Palette["a"] = &DocStyle{Fg: "$red"}
	d.Set(1, 0, Cell{Ch: "x", Style: "a"})
	if err := d.RenameStyle("a", "b"); err != nil {
		t.Fatalf("RenameStyle: %v", err)
	}
	if d.Cells[0][1].Style != "b" {
		t.Errorf("cell style = %q; want %q after rename", d.Cells[0][1].Style, "b")
	}
}

func TestDocument_DeleteStyleRemapsCells(t *testing.T) {
	d := NewDocument(2, 1)
	d.Palette["a"] = &DocStyle{}
	d.Set(0, 0, Cell{Ch: "x", Style: "a"})
	if err := d.DeleteStyle("a"); err != nil {
		t.Fatalf("DeleteStyle: %v", err)
	}
	if d.Cells[0][0].Style != "" {
		t.Errorf("cell style after delete = %q; want \"\"", d.Cells[0][0].Style)
	}
}

func TestDocument_ResizePreservesCells(t *testing.T) {
	d := NewDocument(3, 3)
	d.Set(1, 1, Cell{Ch: "x", Style: ""})
	d.Resize(2, 2)
	if d.Cells[1][1].Ch != "x" {
		t.Errorf("cell 1,1 lost during shrink: got %+v", d.Cells[1][1])
	}
	d.Resize(4, 4)
	if d.Cells[1][1].Ch != "x" {
		t.Errorf("cell 1,1 lost during grow: got %+v", d.Cells[1][1])
	}
	if d.Cells[3][3] != EmptyCell {
		t.Errorf("new cell 3,3 = %+v; want empty", d.Cells[3][3])
	}
}

func TestDocument_DefaultStyleEditableButNotRenamable(t *testing.T) {
	d := NewDocument(2, 2)
	if err := d.RenameStyle("default", "x"); err == nil {
		t.Error("expected RenameStyle('default', ...) to fail")
	}
	if err := d.DeleteStyle("default"); err == nil {
		t.Error("expected DeleteStyle('default') to fail")
	}
}

func writeFile(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o644)
}

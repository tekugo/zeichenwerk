package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

// Ensure Flex implements Container interface.
var _ Container = (*Flex)(nil)

func TestNewFlex(t *testing.T) {
	f := NewFlex("flex1", "cls", Start, 2)

	if f.ID() != "flex1" {
		t.Errorf("ID() = %q; want %q", f.ID(), "flex1")
	}
	if f.alignment != Start {
		t.Errorf("alignment = %s; want %q", f.alignment.String(), "start")
	}
	if f.spacing != 2 {
		t.Errorf("spacing = %d; want 2", f.spacing)
	}
	if len(f.children) != 0 {
		t.Errorf("expected no children, got %d", len(f.children))
	}
}

func TestFlex_Add(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")

	f.Add(c1)
	f.Add(c2)

	if len(f.children) != 2 {
		t.Errorf("expected 2 children, got %d", len(f.children))
	}
	if f.children[0].ID() != "c1" {
		t.Errorf("children[0].ID() = %q; want %q", f.children[0].ID(), "c1")
	}
	if f.children[1].ID() != "c2" {
		t.Errorf("children[1].ID() = %q; want %q", f.children[1].ID(), "c2")
	}
}

func TestFlex_Add_Nil(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	f.Add(nil) // should not panic or add anything
	if len(f.children) != 0 {
		t.Errorf("nil widget should not be added; got %d children", len(f.children))
	}
}

func TestFlex_Add_SetsParent(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	c := NewComponent("c", "")
	f.Add(c)
	if c.Parent() != f {
		t.Error("Add() should set the child's parent to the flex container")
	}
}

func TestFlex_Children(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	f.Add(c1)
	f.Add(c2)

	children := f.Children()
	if len(children) != 2 {
		t.Errorf("Children() returned %d items; want 2", len(children))
	}
	if children[0].ID() != "c1" || children[1].ID() != "c2" {
		t.Errorf("Children() returned wrong order: [%s, %s]", children[0].ID(), children[1].ID())
	}
}

func TestFlex_Hint_Empty(t *testing.T) {
	f := NewFlex("flex", "", Start, 5)
	w, h := f.Hint()
	if w != 0 || h != 0 {
		t.Errorf("Hint() for empty flex = %d,%d; want 0,0", w, h)
	}
}

func TestFlex_Hint_SingleChild(t *testing.T) {
	f := NewFlex("flex", "", Start, 5)
	c := NewComponent("c", "")
	c.SetHint(30, 10)
	f.Add(c)

	w, h := f.Hint()
	// No spacing for a single child
	if w != 30 {
		t.Errorf("Hint() width = %d; want 30", w)
	}
	if h != 10 {
		t.Errorf("Hint() height = %d; want 10", h)
	}
}

func TestFlex_Hint_Horizontal(t *testing.T) {
	f := NewFlex("flex", "", Start, 2)
	c1 := NewComponent("c1", "")
	c1.SetHint(10, 5)
	c2 := NewComponent("c2", "")
	c2.SetHint(20, 8)
	f.Add(c1)
	f.Add(c2)

	w, h := f.Hint()
	// Width = sum of widths + spacing between = 10 + 20 + 2 = 32
	// Height = max child height = 8
	if w != 32 {
		t.Errorf("Hint() width = %d; want 32", w)
	}
	if h != 8 {
		t.Errorf("Hint() height = %d; want 8", h)
	}
}

func TestFlex_Hint_Vertical(t *testing.T) {
	f := NewFlex("flex", "", Start, 3)
	c1 := NewComponent("c1", "")
	c1.SetHint(10, 5)
	c2 := NewComponent("c2", "")
	c2.SetHint(20, 8)
	f.Add(c1)
	f.Add(c2)

	w, h := f.Hint()
	// Width = max child width = 20
	// Height = sum of heights + spacing between = 5 + 8 + 3 = 16
	if w != 20 {
		t.Errorf("Hint() width = %d; want 20", w)
	}
	if h != 16 {
		t.Errorf("Hint() height = %d; want 16", h)
	}
}

func TestFlex_Hint_NoSpacingOnFirst(t *testing.T) {
	// Spacing is only added between children, not before the first
	f := NewFlex("flex", "", Start, 10)
	c1 := NewComponent("c1", "")
	c1.SetHint(5, 3)
	c2 := NewComponent("c2", "")
	c2.SetHint(5, 3)
	c3 := NewComponent("c3", "")
	c3.SetHint(5, 3)
	f.Add(c1)
	f.Add(c2)
	f.Add(c3)

	w, _ := f.Hint()
	// 5 + 10 + 5 + 10 + 5 = 35
	if w != 35 {
		t.Errorf("Hint() width = %d; want 35 (spacing between, not before first)", w)
	}
}

func TestFlex_LayoutHorizontal_FixedSizes(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	f.SetBounds(0, 0, 60, 20)

	c1 := NewComponent("c1", "")
	c1.SetHint(20, 10)
	c2 := NewComponent("c2", "")
	c2.SetHint(40, 15)
	f.Add(c1)
	f.Add(c2)
	f.Layout()

	x1, _, w1, _ := c1.Bounds()
	x2, _, w2, _ := c2.Bounds()

	if x1 != 0 {
		t.Errorf("c1 x = %d; want 0", x1)
	}
	if w1 != 20 {
		t.Errorf("c1 width = %d; want 20", w1)
	}
	if x2 != 20 {
		t.Errorf("c2 x = %d; want 20", x2)
	}
	if w2 != 40 {
		t.Errorf("c2 width = %d; want 40", w2)
	}
}

func TestFlex_LayoutHorizontal_WithSpacing(t *testing.T) {
	f := NewFlex("flex", "", Start, 5)
	f.SetBounds(0, 0, 65, 20)

	c1 := NewComponent("c1", "")
	c1.SetHint(20, 10)
	c2 := NewComponent("c2", "")
	c2.SetHint(40, 15)
	f.Add(c1)
	f.Add(c2)
	f.Layout()

	x1, _, _, _ := c1.Bounds()
	x2, _, _, _ := c2.Bounds()

	if x1 != 0 {
		t.Errorf("c1 x = %d; want 0", x1)
	}
	// c2 starts after c1's width (20) plus spacing (5)
	if x2 != 25 {
		t.Errorf("c2 x = %d; want 25", x2)
	}
}

func TestFlex_LayoutHorizontal_FractionalSizes(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	f.SetBounds(0, 0, 60, 20)

	c1 := NewComponent("c1", "")
	c1.SetHint(-1, 10) // 1 fractional unit
	c2 := NewComponent("c2", "")
	c2.SetHint(-2, 10) // 2 fractional units
	f.Add(c1)
	f.Add(c2)
	f.Layout()

	_, _, w1, _ := c1.Bounds()
	_, _, w2, _ := c2.Bounds()

	// 3 total fractions across 60px: c1=20, c2=40 (last gets remainder)
	if w1 != 20 {
		t.Errorf("c1 width = %d; want 20", w1)
	}
	if w2 != 40 {
		t.Errorf("c2 width = %d; want 40", w2)
	}
}

func TestFlex_LayoutVertical_FixedSizes(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	f.SetFlag(FlagVertical, true)
	f.SetBounds(0, 0, 60, 30)

	c1 := NewComponent("c1", "")
	c1.SetHint(20, 10)
	c2 := NewComponent("c2", "")
	c2.SetHint(30, 20)
	f.Add(c1)
	f.Add(c2)
	f.Layout()

	_, y1, _, h1 := c1.Bounds()
	_, y2, _, h2 := c2.Bounds()

	if y1 != 0 {
		t.Errorf("c1 y = %d; want 0", y1)
	}
	if h1 != 10 {
		t.Errorf("c1 height = %d; want 10", h1)
	}
	if y2 != 10 {
		t.Errorf("c2 y = %d; want 10", y2)
	}
	if h2 != 20 {
		t.Errorf("c2 height = %d; want 20", h2)
	}
}

func TestFlex_LayoutVertical_WithSpacing(t *testing.T) {
	f := NewFlex("flex", "", Start, 4)
	f.SetFlag(FlagVertical, true)
	f.SetBounds(0, 0, 60, 34)

	c1 := NewComponent("c1", "")
	c1.SetHint(20, 10)
	c2 := NewComponent("c2", "")
	c2.SetHint(30, 20)
	f.Add(c1)
	f.Add(c2)
	f.Layout()

	_, y1, _, _ := c1.Bounds()
	_, y2, _, _ := c2.Bounds()

	if y1 != 0 {
		t.Errorf("c1 y = %d; want 0", y1)
	}
	// c2 starts after c1's height (10) plus spacing (4)
	if y2 != 14 {
		t.Errorf("c2 y = %d; want 14", y2)
	}
}

func TestFlex_LayoutVertical_FractionalSizes(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	f.SetFlag(FlagVertical, true)
	f.SetBounds(0, 0, 60, 30)

	c1 := NewComponent("c1", "")
	c1.SetHint(10, -1) // 1 fractional unit
	c2 := NewComponent("c2", "")
	c2.SetHint(10, -2) // 2 fractional units
	f.Add(c1)
	f.Add(c2)
	f.Layout()

	_, _, _, h1 := c1.Bounds()
	_, _, _, h2 := c2.Bounds()

	// 3 total fractions across 30px: c1=10, c2=20 (last gets remainder)
	if h1 != 10 {
		t.Errorf("c1 height = %d; want 10", h1)
	}
	if h2 != 20 {
		t.Errorf("c2 height = %d; want 20", h2)
	}
}

func TestFlex_LayoutHorizontal_AlignStart(t *testing.T) {
	f := NewFlex("flex", "", Start, 0)
	f.SetBounds(0, 0, 60, 30)

	c := NewComponent("c", "")
	c.SetHint(20, 10)
	f.Add(c)
	f.Layout()

	_, y, _, h := c.Bounds()
	if y != 0 {
		t.Errorf("start align: y = %d; want 0", y)
	}
	if h != 10 {
		t.Errorf("start align: height = %d; want 10", h)
	}
}

func TestFlex_LayoutHorizontal_AlignCenter(t *testing.T) {
	f := NewFlex("flex", "", Center, 0)
	f.SetBounds(0, 0, 60, 30)

	c := NewComponent("c", "")
	c.SetHint(20, 10)
	f.Add(c)
	f.Layout()

	_, y, _, h := c.Bounds()
	// center in 30px with size 10: (30-10)/2 = 10
	if y != 10 {
		t.Errorf("center align: y = %d; want 10", y)
	}
	if h != 10 {
		t.Errorf("center align: height = %d; want 10", h)
	}
}

func TestFlex_LayoutHorizontal_AlignEnd(t *testing.T) {
	f := NewFlex("flex", "", End, 0)
	f.SetBounds(0, 0, 60, 30)

	c := NewComponent("c", "")
	c.SetHint(20, 10)
	f.Add(c)
	f.Layout()

	_, y, _, h := c.Bounds()
	// end: 30 - 10 = 20
	if y != 20 {
		t.Errorf("end align: y = %d; want 20", y)
	}
	if h != 10 {
		t.Errorf("end align: height = %d; want 10", h)
	}
}

func TestFlex_LayoutHorizontal_AlignStretch(t *testing.T) {
	f := NewFlex("flex", "", Stretch, 0)
	f.SetBounds(0, 0, 60, 30)

	c := NewComponent("c", "")
	c.SetHint(20, 10)
	f.Add(c)
	f.Layout()

	_, y, _, h := c.Bounds()
	if y != 0 {
		t.Errorf("stretch align: y = %d; want 0", y)
	}
	// stretched to fill container height
	if h != 30 {
		t.Errorf("stretch align: height = %d; want 30", h)
	}
}

func TestAlign(t *testing.T) {
	tests := []struct {
		name      string
		alignment Alignment
		start     int
		end       int
		size      int
		wantPos   int
		wantSize  int
	}{
		{
			name: "start", alignment: Start,
			start: 0, end: 100, size: 30,
			wantPos: 0, wantSize: 30,
		},
		{
			name: "center fits", alignment: Center,
			start: 0, end: 100, size: 30,
			wantPos: 35, wantSize: 30, // (100-30)/2 = 35
		},
		{
			name: "center too large", alignment: Center,
			start: 0, end: 20, size: 30,
			wantPos: 10, wantSize: 30, // space/2 = 10
		},
		{
			name: "end", alignment: End,
			start: 0, end: 100, size: 30,
			wantPos: 70, wantSize: 30, // 100-30 = 70
		},
		{
			name: "stretch", alignment: Stretch,
			start: 0, end: 100, size: 30,
			wantPos: 0, wantSize: 100, // fills available space
		},
		{
			name: "unknown defaults to start", alignment: Default,
			start: 0, end: 100, size: 30,
			wantPos: 0, wantSize: 30,
		},
		{
			name: "start with offset", alignment: Start,
			start: 10, end: 110, size: 20,
			wantPos: 10, wantSize: 20,
		},
		{
			name: "end with offset", alignment: End,
			start: 10, end: 110, size: 20,
			wantPos: 90, wantSize: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, size := align(tt.alignment, tt.start, tt.end, tt.size)
			if pos != tt.wantPos {
				t.Errorf("align(%s, %d, %d, %d) pos = %d; want %d",
					tt.alignment.String(), tt.start, tt.end, tt.size, pos, tt.wantPos)
			}
			if size != tt.wantSize {
				t.Errorf("align(%s, %d, %d, %d) size = %d; want %d",
					tt.alignment.String(), tt.start, tt.end, tt.size, size, tt.wantSize)
			}
		})
	}
}

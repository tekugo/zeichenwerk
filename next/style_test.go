package next

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyleNew(t *testing.T) {
	style := NewStyle("button").WithBorder("thin")
	assert.Equal(t, "button", style.selector)
	assert.Nil(t, style.parent)
}

func TestStyleParent(t *testing.T) {
	style := NewStyle("button").WithBorder("thin").WithParent(&DefaultStyle)
	assert.Equal(t, "button", style.selector)
	assert.NotNil(t, style.parent)
	assert.NotNil(t, style.Margin())
	assert.NotNil(t, style.Padding())
}

func TestStyleValues(t *testing.T) {
	style := NewStyle("button").
		WithBackground("bg").
		WithForeground("fg").
		WithBorder("border").
		WithCursor("cursor").
		WithMargin(1, 2, 3, 4).
		WithPadding(5, 6, 7, 8)

	assert.Equal(t, "bg", style.Background())
	assert.Equal(t, "fg", style.Foreground())
	assert.Equal(t, "border", style.Border())
	assert.Equal(t, "cursor", style.Cursor())
	assert.Equal(t, 1, style.Margin().Top)
	assert.Equal(t, 2, style.Margin().Right)
	assert.Equal(t, 3, style.Margin().Bottom)
	assert.Equal(t, 4, style.Margin().Left)
	assert.Equal(t, 5, style.Padding().Top)
	assert.Equal(t, 6, style.Padding().Right)
	assert.Equal(t, 7, style.Padding().Bottom)
	assert.Equal(t, 8, style.Padding().Left)
}

func TestStyleFix(t *testing.T) {
	style := NewStyle("button").WithBorder("thin").Fix()
	assert.True(t, style.fixed)
	next := style.WithBackground("black")
	assert.NotSame(t, style, next)
	assert.Same(t, style, next.parent)
}

func TestStyleMod(t *testing.T) {
	style := NewStyle("button").WithBorder("thin")
	assert.False(t, style.fixed)
	next := style.WithBackground("black")
	assert.Same(t, style, next)
}

// TestStyleConstructorEdgeCases tests NewStyle with different parameter scenarios
func TestStyleConstructorEdgeCases(t *testing.T) {
	// Test with no parameters
	style1 := NewStyle()
	assert.Equal(t, "", style1.selector)
	assert.Nil(t, style1.parent)
	assert.False(t, style1.fixed)

	// Test with selector parameter
	style2 := NewStyle("test-selector")
	assert.Equal(t, "test-selector", style2.selector)
	assert.Nil(t, style2.parent)
	assert.False(t, style2.fixed)
}

// TestStylePropertyAccessorsWithInheritance tests property getters with parent inheritance
func TestStylePropertyAccessorsWithInheritance(t *testing.T) {
	parent := NewStyle("parent").
		WithBackground("parent-bg").
		WithForeground("parent-fg").
		WithBorder("parent-border").
		WithCursor("parent-cursor").
		WithFont("parent-font").
		WithMargin(1, 2, 3, 4).
		WithPadding(5, 6, 7, 8)

	child := NewStyle("child").WithParent(parent)

	// Test inheritance from parent
	assert.Equal(t, "parent-bg", child.Background())
	assert.Equal(t, "parent-fg", child.Foreground())
	assert.Equal(t, "parent-border", child.Border())
	assert.Equal(t, "parent-cursor", child.Cursor())
	assert.Equal(t, 1, child.Margin().Top)
	assert.Equal(t, 5, child.Padding().Top)

	// Test child overrides
	child = child.WithBackground("child-bg").WithForeground("child-fg")
	assert.Equal(t, "child-bg", child.Background())
	assert.Equal(t, "child-fg", child.Foreground())
	assert.Equal(t, "parent-border", child.Border()) // Still inherited
}

// TestStylePropertyAccessorsWithoutParent tests property getters without parent
func TestStylePropertyAccessorsWithoutParent(t *testing.T) {
	style := NewStyle("test")

	// Test empty values without parent
	assert.Equal(t, "", style.Background())
	assert.Equal(t, "", style.Foreground())
	assert.Equal(t, "", style.Border())
	assert.Equal(t, "", style.Cursor())
	assert.NotNil(t, style.Margin())
	assert.Equal(t, 0, style.Margin().Top)
	assert.NotNil(t, style.Padding())
	assert.Equal(t, 0, style.Padding().Top)
}

// TestStyleFixed tests the Fixed() method
func TestStyleFixed(t *testing.T) {
	style := NewStyle("test")
	assert.False(t, style.Fixed())

	style.Fix()
	assert.True(t, style.Fixed())
}

// TestStyleFont tests the Font() method
func TestStyleFont(t *testing.T) {
	// Test without parent
	style := NewStyle("test").WithFont("arial")
	assert.Equal(t, "arial", style.Font())

	// Test with parent inheritance
	parent := NewStyle("parent").WithFont("parent-font")
	child := NewStyle("child").WithParent(parent)
	assert.Equal(t, "parent-font", child.Font())

	// Test child override
	child = child.WithFont("child-font")
	assert.Equal(t, "child-font", child.Font())

	// Test empty font with parent
	childEmpty := NewStyle("child-empty").WithParent(parent)
	assert.Equal(t, "parent-font", childEmpty.Font())
}

// TestStyleParentMethod tests the Parent() method
func TestStyleParentMethod(t *testing.T) {
	parent := NewStyle("parent")
	child := NewStyle("child").WithParent(parent)

	assert.Nil(t, parent.Parent())
	assert.Equal(t, parent, child.Parent())
}

// TestStyleWithColors tests the WithColors() method
func TestStyleWithColors(t *testing.T) {
	style := NewStyle("test").WithColors("red", "blue")
	assert.Equal(t, "red", style.Foreground())
	assert.Equal(t, "blue", style.Background())
}

// TestStyleWithFont tests the WithFont() method
func TestStyleWithFont(t *testing.T) {
	style := NewStyle("test").WithFont("helvetica")
	assert.Equal(t, "helvetica", style.Font())
}

// TestStyleHorizontal tests the Horizontal() spacing calculation method
func TestStyleHorizontal(t *testing.T) {
	// Test with no border
	style := NewStyle("test").
		WithMargin(1, 2, 3, 4). // left=4, right=2
		WithPadding(5, 6, 7, 8) // left=8, right=6
	expected := 4 + 2 + 8 + 6 // margin + padding, no border
	assert.Equal(t, expected, style.Horizontal())

	// Test with border
	style = style.WithBorder("thin")
	expected = 4 + 2 + 8 + 6 + 2 // margin + padding + border (2)
	assert.Equal(t, expected, style.Horizontal())

	// Test with "none" border (should not add border space)
	style = style.WithBorder("none")
	expected = 4 + 2 + 8 + 6 // margin + padding, no border
	assert.Equal(t, expected, style.Horizontal())

	// Test with empty border
	style = style.WithBorder("")
	expected = 4 + 2 + 8 + 6 // margin + padding, no border
	assert.Equal(t, expected, style.Horizontal())

	// Test with nil margin and padding
	style = NewStyle("test")
	assert.Equal(t, 0, style.Horizontal())
}

// TestStyleVertical tests the Vertical() spacing calculation method
func TestStyleVertical(t *testing.T) {
	// Test with no border
	style := NewStyle("test").
		WithMargin(1, 2, 3, 4). // top=1, bottom=3
		WithPadding(5, 6, 7, 8) // top=5, bottom=7
	expected := 1 + 3 + 5 + 7 // margin + padding, no border
	assert.Equal(t, expected, style.Vertical())

	// Test with border
	style = style.WithBorder("thick")
	expected = 1 + 3 + 5 + 7 + 2 // margin + padding + border (2)
	assert.Equal(t, expected, style.Vertical())

	// Test with "none" border
	style = style.WithBorder("none")
	expected = 1 + 3 + 5 + 7 // margin + padding, no border
	assert.Equal(t, expected, style.Vertical())

	// Test with nil margin and padding
	style = NewStyle("test")
	assert.Equal(t, 0, style.Vertical())
}

// TestStyleInfo tests the Info() method
func TestStyleInfo(t *testing.T) {
	style := NewStyle("test-selector").
		WithBackground("red").
		WithForeground("white").
		WithBorder("thick").
		WithCursor("pointer").
		WithMargin(1, 2, 3, 4).
		WithPadding(5, 6, 7, 8)

	info := style.Info()
	assert.Contains(t, info, "test-selector")
	assert.Contains(t, info, "red")
	assert.Contains(t, info, "white")
	assert.Contains(t, info, "thick")
	assert.Contains(t, info, "pointer")
	assert.Contains(t, info, "(1 2 3 4)")
	assert.Contains(t, info, "(5 6 7 8)")
}

// TestStyleInheritanceEdgeCases tests edge cases in inheritance
func TestStyleInheritanceEdgeCases(t *testing.T) {
	// Test with parent that has empty values
	parent := NewStyle("parent") // No values set
	child := NewStyle("child").WithParent(parent)

	assert.Equal(t, "", child.Background())
	assert.Equal(t, "", child.Foreground())
	assert.Equal(t, "", child.Border())
	assert.Equal(t, "", child.Cursor())

	// Test margin/padding inheritance when parent has nil values
	assert.NotNil(t, child.Margin())
	assert.NotNil(t, child.Padding())
	assert.Equal(t, 0, child.Margin().Top)
	assert.Equal(t, 0, child.Padding().Top)
}

// TestStyleSpacingInheritance tests margin and padding inheritance
func TestStyleSpacingInheritance(t *testing.T) {
	parent := NewStyle("parent").
		WithMargin(10, 20, 30, 40).
		WithPadding(1, 2, 3, 4)

	child := NewStyle("child").WithParent(parent)

	// Child should inherit parent's spacing
	assert.Equal(t, 10, child.Margin().Top)
	assert.Equal(t, 20, child.Margin().Right)
	assert.Equal(t, 1, child.Padding().Top)
	assert.Equal(t, 2, child.Padding().Right)

	// Child can override spacing
	child = child.WithMargin(100).WithPadding(200)
	assert.Equal(t, 100, child.Margin().Top)
	assert.Equal(t, 200, child.Padding().Top)
}

// TestDefaultStyleBehavior tests behavior with DefaultStyle
func TestDefaultStyleBehavior(t *testing.T) {
	// DefaultStyle should be fixed
	assert.True(t, DefaultStyle.Fixed())

	// DefaultStyle should have default values
	assert.Equal(t, "black", DefaultStyle.Background())
	assert.Equal(t, "white", DefaultStyle.Foreground())
	assert.Equal(t, "none", DefaultStyle.Border())

	// Child of DefaultStyle should inherit values
	child := NewStyle("child").WithParent(&DefaultStyle)
	assert.Equal(t, "black", child.Background())
	assert.Equal(t, "white", child.Foreground())
}

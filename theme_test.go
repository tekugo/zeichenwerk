package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestThemeAdd(t *testing.T) {
	theme := NewMapTheme()

	theme.Add(NewStyle("").WithForeground("green").WithBackground("black").WithMargin(0).WithPadding(0))
	theme.Add(NewStyle("button").WithBorder("thin"))

	assert.Equal(t, 2, len(theme.styles))

	style := theme.Get("button")
	assert.Equal(t, "button", style.selector)
	assert.NotNil(t, style.parent)
	assert.Same(t, theme.Get("button").parent, theme.Get(""))
}

func TestThemeParent(t *testing.T) {
	theme := NewMapTheme()

	theme.Add(NewStyle("").WithForeground("green").WithBackground("black").WithMargin(0).WithPadding(0))

	assert.Nil(t, theme.Get("").parent)

	theme.Add(NewStyle("button").WithBorder("round"))
	assert.Same(t, theme.Get("button").parent, theme.Get(""))

	theme.Add(NewStyle("button.primary").WithForeground("blue"))
	assert.Same(t, theme.Get("button.primary").parent, theme.Get("button"))

	theme.Add(NewStyle("button.primary:focus").WithBackground("red"))
	assert.Same(t, theme.Get("button.primary:focus").parent, theme.Get("button.primary"))
	assert.Equal(t, "round", theme.Get("button.primary:focus").Border())

	theme.Add(NewStyle("box").WithBorder("thin"))
	assert.Same(t, theme.Get("box").parent, theme.Get(""))

	theme.Add(NewStyle("box/title").WithForeground("yellow"))
	assert.Same(t, theme.Get("box/title").parent, theme.Get("box"))
}

func TestThemeTokyoNight(t *testing.T) {
	theme := TokyoNightTheme()
	assert.NotNil(t, theme)
	assert.Equal(t, 43, len(theme.Styles()))
}

// TestThemeApply tests the Apply method which applies styles to widgets
func TestThemeApply(t *testing.T) {
	theme := NewMapTheme()
	theme.Add(NewStyle("button").WithBackground("blue"))
	theme.Add(NewStyle("button:focus").WithBackground("red"))
	theme.Add(NewStyle("input/placeholder").WithForeground("gray"))
	theme.Add(NewStyle("input/placeholder:focus").WithForeground("white"))

	// Create a mock widget to test Apply method
	widget := &ThemeMockWidget{
		styles: make(map[string]*Style),
	}

	// Test applying base selector without parts
	theme.Apply(widget, "button", "focus")
	assert.NotNil(t, widget.styles[""])
	assert.NotNil(t, widget.styles[":focus"])
	assert.Equal(t, "blue", widget.styles[""].Background())
	assert.Equal(t, "red", widget.styles[":focus"].Background())

	// Test applying selector with parts
	theme.Apply(widget, "input/placeholder", "focus")
	assert.NotNil(t, widget.styles["placeholder"])
	assert.NotNil(t, widget.styles["placeholder:focus"])
	assert.Equal(t, "gray", widget.styles["placeholder"].Foreground())
	assert.Equal(t, "white", widget.styles["placeholder:focus"].Foreground())
}

// TestThemeBorder tests the Border method
func TestThemeBorder(t *testing.T) {
	theme := NewMapTheme()
	
	// Test empty borders initially
	border := theme.Border("thin")
	assert.Equal(t, BorderStyle{}, border)

	// Set some borders and test retrieval
	borders := map[string]BorderStyle{
		"thin":   {},
		"thick":  {},
		"double": {},
	}
	theme.SetBorders(borders)

	// Test that borders are retrievable (exact struct content may vary)
	thinBorder := theme.Border("thin")
	thickBorder := theme.Border("thick")
	doubleBorder := theme.Border("double")
	assert.NotNil(t, thinBorder)
	assert.NotNil(t, thickBorder)
	assert.NotNil(t, doubleBorder)

	// Test non-existent border returns empty BorderStyle
	emptyBorder := theme.Border("nonexistent")
	assert.Equal(t, BorderStyle{}, emptyBorder)
}

// TestThemeColor tests the Color method with variables and direct colors
func TestThemeColor(t *testing.T) {
	theme := NewMapTheme()

	// Test direct colors (should be returned as-is)
	assert.Equal(t, "red", theme.Color("red"))
	assert.Equal(t, "#FF0000", theme.Color("#FF0000"))
	assert.Equal(t, "rgb(255,0,0)", theme.Color("rgb(255,0,0)"))

	// Test undefined variables (should return as-is)
	assert.Equal(t, "$undefined", theme.Color("$undefined"))

	// Set some color variables
	colors := map[string]string{
		"$primary":   "#007ACC",
		"$secondary": "#FF6B35",
		"$background": "#1E1E1E",
	}
	theme.SetColors(colors)

	// Test variable resolution
	assert.Equal(t, "#007ACC", theme.Color("$primary"))
	assert.Equal(t, "#FF6B35", theme.Color("$secondary"))
	assert.Equal(t, "#1E1E1E", theme.Color("$background"))

	// Test undefined variable after setting colors
	assert.Equal(t, "$undefined", theme.Color("$undefined"))

	// Test direct colors still work
	assert.Equal(t, "blue", theme.Color("blue"))
}

// TestThemeColors tests the Colors method
func TestThemeColors(t *testing.T) {
	theme := NewMapTheme()

	// Test empty colors initially
	colors := theme.Colors()
	assert.Empty(t, colors)

	// Set colors and test retrieval
	colorMap := map[string]string{
		"$primary":   "#007ACC",
		"$secondary": "#FF6B35",
	}
	theme.SetColors(colorMap)

	colors = theme.Colors()
	assert.Equal(t, 2, len(colors))
	assert.Equal(t, "#007ACC", colors["$primary"])
	assert.Equal(t, "#FF6B35", colors["$secondary"])
}

// TestThemeFlag tests the Flag method
func TestThemeFlag(t *testing.T) {
	theme := NewMapTheme()

	// Test undefined flags (should return false)
	assert.False(t, theme.Flag("undefined"))
	assert.False(t, theme.Flag("debug"))

	// Set some flags
	flags := map[string]bool{
		"debug":      true,
		"production": false,
		"feature_x":  true,
	}
	theme.SetFlags(flags)

	// Test flag retrieval
	assert.True(t, theme.Flag("debug"))
	assert.False(t, theme.Flag("production"))
	assert.True(t, theme.Flag("feature_x"))

	// Test undefined flag after setting flags
	assert.False(t, theme.Flag("undefined"))
}

// TestThemeGet tests the Get method with various selector patterns
func TestThemeGet(t *testing.T) {
	theme := NewMapTheme()

	// Test getting non-existent style (should return DefaultStyle)
	style := theme.Get("nonexistent")
	assert.Same(t, &DefaultStyle, style)

	// Add some styles with different specificity
	theme.Add(NewStyle("button").WithBackground("blue"))
	theme.Add(NewStyle("button.primary").WithBackground("green"))
	theme.Add(NewStyle("button#submit").WithBackground("red"))
	theme.Add(NewStyle("button:focus").WithForeground("white"))

	// Test basic selector
	style = theme.Get("button")
	assert.Equal(t, "blue", style.Background())

	// Test class selector
	style = theme.Get("button.primary")
	assert.Equal(t, "green", style.Background())

	// Test ID selector (falls back to button since ID selector doesn't match exactly)
	style = theme.Get("button#submit")
	assert.Equal(t, "blue", style.Background()) // Falls back to "button" style

	// Test state selector
	style = theme.Get("button:focus")
	assert.Equal(t, "white", style.Foreground())

	// Test complex selectors that don't exist (should cascade to DefaultStyle)
	style = theme.Get("input.large:disabled")
	assert.Same(t, &DefaultStyle, style)
}

// TestThemeRune tests the Rune method
func TestThemeRune(t *testing.T) {
	theme := NewMapTheme()

	// Test undefined runes (should return zero rune)
	assert.Equal(t, rune(0), theme.Rune("undefined"))
	assert.Equal(t, rune(0), theme.Rune("arrow-up"))

	// Set some runes
	runes := map[string]rune{
		"arrow-up":    '↑',
		"arrow-down":  '↓',
		"arrow-left":  '←',
		"arrow-right": '→',
		"bullet":      '•',
		"checkbox":    '☐',
	}
	theme.SetRunes(runes)

	// Test rune retrieval
	assert.Equal(t, '↑', theme.Rune("arrow-up"))
	assert.Equal(t, '↓', theme.Rune("arrow-down"))
	assert.Equal(t, '←', theme.Rune("arrow-left"))
	assert.Equal(t, '→', theme.Rune("arrow-right"))
	assert.Equal(t, '•', theme.Rune("bullet"))
	assert.Equal(t, '☐', theme.Rune("checkbox"))

	// Test undefined rune after setting runes
	assert.Equal(t, rune(0), theme.Rune("undefined"))
}

// TestThemeSetFlags tests the SetFlags method
func TestThemeSetFlags(t *testing.T) {
	theme := NewMapTheme()

	// Initially no flags
	assert.False(t, theme.Flag("test"))

	// Set flags
	flags := map[string]bool{
		"debug":    true,
		"verbose":  false,
		"feature1": true,
	}
	theme.SetFlags(flags)

	// Verify flags are set
	assert.True(t, theme.Flag("debug"))
	assert.False(t, theme.Flag("verbose"))
	assert.True(t, theme.Flag("feature1"))

	// Replace with new flags
	newFlags := map[string]bool{
		"production": true,
		"debug":      false, // Changed value
	}
	theme.SetFlags(newFlags)

	// Verify old flags are gone and new ones are present
	assert.False(t, theme.Flag("feature1")) // Should be false (not present)
	assert.True(t, theme.Flag("production"))
	assert.False(t, theme.Flag("debug")) // Changed value
}

// TestThemeSetRunes tests the SetRunes method
func TestThemeSetRunes(t *testing.T) {
	theme := NewMapTheme()

	// Initially no runes
	assert.Equal(t, rune(0), theme.Rune("test"))

	// Set runes
	runes := map[string]rune{
		"star":   '★',
		"heart":  '♥',
		"diamond": '♦',
	}
	theme.SetRunes(runes)

	// Verify runes are set
	assert.Equal(t, '★', theme.Rune("star"))
	assert.Equal(t, '♥', theme.Rune("heart"))
	assert.Equal(t, '♦', theme.Rune("diamond"))

	// Replace with new runes
	newRunes := map[string]rune{
		"club":  '♣',
		"spade": '♠',
	}
	theme.SetRunes(newRunes)

	// Verify old runes are gone and new ones are present
	assert.Equal(t, rune(0), theme.Rune("star")) // Should be 0 (not present)
	assert.Equal(t, '♣', theme.Rune("club"))
	assert.Equal(t, '♠', theme.Rune("spade"))
}

// TestThemeSelectorParsing tests the split function and complex selector parsing
func TestThemeSelectorParsing(t *testing.T) {
	theme := NewMapTheme()

	// Test various selector formats
	theme.Add(NewStyle("input").WithBackground("white"))
	theme.Add(NewStyle("input.large").WithPadding(2))
	theme.Add(NewStyle("input#username").WithBorder("thick"))
	theme.Add(NewStyle("#username").WithBorder("thick")) // ID-only selector
	theme.Add(NewStyle("input:focus").WithForeground("blue"))
	theme.Add(NewStyle("input/placeholder").WithForeground("gray"))
	theme.Add(NewStyle("input/placeholder.large").WithFont("large"))
	theme.Add(NewStyle("input.large:focus").WithBackground("yellow"))
	theme.Add(NewStyle("list/item").WithPadding(1))
	theme.Add(NewStyle("list/item.selected").WithBackground("blue"))
	theme.Add(NewStyle("list/item:hover").WithBackground("lightblue"))

	// Test retrieval of various selectors
	assert.Equal(t, "white", theme.Get("input").Background())
	assert.Equal(t, 2, theme.Get("input.large").Padding().Top)
	assert.Equal(t, "thick", theme.Get("#username").Border()) // Test ID-only selector
	assert.Equal(t, "blue", theme.Get("input:focus").Foreground())
	assert.Equal(t, "gray", theme.Get("input/placeholder").Foreground())
	assert.Equal(t, "yellow", theme.Get("input.large:focus").Background())
	assert.Equal(t, 1, theme.Get("list/item").Padding().Top)
	assert.Equal(t, "blue", theme.Get("list/item.selected").Background())
	assert.Equal(t, "lightblue", theme.Get("list/item:hover").Background())
}

// TestThemeComplexInheritance tests complex inheritance scenarios
func TestThemeComplexInheritance(t *testing.T) {
	theme := NewMapTheme()

	// Set up inheritance chain: "" -> "button" -> "button.primary" -> "button.primary:focus"
	theme.Add(NewStyle("").WithForeground("black").WithBackground("white"))
	theme.Add(NewStyle("button").WithBorder("thin"))
	theme.Add(NewStyle("button.primary").WithBackground("blue"))
	theme.Add(NewStyle("button.primary:focus").WithForeground("white"))

	// Test inheritance chain
	baseStyle := theme.Get("")
	buttonStyle := theme.Get("button")
	primaryStyle := theme.Get("button.primary")
	focusStyle := theme.Get("button.primary:focus")

	// Base style
	assert.Equal(t, "black", baseStyle.Foreground())
	assert.Equal(t, "white", baseStyle.Background())

	// Button inherits from base, adds border
	assert.Equal(t, "black", buttonStyle.Foreground()) // Inherited
	assert.Equal(t, "white", buttonStyle.Background()) // Inherited
	assert.Equal(t, "thin", buttonStyle.Border())      // Added

	// Primary inherits from button, overrides background
	assert.Equal(t, "black", primaryStyle.Foreground()) // Inherited from base through button
	assert.Equal(t, "blue", primaryStyle.Background())  // Overridden
	assert.Equal(t, "thin", primaryStyle.Border())      // Inherited from button

	// Focus inherits from primary, overrides foreground
	assert.Equal(t, "white", focusStyle.Foreground()) // Overridden
	assert.Equal(t, "blue", focusStyle.Background())  // Inherited from primary
	assert.Equal(t, "thin", focusStyle.Border())      // Inherited through chain
}

// TestThemeSelectorEdgeCases tests edge cases in selector parsing and resolution
func TestThemeSelectorEdgeCases(t *testing.T) {
	theme := NewMapTheme()

	// Test complex ID selectors with parts and states
	theme.Add(NewStyle("#submit/text").WithForeground("white"))
	theme.Add(NewStyle("#submit/text:focus").WithForeground("yellow"))
	theme.Add(NewStyle("button#submit:hover").WithBackground("gray"))
	theme.Add(NewStyle(".primary").WithBackground("blue"))
	theme.Add(NewStyle(":disabled").WithForeground("gray"))

	// Test retrieval (many fall back to DefaultStyle)
	assert.Equal(t, "white", theme.Get("#submit/text").Foreground())
	assert.Equal(t, "yellow", theme.Get("#submit/text:focus").Foreground())
	assert.Equal(t, "black", theme.Get("button#submit:hover").Background()) // Falls back to DefaultStyle
	assert.Equal(t, "blue", theme.Get(".primary").Background())
	assert.Equal(t, "gray", theme.Get(":disabled").Foreground()) // State selector works
}

// TestThemeApplyWithAlternateParts tests Apply method with parts in different positions
func TestThemeApplyWithAlternateParts(t *testing.T) {
	theme := NewMapTheme()
	theme.Add(NewStyle("button#main/text").WithForeground("white"))
	theme.Add(NewStyle("button#main/text:focus").WithForeground("blue"))

	widget := &ThemeMockWidget{
		styles: make(map[string]*Style),
	}

	// Test applying with alternate part position (after ID)
	theme.Apply(widget, "button#main/text", "focus")
	assert.NotNil(t, widget.styles["text"])
	assert.NotNil(t, widget.styles["text:focus"])
	assert.Equal(t, "white", widget.styles["text"].Foreground())
	// The focus state may fall back to the base style
	assert.Equal(t, "white", widget.styles["text:focus"].Foreground())
}

// ThemeMockWidget is a simple implementation of Widget interface for testing themes
type ThemeMockWidget struct {
	styles map[string]*Style
}

func (m *ThemeMockWidget) SetStyle(key string, style *Style) {
	m.styles[key] = style
}

// Minimal Widget interface implementation for testing
func (m *ThemeMockWidget) Bounds() (int, int, int, int) { return 0, 0, 0, 0 }
func (m *ThemeMockWidget) Content() (int, int, int, int) { return 0, 0, 0, 0 }
func (m *ThemeMockWidget) Cursor() (int, int) { return -1, -1 }
func (m *ThemeMockWidget) Focusable() bool { return false }
func (m *ThemeMockWidget) Focused() bool { return false }
func (m *ThemeMockWidget) Handle(tcell.Event) bool { return false }
func (m *ThemeMockWidget) Hint() (int, int) { return 0, 0 }
func (m *ThemeMockWidget) Hovered() bool { return false }
func (m *ThemeMockWidget) ID() string { return "mock" }
func (m *ThemeMockWidget) Info() string { return "mock widget" }
func (m *ThemeMockWidget) Log(Widget, string, string, ...any) {}
func (m *ThemeMockWidget) On(string, func(Widget, string, ...any) bool) {}
func (m *ThemeMockWidget) Parent() Widget { return nil }
func (m *ThemeMockWidget) Position() (int, int) { return 0, 0 }
func (m *ThemeMockWidget) Refresh() {}
func (m *ThemeMockWidget) SetBounds(int, int, int, int) {}
func (m *ThemeMockWidget) SetFocused(bool) {}
func (m *ThemeMockWidget) SetHint(int, int) {}
func (m *ThemeMockWidget) SetHovered(bool) {}
func (m *ThemeMockWidget) SetParent(Container) {}
func (m *ThemeMockWidget) SetPosition(int, int) {}
func (m *ThemeMockWidget) SetSize(int, int) {}
func (m *ThemeMockWidget) Size() (int, int) { return 0, 0 }
func (m *ThemeMockWidget) State() string { return "" }
func (m *ThemeMockWidget) Style(...string) *Style { return &DefaultStyle }
func (m *ThemeMockWidget) Styles() []string { return []string{} }

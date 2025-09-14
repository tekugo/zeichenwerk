package zeichenwerk

// NordTheme creates a new theme inspired by the Nord color palette.
// Nord is a arctic, north-bluish color palette with a clean and minimal design.
// The theme features cool blues, whites, and subtle accent colors that provide
// excellent readability and a calming, professional appearance.
//
// Color palette inspiration:
//   - Polar Night: Dark blues and grays for backgrounds
//   - Snow Storm: Light colors for text and highlights  
//   - Frost: Blue accent colors for interactive elements
//   - Aurora: Colorful accents for warnings, errors, and highlights
//
// Returns:
//   - Theme: A complete Nord-inspired theme ready for use
func NordTheme() Theme {
	t := NewMapTheme()

	AddUnicodeBorders(t)

	// Nord color palette
	// Polar Night (dark backgrounds)
	// Snow Storm (light foregrounds)
	// Frost (blue accents)
	// Aurora (colorful highlights)
	t.SetColors(map[string]string{
		// Polar Night - Dark backgrounds
		"$bg0":     "#2e3440", // Nord0 - darkest background
		"$bg1":     "#3b4252", // Nord1 - dark background
		"$bg2":     "#434c5e", // Nord2 - medium background
		"$bg3":     "#4c566a", // Nord3 - light background

		// Snow Storm - Light foregrounds
		"$fg0":     "#eceff4", // Nord4 - lightest foreground
		"$fg1":     "#e5e9f0", // Nord5 - light foreground
		"$fg2":     "#d8dee9", // Nord6 - medium foreground

		// Frost - Blue accents
		"$frost1":  "#8fbcbb", // Nord7 - light blue
		"$frost2":  "#88c0d0", // Nord8 - medium blue
		"$frost3":  "#81a1c1", // Nord9 - dark blue
		"$frost4":  "#5e81ac", // Nord10 - darkest blue

		// Aurora - Colorful highlights
		"$red":     "#bf616a", // Nord11 - red
		"$orange":  "#d08770", // Nord12 - orange
		"$yellow":  "#ebcb8b", // Nord13 - yellow
		"$green":   "#a3be8c", // Nord14 - green
		"$purple":  "#b48ead", // Nord15 - purple

		// Aliases for compatibility
		"$blue":    "#81a1c1", // Alias to frost3
		"$cyan":    "#88c0d0", // Alias to frost2
		"$aqua":    "#8fbcbb", // Alias to frost1
		"$magenta": "#b48ead", // Alias to purple
		"$gray":    "#4c566a", // Alias to bg3
	})

	t.SetStyles(map[string]*Style{
		// Default widget styles
		"": NewStyle("$fg0", "$bg0").SetMargin(0).SetPadding(0),

		// Button styles
		"button":         NewStyle("$fg0", "$frost3").SetBorder("lines").SetPadding(0, 2),
		"button:focus":   NewStyle("$bg0", "$frost2"),
		"button:hover":   NewStyle("$bg0", "$frost1"),
		"button:pressed": NewStyle("$bg0", "$frost4"),
		"button:disabled": NewStyle("$bg3", "$bg1"),

		// Checkbox styles
		"checkbox":          NewStyle("$fg0", "").SetPadding(0),
		"checkbox:focus":    NewStyle("$frost2", ""),
		"checkbox:hover":    NewStyle("$frost1", ""),
		"checkbox:disabled": NewStyle("$bg3", ""),

		// Input styles
		"input":             NewStyle("$fg0", "$bg2").SetCursor("*bar").SetBorder("thin"),
		"input:focus":       NewStyle("$fg0", "$bg1").SetBorder("double"),
		"input:placeholder": NewStyle("$bg3", "$bg2"),

		// Label styles
		"label": NewStyle("$fg0", ""),

		// List styles
		"list":                NewStyle("$fg0", "$bg1").SetBorder("thin"),
		"list:focus":          NewStyle("$fg0", "$bg1").SetBorder("double"),
		"list:disabled":       NewStyle("$bg3", "$bg2"),
		"list/highlight":      NewStyle("$bg0", "$bg3"),
		"list/highlight:focus": NewStyle("$bg0", "$frost2"),

		// Progress bar styles
		"progress-bar":     NewStyle("$bg3", "$bg1").SetRender("unicode").SetBorder("thin"),
		"progress-bar/bar": NewStyle("$frost2", ""),

		// Grid styles
		"grid": NewStyle("$bg3", "$bg0").SetBorder("thin"),

		// Box styles
		"box":       NewStyle("$fg0", "").SetBorder("round"),
		"box:focus": NewStyle("$frost2", "").SetBorder("double"),

		// Flex styles
		"flex": NewStyle("$fg0", ""),

		// Scroller styles
		"scroller":       NewStyle("$fg1", "$bg1").SetBorder("thin"),
		"scroller:focus": NewStyle("$frost2", "$bg1").SetBorder("double"),

		// Tabs styles
		"tabs":                     NewStyle("$fg1", "$bg1"),
		"tabs:focus":               NewStyle("$frost2", "$bg1"),
		"tabs/line":                NewStyle("$bg3", ""),
		"tabs/line:focus":          NewStyle("$frost2", ""),
		"tabs/highlight":           NewStyle("$bg0", "$fg2"),
		"tabs/highlight:focus":     NewStyle("$bg0", "$frost2"),
		"tabs/highlight-line":      NewStyle("$fg2", ""),
		"tabs/highlight-line:focus": NewStyle("$frost2", ""),

		// Separator styles
		"separator": NewStyle("$bg3", ""),

		// Text styles
		"text": NewStyle("$fg0", "").SetBorder("thin"),

		// Header style
		".header": NewStyle("$fg0", "$bg2"),

		// Footer style
		".footer": NewStyle("$fg1", "$bg2"),

		// Inspector style
		".inspector":          NewStyle("$fg0", "$bg1").SetBorder("double"),
		"box.inspector":       NewStyle("$fg0", "$bg1"),
		"box.inspector:title": NewStyle("$frost2", ""),

		// Popup styles
		".popup":            NewStyle("$fg0", "$bg1").SetBorder("double"),
		"flex/shadow.popup": NewStyle("$bg2", "$bg0"),
		"button.popup":      NewStyle("$fg0", "$frost3"),
		".popup#title":      NewStyle("$bg0", "$frost2"),

		// Shortcut key styles
		".shortcut": NewStyle("$frost2", "$bg2").SetPadding(0, 1),

		// Debug log styles
		"#debug-log": NewStyle("$green", "$bg1"),

		// Special class styles for status indicators
		".success": NewStyle("$green", ""),
		".warning": NewStyle("$yellow", ""),
		".error":   NewStyle("$red", ""),
		".info":    NewStyle("$frost2", ""),

		// Disabled state for any widget
		":disabled": NewStyle("$bg3", "$bg1"),

		// Spacer (invisible)
		"spacer": NewStyle("", ""),
	})

	return t
}
package zeichenwerk

// GruvboxDarkTheme creates a new theme inspired by the Gruvbox Dark color palette.
// Gruvbox is a retro groove color scheme with warm, earthy colors designed for
// comfortable long-term use. The dark variant features warm dark backgrounds
// with bright, saturated foreground colors that provide excellent contrast.
//
// Color philosophy:
//   - Warm, earthy base colors that are easy on the eyes
//   - High contrast between background and foreground
//   - Retro aesthetic with modern usability
//   - Carefully balanced color relationships
//
// Returns:
//   - Theme: A complete Gruvbox Dark theme ready for use
func GruvboxDarkTheme() Theme {
	t := NewMapTheme()

	AddUnicodeBorders(t)

	// Gruvbox Dark color palette
	t.SetColors(map[string]string{
		// Dark backgrounds
		"$bg0":     "#282828", // Hard dark
		"$bg1":     "#3c3836", // Medium dark
		"$bg2":     "#504945", // Soft dark
		"$bg3":     "#665c54", // Light dark
		"$bg4":     "#7c6f64", // Lightest dark

		// Light foregrounds
		"$fg0":     "#fbf1c7", // Hard light
		"$fg1":     "#ebdbb2", // Medium light
		"$fg2":     "#d5c4a1", // Soft light
		"$fg3":     "#bdae93", // Dark light
		"$fg4":     "#a89984", // Darkest light

		// Neutral colors
		"$gray":    "#928374", // Gray
		"$red":     "#fb4934", // Bright red
		"$green":   "#b8bb26", // Bright green
		"$yellow":  "#fabd2f", // Bright yellow
		"$blue":    "#83a598", // Bright blue
		"$purple":  "#d3869b", // Bright purple
		"$aqua":    "#8ec07c", // Bright aqua
		"$orange":  "#fe8019", // Bright orange

		// Faded colors for subtle elements
		"$red_dim":     "#cc241d", // Faded red
		"$green_dim":   "#98971a", // Faded green
		"$yellow_dim":  "#d79921", // Faded yellow
		"$blue_dim":    "#458588", // Faded blue
		"$purple_dim":  "#b16286", // Faded purple
		"$aqua_dim":    "#689d6a", // Faded aqua
		"$orange_dim":  "#d65d0e", // Faded orange

		// Aliases for compatibility
		"$cyan":    "#8ec07c", // Alias to aqua
		"$magenta": "#d3869b", // Alias to purple
	})

	t.SetStyles(map[string]*Style{
		// Default widget styles
		"": NewStyle("$fg1", "$bg0").SetMargin(0).SetPadding(0),

		// Button styles
		"button":         NewStyle("$bg0", "$yellow").SetBorder("lines").SetPadding(0, 2),
		"button:focus":   NewStyle("$bg0", "$orange"),
		"button:hover":   NewStyle("$bg0", "$yellow_dim"),
		"button:pressed": NewStyle("$fg0", "$orange_dim"),
		"button:disabled": NewStyle("$bg3", "$bg1"),

		// Checkbox styles
		"checkbox":          NewStyle("$fg1", "").SetPadding(0),
		"checkbox:focus":    NewStyle("$yellow", ""),
		"checkbox:hover":    NewStyle("$orange", ""),
		"checkbox:disabled": NewStyle("$bg3", ""),

		// Input styles
		"input":             NewStyle("$fg0", "$bg1").SetCursor("*bar").SetBorder("thin"),
		"input:focus":       NewStyle("$fg0", "$bg0").SetBorder("double"),
		"input:placeholder": NewStyle("$bg4", "$bg1"),

		// Label styles
		"label": NewStyle("$fg1", ""),

		// List styles
		"list":                NewStyle("$fg1", "$bg1").SetBorder("thin"),
		"list:focus":          NewStyle("$fg1", "$bg1").SetBorder("double"),
		"list:disabled":       NewStyle("$bg3", "$bg2"),
		"list/highlight":      NewStyle("$bg0", "$bg3"),
		"list/highlight:focus": NewStyle("$bg0", "$yellow"),

		// Progress bar styles
		"progress-bar":     NewStyle("$bg3", "$bg1").SetRender("unicode").SetBorder("thin"),
		"progress-bar/bar": NewStyle("$green", ""),

		// Grid styles
		"grid": NewStyle("$bg3", "$bg0").SetBorder("thin"),

		// Box styles
		"box":       NewStyle("$fg1", "").SetBorder("round"),
		"box:focus": NewStyle("$yellow", "").SetBorder("double"),

		// Flex styles
		"flex": NewStyle("$fg1", ""),

		// Scroller styles
		"scroller":       NewStyle("$fg2", "$bg1").SetBorder("thin"),
		"scroller:focus": NewStyle("$yellow", "$bg1").SetBorder("double"),

		// Tabs styles
		"tabs":                     NewStyle("$fg2", "$bg1"),
		"tabs:focus":               NewStyle("$yellow", "$bg1"),
		"tabs/line":                NewStyle("$bg3", ""),
		"tabs/line:focus":          NewStyle("$yellow", ""),
		"tabs/highlight":           NewStyle("$bg0", "$fg3"),
		"tabs/highlight:focus":     NewStyle("$bg0", "$yellow"),
		"tabs/highlight-line":      NewStyle("$fg3", ""),
		"tabs/highlight-line:focus": NewStyle("$yellow", ""),

		// Separator styles
		"separator": NewStyle("$bg3", ""),

		// Text styles
		"text": NewStyle("$fg1", "").SetBorder("thin"),

		// Header style
		".header": NewStyle("$fg0", "$bg2"),

		// Footer style
		".footer": NewStyle("$fg2", "$bg2"),

		// Inspector style
		".inspector":          NewStyle("$fg1", "$bg1").SetBorder("double"),
		"box.inspector":       NewStyle("$fg1", "$bg1"),
		"box.inspector:title": NewStyle("$yellow", ""),

		// Popup styles
		".popup":            NewStyle("$fg1", "$bg1").SetBorder("double"),
		"flex/shadow.popup": NewStyle("$bg2", "$bg0"),
		"button.popup":      NewStyle("$bg0", "$yellow"),
		".popup#title":      NewStyle("$bg0", "$orange"),

		// Shortcut key styles
		".shortcut": NewStyle("$orange", "$bg2").SetPadding(0, 1),

		// Debug log styles
		"#debug-log": NewStyle("$green", "$bg1"),

		// Status indicator styles
		".success": NewStyle("$green", ""),
		".warning": NewStyle("$yellow", ""),
		".error":   NewStyle("$red", ""),
		".info":    NewStyle("$blue", ""),

		// Disabled state for any widget
		":disabled": NewStyle("$bg3", "$bg1"),

		// Spacer (invisible)
		"spacer": NewStyle("", ""),
	})

	return t
}

// GruvboxLightTheme creates a new theme inspired by the Gruvbox Light color palette.
// This is the light variant of Gruvbox, featuring warm light backgrounds with
// darker foreground colors while maintaining the same retro aesthetic and
// excellent readability of the dark variant.
//
// Color philosophy:
//   - Warm, cream-colored backgrounds that reduce eye strain
//   - Dark, high-contrast foreground colors
//   - Maintains Gruvbox's retro aesthetic in light mode
//   - Inverted color relationships from the dark theme
//
// Returns:
//   - Theme: A complete Gruvbox Light theme ready for use
func GruvboxLightTheme() Theme {
	t := NewMapTheme()

	AddUnicodeBorders(t)

	// Gruvbox Light color palette (inverted from dark)
	t.SetColors(map[string]string{
		// Light backgrounds
		"$bg0":     "#fbf1c7", // Hard light
		"$bg1":     "#ebdbb2", // Medium light
		"$bg2":     "#d5c4a1", // Soft light
		"$bg3":     "#bdae93", // Dark light
		"$bg4":     "#a89984", // Darkest light

		// Dark foregrounds
		"$fg0":     "#282828", // Hard dark
		"$fg1":     "#3c3836", // Medium dark
		"$fg2":     "#504945", // Soft dark
		"$fg3":     "#665c54", // Light dark
		"$fg4":     "#7c6f64", // Lightest dark

		// Neutral colors (darker variants for light theme)
		"$gray":    "#928374", // Gray
		"$red":     "#cc241d", // Dark red
		"$green":   "#98971a", // Dark green
		"$yellow":  "#d79921", // Dark yellow
		"$blue":    "#458588", // Dark blue
		"$purple":  "#b16286", // Dark purple
		"$aqua":    "#689d6a", // Dark aqua
		"$orange":  "#d65d0e", // Dark orange

		// Bright colors for highlights
		"$red_bright":     "#fb4934", // Bright red
		"$green_bright":   "#b8bb26", // Bright green
		"$yellow_bright":  "#fabd2f", // Bright yellow
		"$blue_bright":    "#83a598", // Bright blue
		"$purple_bright":  "#d3869b", // Bright purple
		"$aqua_bright":    "#8ec07c", // Bright aqua
		"$orange_bright":  "#fe8019", // Bright orange

		// Aliases for compatibility
		"$cyan":    "#689d6a", // Alias to aqua
		"$magenta": "#b16286", // Alias to purple
	})

	t.SetStyles(map[string]*Style{
		// Default widget styles
		"": NewStyle("$fg1", "$bg0").SetMargin(0).SetPadding(0),

		// Button styles
		"button":         NewStyle("$bg0", "$yellow").SetBorder("lines").SetPadding(0, 2),
		"button:focus":   NewStyle("$bg0", "$orange"),
		"button:hover":   NewStyle("$bg0", "$yellow_bright"),
		"button:pressed": NewStyle("$fg0", "$orange_bright"),
		"button:disabled": NewStyle("$bg3", "$bg1"),

		// Checkbox styles
		"checkbox":          NewStyle("$fg1", "").SetPadding(0),
		"checkbox:focus":    NewStyle("$orange", ""),
		"checkbox:hover":    NewStyle("$yellow", ""),
		"checkbox:disabled": NewStyle("$bg3", ""),

		// Input styles
		"input":             NewStyle("$fg0", "$bg1").SetCursor("*bar").SetBorder("thin"),
		"input:focus":       NewStyle("$fg0", "$bg0").SetBorder("double"),
		"input:placeholder": NewStyle("$bg4", "$bg1"),

		// Label styles
		"label": NewStyle("$fg1", ""),

		// List styles
		"list":                NewStyle("$fg1", "$bg1").SetBorder("thin"),
		"list:focus":          NewStyle("$fg1", "$bg1").SetBorder("double"),
		"list:disabled":       NewStyle("$bg3", "$bg2"),
		"list/highlight":      NewStyle("$bg0", "$bg3"),
		"list/highlight:focus": NewStyle("$bg0", "$orange"),

		// Progress bar styles
		"progress-bar":     NewStyle("$bg3", "$bg1").SetRender("unicode").SetBorder("thin"),
		"progress-bar/bar": NewStyle("$green", ""),

		// Grid styles
		"grid": NewStyle("$bg3", "$bg0").SetBorder("thin"),

		// Box styles
		"box":       NewStyle("$fg1", "").SetBorder("round"),
		"box:focus": NewStyle("$orange", "").SetBorder("double"),

		// Flex styles
		"flex": NewStyle("$fg1", ""),

		// Scroller styles
		"scroller":       NewStyle("$fg2", "$bg1").SetBorder("thin"),
		"scroller:focus": NewStyle("$orange", "$bg1").SetBorder("double"),

		// Tabs styles
		"tabs":                     NewStyle("$fg2", "$bg1"),
		"tabs:focus":               NewStyle("$orange", "$bg1"),
		"tabs/line":                NewStyle("$bg3", ""),
		"tabs/line:focus":          NewStyle("$orange", ""),
		"tabs/highlight":           NewStyle("$bg0", "$fg3"),
		"tabs/highlight:focus":     NewStyle("$bg0", "$orange"),
		"tabs/highlight-line":      NewStyle("$fg3", ""),
		"tabs/highlight-line:focus": NewStyle("$orange", ""),

		// Separator styles
		"separator": NewStyle("$bg3", ""),

		// Text styles
		"text": NewStyle("$fg1", "").SetBorder("thin"),

		// Header style
		".header": NewStyle("$fg0", "$bg2"),

		// Footer style
		".footer": NewStyle("$fg2", "$bg2"),

		// Inspector style
		".inspector":          NewStyle("$fg1", "$bg1").SetBorder("double"),
		"box.inspector":       NewStyle("$fg1", "$bg1"),
		"box.inspector:title": NewStyle("$orange", ""),

		// Popup styles
		".popup":            NewStyle("$fg1", "$bg1").SetBorder("double"),
		"flex/shadow.popup": NewStyle("$bg2", "$bg0"),
		"button.popup":      NewStyle("$bg0", "$orange"),
		".popup#title":      NewStyle("$bg0", "$yellow"),

		// Shortcut key styles
		".shortcut": NewStyle("$yellow", "$bg2").SetPadding(0, 1),

		// Debug log styles
		"#debug-log": NewStyle("$green", "$bg1"),

		// Status indicator styles
		".success": NewStyle("$green", ""),
		".warning": NewStyle("$yellow", ""),
		".error":   NewStyle("$red", ""),
		".info":    NewStyle("$blue", ""),

		// Disabled state for any widget
		":disabled": NewStyle("$bg3", "$bg1"),

		// Spacer (invisible)
		"spacer": NewStyle("", ""),
	})

	return t
}
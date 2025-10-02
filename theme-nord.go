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
		"$bg0": "#2e3440", // Nord0 - darkest background
		"$bg1": "#3b4252", // Nord1 - dark background
		"$bg2": "#434c5e", // Nord2 - medium background
		"$bg3": "#4c566a", // Nord3 - light background

		// Snow Storm - Light foregrounds
		"$fg0": "#eceff4", // Nord4 - lightest foreground
		"$fg1": "#e5e9f0", // Nord5 - light foreground
		"$fg2": "#d8dee9", // Nord6 - medium foreground

		// Frost - Blue accents
		"$frost1": "#8fbcbb", // Nord7 - light blue
		"$frost2": "#88c0d0", // Nord8 - medium blue
		"$frost3": "#81a1c1", // Nord9 - dark blue
		"$frost4": "#5e81ac", // Nord10 - darkest blue

		// Aurora - Colorful highlights
		"$red":    "#bf616a", // Nord11 - red
		"$orange": "#d08770", // Nord12 - orange
		"$yellow": "#ebcb8b", // Nord13 - yellow
		"$green":  "#a3be8c", // Nord14 - green
		"$purple": "#b48ead", // Nord15 - purple

		// Aliases for compatibility
		"$blue":    "#81a1c1", // Alias to frost3
		"$cyan":    "#88c0d0", // Alias to frost2
		"$aqua":    "#8fbcbb", // Alias to frost1
		"$magenta": "#b48ead", // Alias to purple
		"$gray":    "#4c566a", // Alias to bg3
	})

	return t
}


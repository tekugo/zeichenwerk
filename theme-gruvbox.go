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
		"$bg0": "#282828", // Hard dark
		"$bg1": "#3c3836", // Medium dark
		"$bg2": "#504945", // Soft dark
		"$bg3": "#665c54", // Light dark
		"$bg4": "#7c6f64", // Lightest dark

		// Light foregrounds
		"$fg0": "#fbf1c7", // Hard light
		"$fg1": "#ebdbb2", // Medium light
		"$fg2": "#d5c4a1", // Soft light
		"$fg3": "#bdae93", // Dark light
		"$fg4": "#a89984", // Darkest light

		// Neutral colors
		"$gray":   "#928374", // Gray
		"$red":    "#fb4934", // Bright red
		"$green":  "#b8bb26", // Bright green
		"$yellow": "#fabd2f", // Bright yellow
		"$blue":   "#83a598", // Bright blue
		"$purple": "#d3869b", // Bright purple
		"$aqua":   "#8ec07c", // Bright aqua
		"$orange": "#fe8019", // Bright orange

		// Faded colors for subtle elements
		"$red_dim":    "#cc241d", // Faded red
		"$green_dim":  "#98971a", // Faded green
		"$yellow_dim": "#d79921", // Faded yellow
		"$blue_dim":   "#458588", // Faded blue
		"$purple_dim": "#b16286", // Faded purple
		"$aqua_dim":   "#689d6a", // Faded aqua
		"$orange_dim": "#d65d0e", // Faded orange

		// Aliases for compatibility
		"$cyan":    "#8ec07c", // Alias to aqua
		"$magenta": "#d3869b", // Alias to purple
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
		"$bg0": "#fbf1c7", // Hard light
		"$bg1": "#ebdbb2", // Medium light
		"$bg2": "#d5c4a1", // Soft light
		"$bg3": "#bdae93", // Dark light
		"$bg4": "#a89984", // Darkest light

		// Dark foregrounds
		"$fg0": "#282828", // Hard dark
		"$fg1": "#3c3836", // Medium dark
		"$fg2": "#504945", // Soft dark
		"$fg3": "#665c54", // Light dark
		"$fg4": "#7c6f64", // Lightest dark

		// Neutral colors (darker variants for light theme)
		"$gray":   "#928374", // Gray
		"$red":    "#cc241d", // Dark red
		"$green":  "#98971a", // Dark green
		"$yellow": "#d79921", // Dark yellow
		"$blue":   "#458588", // Dark blue
		"$purple": "#b16286", // Dark purple
		"$aqua":   "#689d6a", // Dark aqua
		"$orange": "#d65d0e", // Dark orange

		// Bright colors for highlights
		"$red_bright":    "#fb4934", // Bright red
		"$green_bright":  "#b8bb26", // Bright green
		"$yellow_bright": "#fabd2f", // Bright yellow
		"$blue_bright":   "#83a598", // Bright blue
		"$purple_bright": "#d3869b", // Bright purple
		"$aqua_bright":   "#8ec07c", // Bright aqua
		"$orange_bright": "#fe8019", // Bright orange

		// Aliases for compatibility
		"$cyan":    "#689d6a", // Alias to aqua
		"$magenta": "#b16286", // Alias to purple
	})

	return t
}


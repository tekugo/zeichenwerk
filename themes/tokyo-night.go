package themes

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// TokyoNightTheme returns a Theme styled after the Tokyo Night colour palette.
// It registers Unicode borders via AddUnicodeBorders and sets colours, styles,
// and string indicators for all built-in widgets.
func TokyoNight() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)
	AddNerdStrings(t)
	AddDefaultStyles(t)

	t.SetColors(map[string]string{
		"$bg0":     "#1a1b26",
		"$bg1":     "#1e1e2e",
		"$bg2":     "#1b263b",
		"$bg3":     "#292e42", // highlight background
		"$fg0":     "#c0caf5",
		"$fg1":     "#a9b1d6", // secondary text
		"$fg2":     "#565f89", // muted / comment text
		"$gray":    "#3b4261", // decorative / line numbers
		"$blue":    "#7aa2f7",
		"$cyan":    "#2ac3de",
		"$aqua":    "#89ddff",
		"$magenta": "#bb9af7",
		"$red":     "#f7768e",
		"$orange":  "#ff9e64",
		"$yellow":  "#e0af68",
		"$green":   "#9ece6a",
	})

	return t
}

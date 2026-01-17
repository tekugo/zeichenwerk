package next

func TokyoNightTheme() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)

	t.SetColors(map[string]string{
		"$bg0":     "#1a1b26",
		"$bg1":     "#1e1e2e",
		"$bg2":     "#1b263b",
		"$fg0":     "#c0caf5",
		"$fg1":     "#565f89",
		"$gray":    "#414868",
		"$blue":    "#7aa2f7",
		"$cyan":    "#2ac3de",
		"$aqua":    "#89ddff",
		"$magenta": "#bb9af7",
		"$red":     "#f7768e",
		"$orange":  "#ff9e64",
		"$yellow":  "#e0af68",
		"$green":   "#9ece6a",
	})

	t.SetStyles(
		NewStyle("").WithColors("$fg0", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("box").WithBorder("thin"),
		NewStyle("flex"),
		NewStyle("grid").WithBorder("thin"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg1"),
		NewStyle("list/highlight:focus").WithColors("$bg0", "$red"),
		NewStyle("static").WithColors("$fg0", "$bg1").WithMargin(0).WithPadding(0),
	)

	return t
}

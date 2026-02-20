package zeichenwerk

func TokyoNightTheme() Theme {
	t := NewMapTheme()

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
		NewStyle("button").WithColors("$bg0", "$blue").WithBorder("lines").WithPadding(0, 2),
		NewStyle("button:focus").WithForeground("$fg0").WithBackground("$blue"),
		NewStyle("button:hover").WithColors("$red", "$blue"),
		NewStyle("button.dialog").WithColors("$fg1", "$bg2").WithBorder("none").WithPadding(0, 2),
		NewStyle("dialog").WithColors("$fg0", "$blue").WithBorder("thick").WithPadding(1, 2),
		NewStyle("digits").WithForeground("$green"),
		NewStyle("editor").WithCursor("*block"),
		NewStyle("editor/current-line").WithBackground("$gray"),
		NewStyle("editor/current-line-number").WithBackground("$gray"),
		NewStyle("editor/line-numbers").WithForeground("$green"),
		NewStyle("editor/separator").WithForeground("$gray"),
		NewStyle("form").WithColors("$fg0", "$bg0"),
		NewStyle("form-group").WithColors("$fg1", "$bg0"),
		NewStyle("grid").WithColors("$fg1", "$bg0").WithBorder("thin"),
		NewStyle("input").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("input:focus").WithColors("$bg0", "$blue"),
		NewStyle("label").WithForeground("fg0"),
		NewStyle("list/highlight:focus").WithColors("$bg0", "$red"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg1"),
		NewStyle("progress-bar").WithForeground("$fg1"),
		NewStyle("progress-bar/bar").WithForeground("$orange"),
		NewStyle("scroller").WithColors("$fg1", "$bg2"),
		NewStyle("table").WithColors("", ""),
		NewStyle("table/grid").WithColors("$fg1", "$bg0").WithBorder("thin"),
		NewStyle("table/header").WithColors("$fg0", "$bg0"),
		NewStyle("table/highlight").WithColors("$bg0", "$fg1"),
		NewStyle("table/highlight:focus").WithColors("$bg0", "$red"),
		NewStyle("tabs/highlight").WithColors("$bg0", "$fg0"),
		NewStyle("tabs/highlight-line").WithForeground("$bg3"),
		NewStyle("tabs/line:focus").WithForeground("$blue"),
		NewStyle("tabs/highlight:focus").WithColors("$bg0", "$orange"),
		NewStyle("tabs/highlight-line:focus").WithForeground("$orange"),

		// Dialog styles
		NewStyle(".dialog").WithColors("$fg0", "$blue"),
		NewStyle("button.dialog").WithColors("$fg0", "$red"),

		// Header & Footer style
		NewStyle(".header").WithColors("$fg0", "$fg1"),
		NewStyle(".footer").WithColors("$fg0", "$fg1"),

		// Inspector style
		NewStyle(".inspector").WithColors("", "$bg2"),
		NewStyle("box.inspector:title").WithColors("$cyan", ""),
		NewStyle(".shortcut").WithColors("$cyan", "$fg1").WithPadding(0, 1),
		NewStyle("#debug-log").WithColors("$green", "$bg1"),
	)

	return t
}

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

	t.SetStyles(map[string]*Style{
		// Default widget styles
		"":                          NewStyle("$fg0", "$bg0").SetMargin(0).SetPadding(0),
		"button":                    NewStyle("$bg0", "$blue").SetBorder("lines").SetPadding(0, 2),
		"button:focus":              NewStyle("$fg0", "$blue"),
		"button:hover":              NewStyle("$red", "$blue"),
		"editor":                    NewStyle("", "").SetCursor("*block"),
		"grid":                      NewStyle("$fg1", "$bg0").SetBorder("thin"),
		"input":                     NewStyle("$fg0", "$bg2").SetCursor("*bar"),
		"input:focus":               NewStyle("$bg0", "$blue"),
		"label":                     NewStyle("fg0", ""),
		"list/highlight:focus":      NewStyle("$bg0", "$red"),
		"list/highlight":            NewStyle("$bg0", "$fg1"),
		"progress-bar":              NewStyle("$fg1", "").SetRender("unicode"),
		"scroller":                  NewStyle("$fg1", "$bg2"),
		"progress-bar/bar":          NewStyle("$orange", ""),
		"table":                     NewStyle("", "").SetBorder("thin"),
		"table/highlight":           NewStyle("$bg0", "fg1"),
		"table/highlight:focus":     NewStyle("$bg0", "$red"),
		"tabs/highlight":            NewStyle("$bg0", "$fg0"),
		"tabs/highlight-line":       NewStyle("$bg3", ""),
		"tabs/line:focus":           NewStyle("$blue", ""),
		"tabs/highlight:focus":      NewStyle("$bg0", "$orange"),
		"tabs/highlight-line:focus": NewStyle("$orange", ""),

		// Header style
		".header": NewStyle("$fg0", "$fg1"),

		// Inspector style
		".inspector":          NewStyle("", "$bg2"),
		"box.inspector:title": NewStyle("$cyan", ""),

		".footer": NewStyle("$fg0", "$fg1"),

		".popup":            NewStyle("", "$bg2"),
		"flex/shadow.popup": NewStyle("$bg1", "black"),
		"button.popup":      NewStyle("", "$cyan"),
		".popup#title":      NewStyle("$bg0", "$fg1"),

		".shortcut": NewStyle("$cyan", "$fg1").SetPadding(0, 1),

		"#debug-log": NewStyle("$green", "$bg1"),
	})

	return t
}

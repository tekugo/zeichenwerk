package zeichenwerk

func MidnightNeonTheme() Theme {
	t := NewMapTheme()

	AddUnicodeBorders(t)

	t.SetColors(map[string]string{
		"$bg0":     "#0f1117",
		"$bg1":     "#1a1c23",
		"$bg2":     "#242730",
		"$bg3":     "#2f323d",
		"$fg0":     "#5ee9f0",
		"$fg1":     "#c7ccd9",
		"$fg2":     "#a0a4b3",
		"$fg3":     "#6c7384",
		"$blue":    "#5aaaff",
		"$cyan":    "#40e000",
		"$green":   "#4cd964",
		"$yellow":  "#ffd866",
		"$orange":  "#ff9f43",
		"$magenta": "#c792ea",
	})

	t.SetStyles(map[string]*Style{
		// Default widget styles
		"":                          NewStyle("$fg0", "$bg0").SetMargin(0).SetPadding(0),
		"button":                    NewStyle("$fg1", "$bg3").SetBorder("lines").SetPadding(0, 2),
		"button:focus":              NewStyle("$cyan", "$bg3"),
		"button:hover":              NewStyle("$fg0", "$bg3"),
		"grid":                      NewStyle("$fg2", "$bg0").SetBorder("thin"),
		"input":                     NewStyle("$fg0", "$bg2").SetCursor("*bar").SetBorder("round"),
		"input:focus":               NewStyle("$bg0", "$blue"),
		"list/highlight:focus":      NewStyle("$bg0", "$cyan"),
		"list/highlight":            NewStyle("$bg0", "$fg1"),
		"progress-bar":              NewStyle("$fg1", "").SetRender("unicode"),
		"scroller":                  NewStyle("$fg1", "$bg2"),
		"progress-bar/bar":          NewStyle("$orange", ""),
		"tabs/highlight":            NewStyle("$bg0", "$orange"),
		"tabs/highlight-line":       NewStyle("$orange", ""),
		"tabs/line:focus":           NewStyle("$blue", ""),
		"tabs/highlight:focus":      NewStyle("$bg0", "$cyan"),
		"tabs/highlight-line:focus": NewStyle("$cyan", ""),

		// Header style
		".header": NewStyle("$fg0", "$fg1"),

		// Inspector style
		".inspector":          NewStyle("", "$bg2"),
		"box.inspector:title": NewStyle("$cyan", ""),

		".footer": NewStyle("$fg0", "$fg1"),

		".popup":       NewStyle("", "$fg1"),
		"button.popup": NewStyle("", "$cyan"),
		".popup#title": NewStyle("$bg0", "$fg0"),

		".shortcut": NewStyle("$cyan", "$fg1").SetPadding(0, 1),

		"#popup/shadow": NewStyle("$b1", "black"),
		"#debug-log":    NewStyle("$green", "$bg1"),
	})

	return t
}

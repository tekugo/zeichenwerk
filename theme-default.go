package zeichenwerk

func DefaultTheme() Theme {
	theme := MapTheme{styles: make(map[string]Style)}

	theme.Set("flex", NewStyle("white", ""))
	theme.Set("grid", NewStyle("white", "blue").SetBorder("thick-slashed"))
	theme.Set("input", NewStyle("white", "black").SetCursor("bar"))
	theme.Set("label", NewStyle("yellow", "green"))
	theme.Set("label.header", NewStyle("white", "#774433"))
	theme.Set("list", NewStyle("green", "black").SetBorder("thin").SetMargin(1, 2, 1, 2))
	theme.Set("list:focus", NewStyle("green", "black").SetBorder("double").SetMargin(1, 2, 1, 2))
	theme.Set("button", NewStyle("white", "green").SetBorder("lines").SetPadding(0, 2, 0, 2))
	theme.Set("button:focus", NewStyle("red", "white").SetBorder("lines").SetPadding(0, 2, 0, 2))
	theme.Set("button:hover", NewStyle("white", "dark-red").SetBorder("lines").SetPadding(0, 2, 0, 2))
	theme.Set("flex#header", NewStyle("white", "#774433"))
	theme.Set("flex#footer", NewStyle("white", "#334477"))

	return &theme
}

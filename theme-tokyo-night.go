package zeichenwerk

func TokyoNightTheme() Theme {
	theme := MapTheme{styles: make(map[string]Style), colors: make(map[string]string)}

	theme.colors["$bg"] = "#1a1b26"
	theme.colors["$bg2"] = "#1e1e2e"
	theme.colors["$bg3"] = "#1b263b"

	theme.colors["$fg"] = "#c0caf5"
	theme.colors["$comments"] = "#565f89"
	theme.colors["$gray"] = "#414868"
	theme.colors["$blue"] = "#7aa2f7"
	theme.colors["$cyan"] = "#2ac3de"
	theme.colors["$aqua"] = "#89ddff"
	theme.colors["$magenta"] = "#bb9af7"
	theme.colors["$red"] = "#f7768e"
	theme.colors["$orange"] = "#ff9e64"
	theme.colors["$yellow"] = "#e0af68"
	theme.colors["$green"] = "#9ece6a"

	theme.Set("", NewStyle("$fg", "$bg").SetMargin(0).SetPadding(0))
	theme.Set("button", NewStyle("$bg", "$blue").SetBorder("lines").SetPadding(0, 2))
	theme.Set("button:focus", NewStyle("white", "$blue"))
	theme.Set("grid", NewStyle("$comments", "$bg").SetBorder("thin"))
	theme.Set("input", NewStyle("$fg", "$bg2").SetCursor("*bar").SetBorder("round"))
	theme.Set("input:focus", NewStyle("$bg", "$blue"))
	theme.Set("list", NewStyle("", "").SetBorder("round"))
	theme.Set("list:focus", NewStyle("", "").SetBorder("double"))
	theme.Set("list:highlight", NewStyle("$bg", "$red"))
	theme.Set("list:highlight-blurred", NewStyle("$bg", "$blue"))
	theme.Set("progress-bar", NewStyle("$comments", "").SetRender("unicode"))
	theme.Set("progress-bar:bar", NewStyle("$orange", ""))
	theme.Set(".header", NewStyle("$fg", "$comments"))
	theme.Set(".inspector", NewStyle("", "$bg3"))
	theme.Set("box.inspector:title", NewStyle("$cyan", ""))
	theme.Set(".footer", NewStyle("$fg", "$comments"))
	theme.Set(".popup", NewStyle("", "$comments"))
	theme.Set("button.popup", NewStyle("", "$cyan"))
	theme.Set(".popup#title", NewStyle("$bg", "$fg"))
	theme.Set(".shortcut", NewStyle("$cyan", "$comments").SetPadding(0, 1))
	theme.Set("#popup:shadow", NewStyle("$bg2", "black"))
	theme.Set("#debug-log", NewStyle("$green", "$bg2"))

	return &theme
}

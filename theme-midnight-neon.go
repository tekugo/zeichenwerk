package zeichenwerk

// MidnightNeonTheme returns a Theme styled after a dark midnight palette with
// neon accent colours. It features near-black backgrounds with vivid electric
// highlights — cyan, blue, green, and magenta — for a futuristic aesthetic.
func MidnightNeonTheme() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)

	t.SetColors(map[string]string{
		// Dark backgrounds
		"$bg0": "#0f1117",
		"$bg1": "#1a1c23",
		"$bg2": "#242730",
		"$bg3": "#2f323d",

		// Foregrounds
		"$fg0": "#5ee9f0", // Electric cyan — primary foreground
		"$fg1": "#c7ccd9", // Light grey
		"$fg2": "#a0a4b3", // Medium grey
		"$fg3": "#6c7384", // Dark grey

		// Neon accent colours
		"$blue":    "#5aaaff",
		"$cyan":    "#40e000", // Bright neon green (used as "cyan" in this palette)
		"$green":   "#4cd964",
		"$yellow":  "#ffd866",
		"$orange":  "#ff9f43",
		"$magenta": "#c792ea",
		"$red":     "#ff5f87", // Neon pink-red
		"$aqua":    "#5ee9f0", // Electric cyan (same as $fg0)
		"$purple":  "#c792ea", // Alias for magenta
		"$gray":    "#6c7384", // Alias for fg3
	})

	t.SetStyles(
		NewStyle("").WithColors("$fg0", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("box").WithBorder("thin"),
		NewStyle("button").WithColors("$bg0", "$blue").WithBorder("lines").WithPadding(0, 2),
		NewStyle("button:focused").WithColors("$bg0", "$cyan"),
		NewStyle("button:hovered").WithColors("$bg0", "$aqua"),
		NewStyle("button.dialog").WithBorder("none"),
		NewStyle("button.dialog:focused").WithBorder("none"),
		NewStyle("button.dialog:hovered").WithBorder("none"),
		NewStyle("checkbox").WithColors("$fg1", "$bg0"),
		NewStyle("checkbox:disabled").WithColors("$bg3", "$bg0"),
		NewStyle("checkbox:focused").WithColors("$blue", "$bg0"),
		NewStyle("checkbox:hovered").WithColors("$cyan", "$bg0"),
		NewStyle("collapsible"),
		NewStyle("collapsible/header").WithColors("$fg0", "$bg1"),
		NewStyle("collapsible/header:focused").WithColors("$fg0", "$blue"),
		NewStyle("collapsible/header:hovered").WithColors("$fg0", "$cyan"),
		NewStyle("custom"),
		NewStyle("dialog").WithColors("$fg1", "$bg2").WithBorder("round").WithPadding(1, 2),
		NewStyle("dialog/title").WithColors("$bg0", "$blue").WithBorder("none").WithMargin(0).WithPadding(0, 1),
		NewStyle("flex"),
		NewStyle("form"),
		NewStyle("formgroup").WithColors("$fg1", "$bg1"),
		NewStyle("formgroup:title").WithColors("$blue", "$bg1"),
		NewStyle("formgroup:label").WithColors("$fg3", "$bg1"),
		NewStyle("grid").WithBorder("thin"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg1"),
		NewStyle("list/highlight:focused").WithColors("$bg0", "$cyan"),
		NewStyle("editor/current-line").WithColors("$fg1", "$bg1"),
		NewStyle("editor/current-line-number").WithColors("$cyan", "$bg1"),
		NewStyle("editor/line-numbers").WithColors("$bg3", "$bg0"),
		NewStyle("editor/selection").WithColors("$bg0", "$blue"),
		NewStyle("input").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("input:focused").WithColors("$bg0", "$blue"),
		NewStyle("typeahead").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("typeahead:focused").WithColors("$bg0", "$blue"),
		NewStyle("typeahead/suggestion").WithColors("$fg2", "$bg2"),
		NewStyle("typeahead/suggestion:focused").WithColors("$fg2", "$blue"),
		NewStyle("rule"),
		NewStyle("scanner").WithColors("$cyan", "$bg0"),
		NewStyle("static").WithColors("$fg1", "$bg1").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg2").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$blue"),
		NewStyle("styled").WithColors("$fg1", "$bg1"),
		NewStyle("switcher"),
		NewStyle("table").WithColors("", "").WithBorder("thin $fg3"),
		NewStyle("table:focused").WithBorder("double $fg1"),
		NewStyle("table/grid").WithColors("$fg3", "$bg0").WithBorder("thin"),
		NewStyle("table/grid:focused").WithBorder("double-thin"),
		NewStyle("table/header").WithColors("$fg0", "$bg0"),
		NewStyle("table/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("table/highlight:focused").WithColors("$bg0", "$cyan"),
		NewStyle("tabs/highlight").WithColors("$bg0", "$orange"),
		NewStyle("tabs/highlight-line").WithForeground("$orange"),
		NewStyle("tabs/line:focused").WithForeground("$blue"),
		NewStyle("tabs/highlight:focused").WithColors("$bg0", "$cyan"),
		NewStyle("tabs/highlight-line:focused").WithForeground("$cyan"),
		NewStyle("text"),
		NewStyle("tree").WithColors("$fg1", "$bg0"),
		NewStyle("tree/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("tree/highlight:focused").WithColors("$bg0", "$cyan"),
		NewStyle("tree/indent").WithColors("$bg3", ""),
		NewStyle("viewport"),
	)

	t.SetStrings(map[string]string{
		// ---- Collapsible ----
		"collapsible.expanded":  "▼ ",
		"collapsible.collapsed": "▶ ",

		// ---- Progress bar ----
		"progress.h.prefix":        "",
		"progress.h.suffix":        "",
		"progress.h.start.filled":  "█",
		"progress.h.start.empty":   "░",
		"progress.h.middle.filled": "█",
		"progress.h.middle.empty":  "░",
		"progress.h.end.filled":    "█",
		"progress.h.end.empty":     "░",
		"progress.v.prefix":        "",
		"progress.v.suffix":        "",
		"progress.v.start.filled":  "█",
		"progress.v.start.empty":   "░",
		"progress.v.middle.filled": "█",
		"progress.v.middle.empty":  "░",
		"progress.v.end.filled":    "█",
		"progress.v.end.empty":     "░",

		// ---- Select ----
		"select.dropdown": " ▼",

		// ---- Tree ----
		"tree.expanded":  "▼ ",
		"tree.collapsed": "▶ ",
		"tree.branch":    "├─",
		"tree.last":      "└─",
		"tree.trunk":     "│ ",
	})

	return t
}

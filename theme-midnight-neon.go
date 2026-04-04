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
		"$fg0": "#e0e4f0", // Near-white — comfortable default text
		"$fg1": "#c7ccd9", // Light grey — secondary text
		"$fg2": "#8892a4", // Medium grey — muted text
		"$fg3": "#5a6478", // Dark grey — very muted / decorative

		// Neon accent colours
		"$blue":    "#4d9ef7",
		"$cyan":    "#00d9ff", // True neon cyan — focused / interactive states
		"$green":   "#4cd964",
		"$yellow":  "#ffd866",
		"$orange":  "#ff9f43",
		"$magenta": "#c792ea",
		"$red":     "#ff5f87", // Neon pink-red
		"$aqua":    "#5ee9f0", // Electric blue-cyan — accent / headers
		"$purple":  "#b07ef7", // Distinct purple
		"$gray":    "#5a6478", // Alias for $fg3
	})

	t.SetStyles(
		NewStyle("").WithColors("$fg0", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("box").WithBorder("thin"),
		NewStyle("button").WithColors("$fg0", "$blue").WithBorder("none").WithPadding(0, 2),
		NewStyle("button:focused").WithColors("$bg0", "$cyan"),
		NewStyle("button:hovered").WithColors("$bg0", "$aqua"),
		NewStyle("button.dialog").WithColors("$fg1", "$bg2").WithBorder("none"),
		NewStyle("button.dialog:focused").WithColors("$bg0", "$cyan").WithBorder("none"),
		NewStyle("button.dialog:hovered").WithColors("$bg0", "$aqua").WithBorder("none"),
		NewStyle("checkbox").WithColors("$fg1", "$bg0"),
		NewStyle("checkbox:disabled").WithColors("$bg3", "$bg0"),
		NewStyle("checkbox:focused").WithColors("$cyan", "$bg0"),
		NewStyle("checkbox:hovered").WithColors("$aqua", "$bg0"),
		NewStyle("collapsible"),
		NewStyle("collapsible/header").WithColors("$fg1", "$bg1"),
		NewStyle("collapsible/header:focused").WithColors("$fg0", "$blue"),
		NewStyle("collapsible/header:hovered").WithColors("$fg0", "$aqua"),
		NewStyle("custom"),
		NewStyle("dialog").WithColors("$fg1", "$bg2").WithBorder("round").WithPadding(1, 2),
		NewStyle("dialog/title").WithColors("$bg0", "$blue").WithBorder("none").WithMargin(0).WithPadding(0, 1),
		NewStyle("flex"),
		NewStyle("flex.dialog").WithColors("$fg1", "$bg2"),
		NewStyle("form"),
		NewStyle("formgroup").WithColors("$fg1", "$bg1"),
		NewStyle("formgroup:title").WithColors("$aqua", "$bg1"),
		NewStyle("formgroup:label").WithColors("$fg3", "$bg1"),
		NewStyle("grid").WithBorder("thin"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("list/highlight:focused").WithColors("$bg0", "$cyan"),
		NewStyle("editor/current-line").WithColors("$fg0", "$bg1"),
		NewStyle("editor/current-line-number").WithColors("$cyan", "$bg1"),
		NewStyle("editor/line-numbers").WithColors("$fg3", "$bg0"),
		NewStyle("editor/selection").WithColors("$bg0", "$blue"),
		NewStyle("input").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("input:focused").WithColors("$bg0", "$blue"),
		NewStyle("typeahead").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("typeahead:focused").WithColors("$bg0", "$blue"),
		NewStyle("typeahead/suggestion").WithColors("$fg2", "$bg2"),
		NewStyle("typeahead/suggestion:focused").WithColors("$fg2", "$blue"),
		NewStyle("rule"),
		NewStyle("scanner").WithColors("$aqua", "$bg0"),
		NewStyle("static").WithColors("$fg1", "").WithMargin(0).WithPadding(0),
		NewStyle("static.dialog").WithColors("$fg1", "$bg2").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg2").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$cyan"),
		NewStyle("styled").WithColors("$fg1", "$bg1"),
		NewStyle("switcher"),
		NewStyle("table").WithColors("", "").WithBorder("thin $fg3"),
		NewStyle("table:focused").WithBorder("double $aqua"),
		NewStyle("table/grid").WithColors("$fg3", "$bg0").WithBorder("thin"),
		NewStyle("table/grid:focused").WithBorder("double-thin"),
		NewStyle("table/header").WithColors("$aqua", "$bg0").WithFont("bold"),
		NewStyle("table/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("table/highlight:focused").WithColors("$bg0", "$cyan"),
		NewStyle("table/cell").WithColors("$fg0", "$bg3"),
		NewStyle("table/cell:focused").WithColors("$bg0", "$aqua").WithFont("bold"),
		NewStyle("tabs/highlight").WithColors("$bg0", "$orange"),
		NewStyle("tabs/highlight-line").WithForeground("$orange"),
		NewStyle("tabs/line:focused").WithForeground("$aqua"),
		NewStyle("tabs/highlight:focused").WithColors("$bg0", "$cyan"),
		NewStyle("tabs/highlight-line:focused").WithForeground("$cyan"),
		NewStyle("text"),
		NewStyle("tree").WithColors("$fg1", "$bg0"),
		NewStyle("tree/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("tree/highlight:focused").WithColors("$bg0", "$cyan"),
		NewStyle("tree/indent").WithColors("$fg3", ""),
		NewStyle("viewport"),
		NewStyle("terminal").WithColors("$fg0", "$bg0"),
		NewStyle("terminal:focused").WithColors("$fg0", "$bg0"),
		NewStyle("shortcuts").WithColors("$fg2", "$bg0"),
		NewStyle("shortcuts/key").WithForeground("$cyan").WithFont("bold"),
		NewStyle("shortcuts/label").WithForeground("$fg1"),
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

		// ---- Shortcuts ----
		"shortcuts.prefix":    "",
		"shortcuts.separator": "   ",
		"shortcuts.suffix":    "",

		// ---- Tree ----
		"tree.expanded":  "▼ ",
		"tree.collapsed": "▶ ",
		"tree.branch":    "├─",
		"tree.last":      "└─",
		"tree.trunk":     "│ ",
	})

	return t
}

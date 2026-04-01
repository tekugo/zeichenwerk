package zeichenwerk

// TokyoNightTheme returns a Theme styled after the Tokyo Night colour palette.
// It registers Unicode borders via AddUnicodeBorders and sets colours, styles,
// and string indicators for all built-in widgets.
func TokyoNightTheme() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)

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

	t.SetStyles(
		NewStyle("").WithColors("$fg0", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("box").WithBorder("thin"),
		NewStyle("button").WithColors("$bg0", "$blue").WithBorder("none").WithPadding(0, 2),
		NewStyle("button:focused").WithColors("$fg0", "$blue"),
		NewStyle("button:hovered").WithColors("$red", "$blue"),
		NewStyle("button.dialog").WithBorder("none"),
		NewStyle("button.dialog:focused").WithBorder("none"),
		NewStyle("button.dialog:hovered").WithBorder("none"),
		NewStyle("checkbox").WithColors("$fg2", "$bg0"),
		NewStyle("checkbox:disabled").WithColors("$gray", "$bg0"),
		NewStyle("checkbox:focused").WithColors("$fg0", "$bg0"),
		NewStyle("checkbox:hovered").WithColors("$aqua", "$bg0"),
		NewStyle("dialog").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(1, 2),
		NewStyle("dialog/title").WithColors("$bg1", "$blue").WithBorder("none").WithMargin(0).WithPadding(0, 1),
		NewStyle("collapsible"),
		NewStyle("collapsible/header").WithColors("$fg0", "$bg1"),
		NewStyle("collapsible/header:focused").WithColors("$fg0", "$blue"),
		NewStyle("collapsible/header:hovered").WithColors("$fg0", "$cyan"),
		NewStyle("custom"),
		NewStyle("flex"),
		NewStyle("form"),
		NewStyle("formgroup").WithColors("$fg0", "$bg1"),
		NewStyle("formgroup:title").WithColors("$blue", "$bg1"),
		NewStyle("formgroup:label").WithColors("$fg2", "$bg1"),
		NewStyle("grid").WithBorder("thin"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("list/highlight:focused").WithColors("$bg0", "$blue"),
		NewStyle("editor/current-line").WithColors("$fg0", "$bg1"),
		NewStyle("editor/current-line-number").WithColors("$blue", "$bg1"),
		NewStyle("editor/line-numbers").WithColors("$gray", "$bg0"),
		NewStyle("editor/selection").WithColors("$bg0", "$blue"),
		NewStyle("input").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("input:focused").WithColors("$bg0", "$blue"),
		NewStyle("typeahead").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("typeahead:focused").WithColors("$bg0", "$blue"),
		NewStyle("typeahead/suggestion").WithColors("$fg1", "$bg2"),
		NewStyle("typeahead/suggestion:focused").WithColors("$fg1", "$blue"),
		NewStyle("rule"),
		NewStyle("scanner").WithColors("$cyan", "$bg0"),
		NewStyle("static").WithColors("$fg0", "$bg1").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg2").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$blue"),
		NewStyle("styled").WithColors("$fg0", "$bg1"),
		NewStyle("switcher"),
		NewStyle("table").WithColors("", "").WithBorder("thin $fg1"),
		NewStyle("table:focused").WithBorder("double $fg0"),
		NewStyle("table/grid").WithColors("$fg1", "$bg0").WithBorder("thin"),
		NewStyle("table/grid:focused").WithBorder("double-thin"),
		NewStyle("table/header").WithColors("$fg0", "$bg0"),
		NewStyle("table/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("table/highlight:focused").WithColors("$bg0", "$blue").WithFont("bold"),
		NewStyle("tabs/highlight").WithColors("$bg0", "$fg0"),
		NewStyle("tabs/highlight-line").WithForeground("$bg3"),
		NewStyle("tabs/line:focused").WithForeground("$blue"),
		NewStyle("tabs/highlight:focused").WithColors("$bg0", "$orange"),
		NewStyle("tabs/highlight-line:focused").WithForeground("$orange"),
		NewStyle("text"),
		NewStyle("tree").WithColors("$fg0", "$bg0"),
		NewStyle("tree/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("tree/highlight:focused").WithColors("$bg0", "$blue"),
		NewStyle("tree/indent").WithColors("$gray", ""),
		NewStyle("viewport"),
	)

	t.SetStrings(map[string]string{
		// ---- Collapsible ----
		"collapsible.expanded":  "▼ ",
		"collapsible.collapsed": "▶ ",

		// ---- Progress bar ----
		// Horizontal orientation
		"progress.h.prefix":        "",
		"progress.h.suffix":        "",
		"progress.h.start.filled":  "\uEE03",
		"progress.h.start.empty":   "\uEE00",
		"progress.h.middle.filled": "\uEE04",
		"progress.h.middle.empty":  "\uEE01",
		"progress.h.end.filled":    "\uEE05",
		"progress.h.end.empty":     "\uEE02",
		// Vertical orientation
		"progress.v.prefix":        "",
		"progress.v.suffix":        "",
		"progress.v.start.filled":  "█",
		"progress.v.start.empty":   "░",
		"progress.v.middle.filled": "█",
		"progress.v.middle.empty":  "░",
		"progress.v.end.filled":    "█",
		"progress.v.end.empty":     "░",

		// ---- Select ----
		"select.dropdown": " \u25BC",

		// ---- Tree ----
		"tree.expanded":  " ▼",
		"tree.collapsed": " ▶",
		"tree.branch":    " ├─",
		"tree.last":      " └─",
		"tree.trunk":     " │ ",
	})

	return t
}

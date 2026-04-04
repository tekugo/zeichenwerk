package zeichenwerk

// NordTheme returns a Theme styled after the Nord colour palette.
// Nord is an arctic, north-bluish colour scheme with a clean, minimal design.
// The theme features cool blues, whites, and subtle accent colours that provide
// excellent readability and a calming, professional appearance.
//
// Colour groups:
//   - Polar Night ($bg0–$bg3): dark blue-grey backgrounds
//   - Snow Storm ($fg0–$fg2): near-white foreground tones
//   - Frost ($frost1–$frost4, $blue, $cyan, $aqua): blue accent range
//   - Aurora ($red, $orange, $yellow, $green, $purple): colourful highlights
func NordTheme() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)

	t.SetColors(map[string]string{
		// Polar Night — dark backgrounds
		"$bg0": "#2e3440", // Nord0 — darkest background
		"$bg1": "#3b4252", // Nord1 — dark background
		"$bg2": "#434c5e", // Nord2 — medium background
		"$bg3": "#4c566a", // Nord3 — light background

		// Snow Storm — light foregrounds
		"$fg0": "#eceff4", // Nord6 — lightest foreground
		"$fg1": "#e5e9f0", // Nord5 — light foreground
		"$fg2": "#d8dee9", // Nord4 — medium foreground

		// Frost — blue accent range
		"$frost1": "#8fbcbb", // Nord7 — light teal
		"$frost2": "#88c0d0", // Nord8 — medium blue
		"$frost3": "#81a1c1", // Nord9 — steel blue
		"$frost4": "#5e81ac", // Nord10 — deep blue

		// Aurora — colourful highlights
		"$red":    "#bf616a", // Nord11
		"$orange": "#d08770", // Nord12
		"$yellow": "#ebcb8b", // Nord13
		"$green":  "#a3be8c", // Nord14
		"$purple": "#b48ead", // Nord15

		// Aliases for widget style compatibility
		"$blue":    "#81a1c1", // → frost3
		"$cyan":    "#88c0d0", // → frost2
		"$aqua":    "#8fbcbb", // → frost1
		"$magenta": "#b48ead", // → purple
		"$gray":    "#4c566a", // → bg3
	})

	t.SetStyles(
		NewStyle("").WithColors("$fg0", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("box").WithBorder("thin"),
		NewStyle("button").WithColors("$bg0", "$frost3").WithBorder("none").WithPadding(0, 2),
		NewStyle("button:focused").WithColors("$bg0", "$frost2"),
		NewStyle("button:hovered").WithColors("$bg0", "$frost1"),
		NewStyle("button.dialog").WithColors("$fg0", "$bg2").WithBorder("none"),
		NewStyle("button.dialog:focused").WithColors("$bg0", "$frost2").WithBorder("none"),
		NewStyle("button.dialog:hovered").WithColors("$bg0", "$frost1").WithBorder("none"),
		NewStyle("checkbox").WithColors("$fg1", "$bg0"),
		NewStyle("checkbox:disabled").WithColors("$bg3", "$bg0"),
		NewStyle("checkbox:focused").WithColors("$frost2", "$bg0"),
		NewStyle("checkbox:hovered").WithColors("$frost1", "$bg0"),
		NewStyle("collapsible"),
		NewStyle("collapsible/header").WithColors("$fg0", "$bg1"),
		NewStyle("collapsible/header:focused").WithColors("$fg0", "$frost3"),
		NewStyle("collapsible/header:hovered").WithColors("$fg0", "$frost2"),
		NewStyle("custom"),
		NewStyle("dialog").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(1, 2),
		NewStyle("dialog/title").WithColors("$bg0", "$frost2").WithBorder("none").WithMargin(0).WithPadding(0, 1),
		NewStyle("flex"),
		NewStyle("flex.dialog").WithColors("$fg0", "$bg2"),
		NewStyle("form"),
		NewStyle("formgroup").WithColors("$fg0", "$bg1"),
		NewStyle("formgroup:title").WithColors("$frost2", "$bg1"),
		NewStyle("formgroup:label").WithColors("$bg3", "$bg1"),
		NewStyle("grid").WithBorder("thin"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("list/highlight:focused").WithColors("$bg0", "$frost2"),
		NewStyle("editor/current-line").WithColors("$fg0", "$bg1"),
		NewStyle("editor/current-line-number").WithColors("$frost2", "$bg1"),
		NewStyle("editor/line-numbers").WithColors("$bg3", "$bg0"),
		NewStyle("editor/selection").WithColors("$bg0", "$frost3"),
		NewStyle("input").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("input:focused").WithColors("$bg0", "$frost2"),
		NewStyle("typeahead").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("typeahead:focused").WithColors("$bg0", "$frost2"),
		NewStyle("typeahead/suggestion").WithColors("$fg2", "$bg2"),
		NewStyle("typeahead/suggestion:focused").WithColors("$fg2", "$frost2"),
		NewStyle("rule"),
		NewStyle("scanner").WithColors("$frost1", "$bg0"),
		NewStyle("static").WithColors("$fg0", "$bg1").WithMargin(0).WithPadding(0),
		NewStyle("static.dialog").WithColors("$fg0", "$bg2").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg2").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$frost2"),
		NewStyle("styled").WithColors("$fg0", "$bg1"),
		NewStyle("switcher"),
		NewStyle("table").WithColors("", "").WithBorder("thin $fg2"),
		NewStyle("table:focused").WithBorder("double $fg0"),
		NewStyle("table/grid").WithColors("$fg2", "$bg0").WithBorder("thin"),
		NewStyle("table/grid:focused").WithBorder("double-thin"),
		NewStyle("table/header").WithColors("$fg0", "$bg0"),
		NewStyle("table/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("table/highlight:focused").WithColors("$bg0", "$frost2"),
		NewStyle("table/cell").WithColors("$fg0", "$bg3"),
		NewStyle("table/cell:focused").WithColors("$bg0", "$frost1").WithFont("bold"),
		NewStyle("tabs/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("tabs/highlight-line").WithForeground("$fg2"),
		NewStyle("tabs/line:focused").WithForeground("$frost2"),
		NewStyle("tabs/highlight:focused").WithColors("$bg0", "$frost2"),
		NewStyle("tabs/highlight-line:focused").WithForeground("$frost2"),
		NewStyle("text"),
		NewStyle("tree").WithColors("$fg0", "$bg0"),
		NewStyle("tree/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("tree/highlight:focused").WithColors("$bg0", "$frost2"),
		NewStyle("tree/indent").WithColors("$bg3", ""),
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

package zeichenwerk

// GruvboxDarkTheme returns a Theme styled after the Gruvbox Dark colour palette.
// Gruvbox is a retro groove colour scheme with warm, earthy tones designed for
// comfortable long-term use. The dark variant features warm dark backgrounds
// with bright, saturated foreground colours that provide excellent contrast.
func GruvboxDarkTheme() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)

	t.SetColors(map[string]string{
		// Dark backgrounds
		"$bg0": "#282828", // Hard dark
		"$bg1": "#3c3836", // Medium dark
		"$bg2": "#504945", // Soft dark
		"$bg3": "#665c54", // Light dark
		"$bg4": "#7c6f64", // Lightest dark

		// Light foregrounds
		"$fg0": "#fbf1c7", // Hard light
		"$fg1": "#ebdbb2", // Medium light
		"$fg2": "#d5c4a1", // Soft light
		"$fg3": "#bdae93", // Dark light
		"$fg4": "#a89984", // Darkest light

		// Neutral and accent colours
		"$gray":   "#928374", // Gray
		"$red":    "#fb4934", // Bright red
		"$green":  "#b8bb26", // Bright green
		"$yellow": "#fabd2f", // Bright yellow
		"$blue":   "#83a598", // Bright blue
		"$purple": "#d3869b", // Bright purple
		"$aqua":   "#8ec07c", // Bright aqua
		"$orange": "#fe8019", // Bright orange

		// Faded variants for subtle elements
		"$red_dim":    "#cc241d", // Faded red
		"$green_dim":  "#98971a", // Faded green
		"$yellow_dim": "#d79921", // Faded yellow
		"$blue_dim":   "#458588", // Faded blue
		"$purple_dim": "#b16286", // Faded purple
		"$aqua_dim":   "#689d6a", // Faded aqua
		"$orange_dim": "#d65d0e", // Faded orange

		// Aliases
		"$cyan":    "#8ec07c", // Alias for aqua
		"$magenta": "#d3869b", // Alias for purple
	})

	t.SetStyles(
		NewStyle("").WithColors("$fg1", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("box").WithBorder("thin"),
		NewStyle("button").WithColors("$bg0", "$yellow").WithBorder("none").WithPadding(0, 2),
		NewStyle("button:focused").WithColors("$bg0", "$orange"),
		NewStyle("button:hovered").WithColors("$bg0", "$yellow_dim"),
		NewStyle("button.dialog").WithColors("$fg0", "$bg2").WithBorder("none"),
		NewStyle("button.dialog:focused").WithColors("$bg0", "$orange").WithBorder("none"),
		NewStyle("button.dialog:hovered").WithColors("$bg0", "$yellow_dim").WithBorder("none"),
		NewStyle("checkbox").WithColors("$fg1", "$bg0"),
		NewStyle("checkbox:disabled").WithColors("$bg3", "$bg0"),
		NewStyle("checkbox:focused").WithColors("$yellow", "$bg0"),
		NewStyle("checkbox:hovered").WithColors("$orange", "$bg0"),
		NewStyle("collapsible"),
		NewStyle("collapsible/header").WithColors("$fg0", "$bg1"),
		NewStyle("collapsible/header:focused").WithColors("$fg0", "$yellow"),
		NewStyle("collapsible/header:hovered").WithColors("$fg0", "$aqua"),
		NewStyle("custom"),
		NewStyle("dialog").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(1, 2),
		NewStyle("dialog/title").WithColors("$bg1", "$yellow").WithBorder("none").WithMargin(0).WithPadding(0, 1),
		NewStyle("flex"),
		NewStyle("flex.dialog").WithColors("$fg0", "$bg2"),
		NewStyle("form"),
		NewStyle("formgroup").WithColors("$fg0", "$bg1"),
		NewStyle("formgroup:title").WithColors("$yellow", "$bg1"),
		NewStyle("formgroup:label").WithColors("$fg4", "$bg1"),
		NewStyle("grid").WithBorder("thin"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg4"),
		NewStyle("list/highlight:focused").WithColors("$bg0", "$yellow"),
		NewStyle("editor/current-line").WithColors("$fg0", "$bg1"),
		NewStyle("editor/current-line-number").WithColors("$yellow", "$bg1"),
		NewStyle("editor/line-numbers").WithColors("$gray", "$bg0"),
		NewStyle("editor/selection").WithColors("$bg0", "$blue"),
		NewStyle("input").WithColors("$fg0", "$bg1").WithCursor("*bar"),
		NewStyle("input:focused").WithColors("$bg0", "$yellow"),
		NewStyle("typeahead").WithColors("$fg0", "$bg1").WithCursor("*bar"),
		NewStyle("typeahead:focused").WithColors("$bg0", "$yellow"),
		NewStyle("typeahead/suggestion").WithColors("$fg2", "$bg1"),
		NewStyle("typeahead/suggestion:focused").WithColors("$fg2", "$yellow"),
		NewStyle("rule"),
		NewStyle("scanner").WithColors("$aqua", "$bg0"),
		NewStyle("sparkline").WithColors("$aqua", "$bg0"),
		NewStyle("sparkline/high").WithColors("$orange", "$bg0"),
		NewStyle("heatmap").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/header").WithColors("$fg2", "$bg0"),
		NewStyle("heatmap/zero").WithColors("$fg4", "$bg2"),
		NewStyle("heatmap/mid").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/max").WithColors("$bg0", "$green"),
		NewStyle("static").WithColors("$fg0", "$bg1").WithMargin(0).WithPadding(0),
		NewStyle("static.dialog").WithColors("$fg0", "$bg2").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg1").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$yellow"),
		NewStyle("styled").WithColors("$fg0", "$bg1"),
		NewStyle("switcher"),
		NewStyle("table").WithColors("", "").WithBorder("thin $fg3"),
		NewStyle("table:focused").WithBorder("double $fg0"),
		NewStyle("table/grid").WithColors("$fg3", "$bg0").WithBorder("thin"),
		NewStyle("table/grid:focused").WithBorder("double-thin"),
		NewStyle("table/header").WithColors("$fg0", "$bg0"),
		NewStyle("table/highlight").WithColors("$bg0", "$fg4"),
		NewStyle("table/highlight:focused").WithColors("$bg0", "$yellow"),
		NewStyle("table/cell").WithColors("$fg0", "$bg3"),
		NewStyle("table/cell:focused").WithColors("$bg0", "$orange").WithFont("bold"),
		NewStyle("tabs/highlight").WithColors("$bg0", "$fg3"),
		NewStyle("tabs/highlight-line").WithForeground("$fg3"),
		NewStyle("tabs/line:focused").WithForeground("$yellow"),
		NewStyle("tabs/highlight:focused").WithColors("$bg0", "$yellow"),
		NewStyle("tabs/highlight-line:focused").WithForeground("$yellow"),
		NewStyle("text"),
		NewStyle("tree").WithColors("$fg1", "$bg0"),
		NewStyle("tree/highlight").WithColors("$bg0", "$fg4"),
		NewStyle("tree/highlight:focused").WithColors("$bg0", "$yellow"),
		NewStyle("tree/indent").WithColors("$gray", ""),
		NewStyle("viewport"),
		NewStyle("terminal").WithColors("$fg0", "$bg0"),
		NewStyle("terminal:focused").WithColors("$fg0", "$bg0"),
		NewStyle("shortcuts").WithColors("$fg2", "$bg0"),
		NewStyle("shortcuts/key").WithForeground("$yellow").WithFont("bold"),
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

// GruvboxLightTheme returns a Theme styled after the Gruvbox Light colour palette.
// This is the light variant of Gruvbox, featuring warm cream-coloured backgrounds
// with darker foreground colours while maintaining the same retro aesthetic.
func GruvboxLightTheme() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)

	t.SetColors(map[string]string{
		// Light backgrounds (inverted from dark)
		"$bg0": "#fbf1c7", // Hard light
		"$bg1": "#ebdbb2", // Medium light
		"$bg2": "#d5c4a1", // Soft light
		"$bg3": "#bdae93", // Dark light
		"$bg4": "#a89984", // Darkest light

		// Dark foregrounds (inverted from dark)
		"$fg0": "#282828", // Hard dark
		"$fg1": "#3c3836", // Medium dark
		"$fg2": "#504945", // Soft dark
		"$fg3": "#665c54", // Light dark
		"$fg4": "#7c6f64", // Lightest dark

		// Neutral and accent colours (darker variants for light theme)
		"$gray":   "#928374", // Gray
		"$red":    "#cc241d", // Dark red
		"$green":  "#98971a", // Dark green
		"$yellow": "#d79921", // Dark yellow
		"$blue":   "#458588", // Dark blue
		"$purple": "#b16286", // Dark purple
		"$aqua":   "#689d6a", // Dark aqua
		"$orange": "#d65d0e", // Dark orange

		// Bright variants for highlights
		"$red_bright":    "#fb4934", // Bright red
		"$green_bright":  "#b8bb26", // Bright green
		"$yellow_bright": "#fabd2f", // Bright yellow
		"$blue_bright":   "#83a598", // Bright blue
		"$purple_bright": "#d3869b", // Bright purple
		"$aqua_bright":   "#8ec07c", // Bright aqua
		"$orange_bright": "#fe8019", // Bright orange

		// Aliases
		"$cyan":    "#689d6a", // Alias for aqua
		"$magenta": "#b16286", // Alias for purple
	})

	t.SetStyles(
		NewStyle("").WithColors("$fg1", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("box").WithBorder("thin"),
		NewStyle("button").WithColors("$bg0", "$orange").WithBorder("none").WithPadding(0, 2),
		NewStyle("button:focused").WithColors("$bg0", "$red"),
		NewStyle("button:hovered").WithColors("$bg0", "$orange_bright"),
		NewStyle("button.dialog").WithColors("$fg0", "$bg2").WithBorder("none"),
		NewStyle("button.dialog:focused").WithColors("$bg0", "$red").WithBorder("none"),
		NewStyle("button.dialog:hovered").WithColors("$bg0", "$orange_bright").WithBorder("none"),
		NewStyle("checkbox").WithColors("$fg1", "$bg0"),
		NewStyle("checkbox:disabled").WithColors("$bg3", "$bg0"),
		NewStyle("checkbox:focused").WithColors("$orange", "$bg0"),
		NewStyle("checkbox:hovered").WithColors("$red", "$bg0"),
		NewStyle("collapsible"),
		NewStyle("collapsible/header").WithColors("$fg0", "$bg1"),
		NewStyle("collapsible/header:focused").WithColors("$fg0", "$orange"),
		NewStyle("collapsible/header:hovered").WithColors("$fg0", "$aqua"),
		NewStyle("custom"),
		NewStyle("dialog").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(1, 2),
		NewStyle("dialog/title").WithColors("$bg1", "$orange").WithBorder("none").WithMargin(0).WithPadding(0, 1),
		NewStyle("flex"),
		NewStyle("flex.dialog").WithColors("$fg0", "$bg2"),
		NewStyle("form"),
		NewStyle("formgroup").WithColors("$fg0", "$bg1"),
		NewStyle("formgroup:title").WithColors("$orange", "$bg1"),
		NewStyle("formgroup:label").WithColors("$fg4", "$bg1"),
		NewStyle("grid").WithBorder("thin"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg4"),
		NewStyle("list/highlight:focused").WithColors("$bg0", "$orange"),
		NewStyle("editor/current-line").WithColors("$fg0", "$bg1"),
		NewStyle("editor/current-line-number").WithColors("$orange", "$bg1"),
		NewStyle("editor/line-numbers").WithColors("$gray", "$bg0"),
		NewStyle("editor/selection").WithColors("$bg0", "$blue"),
		NewStyle("input").WithColors("$fg0", "$bg1").WithCursor("*bar"),
		NewStyle("input:focused").WithColors("$bg0", "$orange"),
		NewStyle("typeahead").WithColors("$fg0", "$bg1").WithCursor("*bar"),
		NewStyle("typeahead:focused").WithColors("$bg0", "$orange"),
		NewStyle("typeahead/suggestion").WithColors("$fg2", "$bg1"),
		NewStyle("typeahead/suggestion:focused").WithColors("$fg2", "$orange"),
		NewStyle("rule"),
		NewStyle("scanner").WithColors("$aqua", "$bg0"),
		NewStyle("sparkline").WithColors("$aqua", "$bg0"),
		NewStyle("sparkline/high").WithColors("$orange", "$bg0"),
		NewStyle("heatmap").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/header").WithColors("$fg2", "$bg0"),
		NewStyle("heatmap/zero").WithColors("$fg4", "$bg2"),
		NewStyle("heatmap/mid").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/max").WithColors("$bg0", "$green"),
		NewStyle("static").WithColors("$fg0", "$bg1").WithMargin(0).WithPadding(0),
		NewStyle("static.dialog").WithColors("$fg0", "$bg2").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg1").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$orange"),
		NewStyle("styled").WithColors("$fg0", "$bg1"),
		NewStyle("switcher"),
		NewStyle("table").WithColors("", "").WithBorder("thin $fg3"),
		NewStyle("table:focused").WithBorder("double $fg0"),
		NewStyle("table/grid").WithColors("$fg3", "$bg0").WithBorder("thin"),
		NewStyle("table/grid:focused").WithBorder("double-thin"),
		NewStyle("table/header").WithColors("$fg0", "$bg0"),
		NewStyle("table/highlight").WithColors("$bg0", "$fg4"),
		NewStyle("table/highlight:focused").WithColors("$bg0", "$orange"),
		NewStyle("table/cell").WithColors("$fg0", "$bg3"),
		NewStyle("table/cell:focused").WithColors("$bg0", "$yellow").WithFont("bold"),
		NewStyle("tabs/highlight").WithColors("$bg0", "$fg3"),
		NewStyle("tabs/highlight-line").WithForeground("$fg3"),
		NewStyle("tabs/line:focused").WithForeground("$orange"),
		NewStyle("tabs/highlight:focused").WithColors("$bg0", "$orange"),
		NewStyle("tabs/highlight-line:focused").WithForeground("$orange"),
		NewStyle("text"),
		NewStyle("tree").WithColors("$fg1", "$bg0"),
		NewStyle("tree/highlight").WithColors("$bg0", "$fg4"),
		NewStyle("tree/highlight:focused").WithColors("$bg0", "$orange"),
		NewStyle("tree/indent").WithColors("$gray", ""),
		NewStyle("viewport"),
		NewStyle("terminal").WithColors("$fg0", "$bg0"),
		NewStyle("terminal:focused").WithColors("$fg0", "$bg0"),
		NewStyle("shortcuts").WithColors("$fg2", "$bg0"),
		NewStyle("shortcuts/key").WithForeground("$blue").WithFont("bold"),
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

package themes

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// GruvboxDarkTheme returns a Theme styled after the Gruvbox Dark colour palette.
// Gruvbox is a retro groove colour scheme with warm, earthy tones designed for
// comfortable long-term use. The dark variant features warm dark backgrounds
// with bright, saturated foreground colours that provide excellent contrast.
func GruvboxDark() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)
	AddUnicodeStrings(t)

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

	t.AddStyles(
		NewStyle("").WithColors("$fg1", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("bar-chart").WithColors("$fg1", "$bg0"),
		NewStyle("bar-chart/s0").WithColors("$blue", "$bg0"),
		NewStyle("bar-chart/s1").WithColors("$green", "$bg0"),
		NewStyle("bar-chart/s2").WithColors("$orange", "$bg0"),
		NewStyle("bar-chart/s3").WithColors("$purple", "$bg0"),
		NewStyle("bar-chart/s4").WithColors("$aqua", "$bg0"),
		NewStyle("bar-chart/s5").WithColors("$yellow", "$bg0"),
		NewStyle("bar-chart/s6").WithColors("$red", "$bg0"),
		NewStyle("bar-chart/s7").WithColors("$blue_dim", "$bg0"),
		NewStyle("bar-chart/axis").WithColors("$fg4", "$bg0"),
		NewStyle("bar-chart/grid").WithColors("$bg3", "$bg0"),
		NewStyle("bar-chart/label").WithColors("$fg2", "$bg0"),
		NewStyle("bar-chart/label:focused").WithColors("$yellow", "$bg0").WithFont("bold"),
		NewStyle("bar-chart/selection").WithColors("$bg0", "$fg4"),
		NewStyle("bar-chart/value").WithColors("$fg1", "$bg0"),
		NewStyle("bar-chart/legend").WithColors("$fg2", "$bg0"),
		NewStyle("button").WithColors("$bg0", "$yellow").WithBorder("none").WithPadding(0, 2),
		NewStyle("button:focused").WithColors("$bg0", "$orange"),
		NewStyle("button:hovered").WithColors("$bg0", "$yellow_dim"),
		NewStyle("button.dialog").WithColors("$fg0", "$bg2").WithBorder("none"),
		NewStyle("button.dialog:focused").WithColors("$bg0", "$orange").WithBorder("none"),
		NewStyle("button.dialog:hovered").WithColors("$bg0", "$yellow_dim").WithBorder("none"),
		NewStyle("checkbox:focused").WithColors("$yellow", "$bg0"),
		NewStyle("checkbox:hovered").WithColors("$orange", "$bg0"),
		NewStyle("collapsible/header").WithColors("$fg0", "$bg1"),
		NewStyle("collapsible/header:focused").WithColors("$fg0", "$yellow"),
		NewStyle("collapsible/header:hovered").WithColors("$fg0", "$aqua"),
		NewStyle("dialog").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(1, 2),
		NewStyle("dialog/title").WithColors("$bg1", "$yellow").WithBorder("none").WithMargin(0).WithPadding(0, 1),
		NewStyle("editor/current-line-number").WithColors("$yellow", "$bg1"),
		NewStyle("editor/line-numbers").WithColors("$gray", "$bg0"),
		NewStyle("editor/selection").WithColors("$bg0", "$blue"),
		NewStyle("formgroup:title").WithColors("$yellow", "$bg1"),
		NewStyle("input:focused").WithColors("$bg0", "$yellow"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg4"),
		NewStyle("list/highlight:focused").WithColors("$bg0", "$yellow"),
		NewStyle("marquee").WithColors("$aqua", "$bg0"),
		NewStyle("typeahead:focused").WithColors("$bg0", "$yellow"),
		NewStyle("typeahead/suggestion:focused").WithColors("$fg2", "$yellow"),
		NewStyle("scanner").WithColors("$aqua", "$bg0"),
		NewStyle("shimmer/band").WithForeground("$yellow"),
		NewStyle("sparkline").WithColors("$aqua", "$bg0"),
		NewStyle("sparkline/high").WithColors("$orange", "$bg0"),
		NewStyle("typewriter").WithColors("$fg0", "$bg0"),
		NewStyle("typewriter/cursor").WithColors("$yellow", "$bg0"),
		NewStyle("heatmap").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/header").WithColors("$fg2", "$bg0"),
		NewStyle("heatmap/zero").WithColors("$fg4", "$bg2"),
		NewStyle("heatmap/mid").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/max").WithColors("$bg0", "$green"),
		NewStyle("breadcrumb").WithColors("$fg0", "$bg0"),
		NewStyle("breadcrumb/segment").WithColors("$fg1", "$bg0"),
		NewStyle("breadcrumb/segment:focused").WithColors("$bg0", "$yellow").WithFont("bold"),
		NewStyle("breadcrumb/separator").WithColors("$fg4", "$bg0"),
		NewStyle("indicator").WithColors("$fg1", ""),
		NewStyle("indicator:debug").WithForeground("$gray"),
		NewStyle("indicator:info").WithForeground("$blue"),
		NewStyle("indicator:success").WithForeground("$green"),
		NewStyle("indicator:warning").WithForeground("$yellow"),
		NewStyle("indicator:error").WithForeground("$red"),
		NewStyle("indicator:fatal").WithForeground("$magenta"),
		NewStyle("static").WithColors("$fg0", "$bg1").WithMargin(0).WithPadding(0),
		NewStyle("static.dialog").WithColors("$fg0", "$bg2").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg1").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$yellow"),
		NewStyle("combo").WithColors("$fg0", "$bg1").WithPadding(0, 1),
		NewStyle("combo:focused").WithColors("$bg0", "$yellow"),
		NewStyle("styled").WithColors("$fg0", "$bg1").WithPadding(0, 1),
		NewStyle("styled/h1").WithFont("bold"),
		NewStyle("styled/h2").WithFont("bold"),
		NewStyle("styled/h3").WithFont("bold underline"),
		NewStyle("styled/h4").WithFont("bold"),
		NewStyle("styled/pre").WithBackground("$bg0"),
		NewStyle("styled/code").WithForeground("$aqua"),
		NewStyle("styled/bq").WithColors("$fg2", "$bg1"),
		NewStyle("styled/hr").WithForeground("$fg3"),
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
		NewStyle("tiles").WithColors("$fg1", "$bg0"),
		NewStyle("tiles:focused").WithBorder("round $yellow"),
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
		NewStyle("commands").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(0, 0),
		NewStyle("commands/input").WithColors("$fg0", "$bg1").WithBorder("none").WithCursor("*bar"),
		NewStyle("commands/item").WithColors("$fg1", "$bg2"),
		NewStyle("commands/item:focused").WithColors("$bg0", "$yellow").WithFont("bold"),
		NewStyle("commands/shortcut").WithColors("$fg4", "$bg2"),
		NewStyle("commands/shortcut:focused").WithColors("$bg1", "$yellow"),
		NewStyle("commands/group").WithColors("$fg4", "$bg2").WithFont("bold"),
	)

	return t
}

package themes

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// MidnightNeonTheme returns a Theme styled after a dark midnight palette with
// neon accent colours. It features near-black backgrounds with vivid electric
// highlights — cyan, blue, green, and magenta — for a futuristic aesthetic.
func MidnightNeon() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)
	AddUnicodeStrings(t)

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

	t.AddStyles(
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
		NewStyle("marquee").WithColors("$cyan", "$bg0"),
		NewStyle("shimmer").WithColors("$fg2", "$bg0"),
		NewStyle("shimmer/band").WithForeground("$cyan"),
		NewStyle("scanner").WithColors("$aqua", "$bg0"),
		NewStyle("sparkline").WithColors("$aqua", "$bg0"),
		NewStyle("typewriter").WithColors("$fg0", "$bg0"),
		NewStyle("typewriter/cursor").WithColors("$cyan", "$bg0"),
		NewStyle("sparkline/high").WithColors("$orange", "$bg0"),
		NewStyle("bar-chart").WithColors("$fg0", "$bg0"),
		NewStyle("bar-chart/s0").WithColors("$cyan", "$bg0"),
		NewStyle("bar-chart/s1").WithColors("$green", "$bg0"),
		NewStyle("bar-chart/s2").WithColors("$magenta", "$bg0"),
		NewStyle("bar-chart/s3").WithColors("$blue", "$bg0"),
		NewStyle("bar-chart/s4").WithColors("$orange", "$bg0"),
		NewStyle("bar-chart/s5").WithColors("$red", "$bg0"),
		NewStyle("bar-chart/s6").WithColors("$aqua", "$bg0"),
		NewStyle("bar-chart/s7").WithColors("$yellow", "$bg0"),
		NewStyle("bar-chart/axis").WithColors("$fg3", "$bg0"),
		NewStyle("bar-chart/grid").WithColors("$bg3", "$bg0"),
		NewStyle("bar-chart/label").WithColors("$fg2", "$bg0"),
		NewStyle("bar-chart/label:focused").WithColors("$cyan", "$bg0").WithFont("bold"),
		NewStyle("bar-chart/selection").WithColors("$bg0", "$fg2"),
		NewStyle("bar-chart/value").WithColors("$fg0", "$bg0"),
		NewStyle("bar-chart/legend").WithColors("$fg2", "$bg0"),
		NewStyle("heatmap").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/header").WithColors("$fg2", "$bg0"),
		NewStyle("heatmap/zero").WithColors("$fg2", "$bg2"),
		NewStyle("heatmap/mid").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/max").WithColors("$bg0", "$green"),
		NewStyle("breadcrumb").WithColors("$fg0", "$bg0"),
		NewStyle("breadcrumb/segment").WithColors("$fg1", "$bg0"),
		NewStyle("breadcrumb/segment:focused").WithColors("$bg0", "$cyan").WithFont("bold"),
		NewStyle("breadcrumb/separator").WithColors("$fg3", "$bg0"),
		NewStyle("static").WithColors("$fg1", "").WithMargin(0).WithPadding(0),
		NewStyle("static.dialog").WithColors("$fg1", "$bg2").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg2").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$cyan"),
		NewStyle("combo").WithColors("$fg0", "$bg2").WithPadding(0, 1),
		NewStyle("combo:focused").WithColors("$bg0", "$cyan"),
		NewStyle("styled").WithColors("$fg1", "$bg1").WithPadding(0, 1),
		NewStyle("styled/h1").WithColors("$fg0", "$bg1").WithFont("bold"),
		NewStyle("styled/h2").WithColors("$fg0", "$bg1").WithFont("bold"),
		NewStyle("styled/h3").WithColors("$fg0", "$bg1").WithFont("bold underline"),
		NewStyle("styled/h4").WithColors("$fg0", "$bg1").WithFont("bold"),
		NewStyle("styled/pre").WithBackground("$bg0"),
		NewStyle("styled/code").WithForeground("$aqua"),
		NewStyle("styled/bq").WithColors("$fg2", "$bg1"),
		NewStyle("styled/hr").WithForeground("$fg3"),
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
		NewStyle("tiles").WithColors("$fg1", "$bg0"),
		NewStyle("tiles:focused").WithBorder("round $cyan"),
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
		NewStyle("commands").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(0, 0),
		NewStyle("commands/input").WithColors("$fg0", "$bg3").WithBorder("none").WithCursor("*bar"),
		NewStyle("commands/item").WithColors("$fg1", "$bg2"),
		NewStyle("commands/item:focused").WithColors("$bg0", "$cyan").WithFont("bold"),
		NewStyle("commands/shortcut").WithColors("$fg3", "$bg2"),
		NewStyle("commands/shortcut:focused").WithColors("$bg1", "$cyan"),
		NewStyle("commands/group").WithColors("$fg3", "$bg2").WithFont("bold"),
	)

	return t
}

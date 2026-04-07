package zeichenwerk

// LipstickTheme returns a Theme inspired by the Charm / Lipgloss aesthetic.
// It features warm dark backgrounds, Charm's signature fuchsia and indigo
// accents, and cream foreground text — giving a polished, modern feel that
// matches the look of Bubble Tea applications.
//
// Colour groups:
//   - Ink ($bg0–$bg3): warm near-black backgrounds with a subtle purple tint
//   - Cream ($fg0–$fg2): warm off-white foreground tones
//   - Accents: fuchsia, indigo, green, and supporting colours from the Charm palette
func LipstickTheme() *Theme {
	t := NewTheme()

	AddUnicodeBorders(t)

	t.SetColors(map[string]string{
		// Ink — warm dark backgrounds with a faint purple undertone
		"$bg0": "#1c1c1e", // Near-black
		"$bg1": "#2a2a2e", // Dark surface
		"$bg2": "#36363f", // Elevated surface
		"$bg3": "#46465a", // Highlight / selection background

		// Cream — warm off-white foreground tones
		"$fg0": "#fffdf5", // Charm cream — primary text
		"$fg1": "#d9d4cf", // Slightly dimmed
		"$fg2": "#9090aa", // Muted / secondary — purple-gray

		// Charm accent palette
		"$fuchsia": "#f780e2", // Charm signature fuchsia — primary accent
		"$pink":    "#f25d94", // Warm pink — hover / secondary accent
		"$indigo":  "#7571f9", // Charm indigo — focused states
		"$purple":  "#7d56f4", // Deeper purple — highlights
		"$green":   "#02bf87", // Charm emerald — positive / active
		"$cyan":    "#14f9d5", // Charm cyan — information
		"$yellow":  "#edff82", // Charm yellow — warnings
		"$red":     "#ff4672", // Charm red — errors / destructive
		"$gray":    "#8a8aaa", // Muted purple-gray — readable on dark backgrounds

		// Aliases for widget style compatibility
		"$blue":    "#7571f9", // → indigo
		"$aqua":    "#14f9d5", // → cyan
		"$magenta": "#f780e2", // → fuchsia
		"$orange":  "#f25d94", // → pink
	})

	t.SetStyles(
		NewStyle("").WithColors("$fg0", "$bg0").WithMargin(0).WithPadding(0),
		NewStyle("box").WithBorder("round"),
		NewStyle("button").WithColors("$bg0", "$fuchsia").WithBorder("none").WithPadding(0, 2),
		NewStyle("button:focused").WithColors("$bg0", "$indigo"),
		NewStyle("button:hovered").WithColors("$bg0", "$pink"),
		NewStyle("button.dialog").WithColors("$fg0", "$bg2").WithBorder("none"),
		NewStyle("button.dialog:focused").WithColors("$bg0", "$indigo").WithBorder("none"),
		NewStyle("button.dialog:hovered").WithColors("$bg0", "$pink").WithBorder("none"),
		NewStyle("checkbox").WithColors("$fg1", "$bg0"),
		NewStyle("checkbox:disabled").WithColors("$gray", "$bg0"),
		NewStyle("checkbox:focused").WithColors("$fuchsia", "$bg0"),
		NewStyle("checkbox:hovered").WithColors("$pink", "$bg0"),
		NewStyle("collapsible"),
		NewStyle("collapsible/header").WithColors("$fg0", "$bg1"),
		NewStyle("collapsible/header:focused").WithColors("$fg0", "$indigo"),
		NewStyle("collapsible/header:hovered").WithColors("$fg0", "$fuchsia"),
		NewStyle("custom"),
		NewStyle("dialog").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(1, 2),
		NewStyle("dialog/title").WithColors("$bg0", "$fuchsia").WithBorder("none").WithMargin(0).WithPadding(0, 1),
		NewStyle("flex"),
		NewStyle("flex.dialog").WithColors("$fg0", "$bg2"),
		NewStyle("form"),
		NewStyle("formgroup").WithColors("$fg0", "$bg1"),
		NewStyle("formgroup:title").WithColors("$fuchsia", "$bg1"),
		NewStyle("formgroup:label").WithColors("$fg2", "$bg1"),
		NewStyle("grid").WithBorder("round"),
		NewStyle("list/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("list/highlight:focused").WithColors("$bg0", "$fuchsia"),
		NewStyle("editor/current-line").WithColors("$fg0", "$bg1"),
		NewStyle("editor/current-line-number").WithColors("$fuchsia", "$bg1"),
		NewStyle("editor/line-numbers").WithColors("$fg2", "$bg0"),
		NewStyle("editor/selection").WithColors("$bg0", "$indigo"),
		NewStyle("input").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("input:focused").WithColors("$bg0", "$fuchsia"),
		NewStyle("typeahead").WithColors("$fg0", "$bg2").WithCursor("*bar"),
		NewStyle("typeahead:focused").WithColors("$bg0", "$fuchsia"),
		NewStyle("typeahead/suggestion").WithColors("$fg2", "$bg2"),
		NewStyle("typeahead/suggestion:focused").WithColors("$fg2", "$fuchsia"),
		NewStyle("rule"),
		NewStyle("scanner").WithColors("$cyan", "$bg0"),
		NewStyle("sparkline").WithColors("$cyan", "$bg0"),
		NewStyle("sparkline/high").WithColors("$yellow", "$bg0"),
		NewStyle("heatmap").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/header").WithColors("$fg2", "$bg0"),
		NewStyle("heatmap/zero").WithColors("$fg2", "$bg2"),
		NewStyle("heatmap/mid").WithColors("$fg0", "$bg0"),
		NewStyle("heatmap/max").WithColors("$bg0", "$green"),
		NewStyle("static").WithColors("$fg0", "$bg1").WithMargin(0).WithPadding(0),
		NewStyle("static.dialog").WithColors("$fg0", "$bg2").WithMargin(0).WithPadding(0),
		NewStyle("select").WithColors("$fg0", "$bg2").WithPadding(0, 1),
		NewStyle("select:focused").WithColors("$bg0", "$fuchsia"),
		NewStyle("combo").WithColors("$fg0", "$bg2").WithPadding(0, 1),
		NewStyle("combo:focused").WithColors("$bg0", "$fuchsia"),
		NewStyle("styled").WithColors("$fg0", "$bg1").WithPadding(0, 1),
		NewStyle("styled/h1").WithFont("bold"),
		NewStyle("styled/h2").WithFont("bold"),
		NewStyle("styled/h3").WithFont("bold underline"),
		NewStyle("styled/h4").WithFont("bold"),
		NewStyle("styled/pre").WithBackground("$bg0"),
		NewStyle("styled/code").WithForeground("$fuchsia"),
		NewStyle("styled/bq").WithColors("$fg2", "$bg1"),
		NewStyle("styled/hr").WithForeground("$gray"),
		NewStyle("switcher"),
		NewStyle("table").WithColors("", "").WithBorder("thin $fg2"),
		NewStyle("table:focused").WithBorder("double $fuchsia"),
		NewStyle("table/grid").WithColors("$fg2", "$bg0").WithBorder("thin"),
		NewStyle("table/grid:focused").WithBorder("double-thin"),
		NewStyle("table/header").WithColors("$fuchsia", "$bg0").WithFont("bold"),
		NewStyle("table/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("table/highlight:focused").WithColors("$bg0", "$fuchsia"),
		NewStyle("table/cell").WithColors("$fg0", "$bg3"),
		NewStyle("table/cell:focused").WithColors("$bg0", "$indigo").WithFont("bold"),
		NewStyle("tabs/highlight").WithColors("$bg0", "$fg1"),
		NewStyle("tabs/highlight-line").WithForeground("$fg2"),
		NewStyle("tabs/line:focused").WithForeground("$indigo"),
		NewStyle("tabs/highlight:focused").WithColors("$bg0", "$fuchsia"),
		NewStyle("tabs/highlight-line:focused").WithForeground("$fuchsia"),
		NewStyle("text"),
		NewStyle("tree").WithColors("$fg0", "$bg0"),
		NewStyle("tree/highlight").WithColors("$bg0", "$fg2"),
		NewStyle("tree/highlight:focused").WithColors("$bg0", "$fuchsia"),
		NewStyle("tree/indent").WithColors("$fg2", ""),
		NewStyle("viewport"),
		NewStyle("terminal").WithColors("$fg0", "$bg0"),
		NewStyle("terminal:focused").WithColors("$fg0", "$bg0"),
		NewStyle("shortcuts").WithColors("$fg2", "$bg0"),
		NewStyle("shortcuts/key").WithForeground("$fuchsia").WithFont("bold"),
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

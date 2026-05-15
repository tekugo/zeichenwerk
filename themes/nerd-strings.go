package themes

import (
	. "github.com/tekugo/zeichenwerk/core"
)

func AddNerdStrings(theme *Theme) {
	theme.SetStrings(map[string]string{
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

		// ---- Radio ----
		"radio.on":  "\uF111",
		"radio.off": "\uF10C",

		// ---- Slider ----
		"slider.compact.track":    "\u2501",
		"slider.compact.thumb":    "\u2503",
		"slider.box.top-left":     "\u256D",
		"slider.box.top-right":    "\u256E",
		"slider.box.bottom-left":  "\u2570",
		"slider.box.bottom-right": "\u256F",
		"slider.box.horizontal":   "\u2500",
		"slider.box.thumb-top":    "\u2565",
		"slider.box.thumb-bottom": "\u2568",

		// ---- Select ----
		"select.dropdown": " \u25BC",

		// ---- Shortcuts ----
		"shortcuts.prefix":    "",
		"shortcuts.separator": "   ",
		"shortcuts.suffix":    "",

		// ---- Tree ----
		"tree.expanded":  " ▼",
		"tree.collapsed": " ▶",
		"tree.branch":    " ├─",
		"tree.last":      " └─",
		"tree.trunk":     " │ ",

		// ---- Bar Chart ----
		"bar-chart.corner": "└",
		"bar-chart.hline":  "─",
		"bar-chart.vline":  "│",
		"bar-chart.tick-x": "┬",
		"bar-chart.tick-y": "┤",
		"bar-chart.grid":   "┄",
		"bar-chart.swatch": "█",

		// ---- Breadcrumb ----
		"breadcrumb.separator": " › ",
		"breadcrumb.overflow":  "…",

		// ---- Typewriter ----
		"typewriter.cursor": "▌",
	})
}

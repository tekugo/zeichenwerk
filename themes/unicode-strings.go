package themes

import (
	. "github.com/tekugo/zeichenwerk/core"
)

func AddUnicodeStrings(theme *Theme) {
	theme.SetStrings(map[string]string{
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

		// ---- Radio ----
		"radio.on":  "◉",
		"radio.off": "○",

		// ---- Slider ----
		// Compact (height 1) — heavy horizontal track with a heavy vertical thumb.
		"slider.compact.track": "━",
		"slider.compact.thumb": "┃",
		// Two-row (height ≥ 2) — rounded outer box, double-stem thumb (╥/╨)
		// joining the single-line border so the thumb stands out without
		// breaking the box's horizontal flow.
		"slider.box.top-left":     "╭",
		"slider.box.top-right":    "╮",
		"slider.box.bottom-left":  "╰",
		"slider.box.bottom-right": "╯",
		"slider.box.horizontal":   "─",
		"slider.box.thumb-top":    "╥",
		"slider.box.thumb-bottom": "╨",

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

package main

import (
	"fmt"
	"time"

	z "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/compose"
)

var info = `
# Configuring Claude Code for OTLP

Messwerk receives Claude Code telemetry over **OTLP/gRPC** and displays token usage, cost, and activity in real time.

## Quick start

Add the following to your Claude Code settings.json (user, project, or local scope):

` + "```" + `
{
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "OTEL_METRICS_EXPORTER": "otlp",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317"
  }
}
` + "```" + `

Then start messwerk before launching Claude Code:

` + "```" + `sh
messwerk
` + "```" + ` 

## Variables

| Variable | Value | Purpose |
|---|---|---|
| CLAUDE_CODE_ENABLE_TELEMETRY | 1 | Enables telemetry collection |
| OTEL_METRICS_EXPORTER | otlp | Routes metrics to the OTLP exporter |
| OTEL_EXPORTER_OTLP_PROTOCOL | grpc | Uses gRPC transport |
| OTEL_EXPORTER_OTLP_ENDPOINT | http://localhost:4317 | Messwerk's listen address |

## Options

**Custom port** — if 4317 is already in use, start messwerk on a different port and update the endpoint accordingly:

` + "```" + `sh
messwerk -port 4318
` + "```" + ` 

` + "```" + `json
"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318"
` + "```" + `

**Faster updates** — by default Claude Code exports metrics every 60 seconds. For a more responsive display during development, reduce the interval:

` + "```" + `json
"OTEL_METRIC_EXPORT_INTERVAL": "10000"
` + "```" + ` 
`

func buildUI(theme *z.Theme, store *Store) *z.UI {
	mp := &metricsProvider{}
	lp := &logProvider{}

	ui := UI(theme,
		Flex("main", "", false, "stretch", 0,
			Flex("header", "", true, "stretch", 0,
				Static("title", "", " messwerk  ·  Claude Code OTLP Monitor", Hint(-1, 1), Font("bold")),
				Clock("clock", "", time.Second, time.DateTime, "  "),
			),
			Grid("grid", "", []int{-1}, []int{41, -1}, true, Hint(0, -1),
				Cell(0, 1, 1, 1,
					Deck("sessions", "", deckRenderer(theme, store), 5),
				),
				Cell(1, 1, 1, 1,
					Switcher("content", "",
						Flex("info-panel", "", false, "stretch", 1, Padding(1),
							Flex("info-buttons", "", true, "strech", 0,
								Spacer("", Hint(-1, 0)),
								Button("info-copy", "", "Copy to Clipboard"),
							),
							Styled("info-text", "", info, Hint(0, -1), Padding(1, 2)),
						),
						Grid("all", "", []int{0, 0, 0, -1}, []int{26, 0, -1}, false, Border("none"), Padding(1, 2),
							Cell(0, 0, 1, 1, Card("total-input", "", "Input",
								Digits("total-input-value", "", "0", Flag(z.FlagRight, true), Fg("$blue")),
								Static("total-input-footer", "", "tokens", Fg("$gray"), Flag(z.FlagRight)),
							)),
							Cell(1, 0, 2, 1, Card("total-input-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("total-input-sparkline", "", Hint(30, 3), Fg("$blue")),
								Static("total-input-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							Cell(0, 1, 1, 1, Card("total-output", "", "Output",
								Digits("total-output-value", "", "0", Flag(z.FlagRight, true), Fg("$green")),
								Static("totla-output-footer", "", "tokens", Fg("$gray"), Flag(z.FlagRight)),
							)),
							Cell(1, 1, 2, 1, Card("total-output-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("total-output-sparkline", "", Hint(30, 3), Fg("$green")),
								Static("total-output-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							Cell(0, 2, 1, 1, Card("total-cost", "", "Cost",
								Digits("total-cost-value", "", "0.00", Flag(z.FlagRight, true), Fg("$cyan")),
								Static("total-cost-footer", "", "US$", Fg("$gray"), Flag(z.FlagRight)),
							)),
							Cell(1, 2, 2, 1, Card("total-cost-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("total-cost-sparkline", "", Hint(30, 3), Fg("$cyan")),
								Static("total-cost-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
						),
						Grid("session", "", []int{0, 0, 0, 0, 0, -1}, []int{26, 26, -1}, false, Border("none"), Padding(1, 2),
							// Row 0: Session header
							Cell(0, 0, 3, 1,
								Flex("session-header", "", true, "stretch", 1, Margin(0, 0, 1, 0),
									Static("session-id", "", "–", Hint(-1, 1), Font("bold"), Fg("$cyan")),
									Static("session-info", "", "–", Fg("$gray")),
								),
							),
							// Row 1: Stats
							Cell(0, 1, 1, 1, Card("session-cache-card", "", "Cache",
								Progress("session-cache-bar", "", true, Margin(1)),
								Static("session-cache-label", "", "–", Fg("$gray"), Flag(z.FlagRight)),
							)),
							Cell(1, 1, 1, 1, Card("session-accept-card", "", "Edits", Margin(0, 0, 0, 2),
								Progress("session-accept-bar", "", true, Margin(1)),
								Static("session-accept-label", "", "–", Fg("$gray"), Flag(z.FlagRight)),
							)),
							Cell(2, 1, 1, 1, Card("session-lines-card", "", "Lines", Margin(0, 0, 0, 2),
								Flex("session-lines-row", "", true, "center", 1,
									Static("session-lines-added", "", "+0", Fg("$green")),
									Spacer("", Hint(-1, 0)),
									Static("session-lines-removed", "", "-0", Fg("$red")),
								),
							)),
							// Rows 2–4: Metrics + sparklines
							Cell(0, 2, 1, 1, Card("session-input", "", "Input",
								Digits("session-input-value", "", "0", Flag(z.FlagRight, true), Fg("$blue")),
								Static("session-input-footer", "", "tokens", Fg("$gray"), Flag(z.FlagRight)),
							)),
							Cell(1, 2, 2, 1, Card("session-input-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("session-input-sparkline", "", Hint(30, 3), Fg("$blue")),
								Static("session-input-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							Cell(0, 3, 1, 1, Card("session-output", "", "Output",
								Digits("session-output-value", "", "0", Flag(z.FlagRight, true), Fg("$green")),
								Static("session-output-footer", "", "tokens", Fg("$gray"), Flag(z.FlagRight)),
							)),
							Cell(1, 3, 2, 1, Card("session-output-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("session-output-sparkline", "", Hint(30, 3), Fg("$green")),
								Static("session-output-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							Cell(0, 4, 1, 1, Card("session-cost", "", "Cost",
								Digits("session-cost-value", "", "0.00", Flag(z.FlagRight, true), Fg("$cyan")),
								Static("session-cost-footer", "", "US$", Fg("$gray"), Flag(z.FlagRight)),
							)),
							Cell(1, 4, 2, 1, Card("session-cost-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("session-cost-sparkline", "", Hint(30, 3), Fg("$cyan")),
								Static("session-cost-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							// Row 5: Tabs + tables
							Cell(0, 5, 3, 1,
								Flex("session-tables-area", "", false, "stretch", 0, Margin(1, 0, 0, 0),
									Tabs("session-tabs", ""),
									Switcher("session-table-switcher", "", Hint(0, -1),
										Table("session-metrics-table", "", mp, false, Hint(0, -1)),
										Table("session-logs-table", "", lp, false, Hint(0, -1)),
									),
								),
							),
						),
					),
				),
			),
			Shortcuts("footer", "", []string{"q", "Quit"}),
		),
	)

	deck := z.Find(ui, "sessions").(*z.Deck)
	content := z.Find(ui, "content").(*z.Switcher)

	// Session header
	sessionHeader := z.Find(ui, "session-header").(*z.Flex)
	sessionID := z.Find(ui, "session-id").(*z.Static)
	sessionInfo := z.Find(ui, "session-info").(*z.Static)

	// Stats row
	cacheBar := z.Find(ui, "session-cache-bar").(*z.Progress)
	cacheLabel := z.Find(ui, "session-cache-label").(*z.Static)
	acceptBar := z.Find(ui, "session-accept-bar").(*z.Progress)
	acceptLabel := z.Find(ui, "session-accept-label").(*z.Static)
	linesAdded := z.Find(ui, "session-lines-added").(*z.Static)
	linesRemoved := z.Find(ui, "session-lines-removed").(*z.Static)
	cacheBar.SetTotal(100)
	acceptBar.SetTotal(100)

	// Metrics + sparklines
	costDigits := z.Find(ui, "session-cost-value").(*z.Digits)
	inputDigits := z.Find(ui, "session-input-value").(*z.Digits)
	outputDigits := z.Find(ui, "session-output-value").(*z.Digits)
	inputFooter := z.Find(ui, "session-input-footer").(*z.Static)
	outputFooter := z.Find(ui, "session-output-footer").(*z.Static)
	inputSparkline := z.Find(ui, "session-input-sparkline").(*z.Sparkline)
	outputSparkline := z.Find(ui, "session-output-sparkline").(*z.Sparkline)
	costSparkline := z.Find(ui, "session-cost-sparkline").(*z.Sparkline)

	// Tabs + tables
	sessionTabs := z.Find(ui, "session-tabs").(*z.Tabs)
	tableSwitcher := z.Find(ui, "session-table-switcher").(*z.Switcher)
	sessionTabs.Add("Metrics")
	sessionTabs.Add("Logs")
	sessionTabs.On(z.EvtChange, func(_ z.Widget, _ z.Event, args ...any) bool {
		tableSwitcher.Select(args[0].(int))
		return true
	})

	z.Derived(store.TotalInput, func(n int64) string { return z.Humanize(n) }).
		Bind(z.Find(ui, "total-input-value").(*z.Digits))
	z.Derived(store.TotalOutput, func(n int64) string { return z.Humanize(n) }).
		Bind(z.Find(ui, "total-output-value").(*z.Digits))
	z.Derived(store.TotalCost, func(f float64) string { return fmt.Sprintf("%.4f", f) }).
		Bind(z.Find(ui, "total-cost-value").(*z.Digits))

	z.Find(ui, "total-input-sparkline").(*z.Sparkline).SetProvider(store.Input)
	z.Find(ui, "total-output-sparkline").(*z.Sparkline).SetProvider(store.Output)
	z.Find(ui, "total-cost-sparkline").(*z.Sparkline).SetProvider(store.Cost)
	deck.Set([]any{nil})

	showSession := func(session *Session) {
		agg := session.Aggregate()

		// Header
		id := session.ID
		if len(id) > 18 {
			id = id[:18] + "…"
		}
		sessionID.Set(id)
		start := session.Start.Format("15:04:05")
		duration := z.FormatDuration(time.Since(session.Start))
		sessionInfo.Set(fmt.Sprintf("start %s, duration %s, model %s, os %s/%s", start, duration, session.PrimaryModel(), session.OSType, session.HostArch))
		sessionHeader.Layout()

		// Cache ratio: what fraction of total prompt tokens came from cache
		totalTokens := agg.Input + agg.CacheRead + agg.CacheCreation
		cacheRatio := 0
		if totalTokens > 0 {
			cacheRatio = int(float64(agg.CacheRead) * 100 / float64(totalTokens))
		}
		cacheBar.Set(cacheRatio)
		cacheLabel.Set(fmt.Sprintf("%d%% cached", cacheRatio))

		// Accept ratio: how often code edit suggestions were accepted
		totalEdits := agg.Accepted + agg.Rejected
		acceptRatio := 0
		if totalEdits > 0 {
			acceptRatio = int(float64(agg.Accepted) * 100 / float64(totalEdits))
		}
		acceptBar.Set(acceptRatio)
		acceptLabel.Set(fmt.Sprintf("%d%% accepted", acceptRatio))

		// Lines changed
		linesAdded.Set(fmt.Sprintf("+%s", z.Humanize(agg.LinesAdded)))
		linesRemoved.Set(fmt.Sprintf("-%s", z.Humanize(agg.LinesRemoved)))

		// Metrics + sparklines
		lastSeen := session.LastSeen().Format("15:04:05")
		costDigits.Set(fmt.Sprintf("%.4f", session.TotalCost()))
		inputDigits.Set(z.Humanize(agg.Input))
		outputDigits.Set(z.Humanize(agg.Output))
		inputFooter.Set(lastSeen)
		outputFooter.Set(lastSeen)
		inputSparkline.SetProvider(session.Input)
		outputSparkline.SetProvider(session.Output)
		costSparkline.SetProvider(session.Cost)

		// Tables
		mp.set(session)
		lp.set(session)
	}

	deck.On(z.EvtSelect, func(_ z.Widget, _ z.Event, args ...any) bool {
		index := args[0].(int)
		if index == 0 {
			content.Select(1)
			return true
		}
		sessions := store.Items()
		if index-1 >= len(sessions) {
			return true
		}
		showSession(sessions[index-1])
		content.Select(2)
		return true
	})

	store.SetOnChange(func() {
		sessions := store.Items()
		items := make([]any, 1+len(sessions))
		items[0] = nil
		for i, s := range sessions {
			items[i+1] = s
		}
		deck.Set(items)
		if len(sessions) > 0 && content.Selected() == 0 {
			content.Select(1)
		}
		if idx := deck.Selected(); idx > 0 && idx-1 < len(sessions) {
			showSession(sessions[idx-1])
		}
		ui.Refresh()
	})

	z.Find(ui, "clock").(*z.Clock).Start()
	return ui
}

func deckRenderer(theme *z.Theme, store *Store) z.ItemRender {
	sparkline := z.NewSparkline("deck-dummy", "")
	sparkline.Apply(theme)
	sparkline.SetStyle("", sparkline.Style().Modifiable())
	renderSparkline := func(r *z.Renderer, x, y, w, h int, fg, bg string, data z.DataProvider) {
		sparkline.SetProvider(data)
		sparkline.SetBounds(x, y, w, h)
		sparkline.Style().WithColors(fg, bg)
		sparkline.Render(r)
	}

	return func(r *z.Renderer, x, y, w, h, index int, data any, selected, focused bool) {
		var bg string
		if selected {
			bg = theme.Color("$bg3")
		} else {
			bg = theme.Color("$bg1")
		}

		// Clear background
		r.Set("", bg, "")
		r.Fill(x, y, w, h, " ")

		// Draw indicator if selected
		if selected {
			var fg string
			if focused {
				fg = theme.Color("$blue")
			} else {
				fg = theme.Color("fg2")
			}
			r.Set(fg, bg, "")
			r.Line(x, y, 0, 1, h-2, "▍", "▍", "▍")
		}

		// Content starts at x+2
		cx := x + 2

		// Sparkline data
		var input, output, costTS *z.TimeSeries[float64]
		var model string
		var totalCost float64

		// Session name + second-line data
		if index == 0 {
			r.Set("$cyan", bg, "")
			r.Text(cx, y, "All Sessions", 0)
			input = store.Input
			output = store.Output
			costTS = store.Cost
			totalCost = store.TotalCost.Get()
		} else {
			session := data.(*Session)
			r.Set("$cyan", bg, "")
			r.Text(cx, y, session.ID, 0)
			input = session.Input
			output = session.Output
			costTS = session.Cost
			model = session.PrimaryModel()
			totalCost = session.TotalCost()
		}

		// Second line: model left, cost right-aligned
		if model != "" {
			r.Set("$fg2", bg, "")
			r.Text(cx, y+1, model, 0)
		}
		costStr := fmt.Sprintf("$%.2f", totalCost)
		r.Set("$cyan", bg, "")
		r.Text(x+w-len(costStr)-1, y+1, costStr, 0)

		// Render spark lines
		r.Set("$fg2", bg, "")
		r.Text(cx, y+2, "Input", 0)
		r.Text(cx, y+3, "Output", 0)
		r.Text(cx, y+4, "Cost", 0)
		renderSparkline(r, cx+8, y+2, 30, 1, "$blue", bg, input)
		renderSparkline(r, cx+8, y+3, 30, 1, "$green", bg, output)
		renderSparkline(r, cx+8, y+4, 30, 1, "$cyan", bg, costTS)
	}
}

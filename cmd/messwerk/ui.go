package main

import (
	"fmt"
	"time"

	z "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/compose"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/values"
	"github.com/tekugo/zeichenwerk/widgets"
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

func buildUI(theme *core.Theme, store *Store) *z.UI {
	mp := &metricsProvider{}
	lp := &logProvider{}

	ui := UI(theme,
		VFlex("main", "", core.Stretch, 0,
			HFlex("header", "", core.Stretch, 0,
				Static("title", "", " messwerk  ·  Claude Code OTLP Monitor", Hint(-1, 1), Font("bold")),
				Clock("clock", "", time.Second, time.DateTime, "  "),
			),
			Grid("grid", "", []int{-1}, []int{41, -1}, true, Hint(0, -1),
				Cell(0, 1, 1, 1,
					Deck("sessions", "", deckRenderer(theme, store), 5),
				),
				Cell(1, 1, 1, 1,
					Switcher("content", "",
						VFlex("info-panel", "", core.Stretch, 1, Padding(1),
							HFlex("info-buttons", "", core.Stretch, 0,
								Spacer("", Hint(-1, 0)),
								Button("info-copy", "", "Copy to Clipboard"),
							),
							Styled("info-text", "", info, Hint(0, -1), Padding(1, 2)),
						),
						Grid("all", "", []int{0, 0, 0, -1}, []int{26, 0, -1}, false, Border("none"), Padding(1, 2),
							Cell(0, 0, 1, 1, Card("total-input", "", "Input",
								Digits("total-input-value", "", "0", Flag(widgets.FlagRight, true), Fg("$blue")),
								Static("total-input-footer", "", "tokens", Fg("$gray"), Flag(widgets.FlagRight)),
							)),
							Cell(1, 0, 2, 1, Card("total-input-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("total-input-sparkline", "", Hint(30, 3), Fg("$blue")),
								Static("total-input-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							Cell(0, 1, 1, 1, Card("total-output", "", "Output",
								Digits("total-output-value", "", "0", Flag(widgets.FlagRight, true), Fg("$green")),
								Static("totla-output-footer", "", "tokens", Fg("$gray"), Flag(widgets.FlagRight)),
							)),
							Cell(1, 1, 2, 1, Card("total-output-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("total-output-sparkline", "", Hint(30, 3), Fg("$green")),
								Static("total-output-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							Cell(0, 2, 1, 1, Card("total-cost", "", "Cost",
								Digits("total-cost-value", "", "0.00", Flag(widgets.FlagRight, true), Fg("$cyan")),
								Static("total-cost-footer", "", "US$", Fg("$gray"), Flag(widgets.FlagRight)),
							)),
							Cell(1, 2, 2, 1, Card("total-cost-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("total-cost-sparkline", "", Hint(30, 3), Fg("$cyan")),
								Static("total-cost-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
						),
						Grid("session", "", []int{0, 0, 0, 0, 0, -1}, []int{26, 26, -1}, false, Border("none"), Padding(1, 2),
							// Row 0: Session header
							Cell(0, 0, 3, 1,
								HFlex("session-header", "", core.Stretch, 1, Margin(0, 0, 1, 0),
									Static("session-id", "", "–", Hint(-1, 1), Font("bold"), Fg("$cyan")),
									Static("session-info", "", "–", Fg("$gray")),
								),
							),
							// Row 1: Stats
							Cell(0, 1, 1, 1, Card("session-cache-card", "", "Cache",
								Progress("session-cache-bar", "", true, Margin(1)),
								Static("session-cache-label", "", "–", Fg("$gray"), Flag(widgets.FlagRight)),
							)),
							Cell(1, 1, 1, 1, Card("session-accept-card", "", "Edits", Margin(0, 0, 0, 2),
								Progress("session-accept-bar", "", true, Margin(1)),
								Static("session-accept-label", "", "–", Fg("$gray"), Flag(widgets.FlagRight)),
							)),
							Cell(2, 1, 1, 1, Card("session-lines-card", "", "Lines", Margin(0, 0, 0, 2),
								HFlex("session-lines-row", "", core.Center, 1,
									Static("session-lines-added", "", "+0", Fg("$green")),
									Spacer("", Hint(-1, 0)),
									Static("session-lines-removed", "", "-0", Fg("$red")),
								),
							)),
							// Rows 2–4: Metrics + sparklines
							Cell(0, 2, 1, 1, Card("session-input", "", "Input",
								Digits("session-input-value", "", "0", Flag(widgets.FlagRight, true), Fg("$blue")),
								Static("session-input-footer", "", "tokens", Fg("$gray"), Flag(widgets.FlagRight)),
							)),
							Cell(1, 2, 2, 1, Card("session-input-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("session-input-sparkline", "", Hint(30, 3), Fg("$blue")),
								Static("session-input-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							Cell(0, 3, 1, 1, Card("session-output", "", "Output",
								Digits("session-output-value", "", "0", Flag(widgets.FlagRight, true), Fg("$green")),
								Static("session-output-footer", "", "tokens", Fg("$gray"), Flag(widgets.FlagRight)),
							)),
							Cell(1, 3, 2, 1, Card("session-output-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("session-output-sparkline", "", Hint(30, 3), Fg("$green")),
								Static("session-output-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							Cell(0, 4, 1, 1, Card("session-cost", "", "Cost",
								Digits("session-cost-value", "", "0.00", Flag(widgets.FlagRight, true), Fg("$cyan")),
								Static("session-cost-footer", "", "US$", Fg("$gray"), Flag(widgets.FlagRight)),
							)),
							Cell(1, 4, 2, 1, Card("session-cost-sparkline-card", "", "", Margin(0, 0, 0, 2),
								Sparkline("session-cost-sparkline", "", Hint(30, 3), Fg("$cyan")),
								Static("session-cost-sparkline-footer", "", "Last 60 minutes, 2 minute intervals", Fg("$gray")),
							)),
							// Row 5: Tabs + tables
							Cell(0, 5, 3, 1,
								VFlex("session-tables-area", "", core.Stretch, 0, Margin(1, 0, 0, 0),
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

	deck := core.Find(ui, "sessions").(*widgets.Deck)
	content := core.Find(ui, "content").(*widgets.Switcher)

	// Session header
	sessionHeader := core.Find(ui, "session-header").(*widgets.Flex)
	sessionID := core.Find(ui, "session-id").(*widgets.Static)
	sessionInfo := core.Find(ui, "session-info").(*widgets.Static)

	// Stats row
	cacheBar := core.Find(ui, "session-cache-bar").(*widgets.Progress)
	cacheLabel := core.Find(ui, "session-cache-label").(*widgets.Static)
	acceptBar := core.Find(ui, "session-accept-bar").(*widgets.Progress)
	acceptLabel := core.Find(ui, "session-accept-label").(*widgets.Static)
	linesAdded := core.Find(ui, "session-lines-added").(*widgets.Static)
	linesRemoved := core.Find(ui, "session-lines-removed").(*widgets.Static)
	cacheBar.SetTotal(100)
	acceptBar.SetTotal(100)

	// Metrics + sparklines
	costDigits := core.Find(ui, "session-cost-value").(*widgets.Digits)
	inputDigits := core.Find(ui, "session-input-value").(*widgets.Digits)
	outputDigits := core.Find(ui, "session-output-value").(*widgets.Digits)
	inputFooter := core.Find(ui, "session-input-footer").(*widgets.Static)
	outputFooter := core.Find(ui, "session-output-footer").(*widgets.Static)
	inputSparkline := core.Find(ui, "session-input-sparkline").(*widgets.Sparkline)
	outputSparkline := core.Find(ui, "session-output-sparkline").(*widgets.Sparkline)
	costSparkline := core.Find(ui, "session-cost-sparkline").(*widgets.Sparkline)

	// Tabs + tables
	sessionTabs := core.Find(ui, "session-tabs").(*widgets.Tabs)
	tableSwitcher := core.Find(ui, "session-table-switcher").(*widgets.Switcher)
	sessionTabs.Add("Metrics")
	sessionTabs.Add("Logs")
	sessionTabs.On(widgets.EvtChange, func(_ core.Widget, _ core.Event, args ...any) bool {
		tableSwitcher.Select(args[0].(int))
		return true
	})

	values.Derived(store.TotalInput, func(n int64) string { return core.Humanize(n) }).
		Bind(core.Find(ui, "total-input-value").(*widgets.Digits))
	values.Derived(store.TotalOutput, func(n int64) string { return core.Humanize(n) }).
		Bind(core.Find(ui, "total-output-value").(*widgets.Digits))
	values.Derived(store.TotalCost, func(f float64) string { return fmt.Sprintf("%.2f", f) }).
		Bind(core.Find(ui, "total-cost-value").(*widgets.Digits))

	core.Find(ui, "total-input-sparkline").(*widgets.Sparkline).SetProvider(store.Input)
	core.Find(ui, "total-output-sparkline").(*widgets.Sparkline).SetProvider(store.Output)
	core.Find(ui, "total-cost-sparkline").(*widgets.Sparkline).SetProvider(store.Cost)
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
		duration := core.FormatDuration(time.Since(session.Start))
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
		linesAdded.Set(fmt.Sprintf("+%s", core.Humanize(agg.LinesAdded)))
		linesRemoved.Set(fmt.Sprintf("-%s", core.Humanize(agg.LinesRemoved)))

		// Metrics + sparklines
		lastSeen := session.LastSeen().Format("15:04:05")
		costDigits.Set(fmt.Sprintf("%.2f", session.TotalCost()))
		inputDigits.Set(core.Humanize(agg.Input))
		outputDigits.Set(core.Humanize(agg.Output))
		inputFooter.Set(lastSeen)
		outputFooter.Set(lastSeen)
		inputSparkline.SetProvider(session.Input)
		outputSparkline.SetProvider(session.Output)
		costSparkline.SetProvider(session.Cost)

		// Tables
		mp.set(session)
		lp.set(session)
	}

	deck.On(widgets.EvtSelect, func(_ core.Widget, _ core.Event, args ...any) bool {
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
		if len(sessions) > 0 && content.Get() == 0 {
			content.Select(1)
		}
		if idx := deck.Selected(); idx > 0 && idx-1 < len(sessions) {
			showSession(sessions[idx-1])
		}
		ui.Refresh()
	})

	core.Find(ui, "clock").(*widgets.Clock).Start()
	return ui
}

func deckRenderer(theme *core.Theme, store *Store) widgets.ItemRender {
	sparkline := widgets.NewSparkline("deck-dummy", "")
	sparkline.Apply(theme)
	sparkline.SetStyle("", sparkline.Style().Modifiable())
	renderSparkline := func(r *core.Renderer, x, y, w, h int, fg, bg string, data widgets.DataProvider) {
		sparkline.SetProvider(data)
		sparkline.SetBounds(x, y, w, h)
		sparkline.Style().WithColors(fg, bg)
		sparkline.Render(r)
	}

	return func(r *core.Renderer, x, y, w, h, index int, data any, selected, focused bool) {
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
		var input, output, costTS *core.TimeSeries[float64]
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

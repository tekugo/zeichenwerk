package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	. "github.com/tekugo/zeichenwerk"
)

const otlpConfig = `{
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "OTEL_METRICS_EXPORTER": "otlp",
    "OTEL_LOGS_EXPORTER": "otlp",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
    "OTEL_METRIC_EXPORT_INTERVAL": "10000",
    "OTEL_LOGS_EXPORT_INTERVAL": "5000"
  }
}`

// buildUI constructs the messwerk TUI and wires the store's onChange callback.
func buildUI(theme *Theme, store *Store) *UI {
	b := NewBuilder(theme)
	configLines := strings.Split(otlpConfig, "\n")

	b.Flex("main", false, "stretch", 0).
		Tabs("tabs", "Start", "Monitor", "Config", "About").
		Switcher("view", false).Hint(0, -1).
		With(func(b *Builder) {
			// Start pane — OTLP setup instructions shown before data arrives.
			b.Flex("start-pane", false, "stretch", 0).Hint(0, -1).Padding(3, 6).
				Static("info-heading", "Waiting for telemetry data").
				Static("info-desc", "Add this to ~/.claude/settings.json to enable OTLP export to messwerk:").
				Text("info-config", configLines, false, 0).Hint(0, len(configLines)).
				Button("copy-btn", "Copy config to clipboard").
			End()
		}).
		With(func(b *Builder) {
			// Monitor pane — deck on the left, detail on the right.
			b.Grid("nav-grid", 1, 2, false).Hint(0, -1).
				Rows(-1).
				Columns(42, -1).
				Cell(0, 0, 1, 1).Deck("nav", makeItemRender(theme, store), 3).
				Cell(1, 0, 1, 1).Flex("detail", false, "stretch", 0).
				Flex("stats", true, "stretch", 0).Hint(0, 4).
				Flex("stat-input", false, "start", 0).Hint(-1, 0).Padding(0, 1).
				Static("lbl-input", "Input").
				Static("val-input", "—").
				End().
				VRule("thin").
				Flex("stat-output", false, "start", 0).Hint(-1, 0).Padding(0, 1).
				Static("lbl-output", "Output").
				Static("val-output", "—").
				End().
				VRule("thin").
				Flex("stat-cache", false, "start", 0).Hint(-1, 0).Padding(0, 1).
				Static("lbl-cache", "Cache").
				Static("val-cache", "—").
				End().
				VRule("thin").
				Flex("stat-cost", false, "start", 0).Hint(-1, 0).Padding(0, 1).
				Static("lbl-cost", "Cost").
				Static("val-cost", "—").
				End().
				End().
				HRule("thin").
				Sparkline("detail-spark").Hint(0, 3).
				HRule("thin").
				Text("detail-list", nil, false, 500).
				End().
				End()
		}).
		With(func(b *Builder) {
			// Config pane (placeholder).
			b.Flex("config-pane", false, "stretch", 0).Hint(0, -1).Padding(3, 6).
				Static("config-placeholder", "Config inspector — coming soon").
			End()
		}).
		With(func(b *Builder) {
			// About pane (placeholder).
			b.Flex("about-pane", false, "stretch", 0).Hint(0, -1).Padding(3, 6).
				Static("about-placeholder", "About — coming soon").
			End()
		}).
		End().
		Shortcuts("footer", "↑↓", "navigate", "←→", "switch tabs", "q", "quit").
		End()

	ui := b.Build()

	tabs := Find(ui, "tabs").(*Tabs)
	viewSw := Find(ui, "view").(*Switcher)
	deck := Find(ui, "nav").(*Deck)
	detailSpark := Find(ui, "detail-spark").(*Sparkline)
	detailSpark.SetMode(Absolute)
	detailSpark.SetMin(0)
	valInput := Find(ui, "val-input").(*Static)
	valOutput := Find(ui, "val-output").(*Static)
	valCache := Find(ui, "val-cache").(*Static)
	valCost := Find(ui, "val-cost").(*Static)
	detailList := Find(ui, "detail-list").(*Text)
	copyBtn := Find(ui, "copy-btn").(*Button)

	// currentTab tracks the active tab index so refresh() can auto-switch.
	var currentTab int

	// Wire both EvtChange (arrow key) and EvtActivate (Enter) on the tabs
	// so the view switcher follows navigation immediately in both cases.
	switchToTab := func(idx int) {
		currentTab = idx
		viewSw.Select(idx)
	}
	tabs.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			if idx, ok := data[0].(int); ok {
				switchToTab(idx)
			}
		}
		return true
	})
	tabs.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			if idx, ok := data[0].(int); ok {
				switchToTab(idx)
			}
		}
		return true
	})

	// Copy button: write OTLP config to clipboard and show brief confirmation.
	copyBtn.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if err := copyToClipboard(otlpConfig); err == nil {
			copyBtn.Set("✓ Copied!")
			Redraw(copyBtn)
			go func() {
				time.Sleep(2 * time.Second)
				copyBtn.Set("Copy config to clipboard")
				Redraw(copyBtn)
			}()
		}
		return true
	})

	var currentItems []*SessionItem

	updateDetail := func(item *SessionItem) {
		valInput.SetText(formatTokens(item.InputTokens))
		valOutput.SetText(formatTokens(item.OutputTokens))
		valCache.SetText(formatTokens(item.CacheTokens))
		valCost.SetText(fmt.Sprintf("$%.4f", item.TotalCost))

		vals := item.Sparkline.Values()
		max := 1.0
		for _, v := range vals {
			if v > max {
				max = v
			}
		}
		detailSpark.SetMax(max)
		detailSpark.SetValues(vals)

		detailList.Clear()
		if item.ID == "__gesamt__" && len(currentItems) > 1 {
			for _, s := range currentItems[1:] {
				_, dot := statusDot(s.Status, theme)
				line := fmt.Sprintf("%s %-28s  %8s tok  $%.4f  %s",
					dot, truncate(s.Name, 28),
					formatTokens(s.TotalTokens), s.TotalCost,
					sessionAge(s.Sparkline))
				detailList.Add(line)
			}
		}
	}

	refresh := func() {
		items := store.Items()
		currentItems = items

		// Auto-switch from Start to Monitor when first real session appears.
		if currentTab == 0 && len(items) > 1 {
			tabs.Select(1) // fires EvtChange+EvtActivate → switchToTab(1)
		}

		// Update deck, preserving the user's current selection if possible.
		any := make([]any, len(items))
		for i, it := range items {
			any[i] = it
		}
		prevIdx := deck.Selected()
		deck.SetItems(any) // resets index to 0, queues Redraw(deck)
		if prevIdx > 0 && prevIdx < len(items) {
			// deck.Select fires EvtSelect → updateDetail via the handler below.
			deck.Select(prevIdx)
		} else if len(items) > 0 {
			updateDetail(items[deck.Selected()])
		}

		// Guarantee a full screen redraw so every visible widget reflects
		// the latest data, regardless of how the Refresh chain is routed.
		ui.Refresh()
	}

	deck.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		idx, ok := data[0].(int)
		if !ok || idx < 0 || idx >= len(currentItems) {
			return false
		}
		updateDetail(currentItems[idx])
		return true
	})

	store.SetOnChange(refresh)
	refresh()
	return ui
}

// copyToClipboard writes text to the system clipboard.
func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// makeItemRender returns the Deck ItemRender for session cards.
// Each card occupies 3 rows:
//
//	Row 0: status dot + session name
//	Row 1: token count + cost (right-aligned)
//	Row 2: sparkline (last 20 minutes, relative scale)
func makeItemRender(theme *Theme, store *Store) ItemRender {
	return func(r *Renderer, x, y, w, h, _ int, data any, selected, focused bool) {
		item := data.(*SessionItem)

		bg := theme.Color("$bg1")
		if selected {
			bg = theme.Color("$bg3")
		}
		r.Set("", bg, "")
		r.Fill(x, y, w, h, " ")

		// Selection indicator.
		indicatorFg := theme.Color("$fg2")
		indicator := " "
		if selected {
			indicator = "▍"
			if focused {
				indicatorFg = theme.Color("$blue")
			}
		}
		r.Set(indicatorFg, bg, "")
		for row := 0; row < h; row++ {
			r.Put(x, y+row, indicator)
		}

		contentX := x + 2

		// Row 0: status dot + name.
		dotFg, dot := statusDot(item.Status, theme)
		r.Set(dotFg, bg, "")
		r.Put(contentX, y, dot)

		nameFg := theme.Color("$fg1")
		nameFont := ""
		if selected {
			nameFg = theme.Color("$fg0")
			nameFont = "bold"
		}
		r.Set(nameFg, bg, nameFont)
		r.Text(contentX+2, y, item.Name, w-(contentX+2-x))

		// Row 1: tokens + cost, right-aligned.
		if h > 1 {
			tokStr := formatTokens(item.TotalTokens)
			costStr := fmt.Sprintf("$%.2f", item.TotalCost)
			line1 := tokStr + " tok  " + costStr
			startX := x + w - len(line1) - 1
			if startX < contentX {
				startX = contentX
			}
			r.Set(theme.Color("$fg2"), bg, "")
			r.Text(startX, y+1, line1, w-(startX-x))
		}

		// Row 2: sparkline via the Sparkline component.
		if h > 2 {
			sp := item.Sparkline
			sp.Apply(theme)
			spStyle := sp.Style()
			spFg := theme.Color(spStyle.Foreground())
			sp.SetStyle("", NewStyle().WithForeground(spStyle.Foreground()).WithBackground(bg))
			spW := w - (contentX - x)
			sp.SetBounds(contentX, y+2, spW, 1)
			sp.Render(r)

			// Draw ▁ baseline for leading columns with no data.
			emptyCount := spW - len(sp.Values())
			if emptyCount > 0 {
				r.Set(spFg, bg, "")
				for i := 0; i < emptyCount; i++ {
					r.Put(contentX+i, y+2, "▁")
				}
			}
		}
	}
}

// statusDot returns the (fg colour, glyph) for the given status.
func statusDot(st SessionStatus, theme *Theme) (string, string) {
	switch st {
	case StatusActive:
		return theme.Color("$green"), "●"
	case StatusIdle:
		return theme.Color("$yellow"), "●"
	default:
		return theme.Color("$fg2"), "○"
	}
}

// sessionAge returns a human-readable "last seen" string derived from the
// most recent non-zero sparkline bucket timestamp.
func sessionAge(sp *Sparkline) string {
	vals := sp.Values()
	idle := 0
	for i := len(vals) - 1; i >= 0; i-- {
		if vals[i] > 0 {
			break
		}
		idle++
	}
	if len(vals) == 0 || idle == len(vals) {
		return "—"
	}
	d := time.Duration(idle) * time.Minute
	if d < time.Minute {
		return "just now"
	}
	return fmt.Sprintf("%dm ago", int(d.Minutes()))
}

// truncate shortens s to at most n runes, appending … if needed.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// formatTokens formats a token count compactly: 123, 12.4k, 1.2M.
func formatTokens(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

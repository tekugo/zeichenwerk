// Demo2 — per-widget reference demo with documentation, Builder snippets,
// and Compose snippets shown alongside a live demo of every widget.
package main

import (
	"flag"
	"os"
	"strings"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/inspector"
	"github.com/tekugo/zeichenwerk/themes"
	. "github.com/tekugo/zeichenwerk/values"
	. "github.com/tekugo/zeichenwerk/widgets"
)

func main() {
	registerEntries(
		containerEntries,
		inputEntries,
		displayEntries,
		animatedEntries,
		customEntries,
	)

	t := flag.String("t", "tokyo", "Theme: tokyo, midnight, nord, gruvbox-dark, gruvbox-light, lipstick")
	dump := flag.Bool("dump", false, "Dump widget hierarchy and exit")
	flag.Parse()

	theme := pickTheme(*t)
	ui := buildUI(theme)
	if *dump {
		ui.SetBounds(0, 0, 140, 40)
		ui.Layout()
		ui.Dump(os.Stdout, DumpOptions{Style: false})
		return
	}
	inspector.Open(ui)
	ui.Run()
}

func pickTheme(name string) *Theme {
	switch name {
	case "midnight":
		return themes.MidnightNeon()
	case "nord":
		return themes.Nord()
	case "gruvbox-dark":
		return themes.GruvboxDark()
	case "gruvbox-light":
		return themes.GruvboxLight()
	case "lipstick":
		return themes.Lipstick()
	default:
		return themes.TokyoNight()
	}
}

func buildUI(theme *Theme) *UI {
	items, indexMap := navItems()

	b := NewBuilder(theme).
		VFlex("root", Stretch, 0).
		// ── Header ─────────────────────────────────────────────────────────
		HFlex("header", Stretch, 1).Padding(0, 1).
		Static("title", "Zeichenwerk Demo 2").
		Spacer().Hint(-1, 0).
		Static("count", "").
		Spacer().Hint(2, 0).
		Static("theme-label", "Theme:").
		Select("theme-select", "tokyo", "Tokyo Night",
			"midnight", "Midnight Neon",
			"nord", "Nord",
			"gruvbox-dark", "Gruvbox Dark",
			"gruvbox-light", "Gruvbox Light",
			"lipstick", "Lipstick").
		End().
		HRule("thin").
		// ── Body ───────────────────────────────────────────────────────────
		HFlex("body", Stretch, 0).Hint(0, -1).
		// Left: navigation
		VFlex("nav-pane", Stretch, 0).Hint(34, 0).
		Box("nav-box", "Widgets").Border("round").Hint(34, -1).
		List("navigation", items...).Hint(0, -1).Flag(FlagSearch).
		End().
		End().
		// Right: a single Box hosting the tabs + switcher; title = current widget.
		Box("widget-box", "").Border("round").Hint(-1, -1).
		VFlex("right-pane", Stretch, 0).Hint(-1, -1).
		Tabs("tabs", "Demo", "Documentation", "Builder code", "Compose code").
		Switcher("page-switcher", true).Hint(-1, -1).
		// 0: Demo pane — switcher containing every widget demo
		Switcher("demo-switcher", false).Hint(-1, -1).
		End().
		// 1: Docs pane
		Styled("docs", "").Hint(-1, -1).
		// 2: Builder code pane
		Styled("builder-code", "").Hint(-1, -1).
		// 3: Compose code pane
		Styled("compose-code", "").Hint(-1, -1).
		End(). // end page-switcher
		End(). // end right-pane VFlex
		End(). // end widget-box
		End(). // end body HFlex
		// ── Footer ─────────────────────────────────────────────────────────
		HFlex("footer", Center, 0).
		Shortcuts("footer-shortcuts",
			"↑↓", "navigate",
			"Tab", "focus",
			"1-4", "tabs",
			"/", "search",
			"q", "quit").
		Spacer().Hint(-1, 0).
		Static("status", "Select a widget on the left").
		End()

	// Add all demo widgets into the demo switcher.
	demoSwitcher := b.Find("demo-switcher").(*Switcher)
	_ = demoSwitcher // populated below via b.With
	for _, e := range allEntries {
		// Build a stand-alone container per entry into a temporary builder.
		nb := NewBuilder(theme)
		e.DemoFn(nb)
		w := nb.Container()
		demoSwitcher.Add(w)
	}

	ui := b.Build()

	// Theme switching.
	all := map[string]*Theme{
		"tokyo":         themes.TokyoNight(),
		"midnight":      themes.MidnightNeon(),
		"nord":          themes.Nord(),
		"gruvbox-dark":  themes.GruvboxDark(),
		"gruvbox-light": themes.GruvboxLight(),
		"lipstick":      themes.Lipstick(),
	}
	Find(ui, "theme-select").On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 1 {
			if k, ok := data[0].(string); ok {
				if t, found := all[k]; found {
					ui.SetTheme(t)
				}
			}
		}
		return true
	})

	// Show the entry count and initial selection.
	Update(ui, "count", formatCount(len(allEntries)))

	// Wire up navigation. Section-header rows aren't selectable; if the user
	// lands on one, hop further in the direction they were moving so they can
	// pass through it instead of getting bounced back.
	nav := Find(ui, "navigation").(*List)
	isHeader := func(i int) bool {
		return i < 0 || i >= len(items) || strings.HasPrefix(items[i], "─ ")
	}
	switchToEntry := func(rowIndex int) {
		entryIdx, ok := indexMap[rowIndex]
		if !ok {
			return
		}
		showEntry(ui, entryIdx)
	}
	prevIdx := -1
	nav.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		if len(data) != 1 {
			return true
		}
		idx, ok := data[0].(int)
		if !ok {
			return true
		}
		if !isHeader(idx) {
			prevIdx = idx
			switchToEntry(idx)
			return true
		}
		// Header row — figure out the desired direction. Default to forward
		// (covers the initial state and clicks).
		dir := 1
		if prevIdx >= 0 && idx < prevIdx {
			dir = -1
		}
		// Walk past any consecutive headers in that direction.
		next := idx + dir
		for next >= 0 && next < len(items) && isHeader(next) {
			next += dir
		}
		// If we ran off either end, reverse direction and try again so the
		// list never stalls on a header.
		if next < 0 || next >= len(items) {
			dir = -dir
			next = idx + dir
			for next >= 0 && next < len(items) && isHeader(next) {
				next += dir
			}
		}
		if next >= 0 && next < len(items) {
			nav.Select(next)
			prevIdx = next
			switchToEntry(next)
		}
		return true
	})
	nav.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 1 {
			if idx, ok := data[0].(int); ok {
				switchToEntry(idx)
			}
		}
		return true
	})

	// Number-key shortcuts on the right pane to switch tabs.
	tabs := Find(ui, "tabs").(*Tabs)
	switcher := Find(ui, "page-switcher").(*Switcher)
	OnKey(Find(ui, "root"), func(e *tcell.EventKey) bool {
		if e.Key() != tcell.KeyRune {
			return false
		}
		switch e.Str() {
		case "1", "2", "3", "4":
			i := int(e.Str()[0] - '1')
			tabs.Set(i)
			switcher.Select(i)
			return true
		case "q", "Q":
			ui.Quit()
			return true
		}
		return false
	})

	// Pick the first non-header row by default.
	for i := range items {
		if !isHeader(i) {
			nav.Select(i)
			showEntry(ui, indexMap[i])
			break
		}
	}

	return ui
}

// showEntry updates the four panes for the given entry index.
func showEntry(ui *UI, entryIdx int) {
	if entryIdx < 0 || entryIdx >= len(allEntries) {
		return
	}
	e := allEntries[entryIdx]

	// Switch the demo switcher to the entry's pane.
	if ds, ok := Find(ui, "demo-switcher").(*Switcher); ok {
		ds.Select(entryIdx)
	}

	// Update the docs / builder / compose Styled widgets.
	if w, ok := Find(ui, "docs").(*Styled); ok {
		w.SetText(buildDocText(e))
	}
	if w, ok := Find(ui, "builder-code").(*Styled); ok {
		w.SetText(codeBlock("Builder API", e.Builder))
	}
	if w, ok := Find(ui, "compose-code").(*Styled); ok {
		w.SetText(codeBlock("Compose API", e.Compose))
	}

	// Box title shows the widget name; the tabs inside provide the sub-header.
	if box, ok := Find(ui, "widget-box").(*Box); ok {
		box.Set(e.Name)
	}
	Update(ui, "status", e.Category+"  ·  "+e.Summary)
}

func buildDocText(e Entry) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(e.Name)
	b.WriteString("\n\n")
	b.WriteString("> ")
	b.WriteString(e.Summary)
	b.WriteString("\n\n")
	doc := e.Doc()
	// Strip the leading `# Title` line (we already render our own header).
	if i := strings.Index(doc, "\n"); i >= 0 && strings.HasPrefix(doc, "# ") {
		doc = doc[i+1:]
	}
	b.WriteString(strings.TrimSpace(doc))
	return b.String()
}

func codeBlock(title, code string) string {
	if code == "" {
		return "# " + title + "\n\n_No example available._"
	}
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n```\n")
	b.WriteString(strings.TrimSpace(code))
	b.WriteString("\n```\n")
	return b.String()
}

func formatCount(n int) string {
	return "  " + itoa(n) + " widgets  "
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk"
)

// buildUI constructs the full triebwerk widget tree.
func buildUI(theme *Theme, dir string) *UI {
	var deck *Deck

	b := NewBuilder(theme)

	b.Flex("root", false, "stretch", 0).
		// ── Header ──────────────────────────────────────────────────────────
		Static("header", headerText(dir)).Hint(0, 1).
		// ── Body ────────────────────────────────────────────────────────────
		Flex("body", true, "stretch", 0).Hint(0, -1).
			Box("targets-box", "Targets").Hint(34, 0).Border("thin").
				Flex("targets-panel", false, "stretch", 0).
					Add(buildDeck(theme, &deck)).Hint(0, -1).
					Flex("filter-section", false, "stretch", 0).Margin(1, 0, 0, 0).
						Static("filter-title", " filter").Foreground("$fg2").
						Flex("filter-bar", true, "start", 2).Hint(0, 1).Padding(0, 1).
							Checkbox("filter-make", "make", true).
							Checkbox("filter-just", "just", true).
						End().
					End().
				End().
			End().
			Box("output-box", "Output").Hint(-1, 0).Border("thin").
				Terminal("output").Hint(0, -1).
			End().
		End().
		// ── Footer ──────────────────────────────────────────────────────────
		Flex("footer", true, "center", 0).Hint(0, 1).
			Scanner("watch-scanner", 3, "circles").
			Static("footer-status", "").Hint(0, 1).
			Spacer().Hint(-1, 1).
			Shortcuts("footer-shortcuts", "r", "run", "w", "watch", "d", "dir", "m", "make", "j", "just", "c", "clear", "q", "quit").Padding(0, 1, 0, 0).
		End()

	ui := b.Build()

	allTargets, _ := loadTargets(dir)
	deck.SetItems(toItems(allTargets, dir))

	cbMake := Find(ui, "filter-make").(*Checkbox)
	cbJust := Find(ui, "filter-just").(*Checkbox)

	refilter := func(_ Widget, _ Event, _ ...any) bool {
		showMake := cbMake.Flag(FlagChecked)
		showJust := cbJust.Flag(FlagChecked)
		var filtered []Target
		for _, t := range allTargets {
			if t.Runner == "" || (t.Runner == "make" && showMake) || (t.Runner == "just" && showJust) {
				filtered = append(filtered, t)
			}
		}
		deck.SetItems(toItems(filtered, dir))
		return true
	}

	cbMake.On(EvtChange, refilter)
	cbJust.On(EvtChange, refilter)

	scanner := Find(ui, "watch-scanner").(*Scanner)
	shortcuts := Find(ui, "footer-shortcuts").(*Shortcuts)

	runner := NewRunner(
		Find(ui, "output").(*Terminal),
		Find(ui, "footer-status").(*Static),
		scanner,
		dir,
	)

	cfg, _ := loadConfig(dir)
	watcher := NewWatcher(runner, dir)

	toggleWatch := func(target Target) {
		if target.Runner == "" {
			return
		}
		if watcher.IsWatching(target.Name) {
			watcher.Stop(target.Name)
			if watcher.Count() == 0 {
				runner.SetWatchActive(false)
			}
			shortcuts.SetPairs("r", "run", "w", "watch", "d", "dir", "m", "make", "j", "just", "c", "clear", "q", "quit")
		} else {
			pattern := cfg.pattern(target.Name)
			if pattern == "" {
				ui.Prompt("Watch", "Glob pattern (e.g. **/*.go):", func(p string) {
					if p == "" {
						return
					}
					cfg.setPattern(dir, target.Name, p)
					saveConfig(dir, cfg)
					watcher.Start(target, p)
					runner.SetWatchActive(true)
					shortcuts.SetPairs("r", "run", "w", "stop", "d", "dir", "m", "make", "j", "just", "c", "clear", "q", "quit")
				}, nil)
			} else {
				watcher.Start(target, pattern)
				runner.SetWatchActive(true)
				shortcuts.SetPairs("r", "run", "w", "stop", "d", "dir", "m", "make", "j", "just", "c", "clear", "q", "quit")
			}
		}
	}

	OnKey(deck, func(e *tcell.EventKey) bool {
		if e.Key() != tcell.KeyRune {
			return false
		}
		switch e.Str() {
		case "m":
			cbMake.Toggle()
			return true
		case "j":
			cbJust.Toggle()
			return true
		case "r":
			if idx := deck.Selected(); idx >= 0 {
				runner.Enqueue(deck.Items()[idx].(Target))
			}
			return true
		case "w":
			if idx := deck.Selected(); idx >= 0 {
				toggleWatch(deck.Items()[idx].(Target))
			}
			return true
		case "c":
			runner.ClearTerminal()
			return true
		}
		return false
	})

	deck.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if idx := deck.Selected(); idx >= 0 {
			runner.Enqueue(deck.Items()[idx].(Target))
		}
		return true
	})

	return ui
}

// headerText formats the header line showing the project directory.
func headerText(dir string) string {
	home, _ := os.UserHomeDir()
	display := dir
	if rel, err := filepath.Rel(home, dir); err == nil && !filepath.IsAbs(rel) {
		display = "~/" + rel
	}
	return fmt.Sprintf("  triebwerk  —  %s", display)
}

// buildDeck creates the target Deck widget and writes its pointer to dst.
func buildDeck(theme *Theme, dst **Deck) *Deck {
	d := NewDeck("targets", "", renderTargetCard(theme), 3)
	*dst = d
	return d
}

// renderTargetCard returns an ItemRender function for target cards.
//
// Each card occupies 3 rows:
//
//	Row 0:  ▍ [badge] name
//	Row 1:    description
//	Row 2:  (empty gap between cards)
func renderTargetCard(theme *Theme) ItemRender {
	return func(r *Renderer, x, y, w, h, _ int, data any, selected, focused bool) {
		target := data.(Target)

		bg := theme.Color("$bg1")
		if selected {
			bg = theme.Color("$bg3")
		}

		r.Set("", bg, "")
		r.Fill(x, y, w, h, " ")

		// Left selection indicator.
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

		// Badge: [make] or [just]
		badgeFg := theme.Color("$blue")
		if target.Runner == "just" {
			badgeFg = theme.Color("$green")
		}
		badge := "[" + target.Runner + "]"
		badgeX := x + 2
		r.Set(badgeFg, bg, "")
		r.Text(badgeX, y, badge, len(badge))

		// Target name.
		nameFont := ""
		if selected {
			nameFont = "bold"
		}
		nameX := badgeX + len(badge) + 1
		r.Set(theme.Color("$fg0"), bg, nameFont)
		r.Text(nameX, y, target.Name, w-(nameX-x))

		// Description (row 1).
		if h > 1 && target.Description != "" {
			r.Set(theme.Color("$fg2"), bg, "")
			r.Text(badgeX, y+1, target.Description, w-(badgeX-x))
		}
	}
}

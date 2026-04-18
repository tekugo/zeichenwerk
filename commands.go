package zeichenwerk

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

// Command represents a named action that can be triggered from the commands
// palette.
type Command struct {
	Name     string // display label; used for fuzzy matching
	Shortcut string // hint string shown on the right ("Ctrl+O", ""); display only
	Group    string // optional section name; empty = ungrouped
	Action   func() // executed when the command is confirmed
}

// rankedCommand is a display-list entry produced by filterCommands.
// When isHeader is true the entry is a non-selectable group header row
// (cmd.Name holds the group label).
type rankedCommand struct {
	cmd      *Command
	score    int
	isHeader bool
}

// Commands is the command registry and palette controller. It is a lazy
// singleton owned by the UI; obtain it via [UI.Commands].
type Commands struct {
	ui       *UI
	entries  []*Command // registration order preserved
	maxItems int        // max visible rows before scrolling (default 10)
	width    int        // explicit popup width override; 0 = auto
	open     bool       // true while the popup is in the layer stack
}

func newCommands(ui *UI) *Commands {
	return &Commands{
		ui:       ui,
		maxItems: 10,
	}
}

// All returns a snapshot of the current registration slice.
func (c *Commands) All() []*Command {
	result := make([]*Command, len(c.entries))
	copy(result, c.entries)
	return result
}

// Close dismisses the palette. No-op when the palette is not open.
func (c *Commands) Close() {
	if !c.open {
		return
	}
	c.ui.Close() // triggers EvtClose → c.open = false
}

// IsOpen reports whether the palette is currently displayed.
func (c *Commands) IsOpen() bool {
	return c.open
}

// Register appends a command at the end of the registration list. Returns the
// *Command for chaining or later removal. group may be empty. Commands with the
// same group string appear under a shared header in the palette. name must be
// non-empty; shortcut and action may be empty/nil.
func (c *Commands) Register(group, name, shortcut string, action func()) *Command {
	cmd := &Command{Name: name, Shortcut: shortcut, Group: group, Action: action}
	c.entries = append(c.entries, cmd)
	return cmd
}

// SetMaxItems sets the maximum number of command rows visible before the list
// scrolls. Clamped to minimum 3. Default 10.
func (c *Commands) SetMaxItems(n int) {
	if n < 3 {
		n = 3
	}
	c.maxItems = n
}

// SetWidth forces the popup to a fixed column width, overriding auto-sizing.
// 0 = auto.
func (c *Commands) SetWidth(n int) {
	c.width = n
}

// Unregister removes the first command with the matching name (exact,
// case-sensitive). Returns true if found. No-op while the palette is open.
func (c *Commands) Unregister(name string) bool {
	if c.open {
		return false
	}
	for i, cmd := range c.entries {
		if cmd.Name == name {
			c.entries = append(c.entries[:i], c.entries[i+1:]...)
			return true
		}
	}
	return false
}

// ---- Display --------------------------------------------------------------

// Open displays the commands palette. No-op if already open.
//
// It builds a transient popup containing a Filter input and a commandsPanel,
// focuses the filter, and wires all interaction handlers.
func (c *Commands) Open() {
	if c.open {
		return
	}

	initialItems := c.filterCommands("")
	w, h := c.computePopupSize(initialItems)

	// Build dialog + flex + filter via the builder API.
	// The builder stack will be [Dialog, Flex] after the Filter call since
	// Filter is not a Container.
	b := c.ui.NewBuilder()
	b.Dialog("commands-dialog", "Commands").
		Flex("commands-flex", "stretch", 0).Flag(FlagVertical).
		Filter("commands-input").Hint(0, 1)

	// Manually add the commandsPanel to the Flex.
	flex := b.stack.Peek().(*Flex)
	panel := newCommandsPanel(c.maxItems)
	panel.SetItems(initialItems)
	flex.Add(panel)

	theme := c.ui.renderer.theme
	panel.Apply(theme)

	dialog := b.stack[0].(*Dialog)

	// Override dialog border/background with the "commands" theme style.
	cmdStyle := theme.Get("commands")
	dialog.SetStyle("", cmdStyle)

	// Override filter style with "commands/input" if defined.
	input := Find(dialog, "commands-input").(*Filter)
	inputStyle := theme.Get("commands/input")
	if inputStyle != &DefaultStyle {
		input.SetStyle("", inputStyle)
	}

	// Reset c.open when the dialog layer is removed (Escape, Close, or Enter).
	dialog.On(EvtClose, func(_ Widget, _ Event, _ ...any) bool {
		c.open = false
		return false
	})

	// Re-filter the panel on every keystroke.
	OnChange(input, func(text string) bool {
		items := c.filterCommands(text)
		panel.SetItems(items)
		c.ui.Refresh()
		return false
	})

	// Navigation and confirmation keys — prepended so they run before the
	// Filter's own handlers.
	OnKey(input, func(e *tcell.EventKey) bool {
		switch e.Key() {
		case tcell.KeyUp:
			panel.move(-1)
			c.ui.Refresh()
			return true
		case tcell.KeyDown:
			panel.move(+1)
			c.ui.Refresh()
			return true
		case tcell.KeyHome:
			panel.home()
			c.ui.Refresh()
			return true
		case tcell.KeyEnd:
			panel.end()
			c.ui.Refresh()
			return true
		case tcell.KeyEnter:
			if cmd := panel.focused(); cmd != nil {
				c.ui.Close()
				if cmd.Action != nil {
					cmd.Action()
				}
			}
			return true
		}
		return false
	})

	// Panel double-click activates the command (after the popup closes).
	panel.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			if cmd, ok := data[0].(*Command); ok {
				c.ui.Close()
				if cmd.Action != nil {
					cmd.Action()
				}
			}
		}
		return true
	})

	c.ui.Popup(-1, -1, w, h, dialog)
	c.open = true
}

// computePopupSize returns the popup dimensions based on the initial item list
// and the configured width/maxItems settings.
func (c *Commands) computePopupSize(items []rankedCommand) (w, h int) {
	visibleRows := len(items)
	if visibleRows > c.maxItems {
		visibleRows = c.maxItems
	}
	// 1 title + 1 border-top + 1 filter + 1 separator + rows + 1 border-bottom
	h = 5 + visibleRows

	if c.width > 0 {
		return c.width, h
	}

	nameW := 0
	shortcutW := 0
	for _, cmd := range c.entries {
		if n := utf8.RuneCountInString(cmd.Name); n > nameW {
			nameW = n
		}
		if s := utf8.RuneCountInString(cmd.Shortcut); s > shortcutW {
			shortcutW = s
		}
	}
	minW := nameW
	if shortcutW > 0 {
		minW += shortcutW + 3
	}
	minW += 4 // padding + border
	w = minW
	if w < 44 {
		w = 44
	}
	if c.ui != nil && c.ui.width > 0 {
		if maxW := c.ui.width - 4; w > maxW {
			w = maxW
		}
	}
	return w, h
}

// ── Fuzzy matching ────────────────────────────────────────────────────────────

// fuzzyMatch reports whether all characters in query appear in name in order
// (case-insensitive). When matched is true, score reflects word-boundary (+5),
// adjacency (+3), and exact-case (+1) bonuses.
func fuzzyMatch(query, name string) (matched bool, score int) {
	if query == "" {
		return true, 0
	}
	qRunes := []rune(query)
	nRunes := []rune(name)
	qLower := []rune(strings.ToLower(query))
	nLower := []rune(strings.ToLower(name))

	qi := 0
	positions := make([]int, 0, len(qRunes))
	for ni := range nLower {
		if qi < len(qLower) && nLower[ni] == qLower[qi] {
			positions = append(positions, ni)
			qi++
		}
	}
	if qi < len(qLower) {
		return false, 0
	}

	for i, pos := range positions {
		// Word-boundary bonus
		if pos == 0 {
			score += 5
		} else {
			prev := nRunes[pos-1]
			if prev == ' ' || prev == '/' || prev == '_' || prev == '-' {
				score += 5
			}
		}
		// Adjacency bonus
		if i > 0 && pos == positions[i-1]+1 {
			score += 3
		}
		// Exact case match bonus
		if nRunes[pos] == qRunes[i] {
			score += 1
		}
	}
	return true, score
}

// filterCommands scores and sorts the registered commands against query.
// When query is empty all commands are returned in registration order.
// When groups are in use they are preserved in first-appearance order.
func (c *Commands) filterCommands(query string) []rankedCommand {
	// Detect grouping
	hasGroups := false
	for _, cmd := range c.entries {
		if cmd.Group != "" {
			hasGroups = true
			break
		}
	}

	type scored struct {
		cmd   *Command
		score int
	}

	// Score and filter all commands
	results := make([]scored, 0, len(c.entries))
	for _, cmd := range c.entries {
		if query == "" {
			results = append(results, scored{cmd: cmd})
		} else {
			if ok, s := fuzzyMatch(query, cmd.Name); ok {
				results = append(results, scored{cmd: cmd, score: s})
			}
		}
	}

	if !hasGroups {
		if query != "" {
			slices.SortStableFunc(results, func(a, b scored) int {
				return b.score - a.score
			})
		}
		out := make([]rankedCommand, len(results))
		for i, s := range results {
			out[i] = rankedCommand{cmd: s.cmd, score: s.score}
		}
		return out
	}

	// Build score lookup (only for matched commands)
	matchedCmds := make(map[*Command]int, len(results))
	for _, r := range results {
		matchedCmds[r.cmd] = r.score
	}

	// Collect group order from first appearance in registration list
	groupOrder := make([]string, 0)
	groupSeen := make(map[string]bool)
	for _, cmd := range c.entries {
		if !groupSeen[cmd.Group] {
			groupOrder = append(groupOrder, cmd.Group)
			groupSeen[cmd.Group] = true
		}
	}

	// Collect matching items per group, preserving registration order
	groupItems := make(map[string][]scored, len(groupOrder))
	for _, cmd := range c.entries {
		score, ok := matchedCmds[cmd]
		if !ok {
			continue
		}
		groupItems[cmd.Group] = append(groupItems[cmd.Group], scored{cmd: cmd, score: score})
	}

	// Sort within each group by score descending (stable = ties keep registration order)
	if query != "" {
		for g := range groupItems {
			slices.SortStableFunc(groupItems[g], func(a, b scored) int {
				return b.score - a.score
			})
		}
	}

	// Build output: group header + items for each group that has matches
	var out []rankedCommand
	for _, group := range groupOrder {
		items := groupItems[group]
		if len(items) == 0 {
			continue
		}
		if group != "" {
			out = append(out, rankedCommand{
				cmd:      &Command{Name: group},
				isHeader: true,
			})
		}
		for _, s := range items {
			out = append(out, rankedCommand{cmd: s.cmd, score: s.score})
		}
	}
	return out
}

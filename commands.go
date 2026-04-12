package zeichenwerk

import (
	"slices"
	"strings"
	"time"
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

// ── commandsPanel ─────────────────────────────────────────────────────────────

// commandsPanel is an unexported widget that renders the scrollable command
// list inside the commands palette popup.
type commandsPanel struct {
	Component
	items    []rankedCommand
	index    int // absolute index into items; -1 when no selectable item
	offset   int // first visible item index (scroll position)
	maxItems int

	lastClickIndex int
	lastClickTime  time.Time
}

func newCommandsPanel(maxItems int) *commandsPanel {
	p := &commandsPanel{
		Component:      Component{id: "commands-panel"},
		index:          -1,
		maxItems:       maxItems,
		lastClickIndex: -1,
	}
	OnMouse(p, p.handleMouse)
	return p
}

// Apply sets visual styles for the panel rows.
func (p *commandsPanel) Apply(theme *Theme) {
	theme.Apply(p, "commands/item", "focused")
	theme.Apply(p, "commands/shortcut", "focused")
	theme.Apply(p, "commands/group")
}

// Hint returns the preferred size: width deferred to parent, height =
// min(len(items), maxItems) + 1 for the separator line.
func (p *commandsPanel) Hint() (int, int) {
	h := len(p.items)
	if h > p.maxItems {
		h = p.maxItems
	}
	if h < 0 {
		h = 0
	}
	return 0, h + 1
}

// SetItems replaces the visible list, resetting scroll and selection to the
// first non-header item.
func (p *commandsPanel) SetItems(items []rankedCommand) {
	p.items = items
	p.offset = 0
	p.index = -1
	for i, item := range items {
		if !item.isHeader {
			p.index = i
			break
		}
	}
}

// focused returns the currently selected Command, or nil when nothing is
// selected.
func (p *commandsPanel) focused() *Command {
	if p.index < 0 || p.index >= len(p.items) {
		return nil
	}
	return p.items[p.index].cmd
}

// move adjusts the selection by delta, skipping group header rows.
func (p *commandsPanel) move(delta int) {
	if len(p.items) == 0 || p.index < 0 {
		return
	}
	newIdx := p.index + delta
	if delta > 0 {
		for newIdx < len(p.items) && p.items[newIdx].isHeader {
			newIdx++
		}
		if newIdx >= len(p.items) {
			return
		}
	} else {
		for newIdx >= 0 && p.items[newIdx].isHeader {
			newIdx--
		}
		if newIdx < 0 {
			return
		}
	}
	p.index = newIdx
	p.ensureVisible()
}

// home jumps to the first selectable item.
func (p *commandsPanel) home() {
	for i, item := range p.items {
		if !item.isHeader {
			p.index = i
			p.offset = 0
			return
		}
	}
}

// end jumps to the last selectable item.
func (p *commandsPanel) end() {
	for i := len(p.items) - 1; i >= 0; i-- {
		if !p.items[i].isHeader {
			p.index = i
			p.ensureVisible()
			return
		}
	}
}

func (p *commandsPanel) ensureVisible() {
	if p.index < p.offset {
		p.offset = p.index
	} else if p.index >= p.offset+p.maxItems {
		p.offset = p.index - p.maxItems + 1
	}
}

// Render draws the separator line and all visible command rows.
func (p *commandsPanel) Render(r *Renderer) {
	if p.Flag(FlagHidden) {
		return
	}
	p.Component.Render(r)
	cx, cy, cw, _ := p.Content()

	// Separator between the filter input and the item list.
	sepStyle := p.Style("group")
	r.Set(sepStyle.Foreground(), sepStyle.Background(), "")
	r.Fill(cx, cy, cw, 1, "─")

	end := p.offset + p.maxItems
	if end > len(p.items) {
		end = len(p.items)
	}

	for row, item := range p.items[p.offset:end] {
		y := cy + 1 + row // +1 to skip the separator
		absIdx := p.offset + row

		if item.isHeader {
			style := p.Style("group")
			r.Set(style.Foreground(), style.Background(), style.Font())
			r.Fill(cx, y, cw, 1, " ")
			label := "── " + item.cmd.Name + " "
			labelW := utf8.RuneCountInString(label)
			r.Text(cx, y, label, cw)
			if labelW < cw {
				r.Repeat(cx+labelW, y, 1, 0, cw-labelW, "─")
			}
		} else {
			isFocused := absIdx == p.index
			var itemStyle, shortStyle *Style
			if isFocused {
				itemStyle = p.Style("item:focused")
				shortStyle = p.Style("shortcut:focused")
			} else {
				itemStyle = p.Style("item")
				shortStyle = p.Style("shortcut")
			}

			r.Set(itemStyle.Foreground(), itemStyle.Background(), itemStyle.Font())
			r.Fill(cx, y, cw, 1, " ")

			// Focus indicator (›) or blank
			if isFocused {
				r.Text(cx, y, "›", 1)
			}

			// Name — truncated to leave room for the shortcut
			shortcutW := 0
			if item.cmd.Shortcut != "" {
				shortcutW = utf8.RuneCountInString(item.cmd.Shortcut)
			}
			nameW := cw - 2
			if shortcutW > 0 {
				nameW = cw - 2 - shortcutW - 1
			}
			if nameW < 0 {
				nameW = 0
			}
			r.Text(cx+2, y, item.cmd.Name, nameW)

			// Shortcut — right-aligned in its own style
			if item.cmd.Shortcut != "" && shortcutW > 0 {
				r.Set(shortStyle.Foreground(), shortStyle.Background(), shortStyle.Font())
				r.Text(cx+cw-shortcutW, y, item.cmd.Shortcut, shortcutW)
			}
		}
	}
}

// handleMouse selects on single click; activates on double-click.
func (p *commandsPanel) handleMouse(e *tcell.EventMouse) bool {
	if e.Buttons() != tcell.Button1 {
		return false
	}
	_, my := e.Position()
	if my < p.y+1 || my >= p.y+p.height { // +1 to skip the separator row
		return false
	}
	row := my - p.y - 1 // -1 to account for the separator row
	absIdx := p.offset + row
	if absIdx < 0 || absIdx >= len(p.items) {
		return false
	}
	item := p.items[absIdx]
	if item.isHeader {
		return false
	}

	now := time.Now()
	if absIdx == p.lastClickIndex && now.Sub(p.lastClickTime) < doubleClickThreshold {
		// Double-click: activate the command
		p.lastClickIndex = -1
		p.Dispatch(p, EvtActivate, item.cmd)
		return true
	}

	// Single click: move selection
	p.index = absIdx
	p.lastClickIndex = absIdx
	p.lastClickTime = now
	p.ensureVisible()
	Redraw(p)
	return true
}

// ── Commands ──────────────────────────────────────────────────────────────────

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

// Register appends a command at the end of the registration list. Returns the
// *Command for chaining or later removal. name must be non-empty; shortcut and
// action may be empty/nil.
func (c *Commands) Register(name, shortcut string, action func()) *Command {
	cmd := &Command{Name: name, Shortcut: shortcut, Action: action}
	c.entries = append(c.entries, cmd)
	return cmd
}

// RegisterGroup is like Register but assigns a group label. Commands with the
// same group string appear under a shared header in the palette.
func (c *Commands) RegisterGroup(group, name, shortcut string, action func()) *Command {
	cmd := &Command{Name: name, Shortcut: shortcut, Group: group, Action: action}
	c.entries = append(c.entries, cmd)
	return cmd
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

// All returns a snapshot of the current registration slice.
func (c *Commands) All() []*Command {
	result := make([]*Command, len(c.entries))
	copy(result, c.entries)
	return result
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

// IsOpen reports whether the palette is currently displayed.
func (c *Commands) IsOpen() bool {
	return c.open
}

// Close dismisses the palette. No-op when the palette is not open.
func (c *Commands) Close() {
	if !c.open {
		return
	}
	c.ui.Close() // triggers EvtClose → c.open = false
}

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
		Flex("commands-flex", false, "stretch", 0).
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

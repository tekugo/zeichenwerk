package zeichenwerk

import (
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
	. "github.com/tekugo/zeichenwerk/v2/widgets"
)

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
		Component:      *NewComponent("commands-panel", ""),
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
	_, y, _, height := p.Bounds()
	_, my := e.Position()
	if my < y+1 || my >= y+height { // +1 to skip the separator row
		return false
	}
	row := my - y - 1 // -1 to account for the separator row
	absIdx := p.offset + row
	if absIdx < 0 || absIdx >= len(p.items) {
		return false
	}
	item := p.items[absIdx]
	if item.isHeader {
		return false
	}

	now := time.Now()
	if absIdx == p.lastClickIndex && now.Sub(p.lastClickTime) < DoubleClickThreshold {
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

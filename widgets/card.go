package widgets

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// Card is a bordered container with a title rendered in the top border line,
// a main content area, and an optional fixed-height footer.
//
// The first widget added via [Card.Add] becomes the content; the second
// becomes the footer. Further calls replace the footer (same behaviour as
// [Box] for its single child).
//
// Layout:
//   - When a footer is present, the content widget fills all available height
//     minus the footer's hint height; the footer is pinned to the bottom.
//   - When no footer is present, the content fills the entire content area.
//
// Style selectors: "card", "card/title".
type Card struct {
	Component
	Title   string // Optional title drawn inline with the top border
	content Widget
	footer  Widget
}

// NewCard creates a new card container with the given id, class, and title.
func NewCard(id, class, title string) *Card {
	return &Card{
		Component: Component{id: id, class: class},
		Title:     title,
	}
}

// ---- Widget interface -------------------------------------------------------

// Apply registers the "card" and "card/title" theme styles.
func (c *Card) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("card"))
}

// Hint returns the preferred content size: the maximum of content and footer
// widths, and the sum of their heights (each including their own style overhead).
func (c *Card) Hint() (int, int) {
	if c.hwidth != 0 || c.hheight != 0 {
		return c.hwidth, c.hheight
	}
	var w, h int
	if c.content != nil {
		cw, ch := c.content.Hint()
		s := c.content.Style()
		cw += s.Horizontal()
		ch += s.Vertical()
		if cw > w {
			w = cw
		}
		h += ch
	}
	if c.footer != nil {
		fw, fh := c.footer.Hint()
		s := c.footer.Style()
		fw += s.Horizontal()
		fh += s.Vertical()
		if fw > w {
			w = fw
		}
		h += fh
	}
	return w, h
}

// Render draws the card border, the title inline with the top border, then
// the content and footer children.
func (c *Card) Render(r *Renderer) {
	c.Component.Render(r)

	state := c.State()
	if state != "" {
		state = ":" + state
	}
	style := c.Style(state)

	if c.Title != "" {
		titleStyle := c.Style("title")
		r.Set(titleStyle.Foreground(), titleStyle.Background(), titleStyle.Font())
		r.Text(c.x+style.Margin().Left+2, c.y+style.Margin().Top, " "+c.Title+" ", 0)
	}
	if c.content != nil {
		c.content.Render(r)
	}
	if c.footer != nil {
		c.footer.Render(r)
	}
}

// ---- Container interface ---------------------------------------------------

// Add routes the first widget to the content slot and the second to the
// footer slot. Further calls replace the footer.
func (c *Card) Add(widget Widget, _ ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	if c.content == nil {
		c.content = widget
		widget.SetParent(c)
	} else {
		if c.footer != nil {
			c.footer.SetParent(nil)
		}
		c.footer = widget
		widget.SetParent(c)
	}
	return nil
}

// Children returns the content widget followed by the footer widget (if set).
func (c *Card) Children() []Widget {
	out := make([]Widget, 0, 2)
	if c.content != nil {
		out = append(out, c.content)
	}
	if c.footer != nil {
		out = append(out, c.footer)
	}
	return out
}

// Layout positions the content to fill the available space minus the footer
// height, and pins the footer to the bottom of the content area.
func (c *Card) Layout() error {
	cx, cy, cw, ch := c.Content()

	if c.footer != nil {
		_, fh := c.footer.Hint()
		s := c.footer.Style()
		fh += s.Vertical()
		if fh > ch {
			fh = 0 // footer doesn't fit; collapse it
		}
		contentH := ch - fh
		if c.content != nil {
			c.content.SetBounds(cx, cy, cw, contentH)
		}
		c.footer.SetBounds(cx, cy+contentH, cw, fh)
	} else if c.content != nil {
		c.content.SetBounds(cx, cy, cw, ch)
	}
	return Layout(c)
}

// ---- Setter ----------------------------------------------------------------

// Set updates the card's title and triggers a refresh.
func (c *Card) Set(value string) {
	c.Title = value
	c.Refresh()
}

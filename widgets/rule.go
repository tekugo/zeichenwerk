package widgets

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// Rule is a horizontal or vertical line as a visual divider for content.
type Rule struct {
	Component
	horizontal bool
	style      string
}

// NewHRule creates a horizontal rule using the named border style.
// The rule has a fixed hint height of 1.
func NewHRule(class, style string) *Rule {
	return &Rule{
		Component:  Component{id: "hrule", class: class, hheight: 1},
		horizontal: true,
		style:      style,
	}
}

// NewVRule creates a vertical rule using the named border style.
// The rule has a fixed hint width of 1.
func NewVRule(class, style string) *Rule {
	return &Rule{
		Component:  Component{id: "vrule", class: class, hwidth: 1},
		horizontal: false,
		style:      style,
	}
}

// Apply applies a theme's styles to the component.
func (c *Rule) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("rule"))
}

// Render draws the rule as a horizontal or vertical line using the configured
// border style's inner-H or inner-V character. If the theme does not define
// the requested border, the method falls back to the "default" border and
// finally skips rendering so a missing theme asset degrades gracefully
// instead of crashing the renderer.
func (c *Rule) Render(r *Renderer) {
	c.Component.Render(r)

	b := r.Theme.Border(c.style)
	if b == nil {
		b = r.Theme.Border("default")
	}
	if b == nil {
		return
	}

	x, y, w, h := c.Content()
	if c.horizontal {
		r.Line(x, y, 1, 0, w-2, b.InnerH, b.InnerH, b.InnerH)
	} else {
		r.Line(x, y, 0, 1, h-2, b.InnerV, b.InnerV, b.InnerV)
	}
}

package zeichenwerk

// Rule is a horizontal or vertical line as a visual divider for content.
type Rule struct {
	Component
	horizontal bool
	style      string
}

func NewHRule(style string) *Rule {
	return &Rule{
		Component:  Component{id: "hrule", hheight: 1},
		horizontal: true,
		style:      style,
	}
}

func NewVRule(style string) *Rule {
	return &Rule{
		Component:  Component{id: "vrule", hwidth: 1},
		horizontal: false,
		style:      style,
	}
}

// Render the rule
func (c *Rule) Render(r *Renderer) {
	c.Component.Render(r)

	x, y, w, h := c.Content()
	b := r.theme.Border(c.style)
	if c.horizontal {
		r.Line(x, y, 1, 0, w-2, b.InnerH, b.InnerH, b.InnerH)
	} else {
		r.Line(x, y, 0, 1, h-2, b.InnerV, b.InnerV, b.InnerV)
	}
}

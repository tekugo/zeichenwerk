package zeichenwerk

// Custom component is basically just a stub component, which gets a custom
// render function. They still have all the basic component functionality,
// including event handlers, styling and focusing.
type Custom struct {
	Component
	renderer func(Widget, *Renderer)
}

// Creates a new custom component.
func NewCustom(id, class string, fn func(Widget, *Renderer)) *Custom {
	return &Custom{
		Component: Component{id: id, class: class},
		renderer:  fn,
	}
}

// Apply applies a theme style to the component.
func (c *Custom) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("custom"), "disbled", "focused", "hovered")
}

// Renders the custom component. The rendering is delegated to the renderer
// function, which was passed during construction. If the component should
// render border and styling, you have to call `Component.Render()` on the
// passed widget.
func (c *Custom) Render(renderer *Renderer) {
	c.Component.Render(renderer)

	if c.renderer != nil {
		c.renderer(c, renderer)
	}
}

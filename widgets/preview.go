package widgets

import (
	. "github.com/tekugo/zeichenwerk/core"
)

// Preview is a single-child container that *lies* about its
// children: Children() returns an empty slice, hiding the wrapped
// subtree from every framework walk that uses Children to traverse —
// focus traversal (Tab / Shift-Tab), mouse hit testing (FindAt),
// generic event dispatch, and the standard layer-render walker all
// stop at the Preview node. Layout and Render bypass the lie and
// drive the wrapped widget directly, so the subtree still appears,
// lays out correctly, and updates as its state changes.
//
// The widget is the right shape for a designer's preview pane: the
// designer wants to render a tree exactly the way the production UI
// would, but doesn't want any of the wrapped widgets to receive
// focus or react to mouse / keyboard input. Setting FlagSkip on
// every preview widget would cover focus traversal but would still
// allow mouse hit-testing to land on inner widgets, would leak into
// the form's Skip checkbox, and would leak into codegen as
// .Flag(FlagSkip, true) calls — none of which are desirable.
//
// Inspector tooling that needs to walk the previewed subtree
// (designer tree, codegen) calls Target() instead of Children();
// the Preview is invisible to the framework but transparent to
// callers that ask it directly.
type Preview struct {
	Component
	target Widget
}

// NewPreview creates a Preview that wraps target. Pass nil to
// build an empty Preview that gets a target via Add later.
func NewPreview(id, class string, target Widget) *Preview {
	p := &Preview{
		Component: Component{id: id, class: class},
		target:    target,
	}
	if target != nil {
		target.SetParent(p)
	}
	return p
}

// Apply applies the theme. Preview uses the "preview" selector; it
// falls back to whatever the theme's default-empty selector picks
// up when "preview" isn't registered explicitly, which is the
// common case (themes don't customise it today).
func (p *Preview) Apply(theme *Theme) {
	theme.Apply(p, p.Selector("preview"))
}

// Target returns the wrapped widget, or nil when no target has
// been set. Inspector tooling that needs to walk into the preview
// uses this accessor; the framework itself sees a childless leaf.
func (p *Preview) Target() Widget {
	return p.target
}

// SetTarget replaces the wrapped widget. The previous target's
// parent reference is cleared so Find walks rooted at the old
// target don't mistakenly think it's still inside this Preview.
func (p *Preview) SetTarget(w Widget) {
	if p.target != nil {
		p.target.SetParent(nil)
	}
	p.target = w
	if w != nil {
		w.SetParent(p)
	}
}

// Add satisfies Container so the Builder can use Preview in chain
// construction (b.Preview(id) ... End()). The first widget added
// becomes the target; subsequent Adds replace the previous target.
// Returns ErrChildIsNil for nil widgets, matching Box's contract.
func (p *Preview) Add(widget Widget, params ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	p.SetTarget(widget)
	return nil
}

// Children returns an empty slice. This is the load-bearing lie:
// every framework walker that uses Children to find descendants
// (focus, hit-testing, event dispatch, render-walker, Find,
// FindAll, Traverse) stops here. The wrapped target is reached
// only through Target().
func (p *Preview) Children() []Widget {
	return nil
}

// Insert sets the preview target. Only index 0 is valid; any other
// value returns ErrFull. The Preview deliberately hides its target
// from Children() walks, but the underlying single-slot semantics
// match every other one-child container.
func (p *Preview) Insert(index int, widget Widget, _ ...any) error {
	if widget == nil {
		return ErrChildIsNil
	}
	if index != 0 {
		return ErrFull
	}
	p.SetTarget(widget)
	return nil
}

// Remove clears the preview target if it matches child. Returns
// ErrNotFound otherwise.
func (p *Preview) Remove(child Widget) error {
	if child == nil {
		return ErrChildIsNil
	}
	if p.target != child {
		return ErrNotFound
	}
	p.SetTarget(nil)
	return nil
}

// Hint returns the preferred size of the wrapped target plus the
// Preview's own style overhead, falling back to (0, 0) when there
// is no target.
func (p *Preview) Hint() (int, int) {
	if p.hwidth != 0 || p.hheight != 0 {
		return p.hwidth, p.hheight
	}
	if p.target == nil {
		return 0, 0
	}
	w, h := p.target.Hint()
	style := p.target.Style()
	return w + style.Horizontal(), h + style.Vertical()
}

// Layout sizes the target to Preview's content area and triggers
// recursive layout of the target subtree. The Layout helper used
// by other containers walks Children, which is empty for Preview;
// we have to invoke target.Layout explicitly.
func (p *Preview) Layout() error {
	if p.target == nil {
		return nil
	}
	cx, cy, cw, ch := p.Content()
	p.target.SetBounds(cx, cy, cw, ch)
	if c, ok := p.target.(Container); ok {
		return c.Layout()
	}
	return nil
}

// Render draws the Preview's own component chrome (border etc.)
// and then renders the wrapped target. The framework's standard
// render path renders Preview's Children, which is empty, so
// without this explicit call the target wouldn't appear.
func (p *Preview) Render(r *Renderer) {
	p.Component.Render(r)
	if p.target != nil {
		p.target.Render(r)
	}
}

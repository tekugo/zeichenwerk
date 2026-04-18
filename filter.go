package zeichenwerk

// ==== AI ===================================================================

// Filter is a standalone input widget that progressively filters a bound [List]
// or [Tree] as the user types. It embeds [Typeahead] and adds a Bind/Unbind
// mechanism plus the filtering side-effect on every EvtChange.
//
// Ghost-text (prefix completion) and list filtering intentionally use different
// matching semantics: ghost text shows the first item whose text starts with the
// typed text; the bound widget shows all items that contain it as a substring.
type Filter struct {
	Typeahead
	bound Filterable
}

// NewFilter creates a new Filter widget with the default placeholder "Filter…".
// An EvtChange handler is prepended so applyFilter runs before any
// user-registered handlers.
func NewFilter(id, class string) *Filter {
	f := &Filter{}
	// Initialise the embedded Typeahead/Input fields directly so that all
	// method values and closures registered below are bound to the embedded
	// structs inside f, not to a separate heap-allocated copy.
	f.Component = Component{id: id, class: class, hheight: 1}
	f.buf = NewGapBufferFromString("", 16)
	f.placeholder = "Filter…"
	f.mask = "*"
	f.SetFlag(FlagFocusable, true)
	f.SetFlag(FlagMasked, false)
	f.SetFlag(FlagReadonly, false)
	f.Input.refresh = func() { Redraw(f) }

	// Key handlers — same order as NewTypeahead: Input's handler is registered
	// first so that On (which prepends) puts Typeahead's handler in front.
	OnKey(f, f.Input.handleKey)
	OnKey(f, f.Typeahead.handleKey)

	// Suggestion update is registered before applyFilter so that it runs after
	// applyFilter (On prepends, so the last-registered handler runs first).
	OnChange(f, func(value string) bool {
		f.Typeahead.updateSuggestion(value)
		return false
	})

	// applyFilter: prepended last, so it runs first on every text change.
	OnChange(f, func(value string) bool {
		f.applyFilter()
		return false
	})

	return f
}

// Apply registers theme styles for the filter widget. In addition to the
// inherited "typeahead" and "typeahead/suggestion" selectors it applies the
// "filter" selector, enabling callers to style the filter field independently
// from plain inputs. Falls back to "typeahead" styles when "filter" is not
// defined in the theme.
func (f *Filter) Apply(theme *Theme) {
	theme.Apply(f, f.Selector("typeahead"), "disabled", "focused", "hovered")
	theme.Apply(f, f.Selector("typeahead/suggestion"), "focused")
	theme.Apply(f, f.Selector("filter"), "disabled", "focused", "hovered")
}

// Bind sets the bound widget. If w also implements [Suggester] its Suggest
// method is wired as the ghost-text provider. applyFilter is called immediately
// so the bound widget's visible content reflects the current input text.
func (f *Filter) Bind(w Filterable) {
	f.bound = w
	if s, ok := w.(Suggester); ok {
		f.SetSuggest(s.Suggest)
	} else {
		f.SetSuggest(nil)
	}
	f.applyFilter()
}

// Unbind clears the binding: resets the bound widget's filter, detaches the
// suggest function, and sets bound to nil. No-op when nothing is bound.
func (f *Filter) Unbind() {
	if f.bound == nil {
		return
	}
	f.bound.Filter("")
	f.SetSuggest(nil)
	f.bound = nil
}

// Bound returns the currently bound widget, or nil if nothing is bound.
func (f *Filter) Bound() Filterable {
	return f.bound
}

// Clear clears the input text. The bound widget's filter is reset automatically
// via the EvtChange → applyFilter chain.
func (f *Filter) Clear() {
	f.Input.Clear()
}

func (f *Filter) applyFilter() {
	if f.bound == nil {
		return
	}
	f.bound.Filter(f.Get())
}

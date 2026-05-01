package widgets

import (
	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ==== AI ===================================================================

// Typeahead is a single-line text input that shows an inline ghost-text
// suggestion after the cursor. The suggestion is provided by a caller-supplied
// callback and accepted with Tab or → at end-of-text.
type Typeahead struct {
	Input
	suggest    func(string) []string
	suggestion string
}

// NewTypeahead creates a new Typeahead widget. Params are identical to
// NewInput: [0] initial text, [1] placeholder, [2] mask character.
func NewTypeahead(id, class string, params ...string) *Typeahead {
	var text, placeholder, mask string
	if len(params) > 0 {
		text = params[0]
	}
	if len(params) > 1 {
		placeholder = params[1]
	}
	if len(params) > 2 {
		mask = params[2]
	} else {
		mask = "*"
	}

	t := &Typeahead{}

	// Initialise the embedded Input fields directly so that method values
	// registered below are bound to &t.Input, not to a separate heap copy.
	t.Component = Component{id: id, class: class, hheight: 1}
	t.buf = NewGapBufferFromString(text, 16)
	t.placeholder = placeholder
	t.mask = mask
	t.SetFlag(FlagFocusable, true)
	t.SetFlag(FlagMasked, false)
	t.SetFlag(FlagReadonly, false)
	t.refresh = func() { Redraw(t) }

	// Input's handler runs last; Typeahead's handler is prepended next.
	OnKey(t, t.Input.handleKey)
	OnKey(t, t.handleKey)

	// Update suggestion after every text change, including SetText.
	OnChange(t, func(value string) bool {
		t.updateSuggestion(value)
		return false
	})
	return t
}

// ---- Widget methods -------------------------------------------------------

// Apply applies theme styles for the typeahead widget.
func (t *Typeahead) Apply(theme *Theme) {
	theme.Apply(t, t.Selector("typeahead"), "disabled", "focused", "hovered")
	theme.Apply(t, t.Selector("typeahead/suggestion"), "focused")
}

// Refresh queues a redraw for the typeahead.
func (t *Typeahead) Refresh() {
	Redraw(t)
}

// Render draws the input text and, when a suggestion is active, the ghost-text
// suffix after the cursor.
func (t *Typeahead) Render(r *Renderer) {
	t.Component.Render(r)
	t.Input.Render(r)

	if t.suggestion == "" || t.Flag(FlagMasked) {
		return
	}

	text := t.Get()
	if len([]rune(t.suggestion)) <= len([]rune(text)) {
		return
	}
	suffix := string([]rune(t.suggestion)[len([]rune(text)):])

	cx, cy, cw, _ := t.Content()
	cursorX, _, _ := t.Cursor()
	ghostX := cursorX
	availW := cw - ghostX
	if availW <= 0 {
		return
	}

	state := ""
	if t.State() != "" {
		state = ":" + t.State()
	}
	style := t.Style("suggestion" + state)
	r.Set(style.Foreground(), style.Background(), style.Font())
	r.Text(cx+ghostX, cy, suffix, availW)
}

// SetSuggest sets the suggestion provider. The function receives the current
// input text and returns candidate strings. Returning nil or an empty slice
// clears the ghost text.
func (t *Typeahead) SetSuggest(fn func(string) []string) {
	t.suggest = fn
}

// ---- Internal methods -----------------------------------------------------

// accept completes the suggestion: sets input text, moves cursor to end,
// clears suggestion, and dispatches EvtAccept.
func (t *Typeahead) accept() {
	accepted := t.suggestion
	t.suggestion = ""
	t.Set(accepted)
	t.End()
	t.Dispatch(t, EvtAccept, accepted)
}

// handleKey intercepts Tab and → for suggestion acceptance, and Esc to clear.
// It runs before Input's handler because OnKey prepends.
func (t *Typeahead) handleKey(evt *tcell.EventKey) bool {
	switch evt.Key() {
	case tcell.KeyTab:
		if t.suggestion != "" {
			t.accept()
			return true
		}
		return false
	case tcell.KeyRight:
		text := t.Get()
		if t.suggestion != "" && t.pos == len([]rune(text)) {
			t.accept()
			return true
		}
		return false
	case tcell.KeyEscape:
		if t.suggestion != "" {
			t.suggestion = ""
			t.Refresh()
		}
		return false
	}
	return false
}

// updateSuggestion updates the ghost-text based on the current input text.
func (t *Typeahead) updateSuggestion(text string) {
	if t.suggest == nil {
		t.suggestion = ""
		t.Refresh()
		return
	}
	candidates := t.suggest(text)
	t.suggestion = ""
	for _, c := range candidates {
		if c != text {
			t.suggestion = c
			break
		}
	}
	t.Refresh()
}

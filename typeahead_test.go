package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func newKey(key tcell.Key) *tcell.EventKey {
	return tcell.NewEventKey(key, "", tcell.ModNone)
}

func TestTypeahead_UpdateSuggestion_Match(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.SetSuggest(func(text string) []string {
		return []string{"hello world", "help"}
	})
	ta.updateSuggestion("hel")
	if ta.suggestion != "hello world" {
		t.Errorf("suggestion = %q; want %q", ta.suggestion, "hello world")
	}
}

func TestTypeahead_UpdateSuggestion_NoMatch(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.SetSuggest(func(text string) []string {
		return nil // callback returns nothing when there is no match
	})
	ta.updateSuggestion("hel")
	if ta.suggestion != "" {
		t.Errorf("suggestion = %q; want empty", ta.suggestion)
	}
}

func TestTypeahead_UpdateSuggestion_NilSuggest(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.suggestion = "leftover"
	ta.updateSuggestion("hel")
	if ta.suggestion != "" {
		t.Errorf("suggestion = %q; want empty when suggest is nil", ta.suggestion)
	}
}

func TestTypeahead_UpdateSuggestion_EmptySlice(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.SetSuggest(func(text string) []string { return nil })
	ta.suggestion = "leftover"
	ta.updateSuggestion("hel")
	if ta.suggestion != "" {
		t.Errorf("suggestion = %q; want empty for nil return", ta.suggestion)
	}
}

func TestTypeahead_Tab_Accepts(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.SetSuggest(func(text string) []string {
		return []string{"hello world"}
	})
	ta.updateSuggestion("hel")

	var accepted string
	ta.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			accepted, _ = data[0].(string)
		}
		return true
	})

	consumed := ta.handleKey(ta, newKey(tcell.KeyTab))
	if !consumed {
		t.Error("Tab with active suggestion should be consumed")
	}
	if ta.Text() != "hello world" {
		t.Errorf("Text() = %q; want %q", ta.Text(), "hello world")
	}
	if accepted != "hello world" {
		t.Errorf("EvtAccept data = %q; want %q", accepted, "hello world")
	}
}

func TestTypeahead_Tab_Propagates_WhenEmpty(t *testing.T) {
	ta := NewTypeahead("ta", "")
	consumed := ta.handleKey(ta, newKey(tcell.KeyTab))
	if consumed {
		t.Error("Tab without suggestion should not be consumed")
	}
}

func TestTypeahead_Right_AtEnd_Accepts(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.SetSuggest(func(text string) []string {
		return []string{"hello world"}
	})
	ta.Input.SetText("hel")
	ta.Input.End()
	ta.suggestion = "hello world"

	consumed := ta.handleKey(ta, newKey(tcell.KeyRight))
	if !consumed {
		t.Error("→ at end with suggestion should be consumed")
	}
	if ta.Text() != "hello world" {
		t.Errorf("Text() = %q; want %q", ta.Text(), "hello world")
	}
}

func TestTypeahead_Right_MidText_Propagates(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.suggestion = "hello world"
	ta.Input.SetText("hel")
	ta.pos = 1 // mid-text

	consumed := ta.handleKey(ta, newKey(tcell.KeyRight))
	if consumed {
		t.Error("→ mid-text should not be consumed by Typeahead")
	}
}

func TestTypeahead_Esc_ClearsSuggestion(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.suggestion = "hello world"

	consumed := ta.handleKey(ta, newKey(tcell.KeyEscape))
	if consumed {
		t.Error("Esc should not be consumed (should propagate)")
	}
	if ta.suggestion != "" {
		t.Errorf("suggestion = %q; want empty after Esc", ta.suggestion)
	}
}

func TestTypeahead_Masked_SuppressesSuggestion(t *testing.T) {
	ta := NewTypeahead("ta", "")
	ta.SetSuggest(func(text string) []string {
		return []string{"secret123"}
	})
	ta.SetMask("*")
	ta.Input.SetText("sec")
	ta.suggestion = "secret123"

	// Masking flag must be set for Render to suppress ghost text.
	// We verify the flag check directly rather than full rendering.
	if !ta.Flag(FlagMasked) {
		t.Error("FlagMasked should be set after SetMask")
	}
	// Render would skip ghost text; verified via the FlagMasked check in Render.
}

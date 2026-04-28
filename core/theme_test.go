package core

import "testing"

// ── Color ─────────────────────────────────────────────────────────────────────

func TestTheme_Color_DirectPassthrough(t *testing.T) {
	theme := NewTheme()
	got := theme.Color("#ff0000")
	if got != "#ff0000" {
		t.Errorf("Color(\"#ff0000\") = %q; want unchanged", got)
	}
}

func TestTheme_Color_VariableResolution(t *testing.T) {
	theme := NewTheme()
	theme.SetColors(map[string]string{"$primary": "#007acc"})
	got := theme.Color("$primary")
	if got != "#007acc" {
		t.Errorf("Color(\"$primary\") = %q; want %q", got, "#007acc")
	}
}

func TestTheme_Color_UnknownVariable_Passthrough(t *testing.T) {
	theme := NewTheme()
	got := theme.Color("$unknown")
	if got != "$unknown" {
		t.Errorf("Color(\"$unknown\") = %q; want %q (unresolved returned as-is)", got, "$unknown")
	}
}

// ── String ────────────────────────────────────────────────────────────────────

func TestTheme_String_Found(t *testing.T) {
	theme := NewTheme()
	// "collapsible.expanded" is set in NewTheme
	got := theme.String("collapsible.expanded")
	if got == "" {
		t.Error("String(\"collapsible.expanded\") should return a non-empty default")
	}
}

func TestTheme_String_Missing_ReturnsEmpty(t *testing.T) {
	theme := NewTheme()
	got := theme.String("nonexistent.key")
	if got != "" {
		t.Errorf("String(\"nonexistent.key\") = %q; want empty", got)
	}
}

func TestTheme_SetStrings_OverridesDefault(t *testing.T) {
	theme := NewTheme()
	theme.SetStrings(map[string]string{"mykey": "myval"})
	if theme.String("mykey") != "myval" {
		t.Errorf("String(\"mykey\") = %q; want %q", theme.String("mykey"), "myval")
	}
}

// ── Flag ─────────────────────────────────────────────────────────────────────

func TestTheme_Flag_Missing_ReturnsFalse(t *testing.T) {
	theme := NewTheme()
	if theme.Flag("nonexistent") {
		t.Error("Flag(\"nonexistent\") should return false")
	}
}

func TestTheme_Flag_Set_ReturnsTrue(t *testing.T) {
	theme := NewTheme()
	theme.SetFlags(map[string]bool{"feature.x": true})
	if !theme.Flag("feature.x") {
		t.Error("Flag(\"feature.x\") should return true after SetFlags")
	}
}

// ── Border ────────────────────────────────────────────────────────────────────

func TestTheme_Border_Registered(t *testing.T) {
	theme := NewTheme()
	theme.SetBorders(map[string]*Border{
		"thin": {},
	})
	b := theme.Border("thin")
	if b == nil {
		t.Error("Border(\"thin\") should not be nil after AddUnicodeBorders")
	}
}

func TestTheme_Border_Unregistered_ReturnsNil(t *testing.T) {
	theme := NewTheme()
	b := theme.Border("nonexistent")
	if b != nil {
		t.Errorf("Border(\"nonexistent\") = %v; want nil", b)
	}
}

// ── Get ───────────────────────────────────────────────────────────────────────

func TestTheme_Get_FindsStyle(t *testing.T) {
	theme := NewTheme()
	style := NewStyle("button").WithColors("white", "blue")
	theme.AddStyles(style)
	got := theme.Get("button")
	if got == &DefaultStyle {
		t.Error("Get(\"button\") should find the registered style, not DefaultStyle")
	}
}

func TestTheme_Get_Unknown_ReturnsDefaultStyle(t *testing.T) {
	theme := NewTheme()
	got := theme.Get("unknown-widget")
	if got != &DefaultStyle {
		t.Errorf("Get(unknown) should return &DefaultStyle; got %v", got)
	}
}

// ── SetColors ────────────────────────────────────────────────────────────────

func TestTheme_SetColors_ReplacesRegistry(t *testing.T) {
	theme := NewTheme()
	theme.SetColors(map[string]string{"$a": "red", "$b": "blue"})
	if theme.Color("$a") != "red" || theme.Color("$b") != "blue" {
		t.Error("SetColors should register the provided color variables")
	}
	// Colors() returns the map directly
	if len(theme.Colors()) != 2 {
		t.Errorf("Colors() len = %d; want 2", len(theme.Colors()))
	}
}

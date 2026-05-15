package designer

import (
	"reflect"

	"github.com/tekugo/zeichenwerk/widgets"
)

// RegisterDefaults populates d with every widget kind that ships with
// the zeichenwerk widgets package. Register panics on a mismatch
// (Type vs form.New() result) — that's a programming error in this
// table, not a runtime condition, so we surface it loudly.
//
// Open calls this internally; drivers using their own Designer can
// call RegisterDefaults explicitly and then add custom kinds via
// Register.
func RegisterDefaults(d *Designer) {
	reg := func(t reflect.Type, mk func() WidgetForm) {
		if err := d.Register(Kind{Type: t, Make: mk}); err != nil {
			panic(err)
		}
	}

	reg(reflect.TypeOf((*widgets.Static)(nil)),
		func() WidgetForm { return &widgets.StaticForm{} })
	reg(reflect.TypeOf((*widgets.Grid)(nil)),
		func() WidgetForm { return &widgets.GridForm{} })
	reg(reflect.TypeOf((*widgets.Flex)(nil)),
		func() WidgetForm { return &widgets.FlexForm{} })
	reg(reflect.TypeOf((*widgets.Input)(nil)),
		func() WidgetForm { return &widgets.InputForm{} })
	reg(reflect.TypeOf((*widgets.Button)(nil)),
		func() WidgetForm { return &widgets.ButtonForm{} })
	reg(reflect.TypeOf((*widgets.Box)(nil)),
		func() WidgetForm { return &widgets.BoxForm{} })
	reg(reflect.TypeOf((*widgets.Card)(nil)),
		func() WidgetForm { return &widgets.CardForm{} })
	reg(reflect.TypeOf((*widgets.Checkbox)(nil)),
		func() WidgetForm { return &widgets.CheckboxForm{} })
	reg(reflect.TypeOf((*widgets.List)(nil)),
		func() WidgetForm { return &widgets.ListForm{} })
	reg(reflect.TypeOf((*widgets.Breadcrumb)(nil)),
		func() WidgetForm { return &widgets.BreadcrumbForm{} })
	reg(reflect.TypeOf((*widgets.Clock)(nil)),
		func() WidgetForm { return &widgets.ClockForm{} })
	reg(reflect.TypeOf((*widgets.Collapsible)(nil)),
		func() WidgetForm { return &widgets.CollapsibleForm{} })
	reg(reflect.TypeOf((*widgets.Combo)(nil)),
		func() WidgetForm { return &widgets.ComboForm{} })
	reg(reflect.TypeOf((*widgets.Deck)(nil)),
		func() WidgetForm { return &widgets.DeckForm{} })
	reg(reflect.TypeOf((*widgets.Dialog)(nil)),
		func() WidgetForm { return &widgets.DialogForm{} })
	reg(reflect.TypeOf((*widgets.Digits)(nil)),
		func() WidgetForm { return &widgets.DigitsForm{} })
	reg(reflect.TypeOf((*widgets.Editor)(nil)),
		func() WidgetForm { return &widgets.EditorForm{} })
	reg(reflect.TypeOf((*widgets.Filter)(nil)),
		func() WidgetForm { return &widgets.FilterForm{} })
	reg(reflect.TypeOf((*widgets.Indicator)(nil)),
		func() WidgetForm { return &widgets.IndicatorForm{} })
	reg(reflect.TypeOf((*widgets.Marquee)(nil)),
		func() WidgetForm { return &widgets.MarqueeForm{} })
	reg(reflect.TypeOf((*widgets.Progress)(nil)),
		func() WidgetForm { return &widgets.ProgressForm{} })
	reg(reflect.TypeOf((*widgets.Radio)(nil)),
		func() WidgetForm { return &widgets.RadioForm{} })
	reg(reflect.TypeOf((*widgets.Slider)(nil)),
		func() WidgetForm { return &widgets.SliderForm{} })
	reg(reflect.TypeOf((*widgets.Rule)(nil)),
		func() WidgetForm { return &widgets.RuleForm{} })
	reg(reflect.TypeOf((*widgets.Scanner)(nil)),
		func() WidgetForm { return &widgets.ScannerForm{} })
	reg(reflect.TypeOf((*widgets.Select)(nil)),
		func() WidgetForm { return &widgets.SelectForm{} })
	reg(reflect.TypeOf((*widgets.Shortcuts)(nil)),
		func() WidgetForm { return &widgets.ShortcutsForm{} })
	reg(reflect.TypeOf((*widgets.Spinner)(nil)),
		func() WidgetForm { return &widgets.SpinnerForm{} })
	reg(reflect.TypeOf((*widgets.Styled)(nil)),
		func() WidgetForm { return &widgets.StyledForm{} })
	reg(reflect.TypeOf((*widgets.Switcher)(nil)),
		func() WidgetForm { return &widgets.SwitcherForm{} })
	reg(reflect.TypeOf((*widgets.Table)(nil)),
		func() WidgetForm { return &widgets.TableForm{} })
	reg(reflect.TypeOf((*widgets.Tabs)(nil)),
		func() WidgetForm { return &widgets.TabsForm{} })
	reg(reflect.TypeOf((*widgets.Terminal)(nil)),
		func() WidgetForm { return &widgets.TerminalForm{} })
	reg(reflect.TypeOf((*widgets.Text)(nil)),
		func() WidgetForm { return &widgets.TextForm{} })
	reg(reflect.TypeOf((*widgets.Tiles)(nil)),
		func() WidgetForm { return &widgets.TilesForm{} })
	// TreeFS and TreeWidgets publish their inner *Tree to the
	// builder, so a single *Tree registration serves all three roles.
	// Specialised TreeFS / TreeWidgets editing happens through
	// dedicated dialogs.
	reg(reflect.TypeOf((*widgets.Tree)(nil)),
		func() WidgetForm { return &widgets.TreeForm{} })
	reg(reflect.TypeOf((*widgets.Typeahead)(nil)),
		func() WidgetForm { return &widgets.TypeaheadForm{} })
	reg(reflect.TypeOf((*widgets.Typewriter)(nil)),
		func() WidgetForm { return &widgets.TypewriterForm{} })
	reg(reflect.TypeOf((*widgets.Viewport)(nil)),
		func() WidgetForm { return &widgets.ViewportForm{} })
}

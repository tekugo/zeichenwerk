package zeichenwerk

type TabSwitcher struct {
	BaseWidget
	tabs     Tabs
	switcher Switcher
	mapping  map[string]string
}

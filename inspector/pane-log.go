package inspector

import (
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// mountLogPane builds the Log tab's Table widget on top of
// ui.Logs() (the framework's circular log buffer, which
// implements TableProvider) and slots it into the popup's
// pre-built log box.
//
// The table re-reads on each render, so no subscription is
// needed — new log entries appear the next time the popup is
// drawn.
func (s *session) mountLogPane() {
	logBox := core.MustFind[*widgets.Box](s.popup, "inspector-log-box")
	table := widgets.NewTable("inspector-log-table", "", s.ui.Logs(), false)
	table.SetHint(0, -1)
	table.Apply(s.theme)
	_ = logBox.Add(table)
}

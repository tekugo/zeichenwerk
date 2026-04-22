package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/tekugo/zeichenwerk/widgets"
)

// ---- metricsProvider -------------------------------------------------------

// metricsProvider implements z.TableProvider for Session.Metrics, newest-first.
type metricsProvider struct {
	mu      sync.RWMutex
	session *Session
}

func (p *metricsProvider) set(s *Session) {
	p.mu.Lock()
	p.session = s
	p.mu.Unlock()
}

var metricsColumns = []widgets.TableColumn{
	{Header: "Time", Width: 8},
	{Header: "Start", Width: 8},
	{Header: "Model", Width: 22},
	{Header: "In", Width: 8, Alignment: widgets.AlignRight},
	{Header: "Out", Width: 8, Alignment: widgets.AlignRight},
	{Header: "CacheR", Width: 8, Alignment: widgets.AlignRight},
	{Header: "CacheC", Width: 8, Alignment: widgets.AlignRight},
	{Header: "Cost", Width: 10, Alignment: widgets.AlignRight},
	{Header: "UserT", Width: 8, Alignment: widgets.AlignRight},
	{Header: "CLIT", Width: 8, Alignment: widgets.AlignRight},
	{Header: "+Lines", Width: 7, Alignment: widgets.AlignRight},
	{Header: "-Lines", Width: 7, Alignment: widgets.AlignRight},
	{Header: "Accept", Width: 6, Alignment: widgets.AlignRight},
	{Header: "Reject", Width: 6, Alignment: widgets.AlignRight},
}

func (p *metricsProvider) Columns() []widgets.TableColumn { return metricsColumns }

func (p *metricsProvider) Length() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.session == nil {
		return 0
	}
	return len(p.session.Metrics)
}

func (p *metricsProvider) Str(row, col int) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.session == nil {
		return ""
	}
	n := len(p.session.Metrics)
	if row < 0 || row >= n {
		return ""
	}
	m := p.session.Metrics[n-1-row] // newest-first
	switch col {
	case 0:
		return m.Time.Format("15:04:05")
	case 1:
		return m.Start.Format("15:04:05")
	case 2:
		return m.Model
	case 3:
		return fmt.Sprintf("%d", m.Input)
	case 4:
		return fmt.Sprintf("%d", m.Output)
	case 5:
		return fmt.Sprintf("%d", m.CacheRead)
	case 6:
		return fmt.Sprintf("%d", m.CacheCreation)
	case 7:
		return fmt.Sprintf("$%.6f", m.Cost)
	case 8:
		return fmt.Sprintf("%.1fs", m.ActiveUser)
	case 9:
		return fmt.Sprintf("%.1fs", m.ActiveCLI)
	case 10:
		return fmt.Sprintf("+%d", m.LinesAdded)
	case 11:
		return fmt.Sprintf("-%d", m.LinesRemoved)
	case 12:
		return fmt.Sprintf("%d", m.Accepted)
	case 13:
		return fmt.Sprintf("%d", m.Rejected)
	}
	return ""
}

// ---- logProvider -----------------------------------------------------------

// logProvider implements z.TableProvider for Session.Log, newest-first.
type logProvider struct {
	mu      sync.RWMutex
	session *Session
}

func (p *logProvider) set(s *Session) {
	p.mu.Lock()
	p.session = s
	p.mu.Unlock()
}

var logColumns = []widgets.TableColumn{
	{Header: "Time", Width: 8},
	{Header: "Body", Width: 32},
	{Header: "Attrs", Width: 60},
}

func (p *logProvider) Columns() []widgets.TableColumn { return logColumns }

func (p *logProvider) Length() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.session == nil {
		return 0
	}
	return len(p.session.Log)
}

func (p *logProvider) Str(row, col int) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.session == nil {
		return ""
	}
	n := len(p.session.Log)
	if row < 0 || row >= n {
		return ""
	}
	l := p.session.Log[n-1-row] // newest-first
	switch col {
	case 0:
		return l.Time.Format("15:04:05")
	case 1:
		return l.Body
	case 2:
		keys := make([]string, 0, len(l.Attrs))
		for k := range l.Attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pairs := make([]string, 0, len(keys))
		for _, k := range keys {
			pairs = append(pairs, k+"="+l.Attrs[k])
		}
		return strings.Join(pairs, "  ")
	}
	return ""
}

package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	z "github.com/tekugo/zeichenwerk"
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

var metricsColumns = []z.TableColumn{
	{Header: "Time", Width: 8},
	{Header: "Start", Width: 8},
	{Header: "Model", Width: 22},
	{Header: "In", Width: 8, Alignment: z.AlignRight},
	{Header: "Out", Width: 8, Alignment: z.AlignRight},
	{Header: "CacheR", Width: 8, Alignment: z.AlignRight},
	{Header: "CacheC", Width: 8, Alignment: z.AlignRight},
	{Header: "Cost", Width: 10, Alignment: z.AlignRight},
	{Header: "UserT", Width: 8, Alignment: z.AlignRight},
	{Header: "CLIT", Width: 8, Alignment: z.AlignRight},
	{Header: "+Lines", Width: 7, Alignment: z.AlignRight},
	{Header: "-Lines", Width: 7, Alignment: z.AlignRight},
	{Header: "Accept", Width: 6, Alignment: z.AlignRight},
	{Header: "Reject", Width: 6, Alignment: z.AlignRight},
}

func (p *metricsProvider) Columns() []z.TableColumn { return metricsColumns }

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

var logColumns = []z.TableColumn{
	{Header: "Time", Width: 8},
	{Header: "Body", Width: 32},
	{Header: "Attrs", Width: 60},
}

func (p *logProvider) Columns() []z.TableColumn { return logColumns }

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

// ---- modelProvider ---------------------------------------------------------

// modelProvider implements z.TableProvider for per-model totals.
// The last row is always an aggregate "Total" row.
type modelProvider struct {
	mu      sync.RWMutex
	session *Session
}

func (p *modelProvider) set(s *Session) {
	p.mu.Lock()
	p.session = s
	p.mu.Unlock()
}

// sortedModels returns non-empty model keys in alphabetical order.
// Must be called with p.mu held.
func (p *modelProvider) sortedModels() []string {
	names := make([]string, 0, len(p.session.Totals))
	for k := range p.session.Totals {
		if k != "" {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	return names
}

var modelColumns = []z.TableColumn{
	{Header: "Model", Width: 26},
	{Header: "Input", Width: 8, Alignment: z.AlignRight},
	{Header: "Output", Width: 8, Alignment: z.AlignRight},
	{Header: "CacheR", Width: 8, Alignment: z.AlignRight},
	{Header: "CacheC", Width: 8, Alignment: z.AlignRight},
	{Header: "Cost", Width: 12, Alignment: z.AlignRight},
	{Header: "+Lines", Width: 7, Alignment: z.AlignRight},
	{Header: "-Lines", Width: 7, Alignment: z.AlignRight},
	{Header: "Accept", Width: 7, Alignment: z.AlignRight},
	{Header: "Reject", Width: 7, Alignment: z.AlignRight},
}

func (p *modelProvider) Columns() []z.TableColumn { return modelColumns }

func (p *modelProvider) Length() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.session == nil {
		return 0
	}
	n := 0
	for k := range p.session.Totals {
		if k != "" {
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return n + 1 // model rows + total row
}

func (p *modelProvider) Str(row, col int) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.session == nil {
		return ""
	}
	models := p.sortedModels()
	n := len(models)
	if n == 0 {
		return ""
	}

	var m *Metrics
	var label string
	if row == n {
		// Total row: aggregate across all models.
		var tot Metrics
		for _, t := range p.session.Totals {
			tot.Add(t)
		}
		m = &tot
		label = "Total"
	} else {
		if row < 0 || row >= n {
			return ""
		}
		m = p.session.Totals[models[row]]
		label = models[row]
	}

	switch col {
	case 0:
		return truncate(label, 26)
	case 1:
		return formatTokens(m.Input)
	case 2:
		return formatTokens(m.Output)
	case 3:
		return formatTokens(m.CacheRead)
	case 4:
		return formatTokens(m.CacheCreation)
	case 5:
		return fmt.Sprintf("$%.4f", m.Cost)
	case 6:
		if m.LinesAdded > 0 {
			return fmt.Sprintf("+%d", m.LinesAdded)
		}
		return ""
	case 7:
		if m.LinesRemoved > 0 {
			return fmt.Sprintf("-%d", m.LinesRemoved)
		}
		return ""
	case 8:
		if m.Accepted > 0 {
			return fmt.Sprintf("%d", m.Accepted)
		}
		return ""
	case 9:
		if m.Rejected > 0 {
			return fmt.Sprintf("%d", m.Rejected)
		}
		return ""
	}
	return ""
}

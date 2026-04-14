package main

import (
	"sync"
	"time"
)

// Values holds the aggregated metric data for a model.
type Metrics struct {
	// Time stamps
	Time  time.Time // ts
	Start time.Time // start_ts

	// Model name (attrs:model)
	Model string

	// Cost (claude_code.cost.usage)
	Cost float64 // [USD]

	// Token usage (claude_code.token.usage)
	Input         int64 // [tokens] attrs:type input
	Output        int64 // [tokens] attrs:type output
	CacheRead     int64 // [tokens] attrs:type cacheRead
	CacheCreation int64 // [tokens] attrs:type cacheCreation

	// Active times (claude_code.active_time.total)
	ActiveUser float64 // [s] attrs:type user
	ActiveCLI  float64 // [s] attrs:type cli

	// Line Count (claude_code.lines_of_code.count)
	LinesAdded   int64 // [] attrs:type added
	LinesRemoved int64 // [] attrs:type removed

	// Decisions (claude_code.code_edit_tool.decision)
	Accepted int64
	Rejected int64
}

func (m *Metrics) Add(other *Metrics) {
	m.Cost += other.Cost
	m.Input += other.Input
	m.Output += other.Output
	m.CacheRead += other.CacheRead
	m.CacheCreation += other.CacheCreation
	m.ActiveUser += other.ActiveUser
	m.ActiveCLI += other.ActiveCLI
	m.LinesAdded += other.LinesAdded
	m.LinesRemoved += other.LinesRemoved
	m.Accepted += other.Accepted
	m.Rejected += other.Rejected
}

func (m *Metrics) Clear() {
	m.Cost = 0
	m.Input = 0
	m.Output = 0
	m.CacheRead = 0
	m.CacheCreation = 0
	m.ActiveUser = 0
	m.ActiveCLI = 0
	m.LinesAdded = 0
	m.LinesRemoved = 0
	m.Accepted = 0
	m.Rejected = 0
}

type SessionStatus int

const (
	StatusActive SessionStatus = iota
	StatusIdle
	StatusEnded
	StatusTimeOut
)

type Session struct {
	Start  time.Time // start time (first seen)
	End    time.Time // only if ended or timed-out
	Status string    // session status

	// resource
	HostArch       string // host.arch
	OSType         string // os.type
	OSVersion      string // os.version
	ServiceName    string // service.name
	ServiceVersion string // service.Version

	// attrs
	// model will be in the metrics
	OrgID           string // organization.id
	ID              string // session.id
	TerminalType    string // terminal.type
	UserAccountID   string // user.account_id
	UserAccountUUID string // user.account_uuid
	UserEmail       string // user.email
	UserID          string // user.id

	Metrics []Metrics
	Log     []Log

	Totals map[string]*Metrics // Session totals per model
}

// Add accumulates m into the matching (Time, Model) entry, creating one if needed.
// Also updates the per-model Totals. Returns true if an existing entry was found.
func (s *Session) Add(m *Metrics) bool {
	total, ok := s.Totals[m.Model]
	if !ok {
		total = &Metrics{Model: m.Model}
		s.Totals[m.Model] = total
	}
	total.Add(m)

	if existing := s.Find(m.Time, m.Model); existing != nil {
		existing.Add(m)
		return true
	}
	entry := Metrics{
		Time:  m.Time,
		Start: m.Start,
		Model: m.Model,
	}
	entry.Add(m)

	s.Metrics = append(s.Metrics, entry)
	return false
}

// Find returns the Metrics entry whose time bucket (truncated to 1 second)
// and model match ts and model, or nil.
// 1-second bucketing groups all data points from the same OTLP export batch
// together, even when individual data points carry slightly different nanosecond
// timestamps.
func (s *Session) Find(ts time.Time, model string) *Metrics {
	bucket := ts.Truncate(time.Second)
	for i := range s.Metrics {
		if s.Metrics[i].Time.Truncate(time.Second).Equal(bucket) && s.Metrics[i].Model == model {
			return &s.Metrics[i]
		}
	}
	return nil
}

type Log struct {
	Time     time.Time         // ts
	Observed time.Time         // observed_ts
	Severity string            // severity
	Body     string            // body
	TraceID  string            // trace_id
	Attrs    map[string]string // attrs
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	order    []string
	onChange func()
}

func NewStore() *Store {
	return &Store{sessions: make(map[string]*Session)}
}

// Find returns the existing session with the given ID, or nil.
func (st *Store) Find(id string) *Session {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.sessions[id]
}

// Get returns the session with the given ID. If none exists, a new empty
// session is created, registered, and returned with isNew = true.
func (st *Store) Get(id string) (s *Session, isNew bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if s = st.sessions[id]; s != nil {
		return s, false
	}
	s = &Session{Totals: make(map[string]*Metrics)}
	st.sessions[id] = s
	st.order = append(st.order, id)
	return s, true
}

// Items returns sessions in insertion order, newest-first.
func (st *Store) Items() []*Session {
	st.mu.RLock()
	defer st.mu.RUnlock()
	n := len(st.order)
	out := make([]*Session, n)
	for i, id := range st.order {
		out[n-1-i] = st.sessions[id]
	}
	return out
}

// SetOnChange registers fn to be called (without holding any lock) whenever
// new data arrives. Pass nil to remove a previously registered callback.
func (st *Store) SetOnChange(fn func()) {
	st.mu.Lock()
	st.onChange = fn
	st.mu.Unlock()
}

// Notify calls the onChange callback, if one is registered.
func (st *Store) Notify() {
	st.mu.RLock()
	fn := st.onChange
	st.mu.RUnlock()
	if fn != nil {
		fn()
	}
}

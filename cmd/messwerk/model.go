package main

import (
	"sync"
	"time"

	z "github.com/tekugo/zeichenwerk"
)

// Create a new TimeSeries for the last hour with 2 minute intervals
func newTS() *z.TimeSeries[float64] {
	start := time.Now().Add(-time.Duration(29) * time.Minute * 2)
	return z.NewTimeSeries[float64](start, time.Minute*2, 30, true)
}

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

	// Deck sparkline series: 30 × 2 min = 60 min
	Input  *z.TimeSeries[float64]
	Output *z.TimeSeries[float64]
	Cost   *z.TimeSeries[float64]
}

// LastSeen returns the timestamp of the most recently received metrics entry,
// or Start if no metrics have been received yet.
func (s *Session) LastSeen() time.Time {
	if len(s.Metrics) == 0 {
		return s.Start
	}
	return s.Metrics[len(s.Metrics)-1].Time
}

// Aggregate returns a single Metrics with all per-model totals summed together.
func (s *Session) Aggregate() Metrics {
	var result Metrics
	for _, m := range s.Totals {
		result.Add(m)
	}
	return result
}

// PrimaryModel returns the name of the model with the highest cost in this session,
// or an empty string if no metrics have been received.
func (s *Session) PrimaryModel() string {
	var best string
	var bestCost float64
	for model, m := range s.Totals {
		if m.Cost > bestCost {
			bestCost = m.Cost
			best = model
		}
	}
	return best
}

// TotalCost returns the sum of cost across all model totals for this session.
func (s *Session) TotalCost() float64 {
	var total float64
	for _, m := range s.Totals {
		total += m.Cost
	}
	return total
}

// LastInput returns the input token count from the most recent metrics entry,
// or 0 if no metrics have been received yet.
func (s *Session) LastInput() int64 {
	if len(s.Metrics) == 0 {
		return 0
	}
	return s.Metrics[len(s.Metrics)-1].Input
}

// LastOutput returns the output token count from the most recent metrics entry,
// or 0 if no metrics have been received yet.
func (s *Session) LastOutput() int64 {
	if len(s.Metrics) == 0 {
		return 0
	}
	return s.Metrics[len(s.Metrics)-1].Output
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

	// Total time lines for all sessions
	Input  *z.TimeSeries[float64]
	Output *z.TimeSeries[float64]
	Cost   *z.TimeSeries[float64]

	// Reactive running totals
	TotalCost   *z.Value[float64]
	TotalInput  *z.Value[int64]
	TotalOutput *z.Value[int64]

	totalCost   float64
	totalInput  int64
	totalOutput int64
}

func NewStore() *Store {
	return &Store{
		sessions:    make(map[string]*Session),
		Input:       newTS(),
		Output:      newTS(),
		Cost:        newTS(),
		TotalCost:   z.NewValue[float64](0),
		TotalInput:  z.NewValue[int64](0),
		TotalOutput: z.NewValue[int64](0),
	}
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
	s = &Session{
		Totals: make(map[string]*Metrics),
		Input:  newTS(),
		Output: newTS(),
		Cost:   newTS(),
	}
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

// addToSeries adds m's token and cost values to the session's deck-scale TimeSeries.
func (s *Session) addToSeries(m *Metrics) {
	s.Input.Add(m.Time, float64(m.Input))
	s.Output.Add(m.Time, float64(m.Output))
	s.Cost.Add(m.Time, m.Cost)
}

// addToTotal adds m's token and cost values to the store's overview-scale TimeSeries.
func (st *Store) addToTotal(m *Metrics) {
	st.Input.Add(m.Time, float64(m.Input))
	st.Output.Add(m.Time, float64(m.Output))
	st.Cost.Add(m.Time, m.Cost)
	st.totalInput += m.Input
	st.totalOutput += m.Output
	st.totalCost += m.Cost
	st.TotalInput.Set(st.totalInput)
	st.TotalOutput.Set(st.totalOutput)
	st.TotalCost.Set(st.totalCost)
}

// LastSeen returns the most recent LastSeen timestamp across all sessions,
// or the zero time if the store is empty.
func (st *Store) LastSeen() time.Time {
	st.mu.RLock()
	defer st.mu.RUnlock()
	var t time.Time
	for _, s := range st.sessions {
		if ls := s.LastSeen(); ls.After(t) {
			t = ls
		}
	}
	return t
}

// TouchAll advances all session and store TimeSeries to now and triggers a UI refresh.
func (st *Store) TouchAll(now time.Time) {
	st.Input.Touch(now)
	st.Output.Touch(now)
	st.Cost.Touch(now)
	st.mu.RLock()
	for _, s := range st.sessions {
		s.Input.Touch(now)
		s.Output.Touch(now)
		s.Cost.Touch(now)
	}
	st.mu.RUnlock()
	st.Notify()
}

package main

import (
	"sort"
	"sync"
	"time"

	. "github.com/tekugo/zeichenwerk"
)

// SessionStatus represents a session's current activity level.
type SessionStatus int

const (
	StatusActive SessionStatus = iota // data within the last idleTimeout/2
	StatusIdle                        // no data for idleTimeout/2 … idleTimeout
	StatusEnded                       // no data for longer than idleTimeout
)

// ── MetricValues ─────────────────────────────────────────────────────────────

// MetricValues holds measured quantities for a time window.
// Used for raw data points and all aggregation granularities.
type MetricValues struct {
	// claude_code.active_time.total
	ActiveTimeUser float64 // seconds, type=user
	ActiveTimeCLI  float64 // seconds, type=cli

	// claude_code.token.usage
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64

	// claude_code.cost.usage
	CostUSD float64

	// claude_code.lines_of_code.count
	LinesAdded   int64
	LinesRemoved int64

	// claude_code.code_edit_tool.decision
	EditDecisionsAccepted int64
	EditDecisionsRejected int64

	// claude_code.session.count
	SessionCount int64
}

func addValues(dst *MetricValues, src MetricValues) {
	dst.ActiveTimeUser += src.ActiveTimeUser
	dst.ActiveTimeCLI += src.ActiveTimeCLI
	dst.InputTokens += src.InputTokens
	dst.OutputTokens += src.OutputTokens
	dst.CacheReadTokens += src.CacheReadTokens
	dst.CacheCreationTokens += src.CacheCreationTokens
	dst.CostUSD += src.CostUSD
	dst.LinesAdded += src.LinesAdded
	dst.LinesRemoved += src.LinesRemoved
	dst.EditDecisionsAccepted += src.EditDecisionsAccepted
	dst.EditDecisionsRejected += src.EditDecisionsRejected
	dst.SessionCount += src.SessionCount
}

// ── Metric / Bucket ──────────────────────────────────────────────────────────

// Metric is a flat representation of one incoming OTLP data point.
// Session identity is not stored here; it lives in Session.
type Metric struct {
	Timestamp      time.Time
	StartTimestamp time.Time
	Name           string

	// Discriminator attributes used when parsing to route the value into
	// the correct MetricValues field.
	Model    string // token.usage, cost.usage
	Decision string // code_edit_tool.decision: "accept" | "reject"
	Language string // code_edit_tool.decision
	Source   string // code_edit_tool.decision
	ToolName string // code_edit_tool.decision

	MetricValues // embedded — only one or two fields non-zero per raw point
}

// Bucket is a time-windowed aggregate of MetricValues.
type Bucket struct {
	Start time.Time
	End   time.Time
	MetricValues
}

const bucketWidth = 2 * time.Minute
const maxBuckets = 20

func mergeToBuckets(buckets []Bucket, m Metric, width time.Duration) []Bucket {
	start := m.Timestamp.Truncate(width)
	end := start.Add(width)
	for i := range buckets {
		if buckets[i].Start.Equal(start) {
			addValues(&buckets[i].MetricValues, m.MetricValues)
			return buckets
		}
	}
	buckets = append(buckets, Bucket{Start: start, End: end, MetricValues: m.MetricValues})
	if len(buckets) > maxBuckets {
		buckets = buckets[len(buckets)-maxBuckets:]
	}
	return buckets
}

// ── session (internal) ───────────────────────────────────────────────────────

// SessionInfo carries the per-session identity fields extracted from OTLP
// resource/metric attributes.
type SessionInfo struct {
	Name         string
	OrgID        string
	TerminalType string
	UserEmail    string
}

type session struct {
	id           string
	name         string
	orgID        string
	terminalType string
	userEmail    string

	firstSeen time.Time
	lastSeen  time.Time

	mu          sync.Mutex
	total       MetricValues // cumulative lifetime totals
	last        MetricValues // values from the most recent metric window
	lastStartTS time.Time    // StartTimestamp of the most recent window
	metrics     []Metric     // all raw data points (enables bucket reconstruction)
	twoMinute   []Bucket     // 2-minute aggregates for sparklines
}

func (s *session) status(timeout time.Duration) SessionStatus {
	since := time.Since(s.lastSeen)
	if since < timeout/2 {
		return StatusActive
	}
	if since < timeout {
		return StatusIdle
	}
	return StatusEnded
}

func (s *session) addMetric(m Metric) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.firstSeen.IsZero() {
		s.firstSeen = m.Timestamp
	}
	if m.Timestamp.After(s.lastSeen) {
		s.lastSeen = m.Timestamp
	}

	// Reset "last" whenever a new metric window starts.
	if !m.StartTimestamp.Equal(s.lastStartTS) {
		s.last = MetricValues{}
		s.lastStartTS = m.StartTimestamp
	}
	addValues(&s.last, m.MetricValues)
	addValues(&s.total, m.MetricValues)

	s.metrics = append(s.metrics, m)
	s.twoMinute = mergeToBuckets(s.twoMinute, m, bucketWidth)
}

// touchBucket ensures a zero-value bucket exists for now, advancing the
// visible window even when no telemetry is arriving.
func (s *session) touchBucket(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.twoMinute = mergeToBuckets(s.twoMinute, Metric{Timestamp: now}, bucketWidth)
}

// ── OTLPLog ──────────────────────────────────────────────────────────────────

// otlpEntry is one received OTLP data point, formatted for display.
type otlpEntry struct {
	time      time.Time
	session   string
	metric    string
	tokenType string
	value     string
}

// OTLPLog is a thread-safe circular buffer of incoming OTLP entries that
// implements TableProvider so it can be passed directly to a Table widget.
type OTLPLog struct {
	mu      sync.Mutex
	entries []otlpEntry
	size    int
	start   int
	count   int
}

func newOTLPLog(size int) *OTLPLog { return &OTLPLog{entries: make([]otlpEntry, size), size: size} }

func (l *OTLPLog) add(e otlpEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	idx := (l.start + l.count) % l.size
	l.entries[idx] = e
	if l.count < l.size {
		l.count++
	} else {
		l.start = (l.start + 1) % l.size
	}
}

var otlpLogColumns = []TableColumn{
	{Header: "Time", Width: 8},
	{Header: "Session", Width: 24},
	{Header: "Metric", Width: 26},
	{Header: "Type", Width: 12},
	{Header: "Value", Width: 14},
}

func (l *OTLPLog) Columns() []TableColumn { return otlpLogColumns }

func (l *OTLPLog) Length() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.count
}

func (l *OTLPLog) Str(row, col int) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.entries[(l.start+l.count-row-1)%l.size] // newest first
	switch col {
	case 0:
		return e.time.Format("15:04:05")
	case 1:
		return e.session
	case 2:
		return e.metric
	case 3:
		return e.tokenType
	default:
		return e.value
	}
}

// ── Store ────────────────────────────────────────────────────────────────────

// Store is the in-memory session registry.
type Store struct {
	mu          sync.RWMutex
	sessions    map[string]*session
	order       []string
	idleTimeout time.Duration
	onChange    func()

	totalBuckets []Bucket // 2-minute global aggregates across all sessions

	Log *OTLPLog
}

func newStore(timeout time.Duration) *Store {
	return &Store{
		sessions:    make(map[string]*session),
		idleTimeout: timeout,
		Log:         newOTLPLog(500),
	}
}

// AddMetric records a metric data point for a session.
func (st *Store) AddMetric(sessionID string, info SessionInfo, m Metric) {
	st.mu.Lock()
	s, ok := st.sessions[sessionID]
	if !ok {
		s = &session{
			id:           sessionID,
			name:         info.Name,
			orgID:        info.OrgID,
			terminalType: info.TerminalType,
			userEmail:    info.UserEmail,
		}
		st.sessions[sessionID] = s
		st.order = append(st.order, sessionID)
	}
	onChange := st.onChange
	st.mu.Unlock()

	s.addMetric(m)

	st.mu.Lock()
	st.totalBuckets = mergeToBuckets(st.totalBuckets, m, bucketWidth)
	st.mu.Unlock()

	if onChange != nil {
		onChange()
	}
}

// Tick advances bucket windows for all sessions and the global total,
// ensuring sparklines scroll even when no telemetry is arriving.
func (st *Store) Tick(now time.Time) {
	st.mu.Lock()
	sessions := make([]*session, 0, len(st.sessions))
	for _, s := range st.sessions {
		sessions = append(sessions, s)
	}
	st.totalBuckets = mergeToBuckets(st.totalBuckets, Metric{Timestamp: now}, bucketWidth)
	onChange := st.onChange
	st.mu.Unlock()

	for _, s := range sessions {
		s.touchBucket(now)
	}

	if onChange != nil {
		onChange()
	}
}

// SetOnChange registers a function called after every data update.
func (st *Store) SetOnChange(fn func()) {
	st.mu.Lock()
	st.onChange = fn
	st.mu.Unlock()
}

// ── Snapshots ────────────────────────────────────────────────────────────────

// SessionItem is the read-only snapshot used by the deck and session detail view.
type SessionItem struct {
	ID           string
	Name         string
	OrgID        string
	TerminalType string
	UserEmail    string
	Status       SessionStatus
	FirstSeen    time.Time
	LastSeen     time.Time
	Total        MetricValues // cumulative lifetime totals
	Last         MetricValues // most recent metric window — for big-number display
	Buckets      []Bucket     // 2-min aggregates — UI builds sparklines from these
}

// SessionSummary is used in the total page session list.
type SessionSummary struct {
	ID        string
	Name      string
	Status    SessionStatus
	FirstSeen time.Time
	LastSeen  time.Time
}

// TotalView is the snapshot for the aggregate/total page.
type TotalView struct {
	Total    MetricValues
	Buckets  []Bucket
	Sessions []SessionSummary
}

// Items returns session snapshots (newest-first) and the global TotalView.
func (st *Store) Items() ([]*SessionItem, TotalView) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	type entry struct {
		s    *session
		seen time.Time
	}
	entries := make([]entry, 0, len(st.sessions))
	var totalValues MetricValues

	for _, s := range st.sessions {
		entries = append(entries, entry{s, s.lastSeen})
		s.mu.Lock()
		addValues(&totalValues, s.total)
		s.mu.Unlock()
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].seen.After(entries[j].seen)
	})

	items := make([]*SessionItem, len(entries))
	summaries := make([]SessionSummary, len(entries))
	for i, e := range entries {
		s := e.s
		s.mu.Lock()
		items[i] = &SessionItem{
			ID:           s.id,
			Name:         s.name,
			OrgID:        s.orgID,
			TerminalType: s.terminalType,
			UserEmail:    s.userEmail,
			Status:       s.status(st.idleTimeout),
			FirstSeen:    s.firstSeen,
			LastSeen:     s.lastSeen,
			Total:        s.total,
			Last:         s.last,
			Buckets:      append([]Bucket(nil), s.twoMinute...),
		}
		s.mu.Unlock()
		summaries[i] = SessionSummary{
			ID:        s.id,
			Name:      s.name,
			Status:    s.status(st.idleTimeout),
			FirstSeen: s.firstSeen,
			LastSeen:  s.lastSeen,
		}
	}

	tv := TotalView{
		Total:    totalValues,
		Buckets:  append([]Bucket(nil), st.totalBuckets...),
		Sessions: summaries,
	}
	return items, tv
}

package main

import "time"

// MetricValues holds measured quantities for a time window.
// Used for both raw data points and all aggregation granularities (2-min, hourly, daily, etc.).
// For a raw data point most fields are zero; for a bucket they hold summed totals.
type MetricValues struct {
	// claude_code.active_time.total
	ActiveTimeUser float64 // seconds, type=user
	ActiveTimeCLI  float64 // seconds, type=cli

	// claude_code.token.usage (summed across models for aggregates)
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

// Metric is a flat representation of one incoming OTLP data point.
// Session-identifying fields are not included here — they live in Session.
// Embeds MetricValues so measured values are directly accessible.
type Metric struct {
	Timestamp      time.Time
	StartTimestamp time.Time
	Name           string

	// Discriminator attributes — used when parsing to route the OTLP value
	// into the right MetricValues field.
	Model    string // token.usage, cost.usage — e.g. "claude-sonnet-4-6"
	Decision string // code_edit_tool.decision: "accept" | "reject"
	Language string // code_edit_tool.decision — e.g. "Markdown"
	Source   string // code_edit_tool.decision — e.g. "config"
	ToolName string // code_edit_tool.decision — e.g. "Write"

	MetricValues // embedded — only one or two fields non-zero per raw point
}

// Bucket is a time-windowed aggregate of MetricValues.
type Bucket struct {
	Start time.Time
	End   time.Time
	MetricValues
}

// Session tracks state for one Claude Code CLI session.
// Metrics are not stored directly; they are accumulated into Buckets and Total.
type Session struct {
	ID              string
	OrgID           string
	TerminalType    string
	UserAccountID   string
	UserAccountUUID string
	UserEmail       string
	UserID          string

	FirstSeen time.Time
	LastSeen  time.Time

	// Time-bucketed aggregates for different chart granularities
	TwoMinute []Bucket // sparkline — recent activity
	Hourly    []Bucket // day view
	Daily     []Bucket // week view
	Weekly    []Bucket // month view
	Monthly   []Bucket // long-term view

	// Cumulative totals for the entire session lifetime
	Total MetricValues

	// Raw data points in arrival order — buckets can be reconstructed from these.
	Metrics []Metric
}

func newSession(m Metric, sessionID, orgID, terminalType, userAccountID, userAccountUUID, userEmail, userID string) *Session {
	s := &Session{
		ID:              sessionID,
		OrgID:           orgID,
		TerminalType:    terminalType,
		UserAccountID:   userAccountID,
		UserAccountUUID: userAccountUUID,
		UserEmail:       userEmail,
		UserID:          userID,
		FirstSeen:       m.Timestamp,
		LastSeen:        m.Timestamp,
	}
	s.update(m)
	return s
}

func (s *Session) update(m Metric) {
	if m.Timestamp.After(s.LastSeen) {
		s.LastSeen = m.Timestamp
	}

	s.Metrics = append(s.Metrics, m)
	add(&s.Total, m.MetricValues)

	s.TwoMinute = addToBuckets(s.TwoMinute, m, 2*time.Minute)
	s.Hourly = addToBuckets(s.Hourly, m, time.Hour)
	s.Daily = addToBuckets(s.Daily, m, 24*time.Hour)
	s.Weekly = addToBuckets(s.Weekly, m, 7*24*time.Hour)
	s.Monthly = addToBuckets(s.Monthly, m, 30*24*time.Hour)
}

// add accumulates src into dst field by field.
func add(dst *MetricValues, src MetricValues) {
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

// addToBuckets adds m's values into the bucket covering m.Timestamp, creating one if needed.
func addToBuckets(buckets []Bucket, m Metric, width time.Duration) []Bucket {
	start := m.Timestamp.Truncate(width)
	end := start.Add(width)

	for i := range buckets {
		if buckets[i].Start.Equal(start) {
			add(&buckets[i].MetricValues, m.MetricValues)
			return buckets
		}
	}

	buckets = append(buckets, Bucket{
		Start:        start,
		End:          end,
		MetricValues: m.MetricValues,
	})
	return buckets
}

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
	End   time.Time // non-zero for buckets

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
	Status string    //
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
}

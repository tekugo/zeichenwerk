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

// bucket holds the token-delta sum for one minute.
type bucket struct {
	minute int64 // Unix seconds / 60
	tokens int64
}

// session is the internal per-session state.
type session struct {
	id       string
	name     string
	lastSeen time.Time

	mu        sync.Mutex
	byType    map[string]int64 // latest cumulative tokens per type
	totalCost float64          // accumulated cost across all calls
	buckets   []bucket         // rolling 20-minute history
	sparkline *Sparkline
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

// totalTokens sums the latest cumulative values across all token types.
func (s *session) totalTokens() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	var total int64
	for _, v := range s.byType {
		total += v
	}
	return total
}

// addTokens records a new cumulative token value for type and timestamp.
// Returns the positive delta (0 if unchanged).
func (s *session) addTokens(typ string, cumulative int64, ts time.Time) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	last := s.byType[typ]
	delta := cumulative - last
	if delta < 0 {
		delta = cumulative // counter was reset
	}
	s.byType[typ] = cumulative
	s.lastSeen = ts
	if delta > 0 {
		s.addBucket(ts.Unix()/60, delta)
	}
	return delta
}

func (s *session) addBucket(minute, delta int64) {
	const maxBuckets = 20
	for i := range s.buckets {
		if s.buckets[i].minute == minute {
			s.buckets[i].tokens += delta
			s.rebuildSparkline()
			return
		}
	}
	s.buckets = append(s.buckets, bucket{minute, delta})
	if len(s.buckets) > maxBuckets {
		s.buckets = s.buckets[len(s.buckets)-maxBuckets:]
	}
	s.rebuildSparkline()
}

func (s *session) rebuildSparkline() {
	vs := make([]float64, len(s.buckets))
	for i, b := range s.buckets {
		vs[i] = float64(b.tokens)
	}
	s.sparkline.SetValues(vs)
}

// ---- Store -----------------------------------------------------------------

// Store is the in-memory session registry.
type Store struct {
	mu          sync.RWMutex
	sessions    map[string]*session
	order       []string // insertion order
	idleTimeout time.Duration
	onChange    func()

	gesamtBuckets []bucket
	gesamtSp      *Sparkline
}

// newDeckSparkline creates a Relative-mode sparkline for deck cards.
// Relative scaling ensures visible variation even after a burst of activity.
func newDeckSparkline(id string) *Sparkline {
	return NewSparkline(id, "")
}

func newStore(timeout time.Duration) *Store {
	return &Store{
		sessions:    make(map[string]*session),
		idleTimeout: timeout,
		gesamtSp:    newDeckSparkline("gesamt-sp"),
	}
}

// SetOnChange registers a function called after every data update.
func (st *Store) SetOnChange(fn func()) {
	st.mu.Lock()
	st.onChange = fn
	st.mu.Unlock()
}

// UpdateTokens records a cumulative token count for a session.
func (st *Store) UpdateTokens(sessionID, sessionName, tokenType string, cumulative int64, ts time.Time) {
	st.mu.Lock()
	s, ok := st.sessions[sessionID]
	if !ok {
		s = &session{
			id:        sessionID,
			name:      sessionName,
			lastSeen:  ts,
			byType:    make(map[string]int64),
			sparkline: newDeckSparkline(sessionID + "-sp"),
		}
		st.sessions[sessionID] = s
		st.order = append(st.order, sessionID)
	}
	onChange := st.onChange
	st.mu.Unlock()

	delta := s.addTokens(tokenType, cumulative, ts)
	if delta > 0 {
		st.addGesamtBucket(ts.Unix()/60, delta)
	}
	if onChange != nil {
		onChange()
	}
}

// UpdateCost records the latest cumulative cost for a session.
func (st *Store) UpdateCost(sessionID string, cost float64) {
	st.mu.RLock()
	s := st.sessions[sessionID]
	onChange := st.onChange
	st.mu.RUnlock()
	if s == nil {
		return
	}
	s.mu.Lock()
	s.totalCost += cost
	s.mu.Unlock()
	if onChange != nil {
		onChange()
	}
}

func (st *Store) addGesamtBucket(minute, delta int64) {
	const maxBuckets = 20
	st.mu.Lock()
	defer st.mu.Unlock()
	for i := range st.gesamtBuckets {
		if st.gesamtBuckets[i].minute == minute {
			st.gesamtBuckets[i].tokens += delta
			st.rebuildGesamtSparkline()
			return
		}
	}
	st.gesamtBuckets = append(st.gesamtBuckets, bucket{minute, delta})
	if len(st.gesamtBuckets) > maxBuckets {
		st.gesamtBuckets = st.gesamtBuckets[len(st.gesamtBuckets)-maxBuckets:]
	}
	st.rebuildGesamtSparkline()
}

// rebuildGesamtSparkline must be called under st.mu.
func (st *Store) rebuildGesamtSparkline() {
	vs := make([]float64, len(st.gesamtBuckets))
	for i, b := range st.gesamtBuckets {
		vs[i] = float64(b.tokens)
	}
	st.gesamtSp.SetValues(vs)
}

// ---- SessionItem (deck display snapshot) -----------------------------------

// SessionItem is the read-only snapshot passed to the Deck item renderer.
type SessionItem struct {
	ID           string
	Name         string
	TotalTokens  int64
	InputTokens  int64
	OutputTokens int64
	CacheTokens  int64 // cache_read + cache_write
	TotalCost    float64
	Status       SessionStatus
	Sparkline    *Sparkline
}

// Items returns ordered deck items: Gesamt first, then sessions newest-first.
func (st *Store) Items() []*SessionItem {
	st.mu.RLock()
	defer st.mu.RUnlock()

	type kv struct {
		s    *session
		seen time.Time
	}
	kvs := make([]kv, 0, len(st.sessions))
	var totalTokens, totalInput, totalOutput, totalCache int64
	var totalCost float64
	bestStatus := StatusEnded

	for _, s := range st.sessions {
		kvs = append(kvs, kv{s, s.lastSeen})
		s.mu.Lock()
		inp := s.byType["input"]
		out := s.byType["output"]
		cache := s.byType["cache_read"] + s.byType["cache_write"]
		s.mu.Unlock()
		totalInput += inp
		totalOutput += out
		totalCache += cache
		totalTokens += s.totalTokens()
		totalCost += s.totalCost
		if st := s.status(st.idleTimeout); st < bestStatus {
			bestStatus = st
		}
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].seen.After(kvs[j].seen)
	})

	gesamtStatus := StatusEnded
	if len(kvs) > 0 {
		gesamtStatus = bestStatus
	}

	items := make([]*SessionItem, 0, len(kvs)+1)
	items = append(items, &SessionItem{
		ID:           "__gesamt__",
		Name:         "Gesamt",
		TotalTokens:  totalTokens,
		InputTokens:  totalInput,
		OutputTokens: totalOutput,
		CacheTokens:  totalCache,
		TotalCost:    totalCost,
		Status:       gesamtStatus,
		Sparkline:    st.gesamtSp,
	})
	for _, kv := range kvs {
		s := kv.s
		s.mu.Lock()
		inp := s.byType["input"]
		out := s.byType["output"]
		cache := s.byType["cache_read"] + s.byType["cache_write"]
		s.mu.Unlock()
		items = append(items, &SessionItem{
			ID:           s.id,
			Name:         s.name,
			TotalTokens:  s.totalTokens(),
			InputTokens:  inp,
			OutputTokens: out,
			CacheTokens:  cache,
			TotalCost:    s.totalCost,
			Status:       s.status(st.idleTimeout),
			Sparkline:    s.sparkline,
		})
	}
	return items
}

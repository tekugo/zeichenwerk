package main

import (
	"math/rand"
	"time"
)

// populateSim fills store with two synthetic sessions so the UI can be
// exercised without a live OTLP source.
func populateSim(store *Store) {
	now := time.Now()
	rng := rand.New(rand.NewSource(42))

	type sessionDef struct {
		id      string
		start   time.Duration // relative to now
		model   string
		os      string
		arch    string
		buckets int
	}

	defs := []sessionDef{
		{
			id:      "sim-abc123def456",
			start:   -50 * time.Minute,
			model:   "claude-opus-4-6",
			os:      "linux",
			arch:    "amd64",
			buckets: 25,
		},
		{
			id:      "sim-xyz789uvw012",
			start:   -20 * time.Minute,
			model:   "claude-sonnet-4-6",
			os:      "darwin",
			arch:    "arm64",
			buckets: 10,
		},
	}

	for _, def := range defs {
		s, _ := store.Get(def.id)
		s.ID = def.id
		s.Start = now.Add(def.start)
		s.ServiceName = "claude-code"
		s.ServiceVersion = "1.0.0"
		s.OSType = def.os
		s.HostArch = def.arch
		s.UserEmail = "demo@example.com"

		for i := range def.buckets {
			t := s.Start.Add(time.Duration(i) * 2 * time.Minute)
			m := &Metrics{
				Time:          t,
				Start:         s.Start,
				Model:         def.model,
				Input:         int64(800 + rng.Intn(3200)),
				Output:        int64(150 + rng.Intn(850)),
				CacheRead:     int64(rng.Intn(500)),
				CacheCreation: int64(rng.Intn(200)),
				Cost:          0.008 + rng.Float64()*0.06,
				ActiveUser:    10 + rng.Float64()*50,
				ActiveCLI:     5 + rng.Float64()*20,
				LinesAdded:    int64(rng.Intn(80)),
				LinesRemoved:  int64(rng.Intn(30)),
				Accepted:      int64(rng.Intn(5)),
				Rejected:      int64(rng.Intn(2)),
			}
			s.Add(m)
			s.addToSeries(m)
			store.addToTotal(m)
		}
	}

	store.Notify()
}

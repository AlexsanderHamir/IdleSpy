package tracker

import (
	"sync"
	"time"
	"slices"
)

// GoroutineManager manages statistics for multiple goroutines
type GoroutineId int
type GoroutineManager struct {
	Stats map[GoroutineId]*GoroutineStats
	mu    sync.RWMutex // protect concurrent access to stats
}

// GoroutineStats holds statistics for a single goroutine
type GoroutineStats struct {
	GoroutineId GoroutineId
	SelectStats map[string]*SelectStats
	StartTime   time.Time
	EndTime     time.Time
}

// SelectStats holds statistics for a select case
type SelectStats struct {
	// how long the case was blocked
	BlockedCaseTime time.Duration
	// how many times the case was hit
	CaseHits int
	// individual latencies for percentile calculations
	latencies []time.Duration
	mu        sync.Mutex // protect concurrent access to latencies
}

// AddLatency adds a new latency measurement to the stats
func (s *SelectStats) AddLatency(latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latencies = append(s.latencies, latency)
	s.BlockedCaseTime += latency
	s.CaseHits++
}

// GetPercentile returns the nth percentile latency
func (s *SelectStats) GetPercentile(n float64) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.latencies) == 0 {
		return 0
	}

	// Create a copy of latencies to sort
	latencies := make([]time.Duration, len(s.latencies))
	copy(latencies, s.latencies)

	// Sort latencies
	slices.Sort(latencies)

	// Calculate index for nth percentile
	index := int(float64(len(latencies)-1) * n / 100.0)
	return latencies[index]
}

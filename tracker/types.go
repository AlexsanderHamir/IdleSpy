package tracker

import (
	"sync"
	"time"
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
	CaseTime time.Duration // when no default case is provided, this is the time it takes for the goroutine to do what it needs.
	CaseHits int           // how many times the case was hit
}

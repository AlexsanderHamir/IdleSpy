package tracker

import (
	"maps"
	"time"
)

// GetGoroutineLifetime returns the lifetime duration of a goroutine
func (gs *GoroutineStats) GetGoroutineLifetime() time.Duration {
	if gs.EndTime.IsZero() {
		return time.Since(gs.StartTime)
	}
	return gs.EndTime.Sub(gs.StartTime)
}

// GetTotalSelectTime returns the total time spent in select cases for a goroutine
func (gs *GoroutineStats) GetTotalSelectTime() time.Duration {
	var total time.Duration
	for _, stats := range gs.SelectStats {
		total += stats.CaseTime
	}
	return total
}

// GetSelectCaseStats returns statistics for a specific select case
func (gs *GoroutineStats) GetSelectCaseStats(caseName string) *SelectStats {
	return gs.SelectStats[caseName]
}

// GetSelectStats returns a map of select case statistics
func (gs *GoroutineStats) GetSelectStats() map[string]*SelectStats {
	return maps.Clone(gs.SelectStats)
}

package tracker

import (
	"maps"
	"time"
)

// NewGoroutineManager creates a new goroutine statistics manager
func NewGoroutineManager() *GoroutineManager {
	return &GoroutineManager{
		Stats: make(map[int]*GoroutineStats),
	}
}

// TrackGoroutineStart records the start of a goroutine tracking
func (gm *GoroutineManager) TrackGoroutineStart() int {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	id := getGoroutineID()
	if _, exists := gm.Stats[id]; !exists {
		gm.Stats[id] = &GoroutineStats{
			GoroutineId: id,
			SelectStats: make(map[string]*SelectStats),
			StartTime:   time.Now(),
		}
	}
	return id
}

// TrackGoroutineEnd records the end of a goroutine
func (gm *GoroutineManager) TrackGoroutineEnd() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	id := getGoroutineID()
	if stats, exists := gm.Stats[id]; exists {
		stats.EndTime = time.Now()
	}
}

// TrackSelectCase records statistics for a select case
func (gm *GoroutineManager) TrackSelectCase(caseName string, duration time.Duration) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	id := getGoroutineID()
	stats, exists := gm.Stats[id]
	if !exists {
		stats = &GoroutineStats{
			GoroutineId: id,
			SelectStats: make(map[string]*SelectStats),
			StartTime:   time.Now(),
		}
		gm.Stats[id] = stats
	}

	selectStats, exists := stats.SelectStats[caseName]
	if !exists {
		selectStats = &SelectStats{}
		stats.SelectStats[caseName] = selectStats
	}

	selectStats.CaseHits++
	selectStats.CaseTime += duration
}

// GetGoroutineStats returns statistics for a specific goroutine
func (gm *GoroutineManager) GetGoroutineStats(id int) *GoroutineStats {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.Stats[id]
}

// GetAllStats returns statistics for all goroutines
func (gm *GoroutineManager) GetAllStats() map[int]*GoroutineStats {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return maps.Clone(gm.Stats)
}

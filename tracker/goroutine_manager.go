package tracker

import (
	"fmt"
	"maps"
	"sync"
	"time"
)

// NewGoroutineManager creates a new goroutine statistics manager
func NewGoroutineManager() *GoroutineManager {
	return &GoroutineManager{
		Stats: make(map[GoroutineId]*GoroutineStats),
		mu:    &sync.RWMutex{},
		Wg:    &sync.WaitGroup{},
	}
}

// TrackGoroutineStart records the start of a goroutine tracking
func (gm *GoroutineManager) TrackGoroutineStart() GoroutineId {
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

	gm.Wg.Add(1)
	return id
}

// TrackGoroutineEnd records the end of a goroutine
func (gm *GoroutineManager) TrackGoroutineEnd(id GoroutineId) {
	gm.mu.Lock()
	defer func() {
		gm.Wg.Done()
		gm.mu.Unlock()
	}()

	if stats, exists := gm.Stats[id]; exists {
		stats.EndTime = time.Now()
	}
}

// TrackSelectCase records statistics for a select case
func (gm *GoroutineManager) TrackSelectCase(caseName string, duration time.Duration, id GoroutineId) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

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

	selectStats.AddLatency(duration)
}

// GetGoroutineStats returns statistics for a specific goroutine
func (gm *GoroutineManager) GetGoroutineStats(id GoroutineId) *GoroutineStats {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.Stats[id]
}

// GetAllStats returns statistics for all goroutines
func (gm *GoroutineManager) GetAllStats() map[GoroutineId]*GoroutineStats {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return maps.Clone(gm.Stats)
}

// Done waits for all goroutines to finish and then saves the final stats
func (gm *GoroutineManager) Done() error {
	gm.Wg.Wait()

	if gm.Action == None {
		return nil
	}

	switch gm.FileType {
	case "text":
		gm.handleTextActions()
	case "json":
		gm.handleJsonActions()
	default:
		return fmt.Errorf("invalid file type: %s", gm.FileType)
	}

	return nil
}

package test

import (
	"sync"
	"testing"
	"time"

	"github.com/AlexsanderHamir/IdleSpy/tracker"
)

func TestNewGoroutineManager(t *testing.T) {
	gm := tracker.NewGoroutineManager()
	if gm == nil {
		t.Fatal("NewGoroutineManager returned nil")
	}
	if gm.Stats == nil {
		t.Fatal("stats map was not initialized")
	}
}

func TestTrackGoroutineStartAndEnd(t *testing.T) {
	gm := tracker.NewGoroutineManager()

	// Test tracking in main goroutine
	id := gm.TrackGoroutineStart()
	if id <= 0 {
		t.Errorf("Expected positive goroutine ID, got %d", id)
	}

	// Verify stats were created
	stats := gm.GetGoroutineStats(id)
	if stats == nil {
		t.Fatal("Stats not found for tracked goroutine")
	}
	if stats.StartTime.IsZero() {
		t.Error("Start time was not set")
	}
	if !stats.EndTime.IsZero() {
		t.Error("End time was set before TrackGoroutineEnd")
	}

	// Test end tracking
	gm.TrackGoroutineEnd(id)
	stats = gm.GetGoroutineStats(id)
	if stats.EndTime.IsZero() {
		t.Error("End time was not set after TrackGoroutineEnd")
	}

	// Test lifetime calculation
	lifetime := stats.GetGoroutineLifetime()
	if lifetime <= 0 {
		t.Errorf("Expected positive lifetime, got %v", lifetime)
	}
}

func TestTrackSelectCase(t *testing.T) {
	gm := tracker.NewGoroutineManager()
	id := gm.TrackGoroutineStart()

	// Test tracking a select case
	caseName := "test_case"
	duration := 100 * time.Millisecond
	gm.TrackSelectCase(caseName, duration, id)

	// Verify select stats
	stats := gm.GetGoroutineStats(id)
	selectStats := stats.GetSelectCaseStats(caseName)
	if selectStats == nil {
		t.Fatal("Select stats not found for tracked case")
	}

	if selectStats.GetCaseHits() != 1 {
		t.Errorf("Expected 1 case hit, got %d", selectStats.GetCaseHits())
	}
	if selectStats.GetCaseTime() != duration {
		t.Errorf("Expected case time %v, got %v", duration, selectStats.GetCaseTime())
	}

	// Test multiple hits
	gm.TrackSelectCase(caseName, duration, id)
	selectStats = stats.GetSelectCaseStats(caseName)
	if selectStats.GetCaseHits() != 2 {
		t.Errorf("Expected 2 case hits, got %d", selectStats.GetCaseHits())
	}
	if selectStats.GetCaseTime() != duration*2 {
		t.Errorf("Expected case time %v, got %v", duration*2, selectStats.GetCaseTime())
	}
}

func TestConcurrentTracking(t *testing.T) {
	gm := tracker.NewGoroutineManager()
	var wg sync.WaitGroup
	goroutineCount := 10

	// Launch multiple goroutines that track their own stats
	for range goroutineCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := gm.TrackGoroutineStart()
			defer gm.TrackGoroutineEnd(id)

			// Track some select cases
			gm.TrackSelectCase("case1", 50*time.Millisecond, id)
			gm.TrackSelectCase("case2", 100*time.Millisecond, id)

			// Verify stats are accessible
			stats := gm.GetGoroutineStats(id)
			if stats == nil {
				t.Error("Stats not found for concurrent goroutine")
			}
		}()
	}

	wg.Wait()

	// Verify all goroutines were tracked
	allStats := gm.GetAllStats()
	if len(allStats) != goroutineCount {
		t.Errorf("Expected %d goroutines, got %d", goroutineCount, len(allStats))
	}

	// Verify select stats for each goroutine
	for _, stats := range allStats {
		selectStats := stats.GetSelectStats()
		if len(selectStats) != 2 {
			t.Errorf("Expected 2 select cases, got %d", len(selectStats))
		}

		totalTime := stats.GetTotalSelectTime()
		expectedTime := 150 * time.Millisecond // 50ms + 100ms
		if totalTime != expectedTime {
			t.Errorf("Expected total select time %v, got %v", expectedTime, totalTime)
		}
	}
}

func TestGetGoroutineStats(t *testing.T) {
	gm := tracker.NewGoroutineManager()

	// Test getting stats for non-existent goroutine
	stats := gm.GetGoroutineStats(999)
	if stats != nil {
		t.Error("Expected nil stats for non-existent goroutine")
	}

	// Test getting stats for existing goroutine
	id := gm.TrackGoroutineStart()
	stats = gm.GetGoroutineStats(id)
	if stats == nil {
		t.Fatal("Stats not found for tracked goroutine")
	}
	if stats.GoroutineId != id {
		t.Errorf("Expected goroutine ID %d, got %d", id, stats.GoroutineId)
	}
}

func TestGetAllStats(t *testing.T) {
	gm := tracker.NewGoroutineManager()

	// Track a few goroutines
	id1 := gm.TrackGoroutineStart()
	gm.TrackSelectCase("case1", 100*time.Millisecond, id1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		id := gm.TrackGoroutineStart()
		gm.TrackSelectCase("case2", 200*time.Millisecond, id)
		gm.TrackGoroutineEnd(id)
	}()
	wg.Wait()

	// Get all stats
	allStats := gm.GetAllStats()
	if len(allStats) != 2 {
		t.Errorf("Expected 2 goroutines, got %d", len(allStats))
	}

	// Verify stats for each goroutine
	if stats, exists := allStats[id1]; !exists {
		t.Error("Stats not found for first goroutine")
	} else if len(stats.SelectStats) != 1 {
		t.Errorf("Expected 1 select case for first goroutine, got %d", len(stats.SelectStats))
	}
}

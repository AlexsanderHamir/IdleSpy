package test

import (
	"os"
	"testing"
	"time"

	"github.com/AlexsanderHamir/IdleSpy/tracker"
)

func TestTrackGoroutineStartAndEnd(t *testing.T) {
	gm := tracker.NewGoroutineManager()
	gm.Wg.Add(1)

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
	defer func() {
		os.Remove(".internal.json")
		os.Remove(".internal.txt")
	}()

	latency1 := 50 * time.Millisecond
	latency2 := 100 * time.Millisecond

	gm := tracker.NewGoroutineManager()

	goroutineCount := 10

	gm.FileType = "json"
	gm.Action = tracker.Save
	gm.Wg.Add(goroutineCount)

	for range goroutineCount {
		go func() {
			id := gm.TrackGoroutineStart()
			defer gm.TrackGoroutineEnd(id)

			gm.TrackSelectCase("case1", latency1, id)
			gm.TrackSelectCase("case2", latency2, id)

			stats := gm.GetGoroutineStats(id)
			if stats == nil {
				t.Error("Stats not found for concurrent goroutine")
			}
		}()
	}

	err := gm.Done()
	if err != nil {
		t.Errorf("Error saving stats: %v", err)
	}

	if !tracker.FileExists(".internal.json") {
		t.Error("File does not exist")
	}

	allStats := gm.GetAllStats()
	if len(allStats) != goroutineCount {
		t.Errorf("Expected %d goroutines, got %d", goroutineCount, len(allStats))
	}

	for _, stats := range allStats {
		CheckStatsAccuracy(t, stats, latency1, latency2)
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

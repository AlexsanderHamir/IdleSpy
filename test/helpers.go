package test

import (
	"testing"
	"time"

	"github.com/AlexsanderHamir/IdleSpy/tracker"
)

func CheckStatsAccuracy(t *testing.T, stats *tracker.GoroutineStats, latency1 time.Duration, latency2 time.Duration) {
	selectStats := stats.GetSelectStats()
	if len(selectStats) != 2 {
		t.Errorf("Expected 2 select cases, got %d", len(selectStats))
	}

	if selectStats["case1"].GetCaseHits() != 1 {
		t.Errorf("Expected 1 case hit, got %d", selectStats["case1"].GetCaseHits())
	}

	if selectStats["case2"].GetCaseHits() != 1 {
		t.Errorf("Expected 1 case hit, got %d", selectStats["case2"].GetCaseHits())
	}

	if selectStats["case1"].GetAverage() != latency1 {
		t.Errorf("Expected %v average, got %v", latency1, selectStats["case1"].GetAverage())
	}

	if selectStats["case2"].GetAverage() != latency2 {
		t.Errorf("Expected %v average, got %v", latency2, selectStats["case2"].GetAverage())
	}

	totalTime := stats.GetTotalSelectBlockedTime()
	expectedTime := latency1 + latency2
	if totalTime != expectedTime {
		t.Errorf("Expected total select time %v, got %v", expectedTime, totalTime)
	}
}

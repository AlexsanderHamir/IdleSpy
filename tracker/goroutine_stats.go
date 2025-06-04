package tracker

import (
	"log"
	"maps"
	"strings"
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
		total += stats.BlockedCaseTime
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

// PrintStats prints a summary of goroutine performance statistics
func PrintStats(stats map[GoroutineId]*GoroutineStats, title string) {
	log.Println("\n" + title)
	log.Println(strings.Repeat("=", len(title)))

	for goroutineID, stat := range stats {
		log.Printf("\nGoroutine %d:", goroutineID)
		log.Printf("  Lifetime: %v", stat.GetGoroutineLifetime())
		log.Printf("  Total Select Blocked Time: %v", stat.GetTotalSelectTime())

		log.Println("  Select Case Statistics:")
		for caseName, caseStats := range stat.GetSelectStats() {
			log.Printf("    %s:", caseName)
			log.Printf("      Hits: %d", caseStats.GetCaseHits())
			log.Printf("      Total Blocked Time: %v", caseStats.GetCaseTime())
			if caseStats.GetCaseHits() > 0 {
				log.Printf("      Average Blocked Time: %v", caseStats.GetCaseTime()/time.Duration(caseStats.GetCaseHits()))
				log.Printf("      90th Percentile Blocked Time: %v", caseStats.GetPercentile(90))
				log.Printf("      99th Percentile Blocked Time: %v", caseStats.GetPercentile(99))
			}
		}
	}
}

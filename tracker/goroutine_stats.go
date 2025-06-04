package tracker

import (
	"fmt"
	"io"
	"log"
	"maps"
	"os"
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
func PrintAndSaveStats(stats map[GoroutineId]*GoroutineStats, title string) {
	// Open file for writing
	file, err := os.Create("stats.txt")
	if err != nil {
		log.Printf("Error creating stats file: %v", err)
		return
	}
	defer file.Close()

	// Create a multi-writer to write to both file and stdout
	writer := io.MultiWriter(os.Stdout, file)

	// Write title
	fmt.Fprintln(writer, "\n"+title)
	fmt.Fprintln(writer, strings.Repeat("=", len(title)))

	for goroutineID, stat := range stats {
		fmt.Fprintf(writer, "\nGoroutine %d:\n", goroutineID)
		fmt.Fprintf(writer, "  Lifetime: %v\n", stat.GetGoroutineLifetime())
		fmt.Fprintf(writer, "  Total Select Blocked Time: %v\n", stat.GetTotalSelectTime())

		fmt.Fprintln(writer, "  Select Case Statistics:")
		for caseName, caseStats := range stat.GetSelectStats() {
			fmt.Fprintf(writer, "    %s:\n", caseName)
			fmt.Fprintf(writer, "      Hits: %d\n", caseStats.GetCaseHits())
			fmt.Fprintf(writer, "      Total Blocked Time: %v\n", caseStats.GetCaseTime())
			if caseStats.GetCaseHits() > 0 {
				fmt.Fprintf(writer, "      Average Blocked Time: %v\n", caseStats.GetCaseTime()/time.Duration(caseStats.GetCaseHits()))
				fmt.Fprintf(writer, "      90th Percentile Blocked Time: %v\n", caseStats.GetPercentile(90))
				fmt.Fprintf(writer, "      99th Percentile Blocked Time: %v\n", caseStats.GetPercentile(99))
			}
		}
	}
}

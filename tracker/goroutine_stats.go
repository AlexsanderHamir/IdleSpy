package tracker

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"
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
	file, err := os.Create(fmt.Sprintf("%s.txt", title))
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

// JSONStats represents the complete statistics structure for JSON output
type JSONStats struct {
	Title      string                   `json:"title"`
	Goroutines map[string]GoroutineJSON `json:"goroutines"`
}

// GoroutineJSON represents a single goroutine's statistics in JSON format
type GoroutineJSON struct {
	Lifetime        time.Duration       `json:"lifetime"`
	TotalSelectTime time.Duration       `json:"total_select_blocked_time"`
	SelectCaseStats map[string]CaseJSON `json:"select_case_statistics"`
}

// CaseJSON represents statistics for a single select case in JSON format
type CaseJSON struct {
	Hits             int64         `json:"hits"`
	TotalBlockedTime time.Duration `json:"total_blocked_time"`
	AvgBlockedTime   time.Duration `json:"average_blocked_time,omitempty"`
	Percentile90     time.Duration `json:"percentile_90,omitempty"`
	Percentile99     time.Duration `json:"percentile_99,omitempty"`
}

// SaveStats saves goroutine performance statistics to a JSON file in a directory named after the stage
func SaveStats(stats map[GoroutineId]*GoroutineStats, title string) error {
	// Create static directory if it doesn't exist
	if err := os.MkdirAll("static", 0755); err != nil {
		return fmt.Errorf("error creating static directory: %w", err)
	}

	// Create stage-specific directory under static
	dirName := filepath.Join("static", strings.Split(title, "_")[0]) // Get the stage name (everything before the first underscore)
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dirName, err)
	}

	// Create JSON structure
	jsonStats := JSONStats{
		Title:      title,
		Goroutines: make(map[string]GoroutineJSON),
	}

	// Convert stats to JSON structure
	for goroutineID, stat := range stats {
		goroutineJSON := GoroutineJSON{
			Lifetime:        stat.GetGoroutineLifetime(),
			TotalSelectTime: stat.GetTotalSelectTime(),
			SelectCaseStats: make(map[string]CaseJSON),
		}

		// Convert select case statistics
		for caseName, caseStats := range stat.GetSelectStats() {
			caseJSON := CaseJSON{
				Hits:             int64(caseStats.GetCaseHits()),
				TotalBlockedTime: caseStats.GetCaseTime(),
			}

			if caseStats.GetCaseHits() > 0 {
				caseJSON.AvgBlockedTime = caseStats.GetCaseTime() / time.Duration(caseStats.GetCaseHits())
				caseJSON.Percentile90 = caseStats.GetPercentile(90)
				caseJSON.Percentile99 = caseStats.GetPercentile(99)
			}

			goroutineJSON.SelectCaseStats[caseName] = caseJSON
		}

		jsonStats.Goroutines[fmt.Sprintf("%d", goroutineID)] = goroutineJSON
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(jsonStats, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling stats to JSON: %w", err)
	}

	// Create file path within the directory with .json extension
	filePath := filepath.Join(dirName, fmt.Sprintf("%s.json", title))
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing JSON stats file: %w", err)
	}

	return nil
}

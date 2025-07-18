package tracker

import (
	"encoding/json"
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
func (gs *GoroutineStats) GetTotalSelectBlockedTime() time.Duration {
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
func PrintAndSaveStatsText(stats map[GoroutineId]*GoroutineStats, title string) {
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
		fmt.Fprintf(writer, "  Total Select Blocked Time: %v\n", stat.GetTotalSelectBlockedTime())

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

// SaveStatsText saves a summary of goroutine performance statistics to a text file
func SaveStatsText(stats map[GoroutineId]*GoroutineStats, title string) {
	// Open file for writing
	file, err := os.Create(fmt.Sprintf("%s.txt", title))
	if err != nil {
		log.Printf("Error creating stats file: %v", err)
		return
	}
	defer file.Close()

	// Write title
	fmt.Fprintln(file, "\n"+title)
	fmt.Fprintln(file, strings.Repeat("=", len(title)))

	for goroutineID, stat := range stats {
		fmt.Fprintf(file, "\nGoroutine %d:\n", goroutineID)
		fmt.Fprintf(file, "  Lifetime: %v\n", stat.GetGoroutineLifetime())
		fmt.Fprintf(file, "  Total Select Blocked Time: %v\n", stat.GetTotalSelectBlockedTime())

		fmt.Fprintln(file, "  Select Case Statistics:")
		for caseName, caseStats := range stat.GetSelectStats() {
			fmt.Fprintf(file, "    %s:\n", caseName)
			fmt.Fprintf(file, "      Hits: %d\n", caseStats.GetCaseHits())
			fmt.Fprintf(file, "      Total Blocked Time: %v\n", caseStats.GetCaseTime())
			if caseStats.GetCaseHits() > 0 {
				fmt.Fprintf(file, "      Average Blocked Time: %v\n", caseStats.GetCaseTime()/time.Duration(caseStats.GetCaseHits()))
				fmt.Fprintf(file, "      90th Percentile Blocked Time: %v\n", caseStats.GetPercentile(90))
				fmt.Fprintf(file, "      99th Percentile Blocked Time: %v\n", caseStats.GetPercentile(99))
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

// PrintAndSaveStatsJSON prints and saves goroutine performance statistics as JSON
func PrintAndSaveStatsJSON(stats map[GoroutineId]*GoroutineStats, title string) {
	// Create JSON structure
	jsonStats := JSONStats{
		Title:      title,
		Goroutines: make(map[string]GoroutineJSON),
	}

	// Populate the structure
	for goroutineID, stat := range stats {
		goroutineJSON := GoroutineJSON{
			Lifetime:        stat.GetGoroutineLifetime(),
			TotalSelectTime: stat.GetTotalSelectBlockedTime(),
			SelectCaseStats: make(map[string]CaseJSON),
		}

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

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(jsonStats, "", "  ")
	if err != nil {
		log.Printf("Error marshaling stats to JSON: %v", err)
		return
	}

	// Save to file
	filePath := fmt.Sprintf("%s.json", title)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Printf("Error writing JSON stats file: %v", err)
		return
	}

	// Print to stdout
	fmt.Println(string(jsonData))
}

// SaveStats saves goroutine performance statistics to a JSON file in a directory named after the stage
func SaveStatsJSON(stats map[GoroutineId]*GoroutineStats, title string) error {
	file, err := os.Create(fmt.Sprintf("%s.json", title))
	if err != nil {
		return err
	}
	defer file.Close()

	// Create JSON structure
	jsonStats := JSONStats{
		Title:      title,
		Goroutines: make(map[string]GoroutineJSON),
	}

	// Convert stats to JSON structure
	for goroutineID, stat := range stats {
		goroutineJSON := GoroutineJSON{
			Lifetime:        stat.GetGoroutineLifetime(),
			TotalSelectTime: stat.GetTotalSelectBlockedTime(),
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

	jsonData, err := json.MarshalIndent(jsonStats, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling stats to JSON: %w", err)
	}

	if err := os.WriteFile(file.Name(), jsonData, 0644); err != nil {
		return fmt.Errorf("error writing JSON stats file: %w", err)
	}

	return nil
}

// PrintStatsJSON prints goroutine performance statistics as JSON to stdout
func PrintStatsJSON(stats map[GoroutineId]*GoroutineStats, title string) {
	// Create JSON structure
	jsonStats := JSONStats{
		Title:      title,
		Goroutines: make(map[string]GoroutineJSON),
	}

	// Populate the structure
	for goroutineID, stat := range stats {
		goroutineJSON := GoroutineJSON{
			Lifetime:        stat.GetGoroutineLifetime(),
			TotalSelectTime: stat.GetTotalSelectBlockedTime(),
			SelectCaseStats: make(map[string]CaseJSON),
		}

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

	// Marshal to JSON and print to stdout
	jsonData, err := json.MarshalIndent(jsonStats, "", "  ")
	if err != nil {
		log.Printf("Error marshaling stats to JSON: %v", err)
		return
	}

	fmt.Println(string(jsonData))
}

// PrintStatsText prints a summary of goroutine performance statistics to stdout
func PrintStatsText(stats map[GoroutineId]*GoroutineStats, title string) {
	fmt.Println("\n" + title)
	fmt.Println(strings.Repeat("=", len(title)))

	for goroutineID, stat := range stats {
		fmt.Fprintf(os.Stdout, "\nGoroutine %d:\n", goroutineID)
		fmt.Fprintf(os.Stdout, "  Lifetime: %v\n", stat.GetGoroutineLifetime())
		fmt.Fprintf(os.Stdout, "  Total Select Blocked Time: %v\n", stat.GetTotalSelectBlockedTime())

		fmt.Fprintln(os.Stdout, "  Select Case Statistics:")
		for caseName, caseStats := range stat.GetSelectStats() {
			fmt.Fprintf(os.Stdout, "    %s:\n", caseName)
			fmt.Fprintf(os.Stdout, "      Hits: %d\n", caseStats.GetCaseHits())
			fmt.Fprintf(os.Stdout, "      Total Blocked Time: %v\n", caseStats.GetCaseTime())
			if caseStats.GetCaseHits() > 0 {
				fmt.Fprintf(os.Stdout, "      Average Blocked Time: %v\n", caseStats.GetCaseTime()/time.Duration(caseStats.GetCaseHits()))
				fmt.Fprintf(os.Stdout, "      90th Percentile Blocked Time: %v\n", caseStats.GetPercentile(90))
				fmt.Fprintf(os.Stdout, "      99th Percentile Blocked Time: %v\n", caseStats.GetPercentile(99))
			}
		}
	}
}

var buckets = []time.Duration{
	0,
	10 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	500 * time.Millisecond,
	1 * time.Second,
	5 * time.Second,
	10 * time.Second,
}

func PrintBlockedTimeHistogram(stats map[GoroutineId]*GoroutineStats, title string) {
	histogram := make(map[time.Duration]int)

	for _, b := range buckets {
		histogram[b] = 0
	}

	overflowCount := 0
	for _, stat := range stats {
		blocked := stat.GetTotalSelectBlockedTime()
		placed := false
		for _, b := range buckets {
			if blocked <= b {
				histogram[b]++
				placed = true
				break
			}
		}
		if !placed {
			overflowCount++
		}
	}

	fmt.Printf("\n%s\n%s\n", title, strings.Repeat("=", len(title)))
	for i, b := range buckets {
		var lowerBound time.Duration
		if i > 0 {
			lowerBound = buckets[i-1]
		}
		fmt.Printf("[%v - %v]: %d goroutines\n", lowerBound, b, histogram[b])
	}
	if overflowCount > 0 {
		fmt.Printf("[ > %v ]: %d goroutines\n", buckets[len(buckets)-1], overflowCount)
	}
}

// WriteBlockedTimeHistogramDot writes the blocked time histogram DOT graph to a file named "<stageName>.dot"
func WriteBlockedTimeHistogramDot(stats map[GoroutineId]*GoroutineStats, stageName string) error {
	histogram := make(map[time.Duration]int)
	for _, b := range buckets {
		histogram[b] = 0
	}
	overflowCount := 0
	for _, stat := range stats {
		blocked := stat.GetTotalSelectBlockedTime()
		placed := false
		for _, b := range buckets {
			if blocked <= b {
				histogram[b]++
				placed = true
				break
			}
		}
		if !placed {
			overflowCount++
		}
	}

	var bld strings.Builder

	bld.WriteString("digraph BlockedHistogram {\n")
	bld.WriteString("  label=\"" + stageName + " - Blocked Time Histogram\";\n")
	bld.WriteString("  labelloc=top;\n")
	bld.WriteString("  fontsize=14;\n")
	bld.WriteString("  rankdir=LR;\n")
	bld.WriteString("  node [shape=box, style=filled, fontname=\"Arial\", fontsize=10, fillcolor=lightgray];\n\n")

	for i, b := range buckets {
		var lowerBound time.Duration
		if i > 0 {
			lowerBound = buckets[i-1]
		}
		label := fmt.Sprintf("[%v - %v]\\n%d goroutines", lowerBound, b, histogram[b])
		fmt.Fprintf(&bld, "  bucket_%d [label=\"%s\"];\n", i, label)
	}

	if overflowCount > 0 {
		last := len(buckets)
		label := fmt.Sprintf("> %v\\n%d goroutines", buckets[len(buckets)-1], overflowCount)
		fmt.Fprintf(&bld, "  bucket_%d [label=\"%s\"];\n", last, label)
	}

	for i := 0; i < len(buckets)-1; i++ {
		fmt.Fprintf(&bld, "  bucket_%d -> bucket_%d [style=dashed, arrowsize=0.7];\n", i, i+1)
	}
	if overflowCount > 0 {
		fmt.Fprintf(&bld, "  bucket_%d -> bucket_%d [style=dashed, arrowsize=0.7];\n", len(buckets)-2, len(buckets))
	}

	bld.WriteString("}\n")

	// Clean stageName for a safe filename (optional)
	fileName := stageName + ".dot"
	// Optional: replace spaces with underscores, or sanitize further if needed
	fileName = strings.ReplaceAll(fileName, " ", "_")

	return os.WriteFile(fileName, []byte(bld.String()), 0644)
}

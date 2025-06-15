package visualization

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlexsanderHamir/IdleSpy/sharedtypes"
	"github.com/AlexsanderHamir/IdleSpy/tracker"
)

// JSONStats represents the complete statistics structure for JSON output
type JSONStats struct {
	Title      string                   `json:"title"`
	Goroutines map[string]GoroutineJSON `json:"goroutines"`
}

// GoroutineJSON represents a single goroutine's statistics in JSON format
type GoroutineJSON struct {
	Lifetime               int64                           `json:"lifetime"`
	TotalSelectBlockedTime int64                           `json:"total_select_blocked_time"`
	SelectCaseStats        map[string]sharedtypes.CaseJSON `json:"select_case_statistics"`
}

// CaseStats represents statistics for a single case
type CaseStats struct {
	TotalTime time.Duration
	Count     int
	CaseName  string
	Times     []time.Duration // Store individual blocked times for percentile calculations
}

// GenerateBarChart reads stats from a file and generates a bar chart visualization
func GenerateBarChart(visType sharedtypes.VisualizationType) error {
	statsFile := ".internal.json"
	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("error reading stats file: %w", err)
	}

	err = GenerateBarChartFromJSON(data, visType)
	if err != nil {
		return fmt.Errorf("error generating bar chart: %w", err)
	}

	return nil
}

func GenerateBarChartFromJSON(data []byte, visType sharedtypes.VisualizationType) error {
	stats, goroutineCount, err := ParseJSONToStats(data)
	if err != nil {
		return fmt.Errorf("error parsing stats: %w", err)
	}

	printBarChart(stats, visType, goroutineCount)
	return nil
}

func ParseJSONToStats(data []byte) ([]*sharedtypes.CaseJSON, int, error) {
	var input JSONStats
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, 0, err
	}

	var result []*sharedtypes.CaseJSON
	for _, goroutine := range input.Goroutines {
		for caseName, stat := range goroutine.SelectCaseStats {
			stat.CaseName = caseName
			result = append(result, &stat)
		}
	}

	return result, len(input.Goroutines), nil
}

func printBarChart(caseStats []*sharedtypes.CaseJSON, visType sharedtypes.VisualizationType, goroutineCount int) {
	if len(caseStats) == 0 {
		fmt.Println("No valid statistics found")
		return
	}

	aggregatedStats := tracker.AggregateCaseStats(caseStats)
	var aggregatedSlice []*sharedtypes.CaseJSON
	for _, stat := range aggregatedStats {
		if visType == sharedtypes.AverageTime {
			stat.AvgBlockedTime = stat.AvgBlockedTime / int64(goroutineCount)
		}
		aggregatedSlice = append(aggregatedSlice, stat)
	}
	tracker.SortCaseStats(aggregatedSlice, visType)

	maxValue := tracker.GetMaxValue(aggregatedSlice, visType)

	barWidth := 40
	fmt.Printf("\n%s Across %d Goroutines\n", visType, goroutineCount)
	fmt.Println(strings.Repeat("=", len(visType.String())+30))

	for _, stat := range aggregatedSlice {
		value := tracker.GetValueForCase(stat, visType)
		barLength := int(value / maxValue * float64(barWidth))
		if barLength == 0 && value > 0 {
			barLength = 1
		}

		valueStr := formatDuration(time.Duration(value))
		fmt.Printf("%-20s %s %s\n",
			stat.CaseName,
			strings.Repeat("█", barLength),
			valueStr)
	}
}

func formatDuration(d time.Duration) string {
	if d >= time.Second {
		return fmt.Sprintf("~%.2fs", d.Seconds())
	} else if d >= time.Millisecond {
		return fmt.Sprintf("~%dms", d.Milliseconds())
	} else {
		return fmt.Sprintf("~%dµs", d.Microseconds())
	}
}

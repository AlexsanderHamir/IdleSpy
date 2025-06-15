package visualization

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
)

// JSONStats represents the complete statistics structure for JSON input
type JSONStats struct {
	Title      string                   `json:"title"`
	Goroutines map[string]GoroutineJSON `json:"goroutines"`
}

// GoroutineJSON represents a single goroutine's statistics in JSON format
type GoroutineJSON struct {
	Lifetime               int64               `json:"lifetime"`
	TotalSelectBlockedTime int64               `json:"total_select_blocked_time"`
	SelectCaseStats        map[string]CaseJSON `json:"select_case_statistics"`
}

// CaseJSON represents statistics for a single select case in JSON format
type CaseJSON struct {
	Hits             int64 `json:"hits"`
	TotalBlockedTime int64 `json:"total_blocked_time"`
	AvgBlockedTime   int64 `json:"average_blocked_time"`
	Percentile90     int64 `json:"percentile_90"`
	Percentile99     int64 `json:"percentile_99"`
}

// VisualizationType represents the type of visualization to generate
type VisualizationType int

const (
	TotalBlockedTime VisualizationType = iota
	AverageTime
	Percentile90
	Percentile99
	TotalHits
)

func (vt VisualizationType) String() string {
	switch vt {
	case TotalBlockedTime:
		return "Total"
	case AverageTime:
		return "Average"
	case Percentile90:
		return "90th Percentile"
	case Percentile99:
		return "99th Percentile"
	case TotalHits:
		return "Total Hits"
	default:
		return "Unknown"
	}
}

// CaseStats represents statistics for a single case
type CaseStats struct {
	TotalTime time.Duration
	Count     int
	CaseName  string
	Times     []time.Duration // Store individual blocked times for percentile calculations
}

// GenerateBarChart reads stats from a file and generates a bar chart visualization
func GenerateBarChart(visType VisualizationType) error {
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

func GenerateBarChartFromJSON(data []byte, visType VisualizationType) error {
	stats, goroutineCount, err := ParseJSONToCaseStats(data)
	if err != nil {
		return fmt.Errorf("error parsing stats: %w", err)
	}

	printBarChart(stats, visType, goroutineCount)
	return nil
}

func ParseJSONToCaseStats(data []byte) ([]CaseStats, int, error) {
	var input JSONStats
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, 0, err
	}

	// Aggregate stats per case name
	caseStatsMap := make(map[string]*CaseStats)

	for _, goroutine := range input.Goroutines {
		for caseName, stat := range goroutine.SelectCaseStats {
			entry, exists := caseStatsMap[caseName]
			if !exists {
				entry = &CaseStats{
					CaseName: caseName,
				}
				caseStatsMap[caseName] = entry
			}

			entry.Count += int(stat.Hits)
			entry.TotalTime += time.Duration(stat.TotalBlockedTime)

			for i := int64(0); i < stat.Hits; i++ {
				entry.Times = append(entry.Times, time.Duration(stat.AvgBlockedTime))
			}
		}
	}

	// Convert map to slice
	var result []CaseStats
	for _, v := range caseStatsMap {
		result = append(result, *v)
	}

	return result, len(input.Goroutines), nil
}

func calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	slices.Sort(times)
	index := int(float64(len(times)-1) * percentile / 100.0)
	return times[index]
}

func printBarChart(stats []CaseStats, visType VisualizationType, goroutineCount int) {
	if len(stats) == 0 {
		fmt.Println("No valid statistics found")
		return
	}

	sort.Slice(stats, func(i, j int) bool {
		switch visType {
		case TotalBlockedTime:
			return stats[i].TotalTime > stats[j].TotalTime
		case AverageTime:
			return (stats[i].TotalTime / time.Duration(stats[i].Count)) > (stats[j].TotalTime / time.Duration(stats[j].Count))
		case Percentile90:
			return calculatePercentile(stats[i].Times, 90) > calculatePercentile(stats[j].Times, 90)
		case Percentile99:
			return calculatePercentile(stats[i].Times, 99) > calculatePercentile(stats[j].Times, 99)
		case TotalHits:
			return stats[i].Count > stats[j].Count
		default:
			return false
		}
	})

	var maxValue float64
	switch visType {
	case TotalBlockedTime:
		maxValue = float64(stats[0].TotalTime)
	case AverageTime:
		maxValue = float64(stats[0].TotalTime) / float64(stats[0].Count)
	case Percentile90:
		maxValue = float64(calculatePercentile(stats[0].Times, 90))
	case Percentile99:
		maxValue = float64(calculatePercentile(stats[0].Times, 99))
	case TotalHits:
		maxValue = float64(stats[0].Count)
	}
	barWidth := 40

	fmt.Printf("\n%s Blocked Time Across %d Goroutines\n", visType, goroutineCount)
	fmt.Println(strings.Repeat("=", len(visType.String())+30))

	for _, stat := range stats {
		var value float64
		var valueStr string

		switch visType {
		case TotalBlockedTime:
			value = float64(stat.TotalTime)
			valueStr = formatDuration(stat.TotalTime)
		case AverageTime:
			value = float64(stat.TotalTime) / float64(stat.Count)
			valueStr = formatDuration(time.Duration(value))
		case Percentile90:
			value = float64(calculatePercentile(stat.Times, 90))
			valueStr = formatDuration(time.Duration(value))
		case Percentile99:
			value = float64(calculatePercentile(stat.Times, 99))
			valueStr = formatDuration(time.Duration(value))
		case TotalHits:
			value = float64(stat.Count)
			valueStr = fmt.Sprintf("%d hits", stat.Count)
		}

		barLength := int(value / maxValue * float64(barWidth))
		if barLength == 0 && value > 0 {
			barLength = 1
		}

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

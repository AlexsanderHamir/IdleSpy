package visualization

import (
	"encoding/json"
	"fmt"
	"os"
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
	CaseName         string `json:"case_name"`
	Hits             int64  `json:"hits"`
	TotalBlockedTime int64  `json:"total_blocked_time"`
	AvgBlockedTime   int64  `json:"average_blocked_time"`
	Percentile90     int64  `json:"percentile_90"`
	Percentile99     int64  `json:"percentile_99"`
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
	stats, goroutineCount, err := ParseJSONToStats(data)
	if err != nil {
		return fmt.Errorf("error parsing stats: %w", err)
	}

	printBarChart(stats, visType, goroutineCount)
	return nil
}

func ParseJSONToStats(data []byte) ([]*CaseJSON, int, error) {
	var input JSONStats
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, 0, err
	}

	var result []*CaseJSON
	for _, goroutine := range input.Goroutines {
		for caseName, stat := range goroutine.SelectCaseStats {
			stat.CaseName = caseName
			result = append(result, &stat)
		}
	}

	return result, len(input.Goroutines), nil
}

func printBarChart(caseStats []*CaseJSON, visType VisualizationType, goroutineCount int) {
	if len(caseStats) == 0 {
		fmt.Println("No valid statistics found")
		return
	}

	aggregatedStats := make(map[string]*CaseJSON)
	for _, stat := range caseStats {
		if existing, exists := aggregatedStats[stat.CaseName]; exists {
			existing.Hits += stat.Hits
			existing.TotalBlockedTime += stat.TotalBlockedTime
			existing.AvgBlockedTime += stat.AvgBlockedTime

			if stat.Percentile90 > existing.Percentile90 {
				existing.Percentile90 = stat.Percentile90
			}
			if stat.Percentile99 > existing.Percentile99 {
				existing.Percentile99 = stat.Percentile99
			}
		} else {
			aggregatedStats[stat.CaseName] = &CaseJSON{
				CaseName:         stat.CaseName,
				Hits:             stat.Hits,
				TotalBlockedTime: stat.TotalBlockedTime,
				AvgBlockedTime:   stat.AvgBlockedTime,
				Percentile90:     stat.Percentile90,
				Percentile99:     stat.Percentile99,
			}
		}
	}

	var aggregatedSlice []*CaseJSON
	for _, stat := range aggregatedStats {
		aggregatedSlice = append(aggregatedSlice, stat)
	}

	sort.Slice(aggregatedSlice, func(i, j int) bool {
		switch visType {
		case TotalBlockedTime:
			return aggregatedSlice[i].TotalBlockedTime > aggregatedSlice[j].TotalBlockedTime
		case AverageTime:
			return aggregatedSlice[i].AvgBlockedTime > aggregatedSlice[j].AvgBlockedTime
		case Percentile90:
			return aggregatedSlice[i].Percentile90 > aggregatedSlice[j].Percentile90
		case Percentile99:
			return aggregatedSlice[i].Percentile99 > aggregatedSlice[j].Percentile99
		case TotalHits:
			return aggregatedSlice[i].Hits > aggregatedSlice[j].Hits
		default:
			return false
		}
	})

	var maxValue float64
	switch visType {
	case TotalBlockedTime:
		maxValue = float64(aggregatedSlice[0].TotalBlockedTime)
	case AverageTime:
		maxValue = float64(aggregatedSlice[0].AvgBlockedTime)
	case Percentile90:
		maxValue = float64(aggregatedSlice[0].Percentile90)
	case Percentile99:
		maxValue = float64(aggregatedSlice[0].Percentile99)
	case TotalHits:
		maxValue = float64(aggregatedSlice[0].Hits)
	}

	barWidth := 40
	fmt.Printf("\n%s Blocked Time Across %d Goroutines\n", visType, goroutineCount)
	fmt.Println(strings.Repeat("=", len(visType.String())+30))

	// Print bars for each aggregated case
	for _, stat := range aggregatedSlice {
		var value float64
		switch visType {
		case TotalBlockedTime:
			value = float64(stat.TotalBlockedTime)
		case AverageTime:
			value = float64(stat.AvgBlockedTime)
		case Percentile90:
			value = float64(stat.Percentile90)
		case Percentile99:
			value = float64(stat.Percentile99)
		case TotalHits:
			value = float64(stat.Hits)
		}

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

package visualization

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// VisualizationType represents the type of visualization to generate
type VisualizationType int

const (
	TotalTime VisualizationType = iota
	AverageTime
	Percentile90
	Percentile99
	TotalHits
)

func (vt VisualizationType) String() string {
	switch vt {
	case TotalTime:
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
func GenerateBarChart(statsFile string, visType VisualizationType) error {
	file, err := os.Open(statsFile)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	stats, err := parseStats(file)
	if err != nil {
		return fmt.Errorf("error parsing stats: %w", err)
	}

	printBarChart(stats, visType)
	return nil
}

func calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	slices.Sort(times)
	index := int(float64(len(times)-1) * percentile / 100.0)
	return times[index]
}

func printBarChart(stats []CaseStats, visType VisualizationType) {
	if len(stats) == 0 {
		fmt.Println("No valid statistics found")
		return
	}

	sort.Slice(stats, func(i, j int) bool {
		switch visType {
		case TotalTime:
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
	case TotalTime:
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

	fmt.Printf("\n%s Blocked Time Across Goroutines\n", visType)
	fmt.Println(strings.Repeat("=", len(visType.String())+30))

	for _, stat := range stats {
		var value float64
		var valueStr string

		switch visType {
		case TotalTime:
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

		fmt.Printf("%-20s %s %s (from %d goroutines)\n",
			stat.CaseName,
			strings.Repeat("█", barLength),
			valueStr,
			stat.Count)
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

func parseStats(file *os.File) ([]CaseStats, error) {
	statsMap := make(map[string]*CaseStats)
	scanner := bufio.NewScanner(file)

	// Regex patterns for parsing
	casePattern := regexp.MustCompile(`^\s+([a-z_]+):$`)
	blockedTimePattern := regexp.MustCompile(`Total Blocked Time: ([\d.]+)([a-zµ]+)`)

	var currentCase *CaseStats

	for scanner.Scan() {
		line := scanner.Text()

		// Skip goroutine headers and other non-case lines
		if strings.HasPrefix(line, "Goroutine") ||
			strings.HasPrefix(line, "Worker Performance Statistics") ||
			strings.HasPrefix(line, "Lifetime:") ||
			strings.HasPrefix(line, "Total Select Blocked Time:") ||
			strings.HasPrefix(line, "Select Case Statistics:") {
			continue
		}

		if matches := casePattern.FindStringSubmatch(line); matches != nil {
			caseName := matches[1]
			if _, exists := statsMap[caseName]; !exists {
				statsMap[caseName] = &CaseStats{
					CaseName: caseName,
					Times:    make([]time.Duration, 0),
				}
			}
			currentCase = statsMap[caseName]
			continue
		}

		if currentCase == nil {
			continue
		}

		if matches := blockedTimePattern.FindStringSubmatch(line); matches != nil {
			duration := parseDuration(matches[1] + matches[2])
			currentCase.TotalTime += duration
			currentCase.Count++
			currentCase.Times = append(currentCase.Times, duration)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Convert map to slice
	stats := make([]CaseStats, 0, len(statsMap))
	for _, stat := range statsMap {
		stats = append(stats, *stat)
	}

	return stats, nil
}

func parseDuration(durationStr string) time.Duration {
	durationStr = strings.TrimSpace(strings.ToLower(durationStr))

	if strings.HasSuffix(durationStr, "ns") {
		ns, err := strconv.ParseFloat(strings.TrimSuffix(durationStr, "ns"), 64)
		if err != nil {
			return 0
		}
		return time.Duration(ns * float64(time.Nanosecond))
	} else if strings.HasSuffix(durationStr, "µs") || strings.HasSuffix(durationStr, "µ") {
		usStr := strings.TrimSuffix(strings.TrimSuffix(durationStr, "µs"), "µ")
		us, err := strconv.ParseFloat(usStr, 64)
		if err != nil {
			return 0
		}
		return time.Duration(us * float64(time.Microsecond))
	} else if strings.HasSuffix(durationStr, "ms") {
		ms, err := strconv.ParseFloat(strings.TrimSuffix(durationStr, "ms"), 64)
		if err != nil {
			return 0
		}
		return time.Duration(ms * float64(time.Millisecond))
	} else if strings.HasSuffix(durationStr, "m") {
		minutes, err := strconv.ParseFloat(strings.TrimSuffix(durationStr, "m"), 64)
		if err != nil {
			return 0
		}
		return time.Duration(minutes * float64(time.Minute))
	} else if strings.HasSuffix(durationStr, "s") {
		seconds, err := strconv.ParseFloat(strings.TrimSuffix(durationStr, "s"), 64)
		if err != nil {
			return 0
		}
		return time.Duration(seconds * float64(time.Second))
	}
	return 0
}

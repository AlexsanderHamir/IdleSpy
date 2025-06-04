package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"slices"
)

type caseStats struct {
	totalTime time.Duration
	count     int
	caseName  string
	times     []time.Duration // Store individual blocked times for percentile calculations
}

type visualizationType int

const (
	totalTime visualizationType = iota
	averageTime
	percentile90
	percentile99
)

func (vt visualizationType) String() string {
	switch vt {
	case totalTime:
		return "Total"
	case averageTime:
		return "Average"
	case percentile90:
		return "90th Percentile"
	case percentile99:
		return "99th Percentile"
	default:
		return "Unknown"
	}
}

func parseDuration(durationStr string) time.Duration {
	// Remove any spaces and convert to lowercase
	durationStr = strings.TrimSpace(strings.ToLower(durationStr))

	// Handle different duration formats - check more specific suffixes first
	if strings.HasSuffix(durationStr, "ns") {
		// Convert nanoseconds
		ns, err := strconv.ParseFloat(strings.TrimSuffix(durationStr, "ns"), 64)
		if err != nil {
			fmt.Printf("Error parsing nanoseconds: %v\n", err)
			return 0
		}
		return time.Duration(ns * float64(time.Nanosecond))
	} else if strings.HasSuffix(durationStr, "µs") || strings.HasSuffix(durationStr, "µ") {
		// Convert microseconds - handle both µs and µ suffixes
		usStr := strings.TrimSuffix(strings.TrimSuffix(durationStr, "µs"), "µ")
		us, err := strconv.ParseFloat(usStr, 64)
		if err != nil {
			fmt.Printf("Error parsing microseconds: %v\n", err)
			return 0
		}
		return time.Duration(us * float64(time.Microsecond))
	} else if strings.HasSuffix(durationStr, "ms") {
		// Convert milliseconds
		ms, err := strconv.ParseFloat(strings.TrimSuffix(durationStr, "ms"), 64)
		if err != nil {
			fmt.Printf("Error parsing milliseconds: %v\n", err)
			return 0
		}
		return time.Duration(ms * float64(time.Millisecond))
	} else if strings.HasSuffix(durationStr, "m") {
		// Convert minutes
		minutes, err := strconv.ParseFloat(strings.TrimSuffix(durationStr, "m"), 64)
		if err != nil {
			fmt.Printf("Error parsing minutes: %v\n", err)
			return 0
		}
		return time.Duration(minutes * float64(time.Minute))
	} else if strings.HasSuffix(durationStr, "s") {
		// Convert seconds
		seconds, err := strconv.ParseFloat(strings.TrimSuffix(durationStr, "s"), 64)
		if err != nil {
			fmt.Printf("Error parsing seconds: %v\n", err)
			return 0
		}
		return time.Duration(seconds * float64(time.Second))
	}
	fmt.Printf("Unknown duration format: %s\n", durationStr)
	return 0
}

func calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	// Sort times in ascending order
	slices.Sort(times)

	// Calculate index for the percentile
	index := int(float64(len(times)-1) * percentile / 100.0)
	return times[index]
}

func printBarChart(stats []caseStats, visType visualizationType) {
	if len(stats) == 0 {
		fmt.Println("No valid statistics found")
		return
	}

	// Sort based on visualization type
	sort.Slice(stats, func(i, j int) bool {
		var valueI, valueJ time.Duration
		switch visType {
		case totalTime:
			valueI, valueJ = stats[i].totalTime, stats[j].totalTime
		case averageTime:
			valueI = stats[i].totalTime / time.Duration(stats[i].count)
			valueJ = stats[j].totalTime / time.Duration(stats[j].count)
		case percentile90:
			valueI = calculatePercentile(stats[i].times, 90)
			valueJ = calculatePercentile(stats[j].times, 90)
		case percentile99:
			valueI = calculatePercentile(stats[i].times, 99)
			valueJ = calculatePercentile(stats[j].times, 99)
		}
		return valueI > valueJ
	})

	// Find the maximum value for scaling
	var maxValue time.Duration
	switch visType {
	case totalTime:
		maxValue = stats[0].totalTime
	case averageTime:
		maxValue = stats[0].totalTime / time.Duration(stats[0].count)
	case percentile90:
		maxValue = calculatePercentile(stats[0].times, 90)
	case percentile99:
		maxValue = calculatePercentile(stats[0].times, 99)
	}
	barWidth := 40 // Width of the bar chart in characters

	// Print appropriate title
	fmt.Printf("\n%s Blocked Time Across Goroutines\n", visType)
	fmt.Println(strings.Repeat("=", len(visType.String())+30))

	for _, stat := range stats {
		var value time.Duration
		switch visType {
		case totalTime:
			value = stat.totalTime
		case averageTime:
			value = stat.totalTime / time.Duration(stat.count)
		case percentile90:
			value = calculatePercentile(stat.times, 90)
		case percentile99:
			value = calculatePercentile(stat.times, 99)
		}

		// Calculate bar length
		barLength := int(float64(value) / float64(maxValue) * float64(barWidth))
		if barLength == 0 && value > 0 {
			barLength = 1 // Ensure at least one character for non-zero values
		}

		// Format the duration
		var durationStr string
		if value >= time.Second {
			durationStr = fmt.Sprintf("~%.2fs", value.Seconds())
		} else if value >= time.Millisecond {
			durationStr = fmt.Sprintf("~%dms", value.Milliseconds())
		} else {
			durationStr = fmt.Sprintf("~%dµs", value.Microseconds())
		}

		// Print the bar
		fmt.Printf("%-20s %s %s (from %d goroutines)\n",
			stat.caseName,
			strings.Repeat("█", barLength),
			durationStr,
			stat.count)
	}
}

func main() {
	// Add command line flags for visualization type
	visType := flag.String("type", "total", "Visualization type: total, average, p90, or p99")
	flag.Parse()

	// Parse visualization type
	var selectedType visualizationType
	switch strings.ToLower(*visType) {
	case "total":
		selectedType = totalTime
	case "average":
		selectedType = averageTime
	case "p90":
		selectedType = percentile90
	case "p99":
		selectedType = percentile99
	default:
		fmt.Printf("Unknown visualization type: %s. Using total time.\n", *visType)
		selectedType = totalTime
	}

	file, err := os.Open("../../cmd/visualization/output/stats.txt")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Map to store stats for each case
	statsMap := make(map[string]*caseStats)

	scanner := bufio.NewScanner(file)
	// Support both total and average time patterns
	caseTimeRegex := regexp.MustCompile(`(?:Total|Average) Blocked Time: (.*)`)
	caseNameRegex := regexp.MustCompile(`\s+(\w+):$`)

	var currentCase string
	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a case name line
		if matches := caseNameRegex.FindStringSubmatch(line); matches != nil {
			currentCase = matches[1]
			if _, exists := statsMap[currentCase]; !exists {
				statsMap[currentCase] = &caseStats{
					caseName: currentCase,
					times:    make([]time.Duration, 0),
				}
			}
			continue
		}

		// Check if this is a blocked time line
		if matches := caseTimeRegex.FindStringSubmatch(line); matches != nil && currentCase != "" {
			duration := parseDuration(matches[1])
			stats := statsMap[currentCase]
			stats.totalTime += duration
			stats.count++
			stats.times = append(stats.times, duration)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Convert map to slice for sorting
	var stats []caseStats
	for _, stat := range statsMap {
		if stat.count > 0 { // Only include cases with data
			stats = append(stats, *stat)
		}
	}

	// Print the appropriate visualization
	printBarChart(stats, selectedType)
}

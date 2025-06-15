package visualization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GoroutineStats represents statistics for a single goroutine
type GoroutineStats struct {
	ID           int
	Lifetime     time.Duration
	BlockedTimes []time.Duration
	StartTime    time.Time
	EndTime      time.Time
	Efficiency   float64
}

// GenerateLineGraph reads stats from a file and generates a line graph visualization
func GenerateLineGraph(statsFile string) error {
	file, err := os.Open(statsFile)
	if err != nil {
		return fmt.Errorf("error opening stats file: %w", err)
	}
	defer file.Close()

	if strings.HasSuffix(statsFile, ".json") {
		return GenerateLineGraphFromJSON(statsFile)
	}

	scanner := bufio.NewScanner(file)
	stats, err := parseGoroutineStats(scanner)
	if err != nil {
		return fmt.Errorf("error parsing input: %w", err)
	}

	printLineGraph(stats)
	return nil
}

func GenerateLineGraphFromJSON(statsFile string) error {
	file, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	stats, err := ParseJSONToGoroutineStats(file)
	if err != nil {
		return fmt.Errorf("error parsing stats: %w", err)
	}

	printLineGraph(stats)
	return nil
}

func ParseJSONToGoroutineStats(data []byte) ([]GoroutineStats, error) {
	var input JSONStats
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	var stats []GoroutineStats
	now := time.Now()

	for idStr, g := range input.Goroutines {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid goroutine ID: %s", idStr)
		}

		// Accumulate all simulated individual blocked times
		var blockedTimes []time.Duration
		var totalBlocked time.Duration

		for _, caseStat := range g.SelectCaseStats {
			avg := time.Duration(caseStat.AvgBlockedTime)
			for i := int64(0); i < caseStat.Hits; i++ {
				blockedTimes = append(blockedTimes, avg)
			}
			totalBlocked += time.Duration(caseStat.TotalBlockedTime)
		}

		lifetime := time.Duration(g.Lifetime)
		endTime := now
		startTime := endTime.Add(-lifetime)

		efficiency := 1.0
		if lifetime > 0 {
			efficiency = 1 - float64(totalBlocked)/float64(lifetime)
		}

		stats = append(stats, GoroutineStats{
			ID:           id,
			Lifetime:     lifetime,
			BlockedTimes: blockedTimes,
			StartTime:    startTime,
			EndTime:      endTime,
			Efficiency:   efficiency,
		})
	}

	return stats, nil
}

func parseGoroutineStats(scanner *bufio.Scanner) ([]GoroutineStats, error) {
	var stats []GoroutineStats
	goroutineMap := make(map[int]*GoroutineStats)

	goroutinePattern := regexp.MustCompile(`Goroutine (\d+):`)
	lifetimePattern := regexp.MustCompile(`Lifetime: ([\d.]+)s`)
	blockedTimePattern := regexp.MustCompile(`Total Select Blocked Time: ([\d.]+)([a-zµ]+)`)

	var currentGoroutine *GoroutineStats
	var startTime time.Time

	for scanner.Scan() {
		line := scanner.Text()

		if matches := goroutinePattern.FindStringSubmatch(line); matches != nil {
			id, _ := strconv.Atoi(matches[1])
			startTime = time.Now().Add(-20 * time.Second)
			currentGoroutine = &GoroutineStats{
				ID:           id,
				StartTime:    startTime,
				BlockedTimes: []time.Duration{},
			}
			goroutineMap[id] = currentGoroutine
			continue
		}

		if currentGoroutine == nil {
			continue
		}

		if matches := lifetimePattern.FindStringSubmatch(line); matches != nil {
			lifetime, _ := strconv.ParseFloat(matches[1], 64)
			currentGoroutine.Lifetime = time.Duration(lifetime * float64(time.Second))
			currentGoroutine.EndTime = currentGoroutine.StartTime.Add(currentGoroutine.Lifetime)
			continue
		}

		if matches := blockedTimePattern.FindStringSubmatch(line); matches != nil {
			valueStr := matches[1]
			unit := matches[2]

			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue
			}

			var seconds float64
			switch unit {
			case "s":
				seconds = value
			case "ms":
				seconds = value / 1000
			case "µs":
				seconds = value / 1000000
			case "ns":
				seconds = value / 1000000000
			default:
				continue
			}

			blockedDuration := time.Duration(seconds * float64(time.Second))
			currentGoroutine.BlockedTimes = append(currentGoroutine.BlockedTimes, blockedDuration)

			if currentGoroutine.Lifetime > 0 {
				currentGoroutine.Efficiency = float64(currentGoroutine.Lifetime-blockedDuration) / float64(currentGoroutine.Lifetime)
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	for _, g := range goroutineMap {
		if g.Lifetime > 0 {
			stats = append(stats, *g)
		}
	}

	return stats, nil
}

func printLineGraph(stats []GoroutineStats) {
	if len(stats) == 0 {
		fmt.Println("No valid goroutine statistics found")
		return
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].ID < stats[j].ID
	})

	fmt.Println("\nGoroutine Efficiency Scores")
	fmt.Println(strings.Repeat("=", 30))

	for _, g := range stats {
		efficiency := g.Efficiency
		if efficiency < 0 {
			efficiency = 0
		} else if efficiency > 1 {
			efficiency = 1
		}

		efficiencyPercent := efficiency * 100
		barWidth := 40
		filledWidth := int(efficiency * float64(barWidth))

		fmt.Printf("Goroutine %-4d [%s%s] %.1f%%\n",
			g.ID,
			strings.Repeat("█", filledWidth),
			strings.Repeat("░", barWidth-filledWidth),
			efficiencyPercent)

		fmt.Printf("    Lifetime: %.6fs\n", g.Lifetime.Seconds())
		fmt.Printf("    Blocked: %.6fs\n", sumDurations(g.BlockedTimes).Seconds())
		fmt.Println()
	}

}

func sumDurations(durations []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total
}

package visualization

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GoroutineStats represents statistics for a single goroutine
type GoroutineStats struct {
	ID               int
	Lifetime         time.Duration
	TotalBlockedTime time.Duration
	StartTime        time.Time
	EndTime          time.Time
	Efficiency       float64
}

// GenerateLineGraph reads stats from a file and generates a line graph visualization
func GenerateLineGraph() error {
	statsFile := ".internal.json"
	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("error reading stats file: %w", err)
	}

	err = GenerateLineGraphFromJSON(data)
	if err != nil {
		return fmt.Errorf("error generating line graph: %w", err)
	}

	return nil
}

func GenerateLineGraphFromJSON(data []byte) error {
	stats, err := ParseJSONToGoroutineStats(data)
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

	for idStr, g := range input.Goroutines {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid goroutine ID: %s", idStr)
		}

		totalBlocked := time.Duration(g.TotalSelectBlockedTime)
		lifetime := time.Duration(g.Lifetime)

		var efficiency float64
		if lifetime > 0 {
			efficiency = 1 - float64(totalBlocked)/float64(lifetime)
		}

		stats = append(stats, GoroutineStats{
			ID:               id,
			Lifetime:         lifetime,
			Efficiency:       efficiency,
			TotalBlockedTime: totalBlocked,
		})
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
		fmt.Printf("    Blocked: %.6fs\n", g.TotalBlockedTime.Seconds())
		fmt.Println()
	}

}

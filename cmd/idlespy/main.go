package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlexsanderHamir/IdleSpy/sharedtypes"
	"github.com/AlexsanderHamir/IdleSpy/visualization"
)

const chartDescriptions = `
Available chart types:
  score				 - Shows the efficiency score for each goroutine, ratio of the lifetime of the goroutine and the time it was blocked
  sum-total-blocked-time - Displays the sum of the total blocked time for each select across all goroutines
  avg-blocked-time   - Shows the average blocked time across all goroutines and selects
  p90-blocked-time   - Displays the 90th percentile blocked time for each select across all goroutines
  p99-blocked-time   - Shows the 99th percentile blocked time for each select across all goroutines
  hits				 - Visualizes the total number of hits for each select across all goroutines
`

func main() {
	chartType := flag.String("chart", "score", "Type of chart to generate (see descriptions below)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, chartDescriptions)
	}

	flag.Parse()

	var err error
	switch *chartType {
	case "score":
		err = visualization.GenerateLineGraph()
	case "sum-total-blocked-time":
		err = visualization.GenerateBarChart(sharedtypes.TotalBlockedTime)
	case "avg-blocked-time":
		err = visualization.GenerateBarChart(sharedtypes.AverageTime)
	case "sum-p90-blocked-time":
		err = visualization.GenerateBarChart(sharedtypes.Percentile90)
	case "sum-p99-blocked-time":
		err = visualization.GenerateBarChart(sharedtypes.Percentile99)
	case "sum-hits":
		err = visualization.GenerateBarChart(sharedtypes.TotalHits)
	default:
		fmt.Printf("Error: unknown chart type '%s'\n", *chartType)
		fmt.Print(chartDescriptions)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error generating visualization: %v\n", err)
		os.Exit(1)
	}
}

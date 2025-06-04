package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlexsanderHamir/IdleSpy/visualization"
)

const chartDescriptions = `
Available chart types:
  line      - Shows the efficiency score for each goroutine, ratio of the lifetime of the goroutine and the time it was blocked
  bar-total - Displays the total response time for each select across all goroutines
  bar-avg   - Shows the average response time for each select across all goroutines
  bar-p90   - Displays the 90th percentile response time for each select across all goroutines
  bar-p99   - Shows the 99th percentile response time for each select across all goroutines
  bar-hits  - Visualizes the total number of requests for each select across all goroutines
`

func main() {
	// Define command line flags
	statsFile := flag.String("file", "", "Path to the stats file to visualize")
	chartType := flag.String("chart", "line", "Type of chart to generate (see descriptions below)")

	// Custom usage function to include chart descriptions
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, chartDescriptions)
	}

	flag.Parse()

	if *statsFile == "" {
		fmt.Println("Error: stats file path is required")
		flag.Usage()
		os.Exit(1)
	}

	var err error
	switch *chartType {
	case "line":
		err = visualization.GenerateLineGraph(*statsFile)
	case "bar-total":
		err = visualization.GenerateBarChart(*statsFile, visualization.TotalTime)
	case "bar-avg":
		err = visualization.GenerateBarChart(*statsFile, visualization.AverageTime)
	case "bar-p90":
		err = visualization.GenerateBarChart(*statsFile, visualization.Percentile90)
	case "bar-p99":
		err = visualization.GenerateBarChart(*statsFile, visualization.Percentile99)
	case "bar-hits":
		err = visualization.GenerateBarChart(*statsFile, visualization.TotalHits)
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

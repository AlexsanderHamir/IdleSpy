package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlexsanderHamir/IdleSpy/visualization"
)

func main() {
	// Define command line flags
	statsFile := flag.String("file", "", "Path to the stats file to visualize")
	chartType := flag.String("chart", "line", "Type of chart to generate (line, bar-total, bar-avg, bar-p90, bar-p99, bar-hits)")

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
		fmt.Println("Available types: line, bar-total, bar-avg, bar-p90, bar-p99, bar-hits")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error generating visualization: %v\n", err)
		os.Exit(1)
	}
}

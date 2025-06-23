## IdleSpy: Visual Concurrency Profiler for Go

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexsanderHamir/IdleSpy)](https://goreportcard.com/report/github.com/AlexsanderHamir/IdleSpy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Issues](https://img.shields.io/github/issues/AlexsanderHamir/IdleSpy)
![Last Commit](https://img.shields.io/github/last-commit/AlexsanderHamir/IdleSpy)
![Code Size](https://img.shields.io/github/languages/code-size/AlexsanderHamir/IdleSpy)

IdleSpy is a Go library and CLI tool for analyzing goroutines in high-concurrency applications. It shows where time is spent inside select statements and how often each path is blocked, helping you quickly identify inefficiencies and improve concurrency performance.

## Table of Contents

- [What IdleSpy Tracks](#-what-idlespy-tracks)
- [CLI-Generated Charts](#-cli-generated-charts)
- [Installation](#installation)
- [Tracker Usage](#tracker-usage)
  - [Basic Usage](#basic-usage)
- [CLI Usage](#cli-usage)
  - [Understanding the Statistics](#understanding-the-statistics)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [License](#license)

### 🔍 What IdleSpy Tracks

- **Goroutine Lifetime**: Start, end, and total duration.
- **Select Case Activity**: How many times each case is hit and how long it blocks.
- **Blocking Behavior**: Tracks average, total, and percentile (P90/P99) blocked times per select case.

### 📊 CLI-Generated Charts

IdleSpy's CLI can generate insightful graphs like:

- **`score`** – Efficiency score per goroutine (total lifetime vs blocked time).
- **`total-blocked-time`** – Cumulative blocked time per select case across all goroutines.
- **`avg-blocked-time`** – Average blocking duration per case across all goroutines.
- **`p90-blocked-time` / `p99-blocked-time`** – Long-tail blocking outliers across all goroutines.
- **`hits`** – Frequency of each case execution across across all goroutines.

> Note: Use these charts to identify bottlenecks, uncover starvation issues, and fine-tune your system's concurrency design.

#### Graph Example (`score`)

![Score Graph Example](score_graph_example.png)

This graph shows the efficiency score distribution across goroutines, where higher scores indicate better utilization (less blocking time relative to total lifetime).

## Installation

```bash
go get github.com/AlexsanderHamir/IdleSpy/tracker
go install github.com/AlexsanderHamir/IdleSpy/cmd/idlespy@latest
```

## Tracker Usage

IdleSpy provides a simple API to track goroutine behavior in your Go applications. Here's how to use it:

### Basic Usage

```go
import (
	"context"
	"log"
	"time"

	"github.com/AlexsanderHamir/IdleSpy/tracker"
)


	goroutineCount := 10
	gm := tracker.NewGoroutineManager()
	// "text" or "json": if "json" is selected, only the .internal.json file is generated.
	// If "text" is selected, both .internal.json (for parsing) and .visualization.txt are created.
	gm.FileType = "json"
	gm.Action = tracker.PrintAndSave // Save => save only // Print => print only
	gm.Wg.Add(goroutineCount)

// Simple example of tracking a worker goroutine
func processItems(gm *tracker.GoroutineManager,ctx context.Context, items <-chan string, results chan<- string) {

	// Start tracking this goroutine
	id := gm.TrackGoroutineStart()
	defer gm.TrackGoroutineEnd(id)

	for {
		select {
		case item, ok := <-items:
			if !ok {
				return // channel closed
			}
			startTime := time.Now()
			select {
			case results <- process(item):
				// Track successful processing
				gm.TrackSelectCase("process_success", time.Since(startTime), id)
			case <-ctx.Done():
				// Track cancellation
				gm.TrackSelectCase("process_cancelled", time.Since(startTime), id)
				return
			}

		case <-ctx.Done():
			gm.TrackSelectCase("worker_cancelled", 0)
			return
		}
	}
}

// Wait for all goroutines to finish so you can print the final results (blocking)
err := gm.Done()
	if err != nil {
	 t.Errorf("Error saving stats: %v", err)
}


```

## CLI Usage

Use the CLI tool to generate visualizations of your tracking data:

```bash
# Generate efficiency score chart
idlespy -chart score

# View blocking time distribution across select cases
idlespy -chart total-blocked-time
```

> Note: Run `idlespy -help` for more.

### 📊 Understanding the Statistics

The tracker generates detailed runtime statistics and saves them to a .internal.json file, and optionally to a .visualization.txt file if enabled. An example of the generated data format is shown below:

#### 🧵 Goroutine 35

- **Lifetime:** `19.88s`
- **Total Select Blocked Time:** `2.36s`

**Select Case Statistics:**

| Case Name          | Hits | Total Blocked Time | Avg Blocked Time | 90th %ile | 99th %ile |
| ------------------ | ---- | ------------------ | ---------------- | --------- | --------- |
| `slow_path_output` | 51   | 2.36s              | 46.34ms          | 88.35ms   | 93.23ms   |
| `batch_timeout`    | 58   | 274.84µs           | 4.74µs           | 8.17µs    | 12.83µs   |

### Best Practices

1. **Meaningful Case Names**: Use descriptive names for your select cases to make analysis easier
2. **Track All Cases**: Include tracking for all select cases, including timeouts and cancellations
3. **Consistent Timing**: Always measure from the start of the select statement
4. **Goroutine Management**: Create a new `GoroutineManager` for each logical component of your application

## Contributing

We welcome contributions! Before you start contributing, please ensure you have:

- **Go 1.24.3 or later** installed
- **Git** for version control
- Basic understanding of Go testing and benchmarking

### Quick Setup

```bash
# Fork and clone the repository
git clone https://github.com/AlexsanderHamir/GenPool.git
cd GenPool

# Install dependencies
go mod download
go mod tidy

# Run tests to verify setup
go test -v ./...
go test -bench=. ./...
```

### Development Guidelines

- Write tests for new functionality
- Run benchmarks to ensure no performance regressions
- Follow Go code style guidelines
- Update documentation for user-facing changes
- Ensure all tests pass before submitting PRs

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

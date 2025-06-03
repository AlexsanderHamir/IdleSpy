package main

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/AlexsanderHamir/IdleSpy/tracker"
)

// Data represents the structure of data flowing through our pipeline
type Data struct {
	ID        int
	Value     float64
	Timestamp time.Time
	Error     error
}

// Stage represents a processing stage in our pipeline
type Stage struct {
	name     string
	process  func(context.Context, <-chan Data) <-chan Data
	workers  int
	capacity int
}

// NewStage creates a new pipeline stage
func NewStage(name string, process func(context.Context, <-chan Data) <-chan Data, workers, capacity int) *Stage {
	return &Stage{
		name:     name,
		process:  process,
		workers:  workers,
		capacity: capacity,
	}
}

// Run executes the stage with multiple workers
func (s *Stage) Run(ctx context.Context, input <-chan Data) <-chan Data {
	output := make(chan Data, s.capacity)
	var wg sync.WaitGroup

	for i := range s.workers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d started in stage %s", workerID, s.name)
			for data := range s.process(ctx, input) {
				select {
				case output <- data:
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(output)
		log.Printf("Stage %s completed", s.name)
	}()

	return output
}

// Pipeline represents our complete processing pipeline
type Pipeline struct {
	stages []*Stage
	stats  *tracker.GoroutineManager
}

// NewPipeline creates a new pipeline with the given stages
func NewPipeline(stages ...*Stage) *Pipeline {
	return &Pipeline{
		stages: stages,
		stats:  tracker.NewGoroutineManager(),
	}
}

// Execute runs the entire pipeline with all stages running concurrently
func (p *Pipeline) Execute(ctx context.Context, input <-chan Data) <-chan Data {
	// Create all stage channels upfront
	stageChannels := make([]chan Data, len(p.stages))
	for i := range p.stages {
		stageChannels[i] = make(chan Data, p.stages[i].capacity)
	}

	// Start all stages concurrently
	for i, stage := range p.stages {
		var stageInput <-chan Data
		if i == 0 {
			stageInput = input
		} else {
			stageInput = stageChannels[i-1]
		}

		// Start the stage in a goroutine
		go func(s *Stage, in <-chan Data, out chan<- Data, stageIndex int) {
			defer close(out)
			var wg sync.WaitGroup

			// Start all workers for this stage
			for w := range s.workers {
				wg.Add(1)
				// Calculate unique worker ID for this stage
				workerID := (stageIndex * s.workers) + w
				go func(workerID int) {
					p.stats.TrackGoroutineStart()
					defer func() {
						p.stats.TrackGoroutineEnd()
						wg.Done()
					}()
					log.Printf("Worker %d started in stage %s", workerID, s.name)

					// Process data using the stage's process function
					for data := range s.process(ctx, in) {
						startTime := time.Now()
						select {
						case out <- data:
							p.stats.TrackSelectCase("output_send", time.Since(startTime))
						case <-ctx.Done():
							p.stats.TrackSelectCase("context_done", time.Since(startTime))
							return
						}
					}
				}(workerID)
			}

			wg.Wait()
			log.Printf("Stage %s completed", s.name)
		}(stage, stageInput, stageChannels[i], i)
	}

	// Return the output channel of the last stage
	return stageChannels[len(stageChannels)-1]
}

// GetPipelineStats returns statistics for all goroutines in the pipeline
func (p *Pipeline) GetPipelineStats() map[int]*tracker.GoroutineStats {
	return p.stats.GetAllStats()
}

// PrintPipelineStats prints a summary of the pipeline performance
func (p *Pipeline) PrintPipelineStats() {
	stats := p.GetPipelineStats()
	log.Println("\nPipeline Performance Statistics:")
	log.Println("===============================")

	for goroutineID, stat := range stats {
		log.Printf("\nGoroutine %d:", goroutineID)
		log.Printf("  Lifetime: %v", stat.GetGoroutineLifetime())
		log.Printf("  Total Select Time: %v", stat.GetTotalSelectTime())

		log.Println("  Select Case Statistics:")
		for caseName, caseStats := range stat.GetSelectStats() {
			log.Printf("    %s:", caseName)
			log.Printf("      Hits: %d", caseStats.GetCaseHits())
			log.Printf("      Total Time: %v", caseStats.GetCaseTime())
			if caseStats.GetCaseHits() > 0 {
				log.Printf("      Average Time: %v", caseStats.GetCaseTime()/time.Duration(caseStats.GetCaseHits()))
			}
		}
	}
}

// generateData produces random data for our pipeline
func generateData(ctx context.Context, count int, r *rand.Rand, stats *tracker.GoroutineManager) <-chan Data {
	output := make(chan Data)
	go func() {
		defer close(output)
		stats.TrackGoroutineStart()
		defer stats.TrackGoroutineEnd()

		for i := range count {
			startTime := time.Now()
			select {
			case <-ctx.Done():
				stats.TrackSelectCase("context_done", time.Since(startTime))
				return
			case output <- Data{
				ID:        i,
				Value:     r.Float64() * 100,
				Timestamp: time.Now(),
			}:
				stats.TrackSelectCase("data_send", time.Since(startTime))
				// Simulate variable processing time
				time.Sleep(time.Duration(r.Intn(10)) * time.Millisecond)
			}
		}
	}()
	return output
}

func main() {
	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create local random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create pipeline first
	pipeline := NewPipeline()

	// Create pipeline stages
	stages := []*Stage{
		NewStage("filter", func(ctx context.Context, input <-chan Data) <-chan Data {
			return filterStage(ctx, input, r, pipeline.stats)
		}, 40, 20),
		NewStage("transform", func(ctx context.Context, input <-chan Data) <-chan Data {
			return transformStage(ctx, input, r, pipeline.stats)
		}, 40, 20),
		NewStage("aggregate", func(ctx context.Context, input <-chan Data) <-chan Data {
			return aggregateStage(ctx, input, r, pipeline.stats)
		}, 40, 20),
	}

	// Add stages to pipeline
	pipeline.stages = stages

	// Create and execute pipeline
	input := generateData(ctx, 10000, r, pipeline.stats)
	output := pipeline.Execute(ctx, input)

	// Process results
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pipeline.stats.TrackGoroutineStart()
		defer pipeline.stats.TrackGoroutineEnd()

		for data := range output {
			log.Printf("Processed data: ID=%d, Value=%.2f, Time=%v",
				data.ID, data.Value, data.Timestamp)
			pipeline.stats.TrackSelectCase("result_processing", time.Since(data.Timestamp))
		}
	}()

	// Wait for completion or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Pipeline completed successfully")
		pipeline.PrintPipelineStats() // Print statistics after completion
	case <-ctx.Done():
		log.Println("Pipeline timed out")
		pipeline.PrintPipelineStats() // Print statistics even on timeout
	}
}

// filterStage processes data with a single worker
func filterStage(ctx context.Context, input <-chan Data, r *rand.Rand, stats *tracker.GoroutineManager) <-chan Data {
	output := make(chan Data)
	go func() {
		defer close(output)
		stats.TrackGoroutineStart()
		defer stats.TrackGoroutineEnd()

		for data := range input {
			// Simulate some processing time
			time.Sleep(time.Duration(r.Intn(50)) * time.Millisecond)

			// Filter out values less than 20
			if data.Value >= 20 {
				startTime := time.Now()
				select {
				case output <- data:
					stats.TrackSelectCase("filter_output", time.Since(startTime))
				case <-ctx.Done():
					stats.TrackSelectCase("context_done", time.Since(startTime))
					return
				}
			}
		}
	}()
	return output
}

// transformStage processes data with a single worker
func transformStage(ctx context.Context, input <-chan Data, r *rand.Rand, stats *tracker.GoroutineManager) <-chan Data {
	output := make(chan Data)
	go func() {
		defer close(output)
		stats.TrackGoroutineStart()
		defer stats.TrackGoroutineEnd()

		for data := range input {
			// Simulate complex transformation
			time.Sleep(time.Duration(r.Intn(150)) * time.Millisecond)

			// Apply some transformation
			transformed := Data{
				ID:        data.ID,
				Value:     data.Value * 1.5,
				Timestamp: time.Now(),
			}

			startTime := time.Now()
			select {
			case output <- transformed:
				stats.TrackSelectCase("transform_output", time.Since(startTime))
			case <-ctx.Done():
				stats.TrackSelectCase("context_done", time.Since(startTime))
				return
			}
		}
	}()
	return output
}

// aggregateStage processes data with a single worker
func aggregateStage(ctx context.Context, input <-chan Data, r *rand.Rand, stats *tracker.GoroutineManager) <-chan Data {
	output := make(chan Data)
	batchSize := 5
	batch := make([]Data, 0, batchSize)

	go func() {
		defer close(output)
		stats.TrackGoroutineStart()
		defer stats.TrackGoroutineEnd()

		for data := range input {
			batch = append(batch, data)

			if len(batch) >= batchSize {
				// Simulate batch processing
				time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)

				// Calculate average of the batch
				var sum float64
				for _, d := range batch {
					sum += d.Value
				}
				avg := sum / float64(len(batch))

				aggregated := Data{
					ID:        batch[0].ID,
					Value:     avg,
					Timestamp: time.Now(),
				}

				startTime := time.Now()
				select {
				case output <- aggregated:
					stats.TrackSelectCase("aggregate_batch_output", time.Since(startTime))
					batch = batch[:0]
				case <-ctx.Done():
					stats.TrackSelectCase("context_done", time.Since(startTime))
					return
				}
			}
		}

		// Process remaining items
		if len(batch) > 0 {
			var sum float64
			for _, d := range batch {
				sum += d.Value
			}
			avg := sum / float64(len(batch))

			aggregated := Data{
				ID:        batch[0].ID,
				Value:     avg,
				Timestamp: time.Now(),
			}

			startTime := time.Now()
			select {
			case output <- aggregated:
				stats.TrackSelectCase("aggregate_remaining_output", time.Since(startTime))
			case <-ctx.Done():
				stats.TrackSelectCase("context_done", time.Since(startTime))
				return
			}
		}
	}()
	return output
}

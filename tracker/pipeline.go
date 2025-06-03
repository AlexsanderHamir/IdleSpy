package main

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
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
}

// NewPipeline creates a new pipeline with the given stages
func NewPipeline(stages ...*Stage) *Pipeline {
	return &Pipeline{stages: stages}
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
		go func(s *Stage, in <-chan Data, out chan<- Data) {
			defer close(out)
			var wg sync.WaitGroup

			// Start all workers for this stage
			for w := range s.workers {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()
					log.Printf("Worker %d started in stage %s", workerID, s.name)

					// Process data using the stage's process function
					for data := range s.process(ctx, in) {
						select {
						case out <- data:
						case <-ctx.Done():
							return
						}
					}
				}(w)
			}

			wg.Wait()
			log.Printf("Stage %s completed", s.name)
		}(stage, stageInput, stageChannels[i])
	}

	// Return the output channel of the last stage
	return stageChannels[len(stageChannels)-1]
}

// generateData produces random data for our pipeline
func generateData(ctx context.Context, count int, r *rand.Rand) <-chan Data {
	output := make(chan Data)
	go func() {
		defer close(output)
		for i := range count {
			select {
			case <-ctx.Done():
				return
			case output <- Data{
				ID:        i,
				Value:     r.Float64() * 100,
				Timestamp: time.Now(),
			}:
				// Simulate variable processing time
				time.Sleep(time.Duration(r.Intn(100)) * time.Millisecond)
			}
		}
	}()
	return output
}

// filterStage processes data with a single worker
func filterStage(ctx context.Context, input <-chan Data, r *rand.Rand) <-chan Data {
	output := make(chan Data)
	go func() {
		defer close(output)
		for data := range input {
			// Simulate some processing time
			time.Sleep(time.Duration(r.Intn(50)) * time.Millisecond)

			// Filter out values less than 20
			if data.Value >= 20 {
				select {
				case output <- data:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return output
}

// transformStage processes data with a single worker
func transformStage(ctx context.Context, input <-chan Data, r *rand.Rand) <-chan Data {
	output := make(chan Data)
	go func() {
		defer close(output)
		for data := range input {
			// Simulate complex transformation
			time.Sleep(time.Duration(r.Intn(150)) * time.Millisecond)

			// Apply some transformation
			transformed := Data{
				ID:        data.ID,
				Value:     data.Value * 1.5,
				Timestamp: time.Now(),
			}

			select {
			case output <- transformed:
			case <-ctx.Done():
				return
			}
		}
	}()
	return output
}

// aggregateStage processes data with a single worker
func aggregateStage(ctx context.Context, input <-chan Data, r *rand.Rand) <-chan Data {
	output := make(chan Data)
	batchSize := 5
	batch := make([]Data, 0, batchSize)

	go func() {
		defer close(output)
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

				select {
				case output <- aggregated:
					batch = batch[:0]
				case <-ctx.Done():
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

			select {
			case output <- aggregated:
			case <-ctx.Done():
				return
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

	// Create pipeline stages
	stages := []*Stage{
		NewStage("filter", func(ctx context.Context, input <-chan Data) <-chan Data { return filterStage(ctx, input, r) }, 3, 100),
		NewStage("transform", func(ctx context.Context, input <-chan Data) <-chan Data { return transformStage(ctx, input, r) }, 2, 100),
		NewStage("aggregate", func(ctx context.Context, input <-chan Data) <-chan Data { return aggregateStage(ctx, input, r) }, 1, 100),
	}

	// Create and execute pipeline
	pipeline := NewPipeline(stages...)
	input := generateData(ctx, 100, r)
	output := pipeline.Execute(ctx, input)

	// Process results
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range output {
			log.Printf("Processed data: ID=%d, Value=%.2f, Time=%v",
				data.ID, data.Value, data.Timestamp)
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
	case <-ctx.Done():
		log.Println("Pipeline timed out")
	}
}

package examples

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/AlexsanderHamir/IdleSpy/tracker"
)

// WorkItem represents a unit of work that can be processed in different ways
type WorkItem struct {
	ID        int
	Priority  int    // Higher priority items get processed differently
	Type      string // Type of work: "fast", "slow", or "batch"
	Value     float64
	Timestamp time.Time
}

// Worker represents a worker that can handle different types of work
type Worker struct {
	id    int
	stats *tracker.GoroutineManager
	r     *rand.Rand
}

// NewWorker creates a new worker instance
func NewWorker(id int, stats *tracker.GoroutineManager, r *rand.Rand) *Worker {
	return &Worker{
		id:    id,
		stats: stats,
		r:     r,
	}
}

// Process handles different types of work items with varying processing times
func (w *Worker) Process(ctx context.Context, input <-chan WorkItem) <-chan WorkItem {
	output := make(chan WorkItem)

	go func() {
		defer close(output)
		w.stats.TrackGoroutineStart()
		defer w.stats.TrackGoroutineEnd()

		for item := range input {
			// Each worker can choose different processing paths based on the work item
			switch item.Type {
			case "fast":
				w.processFastPath(ctx, item, output)
			case "slow":
				w.processSlowPath(ctx, item, output)
			case "batch":
				w.processBatchPath(ctx, item, output)
			default:
				// Default processing path
				w.processDefaultPath(ctx, item, output)
			}
		}
	}()

	return output
}

// processFastPath handles high-priority items quickly
func (w *Worker) processFastPath(ctx context.Context, item WorkItem, output chan<- WorkItem) {
	// Simulate quick processing
	time.Sleep(time.Duration(w.r.Intn(20)) * time.Millisecond)

	startTime := time.Now()
	select {
	case output <- WorkItem{
		ID:        item.ID,
		Priority:  item.Priority,
		Type:      "fast",
		Value:     item.Value * 1.2, // Quick boost
		Timestamp: time.Now(),
	}:
		w.stats.TrackSelectCase("fast_path_output", time.Since(startTime))
	case <-ctx.Done():
		w.stats.TrackSelectCase("fast_path_context_done", time.Since(startTime))
		return
	}
}

// processSlowPath handles complex items that need more time
func (w *Worker) processSlowPath(ctx context.Context, item WorkItem, output chan<- WorkItem) {
	// Simulate complex processing
	time.Sleep(time.Duration(w.r.Intn(200)) * time.Millisecond)

	// Multiple select cases to simulate different processing states
	processingDone := make(chan struct{})
	go func() {
		// Simulate some complex computation
		time.Sleep(time.Duration(w.r.Intn(100)) * time.Millisecond)
		close(processingDone)
	}()

	startTime := time.Now()
	select {
	case <-processingDone:
		// Processing completed, now try to send
		select {
		case output <- WorkItem{
			ID:        item.ID,
			Priority:  item.Priority,
			Type:      "slow",
			Value:     item.Value * 2.0, // Bigger transformation
			Timestamp: time.Now(),
		}:
			w.stats.TrackSelectCase("slow_path_output", time.Since(startTime))
		case <-ctx.Done():
			w.stats.TrackSelectCase("slow_path_context_done", time.Since(startTime))
			return
		}
	case <-ctx.Done():
		w.stats.TrackSelectCase("slow_path_early_context_done", time.Since(startTime))
		return
	}
}

// processBatchPath handles items that need to be batched
func (w *Worker) processBatchPath(ctx context.Context, item WorkItem, output chan<- WorkItem) {
	// Simulate batch processing
	batchSize := 3
	batch := make([]WorkItem, 0, batchSize)
	batch = append(batch, item)

	// Try to collect more items for the batch
	timeout := time.After(50 * time.Millisecond)
	collecting := true

	for collecting {
		startTime := time.Now()
		select {
		case <-timeout:
			collecting = false
			w.stats.TrackSelectCase("batch_timeout", time.Since(startTime))
		case <-ctx.Done():
			w.stats.TrackSelectCase("batch_context_done", time.Since(startTime))
			return
		default:
			// Process the batch
			if len(batch) >= batchSize {
				collecting = false
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Process the collected batch
	time.Sleep(time.Duration(w.r.Intn(150)) * time.Millisecond)

	var sum float64
	for _, b := range batch {
		sum += b.Value
	}
	avg := sum / float64(len(batch))

	startTime := time.Now()
	select {
	case output <- WorkItem{
		ID:        item.ID,
		Priority:  item.Priority,
		Type:      "batch",
		Value:     avg,
		Timestamp: time.Now(),
	}:
		w.stats.TrackSelectCase("batch_output", time.Since(startTime))
	case <-ctx.Done():
		w.stats.TrackSelectCase("batch_final_context_done", time.Since(startTime))
		return
	}
}

// processDefaultPath handles regular items
func (w *Worker) processDefaultPath(ctx context.Context, item WorkItem, output chan<- WorkItem) {
	// Simulate regular processing
	time.Sleep(time.Duration(w.r.Intn(50)) * time.Millisecond)

	startTime := time.Now()
	select {
	case output <- WorkItem{
		ID:        item.ID,
		Priority:  item.Priority,
		Type:      "default",
		Value:     item.Value * 1.5,
		Timestamp: time.Now(),
	}:
		w.stats.TrackSelectCase("default_path_output", time.Since(startTime))
	case <-ctx.Done():
		w.stats.TrackSelectCase("default_path_context_done", time.Since(startTime))
		return
	}
}

// generateWorkItems creates a stream of work items with different types
func generateWorkItems(ctx context.Context, count int, r *rand.Rand, stats *tracker.GoroutineManager) <-chan WorkItem {
	output := make(chan WorkItem)
	workTypes := []string{"fast", "slow", "batch", "default"}

	go func() {
		defer close(output)
		stats.TrackGoroutineStart()
		defer stats.TrackGoroutineEnd()

		for i := 0; i < count; i++ {
			workType := workTypes[r.Intn(len(workTypes))]
			priority := r.Intn(10)

			startTime := time.Now()
			select {
			case output <- WorkItem{
				ID:        i,
				Priority:  priority,
				Type:      workType,
				Value:     r.Float64() * 100,
				Timestamp: time.Now(),
			}:
				stats.TrackSelectCase("work_item_generation", time.Since(startTime))
				// Vary the generation rate
				time.Sleep(time.Duration(r.Intn(30)) * time.Millisecond)
			case <-ctx.Done():
				stats.TrackSelectCase("work_generation_context_done", time.Since(startTime))
				return
			}
		}
	}()

	return output
}

// RunSelectsExample demonstrates goroutines handling multiple select cases with varying processing times
func RunSelectsExample() {
	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create local random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create stats tracker
	stats := tracker.NewGoroutineManager()

	// Create workers
	numWorkers := 5
	workers := make([]*Worker, numWorkers)
	for i := range numWorkers {
		workers[i] = NewWorker(i, stats, r)
	}

	// Create work distribution channel
	workChan := generateWorkItems(ctx, 1000, r, stats)

	// Create worker output channels
	workerOutputs := make([]<-chan WorkItem, numWorkers)
	for i, worker := range workers {
		workerOutputs[i] = worker.Process(ctx, workChan)
	}

	// Merge worker outputs
	mergedOutput := make(chan WorkItem)
	var wg sync.WaitGroup

	// Start output merger
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(mergedOutput)
		stats.TrackGoroutineStart()
		defer stats.TrackGoroutineEnd()

		// Create a done channel for each worker output
		done := make([]chan struct{}, len(workerOutputs))
		for i := range done {
			done[i] = make(chan struct{})
		}

		// Start a goroutine for each worker output
		for i, ch := range workerOutputs {
			wg.Add(1)
			go func(ch <-chan WorkItem, done chan struct{}, workerID int) {
				defer wg.Done()
				defer close(done)
				stats.TrackGoroutineStart()
				defer stats.TrackGoroutineEnd()

				for item := range ch {
					startTime := time.Now()
					select {
					case mergedOutput <- item:
						stats.TrackSelectCase("merge_output", time.Since(startTime))
					case <-ctx.Done():
						stats.TrackSelectCase("merge_context_done", time.Since(startTime))
						return
					}
				}
			}(ch, done[i], i)
		}

		// Wait for all workers to finish
		for _, d := range done {
			<-d
		}
	}()

	// Process results
	var processedCount int
	for item := range mergedOutput {
		processedCount++
		log.Printf("Processed work item: ID=%d, Type=%s, Priority=%d, Value=%.2f, Time=%v",
			item.ID, item.Type, item.Priority, item.Value, item.Timestamp)
	}

	log.Printf("Total processed items: %d", processedCount)

	// Print statistics
	log.Println("\nWorker Performance Statistics:")
	log.Println("============================")
	for goroutineID, stat := range stats.GetAllStats() {
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

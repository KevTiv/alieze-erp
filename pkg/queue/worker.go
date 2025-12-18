package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Worker processes jobs from the queue
type Worker struct {
	id       int
	queue    *PostgresQueue
	stopChan chan struct{}
}

// NewWorker creates a new worker
func NewWorker(id int, queue *PostgresQueue) *Worker {
	return &Worker{
		id:       id,
		queue:    queue,
		stopChan: make(chan struct{}),
	}
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context) {
	workerID := fmt.Sprintf("worker-%d", w.id)
	log.Printf("[Worker %d] Starting worker %s", w.id, workerID)

	queueNames := []string{"default", "critical", "low"}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			log.Printf("[Worker %d] Stopping worker %s", w.id, workerID)
			return

		case <-ctx.Done():
			log.Printf("[Worker %d] Context cancelled, stopping worker %s", w.id, workerID)
			return

		case <-ticker.C:
			// Try to dequeue a job
			job, err := w.queue.Dequeue(ctx, workerID, queueNames)
			if err != nil {
				log.Printf("[Worker %d] Error dequeuing job: %v", w.id, err)
				continue
			}

			if job == nil {
				// No jobs available
				continue
			}

			// Process the job
			w.processJob(ctx, job)
		}
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) processJob(ctx context.Context, job *Job) {
	log.Printf("[Worker %d] Processing job %s (type: %s)", w.id, job.ID, job.JobType)

	startTime := time.Now()

	// Get handler for job type
	handler, ok := w.queue.handlers[job.JobType]
	if !ok {
		log.Printf("[Worker %d] No handler registered for job type: %s", w.id, job.JobType)
		errMsg := fmt.Sprintf("no handler registered for job type: %s", job.JobType)
		if err := w.queue.Fail(ctx, job.ID, errMsg); err != nil {
			log.Printf("[Worker %d] Failed to mark job as failed: %v", w.id, err)
		}
		return
	}

	// Execute handler
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		log.Printf("[Worker %d] Failed to marshal payload: %v", w.id, err)
		if err := w.queue.Fail(ctx, job.ID, err.Error()); err != nil {
			log.Printf("[Worker %d] Failed to mark job as failed: %v", w.id, err)
		}
		return
	}

	err = handler(ctx, payloadBytes)
	if err != nil {
		log.Printf("[Worker %d] Job failed: %v", w.id, err)

		// Check if we should retry
		if job.AttemptCount < job.MaxAttempts-1 {
			log.Printf("[Worker %d] Retrying job %s (attempt %d/%d)", w.id, job.ID, job.AttemptCount+1, job.MaxAttempts)
			if err := w.queue.Retry(ctx, job.ID); err != nil {
				log.Printf("[Worker %d] Failed to retry job: %v", w.id, err)
			}
		} else {
			log.Printf("[Worker %d] Max attempts reached, marking job as failed", w.id)
			if err := w.queue.Fail(ctx, job.ID, err.Error()); err != nil {
				log.Printf("[Worker %d] Failed to mark job as failed: %v", w.id, err)
			}
		}
		return
	}

	// Mark job as completed
	duration := time.Since(startTime)
	result := map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
		"completed_at": time.Now(),
	}

	if err := w.queue.Complete(ctx, job.ID, result); err != nil {
		log.Printf("[Worker %d] Failed to mark job as completed: %v", w.id, err)
		return
	}

	log.Printf("[Worker %d] Job completed successfully in %v", w.id, duration)
}

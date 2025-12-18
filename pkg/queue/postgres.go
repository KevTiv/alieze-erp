package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostgresQueue implements Queue interface using PostgreSQL UNLOGGED tables
//
// IMPORTANT: The queue uses UNLOGGED tables for better performance.
// This means queue data is lost on database crash, but can be re-enqueued.
// See migration 20250118000000_convert_queue_to_unlogged.sql
type PostgresQueue struct {
	db       *sql.DB
	handlers map[string]HandlerFunc
	workers  []*Worker
	stopChan chan struct{}
}

// NewPostgresQueue creates a new PostgresQueue instance
func NewPostgresQueue(db *sql.DB) *PostgresQueue {
	return &PostgresQueue{
		db:       db,
		handlers: make(map[string]HandlerFunc),
		workers:  make([]*Worker, 0),
		stopChan: make(chan struct{}),
	}
}

// Enqueue adds a job to the queue
func (q *PostgresQueue) Enqueue(ctx context.Context, job Job) error {
	return q.enqueue(ctx, &job)
}

// EnqueueAt schedules a job to run at a specific time
func (q *PostgresQueue) EnqueueAt(ctx context.Context, job Job, scheduledAt time.Time) error {
	job.ScheduledAt = scheduledAt
	return q.enqueue(ctx, &job)
}

// EnqueueWithDelay schedules a job to run after a delay
func (q *PostgresQueue) EnqueueWithDelay(ctx context.Context, job Job, delay time.Duration) error {
	job.ScheduledAt = time.Now().Add(delay)
	return q.enqueue(ctx, &job)
}

func (q *PostgresQueue) enqueue(ctx context.Context, job *Job) error {
	// Convert payload and metadata to JSONB
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	metadataJSON := []byte("{}")
	if job.Metadata != nil {
		metadataJSON, err = json.Marshal(job.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Call SQL function enqueue_job
	query := `SELECT enqueue_job($1, $2, $3, $4, $5, $6, $7, $8)`

	var jobID uuid.UUID
	err = q.db.QueryRowContext(ctx, query,
		job.OrganizationID,
		job.QueueName,
		job.JobType,
		payloadJSON,
		job.Priority,
		job.ScheduledAt,
		job.MaxAttempts,
		metadataJSON,
	).Scan(&jobID)

	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	job.ID = jobID
	return nil
}

// Dequeue retrieves the next job from the queue
func (q *PostgresQueue) Dequeue(ctx context.Context, workerID string, queueNames []string) (*Job, error) {
	if len(queueNames) == 0 {
		queueNames = []string{"default"}
	}

	// Convert queue names to PostgreSQL array format
	queueNamesJSON, err := json.Marshal(queueNames)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal queue names: %w", err)
	}

	// Call SQL function dequeue_job
	query := `SELECT * FROM dequeue_job($1, $2::text[], $3)`

	lockDuration := 300 // 5 minutes in seconds

	var job Job
	var payloadJSON, resultJSON, metadataJSON []byte

	err = q.db.QueryRowContext(ctx, query, workerID, string(queueNamesJSON), lockDuration).Scan(
		&job.ID,
		&job.OrganizationID,
		&job.QueueName,
		&job.JobType,
		&payloadJSON,
		&resultJSON,
		&job.ErrorMessage,
		&job.Priority,
		&job.ScheduledAt,
		&job.Status,
		&job.AttemptCount,
		&job.MaxAttempts,
		&job.WorkerID,
		&job.LockedAt,
		&job.LockedUntil,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.UpdatedAt,
		&metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if len(resultJSON) > 0 {
		if err := json.Unmarshal(resultJSON, &job.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &job.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &job, nil
}

// Complete marks a job as completed
func (q *PostgresQueue) Complete(ctx context.Context, jobID uuid.UUID, result interface{}) error {
	resultJSON := []byte("{}")
	if result != nil {
		var err error
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	query := `
		UPDATE job_queue
		SET status = $1,
		    result = $2,
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $3
	`

	_, err := q.db.ExecContext(ctx, query, JobStatusCompleted, resultJSON, jobID)
	if err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	return nil
}

// Fail marks a job as failed
func (q *PostgresQueue) Fail(ctx context.Context, jobID uuid.UUID, errorMsg string) error {
	query := `
		UPDATE job_queue
		SET status = $1,
		    error_message = $2,
		    updated_at = NOW()
		WHERE id = $3
	`

	_, err := q.db.ExecContext(ctx, query, JobStatusFailed, errorMsg, jobID)
	if err != nil {
		return fmt.Errorf("failed to fail job: %w", err)
	}

	return nil
}

// Retry resets a job for retry
func (q *PostgresQueue) Retry(ctx context.Context, jobID uuid.UUID) error {
	query := `
		UPDATE job_queue
		SET status = $1,
		    attempt_count = attempt_count + 1,
		    worker_id = NULL,
		    locked_at = NULL,
		    locked_until = NULL,
		    scheduled_at = NOW() + interval '5 minutes',
		    updated_at = NOW()
		WHERE id = $2 AND attempt_count < max_attempts
	`

	_, err := q.db.ExecContext(ctx, query, JobStatusPending, jobID)
	if err != nil {
		return fmt.Errorf("failed to retry job: %w", err)
	}

	return nil
}

// Cancel cancels a job
func (q *PostgresQueue) Cancel(ctx context.Context, jobID uuid.UUID) error {
	query := `
		UPDATE job_queue
		SET status = $1,
		    updated_at = NOW()
		WHERE id = $2 AND status IN ('pending', 'processing')
	`

	_, err := q.db.ExecContext(ctx, query, JobStatusCancelled, jobID)
	if err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	return nil
}

// GetJob retrieves a job by ID
func (q *PostgresQueue) GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error) {
	query := `
		SELECT id, organization_id, queue_name, job_type, payload, result,
		       error_message, priority, scheduled_at, status, attempt_count,
		       max_attempts, worker_id, locked_at, locked_until, created_at,
		       started_at, completed_at, updated_at, metadata
		FROM job_queue
		WHERE id = $1
	`

	var job Job
	var payloadJSON, resultJSON, metadataJSON []byte

	err := q.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID,
		&job.OrganizationID,
		&job.QueueName,
		&job.JobType,
		&payloadJSON,
		&resultJSON,
		&job.ErrorMessage,
		&job.Priority,
		&job.ScheduledAt,
		&job.Status,
		&job.AttemptCount,
		&job.MaxAttempts,
		&job.WorkerID,
		&job.LockedAt,
		&job.LockedUntil,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.UpdatedAt,
		&metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if len(resultJSON) > 0 {
		if err := json.Unmarshal(resultJSON, &job.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &job.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &job, nil
}

// GetStats retrieves queue statistics
func (q *PostgresQueue) GetStats(ctx context.Context) (*QueueStats, error) {
	query := `
		SELECT queue_name, date, jobs_enqueued, jobs_completed, jobs_failed,
		       total_processing_time_ms
		FROM queue_stats
		WHERE date = CURRENT_DATE
		ORDER BY queue_name
	`

	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer rows.Close()

	stats := &QueueStats{}
	for rows.Next() {
		var s QueueStats
		err := rows.Scan(
			&s.QueueName,
			&s.Date,
			&s.JobsEnqueued,
			&s.JobsCompleted,
			&s.JobsFailed,
			&s.TotalProcessingMS,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}

		if s.JobsCompleted > 0 {
			s.AvgProcessingMS = s.TotalProcessingMS / int64(s.JobsCompleted)
		}

		// Aggregate stats
		stats.JobsEnqueued += s.JobsEnqueued
		stats.JobsCompleted += s.JobsCompleted
		stats.JobsFailed += s.JobsFailed
		stats.TotalProcessingMS += s.TotalProcessingMS
	}

	if stats.JobsCompleted > 0 {
		stats.AvgProcessingMS = stats.TotalProcessingMS / int64(stats.JobsCompleted)
	}

	return stats, nil
}

// RegisterHandler registers a handler for a job type
func (q *PostgresQueue) RegisterHandler(jobType string, handler HandlerFunc) {
	q.handlers[jobType] = handler
}

// Start starts the queue workers
func (q *PostgresQueue) Start(ctx context.Context, workerCount int) error {
	if workerCount <= 0 {
		workerCount = 1
	}

	for i := 0; i < workerCount; i++ {
		worker := NewWorker(i+1, q)
		q.workers = append(q.workers, worker)
		go worker.Start(ctx)
	}

	return nil
}

// Stop stops all queue workers
func (q *PostgresQueue) Stop() error {
	close(q.stopChan)

	for _, worker := range q.workers {
		worker.Stop()
	}

	return nil
}

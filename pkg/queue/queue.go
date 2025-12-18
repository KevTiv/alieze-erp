package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Queue defines the interface for job queue operations
//
// The PostgreSQL implementation uses UNLOGGED tables for performance:
// - Faster writes (no WAL overhead)
// - Data lost on crash (acceptable for queue systems)
// - Jobs can be re-enqueued if needed
type Queue interface {
	Enqueue(ctx context.Context, job Job) error
	EnqueueAt(ctx context.Context, job Job, scheduledAt time.Time) error
	EnqueueWithDelay(ctx context.Context, job Job, delay time.Duration) error
	Dequeue(ctx context.Context, workerID string, queueNames []string) (*Job, error)
	Complete(ctx context.Context, jobID uuid.UUID, result interface{}) error
	Fail(ctx context.Context, jobID uuid.UUID, errorMsg string) error
	Retry(ctx context.Context, jobID uuid.UUID) error
	Cancel(ctx context.Context, jobID uuid.UUID) error
	GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error)
	GetStats(ctx context.Context) (*QueueStats, error)
	Start(ctx context.Context, workerCount int) error
	Stop() error
	RegisterHandler(jobType string, handler HandlerFunc)
}

// Job represents a job in the queue
type Job struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty" db:"organization_id"`
	QueueName      string                 `json:"queue_name" db:"queue_name"`
	JobType        string                 `json:"job_type" db:"job_type"`
	Payload        map[string]interface{} `json:"payload" db:"payload"`
	Result         map[string]interface{} `json:"result,omitempty" db:"result"`
	ErrorMessage   *string                `json:"error_message,omitempty" db:"error_message"`
	Priority       int                    `json:"priority" db:"priority"`
	ScheduledAt    time.Time              `json:"scheduled_at" db:"scheduled_at"`
	Status         string                 `json:"status" db:"status"`
	AttemptCount   int                    `json:"attempt_count" db:"attempt_count"`
	MaxAttempts    int                    `json:"max_attempts" db:"max_attempts"`
	WorkerID       *string                `json:"worker_id,omitempty" db:"worker_id"`
	LockedAt       *time.Time             `json:"locked_at,omitempty" db:"locked_at"`
	LockedUntil    *time.Time             `json:"locked_until,omitempty" db:"locked_until"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	StartedAt      *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// JobStatus represents possible job statuses
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
	JobStatusCancelled  = "cancelled"
)

// QueueStats represents queue statistics
type QueueStats struct {
	QueueName        string    `json:"queue_name"`
	Date             time.Time `json:"date"`
	JobsEnqueued     int       `json:"jobs_enqueued"`
	JobsCompleted    int       `json:"jobs_completed"`
	JobsFailed       int       `json:"jobs_failed"`
	TotalProcessingMS int64     `json:"total_processing_ms"`
	AvgProcessingMS  int64     `json:"avg_processing_ms"`
}

// HandlerFunc is a function that processes a job
type HandlerFunc func(ctx context.Context, payload []byte) error

// JobOptions contains options for enqueueing a job
type JobOptions struct {
	OrganizationID *uuid.UUID
	QueueName      string
	Priority       int
	MaxAttempts    int
	ScheduledAt    *time.Time
	Metadata       map[string]interface{}
}

// DefaultOptions returns default job options
func DefaultOptions() *JobOptions {
	return &JobOptions{
		QueueName:   "default",
		Priority:    0,
		MaxAttempts: 3,
	}
}

// NewJob creates a new job with default values
func NewJob(jobType string, payload interface{}, opts *JobOptions) (*Job, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Convert payload to map
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
		return nil, err
	}

	scheduledAt := time.Now()
	if opts.ScheduledAt != nil {
		scheduledAt = *opts.ScheduledAt
	}

	return &Job{
		ID:             uuid.New(),
		OrganizationID: opts.OrganizationID,
		QueueName:      opts.QueueName,
		JobType:        jobType,
		Payload:        payloadMap,
		Priority:       opts.Priority,
		ScheduledAt:    scheduledAt,
		Status:         JobStatusPending,
		AttemptCount:   0,
		MaxAttempts:    opts.MaxAttempts,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       opts.Metadata,
	}, nil
}

// JobType constants for common job types
const (
	JobTypeGenerateEmbedding  = "generate_embedding"
	JobTypeProcessInvoice     = "process_invoice"
	JobTypeDuplicateDetection = "duplicate_detection"
	JobTypeEmailNotification  = "email_notification"
	JobTypeDataExport         = "data_export"
	JobTypeDocumentGeneration = "document_generation"
	JobTypeCalendarSync       = "calendar_sync"
	JobTypeActivityReminder   = "activity_reminder"
	JobTypeLeadAutoAssign     = "lead_auto_assign"
	JobTypeWorkflowExecute    = "workflow_execute"
)

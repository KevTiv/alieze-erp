package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// QueueManager defines interface for managing job queues
type QueueManager interface {
	// EnqueueJob enqueues a job for processing
	EnqueueJob(ctx context.Context, queueName string, jobType string, payload map[string]interface{}, options *ManagerJobOptions) (*uuid.UUID, error)

	// ScheduleJob schedules a job to run at a specific time
	ScheduleJob(ctx context.Context, queueName string, jobType string, payload map[string]interface{}, scheduledAt time.Time, options *ManagerJobOptions) (*uuid.UUID, error)

	// CancelJob cancels a queued or scheduled job
	CancelJob(ctx context.Context, jobID uuid.UUID) error

	// GetJobStatus retrieves the status of a job
	GetJobStatus(ctx context.Context, jobID uuid.UUID) (*ManagerJobStatus, error)

	// GetQueueStats retrieves statistics for a queue
	GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error)

	// StartWorker starts a worker for processing jobs
	StartWorker(ctx context.Context, queueName string, workerID string) error

	// StopWorker stops a worker
	StopWorker(ctx context.Context, workerID string) error
}

// ManagerJobOptions contains options for manager job creation
type ManagerJobOptions struct {
	Priority       int                    `json:"priority,omitempty"`
	Delay          *time.Duration         `json:"delay,omitempty"`
	ScheduledAt    *time.Time             `json:"scheduled_at,omitempty"`
	MaxRetries     *int                   `json:"max_retries,omitempty"`
	Timeout        *time.Duration         `json:"timeout,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty"`
	UserID         *uuid.UUID             `json:"user_id,omitempty"`
}

// ManagerJobStatus represents the status of a job from queue manager
type ManagerJobStatus struct {
	ID           uuid.UUID              `json:"id"`
	QueueName    string                 `json:"queue_name"`
	JobType      string                 `json:"job_type"`
	Status       string                 `json:"status"` // pending, processing, completed, failed, cancelled
	Payload      map[string]interface{} `json:"payload"`
	Result       map[string]interface{} `json:"result,omitempty"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   *int                   `json:"max_retries,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	WorkerID     *string                `json:"worker_id,omitempty"`
}

// QueueManagerAdapter wraps the existing Queue interface to implement QueueManager
type QueueManagerAdapter struct {
	queue Queue
}

// NewQueueManagerAdapter creates a new QueueManager using the existing Queue interface
func NewQueueManagerAdapter(queue Queue) QueueManager {
	return &QueueManagerAdapter{queue: queue}
}

// EnqueueJob enqueues a job for processing
func (a *QueueManagerAdapter) EnqueueJob(ctx context.Context, queueName string, jobType string, payload map[string]interface{}, options *ManagerJobOptions) (*uuid.UUID, error) {
	jobID := uuid.New()

	job := Job{
		ID:             jobID,
		QueueName:      queueName,
		JobType:        jobType,
		Payload:        payload,
		OrganizationID: options.OrganizationID,
	}

	if options != nil {
		if options.Delay != nil && *options.Delay > 0 {
			return &jobID, a.queue.EnqueueWithDelay(ctx, job, *options.Delay)
		}
		if options.ScheduledAt != nil {
			return &jobID, a.queue.EnqueueAt(ctx, job, *options.ScheduledAt)
		}
	}

	return &jobID, a.queue.Enqueue(ctx, job)
}

// ScheduleJob schedules a job to run at a specific time
func (a *QueueManagerAdapter) ScheduleJob(ctx context.Context, queueName string, jobType string, payload map[string]interface{}, scheduledAt time.Time, options *ManagerJobOptions) (*uuid.UUID, error) {
	jobID := uuid.New()

	job := Job{
		ID:             jobID,
		QueueName:      queueName,
		JobType:        jobType,
		Payload:        payload,
		OrganizationID: options.OrganizationID,
	}

	return &jobID, a.queue.EnqueueAt(ctx, job, scheduledAt)
}

// CancelJob cancels a queued or scheduled job
func (a *QueueManagerAdapter) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	return a.queue.Cancel(ctx, jobID)
}

// GetJobStatus retrieves the status of a job
func (a *QueueManagerAdapter) GetJobStatus(ctx context.Context, jobID uuid.UUID) (*ManagerJobStatus, error) {
	job, err := a.queue.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	status := &ManagerJobStatus{
		ID:         job.ID,
		QueueName:  job.QueueName,
		JobType:    job.JobType,
		Status:     job.Status,
		Payload:    job.Payload,
		Result:     job.Result,
		RetryCount: job.AttemptCount,
		CreatedAt:  job.CreatedAt,
		WorkerID:   job.WorkerID,
	}

	if job.ErrorMessage != nil {
		status.ErrorMessage = job.ErrorMessage
	}

	return status, nil
}

// GetQueueStats retrieves statistics for a queue
func (a *QueueManagerAdapter) GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error) {
	return a.queue.GetStats(ctx)
}

// StartWorker starts a worker for processing jobs
func (a *QueueManagerAdapter) StartWorker(ctx context.Context, queueName string, workerID string) error {
	// TODO: Implement worker management
	return nil
}

// StopWorker stops a worker
func (a *QueueManagerAdapter) StopWorker(ctx context.Context, workerID string) error {
	// TODO: Implement worker management
	return nil
}

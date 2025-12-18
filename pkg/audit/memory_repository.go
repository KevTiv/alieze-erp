package audit

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryAuditLogRepository implements AuditLogRepository using in-memory storage
type MemoryAuditLogRepository struct {
	logs []*AuditLog
	mu   sync.Mutex
}

// NewMemoryAuditLogRepository creates a new in-memory audit log repository
func NewMemoryAuditLogRepository() *MemoryAuditLogRepository {
	return &MemoryAuditLogRepository{
		logs: make([]*AuditLog, 0),
	}
}

// Create adds a new audit log entry
func (r *MemoryAuditLogRepository) Create(ctx context.Context, log *AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Add timestamp if not set
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// Generate ID if not set
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	r.logs = append(r.logs, log)
	return nil
}

// Find retrieves audit logs matching the filter criteria
func (r *MemoryAuditLogRepository) Find(ctx context.Context, filter *AuditLogFilter) ([]*AuditLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*AuditLog

	for _, log := range r.logs {
		// Apply filter criteria
		if filter.UserID != nil && log.UserID != *filter.UserID {
			continue
		}
		if filter.OrganizationID != nil && log.OrganizationID != *filter.OrganizationID {
			continue
		}
		if filter.Module != nil && log.Module != *filter.Module {
			continue
		}
		if filter.Table != nil && log.Table != *filter.Table {
			continue
		}
		if filter.Operation != nil && log.Operation != *filter.Operation {
			continue
		}
		if filter.Allowed != nil && log.Allowed != *filter.Allowed {
			continue
		}
		if filter.StartTime != nil && log.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && log.Timestamp.After(*filter.EndTime) {
			continue
		}

		result = append(result, log)

		// Apply limit if specified
		if filter.Limit != nil && len(result) >= *filter.Limit {
			break
		}
	}

	return result, nil
}

// Count returns the number of audit logs matching the filter criteria
func (r *MemoryAuditLogRepository) Count(ctx context.Context, filter *AuditLogFilter) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0

	for _, log := range r.logs {
		// Apply filter criteria
		if filter.UserID != nil && log.UserID != *filter.UserID {
			continue
		}
		if filter.OrganizationID != nil && log.OrganizationID != *filter.OrganizationID {
			continue
		}
		if filter.Module != nil && log.Module != *filter.Module {
			continue
		}
		if filter.Table != nil && log.Table != *filter.Table {
			continue
		}
		if filter.Operation != nil && log.Operation != *filter.Operation {
			continue
		}
		if filter.Allowed != nil && log.Allowed != *filter.Allowed {
			continue
		}
		if filter.StartTime != nil && log.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && log.Timestamp.After(*filter.EndTime) {
			continue
		}

		count++
	}

	return count, nil
}

// DeleteOlderThan removes audit logs older than the specified duration
func (r *MemoryAuditLogRepository) DeleteOlderThan(ctx context.Context, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-duration)
	var newLogs []*AuditLog

	for _, log := range r.logs {
		if log.Timestamp.After(cutoff) {
			newLogs = append(newLogs, log)
		}
	}

	r.logs = newLogs
	return nil
}

// GetAllLogs returns all audit logs (for testing purposes)
func (r *MemoryAuditLogRepository) GetAllLogs() []*AuditLog {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Return a copy to avoid race conditions
	copy := make([]*AuditLog, len(r.logs))
	for i, log := range r.logs {
		copy[i] = log
	}
	return copy
}

// ClearAllLogs clears all audit logs (for testing purposes)
func (r *MemoryAuditLogRepository) ClearAllLogs() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logs = []*AuditLog{}
}

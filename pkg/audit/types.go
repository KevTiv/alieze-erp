package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a permission audit record
type AuditLog struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Module        string    `json:"module" db:"module"`
	Table         string    `json:"table" db:"table"`
	Operation     string    `json:"operation" db:"operation"`
	Permission    string    `json:"permission" db:"permission"`
	Allowed       bool      `json:"allowed" db:"allowed"`
	Query         string    `json:"query" db:"query"`
	IPAddress     string    `json:"ip_address" db:"ip_address"`
	UserAgent     string    `json:"user_agent" db:"user_agent"`
	Metadata      string    `json:"metadata" db:"metadata"`
}

// AuditLogFilter defines filtering criteria for audit logs
type AuditLogFilter struct {
	UserID        *uuid.UUID
	OrganizationID *uuid.UUID
	Module        *string
	Table         *string
	Operation     *string
	Allowed       *bool
	StartTime     *time.Time
	EndTime       *time.Time
	Limit         *int
}

// AuditLogRepository interface for audit log storage
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	Find(ctx context.Context, filter *AuditLogFilter) ([]*AuditLog, error)
	Count(ctx context.Context, filter *AuditLogFilter) (int, error)
	DeleteOlderThan(ctx context.Context, duration time.Duration) error
}

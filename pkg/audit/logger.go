package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Logger interface for audit logging
type Logger interface {
	LogPermissionCheck(ctx context.Context, userID, orgID uuid.UUID,
		module, table, operation, permission string, allowed bool, query string) error

	LogDatabaseOperation(ctx context.Context, userID, orgID uuid.UUID,
		module, table, operation, query string, success bool) error

	LogSecurityEvent(ctx context.Context, userID, orgID uuid.UUID,
		eventType, description string, metadata map[string]interface{}) error
}

// AuditLogger implements the Logger interface
type AuditLogger struct {
	repository AuditLogRepository
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(repository AuditLogRepository) *AuditLogger {
	return &AuditLogger{
		repository: repository,
	}
}

// LogPermissionCheck logs a permission check event
func (l *AuditLogger) LogPermissionCheck(ctx context.Context, userID, orgID uuid.UUID,
	module, table, operation, permission string, allowed bool, query string) error {

	// Extract additional context information
	ipAddress := "unknown"
	userAgent := "unknown"

	// In a real implementation, you would extract these from the HTTP request context
	// For now, we'll use default values

	// Create audit log entry
	log := &AuditLog{
		ID:            uuid.New(),
		Timestamp:     time.Now(),
		UserID:        userID,
		OrganizationID: orgID,
		Module:        module,
		Table:         table,
		Operation:     operation,
		Permission:    permission,
		Allowed:       allowed,
		Query:         query,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Metadata:      "{}",
	}

	// Store the audit log
	return l.repository.Create(ctx, log)
}

// LogDatabaseOperation logs a database operation event
func (l *AuditLogger) LogDatabaseOperation(ctx context.Context, userID, orgID uuid.UUID,
	module, table, operation, query string, success bool) error {

	// Create audit log entry
	log := &AuditLog{
		ID:            uuid.New(),
		Timestamp:     time.Now(),
		UserID:        userID,
		OrganizationID: orgID,
		Module:        module,
		Table:         table,
		Operation:     operation,
		Permission:    "database:" + operation,
		Allowed:       success,
		Query:         query,
		IPAddress:     "unknown",
		UserAgent:     "unknown",
		Metadata:      "{}",
	}

	// Store the audit log
	return l.repository.Create(ctx, log)
}

// LogSecurityEvent logs a general security event
func (l *AuditLogger) LogSecurityEvent(ctx context.Context, userID, orgID uuid.UUID,
	eventType, description string, metadata map[string]interface{}) error {

	// Create audit log entry
	log := &AuditLog{
		ID:            uuid.New(),
		Timestamp:     time.Now(),
		UserID:        userID,
		OrganizationID: orgID,
		Module:        "security",
		Table:         eventType,
		Operation:     "event",
		Permission:    "security:" + eventType,
		Allowed:       true,
		Query:         description,
		IPAddress:     "unknown",
		UserAgent:     "unknown",
		Metadata:      "{}",
	}

	// Store the audit log
	return l.repository.Create(ctx, log)
}

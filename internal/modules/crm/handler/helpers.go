package handler

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Context keys for organization and user data
type contextKey string

const (
	contextKeyOrganizationID contextKey = "organization_id"
	contextKeyUserID         contextKey = "user_id"
)

// getOrganizationIDFromContext extracts the organization ID from context
func getOrganizationIDFromContext(ctx context.Context) (uuid.UUID, error) {
	orgID, ok := ctx.Value(contextKeyOrganizationID).(uuid.UUID)
	if !ok {
		// Try string conversion as fallback
		orgIDStr, ok := ctx.Value(contextKeyOrganizationID).(string)
		if !ok {
			return uuid.Nil, fmt.Errorf("organization ID not found in context")
		}
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid organization ID format: %w", err)
		}
	}
	return orgID, nil
}

// getUserIDFromContext extracts the user ID from context
func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(contextKeyUserID).(uuid.UUID)
	if !ok {
		// Try string conversion as fallback
		userIDStr, ok := ctx.Value(contextKeyUserID).(string)
		if !ok {
			return uuid.Nil, fmt.Errorf("user ID not found in context")
		}
		var err error
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
		}
	}
	return userID, nil
}

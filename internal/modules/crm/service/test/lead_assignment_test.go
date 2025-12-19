package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/internal/testutils"
)

// TestLeadAssignmentSuccess tests the core lead assignment functionality
func TestLeadAssignmentSuccess(t *testing.T) {
	// Setup
	repo := testutils.NewMockAssignmentRuleRepository()
	auth := testutils.NewMockAuthService()
	service := service.NewAssignmentRuleService(repo, auth)
	ctx := context.Background()

	// Test data
	leadID := uuid.Must(uuid.NewV7())
	assigneeID := uuid.Must(uuid.NewV7())
	assigneeName := "John Doe"
	conditions := map[string]interface{}{
		"country": "US",
		"source":  "web",
	}

	// Mock the lead retrieval
	repo.WithGetLeadFunc(func(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
		require.Equal(t, leadID, id)
		return &types.Lead{
			ID:         leadID,
			AssignedTo: nil, // Not assigned yet
			Status:     "new",
		}, nil
	})

	// Mock the assignee selection
	repo.WithGetNextAssigneeFunc(func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
		require.Equal(t, "leads", targetModel)
		return assigneeID, assigneeName, nil
	})

	// Mock the assignment
	repo.WithAssignLeadFunc(func(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error {
		require.Equal(t, leadID, leadID)
		require.Equal(t, assigneeID, userID)
		require.Equal(t, "auto_assignment", reason)
		return nil
	})

	// Execute
	result, err := service.AssignLead(ctx, leadID, conditions)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, leadID, result.LeadID)
	require.Equal(t, assigneeID, result.AssignedToID)
	require.Equal(t, assigneeName, result.AssignedToName)
	require.Equal(t, "auto_assignment", result.Reason)
	require.True(t, result.Changed)
}

// TestLeadAlreadyAssigned tests the case where lead is already assigned to the same user
func TestLeadAlreadyAssigned(t *testing.T) {
	// Setup
	repo := testutils.NewMockAssignmentRuleRepository()
	auth := testutils.NewMockAuthService()
	service := service.NewAssignmentRuleService(repo, auth)
	ctx := context.Background()

	// Test data
	leadID := uuid.Must(uuid.NewV7())
	assigneeID := uuid.Must(uuid.NewV7())
	assigneeName := "John Doe"
	conditions := map[string]interface{}{
		"country": "US",
	}

	// Mock the lead retrieval - already assigned
	repo.WithGetLeadFunc(func(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
		return &types.Lead{
			ID:         leadID,
			AssignedTo: &assigneeID, // Already assigned
			Status:     "new",
		}, nil
	})

	// Mock the assignee selection - same assignee
	repo.WithGetNextAssigneeFunc(func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
		return assigneeID, assigneeName, nil
	})

	// Execute
	result, err := service.AssignLead(ctx, leadID, conditions)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, leadID, result.LeadID)
	require.Equal(t, assigneeID, result.AssignedToID)
	require.Equal(t, assigneeName, result.AssignedToName)
	require.Equal(t, "already_assigned", result.Reason)
	require.False(t, result.Changed)
}

// TestLeadReassignment tests the case where lead is reassigned to a different user
func TestLeadReassignment(t *testing.T) {
	// Setup
	repo := testutils.NewMockAssignmentRuleRepository()
	auth := testutils.NewMockAuthService()
	service := service.NewAssignmentRuleService(repo, auth)
	ctx := context.Background()

	// Test data
	leadID := uuid.Must(uuid.NewV7())
	oldAssigneeID := uuid.Must(uuid.NewV7())
	newAssigneeID := uuid.Must(uuid.NewV7())
	assigneeName := "Jane Smith"
	conditions := map[string]interface{}{
		"country": "US",
	}

	// Mock the lead retrieval - assigned to different user
	repo.WithGetLeadFunc(func(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
		return &types.Lead{
			ID:         leadID,
			AssignedTo: &oldAssigneeID, // Assigned to different user
			Status:     "new",
		}, nil
	})

	// Mock the assignee selection - new assignee
	repo.WithGetNextAssigneeFunc(func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
		return newAssigneeID, assigneeName, nil
	})

	// Mock the assignment
	repo.WithAssignLeadFunc(func(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error {
		require.Equal(t, newAssigneeID, userID)
		require.Equal(t, "reassignment", reason)
		return nil
	})

	// Execute
	result, err := service.AssignLead(ctx, leadID, conditions)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, leadID, result.LeadID)
	require.Equal(t, newAssigneeID, result.AssignedToID)
	require.Equal(t, assigneeName, result.AssignedToName)
	require.Equal(t, "reassignment", result.Reason)
	require.True(t, result.Changed)
}

// TestLeadAssignmentNoAssigneeFound tests the case where no suitable assignee is found
func TestLeadAssignmentNoAssigneeFound(t *testing.T) {
	// Setup
	repo := testutils.NewMockAssignmentRuleRepository()
	auth := testutils.NewMockAuthService()
	service := service.NewAssignmentRuleService(repo, auth)
	ctx := context.Background()

	// Test data
	leadID := uuid.Must(uuid.NewV7())
	conditions := map[string]interface{}{
		"country": "XX", // No assignee for this country
	}

	// Mock the lead retrieval
	repo.WithGetLeadFunc(func(ctx context.Context, id uuid.UUID) (*types.Lead, error) {
		return &types.Lead{
			ID:         leadID,
			AssignedTo: nil,
			Status:     "new",
		}, nil
	})

	// Mock the assignee selection - no assignee found
	repo.WithGetNextAssigneeFunc(func(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
		return uuid.Nil, "", nil
	})

	// Execute
	result, err := service.AssignLead(ctx, leadID, conditions)

	// Assert
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "no suitable assignee found")
}

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/pkg/events"

	"github.com/google/uuid"
)

// SalesTeamService handles sales team business logic
type SalesTeamService struct {
	repo        types.SalesTeamRepository
	authService AuthService
	eventBus    *events.Bus
	logger      *slog.Logger
}

func NewSalesTeamService(repo types.SalesTeamRepository, authService AuthService, eventBus *events.Bus) *SalesTeamService {
	return &SalesTeamService{
		repo:        repo,
		authService: authService,
		eventBus:    eventBus,
		logger:      slog.Default().With("service", "sales-team"),
	}
}

func (s *SalesTeamService) CreateSalesTeam(ctx context.Context, req types.SalesTeamCreateRequest) (*types.SalesTeam, error) {
	// Validation
	if err := s.validateSalesTeam(req); err != nil {
		return nil, fmt.Errorf("invalid sales team: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:sales_teams:create"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Set user
	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate ID
	teamID := uuid.New()

	// Create team
	team := types.SalesTeam{
		ID:             teamID,
		OrganizationID: orgID,
		CompanyID:      req.CompanyID,
		Name:           req.Name,
		Code:           req.Code,
		TeamLeaderID:   req.TeamLeaderID,
		MemberIDs:      req.MemberIDs,
		IsActive:       req.IsActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      &userID,
		UpdatedBy:      &userID,
	}

	created, err := s.repo.Create(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("failed to create sales team: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.sales_team.created", created)

	s.logger.Info("Created sales team", "team_id", created.ID, "name", created.Name)

	return created, nil
}

func (s *SalesTeamService) GetSalesTeam(ctx context.Context, id uuid.UUID) (*types.SalesTeam, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:sales_teams:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get the team
	team, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales team: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if team.OrganizationID != orgID {
		return nil, fmt.Errorf("sales team does not belong to organization: %w", errors.New("access denied"))
	}

	return team, nil
}

func (s *SalesTeamService) ListSalesTeams(ctx context.Context, filter types.SalesTeamFilter) ([]*types.SalesTeam, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:sales_teams:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Set organization filter
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	filter.OrganizationID = orgID

	// List teams
	teams, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales teams: %w", err)
	}

	return teams, nil
}

func (s *SalesTeamService) UpdateSalesTeam(ctx context.Context, id uuid.UUID, req types.SalesTeamUpdateRequest) (*types.SalesTeam, error) {
	// Validation
	if err := s.validateSalesTeamUpdate(req); err != nil {
		return nil, fmt.Errorf("invalid sales team update: %w", err)
	}

	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:sales_teams:update"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing team to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing sales team: %w", err)
	}

	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("sales team does not belong to organization: %w", errors.New("access denied"))
	}

	// Set user
	userID, err := s.authService.GetUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Build update
	team := types.SalesTeam{
		ID:             id,
		OrganizationID: orgID,
		CompanyID:      req.CompanyID,
		Name:           *req.Name,
		Code:           req.Code,
		TeamLeaderID:   req.TeamLeaderID,
		MemberIDs:      *req.MemberIDs,
		IsActive:       *req.IsActive,
		UpdatedAt:      time.Now(),
		UpdatedBy:      &userID,
	}

	// Update
	updated, err := s.repo.Update(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("failed to update sales team: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.sales_team.updated", updated)

	s.logger.Info("Updated sales team", "team_id", updated.ID, "name", updated.Name)

	return updated, nil
}

func (s *SalesTeamService) DeleteSalesTeam(ctx context.Context, id uuid.UUID) error {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:sales_teams:delete"); err != nil {
		return fmt.Errorf("permission denied: %w", err)
	}

	// Verify organization access
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Get existing team to verify organization
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing sales team: %w", err)
	}

	if existing.OrganizationID != orgID {
		return fmt.Errorf("sales team does not belong to organization: %w", errors.New("access denied"))
	}

	// Delete
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete sales team: %w", err)
	}

	// Event
	s.eventBus.Publish(ctx, "crm.sales_team.deleted", existing)

	s.logger.Info("Deleted sales team", "team_id", id)

	return nil
}

func (s *SalesTeamService) GetSalesTeamsByMember(ctx context.Context, memberID uuid.UUID) ([]types.SalesTeam, error) {
	// Permission check
	if err := s.authService.CheckPermission(ctx, "crm:sales_teams:read"); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Get teams by member
	teams, err := s.repo.FindByMember(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales teams by member: %w", err)
	}

	// Filter by organization
	orgID, err := s.authService.GetOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	var filteredTeams []types.SalesTeam
	for _, team := range teams {
		if team.OrganizationID == orgID {
			filteredTeams = append(filteredTeams, team)
		}
	}

	return filteredTeams, nil
}

func (s *SalesTeamService) validateSalesTeam(req types.SalesTeamCreateRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	if len(req.Name) > 255 {
		return errors.New("name must be 255 characters or less")
	}

	if req.Code != nil && len(*req.Code) > 50 {
		return errors.New("code must be 50 characters or less")
	}

	return nil
}

func (s *SalesTeamService) validateSalesTeamUpdate(req types.SalesTeamUpdateRequest) error {
	if req.Name != nil {
		if *req.Name == "" {
			return errors.New("name cannot be empty")
		}

		if len(*req.Name) > 255 {
			return errors.New("name must be 255 characters or less")
		}
	}

	if req.Code != nil && len(*req.Code) > 50 {
		return errors.New("code must be 50 characters or less")
	}

	return nil
}

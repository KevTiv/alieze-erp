package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"alieze-erp/internal/modules/crm/repository"
	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/pkg/rules"

	"github.com/google/uuid"
)

// LeadService handles business logic for leads
type LeadService struct {
	leadRepository *repository.LeadRepository
	ruleEngine     *rules.RuleEngine
	logger         *slog.Logger
}

// NewLeadService creates a new LeadService
type NewLeadServiceOptions struct {
	LeadRepository *repository.LeadRepository
	RuleEngine     *rules.RuleEngine
	Logger         *slog.Logger
}

func NewLeadService(options NewLeadServiceOptions) *LeadService {
	if options.Logger == nil {
		options.Logger = slog.Default().With("service", "lead")
	}

	return &LeadService{
		leadRepository: options.LeadRepository,
		ruleEngine:     options.RuleEngine,
		logger:         options.Logger,
	}
}

// CreateLead creates a new lead
func (s *LeadService) CreateLead(ctx context.Context, orgID uuid.UUID, lead types.LeadEnhancedCreateRequest) (*types.LeadEnhanced, error) {
	// Validate input
	if lead.Name == "" {
		return nil, errors.New("lead name is required")
	}

	// Remove organization ID extraction since it's now passed as parameter

	// Set default values
	if lead.LeadType == "" {
		lead.LeadType = types.LeadTypeLead
	}

	if lead.Priority == "" {
		lead.Priority = types.LeadPriorityMedium
	}

	if lead.Active == false {
		lead.Active = true
	}

	if lead.Probability == 0 {
		lead.Probability = 10 // Default 10% probability
	}

	// Set timestamps
	now := time.Now()

	// Create the lead in database
	createdLead := &types.LeadEnhanced{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		CompanyID:       lead.CompanyID,
		Name:            lead.Name,
		ContactName:     lead.ContactName,
		Email:           lead.Email,
		Phone:           lead.Phone,
		Mobile:          lead.Mobile,
		ContactID:       lead.ContactID,
		UserID:          lead.UserID,
		TeamID:          lead.TeamID,
		LeadType:        lead.LeadType,
		StageID:         lead.StageID,
		Priority:        lead.Priority,
		SourceID:        lead.SourceID,
		MediumID:        lead.MediumID,
		CampaignID:      lead.CampaignID,
		ExpectedRevenue: lead.ExpectedRevenue,
		Probability:     lead.Probability,
		RecurringRevenue: lead.RecurringRevenue,
		RecurringPlan:   lead.RecurringPlan,
		DateOpen:        lead.DateOpen,
		DateClosed:      lead.DateClosed,
		DateDeadline:    lead.DateDeadline,
		DateLastStageUpdate: lead.DateLastStageUpdate,
		Active:          lead.Active,
		WonStatus:       lead.WonStatus,
		LostReasonID:    lead.LostReasonID,
		Street:          lead.Street,
		Street2:         lead.Street2,
		City:            lead.City,
		StateID:         lead.StateID,
		Zip:             lead.Zip,
		CountryID:       lead.CountryID,
		Website:         lead.Website,
		Description:     lead.Description,
		TagIDs:          lead.TagIDs,
		Color:           lead.Color,
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       nil, // Will be set by auth middleware in real implementation
		UpdatedBy:       nil, // Will be set by auth middleware in real implementation
		CustomFields:    lead.CustomFields,
		Metadata:        lead.Metadata,
	}

	// Apply business rules
	var err error
	if s.ruleEngine != nil {
		err = s.ruleEngine.Validate(ctx, "lead", createdLead)
		if err != nil {
			return nil, fmt.Errorf("failed to apply business rules: %w", err)
		}
		s.logger.Info("Applied business rules")
	}

	// Create lead in database
	err = s.leadRepository.Create(ctx, createdLead)
	if err != nil {
		return nil, fmt.Errorf("failed to create lead: %w", err)
	}

	s.logger.Info("Lead created successfully", "lead_id", createdLead.ID, "name", createdLead.Name)

	return createdLead, nil
}

// GetLead retrieves a lead by ID
func (s *LeadService) GetLead(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*types.LeadEnhanced, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid lead ID")
	}

	lead, err := s.leadRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}
	if lead == nil {
		return nil, fmt.Errorf("lead not found")
	}

	// Check organization access using the provided orgID parameter
	if lead.OrganizationID != orgID {
		return nil, fmt.Errorf("lead does not belong to organization")
	}

	return lead, nil
}

// UpdateLead updates an existing lead
func (s *LeadService) UpdateLead(ctx context.Context, orgID uuid.UUID, id uuid.UUID, update types.LeadEnhancedUpdateRequest) (*types.LeadEnhanced, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid lead ID")
	}

	// Get existing lead
	existingLead, err := s.GetLead(ctx, orgID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing lead: %w", err)
	}

	// Apply updates
	if update.Name != nil && *update.Name != "" {
		existingLead.Name = *update.Name
	}
	if update.ContactName != nil {
		existingLead.ContactName = update.ContactName
	}
	if update.Email != nil {
		existingLead.Email = update.Email
	}
	if update.Phone != nil {
		existingLead.Phone = update.Phone
	}
	if update.Mobile != nil {
		existingLead.Mobile = update.Mobile
	}
	if update.ContactID != nil {
		existingLead.ContactID = update.ContactID
	}
	if update.UserID != nil {
		existingLead.UserID = update.UserID
	}
	if update.TeamID != nil {
		existingLead.TeamID = update.TeamID
	}
	if update.LeadType != nil && *update.LeadType != "" {
		existingLead.LeadType = *update.LeadType
	}
	if update.StageID != nil {
		existingLead.StageID = update.StageID
	}
	if update.Priority != nil && *update.Priority != "" {
		existingLead.Priority = *update.Priority
	}
	if update.SourceID != nil {
		existingLead.SourceID = update.SourceID
	}
	if update.MediumID != nil {
		existingLead.MediumID = update.MediumID
	}
	if update.CampaignID != nil {
		existingLead.CampaignID = update.CampaignID
	}
	if update.ExpectedRevenue != nil {
		existingLead.ExpectedRevenue = update.ExpectedRevenue
	}
	if update.Probability != nil {
		existingLead.Probability = *update.Probability
	}
	if update.RecurringRevenue != nil {
		existingLead.RecurringRevenue = update.RecurringRevenue
	}
	if update.RecurringPlan != nil {
		existingLead.RecurringPlan = update.RecurringPlan
	}
	if update.DateOpen != nil {
		existingLead.DateOpen = update.DateOpen
	}
	if update.DateClosed != nil {
		existingLead.DateClosed = update.DateClosed
	}
	if update.DateDeadline != nil {
		existingLead.DateDeadline = update.DateDeadline
	}
	if update.DateLastStageUpdate != nil {
		existingLead.DateLastStageUpdate = update.DateLastStageUpdate
	}
	if update.Active != nil {
		existingLead.Active = *update.Active
	}
	if update.WonStatus != nil {
		existingLead.WonStatus = update.WonStatus
	}
	if update.LostReasonID != nil {
		existingLead.LostReasonID = update.LostReasonID
	}
	if update.Street != nil {
		existingLead.Street = update.Street
	}
	if update.Street2 != nil {
		existingLead.Street2 = update.Street2
	}
	if update.City != nil {
		existingLead.City = update.City
	}
	if update.StateID != nil {
		existingLead.StateID = update.StateID
	}
	if update.Zip != nil {
		existingLead.Zip = update.Zip
	}
	if update.CountryID != nil {
		existingLead.CountryID = update.CountryID
	}
	if update.Website != nil {
		existingLead.Website = update.Website
	}
	if update.Description != nil {
		existingLead.Description = update.Description
	}
	if update.TagIDs != nil {
		existingLead.TagIDs = *update.TagIDs
	}
	if update.Color != nil {
		existingLead.Color = update.Color
	}
	if update.CustomFields != nil {
		existingLead.CustomFields = update.CustomFields
	}
	if update.Metadata != nil {
		existingLead.Metadata = update.Metadata
	}

	existingLead.UpdatedAt = time.Now()

	// Apply business rules
	if s.ruleEngine != nil {
		err := s.ruleEngine.Validate(ctx, "lead", existingLead)
		if err != nil {
			return nil, fmt.Errorf("failed to apply business rules: %w", err)
		}
		s.logger.Info("Applied business rules")
	}

	// Update lead in database
	err = s.leadRepository.Update(ctx, existingLead)
	if err != nil {
		return nil, fmt.Errorf("failed to update lead: %w", err)
	}

	s.logger.Info("Lead updated successfully", "lead_id", existingLead.ID, "name", existingLead.Name)

	return existingLead, nil
}

// DeleteLead deletes a lead
func (s *LeadService) DeleteLead(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid lead ID")
	}

	// Get existing lead to check organization access
	existingLead, err := s.GetLead(ctx, orgID, id)
	if err != nil {
		return fmt.Errorf("failed to get existing lead: %w", err)
	}

	// Delete lead from database
	err = s.leadRepository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}

	s.logger.Info("Lead deleted successfully", "lead_id", id, "name", existingLead.Name)

	return nil
}

// ListLeads retrieves a list of leads with filtering
func (s *LeadService) ListLeads(ctx context.Context, orgID uuid.UUID, filter types.LeadEnhancedFilter) ([]*types.LeadEnhanced, error) {
	// Set organization filter
	filter.OrganizationID = orgID

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list leads: %w", err)
	}

	return leads, nil
}

// CountLeads counts leads matching the filter criteria
func (s *LeadService) CountLeads(ctx context.Context, orgID uuid.UUID, filter types.LeadEnhancedFilter) (int, error) {
	// Set organization filter
	filter.OrganizationID = orgID

	count, err := s.leadRepository.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count leads: %w", err)
	}

	return count, nil
}

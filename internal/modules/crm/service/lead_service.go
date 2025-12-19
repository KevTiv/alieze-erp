package service

import (
	"context"
	"errors"
	"time"

	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/pkg/rules"

	"github.com/google/uuid"
)

// LeadServiceOptions contains the dependencies for the LeadService
type LeadServiceOptions struct {
	LeadRepository types.LeadRepository
	RuleEngine     *rules.RuleEngine
}

// NewLeadService creates a new LeadService
type NewLeadServiceOptions struct {
	LeadRepository types.LeadRepository
	RuleEngine     *rules.RuleEngine
}

// LeadService provides lead management functionality
type LeadService struct {
	repo       types.LeadRepository
	ruleEngine *rules.RuleEngine
}

// NewLeadService creates a new LeadService instance
func NewLeadService(opts NewLeadServiceOptions) *LeadService {
	return &LeadService{
		repo:       opts.LeadRepository,
		ruleEngine: opts.RuleEngine,
	}
}

// CreateLead creates a new lead
func (s *LeadService) CreateLead(ctx context.Context, orgID uuid.UUID, req types.LeadCreateRequest) (types.Lead, error) {
	// Validate the request
	if req.Name == "" {
		return types.Lead{}, errors.New("lead name is required")
	}

	// Set default values
	if req.LeadType == "" {
		req.LeadType = types.LeadTypeLead
	}
	if req.Priority == "" {
		req.Priority = types.LeadPriorityMedium
	}
	if req.Probability == 0 {
		req.Probability = 10
	}

	// Create the lead entity
	lead := types.Lead{
		ID:               uuid.Must(uuid.NewV7()),
		OrganizationID:   orgID,
		CompanyID:        req.CompanyID,
		Name:             req.Name,
		ContactName:      req.ContactName,
		Email:            req.Email,
		Phone:            req.Phone,
		Mobile:           req.Mobile,
		ContactID:        req.ContactID,
		UserID:           req.UserID,
		TeamID:           req.TeamID,
		LeadType:         req.LeadType,
		StageID:          req.StageID,
		Priority:         req.Priority,
		SourceID:         req.SourceID,
		MediumID:         req.MediumID,
		CampaignID:       req.CampaignID,
		ExpectedRevenue:  req.ExpectedRevenue,
		Probability:      req.Probability,
		RecurringRevenue: req.RecurringRevenue,
		RecurringPlan:    req.RecurringPlan,
		DateOpen:         req.DateOpen,
		DateClosed:       req.DateClosed,
		DateDeadline:     req.DateDeadline,
		Active:           req.Active,
		Status:           req.Status,
		AssignedTo:       req.AssignedTo,
		WonStatus:        req.WonStatus,
		LostReasonID:     req.LostReasonID,
		Street:           req.Street,
		Street2:          req.Street2,
		City:             req.City,
		StateID:          req.StateID,
		Zip:              req.Zip,
		CountryID:        req.CountryID,
		Website:          req.Website,
		Description:      req.Description,
		TagIDs:           req.TagIDs,
		Color:            req.Color,
		CustomFields:     req.CustomFields,
		Metadata:         req.Metadata,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Apply assignment rules if available
	if s.ruleEngine != nil {
		// TODO: Implement assignment rule logic
	}

	// Create the lead in the repository
	createdLead, err := s.repo.Create(ctx, lead)
	if err != nil {
		return types.Lead{}, err
	}

	return *createdLead, nil
}

// GetLead retrieves a lead by ID
func (s *LeadService) GetLead(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (types.Lead, error) {
	lead, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return types.Lead{}, err
	}

	// Verify organization ownership
	if lead.OrganizationID != orgID {
		return types.Lead{}, errors.New("lead not found or access denied")
	}

	return *lead, nil
}

// UpdateLead updates an existing lead
func (s *LeadService) UpdateLead(ctx context.Context, orgID uuid.UUID, id uuid.UUID, req types.LeadUpdateRequest) (types.Lead, error) {
	// Get the existing lead
	existingLead, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return types.Lead{}, err
	}

	// Verify organization ownership
	if existingLead.OrganizationID != orgID {
		return types.Lead{}, errors.New("lead not found or access denied")
	}

	// Apply updates
	if req.Name != nil {
		existingLead.Name = *req.Name
	}
	if req.ContactName != nil {
		existingLead.ContactName = req.ContactName
	}
	if req.Email != nil {
		existingLead.Email = req.Email
	}
	if req.Phone != nil {
		existingLead.Phone = req.Phone
	}
	if req.Mobile != nil {
		existingLead.Mobile = req.Mobile
	}
	if req.ContactID != nil {
		existingLead.ContactID = req.ContactID
	}
	if req.UserID != nil {
		existingLead.UserID = req.UserID
	}
	if req.TeamID != nil {
		existingLead.TeamID = req.TeamID
	}
	if req.LeadType != nil {
		existingLead.LeadType = *req.LeadType
	}
	if req.StageID != nil {
		existingLead.StageID = req.StageID
	}
	if req.Priority != nil {
		existingLead.Priority = *req.Priority
	}
	if req.SourceID != nil {
		existingLead.SourceID = req.SourceID
	}
	if req.MediumID != nil {
		existingLead.MediumID = req.MediumID
	}
	if req.CampaignID != nil {
		existingLead.CampaignID = req.CampaignID
	}
	if req.ExpectedRevenue != nil {
		existingLead.ExpectedRevenue = req.ExpectedRevenue
	}
	if req.Probability != nil {
		existingLead.Probability = *req.Probability
	}
	if req.RecurringRevenue != nil {
		existingLead.RecurringRevenue = req.RecurringRevenue
	}
	if req.RecurringPlan != nil {
		existingLead.RecurringPlan = req.RecurringPlan
	}
	if req.DateOpen != nil {
		existingLead.DateOpen = req.DateOpen
	}
	if req.DateClosed != nil {
		existingLead.DateClosed = req.DateClosed
	}
	if req.DateDeadline != nil {
		existingLead.DateDeadline = req.DateDeadline
	}
	if req.Active != nil {
		existingLead.Active = *req.Active
	}
	if req.Status != nil {
		existingLead.Status = req.Status
	}
	if req.AssignedTo != nil {
		existingLead.AssignedTo = req.AssignedTo
	}
	if req.WonStatus != nil {
		existingLead.WonStatus = req.WonStatus
	}
	if req.LostReasonID != nil {
		existingLead.LostReasonID = req.LostReasonID
	}
	if req.Street != nil {
		existingLead.Street = req.Street
	}
	if req.Street2 != nil {
		existingLead.Street2 = req.Street2
	}
	if req.City != nil {
		existingLead.City = req.City
	}
	if req.StateID != nil {
		existingLead.StateID = req.StateID
	}
	if req.Zip != nil {
		existingLead.Zip = req.Zip
	}
	if req.CountryID != nil {
		existingLead.CountryID = req.CountryID
	}
	if req.Website != nil {
		existingLead.Website = req.Website
	}
	if req.Description != nil {
		existingLead.Description = req.Description
	}
	if req.TagIDs != nil {
		existingLead.TagIDs = *req.TagIDs
	}
	if req.Color != nil {
		existingLead.Color = req.Color
	}
	if req.CustomFields != nil {
		existingLead.CustomFields = req.CustomFields
	}
	if req.Metadata != nil {
		existingLead.Metadata = req.Metadata
	}

	existingLead.UpdatedAt = time.Now()

	// Update the lead in the repository
	updatedLead, err := s.repo.Update(ctx, *existingLead)
	if err != nil {
		return types.Lead{}, err
	}

	return *updatedLead, nil
}

// DeleteLead deletes a lead
func (s *LeadService) DeleteLead(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error {
	// Get the existing lead to verify ownership
	lead, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify organization ownership
	if lead.OrganizationID != orgID {
		return errors.New("lead not found or access denied")
	}

	return s.repo.Delete(ctx, id)
}

// ListLeads lists leads with filtering
func (s *LeadService) ListLeads(ctx context.Context, orgID uuid.UUID, filter types.LeadFilter) ([]types.Lead, error) {
	filter.OrganizationID = orgID
	return s.repo.FindAll(ctx, filter)
}

// CountLeads counts leads with filtering
func (s *LeadService) CountLeads(ctx context.Context, orgID uuid.UUID, filter types.LeadFilter) (int, error) {
	filter.OrganizationID = orgID
	return s.repo.Count(ctx, filter)
}

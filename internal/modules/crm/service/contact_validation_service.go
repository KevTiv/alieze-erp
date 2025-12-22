package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/crm/base"
	"github.com/KevTiv/alieze-erp/pkg/crm/errors"
	"github.com/google/uuid"
)

// ContactValidationService handles contact validation operations
type ContactValidationService struct {
	validationRepo repository.ContactValidationRepository
	contactRepo    types.ContactRepository
	authService    base.AuthService
	logger         *slog.Logger
}

// NewContactValidationService creates a new validation service
func NewContactValidationService(
	validationRepo repository.ContactValidationRepository,
	contactRepo types.ContactRepository,
	authService base.AuthService,
	logger *slog.Logger,
) *ContactValidationService {
	return &ContactValidationService{
		validationRepo: validationRepo,
		contactRepo:    contactRepo,
		authService:    authService,
		logger:         logger,
	}
}

// ValidateContact validates a contact against configured rules
func (s *ContactValidationService) ValidateContact(ctx context.Context, orgID uuid.UUID, req types.ContactValidateRequest) (*types.ContactValidationResult, error) {
	// Check authorization
	if err := s.authService.CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Get validation rules
	var rules []*types.ContactValidationRule
	var err error

	if len(req.RuleIDs) > 0 {
		// Get specific rules
		for _, ruleID := range req.RuleIDs {
			rule, err := s.validationRepo.GetValidationRule(ctx, ruleID)
			if err != nil {
				s.logger.Error("failed to get validation rule", "rule_id", ruleID, "error", err)
				continue
			}
			if rule != nil && rule.IsActive {
				rules = append(rules, rule)
			}
		}
	} else {
		// Get all active rules for organization
		filter := types.ValidationRuleFilter{
			OrganizationID: orgID,
			IsActive:       boolPtr(true),
		}
		rules, err = s.validationRepo.ListValidationRules(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to list validation rules: %w", err)
		}
	}

	// Apply validation rules
	result := &types.ContactValidationResult{
		IsValid:      true,
		Errors:       []types.ValidationError{},
		Warnings:     []types.ValidationWarning{},
		Suggestions:  []types.DataSuggestion{},
		QualityScore: s.calculateDataQualityScore(&req.ContactData),
	}

	for _, rule := range rules {
		err := s.applyValidationRule(&req.ContactData, rule, result)
		if err != nil {
			s.logger.Error("failed to apply validation rule", "rule_id", rule.ID, "error", err)
		}
	}

	// Generate suggestions
	suggestions := s.generateSuggestions(&req.ContactData)
	result.Suggestions = append(result.Suggestions, suggestions...)

	return result, nil
}

// GetValidationRules retrieves all active validation rules
func (s *ContactValidationService) GetValidationRules(ctx context.Context, orgID uuid.UUID) ([]*types.ContactValidationRule, error) {
	// Check authorization
	if err := s.authService.CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	filter := types.ValidationRuleFilter{
		OrganizationID: orgID,
		IsActive:       boolPtr(true),
	}

	return s.validationRepo.ListValidationRules(ctx, filter)
}

// EnrichContactData provides data enrichment suggestions (placeholder)
func (s *ContactValidationService) EnrichContactData(ctx context.Context, orgID uuid.UUID, req types.ContactEnrichRequest) (*types.ContactEnrichResponse, error) {
	// Check authorization
	if err := s.authService.CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Placeholder implementation - would integrate with external enrichment API
	response := &types.ContactEnrichResponse{
		SuggestedData: make(map[string]interface{}),
		Confidence:    0.0,
		Source:        "placeholder",
		Applied:       false,
	}

	s.logger.Info("enrichment requested (placeholder)", "org_id", orgID, "contact_id", req.ContactID)

	// TODO: Integrate with actual enrichment provider (Clearbit, FullContact, etc.)
	// For now, return empty suggestions

	return response, nil
}

// GetDataSuggestions provides suggestions for completing contact data
func (s *ContactValidationService) GetDataSuggestions(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) ([]types.DataSuggestion, error) {
	// Check authorization
	if err := s.authService.CheckOrganizationAccess(ctx, orgID); err != nil {
		return nil, errors.ErrOrganizationAccess
	}

	// Get contact
	contact, err := s.contactRepo.FindByID(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	if contact == nil {
		return nil, errors.ErrNotFound
	}

	return s.generateSuggestions(contact), nil
}

// applyValidationRule applies a single validation rule
func (s *ContactValidationService) applyValidationRule(contact *types.Contact, rule *types.ContactValidationRule, result *types.ContactValidationResult) error {
	fieldValue := s.getFieldValue(contact, rule.Field)

	switch rule.RuleType {
	case "required":
		if fieldValue == "" {
			s.addValidationIssue(result, rule, "Field is required")
		}

	case "format", "pattern":
		if fieldValue != "" {
			pattern, ok := rule.ValidationConfig["pattern"].(string)
			if ok {
				matched, err := regexp.MatchString(pattern, fieldValue)
				if err != nil {
					return fmt.Errorf("invalid regex pattern: %w", err)
				}
				if !matched {
					s.addValidationIssue(result, rule, "Field format is invalid")
				}
			}
		}

	case "length":
		if fieldValue != "" {
			minLen, hasMin := rule.ValidationConfig["min"].(float64)
			maxLen, hasMax := rule.ValidationConfig["max"].(float64)

			if hasMin && len(fieldValue) < int(minLen) {
				s.addValidationIssue(result, rule, fmt.Sprintf("Field must be at least %d characters", int(minLen)))
			}
			if hasMax && len(fieldValue) > int(maxLen) {
				s.addValidationIssue(result, rule, fmt.Sprintf("Field must be at most %d characters", int(maxLen)))
			}
		}

	case "enum":
		if fieldValue != "" {
			allowed, ok := rule.ValidationConfig["values"].([]interface{})
			if ok {
				valid := false
				for _, val := range allowed {
					if fieldValue == fmt.Sprint(val) {
						valid = true
						break
					}
				}
				if !valid {
					s.addValidationIssue(result, rule, "Field value is not in allowed list")
				}
			}
		}
	}

	return nil
}

// addValidationIssue adds an error or warning based on severity
func (s *ContactValidationService) addValidationIssue(result *types.ContactValidationResult, rule *types.ContactValidationRule, message string) {
	if rule.ErrorMessage != nil && *rule.ErrorMessage != "" {
		message = *rule.ErrorMessage
	}

	switch rule.Severity {
	case "error":
		result.Errors = append(result.Errors, types.ValidationError{
			Field:   rule.Field,
			Message: message,
			RuleID:  rule.ID.String(),
		})
		result.IsValid = false

	case "warning":
		result.Warnings = append(result.Warnings, types.ValidationWarning{
			Field:   rule.Field,
			Message: message,
		})
	}
}

// getFieldValue extracts field value from contact
func (s *ContactValidationService) getFieldValue(contact *types.Contact, field string) string {
	switch field {
	case "name":
		return contact.Name
	case "email":
		if contact.Email != nil {
			return *contact.Email
		}
	case "phone":
		if contact.Phone != nil {
			return *contact.Phone
		}
	case "street":
		if contact.Street != nil {
			return *contact.Street
		}
	case "city":
		if contact.City != nil {
			return *contact.City
		}
	}
	return ""
}

// calculateDataQualityScore calculates data quality score (0-100)
func (s *ContactValidationService) calculateDataQualityScore(contact *types.Contact) int {
	score := 0
	maxScore := 100

	// Core fields (60 points)
	if contact.Name != "" {
		score += 20
	}
	if contact.Email != nil && *contact.Email != "" {
		score += 20
	}
	if contact.Phone != nil && *contact.Phone != "" {
		score += 20
	}

	// Additional fields (40 points)
	if contact.Street != nil && *contact.Street != "" {
		score += 10
	}
	if contact.City != nil && *contact.City != "" {
		score += 10
	}
	if contact.StateID != nil {
		score += 10
	}
	if contact.CountryID != nil {
		score += 10
	}

	// Normalize to 0-100
	return (score * 100) / maxScore
}

// generateSuggestions generates data completion suggestions
func (s *ContactValidationService) generateSuggestions(contact *types.Contact) []types.DataSuggestion {
	var suggestions []types.DataSuggestion

	// Suggest email if missing
	if contact.Email == nil || *contact.Email == "" {
		suggestions = append(suggestions, types.DataSuggestion{
			Field:          "email",
			SuggestedValue: "Add email address for better communication",
			Confidence:     0.7,
			Source:         "pattern",
		})
	}

	// Suggest phone if missing
	if contact.Phone == nil || *contact.Phone == "" {
		suggestions = append(suggestions, types.DataSuggestion{
			Field:          "phone",
			SuggestedValue: "Add phone number for direct contact",
			Confidence:     0.7,
			Source:         "pattern",
		})
	}

	// Suggest address completion
	if (contact.Street == nil || *contact.Street == "") && (contact.City != nil && *contact.City != "") {
		suggestions = append(suggestions, types.DataSuggestion{
			Field:          "street",
			SuggestedValue: "Complete address with street information",
			Confidence:     0.6,
			Source:         "pattern",
		})
	}

	// Suggest email formatting
	if contact.Email != nil && *contact.Email != "" && !strings.Contains(*contact.Email, "@") {
		suggestions = append(suggestions, types.DataSuggestion{
			Field:          "email",
			CurrentValue:   *contact.Email,
			SuggestedValue: "Email format appears invalid",
			Confidence:     0.9,
			Source:         "pattern",
		})
	}

	return suggestions
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}

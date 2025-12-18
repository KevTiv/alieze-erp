package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// LeadRepository handles enhanced lead data operations
type LeadRepository struct {
	db *sql.DB
}

func NewLeadRepository(db *sql.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// Create inserts a new enhanced lead
func (r *LeadRepository) Create(ctx context.Context, lead *types.LeadEnhanced) error {
	if lead.ID == uuid.Nil {
		lead.ID = uuid.New()
	}

	if lead.OrganizationID == uuid.Nil {
		return errors.New("organization_id is required")
	}

	if lead.Name == "" {
		return errors.New("name is required")
	}

	if lead.LeadType == "" {
		lead.LeadType = types.LeadTypeLead
	}

	if lead.Priority == "" {
		lead.Priority = types.LeadPriorityMedium
	}

	if lead.Probability == 0 {
		lead.Probability = 10
	}

	if lead.Active == false {
		lead.Active = true
	}

	if lead.CreatedAt.IsZero() {
		lead.CreatedAt = time.Now()
	}

	if lead.UpdatedAt.IsZero() {
		lead.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO leads_enhanced (
			id, organization_id, company_id, name, contact_name, email, phone, mobile,
			contact_id, user_id, team_id, lead_type, stage_id, priority, source_id,
			medium_id, campaign_id, expected_revenue, probability, recurring_revenue,
			recurring_plan, date_open, date_closed, date_deadline, date_last_stage_update,
			active, won_status, lost_reason_id, street, street2, city, state_id, zip,
			country_id, website, description, tag_ids, color, created_at, updated_at,
			created_by, updated_by, deleted_at, custom_fields, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
			$29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41,
			$42, $43, $44, $45, $46, $47, $48, $49, $50
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		lead.ID,
		lead.OrganizationID,
		lead.CompanyID,
		lead.Name,
		lead.ContactName,
		lead.Email,
		lead.Phone,
		lead.Mobile,
		lead.ContactID,
		lead.UserID,
		lead.TeamID,
		lead.LeadType,
		lead.StageID,
		lead.Priority,
		lead.SourceID,
		lead.MediumID,
		lead.CampaignID,
		lead.ExpectedRevenue,
		lead.Probability,
		lead.RecurringRevenue,
		lead.RecurringPlan,
		lead.DateOpen,
		lead.DateClosed,
		lead.DateDeadline,
		lead.DateLastStageUpdate,
		lead.Active,
		lead.WonStatus,
		lead.LostReasonID,
		lead.Street,
		lead.Street2,
		lead.City,
		lead.StateID,
		lead.Zip,
		lead.CountryID,
		lead.Website,
		lead.Description,
		lead.TagIDs,
		lead.Color,
		lead.CreatedAt,
		lead.UpdatedAt,
		lead.CreatedBy,
		lead.UpdatedBy,
		lead.DeletedAt,
		lead.CustomFields,
		lead.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create enhanced lead: %w", err)
	}

	return nil
}

// FindByID retrieves an enhanced lead by its ID
func (r *LeadRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.LeadEnhanced, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid lead id")
	}

	query := `
		SELECT id, organization_id, company_id, name, contact_name, email, phone, mobile,
			contact_id, user_id, team_id, lead_type, stage_id, priority, source_id,
			medium_id, campaign_id, expected_revenue, probability, recurring_revenue,
			recurring_plan, date_open, date_closed, date_deadline, date_last_stage_update,
			active, won_status, lost_reason_id, street, street2, city, state_id, zip,
			country_id, website, description, tag_ids, color, created_at, updated_at,
			created_by, updated_by, deleted_at, custom_fields, metadata
		FROM leads_enhanced
		WHERE id = $1 AND deleted_at IS NULL
	`

	var lead types.LeadEnhanced
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lead.ID,
		&lead.OrganizationID,
		&lead.CompanyID,
		&lead.Name,
		&lead.ContactName,
		&lead.Email,
		&lead.Phone,
		&lead.Mobile,
		&lead.ContactID,
		&lead.UserID,
		&lead.TeamID,
		&lead.LeadType,
		&lead.StageID,
		&lead.Priority,
		&lead.SourceID,
		&lead.MediumID,
		&lead.CampaignID,
		&lead.ExpectedRevenue,
		&lead.Probability,
		&lead.RecurringRevenue,
		&lead.RecurringPlan,
		&lead.DateOpen,
		&lead.DateClosed,
		&lead.DateDeadline,
		&lead.DateLastStageUpdate,
		&lead.Active,
		&lead.WonStatus,
		&lead.LostReasonID,
		&lead.Street,
		&lead.Street2,
		&lead.City,
		&lead.StateID,
		&lead.Zip,
		&lead.CountryID,
		&lead.Website,
		&lead.Description,
		&lead.TagIDs,
		&lead.Color,
		&lead.CreatedAt,
		&lead.UpdatedAt,
		&lead.CreatedBy,
		&lead.UpdatedBy,
		&lead.DeletedAt,
		&lead.CustomFields,
		&lead.Metadata,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("enhanced lead not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get enhanced lead: %w", err)
	}

	return &lead, nil
}

// FindAll retrieves all enhanced leads with optional filters
func (r *LeadRepository) FindAll(ctx context.Context, filter types.LeadEnhancedFilter) ([]*types.LeadEnhanced, error) {
	query := `SELECT id, organization_id, company_id, name, contact_name, email, phone, mobile,
		contact_id, user_id, team_id, lead_type, stage_id, priority, source_id,
		medium_id, campaign_id, expected_revenue, probability, recurring_revenue,
		recurring_plan, date_open, date_closed, date_deadline, date_last_stage_update,
		active, won_status, lost_reason_id, street, street2, city, state_id, zip,
		country_id, website, description, tag_ids, color, created_at, updated_at,
		created_by, updated_by, deleted_at, custom_fields, metadata
		FROM leads_enhanced WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	var argIndex = 1

	// Organization filter (required)
	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, filter.OrganizationID)
	argIndex++

	// Name filter
	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	// Email filter
	if filter.Email != nil && *filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	// Phone filter
	if filter.Phone != nil && *filter.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Phone+"%")
		argIndex++
	}

	// Contact name filter
	if filter.ContactName != nil && *filter.ContactName != "" {
		conditions = append(conditions, fmt.Sprintf("contact_name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.ContactName+"%")
		argIndex++
	}

	// Mobile filter
	if filter.Mobile != nil && *filter.Mobile != "" {
		conditions = append(conditions, fmt.Sprintf("mobile ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Mobile+"%")
		argIndex++
	}

	// Company ID filter
	if filter.CompanyID != nil && *filter.CompanyID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("company_id = $%d", argIndex))
		args = append(args, *filter.CompanyID)
		argIndex++
	}

	// Contact ID filter
	if filter.ContactID != nil && *filter.ContactID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("contact_id = $%d", argIndex))
		args = append(args, *filter.ContactID)
		argIndex++
	}

	// User ID filter
	if filter.UserID != nil && *filter.UserID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	// Team ID filter
	if filter.TeamID != nil && *filter.TeamID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("team_id = $%d", argIndex))
		args = append(args, *filter.TeamID)
		argIndex++
	}

	// Lead type filter
	if filter.LeadType != nil && *filter.LeadType != "" {
		conditions = append(conditions, fmt.Sprintf("lead_type = $%d", argIndex))
		args = append(args, *filter.LeadType)
		argIndex++
	}

	// Stage ID filter
	if filter.StageID != nil && *filter.StageID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("stage_id = $%d", argIndex))
		args = append(args, *filter.StageID)
		argIndex++
	}

	// Priority filter
	if filter.Priority != nil && *filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}

	// Source ID filter
	if filter.SourceID != nil && *filter.SourceID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("source_id = $%d", argIndex))
		args = append(args, *filter.SourceID)
		argIndex++
	}

	// Medium ID filter
	if filter.MediumID != nil && *filter.MediumID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("medium_id = $%d", argIndex))
		args = append(args, *filter.MediumID)
		argIndex++
	}

	// Campaign ID filter
	if filter.CampaignID != nil && *filter.CampaignID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("campaign_id = $%d", argIndex))
		args = append(args, *filter.CampaignID)
		argIndex++
	}

	// Expected revenue range filter
	if filter.ExpectedRevenueMin != nil {
		conditions = append(conditions, fmt.Sprintf("expected_revenue >= $%d", argIndex))
		args = append(args, *filter.ExpectedRevenueMin)
		argIndex++
	}
	if filter.ExpectedRevenueMax != nil {
		conditions = append(conditions, fmt.Sprintf("expected_revenue <= $%d", argIndex))
		args = append(args, *filter.ExpectedRevenueMax)
		argIndex++
	}

	// Probability range filter
	if filter.ProbabilityMin != nil {
		conditions = append(conditions, fmt.Sprintf("probability >= $%d", argIndex))
		args = append(args, *filter.ProbabilityMin)
		argIndex++
	}
	if filter.ProbabilityMax != nil {
		conditions = append(conditions, fmt.Sprintf("probability <= $%d", argIndex))
		args = append(args, *filter.ProbabilityMax)
		argIndex++
	}

	// Won status filter
	if filter.WonStatus != nil && *filter.WonStatus != "" {
		conditions = append(conditions, fmt.Sprintf("won_status = $%d", argIndex))
		args = append(args, *filter.WonStatus)
		argIndex++
	}

	// Lost reason ID filter
	if filter.LostReasonID != nil && *filter.LostReasonID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("lost_reason_id = $%d", argIndex))
		args = append(args, *filter.LostReasonID)
		argIndex++
	}

	// Active filter
	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filter.Active)
		argIndex++
	}

	// Country ID filter
	if filter.CountryID != nil && *filter.CountryID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("country_id = $%d", argIndex))
		args = append(args, *filter.CountryID)
		argIndex++
	}

	// State ID filter
	if filter.StateID != nil && *filter.StateID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("state_id = $%d", argIndex))
		args = append(args, *filter.StateID)
		argIndex++
	}

	// City filter
	if filter.City != nil && *filter.City != "" {
		conditions = append(conditions, fmt.Sprintf("city ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.City+"%")
		argIndex++
	}

	// Build WHERE clause
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Order by name
	query += " ORDER BY name ASC"

	// Pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find enhanced leads: %w", err)
	}
	defer rows.Close()

	var leads []*types.LeadEnhanced
	for rows.Next() {
		var lead types.LeadEnhanced
		err := rows.Scan(
			&lead.ID,
			&lead.OrganizationID,
			&lead.CompanyID,
			&lead.Name,
			&lead.ContactName,
			&lead.Email,
			&lead.Phone,
			&lead.Mobile,
			&lead.ContactID,
			&lead.UserID,
			&lead.TeamID,
			&lead.LeadType,
			&lead.StageID,
			&lead.Priority,
			&lead.SourceID,
			&lead.MediumID,
			&lead.CampaignID,
			&lead.ExpectedRevenue,
			&lead.Probability,
			&lead.RecurringRevenue,
			&lead.RecurringPlan,
			&lead.DateOpen,
			&lead.DateClosed,
			&lead.DateDeadline,
			&lead.DateLastStageUpdate,
			&lead.Active,
			&lead.WonStatus,
			&lead.LostReasonID,
			&lead.Street,
			&lead.Street2,
			&lead.City,
			&lead.StateID,
			&lead.Zip,
			&lead.CountryID,
			&lead.Website,
			&lead.Description,
			&lead.TagIDs,
			&lead.Color,
			&lead.CreatedAt,
			&lead.UpdatedAt,
			&lead.CreatedBy,
			&lead.UpdatedBy,
			&lead.DeletedAt,
			&lead.CustomFields,
			&lead.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan enhanced lead: %w", err)
		}
		leads = append(leads, &lead)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during enhanced lead iteration: %w", err)
	}

	return leads, nil
}

// Update modifies an existing enhanced lead
func (r *LeadRepository) Update(ctx context.Context, lead *types.LeadEnhanced) error {
	if lead.ID == uuid.Nil {
		return errors.New("lead id is required")
	}

	if lead.OrganizationID == uuid.Nil {
		return errors.New("organization_id is required")
	}

	if lead.Name == "" {
		return errors.New("name is required")
	}

	lead.UpdatedAt = time.Now()

	query := `
		UPDATE leads_enhanced SET
			organization_id = $1,
			company_id = $2,
			name = $3,
			contact_name = $4,
			email = $5,
			phone = $6,
			mobile = $7,
			contact_id = $8,
			user_id = $9,
			team_id = $10,
			lead_type = $11,
			stage_id = $12,
			priority = $13,
			source_id = $14,
			medium_id = $15,
			campaign_id = $16,
			expected_revenue = $17,
			probability = $18,
			recurring_revenue = $19,
			recurring_plan = $20,
			date_open = $21,
			date_closed = $22,
			date_deadline = $23,
			date_last_stage_update = $24,
			active = $25,
			won_status = $26,
			lost_reason_id = $27,
			street = $28,
			street2 = $29,
			city = $30,
			state_id = $31,
			zip = $32,
			country_id = $33,
			website = $34,
			description = $35,
			tag_ids = $36,
			color = $37,
			updated_at = $38,
			updated_by = $39
		WHERE id = $40 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query,
		lead.OrganizationID,
		lead.CompanyID,
		lead.Name,
		lead.ContactName,
		lead.Email,
		lead.Phone,
		lead.Mobile,
		lead.ContactID,
		lead.UserID,
		lead.TeamID,
		lead.LeadType,
		lead.StageID,
		lead.Priority,
		lead.SourceID,
		lead.MediumID,
		lead.CampaignID,
		lead.ExpectedRevenue,
		lead.Probability,
		lead.RecurringRevenue,
		lead.RecurringPlan,
		lead.DateOpen,
		lead.DateClosed,
		lead.DateDeadline,
		lead.DateLastStageUpdate,
		lead.Active,
		lead.WonStatus,
		lead.LostReasonID,
		lead.Street,
		lead.Street2,
		lead.City,
		lead.StateID,
		lead.Zip,
		lead.CountryID,
		lead.Website,
		lead.Description,
		lead.TagIDs,
		lead.Color,
		lead.UpdatedAt,
		lead.UpdatedBy,
		lead.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update enhanced lead: %w", err)
	}

	return nil
}

// Delete removes an enhanced lead (soft delete)
func (r *LeadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid lead id")
	}

	query := `
		UPDATE leads_enhanced SET
			deleted_at = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete enhanced lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("enhanced lead not found or already deleted")
	}

	return nil
}

// Count counts enhanced leads matching the filter criteria
func (r *LeadRepository) Count(ctx context.Context, filter types.LeadEnhancedFilter) (int, error) {
	query := `SELECT COUNT(*) FROM leads_enhanced WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	var argIndex = 1

	// Organization filter (required)
	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, filter.OrganizationID)
	argIndex++

	// Name filter
	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	// Email filter
	if filter.Email != nil && *filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	// Phone filter
	if filter.Phone != nil && *filter.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Phone+"%")
		argIndex++
	}

	// Contact name filter
	if filter.ContactName != nil && *filter.ContactName != "" {
		conditions = append(conditions, fmt.Sprintf("contact_name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.ContactName+"%")
		argIndex++
	}

	// Mobile filter
	if filter.Mobile != nil && *filter.Mobile != "" {
		conditions = append(conditions, fmt.Sprintf("mobile ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Mobile+"%")
		argIndex++
	}

	// Company ID filter
	if filter.CompanyID != nil && *filter.CompanyID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("company_id = $%d", argIndex))
		args = append(args, *filter.CompanyID)
		argIndex++
	}

	// Contact ID filter
	if filter.ContactID != nil && *filter.ContactID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("contact_id = $%d", argIndex))
		args = append(args, *filter.ContactID)
		argIndex++
	}

	// User ID filter
	if filter.UserID != nil && *filter.UserID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	// Team ID filter
	if filter.TeamID != nil && *filter.TeamID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("team_id = $%d", argIndex))
		args = append(args, *filter.TeamID)
		argIndex++
	}

	// Lead type filter
	if filter.LeadType != nil && *filter.LeadType != "" {
		conditions = append(conditions, fmt.Sprintf("lead_type = $%d", argIndex))
		args = append(args, *filter.LeadType)
		argIndex++
	}

	// Stage ID filter
	if filter.StageID != nil && *filter.StageID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("stage_id = $%d", argIndex))
		args = append(args, *filter.StageID)
		argIndex++
	}

	// Priority filter
	if filter.Priority != nil && *filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}

	// Source ID filter
	if filter.SourceID != nil && *filter.SourceID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("source_id = $%d", argIndex))
		args = append(args, *filter.SourceID)
		argIndex++
	}

	// Medium ID filter
	if filter.MediumID != nil && *filter.MediumID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("medium_id = $%d", argIndex))
		args = append(args, *filter.MediumID)
		argIndex++
	}

	// Campaign ID filter
	if filter.CampaignID != nil && *filter.CampaignID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("campaign_id = $%d", argIndex))
		args = append(args, *filter.CampaignID)
		argIndex++
	}

	// Expected revenue range filter
	if filter.ExpectedRevenueMin != nil {
		conditions = append(conditions, fmt.Sprintf("expected_revenue >= $%d", argIndex))
		args = append(args, *filter.ExpectedRevenueMin)
		argIndex++
	}
	if filter.ExpectedRevenueMax != nil {
		conditions = append(conditions, fmt.Sprintf("expected_revenue <= $%d", argIndex))
		args = append(args, *filter.ExpectedRevenueMax)
		argIndex++
	}

	// Probability range filter
	if filter.ProbabilityMin != nil {
		conditions = append(conditions, fmt.Sprintf("probability >= $%d", argIndex))
		args = append(args, *filter.ProbabilityMin)
		argIndex++
	}
	if filter.ProbabilityMax != nil {
		conditions = append(conditions, fmt.Sprintf("probability <= $%d", argIndex))
		args = append(args, *filter.ProbabilityMax)
		argIndex++
	}

	// Won status filter
	if filter.WonStatus != nil && *filter.WonStatus != "" {
		conditions = append(conditions, fmt.Sprintf("won_status = $%d", argIndex))
		args = append(args, *filter.WonStatus)
		argIndex++
	}

	// Lost reason ID filter
	if filter.LostReasonID != nil && *filter.LostReasonID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("lost_reason_id = $%d", argIndex))
		args = append(args, *filter.LostReasonID)
		argIndex++
	}

	// Active filter
	if filter.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filter.Active)
		argIndex++
	}

	// Country ID filter
	if filter.CountryID != nil && *filter.CountryID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("country_id = $%d", argIndex))
		args = append(args, *filter.CountryID)
		argIndex++
	}

	// State ID filter
	if filter.StateID != nil && *filter.StateID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("state_id = $%d", argIndex))
		args = append(args, *filter.StateID)
		argIndex++
	}

	// City filter
	if filter.City != nil && *filter.City != "" {
		conditions = append(conditions, fmt.Sprintf("city ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.City+"%")
		argIndex++
	}

	// Build WHERE clause
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count enhanced leads: %w", err)
	}

	return count, nil
}

// FindByContact retrieves leads associated with a contact
func (r *LeadRepository) FindByContact(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if contactID == uuid.Nil {
		return nil, errors.New("invalid contact ID")
	}

	query := `
		SELECT id, organization_id, company_id, name, contact_name, email, phone, mobile,
			contact_id, user_id, team_id, lead_type, stage_id, priority, source_id,
			medium_id, campaign_id, expected_revenue, probability, recurring_revenue,
			recurring_plan, date_open, date_closed, date_deadline, date_last_stage_update,
			active, won_status, lost_reason_id, street, street2, city, state_id, zip,
			country_id, website, description, tag_ids, color, created_at, updated_at,
			created_by, updated_by, deleted_at, custom_fields, metadata
		FROM leads_enhanced
		WHERE contact_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, contactID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to find leads by contact: %w", err)
	}
	defer rows.Close()

	var leads []*types.LeadEnhanced
	for rows.Next() {
		var lead types.LeadEnhanced
		err := rows.Scan(
			&lead.ID,
			&lead.OrganizationID,
			&lead.CompanyID,
			&lead.Name,
			&lead.ContactName,
			&lead.Email,
			&lead.Phone,
			&lead.Mobile,
			&lead.ContactID,
			&lead.UserID,
			&lead.TeamID,
			&lead.LeadType,
			&lead.StageID,
			&lead.Priority,
			&lead.SourceID,
			&lead.MediumID,
			&lead.CampaignID,
			&lead.ExpectedRevenue,
			&lead.Probability,
			&lead.RecurringRevenue,
			&lead.RecurringPlan,
			&lead.DateOpen,
			&lead.DateClosed,
			&lead.DateDeadline,
			&lead.DateLastStageUpdate,
			&lead.Active,
			&lead.WonStatus,
			&lead.LostReasonID,
			&lead.Street,
			&lead.Street2,
			&lead.City,
			&lead.StateID,
			&lead.Zip,
			&lead.CountryID,
			&lead.Website,
			&lead.Description,
			&lead.TagIDs,
			&lead.Color,
			&lead.CreatedAt,
			&lead.UpdatedAt,
			&lead.CreatedBy,
			&lead.UpdatedBy,
			&lead.DeletedAt,
			&lead.CustomFields,
			&lead.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lead: %w", err)
		}
		leads = append(leads, &lead)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during lead iteration: %w", err)
	}

	return leads, nil
}

// FindByUser retrieves leads assigned to a user
func (r *LeadRepository) FindByUser(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	query := `
		SELECT id, organization_id, company_id, name, contact_name, email, phone, mobile,
			contact_id, user_id, team_id, lead_type, stage_id, priority, source_id,
			medium_id, campaign_id, expected_revenue, probability, recurring_revenue,
			recurring_plan, date_open, date_closed, date_deadline, date_last_stage_update,
			active, won_status, lost_reason_id, street, street2, city, state_id, zip,
			country_id, website, description, tag_ids, color, created_at, updated_at,
			created_by, updated_by, deleted_at, custom_fields, metadata
		FROM leads_enhanced
		WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to find leads by user: %w", err)
	}
	defer rows.Close()

	var leads []*types.LeadEnhanced
	for rows.Next() {
		var lead types.LeadEnhanced
		err := rows.Scan(
			&lead.ID,
			&lead.OrganizationID,
			&lead.CompanyID,
			&lead.Name,
			&lead.ContactName,
			&lead.Email,
			&lead.Phone,
			&lead.Mobile,
			&lead.ContactID,
			&lead.UserID,
			&lead.TeamID,
			&lead.LeadType,
			&lead.StageID,
			&lead.Priority,
			&lead.SourceID,
			&lead.MediumID,
			&lead.CampaignID,
			&lead.ExpectedRevenue,
			&lead.Probability,
			&lead.RecurringRevenue,
			&lead.RecurringPlan,
			&lead.DateOpen,
			&lead.DateClosed,
			&lead.DateDeadline,
			&lead.DateLastStageUpdate,
			&lead.Active,
			&lead.WonStatus,
			&lead.LostReasonID,
			&lead.Street,
			&lead.Street2,
			&lead.City,
			&lead.StateID,
			&lead.Zip,
			&lead.CountryID,
			&lead.Website,
			&lead.Description,
			&lead.TagIDs,
			&lead.Color,
			&lead.CreatedAt,
			&lead.UpdatedAt,
			&lead.CreatedBy,
			&lead.UpdatedBy,
			&lead.DeletedAt,
			&lead.CustomFields,
			&lead.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lead: %w", err)
		}
		leads = append(leads, &lead)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during lead iteration: %w", err)
	}

	return leads, nil
}

// FindByTeam retrieves leads assigned to a team
func (r *LeadRepository) FindByTeam(ctx context.Context, orgID uuid.UUID, teamID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if teamID == uuid.Nil {
		return nil, errors.New("invalid team ID")
	}

	query := `
		SELECT id, organization_id, company_id, name, contact_name, email, phone, mobile,
			contact_id, user_id, team_id, lead_type, stage_id, priority, source_id,
			medium_id, campaign_id, expected_revenue, probability, recurring_revenue,
			recurring_plan, date_open, date_closed, date_deadline, date_last_stage_update,
			active, won_status, lost_reason_id, street, street2, city, state_id, zip,
			country_id, website, description, tag_ids, color, created_at, updated_at,
			created_by, updated_by, deleted_at, custom_fields, metadata
		FROM leads_enhanced
		WHERE team_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, teamID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to find leads by team: %w", err)
	}
	defer rows.Close()

	var leads []*types.LeadEnhanced
	for rows.Next() {
		var lead types.LeadEnhanced
		err := rows.Scan(
			&lead.ID,
			&lead.OrganizationID,
			&lead.CompanyID,
			&lead.Name,
			&lead.ContactName,
			&lead.Email,
			&lead.Phone,
			&lead.Mobile,
			&lead.ContactID,
			&lead.UserID,
			&lead.TeamID,
			&lead.LeadType,
			&lead.StageID,
			&lead.Priority,
			&lead.SourceID,
			&lead.MediumID,
			&lead.CampaignID,
			&lead.ExpectedRevenue,
			&lead.Probability,
			&lead.RecurringRevenue,
			&lead.RecurringPlan,
			&lead.DateOpen,
			&lead.DateClosed,
			&lead.DateDeadline,
			&lead.DateLastStageUpdate,
			&lead.Active,
			&lead.WonStatus,
			&lead.LostReasonID,
			&lead.Street,
			&lead.Street2,
			&lead.City,
			&lead.StateID,
			&lead.Zip,
			&lead.CountryID,
			&lead.Website,
			&lead.Description,
			&lead.TagIDs,
			&lead.Color,
			&lead.CreatedAt,
			&lead.UpdatedAt,
			&lead.CreatedBy,
			&lead.UpdatedBy,
			&lead.DeletedAt,
			&lead.CustomFields,
			&lead.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lead: %w", err)
		}
		leads = append(leads, &lead)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during lead iteration: %w", err)
	}

	return leads, nil
}

// FindByStage retrieves leads in a specific stage
func (r *LeadRepository) FindByStage(ctx context.Context, orgID uuid.UUID, stageID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if stageID == uuid.Nil {
		return nil, errors.New("invalid stage ID")
	}

	query := `
		SELECT id, organization_id, company_id, name, contact_name, email, phone, mobile,
			contact_id, user_id, team_id, lead_type, stage_id, priority, source_id,
			medium_id, campaign_id, expected_revenue, probability, recurring_revenue,
			recurring_plan, date_open, date_closed, date_deadline, date_last_stage_update,
			active, won_status, lost_reason_id, street, street2, city, state_id, zip,
			country_id, website, description, tag_ids, color, created_at, updated_at,
			created_by, updated_by, deleted_at, custom_fields, metadata
		FROM leads_enhanced
		WHERE stage_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, stageID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to find leads by stage: %w", err)
	}
	defer rows.Close()

	var leads []*types.LeadEnhanced
	for rows.Next() {
		var lead types.LeadEnhanced
		err := rows.Scan(
			&lead.ID,
			&lead.OrganizationID,
			&lead.CompanyID,
			&lead.Name,
			&lead.ContactName,
			&lead.Email,
			&lead.Phone,
			&lead.Mobile,
			&lead.ContactID,
			&lead.UserID,
			&lead.TeamID,
			&lead.LeadType,
			&lead.StageID,
			&lead.Priority,
			&lead.SourceID,
			&lead.MediumID,
			&lead.CampaignID,
			&lead.ExpectedRevenue,
			&lead.Probability,
			&lead.RecurringRevenue,
			&lead.RecurringPlan,
			&lead.DateOpen,
			&lead.DateClosed,
			&lead.DateDeadline,
			&lead.DateLastStageUpdate,
			&lead.Active,
			&lead.WonStatus,
			&lead.LostReasonID,
			&lead.Street,
			&lead.Street2,
			&lead.City,
			&lead.StateID,
			&lead.Zip,
			&lead.CountryID,
			&lead.Website,
			&lead.Description,
			&lead.TagIDs,
			&lead.Color,
			&lead.CreatedAt,
			&lead.UpdatedAt,
			&lead.CreatedBy,
			&lead.UpdatedBy,
			&lead.DeletedAt,
			&lead.CustomFields,
			&lead.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lead: %w", err)
		}
		leads = append(leads, &lead)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during lead iteration: %w", err)
	}

	return leads, nil
}

// CountByStage counts leads by stage for pipeline analytics
func (r *LeadRepository) CountByStage(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	query := `
		SELECT stage_id, COUNT(*)
		FROM leads_enhanced
		WHERE organization_id = $1 AND deleted_at IS NULL
		GROUP BY stage_id
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count leads by stage: %w", err)
	}
	defer rows.Close()

	counts := make(map[uuid.UUID]int)
	for rows.Next() {
		var stageID uuid.UUID
		var count int
		if err := rows.Scan(&stageID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan stage count: %w", err)
		}
		counts[stageID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during stage count iteration: %w", err)
	}

	return counts, nil
}

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// AssignmentRuleRepositoryPostgres implements AssignmentRuleRepository for PostgreSQL
type AssignmentRuleRepositoryPostgres struct {
	db *sql.DB
}

func NewAssignmentRuleRepository(db *sql.DB) types.AssignmentRuleRepository {
	return &AssignmentRuleRepositoryPostgres{db: db}
}

// CreateAssignmentRule creates a new assignment rule
func (r *AssignmentRuleRepositoryPostgres) CreateAssignmentRule(ctx context.Context, rule *types.AssignmentRule) error {
	query := `
		INSERT INTO assignment_rules (
			id, organization_id, name, description, rule_type, target_model,
			priority, is_active, conditions, assignment_config, assign_to_type,
			max_assignments_per_user, assignment_window_start, assignment_window_end,
			active_days, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING created_at, updated_at
	`

	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	assignmentConfigJSON, err := json.Marshal(rule.AssignmentConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal assignment config: %w", err)
	}

	var windowStart, windowEnd *time.Time
	if rule.AssignmentWindowStart != nil {
		windowStart = rule.AssignmentWindowStart
	}
	if rule.AssignmentWindowEnd != nil {
		windowEnd = rule.AssignmentWindowEnd
	}

	err = r.db.QueryRowContext(ctx, query,
		rule.ID,
		rule.OrganizationID,
		rule.Name,
		rule.Description,
		rule.RuleType,
		rule.TargetModel,
		rule.Priority,
		rule.IsActive,
		conditionsJSON,
		assignmentConfigJSON,
		rule.AssignToType,
		rule.MaxAssignmentsPerUser,
		windowStart,
		windowEnd,
		rule.ActiveDays,
		rule.CreatedBy,
		rule.UpdatedBy,
	).Scan(&rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create assignment rule: %w", err)
	}

	return nil
}

// Create implements the repository interface
func (r *AssignmentRuleRepositoryPostgres) Create(ctx context.Context, rule types.AssignmentRule) (*types.AssignmentRule, error) {
	// Generate ID if not provided
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	// Get organization ID from context
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return nil, errors.New("organization ID not found in context")
	}

	// Marshal JSON fields
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	assignmentConfigJSON, err := json.Marshal(rule.AssignmentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal assignment config: %w", err)
	}

	activeDaysJSON, err := json.Marshal(rule.ActiveDays)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal active days: %w", err)
	}

	query := `
		INSERT INTO assignment_rules (
			id, organization_id, name, description, rule_type, target_model,
			priority, is_active, conditions, assignment_config, assign_to_type,
			max_assignments_per_user, assignment_window_start, assignment_window_end,
			active_days, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		rule.ID,
		orgID,
		rule.Name,
		rule.Description,
		rule.RuleType,
		rule.TargetModel,
		rule.Priority,
		rule.IsActive,
		conditionsJSON,
		assignmentConfigJSON,
		rule.AssignToType,
		rule.MaxAssignmentsPerUser,
		rule.AssignmentWindowStart,
		rule.AssignmentWindowEnd,
		activeDaysJSON,
		rule.CreatedBy,
		rule.UpdatedBy,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create assignment rule: %w", err)
	}

	return &rule, nil
}

// FindByTargetModel finds assignment rules by target model
func (r *AssignmentRuleRepositoryPostgres) FindByTargetModel(ctx context.Context, targetModel types.AssignmentTargetModel) ([]types.AssignmentRule, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return nil, errors.New("organization ID not found in context")
	}

	query := `
		SELECT id, organization_id, name, description, rule_type, target_model,
			priority, is_active, conditions, assignment_config, assign_to_type,
			max_assignments_per_user, assignment_window_start, assignment_window_end,
			active_days, created_by, updated_by, created_at, updated_at
		FROM assignment_rules
		WHERE target_model = $1 AND organization_id = $2
		ORDER BY priority ASC
	`

	rows, err := r.db.QueryContext(ctx, query, targetModel, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignment rules: %w", err)
	}
	defer rows.Close()

	var rules []types.AssignmentRule

	for rows.Next() {
		var rule types.AssignmentRule
		var conditionsJSON, assignmentConfigJSON, activeDaysJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.OrganizationID,
			&rule.Name,
			&rule.Description,
			&rule.RuleType,
			&rule.TargetModel,
			&rule.Priority,
			&rule.IsActive,
			&conditionsJSON,
			&assignmentConfigJSON,
			&rule.AssignToType,
			&rule.MaxAssignmentsPerUser,
			&rule.AssignmentWindowStart,
			&rule.AssignmentWindowEnd,
			&activeDaysJSON,
			&rule.CreatedBy,
			&rule.UpdatedBy,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment rule: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		if err := json.Unmarshal(assignmentConfigJSON, &rule.AssignmentConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignment config: %w", err)
		}

		if err := json.Unmarshal(activeDaysJSON, &rule.ActiveDays); err != nil {
			return nil, fmt.Errorf("failed to unmarshal active days: %w", err)
		}

		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignment rules: %w", err)
	}

	return rules, nil
}

// Update updates an existing assignment rule
func (r *AssignmentRuleRepositoryPostgres) Update(ctx context.Context, rule types.AssignmentRule) (*types.AssignmentRule, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return nil, errors.New("organization ID not found in context")
	}

	// Verify the rule belongs to the organization
	if rule.OrganizationID != orgID {
		return nil, fmt.Errorf("assignment rule does not belong to organization")
	}

	// Marshal JSON fields
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	assignmentConfigJSON, err := json.Marshal(rule.AssignmentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal assignment config: %w", err)
	}

	activeDaysJSON, err := json.Marshal(rule.ActiveDays)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal active days: %w", err)
	}

	query := `
		UPDATE assignment_rules SET
			name = $1,
			description = $2,
			rule_type = $3,
			target_model = $4,
			priority = $5,
			is_active = $6,
			conditions = $7,
			assignment_config = $8,
			assign_to_type = $9,
			max_assignments_per_user = $10,
			assignment_window_start = $11,
			assignment_window_end = $12,
			active_days = $13,
			updated_by = $14,
			updated_at = NOW()
		WHERE id = $15 AND organization_id = $16
		RETURNING updated_at
	`

	var updatedAt time.Time
	err = r.db.QueryRowContext(ctx, query,
		rule.Name,
		rule.Description,
		rule.RuleType,
		rule.TargetModel,
		rule.Priority,
		rule.IsActive,
		conditionsJSON,
		assignmentConfigJSON,
		rule.AssignToType,
		rule.MaxAssignmentsPerUser,
		rule.AssignmentWindowStart,
		rule.AssignmentWindowEnd,
		activeDaysJSON,
		rule.UpdatedBy,
		rule.ID,
		orgID,
	).Scan(&updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update assignment rule: %w", err)
	}

	// Set the updated timestamp
	rule.UpdatedAt = updatedAt

	return &rule, nil
}

// Count counts assignment rules matching the filter criteria
func (r *AssignmentRuleRepositoryPostgres) Count(ctx context.Context, filter types.AssignmentRuleFilter) (int, error) {
	query := `SELECT COUNT(*) FROM assignment_rules WHERE target_model = $1`
	args := []interface{}{filter.TargetModel}
	argIndex := 2

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *filter.IsActive)
		argIndex++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count assignment rules: %w", err)
	}

	return count, nil
}

// GetAssignmentRule retrieves an assignment rule by ID
func (r *AssignmentRuleRepositoryPostgres) GetAssignmentRule(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
	query := `
		SELECT id, organization_id, name, description, rule_type, target_model,
		       priority, is_active, conditions, assignment_config, assign_to_type,
		       max_assignments_per_user, assignment_window_start, assignment_window_end,
		       active_days, created_at, updated_at, created_by, updated_by
		FROM assignment_rules
		WHERE id = $1
	`

	var rule types.AssignmentRule
	var conditionsJSON, assignmentConfigJSON []byte
	var windowStart, windowEnd *time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.OrganizationID,
		&rule.Name,
		&rule.Description,
		&rule.RuleType,
		&rule.TargetModel,
		&rule.Priority,
		&rule.IsActive,
		&conditionsJSON,
		&assignmentConfigJSON,
		&rule.AssignToType,
		&rule.MaxAssignmentsPerUser,
		&windowStart,
		&windowEnd,
		&rule.ActiveDays,
		&rule.CreatedAt,
		&rule.UpdatedAt,
		&rule.CreatedBy,
		&rule.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get assignment rule: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	if err := json.Unmarshal(assignmentConfigJSON, &rule.AssignmentConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assignment config: %w", err)
	}

	rule.AssignmentWindowStart = windowStart
	rule.AssignmentWindowEnd = windowEnd

	return &rule, nil
}

// UpdateAssignmentRule updates an existing assignment rule
func (r *AssignmentRuleRepositoryPostgres) UpdateAssignmentRule(ctx context.Context, rule *types.AssignmentRule) error {
	query := `
		UPDATE assignment_rules SET
			name = $1,
			description = $2,
			rule_type = $3,
			target_model = $4,
			priority = $5,
			is_active = $6,
			conditions = $7,
			assignment_config = $8,
			assign_to_type = $9,
			max_assignments_per_user = $10,
			assignment_window_start = $11,
			assignment_window_end = $12,
			active_days = $13,
			updated_by = $14,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $15
		RETURNING updated_at
	`

	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	assignmentConfigJSON, err := json.Marshal(rule.AssignmentConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal assignment config: %w", err)
	}

	var windowStart, windowEnd *time.Time
	if rule.AssignmentWindowStart != nil {
		windowStart = rule.AssignmentWindowStart
	}
	if rule.AssignmentWindowEnd != nil {
		windowEnd = rule.AssignmentWindowEnd
	}

	err = r.db.QueryRowContext(ctx, query,
		rule.Name,
		rule.Description,
		rule.RuleType,
		rule.TargetModel,
		rule.Priority,
		rule.IsActive,
		conditionsJSON,
		assignmentConfigJSON,
		rule.AssignToType,
		rule.MaxAssignmentsPerUser,
		windowStart,
		windowEnd,
		rule.ActiveDays,
		rule.UpdatedBy,
		rule.ID,
	).Scan(&rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update assignment rule: %w", err)
	}

	return nil
}

// DeleteAssignmentRule deletes an assignment rule
func (r *AssignmentRuleRepositoryPostgres) DeleteAssignmentRule(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM assignment_rules WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Delete implements the repository interface
func (r *AssignmentRuleRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM assignment_rules WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListAssignmentRules lists assignment rules with filters
func (r *AssignmentRuleRepositoryPostgres) ListAssignmentRules(ctx context.Context, orgID uuid.UUID, targetModel string, activeOnly bool) ([]*types.AssignmentRule, error) {
	query := `
		SELECT id, organization_id, name, description, rule_type, target_model,
		       priority, is_active, conditions, assignment_config, assign_to_type,
		       max_assignments_per_user, assignment_window_start, assignment_window_end,
		       active_days, created_at, updated_at, created_by, updated_by
		FROM assignment_rules
		WHERE organization_id = $1
	`

	params := []interface{}{orgID}
	paramIndex := 2

	if targetModel != "" {
		query += fmt.Sprintf(" AND target_model = $%d", paramIndex)
		params = append(params, targetModel)
		paramIndex++
	}

	if activeOnly {
		query += fmt.Sprintf(" AND is_active = true")
	}

	query += " ORDER BY priority DESC, name ASC"

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to list assignment rules: %w", err)
	}
	defer rows.Close()

	var rules []*types.AssignmentRule
	for rows.Next() {
		var rule types.AssignmentRule
		var conditionsJSON, assignmentConfigJSON []byte
		var windowStart, windowEnd *time.Time

		err := rows.Scan(
			&rule.ID,
			&rule.OrganizationID,
			&rule.Name,
			&rule.Description,
			&rule.RuleType,
			&rule.TargetModel,
			&rule.Priority,
			&rule.IsActive,
			&conditionsJSON,
			&assignmentConfigJSON,
			&rule.AssignToType,
			&rule.MaxAssignmentsPerUser,
			&windowStart,
			&windowEnd,
			&rule.ActiveDays,
			&rule.CreatedAt,
			&rule.UpdatedAt,
			&rule.CreatedBy,
			&rule.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment rule: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		if err := json.Unmarshal(assignmentConfigJSON, &rule.AssignmentConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignment config: %w", err)
		}

		rule.AssignmentWindowStart = windowStart
		rule.AssignmentWindowEnd = windowEnd

		rules = append(rules, &rule)
	}

	return rules, nil
}

// FindAll finds all assignment rules with pagination
func (r *AssignmentRuleRepositoryPostgres) FindAll(ctx context.Context, limit, offset int) ([]types.AssignmentRule, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return nil, errors.New("organization ID not found in context")
	}

	query := `
		SELECT id, organization_id, name, description, rule_type, target_model,
			priority, is_active, conditions, assignment_config, assign_to_type,
			max_assignments_per_user, assignment_window_start, assignment_window_end,
			active_days, created_by, updated_by, created_at, updated_at
		FROM assignment_rules
		WHERE organization_id = $1
		ORDER BY priority ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find assignment rules: %w", err)
	}
	defer rows.Close()

	var rules []types.AssignmentRule
	for rows.Next() {
		var rule types.AssignmentRule
		var conditionsJSON, assignmentConfigJSON, activeDaysJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.OrganizationID,
			&rule.Name,
			&rule.Description,
			&rule.RuleType,
			&rule.TargetModel,
			&rule.Priority,
			&rule.IsActive,
			&conditionsJSON,
			&assignmentConfigJSON,
			&rule.AssignToType,
			&rule.MaxAssignmentsPerUser,
			&rule.AssignmentWindowStart,
			&rule.AssignmentWindowEnd,
			&activeDaysJSON,
			&rule.CreatedBy,
			&rule.UpdatedBy,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment rule: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
		if err := json.Unmarshal(assignmentConfigJSON, &rule.AssignmentConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignment config: %w", err)
		}
		if err := json.Unmarshal(activeDaysJSON, &rule.ActiveDays); err != nil {
			return nil, fmt.Errorf("failed to unmarshal active days: %w", err)
		}

		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignment rules: %w", err)
	}

	return rules, nil
}

// FindByID finds an assignment rule by ID
func (r *AssignmentRuleRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*types.AssignmentRule, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return nil, errors.New("organization ID not found in context")
	}

	query := `
		SELECT id, organization_id, name, description, rule_type, target_model,
			priority, is_active, conditions, assignment_config, assign_to_type,
			max_assignments_per_user, assignment_window_start, assignment_window_end,
			active_days, created_by, updated_by, created_at, updated_at
		FROM assignment_rules
		WHERE id = $1 AND organization_id = $2
	`

	var rule types.AssignmentRule
	var conditionsJSON, assignmentConfigJSON, activeDaysJSON []byte

	err := r.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&rule.ID,
		&rule.OrganizationID,
		&rule.Name,
		&rule.Description,
		&rule.RuleType,
		&rule.TargetModel,
		&rule.Priority,
		&rule.IsActive,
		&conditionsJSON,
		&assignmentConfigJSON,
		&rule.AssignToType,
		&rule.MaxAssignmentsPerUser,
		&rule.AssignmentWindowStart,
		&rule.AssignmentWindowEnd,
		&activeDaysJSON,
		&rule.CreatedBy,
		&rule.UpdatedBy,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("assignment rule not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find assignment rule: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	if err := json.Unmarshal(assignmentConfigJSON, &rule.AssignmentConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assignment config: %w", err)
	}

	if err := json.Unmarshal(activeDaysJSON, &rule.ActiveDays); err != nil {
		return nil, fmt.Errorf("failed to unmarshal active days: %w", err)
	}

	return &rule, nil
}

// CreateAssignmentHistory creates a new assignment history record
func (r *AssignmentRuleRepositoryPostgres) CreateAssignmentHistory(ctx context.Context, history *types.AssignmentHistory) error {
	query := `
		INSERT INTO assignment_history (
			id, organization_id, rule_id, rule_name, target_model, target_id,
			target_name, assigned_to_type, assigned_to_id, assigned_to_name,
			previous_assigned_to_id, previous_assigned_to_name, assignment_reason,
			metadata, assigned_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING assigned_at
	`

	metadataJSON, err := json.Marshal(history.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		history.ID,
		history.OrganizationID,
		history.RuleID,
		history.RuleName,
		history.TargetModel,
		history.TargetID,
		history.TargetName,
		history.AssignedToType,
		history.AssignedToID,
		history.AssignedToName,
		history.PreviousAssignedToID,
		history.PreviousAssignedToName,
		history.AssignmentReason,
		metadataJSON,
		history.AssignedBy,
	).Scan(&history.AssignedAt)

	if err != nil {
		return fmt.Errorf("failed to create assignment history: %w", err)
	}

	return nil
}

// GetAssignmentHistory retrieves an assignment history record by ID
func (r *AssignmentRuleRepositoryPostgres) GetAssignmentHistory(ctx context.Context, id uuid.UUID) (*types.AssignmentHistory, error) {
	query := `
		SELECT id, organization_id, rule_id, rule_name, target_model, target_id,
		       target_name, assigned_to_type, assigned_to_id, assigned_to_name,
		       previous_assigned_to_id, previous_assigned_to_name, assignment_reason,
		       metadata, assigned_at, assigned_by
		FROM assignment_history
		WHERE id = $1
	`

	var history types.AssignmentHistory
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&history.ID,
		&history.OrganizationID,
		&history.RuleID,
		&history.RuleName,
		&history.TargetModel,
		&history.TargetID,
		&history.TargetName,
		&history.AssignedToType,
		&history.AssignedToID,
		&history.AssignedToName,
		&history.PreviousAssignedToID,
		&history.PreviousAssignedToName,
		&history.AssignmentReason,
		&metadataJSON,
		&history.AssignedAt,
		&history.AssignedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get assignment history: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataJSON, &history.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &history, nil
}

// ListAssignmentHistory lists assignment history records with filters
func (r *AssignmentRuleRepositoryPostgres) ListAssignmentHistory(ctx context.Context, orgID uuid.UUID, targetModel string, limit int) ([]*types.AssignmentHistory, error) {
	query := `
		SELECT id, organization_id, rule_id, rule_name, target_model, target_id,
		       target_name, assigned_to_type, assigned_to_id, assigned_to_name,
		       previous_assigned_to_id, previous_assigned_to_name, assignment_reason,
		       metadata, assigned_at, assigned_by
		FROM assignment_history
		WHERE organization_id = $1
	`

	params := []interface{}{orgID}
	paramIndex := 2

	if targetModel != "" {
		query += fmt.Sprintf(" AND target_model = $%d", paramIndex)
		params = append(params, targetModel)
		paramIndex++
	}

	query += " ORDER BY assigned_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to list assignment history: %w", err)
	}
	defer rows.Close()

	var history []*types.AssignmentHistory
	for rows.Next() {
		var h types.AssignmentHistory
		var metadataJSON []byte

		err := rows.Scan(
			&h.ID,
			&h.OrganizationID,
			&h.RuleID,
			&h.RuleName,
			&h.TargetModel,
			&h.TargetID,
			&h.TargetName,
			&h.AssignedToType,
			&h.AssignedToID,
			&h.AssignedToName,
			&h.PreviousAssignedToID,
			&h.PreviousAssignedToName,
			&h.AssignmentReason,
			&metadataJSON,
			&h.AssignedAt,
			&h.AssignedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment history: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataJSON, &h.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		history = append(history, &h)
	}

	return history, nil
}

// GetUserAssignmentLoad retrieves user assignment load
func (r *AssignmentRuleRepositoryPostgres) GetUserAssignmentLoad(ctx context.Context, userID uuid.UUID, targetModel string) (*types.UserAssignmentLoad, error) {
	query := `
		SELECT id, organization_id, user_id, target_model, active_assignments,
		       total_assignments, last_assigned_at, max_capacity, weight,
		       is_available, unavailable_until, updated_at
		FROM user_assignment_load
		WHERE user_id = $1 AND target_model = $2
	`

	var load types.UserAssignmentLoad
	var unavailableUntil *time.Time

	err := r.db.QueryRowContext(ctx, query, userID, targetModel).Scan(
		&load.ID,
		&load.OrganizationID,
		&load.UserID,
		&load.TargetModel,
		&load.ActiveAssignments,
		&load.TotalAssignments,
		&load.LastAssignedAt,
		&load.MaxCapacity,
		&load.Weight,
		&load.IsAvailable,
		&unavailableUntil,
		&load.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return a new load with default values
			return &types.UserAssignmentLoad{
				UserID:      userID,
				TargetModel: targetModel,
				Weight:      1, // Default weight
				IsAvailable: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to get user assignment load: %w", err)
	}

	if unavailableUntil != nil {
		load.UnavailableUntil = *unavailableUntil
	}

	return &load, nil
}

// UpdateUserAssignmentLoad updates user assignment load
func (r *AssignmentRuleRepositoryPostgres) UpdateUserAssignmentLoad(ctx context.Context, load *types.UserAssignmentLoad) error {
	query := `
		INSERT INTO user_assignment_load (
			id, organization_id, user_id, target_model, active_assignments,
			total_assignments, last_assigned_at, max_capacity, weight,
			is_available, unavailable_until
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (organization_id, user_id, target_model)
		DO UPDATE SET
			active_assignments = EXCLUDED.active_assignments,
			total_assignments = EXCLUDED.total_assignments,
			last_assigned_at = EXCLUDED.last_assigned_at,
			max_capacity = EXCLUDED.max_capacity,
			weight = EXCLUDED.weight,
			is_available = EXCLUDED.is_available,
			unavailable_until = EXCLUDED.unavailable_until,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, updated_at
	`

	var unavailableUntil *time.Time
	if !load.UnavailableUntil.IsZero() {
		unavailableUntil = &load.UnavailableUntil
	}

	err := r.db.QueryRowContext(ctx, query,
		load.ID,
		load.OrganizationID,
		load.UserID,
		load.TargetModel,
		load.ActiveAssignments,
		load.TotalAssignments,
		load.LastAssignedAt,
		load.MaxCapacity,
		load.Weight,
		load.IsAvailable,
		unavailableUntil,
	).Scan(&load.ID, &load.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update user assignment load: %w", err)
	}

	return nil
}

// ListUserAssignmentLoads lists user assignment loads
func (r *AssignmentRuleRepositoryPostgres) ListUserAssignmentLoads(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.UserAssignmentLoad, error) {
	query := `
		SELECT id, organization_id, user_id, target_model, active_assignments,
		       total_assignments, last_assigned_at, max_capacity, weight,
		       is_available, unavailable_until, updated_at
		FROM user_assignment_load
		WHERE organization_id = $1
	`

	params := []interface{}{orgID}
	paramIndex := 2

	if targetModel != "" {
		query += fmt.Sprintf(" AND target_model = $%d", paramIndex)
		params = append(params, targetModel)
		paramIndex++
	}

	query += " ORDER BY active_assignments DESC"

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to list user assignment loads: %w", err)
	}
	defer rows.Close()

	var loads []*types.UserAssignmentLoad
	for rows.Next() {
		var load types.UserAssignmentLoad
		var unavailableUntil *time.Time

		err := rows.Scan(
			&load.ID,
			&load.OrganizationID,
			&load.UserID,
			&load.TargetModel,
			&load.ActiveAssignments,
			&load.TotalAssignments,
			&load.LastAssignedAt,
			&load.MaxCapacity,
			&load.Weight,
			&load.IsAvailable,
			&unavailableUntil,
			&load.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user assignment load: %w", err)
		}

		if unavailableUntil != nil {
			load.UnavailableUntil = *unavailableUntil
		}

		loads = append(loads, &load)
	}

	return loads, nil
}

// GetLead retrieves a lead by ID for assignment purposes
func (r *AssignmentRuleRepositoryPostgres) GetLead(ctx context.Context, leadID uuid.UUID) (*types.Lead, error) {
	query := `
		SELECT id, organization_id, name, email, phone, stage_id, status, assigned_to, active, created_at, updated_at, deleted_at
		FROM leads
		WHERE id = $1
	`

	var lead types.Lead
	var email, phone, stageID, deletedAt interface{}
	var assignedTo sql.NullString

	err := r.db.QueryRowContext(ctx, query, leadID).Scan(
		&lead.ID,
		&lead.OrganizationID,
		&lead.Name,
		&email,
		&phone,
		&stageID,
		&lead.Status,
		&assignedTo,
		&lead.Active,
		&lead.CreatedAt,
		&lead.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	// Handle nullable fields
	if email != nil {
		emailStr := email.(string)
		lead.Email = &emailStr
	}
	if phone != nil {
		phoneStr := phone.(string)
		lead.Phone = &phoneStr
	}
	if stageID != nil {
		stageUUID := stageID.(uuid.UUID)
		lead.StageID = &stageUUID
	}
	if assignedTo.Valid {
		assignedUUID := uuid.Must(uuid.Parse(assignedTo.String))
		lead.AssignedTo = &assignedUUID
	}
	if deletedAt != nil {
		deletedTime := deletedAt.(time.Time)
		lead.DeletedAt = &deletedTime
	}

	return &lead, nil
}

// CreateTerritory creates a new territory
func (r *AssignmentRuleRepositoryPostgres) CreateTerritory(ctx context.Context, territory *types.Territory) error {
	query := `
		INSERT INTO territories (
			id, organization_id, name, description, territory_type,
			conditions, assigned_users, assigned_teams, priority,
			is_active, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING created_at, updated_at
	`

	conditionsJSON, err := json.Marshal(territory.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		territory.ID,
		territory.OrganizationID,
		territory.Name,
		territory.Description,
		territory.TerritoryType,
		conditionsJSON,
		territory.AssignedUsers,
		territory.AssignedTeams,
		territory.Priority,
		territory.IsActive,
		territory.CreatedBy,
		territory.UpdatedBy,
	).Scan(&territory.CreatedAt, &territory.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create territory: %w", err)
	}

	return nil
}

// GetTerritory retrieves a territory by ID
func (r *AssignmentRuleRepositoryPostgres) GetTerritory(ctx context.Context, id uuid.UUID) (*types.Territory, error) {
	query := `
		SELECT id, organization_id, name, description, territory_type,
		       conditions, assigned_users, assigned_teams, priority,
		       is_active, created_at, updated_at, created_by, updated_by
		FROM territories
		WHERE id = $1
	`

	var territory types.Territory
	var conditionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&territory.ID,
		&territory.OrganizationID,
		&territory.Name,
		&territory.Description,
		&territory.TerritoryType,
		&conditionsJSON,
		&territory.AssignedUsers,
		&territory.AssignedTeams,
		&territory.Priority,
		&territory.IsActive,
		&territory.CreatedAt,
		&territory.UpdatedAt,
		&territory.CreatedBy,
		&territory.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get territory: %w", err)
	}

	// Unmarshal conditions
	if err := json.Unmarshal(conditionsJSON, &territory.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	return &territory, nil
}

// UpdateTerritory updates an existing territory
func (r *AssignmentRuleRepositoryPostgres) UpdateTerritory(ctx context.Context, territory *types.Territory) error {
	query := `
		UPDATE territories SET
			name = $1,
			description = $2,
			territory_type = $3,
			conditions = $4,
			assigned_users = $5,
			assigned_teams = $6,
			priority = $7,
			is_active = $8,
			updated_by = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $10
		RETURNING updated_at
	`

	conditionsJSON, err := json.Marshal(territory.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		territory.Name,
		territory.Description,
		territory.TerritoryType,
		conditionsJSON,
		territory.AssignedUsers,
		territory.AssignedTeams,
		territory.Priority,
		territory.IsActive,
		territory.UpdatedBy,
		territory.ID,
	).Scan(&territory.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update territory: %w", err)
	}

	return nil
}

// DeleteTerritory deletes a territory
func (r *AssignmentRuleRepositoryPostgres) DeleteTerritory(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM territories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListTerritories lists territories with filters
func (r *AssignmentRuleRepositoryPostgres) ListTerritories(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*types.Territory, error) {
	query := `
		SELECT id, organization_id, name, description, territory_type,
		       conditions, assigned_users, assigned_teams, priority,
		       is_active, created_at, updated_at, created_by, updated_by
		FROM territories
		WHERE organization_id = $1
	`

	params := []interface{}{orgID}
	// paramIndex := 2

	if activeOnly {
		query += fmt.Sprintf(" AND is_active = true")
	}

	query += " ORDER BY priority DESC, name ASC"

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to list territories: %w", err)
	}
	defer rows.Close()

	var territories []*types.Territory
	for rows.Next() {
		var territory types.Territory
		var conditionsJSON []byte

		err := rows.Scan(
			&territory.ID,
			&territory.OrganizationID,
			&territory.Name,
			&territory.Description,
			&territory.TerritoryType,
			&conditionsJSON,
			&territory.AssignedUsers,
			&territory.AssignedTeams,
			&territory.Priority,
			&territory.IsActive,
			&territory.CreatedAt,
			&territory.UpdatedAt,
			&territory.CreatedBy,
			&territory.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan territory: %w", err)
		}

		// Unmarshal conditions
		if err := json.Unmarshal(conditionsJSON, &territory.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		territories = append(territories, &territory)
	}

	return territories, nil
}

// AssignLead assigns a lead to a user
func (r *AssignmentRuleRepositoryPostgres) AssignLead(ctx context.Context, leadID uuid.UUID, userID uuid.UUID, reason string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update lead assignment
	leadQuery := `UPDATE leads SET assigned_to = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = tx.ExecContext(ctx, leadQuery, userID, leadID)
	if err != nil {
		return fmt.Errorf("failed to update lead assignment: %w", err)
	}

	// Record assignment history
	history := &types.AssignmentHistory{
		ID:               uuid.New(),
		TargetModel:      "leads",
		TargetID:         leadID,
		AssignedToType:   "user",
		AssignedToID:     userID,
		AssignmentReason: reason,
	}

	err = r.CreateAssignmentHistory(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to create assignment history: %w", err)
	}

	return tx.Commit()
}

// GetNextAssignee determines the next assignee based on assignment rules
func (r *AssignmentRuleRepositoryPostgres) GetNextAssignee(ctx context.Context, targetModel string, conditions map[string]interface{}) (uuid.UUID, string, error) {
	// Convert conditions to JSON for SQL query
	conditionsJSON, err := json.Marshal(conditions)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to marshal conditions: %w", err)
	}

	// Find matching assignment rule
	var ruleID uuid.UUID
	var ruleType string
	var ruleConfig json.RawMessage

	query := `
		SELECT id, rule_type, assignment_config
		FROM assignment_rules
		WHERE target_model = $1
		AND is_active = true
		AND conditions @> $2
		ORDER BY priority DESC
		LIMIT 1
	`

	err = r.db.QueryRowContext(ctx, query, targetModel, conditionsJSON).Scan(&ruleID, &ruleType, &ruleConfig)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, "", fmt.Errorf("no matching assignment rule found")
		}
		return uuid.Nil, "", fmt.Errorf("failed to find matching rule: %w", err)
	}

	// Determine assignee based on rule type
	var assigneeID uuid.UUID
	var assigneeName string

	switch ruleType {
	case "round_robin":
		// Call PostgreSQL function for round-robin assignment
		err = r.db.QueryRowContext(ctx, "SELECT get_next_round_robin_user($1)", ruleID).Scan(&assigneeID)
		if err != nil {
			return uuid.Nil, "", fmt.Errorf("failed to get round-robin assignee: %w", err)
		}

	case "weighted":
		// Call PostgreSQL function for weighted assignment
		err = r.db.QueryRowContext(ctx, "SELECT get_weighted_user($1, $2)", ruleID, targetModel).Scan(&assigneeID)
		if err != nil {
			return uuid.Nil, "", fmt.Errorf("failed to get weighted assignee: %w", err)
		}

	case "territory":
		// Match territory and get assigned users
		var territoryID uuid.UUID
		err = r.db.QueryRowContext(ctx, "SELECT match_territory($1, $2)", conditions["organization_id"], conditionsJSON).Scan(&territoryID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, "", fmt.Errorf("failed to match territory: %w", err)
		}

		if territoryID != uuid.Nil {
			// Get users assigned to this territory
			var assignedUsers []uuid.UUID
			query := `SELECT assigned_users FROM territories WHERE id = $1`
			err = r.db.QueryRowContext(ctx, query, territoryID).Scan(&assignedUsers)
			if err != nil {
				return uuid.Nil, "", fmt.Errorf("failed to get territory users: %w", err)
			}

			if len(assignedUsers) > 0 {
				assigneeID = assignedUsers[0] // Simple selection - could be enhanced
			}
		}

	case "custom":
		// For custom rules, we'd need additional logic
		// This is a placeholder for custom implementation
		return uuid.Nil, "", fmt.Errorf("custom assignment rules not yet implemented")
	}

	// Get assignee name
	if assigneeID != uuid.Nil {
		var name string
		query := `SELECT name FROM users WHERE id = $1`
		err = r.db.QueryRowContext(ctx, query, assigneeID).Scan(&name)
		if err != nil {
			return assigneeID, "", fmt.Errorf("failed to get assignee name: %w", err)
		}
		assigneeName = name
	}

	return assigneeID, assigneeName, nil
}

// GetAssignmentStatsByUser retrieves assignment statistics by user
func (r *AssignmentRuleRepositoryPostgres) GetAssignmentStatsByUser(ctx context.Context, orgID uuid.UUID, targetModel string) ([]*types.AssignmentStatsByUser, error) {
	query := `
		SELECT user_id, user_name, user_email, target_model, active_assignments,
		       total_assignments, last_assigned_at, weight, is_available, assignments_today
		FROM assignment_stats_by_user
		WHERE organization_id = $1
	`

	params := []interface{}{orgID}
	paramIndex := 2

	if targetModel != "" {
		query += fmt.Sprintf(" AND target_model = $%d", paramIndex)
		params = append(params, targetModel)
		paramIndex++
	}

	query += " ORDER BY active_assignments DESC"

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment stats by user: %w", err)
	}
	defer rows.Close()

	var stats []*types.AssignmentStatsByUser
	for rows.Next() {
		var stat types.AssignmentStatsByUser
		var lastAssignedAt *time.Time

		err := rows.Scan(
			&stat.UserID,
			&stat.UserName,
			&stat.UserEmail,
			&stat.TargetModel,
			&stat.ActiveAssignments,
			&stat.TotalAssignments,
			&lastAssignedAt,
			&stat.Weight,
			&stat.IsAvailable,
			&stat.AssignmentsToday,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment stats: %w", err)
		}

		if lastAssignedAt != nil {
			stat.LastAssignedAt = *lastAssignedAt
		}

		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetAssignmentRuleEffectiveness retrieves effectiveness metrics for assignment rules
func (r *AssignmentRuleRepositoryPostgres) GetAssignmentRuleEffectiveness(ctx context.Context, orgID uuid.UUID) ([]*types.AssignmentRuleEffectiveness, error) {
	query := `
		SELECT rule_id, rule_name, rule_type, target_model, is_active,
		       total_assignments, assignments_today, assignments_this_week,
		       last_used_at, unique_assignees
		FROM assignment_rule_effectiveness
		WHERE organization_id = $1
		ORDER BY total_assignments DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment rule effectiveness: %w", err)
	}
	defer rows.Close()

	var effectiveness []*types.AssignmentRuleEffectiveness
	for rows.Next() {
		var eff types.AssignmentRuleEffectiveness
		var lastUsedAt *time.Time

		err := rows.Scan(
			&eff.RuleID,
			&eff.RuleName,
			&eff.RuleType,
			&eff.TargetModel,
			&eff.IsActive,
			&eff.TotalAssignments,
			&eff.AssignmentsToday,
			&eff.AssignmentsThisWeek,
			&lastUsedAt,
			&eff.UniqueAssignees,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment rule effectiveness: %w", err)
		}

		if lastUsedAt != nil {
			eff.LastUsedAt = *lastUsedAt
		}

		effectiveness = append(effectiveness, &eff)
	}

	return effectiveness, nil
}

// FindActiveRules finds active assignment rules by target model
func (r *AssignmentRuleRepositoryPostgres) FindActiveRules(ctx context.Context, targetModel types.AssignmentTargetModel) ([]types.AssignmentRule, error) {
	// Get organization ID from context for security
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	if !ok {
		return nil, errors.New("organization ID not found in context")
	}

	query := `
		SELECT id, organization_id, name, description, rule_type, target_model,
			priority, is_active, conditions, assignment_config, assign_to_type,
			max_assignments_per_user, assignment_window_start, assignment_window_end,
			active_days, created_by, updated_by, created_at, updated_at
		FROM assignment_rules
		WHERE organization_id = $1 AND target_model = $2 AND is_active = true
		ORDER BY priority ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, targetModel)
	if err != nil {
		return nil, fmt.Errorf("failed to find active assignment rules: %w", err)
	}
	defer rows.Close()

	var rules []types.AssignmentRule
	for rows.Next() {
		var rule types.AssignmentRule
		var conditionsJSON, assignmentConfigJSON, activeDaysJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.OrganizationID,
			&rule.Name,
			&rule.Description,
			&rule.RuleType,
			&rule.TargetModel,
			&rule.Priority,
			&rule.IsActive,
			&conditionsJSON,
			&assignmentConfigJSON,
			&rule.AssignToType,
			&rule.MaxAssignmentsPerUser,
			&rule.AssignmentWindowStart,
			&rule.AssignmentWindowEnd,
			&activeDaysJSON,
			&rule.CreatedBy,
			&rule.UpdatedBy,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment rule: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
		if err := json.Unmarshal(assignmentConfigJSON, &rule.AssignmentConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignment config: %w", err)
		}
		if err := json.Unmarshal(activeDaysJSON, &rule.ActiveDays); err != nil {
			return nil, fmt.Errorf("failed to unmarshal active days: %w", err)
		}

		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignment rules: %w", err)
	}

	return rules, nil
}

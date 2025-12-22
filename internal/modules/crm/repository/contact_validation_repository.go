package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/google/uuid"
)

// ContactValidationRepository defines the interface for validation rule operations
type ContactValidationRepository interface {
	CreateValidationRule(ctx context.Context, rule *types.ContactValidationRule) error
	GetValidationRule(ctx context.Context, id uuid.UUID) (*types.ContactValidationRule, error)
	UpdateValidationRule(ctx context.Context, rule *types.ContactValidationRule) error
	DeleteValidationRule(ctx context.Context, id uuid.UUID) error
	ListValidationRules(ctx context.Context, filter types.ValidationRuleFilter) ([]*types.ContactValidationRule, error)
	CountValidationRules(ctx context.Context, filter types.ValidationRuleFilter) (int, error)
}

type contactValidationRepository struct {
	db *sql.DB
}

// NewContactValidationRepository creates a new validation repository
func NewContactValidationRepository(db *sql.DB) ContactValidationRepository {
	return &contactValidationRepository{db: db}
}

// CreateValidationRule creates a new validation rule
func (r *contactValidationRepository) CreateValidationRule(ctx context.Context, rule *types.ContactValidationRule) error {
	query := `
		INSERT INTO contact_validation_rules (
			id, organization_id, name, field, rule_type,
			validation_config, error_message, is_active, severity
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	err := r.db.QueryRowContext(ctx, query,
		rule.ID, rule.OrganizationID, rule.Name, rule.Field, rule.RuleType,
		rule.ValidationConfig, rule.ErrorMessage, rule.IsActive, rule.Severity,
	).Scan(&rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create validation rule: %w", err)
	}

	return nil
}

// GetValidationRule retrieves a validation rule by ID
func (r *contactValidationRepository) GetValidationRule(ctx context.Context, id uuid.UUID) (*types.ContactValidationRule, error) {
	query := `
		SELECT id, organization_id, name, field, rule_type,
			   validation_config, error_message, is_active, severity,
			   created_at, updated_at
		FROM contact_validation_rules
		WHERE id = $1 AND deleted_at IS NULL
	`

	rule := &types.ContactValidationRule{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.OrganizationID, &rule.Name, &rule.Field, &rule.RuleType,
		&rule.ValidationConfig, &rule.ErrorMessage, &rule.IsActive, &rule.Severity,
		&rule.CreatedAt, &rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get validation rule: %w", err)
	}

	return rule, nil
}

// UpdateValidationRule updates an existing validation rule
func (r *contactValidationRepository) UpdateValidationRule(ctx context.Context, rule *types.ContactValidationRule) error {
	query := `
		UPDATE contact_validation_rules
		SET name = $2, field = $3, rule_type = $4,
			validation_config = $5, error_message = $6,
			is_active = $7, severity = $8, updated_at = now()
		WHERE id = $1 AND organization_id = $9
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		rule.ID, rule.Name, rule.Field, rule.RuleType,
		rule.ValidationConfig, rule.ErrorMessage, rule.IsActive, rule.Severity,
		rule.OrganizationID,
	).Scan(&rule.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("validation rule not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update validation rule: %w", err)
	}

	return nil
}

// DeleteValidationRule soft deletes a validation rule
func (r *contactValidationRepository) DeleteValidationRule(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE contact_validation_rules SET is_active = false WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete validation rule: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("validation rule not found")
	}

	return nil
}

// ListValidationRules lists validation rules with filtering
func (r *contactValidationRepository) ListValidationRules(ctx context.Context, filter types.ValidationRuleFilter) ([]*types.ContactValidationRule, error) {
	query := `
		SELECT id, organization_id, name, field, rule_type,
			   validation_config, error_message, is_active, severity,
			   created_at, updated_at
		FROM contact_validation_rules
		WHERE organization_id = $1
	`
	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.Field != nil {
		query += fmt.Sprintf(" AND field = $%d", argPos)
		args = append(args, *filter.Field)
		argPos++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filter.IsActive)
		argPos++
	}

	if filter.Severity != nil {
		query += fmt.Sprintf(" AND severity = $%d", argPos)
		args = append(args, *filter.Severity)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list validation rules: %w", err)
	}
	defer rows.Close()

	var rules []*types.ContactValidationRule
	for rows.Next() {
		rule := &types.ContactValidationRule{}
		err := rows.Scan(
			&rule.ID, &rule.OrganizationID, &rule.Name, &rule.Field, &rule.RuleType,
			&rule.ValidationConfig, &rule.ErrorMessage, &rule.IsActive, &rule.Severity,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan validation rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// CountValidationRules counts validation rules with filtering
func (r *contactValidationRepository) CountValidationRules(ctx context.Context, filter types.ValidationRuleFilter) (int, error) {
	query := `SELECT COUNT(*) FROM contact_validation_rules WHERE organization_id = $1`
	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.Field != nil {
		query += fmt.Sprintf(" AND field = $%d", argPos)
		args = append(args, *filter.Field)
		argPos++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filter.IsActive)
		argPos++
	}

	if filter.Severity != nil {
		query += fmt.Sprintf(" AND severity = $%d", argPos)
		args = append(args, *filter.Severity)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count validation rules: %w", err)
	}

	return count, nil
}

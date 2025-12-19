package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"alieze-erp/internal/modules/crm/types"
	"github.com/google/uuid"
)

type contactTagRepository struct {
	db *sql.DB
}

func NewContactTagRepository(db *sql.DB) types.ContactTagRepository {
	return &contactTagRepository{db: db}
}

func (r *contactTagRepository) Create(ctx context.Context, tag types.ContactTag) (*types.ContactTag, error) {
	query := `INSERT INTO contact_tags (id, organization_id, name, color) VALUES ($1, $2, $3, $4) RETURNING id, organization_id, name, color, created_at`

	var created types.ContactTag
	err := r.db.QueryRowContext(ctx, query, tag.ID, tag.OrganizationID, tag.Name, tag.Color).Scan(
		&created.ID, &created.OrganizationID, &created.Name, &created.Color, &created.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create contact tag: %w", err)
	}

	return &created, nil
}

func (r *contactTagRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.ContactTag, error) {
	query := `SELECT id, organization_id, name, color, created_at FROM contact_tags WHERE id = $1`

	var tag types.ContactTag
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tag.ID, &tag.OrganizationID, &tag.Name, &tag.Color, &tag.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("contact tag not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get contact tag: %w", err)
	}

	return &tag, nil
}

func (r *contactTagRepository) FindAll(ctx context.Context, filter types.ContactTagFilter) ([]types.ContactTag, error) {
	query := `SELECT id, organization_id, name, color, created_at FROM contact_tags WHERE organization_id = $1`

	var args []interface{}
	args = append(args, filter.OrganizationID)

	if filter.Name != nil {
		query += " AND name LIKE $2"
		args = append(args, "%"+*filter.Name+"%")
	}

	query += " ORDER BY name"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query contact tags: %w", err)
	}
	defer rows.Close()

	var tags []types.ContactTag
	for rows.Next() {
		var tag types.ContactTag
		if err := rows.Scan(&tag.ID, &tag.OrganizationID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan contact tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contact tags: %w", err)
	}

	return tags, nil
}

func (r *contactTagRepository) Update(ctx context.Context, tag types.ContactTag) (*types.ContactTag, error) {
	query := `UPDATE contact_tags SET name = $1, color = $2 WHERE id = $3 RETURNING id, organization_id, name, color, created_at`

	var updated types.ContactTag
	err := r.db.QueryRowContext(ctx, query, tag.Name, tag.Color, tag.ID).Scan(
		&updated.ID, &updated.OrganizationID, &updated.Name, &updated.Color, &updated.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("contact tag not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update contact tag: %w", err)
	}

	return &updated, nil
}

func (r *contactTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM contact_tags WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contact tag not found: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *contactTagRepository) FindByContact(ctx context.Context, contactID uuid.UUID) ([]types.ContactTag, error) {
	// This would require a contact_tags_contacts junction table
	// For now, return empty slice as the junction table doesn't exist yet
	return []types.ContactTag{}, nil
}

func (r *contactTagRepository) Count(ctx context.Context, filter types.ContactTagFilter) (int, error) {
	query := `SELECT COUNT(*) FROM contact_tags WHERE organization_id = $1`

	var args []interface{}
	args = append(args, filter.OrganizationID)

	if filter.Name != nil {
		query += " AND name LIKE $2"
		args = append(args, "%"+*filter.Name+"%")
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count contact tags: %w", err)
	}

	return count, nil
}

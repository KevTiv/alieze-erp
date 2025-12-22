package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/google/uuid"
)

// ContactSegmentRepository defines the interface for segment operations
type ContactSegmentRepository interface {
	// Segment management
	CreateSegment(ctx context.Context, segment *types.ContactSegment) error
	GetSegment(ctx context.Context, id uuid.UUID) (*types.ContactSegment, error)
	UpdateSegment(ctx context.Context, segment *types.ContactSegment) error
	DeleteSegment(ctx context.Context, id uuid.UUID) error
	ListSegments(ctx context.Context, filter types.SegmentFilter) ([]*types.ContactSegment, error)
	CountSegments(ctx context.Context, filter types.SegmentFilter) (int, error)

	// Membership management
	AddContactsToSegment(ctx context.Context, segmentID uuid.UUID, contactIDs []uuid.UUID, addedBy uuid.UUID, orgID uuid.UUID) error
	RemoveContactsFromSegment(ctx context.Context, segmentID uuid.UUID, contactIDs []uuid.UUID) error
	GetSegmentMembers(ctx context.Context, segmentID uuid.UUID, limit, offset int) ([]*types.Contact, error)
	GetContactSegments(ctx context.Context, contactID uuid.UUID) ([]*types.ContactSegment, error)
	IsContactInSegment(ctx context.Context, segmentID uuid.UUID, contactID uuid.UUID) (bool, error)

	// Dynamic segment evaluation
	ClearSegmentMembers(ctx context.Context, segmentID uuid.UUID) error
	UpdateSegmentMemberCount(ctx context.Context, segmentID uuid.UUID, count int) error
}

type contactSegmentRepository struct {
	db *sql.DB
}

// NewContactSegmentRepository creates a new segment repository
func NewContactSegmentRepository(db *sql.DB) ContactSegmentRepository {
	return &contactSegmentRepository{db: db}
}

// CreateSegment creates a new contact segment
func (r *contactSegmentRepository) CreateSegment(ctx context.Context, segment *types.ContactSegment) error {
	query := `
		INSERT INTO contact_segments (
			id, organization_id, name, description, segment_type,
			criteria, color, icon, member_count, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`

	if segment.ID == uuid.Nil {
		segment.ID = uuid.New()
	}

	err := r.db.QueryRowContext(ctx, query,
		segment.ID, segment.OrganizationID, segment.Name, segment.Description,
		segment.SegmentType, segment.Criteria, segment.Color, segment.Icon,
		segment.MemberCount, segment.CreatedBy,
	).Scan(&segment.CreatedAt, &segment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create segment: %w", err)
	}

	return nil
}

// GetSegment retrieves a segment by ID
func (r *contactSegmentRepository) GetSegment(ctx context.Context, id uuid.UUID) (*types.ContactSegment, error) {
	query := `
		SELECT id, organization_id, name, description, segment_type,
			   criteria, color, icon, member_count, last_calculated_at,
			   created_at, updated_at, created_by
		FROM contact_segments
		WHERE id = $1
	`

	segment := &types.ContactSegment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&segment.ID, &segment.OrganizationID, &segment.Name, &segment.Description,
		&segment.SegmentType, &segment.Criteria, &segment.Color, &segment.Icon,
		&segment.MemberCount, &segment.LastCalculatedAt,
		&segment.CreatedAt, &segment.UpdatedAt, &segment.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get segment: %w", err)
	}

	return segment, nil
}

// UpdateSegment updates an existing segment
func (r *contactSegmentRepository) UpdateSegment(ctx context.Context, segment *types.ContactSegment) error {
	query := `
		UPDATE contact_segments
		SET name = $2, description = $3, criteria = $4,
			color = $5, icon = $6, updated_at = now()
		WHERE id = $1 AND organization_id = $7
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		segment.ID, segment.Name, segment.Description, segment.Criteria,
		segment.Color, segment.Icon, segment.OrganizationID,
	).Scan(&segment.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("segment not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update segment: %w", err)
	}

	return nil
}

// DeleteSegment deletes a segment
func (r *contactSegmentRepository) DeleteSegment(ctx context.Context, id uuid.UUID) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete segment members first
	_, err = tx.ExecContext(ctx, "DELETE FROM contact_segment_members WHERE segment_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete segment members: %w", err)
	}

	// Delete segment
	result, err := tx.ExecContext(ctx, "DELETE FROM contact_segments WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete segment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("segment not found")
	}

	return tx.Commit()
}

// ListSegments lists segments with filtering
func (r *contactSegmentRepository) ListSegments(ctx context.Context, filter types.SegmentFilter) ([]*types.ContactSegment, error) {
	query := `
		SELECT id, organization_id, name, description, segment_type,
			   criteria, color, icon, member_count, last_calculated_at,
			   created_at, updated_at, created_by
		FROM contact_segments
		WHERE organization_id = $1
	`
	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.SegmentType != nil {
		query += fmt.Sprintf(" AND segment_type = $%d", argPos)
		args = append(args, *filter.SegmentType)
		argPos++
	}

	if filter.Name != nil {
		query += fmt.Sprintf(" AND name ILIKE $%d", argPos)
		args = append(args, "%"+*filter.Name+"%")
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}
	defer rows.Close()

	var segments []*types.ContactSegment
	for rows.Next() {
		segment := &types.ContactSegment{}
		err := rows.Scan(
			&segment.ID, &segment.OrganizationID, &segment.Name, &segment.Description,
			&segment.SegmentType, &segment.Criteria, &segment.Color, &segment.Icon,
			&segment.MemberCount, &segment.LastCalculatedAt,
			&segment.CreatedAt, &segment.UpdatedAt, &segment.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

// CountSegments counts segments with filtering
func (r *contactSegmentRepository) CountSegments(ctx context.Context, filter types.SegmentFilter) (int, error) {
	query := `SELECT COUNT(*) FROM contact_segments WHERE organization_id = $1`
	args := []interface{}{filter.OrganizationID}
	argPos := 2

	if filter.SegmentType != nil {
		query += fmt.Sprintf(" AND segment_type = $%d", argPos)
		args = append(args, *filter.SegmentType)
		argPos++
	}

	if filter.Name != nil {
		query += fmt.Sprintf(" AND name ILIKE $%d", argPos)
		args = append(args, "%"+*filter.Name+"%")
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count segments: %w", err)
	}

	return count, nil
}

// AddContactsToSegment adds contacts to a segment
func (r *contactSegmentRepository) AddContactsToSegment(ctx context.Context, segmentID uuid.UUID, contactIDs []uuid.UUID, addedBy uuid.UUID, orgID uuid.UUID) error {
	if len(contactIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert members (ignore duplicates)
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO contact_segment_members (organization_id, segment_id, contact_id, added_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (organization_id, segment_id, contact_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, contactID := range contactIDs {
		_, err = stmt.ExecContext(ctx, orgID, segmentID, contactID, addedBy)
		if err != nil {
			return fmt.Errorf("failed to add contact to segment: %w", err)
		}
	}

	// Update member count
	_, err = tx.ExecContext(ctx, `
		UPDATE contact_segments
		SET member_count = (
			SELECT COUNT(*) FROM contact_segment_members WHERE segment_id = $1
		),
		updated_at = now()
		WHERE id = $1
	`, segmentID)
	if err != nil {
		return fmt.Errorf("failed to update member count: %w", err)
	}

	return tx.Commit()
}

// RemoveContactsFromSegment removes contacts from a segment
func (r *contactSegmentRepository) RemoveContactsFromSegment(ctx context.Context, segmentID uuid.UUID, contactIDs []uuid.UUID) error {
	if len(contactIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete members
	query := `DELETE FROM contact_segment_members WHERE segment_id = $1 AND contact_id = ANY($2)`
	_, err = tx.ExecContext(ctx, query, segmentID, contactIDs)
	if err != nil {
		return fmt.Errorf("failed to remove contacts from segment: %w", err)
	}

	// Update member count
	_, err = tx.ExecContext(ctx, `
		UPDATE contact_segments
		SET member_count = (
			SELECT COUNT(*) FROM contact_segment_members WHERE segment_id = $1
		),
		updated_at = now()
		WHERE id = $1
	`, segmentID)
	if err != nil {
		return fmt.Errorf("failed to update member count: %w", err)
	}

	return tx.Commit()
}

// GetSegmentMembers retrieves contacts in a segment
func (r *contactSegmentRepository) GetSegmentMembers(ctx context.Context, segmentID uuid.UUID, limit, offset int) ([]*types.Contact, error) {
	query := `
		SELECT c.id, c.organization_id, c.name, c.email, c.phone,
			   c.is_customer, c.is_vendor, c.street, c.city, c.state_id,
			   c.country_id, c.created_at, c.updated_at
		FROM contacts c
		INNER JOIN contact_segment_members csm ON c.id = csm.contact_id
		WHERE csm.segment_id = $1
		ORDER BY csm.added_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, segmentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get segment members: %w", err)
	}
	defer rows.Close()

	var contacts []*types.Contact
	for rows.Next() {
		contact := &types.Contact{}
		err := rows.Scan(
			&contact.ID, &contact.OrganizationID, &contact.Name, &contact.Email,
			&contact.Phone, &contact.IsCustomer, &contact.IsVendor, &contact.Street,
			&contact.City, &contact.StateID, &contact.CountryID,
			&contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// GetContactSegments retrieves all segments a contact belongs to
func (r *contactSegmentRepository) GetContactSegments(ctx context.Context, contactID uuid.UUID) ([]*types.ContactSegment, error) {
	query := `
		SELECT s.id, s.organization_id, s.name, s.description, s.segment_type,
			   s.criteria, s.color, s.icon, s.member_count, s.last_calculated_at,
			   s.created_at, s.updated_at, s.created_by
		FROM contact_segments s
		INNER JOIN contact_segment_members csm ON s.id = csm.segment_id
		WHERE csm.contact_id = $1
		ORDER BY s.name
	`

	rows, err := r.db.QueryContext(ctx, query, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact segments: %w", err)
	}
	defer rows.Close()

	var segments []*types.ContactSegment
	for rows.Next() {
		segment := &types.ContactSegment{}
		err := rows.Scan(
			&segment.ID, &segment.OrganizationID, &segment.Name, &segment.Description,
			&segment.SegmentType, &segment.Criteria, &segment.Color, &segment.Icon,
			&segment.MemberCount, &segment.LastCalculatedAt,
			&segment.CreatedAt, &segment.UpdatedAt, &segment.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

// IsContactInSegment checks if a contact is in a segment
func (r *contactSegmentRepository) IsContactInSegment(ctx context.Context, segmentID uuid.UUID, contactID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM contact_segment_members WHERE segment_id = $1 AND contact_id = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, segmentID, contactID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check segment membership: %w", err)
	}

	return exists, nil
}

// ClearSegmentMembers removes all members from a segment (for dynamic segment recalculation)
func (r *contactSegmentRepository) ClearSegmentMembers(ctx context.Context, segmentID uuid.UUID) error {
	query := `DELETE FROM contact_segment_members WHERE segment_id = $1`
	_, err := r.db.ExecContext(ctx, query, segmentID)
	if err != nil {
		return fmt.Errorf("failed to clear segment members: %w", err)
	}

	return nil
}

// UpdateSegmentMemberCount updates the cached member count
func (r *contactSegmentRepository) UpdateSegmentMemberCount(ctx context.Context, segmentID uuid.UUID, count int) error {
	query := `
		UPDATE contact_segments
		SET member_count = $2, last_calculated_at = now(), updated_at = now()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, segmentID, count)
	if err != nil {
		return fmt.Errorf("failed to update segment member count: %w", err)
	}

	return nil
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// contactRepository handles contact data operations
type contactRepository struct {
	db *sql.DB
}

func NewContactRepository(db *sql.DB) types.ContactRepository {
	return &contactRepository{db: db}
}

func (r *contactRepository) Create(ctx context.Context, contact types.Contact) (*types.Contact, error) {
	if contact.ID == uuid.Nil {
		contact.ID = uuid.New()
	}

	if contact.OrganizationID == uuid.Nil {
		return nil, errors.New("organization_id is required")
	}

	if contact.Name == "" {
		return nil, errors.New("name is required")
	}

	query := `
		INSERT INTO contacts (
			id, organization_id, name, email, phone, is_customer, is_vendor,
			street, city, state_id, country_id, created_at, updated_at, deleted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, organization_id, name, email, phone, is_customer, is_vendor,
			street, city, state_id, country_id, created_at, updated_at, deleted_at
	`

	now := time.Now()

	result := r.db.QueryRowContext(ctx, query,
		contact.ID,
		contact.OrganizationID,
		contact.Name,
		contact.Email,
		contact.Phone,
		contact.IsCustomer,
		contact.IsVendor,
		contact.Street,
		contact.City,
		contact.StateID,
		contact.CountryID,
		now,
		now,
		nil,
	)

	var created types.Contact
	err := result.Scan(
		&created.ID,
		&created.OrganizationID,
		&created.Name,
		&created.Email,
		&created.Phone,
		&created.IsCustomer,
		&created.IsVendor,
		&created.Street,
		&created.City,
		&created.StateID,
		&created.CountryID,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return &created, nil
}

func (r *contactRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid contact id")
	}

	query := `
		SELECT id, organization_id, name, email, phone, is_customer, is_vendor,
			street, city, state_id, country_id, created_at, updated_at, deleted_at
		FROM contacts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var contact types.Contact
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&contact.ID,
		&contact.OrganizationID,
		&contact.Name,
		&contact.Email,
		&contact.Phone,
		&contact.IsCustomer,
		&contact.IsVendor,
		&contact.Street,
		&contact.City,
		&contact.StateID,
		&contact.CountryID,
		&contact.CreatedAt,
		&contact.UpdatedAt,
		&contact.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("contact not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return &contact, nil
}

func (r *contactRepository) FindAll(ctx context.Context, filter types.ContactFilter) ([]*types.Contact, error) {
	query := `SELECT id, organization_id, name, email, phone, is_customer, is_vendor,
		street, city, state_id, country_id, created_at, updated_at, deleted_at
		FROM contacts WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	var argIndex = 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, filter.OrganizationID)
	argIndex++

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if filter.Email != nil && *filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	if filter.Phone != nil && *filter.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Phone+"%")
		argIndex++
	}

	if filter.IsCustomer != nil {
		conditions = append(conditions, fmt.Sprintf("is_customer = $%d", argIndex))
		args = append(args, *filter.IsCustomer)
		argIndex++
	}

	if filter.IsVendor != nil {
		conditions = append(conditions, fmt.Sprintf("is_vendor = $%d", argIndex))
		args = append(args, *filter.IsVendor)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY name ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*types.Contact
	for rows.Next() {
		var contact types.Contact
		err := rows.Scan(
			&contact.ID,
			&contact.OrganizationID,
			&contact.Name,
			&contact.Email,
			&contact.Phone,
			&contact.IsCustomer,
			&contact.IsVendor,
			&contact.Street,
			&contact.City,
			&contact.StateID,
			&contact.CountryID,
			&contact.CreatedAt,
			&contact.UpdatedAt,
			&contact.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, &contact)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during contact iteration: %w", err)
	}

	return contacts, nil
}

func (r *contactRepository) Update(ctx context.Context, contact types.Contact) (*types.Contact, error) {
	if contact.ID == uuid.Nil {
		return nil, errors.New("contact id is required")
	}

	if contact.OrganizationID == uuid.Nil {
		return nil, errors.New("organization_id is required")
	}

	if contact.Name == "" {
		return nil, errors.New("name is required")
	}

	contact.UpdatedAt = time.Now()

	query := `
		UPDATE contacts SET
			organization_id = $1,
			name = $2,
			email = $3,
			phone = $4,
			is_customer = $5,
			is_vendor = $6,
			street = $7,
			city = $8,
			state_id = $9,
			country_id = $10,
			updated_at = $11
		WHERE id = $12 AND deleted_at IS NULL
		RETURNING id, organization_id, name, email, phone, is_customer, is_vendor,
			street, city, state_id, country_id, created_at, updated_at, deleted_at
	`

	result := r.db.QueryRowContext(ctx, query,
		contact.OrganizationID,
		contact.Name,
		contact.Email,
		contact.Phone,
		contact.IsCustomer,
		contact.IsVendor,
		contact.Street,
		contact.City,
		contact.StateID,
		contact.CountryID,
		contact.UpdatedAt,
		contact.ID,
	)

	var updated types.Contact
	err := result.Scan(
		&updated.ID,
		&updated.OrganizationID,
		&updated.Name,
		&updated.Email,
		&updated.Phone,
		&updated.IsCustomer,
		&updated.IsVendor,
		&updated.Street,
		&updated.City,
		&updated.StateID,
		&updated.CountryID,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return &updated, nil
}

func (r *contactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid contact id")
	}

	query := `
		UPDATE contacts SET
			deleted_at = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contact not found or already deleted")
	}

	return nil
}

// ContactRelationship methods

func (r *contactRepository) CreateRelationship(ctx context.Context, relationship *types.ContactRelationship) error {
	if relationship.ID == uuid.Nil {
		relationship.ID = uuid.New()
	}

	if relationship.OrganizationID == uuid.Nil {
		return errors.New("organization_id is required")
	}

	if relationship.ContactID == uuid.Nil {
		return errors.New("contact_id is required")
	}

	if relationship.RelatedContactID == uuid.Nil {
		return errors.New("related_contact_id is required")
	}

	if !types.IsValidRelationshipType(relationship.Type) {
		return errors.New("invalid relationship type")
	}

	if relationship.CreatedAt.IsZero() {
		relationship.CreatedAt = time.Now()
	}

	if relationship.UpdatedAt.IsZero() {
		relationship.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO contact_relationships
		(id, organization_id, contact_id, related_contact_id, type, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		relationship.ID,
		relationship.OrganizationID,
		relationship.ContactID,
		relationship.RelatedContactID,
		relationship.Type,
		relationship.Notes,
		relationship.CreatedAt,
		relationship.UpdatedAt,
	)

	return err
}

func (r *contactRepository) FindRelationships(
	ctx context.Context,
	orgID uuid.UUID,
	contactID uuid.UUID,
	relationshipType string,
	limit int,
) ([]*types.ContactRelationship, error) {
	if orgID == uuid.Nil {
		return nil, errors.New("organization_id is required")
	}

	if contactID == uuid.Nil {
		return nil, errors.New("contact_id is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT
			id, organization_id, contact_id, related_contact_id, type, notes, created_at, updated_at
		FROM contact_relationships
		WHERE organization_id = $1 AND contact_id = $2
	`

	params := []interface{}{orgID, contactID}

	if relationshipType != "" {
		if !types.IsValidRelationshipType(types.ContactRelationshipType(relationshipType)) {
			return nil, errors.New("invalid relationship type")
		}
		query += " AND type = $3"
		params = append(params, relationshipType)
	}

	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(params)+1)
	params = append(params, limit)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*types.ContactRelationship
	for rows.Next() {
		var rel types.ContactRelationship
		err := rows.Scan(
			&rel.ID,
			&rel.OrganizationID,
			&rel.ContactID,
			&rel.RelatedContactID,
			&rel.Type,
			&rel.Notes,
			&rel.CreatedAt,
			&rel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		relationships = append(relationships, &rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return relationships, nil
}

func (r *contactRepository) ContactExists(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) (bool, error) {
	if orgID == uuid.Nil {
		return false, errors.New("organization_id is required")
	}

	if contactID == uuid.Nil {
		return false, errors.New("contact_id is required")
	}

	query := `SELECT 1 FROM contacts WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, orgID, contactID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check contact existence: %w", err)
	}

	return exists, nil
}

func (r *contactRepository) AddContactToSegments(
	ctx context.Context,
	orgID uuid.UUID,
	contactID uuid.UUID,
	segmentIDs []string,
) error {
	if orgID == uuid.Nil {
		return errors.New("organization_id is required")
	}

	if contactID == uuid.Nil {
		return errors.New("contact_id is required")
	}

	if len(segmentIDs) == 0 {
		return nil
	}

	// Use transaction for multiple segment assignments
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, segmentID := range segmentIDs {
		// Validate segment ID format
		if _, err := uuid.Parse(segmentID); err != nil {
			return fmt.Errorf("invalid segment ID format: %s", segmentID)
		}

		query := `
			INSERT INTO contact_segments (organization_id, contact_id, segment_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (organization_id, contact_id, segment_id) DO NOTHING
		`

		_, err := tx.ExecContext(ctx, query, orgID, contactID, segmentID)
		if err != nil {
			return fmt.Errorf("failed to add to segment %s: %w", segmentID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *contactRepository) AddContactTags(
	ctx context.Context,
	orgID uuid.UUID,
	contactID uuid.UUID,
	tags []string,
) error {
	if orgID == uuid.Nil {
		return errors.New("organization_id is required")
	}

	if contactID == uuid.Nil {
		return errors.New("contact_id is required")
	}

	if len(tags) == 0 {
		return nil
	}

	// Normalize and deduplicate tags
	tagMap := make(map[string]bool)
	for _, tag := range tags {
		normalized := strings.TrimSpace(strings.ToLower(tag))
		if normalized != "" {
			tagMap[normalized] = true
		}
	}

	if len(tagMap) == 0 {
		return nil
	}

	// Use transaction for multiple tag assignments
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for tag := range tagMap {
		query := `
			INSERT INTO contact_tags (organization_id, contact_id, tag)
			VALUES ($1, $2, $3)
			ON CONFLICT (organization_id, contact_id, tag) DO NOTHING
		`

		_, err := tx.ExecContext(ctx, query, orgID, contactID, tag)
		if err != nil {
			return fmt.Errorf("failed to add tag %s: %w", tag, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *contactRepository) Count(ctx context.Context, filter types.ContactFilter) (int, error) {
	query := `SELECT COUNT(*) FROM contacts WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	var argIndex = 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
	args = append(args, filter.OrganizationID)
	argIndex++

	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if filter.Email != nil && *filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	if filter.Phone != nil && *filter.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Phone+"%")
		argIndex++
	}

	if filter.IsCustomer != nil {
		conditions = append(conditions, fmt.Sprintf("is_customer = $%d", argIndex))
		args = append(args, *filter.IsCustomer)
		argIndex++
	}

	if filter.IsVendor != nil {
		conditions = append(conditions, fmt.Sprintf("is_vendor = $%d", argIndex))
		args = append(args, *filter.IsVendor)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count contacts: %w", err)
	}

	return count, nil
}

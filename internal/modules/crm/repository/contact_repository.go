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

// ContactRepository handles contact data operations
type ContactRepository struct {
	db *sql.DB
}

func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(ctx context.Context, contact types.Contact) (*types.Contact, error) {
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

func (r *ContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
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

func (r *ContactRepository) FindAll(ctx context.Context, filter types.ContactFilter) ([]types.Contact, error) {
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

	var contacts []types.Contact
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
		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during contact iteration: %w", err)
	}

	return contacts, nil
}

func (r *ContactRepository) Update(ctx context.Context, contact types.Contact) (*types.Contact, error) {
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

func (r *ContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
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

func (r *ContactRepository) Count(ctx context.Context, filter types.ContactFilter) (int, error) {
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

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
)

// CountryRepository handles country data operations
type CountryRepository struct {
	db *sql.DB
}

func NewCountryRepository(db *sql.DB) *CountryRepository {
	return &CountryRepository{db: db}
}

func (r *CountryRepository) Create(ctx context.Context, country types.Country) (*types.Country, error) {
	if country.ID == uuid.Nil {
		country.ID = uuid.New()
	}

	if country.Code == "" {
		return nil, errors.New("code is required")
	}

	if country.Name == "" {
		return nil, errors.New("name is required")
	}

	query := `
		INSERT INTO countries (
			id, code, name, phone_code, currency_id, address_format, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, code, name, phone_code, currency_id, address_format, created_at
	`

	now := time.Now()

	var created types.Country
	err := r.db.QueryRowContext(ctx, query,
		country.ID, country.Code, country.Name, country.PhoneCode, country.CurrencyID, country.AddressFormat, now,
	).Scan(
		&created.ID, &created.Code, &created.Name, &created.PhoneCode, &created.CurrencyID, &created.AddressFormat, &created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *CountryRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.Country, error) {
	query := `
		SELECT id, code, name, phone_code, currency_id, address_format, created_at
		FROM countries
		WHERE id = $1
	`

	var country types.Country
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&country.ID, &country.Code, &country.Name, &country.PhoneCode, &country.CurrencyID, &country.AddressFormat, &country.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &country, nil
}

func (r *CountryRepository) GetByCode(ctx context.Context, code string) (*types.Country, error) {
	query := `
		SELECT id, code, name, phone_code, currency_id, address_format, created_at
		FROM countries
		WHERE code = $1
	`

	var country types.Country
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&country.ID, &country.Code, &country.Name, &country.PhoneCode, &country.CurrencyID, &country.AddressFormat, &country.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &country, nil
}

func (r *CountryRepository) List(ctx context.Context, filter types.CountryFilter) ([]types.Country, error) {
	query := `
		SELECT id, code, name, phone_code, currency_id, address_format, created_at
		FROM countries
		WHERE (1=1)
	`

	params := []interface{}{}
	paramIndex := 1

	if filter.Code != nil {
		query += fmt.Sprintf(" AND code = $%d", paramIndex)
		params = append(params, *filter.Code)
		paramIndex++
	}

	if filter.Name != nil {
		query += fmt.Sprintf(" AND name ILIKE $%d", paramIndex)
		params = append(params, "%"+*filter.Name+"%")
		paramIndex++
	}

	query += " ORDER BY name"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIndex)
		params = append(params, filter.Limit)
		paramIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", paramIndex)
		params = append(params, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []types.Country
	for rows.Next() {
		var country types.Country
		err := rows.Scan(
			&country.ID, &country.Code, &country.Name, &country.PhoneCode, &country.CurrencyID, &country.AddressFormat, &country.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}

	return countries, nil
}

func (r *CountryRepository) Update(ctx context.Context, id uuid.UUID, update types.CountryUpdateRequest) (*types.Country, error) {
	if update.Code == nil && update.Name == nil && update.PhoneCode == nil && update.CurrencyID == nil && update.AddressFormat == nil {
		return nil, errors.New("no fields to update")
	}

	query := "UPDATE countries SET "
	params := []interface{}{}
	paramIndex := 1

	if update.Code != nil {
		query += fmt.Sprintf("code = $%d, ", paramIndex)
		params = append(params, *update.Code)
		paramIndex++
	}

	if update.Name != nil {
		query += fmt.Sprintf("name = $%d, ", paramIndex)
		params = append(params, *update.Name)
		paramIndex++
	}

	if update.PhoneCode != nil {
		query += fmt.Sprintf("phone_code = $%d, ", paramIndex)
		params = append(params, *update.PhoneCode)
		paramIndex++
	}

	if update.CurrencyID != nil {
		query += fmt.Sprintf("currency_id = $%d, ", paramIndex)
		params = append(params, *update.CurrencyID)
		paramIndex++
	}

	if update.AddressFormat != nil {
		query += fmt.Sprintf("address_format = $%d, ", paramIndex)
		params = append(params, *update.AddressFormat)
		paramIndex++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, code, name, phone_code, currency_id, address_format, created_at", paramIndex)
	params = append(params, id)

	var updated types.Country
	err := r.db.QueryRowContext(ctx, query, params...).Scan(
		&updated.ID, &updated.Code, &updated.Name, &updated.PhoneCode, &updated.CurrencyID, &updated.AddressFormat, &updated.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *CountryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM countries WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

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

// CurrencyRepository handles currency data operations
type CurrencyRepository struct {
	db *sql.DB
}

func NewCurrencyRepository(db *sql.DB) *CurrencyRepository {
	return &CurrencyRepository{db: db}
}

func (r *CurrencyRepository) Create(ctx context.Context, currency types.Currency) (*types.Currency, error) {
	if currency.ID == uuid.Nil {
		currency.ID = uuid.New()
	}

	if currency.Name == "" {
		return nil, errors.New("name is required")
	}

	if currency.Code == "" {
		return nil, errors.New("code is required")
	}

	if currency.Symbol == "" {
		return nil, errors.New("symbol is required")
	}

	if currency.Position == "" {
		currency.Position = types.CurrencyPositionBefore
	}

	if currency.Rounding == 0 {
		currency.Rounding = 0.01
	}

	if currency.DecimalPlaces == 0 {
		currency.DecimalPlaces = 2
	}

	if !currency.Active {
		currency.Active = true
	}

	query := `
		INSERT INTO currencies (
			id, name, symbol, code, rounding, decimal_places, position, active, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, name, symbol, code, rounding, decimal_places, position, active, created_at
	`

	now := time.Now()

	var created types.Currency
	err := r.db.QueryRowContext(ctx, query,
		currency.ID, currency.Name, currency.Symbol, currency.Code,
		currency.Rounding, currency.DecimalPlaces, currency.Position,
		currency.Active, now,
	).Scan(
		&created.ID, &created.Name, &created.Symbol, &created.Code,
		&created.Rounding, &created.DecimalPlaces, &created.Position,
		&created.Active, &created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *CurrencyRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.Currency, error) {
	query := `
		SELECT id, name, symbol, code, rounding, decimal_places, position, active, created_at
		FROM currencies
		WHERE id = $1
	`

	var currency types.Currency
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&currency.ID, &currency.Name, &currency.Symbol, &currency.Code,
		&currency.Rounding, &currency.DecimalPlaces, &currency.Position,
		&currency.Active, &currency.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &currency, nil
}

func (r *CurrencyRepository) GetByCode(ctx context.Context, code string) (*types.Currency, error) {
	query := `
		SELECT id, name, symbol, code, rounding, decimal_places, position, active, created_at
		FROM currencies
		WHERE code = $1
	`

	var currency types.Currency
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&currency.ID, &currency.Name, &currency.Symbol, &currency.Code,
		&currency.Rounding, &currency.DecimalPlaces, &currency.Position,
		&currency.Active, &currency.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &currency, nil
}

func (r *CurrencyRepository) List(ctx context.Context, filter types.CurrencyFilter) ([]types.Currency, error) {
	query := `
		SELECT id, name, symbol, code, rounding, decimal_places, position, active, created_at
		FROM currencies
		WHERE (1=1)
	`

	params := []interface{}{}
	paramIndex := 1

	if filter.Active != nil {
		query += fmt.Sprintf(" AND active = $%d", paramIndex)
		params = append(params, *filter.Active)
		paramIndex++
	}

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

	var currencies []types.Currency
	for rows.Next() {
		var currency types.Currency
		err := rows.Scan(
			&currency.ID, &currency.Name, &currency.Symbol, &currency.Code,
			&currency.Rounding, &currency.DecimalPlaces, &currency.Position,
			&currency.Active, &currency.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	return currencies, nil
}

func (r *CurrencyRepository) Update(ctx context.Context, id uuid.UUID, update types.CurrencyUpdateRequest) (*types.Currency, error) {
	if update.Name == nil && update.Symbol == nil && update.Code == nil &&
	   update.Rounding == nil && update.DecimalPlaces == nil &&
	   update.Position == nil && update.Active == nil {
		return nil, errors.New("no fields to update")
	}

	query := "UPDATE currencies SET "
	params := []interface{}{}
	paramIndex := 1

	if update.Name != nil {
		query += fmt.Sprintf("name = $%d, ", paramIndex)
		params = append(params, *update.Name)
		paramIndex++
	}

	if update.Symbol != nil {
		query += fmt.Sprintf("symbol = $%d, ", paramIndex)
		params = append(params, *update.Symbol)
		paramIndex++
	}

	if update.Code != nil {
		query += fmt.Sprintf("code = $%d, ", paramIndex)
		params = append(params, *update.Code)
		paramIndex++
	}

	if update.Rounding != nil {
		query += fmt.Sprintf("rounding = $%d, ", paramIndex)
		params = append(params, *update.Rounding)
		paramIndex++
	}

	if update.DecimalPlaces != nil {
		query += fmt.Sprintf("decimal_places = $%d, ", paramIndex)
		params = append(params, *update.DecimalPlaces)
		paramIndex++
	}

	if update.Position != nil {
		query += fmt.Sprintf("position = $%d, ", paramIndex)
		params = append(params, *update.Position)
		paramIndex++
	}

	if update.Active != nil {
		query += fmt.Sprintf("active = $%d, ", paramIndex)
		params = append(params, *update.Active)
		paramIndex++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name, symbol, code, rounding, decimal_places, position, active, created_at", paramIndex)
	params = append(params, id)

	var updated types.Currency
	err := r.db.QueryRowContext(ctx, query, params...).Scan(
		&updated.ID, &updated.Name, &updated.Symbol, &updated.Code,
		&updated.Rounding, &updated.DecimalPlaces, &updated.Position,
		&updated.Active, &updated.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *CurrencyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM currencies WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

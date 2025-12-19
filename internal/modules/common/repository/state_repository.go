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

// StateRepository handles state data operations
type StateRepository struct {
	db *sql.DB
}

func NewStateRepository(db *sql.DB) *StateRepository {
	return &StateRepository{db: db}
}

func (r *StateRepository) Create(ctx context.Context, state types.State) (*types.State, error) {
	if state.ID == uuid.Nil {
		state.ID = uuid.New()
	}

	if state.CountryID == uuid.Nil {
		return nil, errors.New("country_id is required")
	}

	if state.Name == "" {
		return nil, errors.New("name is required")
	}

	if state.Code == "" {
		return nil, errors.New("code is required")
	}

	query := `
		INSERT INTO states (
			id, country_id, name, code, created_at
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING id, country_id, name, code, created_at
	`

	now := time.Now()

	var created types.State
	err := r.db.QueryRowContext(ctx, query,
		state.ID, state.CountryID, state.Name, state.Code, now,
	).Scan(
		&created.ID, &created.CountryID, &created.Name, &created.Code, &created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *StateRepository) GetByID(ctx context.Context, id uuid.UUID) (*types.State, error) {
	query := `
		SELECT id, country_id, name, code, created_at
		FROM states
		WHERE id = $1
	`

	var state types.State
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&state.ID, &state.CountryID, &state.Name, &state.Code, &state.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &state, nil
}

func (r *StateRepository) ListByCountry(ctx context.Context, countryID uuid.UUID) ([]types.State, error) {
	query := `
		SELECT id, country_id, name, code, created_at
		FROM states
		WHERE country_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []types.State
	for rows.Next() {
		var state types.State
		err := rows.Scan(
			&state.ID, &state.CountryID, &state.Name, &state.Code, &state.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}

	return states, nil
}

func (r *StateRepository) List(ctx context.Context, filter types.StateFilter) ([]types.State, error) {
	query := `
		SELECT id, country_id, name, code, created_at
		FROM states
		WHERE (1=1)
	`

	params := []interface{}{}
	paramIndex := 1

	if filter.CountryID != nil {
		query += fmt.Sprintf(" AND country_id = $%d", paramIndex)
		params = append(params, *filter.CountryID)
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

	var states []types.State
	for rows.Next() {
		var state types.State
		err := rows.Scan(
			&state.ID, &state.CountryID, &state.Name, &state.Code, &state.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}

	return states, nil
}

func (r *StateRepository) Update(ctx context.Context, id uuid.UUID, update types.StateUpdateRequest) (*types.State, error) {
	if update.CountryID == nil && update.Name == nil && update.Code == nil {
		return nil, errors.New("no fields to update")
	}

	query := "UPDATE states SET "
	params := []interface{}{}
	paramIndex := 1

	if update.CountryID != nil {
		query += fmt.Sprintf("country_id = $%d, ", paramIndex)
		params = append(params, *update.CountryID)
		paramIndex++
	}

	if update.Name != nil {
		query += fmt.Sprintf("name = $%d, ", paramIndex)
		params = append(params, *update.Name)
		paramIndex++
	}

	if update.Code != nil {
		query += fmt.Sprintf("code = $%d, ", paramIndex)
		params = append(params, *update.Code)
		paramIndex++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, country_id, name, code, created_at", paramIndex)
	params = append(params, id)

	var updated types.State
	err := r.db.QueryRowContext(ctx, query, params...).Scan(
		&updated.ID, &updated.CountryID, &updated.Name, &updated.Code, &updated.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *StateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM states WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

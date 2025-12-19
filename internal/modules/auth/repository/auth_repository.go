package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KevTiv/alieze-erp/internal/modules/auth/types"

	"github.com/google/uuid"
)

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateUser(ctx context.Context, user types.User) (*types.User, error) {
	query := `
		INSERT INTO auth.users
		(id, email, encrypted_password, email_confirmed_at, confirmed_at, created_at, updated_at, is_super_admin)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, email, encrypted_password, email_confirmed_at, confirmed_at, created_at, updated_at, is_super_admin
	`

	var created types.User
	var emailConfirmedAt, confirmedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Email, user.EncryptedPassword, user.EmailConfirmedAt, user.ConfirmedAt,
		user.CreatedAt, user.UpdatedAt, user.IsSuperAdmin,
	).Scan(
		&created.ID, &created.Email, &created.EncryptedPassword, &emailConfirmedAt, &confirmedAt,
		&created.CreatedAt, &created.UpdatedAt, &created.IsSuperAdmin,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Handle nullable time fields
	if emailConfirmedAt.Valid {
		created.EmailConfirmedAt = &emailConfirmedAt.Time
	}
	if confirmedAt.Valid {
		created.ConfirmedAt = &confirmedAt.Time
	}

	return &created, nil
}

func (r *authRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*types.User, error) {
	query := `
		SELECT id, email, encrypted_password, email_confirmed_at, confirmed_at, last_sign_in_at,
		       created_at, updated_at, is_super_admin
		FROM auth.users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user types.User
	var emailConfirmedAt, confirmedAt, lastSignInAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.EncryptedPassword, &emailConfirmedAt, &confirmedAt,
		&lastSignInAt, &user.CreatedAt, &user.UpdatedAt, &user.IsSuperAdmin,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	// Handle nullable time fields
	if emailConfirmedAt.Valid {
		user.EmailConfirmedAt = &emailConfirmedAt.Time
	}
	if confirmedAt.Valid {
		user.ConfirmedAt = &confirmedAt.Time
	}
	if lastSignInAt.Valid {
		user.LastSignInAt = &lastSignInAt.Time
	}

	return &user, nil
}

func (r *authRepository) FindUserByEmail(ctx context.Context, email string) (*types.User, error) {
	query := `
		SELECT id, email, encrypted_password, email_confirmed_at, confirmed_at, last_sign_in_at,
		       created_at, updated_at, is_super_admin
		FROM auth.users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user types.User
	var emailConfirmedAt, confirmedAt, lastSignInAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.EncryptedPassword, &emailConfirmedAt, &confirmedAt,
		&lastSignInAt, &user.CreatedAt, &user.UpdatedAt, &user.IsSuperAdmin,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	// Handle nullable time fields
	if emailConfirmedAt.Valid {
		user.EmailConfirmedAt = &emailConfirmedAt.Time
	}
	if confirmedAt.Valid {
		user.ConfirmedAt = &confirmedAt.Time
	}
	if lastSignInAt.Valid {
		user.LastSignInAt = &lastSignInAt.Time
	}

	return &user, nil
}

func (r *authRepository) UpdateUser(ctx context.Context, user types.User) (*types.User, error) {
	query := `
		UPDATE auth.users
		SET email = $2, email_confirmed_at = $3, confirmed_at = $4, last_sign_in_at = $5,
		    updated_at = $6, is_super_admin = $7
		WHERE id = $1
		RETURNING id, email, encrypted_password, email_confirmed_at, confirmed_at, last_sign_in_at,
		          created_at, updated_at, is_super_admin
	`

	var updated types.User
	var emailConfirmedAt, confirmedAt, lastSignInAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Email, user.EmailConfirmedAt, user.ConfirmedAt, user.LastSignInAt,
		user.UpdatedAt, user.IsSuperAdmin,
	).Scan(
		&updated.ID, &updated.Email, &updated.EncryptedPassword, &emailConfirmedAt, &confirmedAt,
		&lastSignInAt, &updated.CreatedAt, &updated.UpdatedAt, &updated.IsSuperAdmin,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Handle nullable time fields
	if emailConfirmedAt.Valid {
		updated.EmailConfirmedAt = &emailConfirmedAt.Time
	}
	if confirmedAt.Valid {
		updated.ConfirmedAt = &confirmedAt.Time
	}
	if lastSignInAt.Valid {
		updated.LastSignInAt = &lastSignInAt.Time
	}

	return &updated, nil
}

func (r *authRepository) CreateOrganization(ctx context.Context, name string, createdBy uuid.UUID) (*uuid.UUID, error) {
	query := `
		INSERT INTO organizations (name, slug, created_by, updated_by)
		VALUES ($1, $2, $3, $3)
		RETURNING id
	`

	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, name, generateSlug(name), createdBy).Scan(&orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &orgID, nil
}

func (r *authRepository) FindOrganizationByID(ctx context.Context, id uuid.UUID) (*string, error) {
	query := `SELECT name FROM organizations WHERE id = $1 AND deleted_at IS NULL`

	var name string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	return &name, nil
}

func (r *authRepository) CreateOrganizationUser(ctx context.Context, orgUser types.OrganizationUser) (*types.OrganizationUser, error) {
	query := `
		INSERT INTO organization_users
		(id, organization_id, user_id, role, is_active, joined_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, organization_id, user_id, role, is_active, joined_at, created_at, updated_at
	`

	var created types.OrganizationUser
	err := r.db.QueryRowContext(ctx, query,
		orgUser.ID, orgUser.OrganizationID, orgUser.UserID, orgUser.Role,
		orgUser.IsActive, orgUser.JoinedAt, orgUser.CreatedAt, orgUser.UpdatedAt,
	).Scan(
		&created.ID, &created.OrganizationID, &created.UserID, &created.Role,
		&created.IsActive, &created.JoinedAt, &created.CreatedAt, &created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create organization user: %w", err)
	}

	return &created, nil
}

func (r *authRepository) FindOrganizationUser(ctx context.Context, orgID, userID uuid.UUID) (*types.OrganizationUser, error) {
	query := `
		SELECT id, organization_id, user_id, role, is_active, joined_at, created_at, updated_at
		FROM organization_users
		WHERE organization_id = $1 AND user_id = $2
	`

	var orgUser types.OrganizationUser
	var joinedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, orgID, userID).Scan(
		&orgUser.ID, &orgUser.OrganizationID, &orgUser.UserID, &orgUser.Role,
		&orgUser.IsActive, &joinedAt, &orgUser.CreatedAt, &orgUser.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find organization user: %w", err)
	}

	// Handle nullable time field
	if joinedAt.Valid {
		orgUser.JoinedAt = joinedAt.Time
	}

	return &orgUser, nil
}

func (r *authRepository) FindOrganizationUsersByUserID(ctx context.Context, userID uuid.UUID) ([]types.OrganizationUser, error) {
	query := `
		SELECT id, organization_id, user_id, role, is_active, joined_at, created_at, updated_at
		FROM organization_users
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find organization users by user id: %w", err)
	}
	defer rows.Close()

	var orgUsers []types.OrganizationUser
	for rows.Next() {
		var orgUser types.OrganizationUser
		var joinedAt sql.NullTime
		err := rows.Scan(
			&orgUser.ID, &orgUser.OrganizationID, &orgUser.UserID, &orgUser.Role,
			&orgUser.IsActive, &joinedAt, &orgUser.CreatedAt, &orgUser.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization user: %w", err)
		}

		// Handle nullable time field
		if joinedAt.Valid {
			orgUser.JoinedAt = joinedAt.Time
		}

		orgUsers = append(orgUsers, orgUser)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating organization users: %w", err)
	}

	return orgUsers, nil
}

func (r *authRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, encryptedPassword string) error {
	query := `UPDATE auth.users SET encrypted_password = $2, updated_at = now() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID, encryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}

// Helper function to generate slug from organization name
func generateSlug(name string) string {
	// Simple slug generation - replace spaces with hyphens and lowercase
	slug := ""
	for _, c := range name {
		if c == ' ' {
			slug += "-"
		} else {
			slug += string(c)
		}
	}
	return slug
}

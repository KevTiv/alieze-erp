package repository

import (
	"context"

	"alieze-erp/internal/modules/auth/types"

	"github.com/google/uuid"
)

// AuthRepository defines the interface for authentication data access
type AuthRepository interface {
	// User operations
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (*domain.User, error)

	// Organization operations
	CreateOrganization(ctx context.Context, name string, createdBy uuid.UUID) (*uuid.UUID, error)
	FindOrganizationByID(ctx context.Context, id uuid.UUID) (*string, error)

	// Organization user operations
	CreateOrganizationUser(ctx context.Context, orgUser domain.OrganizationUser) (*domain.OrganizationUser, error)
	FindOrganizationUser(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationUser, error)
	FindOrganizationUsersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.OrganizationUser, error)

	// Password operations
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, encryptedPassword string) error
}

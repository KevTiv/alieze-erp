package repository

import (
	"context"
	"errors"
	"time"

	"alieze-erp/internal/modules/auth/types"

	"github.com/google/uuid"
)

// MockAuthRepository is a mock implementation of AuthRepository for testing
type MockAuthRepository struct {
	users             map[uuid.UUID]domain.User
	usersByEmail      map[string]domain.User
	organizations     map[uuid.UUID]string
	organizationUsers map[uuid.UUID][]domain.OrganizationUser
	passwordUpdates   map[uuid.UUID]string
	errors            map[string]error
}

func NewMockAuthRepository() *MockAuthRepository {
	return &MockAuthRepository{
		users:             make(map[uuid.UUID]domain.User),
		usersByEmail:      make(map[string]domain.User),
		organizations:     make(map[uuid.UUID]string),
		organizationUsers: make(map[uuid.UUID][]domain.OrganizationUser),
		passwordUpdates:   make(map[uuid.UUID]string),
		errors:            make(map[string]error),
	}
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	if err, exists := m.errors["CreateUser"]; exists {
		return nil, err
	}

	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	return &user, nil
}

func (m *MockAuthRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if err, exists := m.errors["FindUserByID"]; exists {
		return nil, err
	}

	if user, exists := m.users[id]; exists {
		return &user, nil
	}
	return nil, nil
}

func (m *MockAuthRepository) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if err, exists := m.errors["FindUserByEmail"]; exists {
		return nil, err
	}

	if user, exists := m.usersByEmail[email]; exists {
		return &user, nil
	}
	return nil, nil
}

func (m *MockAuthRepository) UpdateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	if err, exists := m.errors["UpdateUser"]; exists {
		return nil, err
	}

	if existing, exists := m.users[user.ID]; exists {
		user.EncryptedPassword = existing.EncryptedPassword // Preserve password
		m.users[user.ID] = user
		m.usersByEmail[user.Email] = user
		return &user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockAuthRepository) CreateOrganization(ctx context.Context, name string, createdBy uuid.UUID) (*uuid.UUID, error) {
	if err, exists := m.errors["CreateOrganization"]; exists {
		return nil, err
	}

	orgID := uuid.New()
	m.organizations[orgID] = name
	return &orgID, nil
}

func (m *MockAuthRepository) FindOrganizationByID(ctx context.Context, id uuid.UUID) (*string, error) {
	if err, exists := m.errors["FindOrganizationByID"]; exists {
		return nil, err
	}

	if name, exists := m.organizations[id]; exists {
		return &name, nil
	}
	return nil, nil
}

func (m *MockAuthRepository) CreateOrganizationUser(ctx context.Context, orgUser domain.OrganizationUser) (*domain.OrganizationUser, error) {
	if err, exists := m.errors["CreateOrganizationUser"]; exists {
		return nil, err
	}

	m.organizationUsers[orgUser.UserID] = append(m.organizationUsers[orgUser.UserID], orgUser)
	return &orgUser, nil
}

func (m *MockAuthRepository) FindOrganizationUser(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationUser, error) {
	if err, exists := m.errors["FindOrganizationUser"]; exists {
		return nil, err
	}

	if orgUsers, exists := m.organizationUsers[userID]; exists {
		for _, orgUser := range orgUsers {
			if orgUser.OrganizationID == orgID {
				return &orgUser, nil
			}
		}
	}
	return nil, nil
}

func (m *MockAuthRepository) FindOrganizationUsersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.OrganizationUser, error) {
	if err, exists := m.errors["FindOrganizationUsersByUserID"]; exists {
		return nil, err
	}

	if orgUsers, exists := m.organizationUsers[userID]; exists {
		return orgUsers, nil
	}
	return []domain.OrganizationUser{}, nil
}

func (m *MockAuthRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, encryptedPassword string) error {
	if err, exists := m.errors["UpdateUserPassword"]; exists {
		return err
	}

	if user, exists := m.users[userID]; exists {
		user.EncryptedPassword = encryptedPassword
		m.users[userID] = user
		m.passwordUpdates[userID] = encryptedPassword
		return nil
	}
	return errors.New("user not found")
}

// Helper methods for testing
func (m *MockAuthRepository) AddUser(user domain.User) {
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
}

func (m *MockAuthRepository) AddOrganization(orgID uuid.UUID, name string) {
	m.organizations[orgID] = name
}

func (m *MockAuthRepository) AddOrganizationUser(orgUser domain.OrganizationUser) {
	m.organizationUsers[orgUser.UserID] = append(m.organizationUsers[orgUser.UserID], orgUser)
}

func (m *MockAuthRepository) SetError(method string, err error) {
	m.errors[method] = err
}

func (m *MockAuthRepository) ClearErrors() {
	m.errors = make(map[string]error)
}

// CreateTestUser creates a test user with default values
func CreateTestUser() domain.User {
	now := time.Now()
	return domain.User{
		ID:                uuid.New(),
		Email:             "test@example.com",
		EncryptedPassword: "hashed-password",
		EmailConfirmedAt:  &now,
		ConfirmedAt:       &now,
		CreatedAt:         now,
		UpdatedAt:         now,
		IsSuperAdmin:      false,
	}
}

// CreateTestOrganizationUser creates a test organization user
func CreateTestOrganizationUser(userID, orgID uuid.UUID) domain.OrganizationUser {
	now := time.Now()
	return domain.OrganizationUser{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           "owner",
		IsActive:       true,
		JoinedAt:       now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

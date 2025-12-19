package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/internal/testutils"
)

type ContactServiceTestSuite struct {
	suite.Suite
	service   *service.ContactService
	repo      *testutils.MockContactRepository
	auth      *testutils.MockAuthService
	ctx       context.Context
	orgID     uuid.UUID
	userID    uuid.UUID
	contactID uuid.UUID
}

func (s *ContactServiceTestSuite) SetupTest() {
	s.T().Log("Setting up test")

	s.repo = testutils.NewMockContactRepository()
	s.auth = testutils.NewMockAuthService()
	s.service = service.NewContactService(s.repo, s.auth)
	s.ctx = context.Background()
	s.orgID = uuid.Must(uuid.NewV7())
	s.userID = uuid.Must(uuid.NewV7())
	s.contactID = uuid.Must(uuid.NewV7())

	// Default mock behavior
	s.auth.WithOrganizationID(s.orgID)
	s.auth.WithUserID(s.userID)
	s.auth.AllowPermission("contacts:create")
	s.auth.AllowPermission("contacts:read")
	s.auth.AllowPermission("contacts:update")
	s.auth.AllowPermission("contacts:delete")
}

func (s *ContactServiceTestSuite) TearDownTest() {
	s.T().Log("Tearing down test")
	// Clean up if needed
}

func (s *ContactServiceTestSuite) TestCreateContactSuccess() {
	s.T().Run("CreateContact - Success", func(t *testing.T) {
		// Setup test data
		contact := types.Contact{
			Name:       "John Doe",
			Email:      stringPtr("john@example.com"),
			Phone:      stringPtr("1234567890"),
			IsCustomer: true,
		}

		// Mock repository behavior
		expectedContact := contact
		expectedContact.ID = s.contactID
		expectedContact.OrganizationID = s.orgID
		expectedContact.CreatedAt = time.Now()
		expectedContact.UpdatedAt = time.Now()

		s.repo.WithCreateFunc(func(ctx context.Context, c types.Contact) (*types.Contact, error) {
			require.Equal(t, s.orgID, c.OrganizationID)
			require.Equal(t, contact.Name, c.Name)
			require.Equal(t, contact.Email, c.Email)
			require.Equal(t, contact.Phone, c.Phone)
			require.Equal(t, contact.IsCustomer, c.IsCustomer)
			return &expectedContact, nil
		})

		// Execute
		created, err := s.service.CreateContact(s.ctx, contact)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, expectedContact.ID, created.ID)
		require.Equal(t, expectedContact.OrganizationID, created.OrganizationID)
		require.Equal(t, expectedContact.Name, created.Name)
		require.Equal(t, expectedContact.Email, created.Email)
		require.Equal(t, expectedContact.Phone, created.Phone)
		require.Equal(t, expectedContact.IsCustomer, created.IsCustomer)
		require.NotZero(t, created.CreatedAt)
		require.NotZero(t, created.UpdatedAt)
	})
}

func (s *ContactServiceTestSuite) TestCreateContactValidationError() {
	s.T().Run("CreateContact - Validation Error", func(t *testing.T) {
		// Test cases with validation errors
		testCases := []struct {
			name        string
			contact     types.Contact
			expectedErr string
		}{
			{
				name: "Empty Name",
				contact: types.Contact{
					Email: stringPtr("john@example.com"),
				},
				expectedErr: "contact name is required",
			},
			{
				name: "Invalid Email",
				contact: types.Contact{
					Name:  "John Doe",
					Email: stringPtr("invalid-email"),
				},
				expectedErr: "invalid email format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Execute
				created, err := s.service.CreateContact(s.ctx, tc.contact)

				// Assert
				require.Error(t, err)
				require.Nil(t, created)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})
}

func (s *ContactServiceTestSuite) TestCreateContactPermissionError() {
	s.T().Run("CreateContact - Permission Error", func(t *testing.T) {
		// Setup test data
		contact := types.Contact{
			Name:  "John Doe",
			Email: stringPtr("john@example.com"),
		}

		// Mock permission denial
		s.auth.DenyPermission("contacts:create")

		// Execute
		created, err := s.service.CreateContact(s.ctx, contact)

		// Assert
		require.Error(t, err)
		require.Nil(t, created)
		require.Contains(t, err.Error(), "permission denied")
	})
}

func (s *ContactServiceTestSuite) TestGetContactSuccess() {
	s.T().Run("GetContact - Success", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID
		expectedContact := types.Contact{
			ID:             contactID,
			OrganizationID: s.orgID,
			Name:           "John Doe",
			Email:          stringPtr("john@example.com"),
			Phone:          stringPtr("1234567890"),
			IsCustomer:     true,
		}

		// Mock repository behavior
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
			require.Equal(t, contactID, id)
			return &expectedContact, nil
		})

		// Execute
		contact, err := s.service.GetContact(s.ctx, contactID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, contact)
		require.Equal(t, expectedContact.ID, contact.ID)
		require.Equal(t, expectedContact.OrganizationID, contact.OrganizationID)
		require.Equal(t, expectedContact.Name, contact.Name)
		require.Equal(t, expectedContact.Email, contact.Email)
		require.Equal(t, expectedContact.Phone, contact.Phone)
		require.Equal(t, expectedContact.IsCustomer, contact.IsCustomer)
	})
}

func (s *ContactServiceTestSuite) TestGetContactPermissionError() {
	s.T().Run("GetContact - Permission Error", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID

		// Mock permission denial
		s.auth.DenyPermission("contacts:read")

		// Execute
		contact, err := s.service.GetContact(s.ctx, contactID)

		// Assert
		require.Error(t, err)
		require.Nil(t, contact)
		require.Contains(t, err.Error(), "permission denied")
	})
}

func (s *ContactServiceTestSuite) TestGetContactOrganizationMismatch() {
	s.T().Run("GetContact - Organization Mismatch", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID
		otherOrgID := uuid.Must(uuid.NewV7())

		// Mock repository behavior - return contact from different organization
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
			return &types.Contact{
				ID:             contactID,
				OrganizationID: otherOrgID, // Different organization
				Name:           "John Doe",
			}, nil
		})

		// Execute
		contact, err := s.service.GetContact(s.ctx, contactID)

		// Assert
		require.Error(t, err)
		require.Nil(t, contact)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

func (s *ContactServiceTestSuite) TestListContactsSuccess() {
	s.T().Run("ListContacts - Success", func(t *testing.T) {
		// Setup test data
		filter := types.ContactFilter{
			Name:  stringPtr("John"),
			Limit: 10,
		}

		// Mock repository behavior
		expectedContacts := []types.Contact{
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "John Doe",
				Email:          stringPtr("john@example.com"),
				IsCustomer:     true,
			},
			{
				ID:             uuid.Must(uuid.NewV7()),
				OrganizationID: s.orgID,
				Name:           "John Smith",
				Email:          stringPtr("john.smith@example.com"),
				IsCustomer:     false,
			},
		}
		expectedCount := 2

		s.repo.WithFindAllFunc(func(ctx context.Context, f types.ContactFilter) ([]types.Contact, error) {
			require.Equal(t, s.orgID, f.OrganizationID)
			require.Equal(t, filter.Name, f.Name)
			require.Equal(t, filter.Limit, f.Limit)
			return expectedContacts, nil
		})

		s.repo.WithCountFunc(func(ctx context.Context, f types.ContactFilter) (int, error) {
			return expectedCount, nil
		})

		// Execute
		contacts, count, err := s.service.ListContacts(s.ctx, filter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, contacts)
		require.Len(t, contacts, 2)
		require.Equal(t, expectedCount, count)
		require.Equal(t, expectedContacts[0].Name, contacts[0].Name)
		require.Equal(t, expectedContacts[1].Name, contacts[1].Name)
	})
}

func (s *ContactServiceTestSuite) TestUpdateContactSuccess() {
	s.T().Run("UpdateContact - Success", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID
		contact := types.Contact{
			ID:         contactID,
			Name:       "John Doe Updated",
			Email:      stringPtr("john.updated@example.com"),
			Phone:      stringPtr("5555555555"),
			IsCustomer: false,
			IsVendor:   true,
		}

		// Mock repository behavior
		existingContact := types.Contact{
			ID:             contactID,
			OrganizationID: s.orgID,
			Name:           "John Doe",
			Email:          stringPtr("john@example.com"),
			IsCustomer:     true,
		}

		expectedContact := contact
		expectedContact.OrganizationID = s.orgID
		expectedContact.UpdatedAt = time.Now()

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
			require.Equal(t, contactID, id)
			return &existingContact, nil
		})

		s.repo.WithUpdateFunc(func(ctx context.Context, c types.Contact) (*types.Contact, error) {
			require.Equal(t, contactID, c.ID)
			require.Equal(t, s.orgID, c.OrganizationID)
			require.Equal(t, contact.Name, c.Name)
			require.Equal(t, contact.Email, c.Email)
			require.Equal(t, contact.Phone, c.Phone)
			require.Equal(t, contact.IsCustomer, c.IsCustomer)
			require.Equal(t, contact.IsVendor, c.IsVendor)
			return &expectedContact, nil
		})

		// Execute
		updated, err := s.service.UpdateContact(s.ctx, contact)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, updated)
		require.Equal(t, expectedContact.ID, updated.ID)
		require.Equal(t, expectedContact.OrganizationID, updated.OrganizationID)
		require.Equal(t, expectedContact.Name, updated.Name)
		require.Equal(t, expectedContact.Email, updated.Email)
		require.Equal(t, expectedContact.Phone, updated.Phone)
		require.Equal(t, expectedContact.IsCustomer, updated.IsCustomer)
		require.Equal(t, expectedContact.IsVendor, updated.IsVendor)
		require.NotZero(t, updated.UpdatedAt)
	})
}

func (s *ContactServiceTestSuite) TestUpdateContactValidationError() {
	s.T().Run("UpdateContact - Validation Error", func(t *testing.T) {
		// Test cases with validation errors
		testCases := []struct {
			name        string
			contact     types.Contact
			expectedErr string
			mockRepo    bool // Whether to mock the repository for organization check
		}{
			{
				name: "Empty ID",
				contact: types.Contact{
					Name: "John Doe",
				},
				expectedErr: "contact id is required",
				mockRepo:    false,
			},
			{
				name: "Empty Name",
				contact: types.Contact{
					ID: s.contactID,
				},
				expectedErr: "contact name is required",
				mockRepo:    false,
			},
			{
				name: "Invalid Email",
				contact: types.Contact{
					ID:    s.contactID,
					Name:  "John Doe",
					Email: stringPtr("invalid-email"),
				},
				expectedErr: "invalid email format",
				mockRepo:    true, // Need to mock repo for organization check
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// For cases that need organization check, mock the repository
				if tc.mockRepo {
					existingContact := types.Contact{
						ID:             s.contactID,
						OrganizationID: s.orgID,
						Name:           "John Doe",
					}

					s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
						return &existingContact, nil
					})
				}

				// Execute
				updated, err := s.service.UpdateContact(s.ctx, tc.contact)

				// Assert
				require.Error(t, err)
				require.Nil(t, updated)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})
}

func (s *ContactServiceTestSuite) TestUpdateContactPermissionError() {
	s.T().Run("UpdateContact - Permission Error", func(t *testing.T) {
		// Setup test data
		contact := types.Contact{
			ID:   s.contactID,
			Name: "John Doe",
		}

		// Mock repository behavior - contact belongs to the organization
		existingContact := types.Contact{
			ID:             s.contactID,
			OrganizationID: s.orgID,
			Name:           "John Doe",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
			return &existingContact, nil
		})

		// Mock permission denial
		s.auth.DenyPermission("contacts:update")

		// Execute
		updated, err := s.service.UpdateContact(s.ctx, contact)

		// Assert
		require.Error(t, err)
		require.Nil(t, updated)
		require.Contains(t, err.Error(), "permission denied")
	})
}

func (s *ContactServiceTestSuite) TestUpdateContactOrganizationMismatch() {
	s.T().Run("UpdateContact - Organization Mismatch", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID
		otherOrgID := uuid.Must(uuid.NewV7())
		contact := types.Contact{
			ID:   contactID,
			Name: "John Doe",
		}

		// Mock repository behavior - return contact from different organization
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
			return &types.Contact{
				ID:             contactID,
				OrganizationID: otherOrgID, // Different organization
				Name:           "John Doe",
			}, nil
		})

		// Execute
		updated, err := s.service.UpdateContact(s.ctx, contact)

		// Assert
		require.Error(t, err)
		require.Nil(t, updated)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

func (s *ContactServiceTestSuite) TestDeleteContactSuccess() {
	s.T().Run("DeleteContact - Success", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID

		// Mock repository behavior
		existingContact := types.Contact{
			ID:             contactID,
			OrganizationID: s.orgID,
			Name:           "John Doe",
		}

		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
			require.Equal(t, contactID, id)
			return &existingContact, nil
		})

		s.repo.WithDeleteFunc(func(ctx context.Context, id uuid.UUID) error {
			require.Equal(t, contactID, id)
			return nil
		})

		// Execute
		err := s.service.DeleteContact(s.ctx, contactID)

		// Assert
		require.NoError(t, err)
	})
}

func (s *ContactServiceTestSuite) TestDeleteContactPermissionError() {
	s.T().Run("DeleteContact - Permission Error", func(t *testing.T) {
		// Setup test data - contact belongs to the organization
		existingContact := types.Contact{
			ID:             s.contactID,
			OrganizationID: s.orgID,
			Name:           "John Doe",
		}

		// Mock repository behavior
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
			return &existingContact, nil
		})

		// Mock permission denial
		s.auth.DenyPermission("contacts:delete")

		// Execute
		err := s.service.DeleteContact(s.ctx, s.contactID)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "permission denied")
	})
}

func (s *ContactServiceTestSuite) TestDeleteContactOrganizationMismatch() {
	s.T().Run("DeleteContact - Organization Mismatch", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID
		otherOrgID := uuid.Must(uuid.NewV7())

		// Mock repository behavior - return contact from different organization
		s.repo.WithFindByIDFunc(func(ctx context.Context, id uuid.UUID) (*types.Contact, error) {
			return &types.Contact{
				ID:             contactID,
				OrganizationID: otherOrgID, // Different organization
				Name:           "John Doe",
			}, nil
		})

		// Execute
		err := s.service.DeleteContact(s.ctx, contactID)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not belong to organization")
	})
}

// Run the test suite
func TestContactServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ContactServiceTestSuite))
}

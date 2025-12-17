package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"alieze-erp/internal/modules/crm/domain"
	"alieze-erp/internal/modules/crm/repository"
	"alieze-erp/internal/testutils"
)

type ContactRepositoryTestSuite struct {
	suite.Suite
	repo      *repository.ContactRepository
	mockDB    *testutils.MockDB
	ctx       context.Context
	orgID     uuid.UUID
	contactID uuid.UUID
}

func (s *ContactRepositoryTestSuite) SetupTest() {
	s.T().Log("Setting up test")

	s.mockDB = testutils.SetupMockDB(s.T())
	s.repo = repository.NewContactRepository(s.mockDB.DB)
	s.ctx = context.Background()
	s.orgID = uuid.Must(uuid.NewV7())
	s.contactID = uuid.Must(uuid.NewV7())
}

func (s *ContactRepositoryTestSuite) TearDownTest() {
	s.T().Log("Tearing down test")
	s.mockDB.Close()
}

func (s *ContactRepositoryTestSuite) TestCreateContactSuccess() {
	s.T().Run("CreateContact - Success", func(t *testing.T) {
		// Setup test data
		contact := types.Contact{
			ID:             s.contactID,
			OrganizationID: s.orgID,
			Name:           "John Doe",
			Email:          stringPtr("john@example.com"),
			Phone:          stringPtr("1234567890"),
			IsCustomer:     true,
			IsVendor:       false,
			Street:         stringPtr("123 Main St"),
			City:           stringPtr("New York"),
			StateID:        nil,
			CountryID:      nil,
		}

		// Expected query and result
		now := time.Now()
		s.mockDB.Mock.ExpectQuery("INSERT INTO contacts").
			WithArgs(
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
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				nil,              // deleted_at
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "email", "phone", "is_customer", "is_vendor",
				"street", "city", "state_id", "country_id", "created_at", "updated_at", "deleted_at",
			}).AddRow(
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
			))

		// Execute
		created, err := s.repo.Create(s.ctx, contact)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, contact.ID, created.ID)
		require.Equal(t, contact.Name, created.Name)
		require.Equal(t, contact.Email, created.Email)
		require.Equal(t, contact.Phone, created.Phone)
		require.Equal(t, contact.IsCustomer, created.IsCustomer)
		require.Equal(t, contact.IsVendor, created.IsVendor)
		require.Equal(t, contact.Street, created.Street)
		require.Equal(t, contact.City, created.City)
		require.Equal(t, contact.StateID, created.StateID)
		require.Equal(t, contact.CountryID, created.CountryID)
		require.NotZero(t, created.CreatedAt)
		require.NotZero(t, created.UpdatedAt)
		require.Nil(t, created.DeletedAt)

		// Ensure all expectations were met
		err = s.mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}

func (s *ContactRepositoryTestSuite) TestCreateContactValidationError() {
	s.T().Run("CreateContact - Validation Error", func(t *testing.T) {
		// Test cases with validation errors
		testCases := []struct {
			name        string
			contact     types.Contact
			expectedErr string
		}{
			{
				name: "Empty OrganizationID",
				contact: types.Contact{
					ID:   s.contactID,
					Name: "John Doe",
				},
				expectedErr: "organization_id is required",
			},
			{
				name: "Empty Name",
				contact: types.Contact{
					ID:             s.contactID,
					OrganizationID: s.orgID,
				},
				expectedErr: "name is required",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Execute
				created, err := s.repo.Create(s.ctx, tc.contact)

				// Assert
				require.Error(t, err)
				require.Nil(t, created)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})
}

func (s *ContactRepositoryTestSuite) TestFindByIDSuccess() {
	s.T().Run("FindByID - Success", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID
		email := "john@example.com"
		phone := "1234567890"
		now := time.Now()

		// Expected query and result
		s.mockDB.Mock.ExpectQuery("SELECT.*FROM contacts").
			WithArgs(contactID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "email", "phone", "is_customer", "is_vendor",
				"street", "city", "state_id", "country_id", "created_at", "updated_at", "deleted_at",
			}).AddRow(
				contactID,
				s.orgID,
				"John Doe",
				&email,
				&phone,
				true,
				false,
				"123 Main St",
				"New York",
				nil,
				nil,
				now,
				now,
				nil,
			))

		// Execute
		contact, err := s.repo.FindByID(s.ctx, contactID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, contact)
		require.Equal(t, contactID, contact.ID)
		require.Equal(t, s.orgID, contact.OrganizationID)
		require.Equal(t, "John Doe", contact.Name)
		require.Equal(t, "john@example.com", *contact.Email)
		require.Equal(t, "1234567890", *contact.Phone)
		require.True(t, contact.IsCustomer)
		require.False(t, contact.IsVendor)
		require.Equal(t, "123 Main St", *contact.Street)
		require.Equal(t, "New York", *contact.City)
		require.Nil(t, contact.StateID)
		require.Nil(t, contact.CountryID)
		require.NotZero(t, contact.CreatedAt)
		require.NotZero(t, contact.UpdatedAt)
		require.Nil(t, contact.DeletedAt)

		// Ensure all expectations were met
		err = s.mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}

func (s *ContactRepositoryTestSuite) TestFindByIDNotFound() {
	s.T().Run("FindByID - Not Found", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID

		// Expected query and result
		s.mockDB.Mock.ExpectQuery("SELECT.*FROM contacts").
			WithArgs(contactID).
			WillReturnError(sql.ErrNoRows)

		// Execute
		contact, err := s.repo.FindByID(s.ctx, contactID)

		// Assert
		require.Error(t, err)
		require.Nil(t, contact)
		require.Contains(t, err.Error(), "contact not found")

		// Ensure all expectations were met
		err = s.mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}

func (s *ContactRepositoryTestSuite) TestFindAllSuccess() {
	s.T().Run("FindAll - Success", func(t *testing.T) {
		// Setup test data
		filter := types.ContactFilter{
			OrganizationID: s.orgID,
			Name:           stringPtr("John"),
			Limit:          10,
			Offset:         0,
		}

		// Expected query and result
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "organization_id", "name", "email", "phone", "is_customer", "is_vendor",
			"street", "city", "state_id", "country_id", "created_at", "updated_at", "deleted_at",
		}).
			AddRow(uuid.Must(uuid.NewV7()), s.orgID, "John Doe", "john@example.com", "1234567890", true, false, "123 Main St", "New York", nil, nil, now, now, nil).
			AddRow(uuid.Must(uuid.NewV7()), s.orgID, "John Smith", "john.smith@example.com", "9876543210", false, true, "456 Oak Ave", "Boston", nil, nil, now, now, nil)

		s.mockDB.Mock.ExpectQuery("SELECT.*FROM contacts").
			WithArgs(s.orgID, "%John%").
			WillReturnRows(rows)

		// Execute
		contacts, err := s.repo.FindAll(s.ctx, filter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, contacts)
		require.Len(t, contacts, 2)
		require.Equal(t, "John Doe", contacts[0].Name)
		require.Equal(t, "John Smith", contacts[1].Name)

		// Ensure all expectations were met
		err = s.mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}

func (s *ContactRepositoryTestSuite) TestUpdateContactSuccess() {
	s.T().Run("UpdateContact - Success", func(t *testing.T) {
		// Setup test data
		contact := types.Contact{
			ID:             s.contactID,
			OrganizationID: s.orgID,
			Name:           "John Doe Updated",
			Email:          stringPtr("john.updated@example.com"),
			Phone:          stringPtr("5555555555"),
			IsCustomer:     false,
			IsVendor:       true,
			Street:         stringPtr("456 Updated St"),
			City:           stringPtr("Updated City"),
			StateID:        nil,
			CountryID:      nil,
		}

		// Expected query and result
		now := time.Now()
		contact.UpdatedAt = now

		s.mockDB.Mock.ExpectQuery("UPDATE contacts").
			WithArgs(
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
				sqlmock.AnyArg(), // updated_at - use AnyArg to avoid time precision issues
				contact.ID,
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "organization_id", "name", "email", "phone", "is_customer", "is_vendor",
				"street", "city", "state_id", "country_id", "created_at", "updated_at", "deleted_at",
			}).AddRow(
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
				now.Add(-time.Hour), // created_at (different from updated_at)
				now,                 // updated_at
				nil,
			))

		// Execute
		updated, err := s.repo.Update(s.ctx, contact)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, updated)
		require.Equal(t, contact.ID, updated.ID)
		require.Equal(t, contact.Name, updated.Name)
		require.Equal(t, contact.Email, updated.Email)
		require.Equal(t, contact.Phone, updated.Phone)
		require.Equal(t, contact.IsCustomer, updated.IsCustomer)
		require.Equal(t, contact.IsVendor, updated.IsVendor)
		require.Equal(t, contact.Street, updated.Street)
		require.Equal(t, contact.City, updated.City)
		require.Equal(t, contact.StateID, updated.StateID)
		require.Equal(t, contact.CountryID, updated.CountryID)
		require.NotZero(t, updated.CreatedAt)
		require.NotZero(t, updated.UpdatedAt)
		require.Nil(t, updated.DeletedAt)

		// Ensure all expectations were met
		err = s.mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}

func (s *ContactRepositoryTestSuite) TestDeleteContactSuccess() {
	s.T().Run("DeleteContact - Success", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID

		// Expected query and result
		result := sqlmock.NewResult(0, 1) // 1 row affected
		s.mockDB.Mock.ExpectExec("UPDATE contacts").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), contactID).
			WillReturnResult(result)

		// Execute
		err := s.repo.Delete(s.ctx, contactID)

		// Assert
		require.NoError(t, err)

		// Ensure all expectations were met
		err = s.mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}

func (s *ContactRepositoryTestSuite) TestDeleteContactNotFound() {
	s.T().Run("DeleteContact - Not Found", func(t *testing.T) {
		// Setup test data
		contactID := s.contactID
		// Expected query and result
		result := sqlmock.NewResult(0, 0) // 0 rows affected
		s.mockDB.Mock.ExpectExec("UPDATE contacts").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), contactID).
			WillReturnResult(result)

		// Execute
		err := s.repo.Delete(s.ctx, contactID)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found or already deleted")

		// Ensure all expectations were met
		err = s.mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}

func (s *ContactRepositoryTestSuite) TestCountContactsSuccess() {
	s.T().Run("Count - Success", func(t *testing.T) {
		// Setup test data
		filter := types.ContactFilter{
			OrganizationID: s.orgID,
			IsCustomer:     boolPtr(true),
		}

		// Expected query and result
		count := 5
		rows := sqlmock.NewRows([]string{"count"}).AddRow(count)

		s.mockDB.Mock.ExpectQuery("SELECT COUNT").
			WithArgs(s.orgID, true).
			WillReturnRows(rows)

		// Execute
		total, err := s.repo.Count(s.ctx, filter)

		// Assert
		require.NoError(t, err)
		require.Equal(t, count, total)

		// Ensure all expectations were met
		err = s.mockDB.Mock.ExpectationsWereMet()
		require.NoError(t, err, "there were unmet expectations")
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// Run the test suite
func TestContactRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ContactRepositoryTestSuite))
}

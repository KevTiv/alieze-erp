package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/queue"
)

func TestContactExportService_ExportContacts(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("CreateExportJob", ctx, mock.MatchedBy(func(j *types.ContactExportJob) bool {
		return j.FileFormat == "csv" && j.Status == "pending"
	})).Return(nil)
	mockQueue.On("Enqueue", ctx, mock.MatchedBy(func(j *queue.Job) bool {
		return j.Type == "contact.export"
	})).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	req := types.ExportContactsRequest{
		FileFormat:     "csv",
		FilterCriteria: types.JSONBMap{},
		SelectedFields: types.JSONBMap{"email": true, "phone": true},
	}

	result, err := svc.ExportContacts(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "csv", result.FileFormat)
	assert.Equal(t, "pending", result.Status)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestContactExportService_ExportContacts_InvalidFormat(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)

	// Test with invalid format
	req := types.ExportContactsRequest{
		FileFormat: "pdf", // Invalid format
	}

	result, err := svc.ExportContacts(ctx, orgID, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid file format")

	mockAuth.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "CreateExportJob")
	mockQueue.AssertNotCalled(t, "Enqueue")
}

func TestContactExportService_GenerateCSV(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	// Test data
	email1 := "john@example.com"
	phone1 := "+1234567890"
	email2 := "jane@example.com"
	phone2 := "+0987654321"

	contacts := []*types.Contact{
		{
			ID:        uuid.New(),
			Email:     &email1,
			Phone:     &phone1,
			FirstName: strPtr("John"),
			LastName:  strPtr("Doe"),
		},
		{
			ID:        uuid.New(),
			Email:     &email2,
			Phone:     &phone2,
			FirstName: strPtr("Jane"),
			LastName:  strPtr("Smith"),
		},
	}

	selectedFields := types.JSONBMap{} // Empty means all fields

	// Test
	csvData, err := svc.GenerateCSV(contacts, selectedFields)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, csvData)
	assert.Greater(t, len(csvData), 0)

	// Verify CSV contains expected data
	csvString := string(csvData)
	assert.Contains(t, csvString, "email")
	assert.Contains(t, csvString, "john@example.com")
	assert.Contains(t, csvString, "jane@example.com")
}

func TestContactExportService_GenerateXLSX(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	// Test data
	email := "john@example.com"
	contacts := []*types.Contact{
		{
			ID:        uuid.New(),
			Email:     &email,
			FirstName: strPtr("John"),
			LastName:  strPtr("Doe"),
		},
	}

	selectedFields := types.JSONBMap{}

	// Test
	xlsxData, err := svc.GenerateXLSX(contacts, selectedFields)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, xlsxData)
	assert.Greater(t, len(xlsxData), 0)

	// XLSX files have a specific header
	assert.True(t, len(xlsxData) > 4) // Minimum XLSX file size
}

func TestContactExportService_GetExportJob(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	jobID := uuid.New()
	fileURL := "https://storage.example.com/file.csv"
	job := &types.ContactExportJob{
		ID:             jobID,
		OrganizationID: orgID,
		FileFormat:     "csv",
		Status:         "completed",
		TotalRecords:   100,
		FileURL:        &fileURL,
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("GetExportJob", ctx, jobID).Return(job, nil)

	// Test
	result, err := svc.GetExportJob(ctx, orgID, jobID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, jobID, result.ID)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, 100, result.TotalRecords)
	assert.NotNil(t, result.FileURL)
	assert.Equal(t, fileURL, *result.FileURL)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactExportService_ListExportJobs(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	jobs := []*types.ContactExportJob{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			FileFormat:     "csv",
			Status:         "completed",
		},
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			FileFormat:     "xlsx",
			Status:         "pending",
		},
	}

	filter := types.ExportJobFilter{
		OrganizationID: orgID,
		Limit:          50,
		Offset:         0,
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("ListExportJobs", ctx, filter).Return(jobs, nil)
	mockRepo.On("CountExportJobs", ctx, filter).Return(2, nil)

	// Test
	result, err := svc.ListExportJobs(ctx, orgID, filter)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Jobs, 2)
	assert.Equal(t, 2, result.Total)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactExportService_GetExportJob_WrongOrganization(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	otherOrgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	jobID := uuid.New()
	job := &types.ContactExportJob{
		ID:             jobID,
		OrganizationID: otherOrgID, // Different organization
		FileFormat:     "csv",
		Status:         "completed",
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("GetExportJob", ctx, jobID).Return(job, nil)

	// Test
	result, err := svc.GetExportJob(ctx, orgID, jobID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "does not belong")

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactExportService_Unauthorized(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock unauthorized
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(false)

	// Test
	req := types.ExportContactsRequest{
		FileFormat: "csv",
	}

	result, err := svc.ExportContacts(ctx, orgID, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized")

	mockAuth.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "CreateExportJob")
}

func TestContactExportService_GenerateCSV_SelectedFields(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactExportService(
		mockRepo,
		mockContactRepo,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	// Test data
	email := "john@example.com"
	phone := "+1234567890"
	contacts := []*types.Contact{
		{
			ID:        uuid.New(),
			Email:     &email,
			Phone:     &phone,
			FirstName: strPtr("John"),
			LastName:  strPtr("Doe"),
			Company:   strPtr("Acme Inc"),
		},
	}

	// Only export email and phone
	selectedFields := types.JSONBMap{
		"email": true,
		"phone": true,
	}

	// Test
	csvData, err := svc.GenerateCSV(contacts, selectedFields)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, csvData)

	csvString := string(csvData)
	assert.Contains(t, csvString, "email")
	assert.Contains(t, csvString, "phone")
	assert.Contains(t, csvString, "john@example.com")
	assert.Contains(t, csvString, "+1234567890")
}

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

// MockContactImportExportRepository is a mock implementation
type MockContactImportExportRepository struct {
	mock.Mock
}

func (m *MockContactImportExportRepository) CreateImportJob(ctx context.Context, job *types.ContactImportJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockContactImportExportRepository) GetImportJob(ctx context.Context, id uuid.UUID) (*types.ContactImportJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactImportJob), args.Error(1)
}

func (m *MockContactImportExportRepository) UpdateImportJob(ctx context.Context, job *types.ContactImportJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockContactImportExportRepository) ListImportJobs(ctx context.Context, filter types.ImportJobFilter) ([]*types.ContactImportJob, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ContactImportJob), args.Error(1)
}

func (m *MockContactImportExportRepository) CountImportJobs(ctx context.Context, filter types.ImportJobFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockContactImportExportRepository) CreateExportJob(ctx context.Context, job *types.ContactExportJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockContactImportExportRepository) GetExportJob(ctx context.Context, id uuid.UUID) (*types.ContactExportJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactExportJob), args.Error(1)
}

func (m *MockContactImportExportRepository) UpdateExportJob(ctx context.Context, job *types.ContactExportJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockContactImportExportRepository) ListExportJobs(ctx context.Context, filter types.ExportJobFilter) ([]*types.ContactExportJob, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ContactExportJob), args.Error(1)
}

func (m *MockContactImportExportRepository) CountExportJobs(ctx context.Context, filter types.ExportJobFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

// MockQueueManager is a mock for queue operations
type MockQueueManager struct {
	mock.Mock
}

func (m *MockQueueManager) Enqueue(ctx context.Context, job *queue.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockQueueManager) Dequeue(ctx context.Context) (*queue.Job, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*queue.Job), args.Error(1)
}

func (m *MockQueueManager) Complete(ctx context.Context, jobID uuid.UUID) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *MockQueueManager) Fail(ctx context.Context, jobID uuid.UUID, errorMsg string) error {
	args := m.Called(ctx, jobID, errorMsg)
	return args.Error(0)
}

func TestContactImportService_ImportContacts(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockValidationService := &service.ContactValidationService{} // Can be nil for this test
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactImportService(
		mockRepo,
		mockContactRepo,
		mockValidationService,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("CreateImportJob", ctx, mock.MatchedBy(func(j *types.ContactImportJob) bool {
		return j.FileName == "contacts.csv" && j.FileFormat == "csv" && j.Status == "pending"
	})).Return(nil)
	mockQueue.On("Enqueue", ctx, mock.MatchedBy(func(j *queue.Job) bool {
		return j.Type == "contact.import"
	})).Return(nil)
	mockEvents.On("Publish", ctx, mock.Anything).Return()

	// Test
	req := types.ImportContactsRequest{
		FileName:          "contacts.csv",
		FileFormat:        "csv",
		FileSize:          1024,
		FileData:          "base64data",
		FieldMapping:      types.JSONBMap{"Email": "email"},
		DuplicateHandling: "skip",
	}

	result, err := svc.ImportContacts(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "contacts.csv", result.FileName)
	assert.Equal(t, "csv", result.FileFormat)
	assert.Equal(t, "pending", result.Status)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestContactImportService_ImportContacts_InvalidFormat(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockValidationService := &service.ContactValidationService{}
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactImportService(
		mockRepo,
		mockContactRepo,
		mockValidationService,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock expectations
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(true)

	// Test with invalid format
	req := types.ImportContactsRequest{
		FileName:   "contacts.txt",
		FileFormat: "txt", // Invalid format
	}

	result, err := svc.ImportContacts(ctx, orgID, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid file format")

	mockAuth.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "CreateImportJob")
	mockQueue.AssertNotCalled(t, "Enqueue")
}

func TestContactImportService_ParseCSV(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockValidationService := &service.ContactValidationService{}
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactImportService(
		mockRepo,
		mockContactRepo,
		mockValidationService,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	// Test CSV parsing
	csvData := `Email,FirstName,LastName
john@example.com,John,Doe
jane@example.com,Jane,Smith`

	records, err := svc.ParseCSV(csvData)

	// Assertions
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "john@example.com", records[0]["Email"])
	assert.Equal(t, "John", records[0]["FirstName"])
	assert.Equal(t, "Doe", records[0]["LastName"])
	assert.Equal(t, "jane@example.com", records[1]["Email"])
}

func TestContactImportService_GetImportMapping(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockValidationService := &service.ContactValidationService{}
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactImportService(
		mockRepo,
		mockContactRepo,
		mockValidationService,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)

	// Test
	req := types.GetImportMappingRequest{
		Headers: []string{"Email", "Phone Number", "First Name", "Company Name"},
	}

	result, err := svc.GetImportMapping(ctx, orgID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.SuggestedMapping)

	// Check that common mappings are suggested
	mapping := result.SuggestedMapping
	if emailMapping, ok := mapping["Email"]; ok {
		assert.Equal(t, "email", emailMapping)
	}
	if phoneMapping, ok := mapping["Phone Number"]; ok {
		assert.Equal(t, "phone", phoneMapping)
	}

	mockAuth.AssertExpectations(t)
}

func TestContactImportService_GetImportJob(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockValidationService := &service.ContactValidationService{}
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactImportService(
		mockRepo,
		mockContactRepo,
		mockValidationService,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	jobID := uuid.New()
	job := &types.ContactImportJob{
		ID:             jobID,
		OrganizationID: orgID,
		FileName:       "contacts.csv",
		Status:         "completed",
		TotalRows:      100,
		SuccessfulRows: 95,
		FailedRows:     5,
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("GetImportJob", ctx, jobID).Return(job, nil)

	// Test
	result, err := svc.GetImportJob(ctx, orgID, jobID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, jobID, result.ID)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, 100, result.TotalRows)
	assert.Equal(t, 95, result.SuccessfulRows)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactImportService_ListImportJobs(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockValidationService := &service.ContactValidationService{}
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactImportService(
		mockRepo,
		mockContactRepo,
		mockValidationService,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	jobs := []*types.ContactImportJob{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			FileName:       "import1.csv",
			Status:         "completed",
		},
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			FileName:       "import2.csv",
			Status:         "pending",
		},
	}

	filter := types.ImportJobFilter{
		OrganizationID: orgID,
		Limit:          50,
		Offset:         0,
	}

	// Mock expectations
	mockAuth.On("CanRead", ctx, userID, orgID, "contact").Return(true)
	mockRepo.On("ListImportJobs", ctx, filter).Return(jobs, nil)
	mockRepo.On("CountImportJobs", ctx, filter).Return(2, nil)

	// Test
	result, err := svc.ListImportJobs(ctx, orgID, filter)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Jobs, 2)
	assert.Equal(t, 2, result.Total)

	mockAuth.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestContactImportService_Unauthorized(t *testing.T) {
	// Setup
	mockRepo := new(MockContactImportExportRepository)
	mockContactRepo := new(MockContactRepositoryForValidation)
	mockValidationService := &service.ContactValidationService{}
	mockAuth := new(MockAuthorizationService)
	mockEvents := new(MockEventPublisher)
	mockQueue := new(MockQueueManager)

	svc := service.NewContactImportService(
		mockRepo,
		mockContactRepo,
		mockValidationService,
		mockAuth,
		mockEvents,
		mockQueue,
	)

	orgID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user_id", userID)

	// Mock unauthorized
	mockAuth.On("CanWrite", ctx, userID, orgID, "contact").Return(false)

	// Test
	req := types.ImportContactsRequest{
		FileName:   "contacts.csv",
		FileFormat: "csv",
	}

	result, err := svc.ImportContacts(ctx, orgID, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized")

	mockAuth.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "CreateImportJob")
}

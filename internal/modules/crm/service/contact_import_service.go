package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/auth"
	"github.com/KevTiv/alieze-erp/pkg/events"
	"github.com/KevTiv/alieze-erp/pkg/queue"
)

// ContactImportService handles contact import operations
type ContactImportService struct {
	repo              repository.ContactImportExportRepository
	contactRepo       types.ContactRepository
	validationService *ContactValidationService
	authService       auth.AuthorizationService
	eventPublisher    events.EventPublisher
	queueManager      queue.QueueManager
}

func NewContactImportService(
	repo repository.ContactImportExportRepository,
	contactRepo types.ContactRepository,
	validationService *ContactValidationService,
	authService auth.AuthorizationService,
	eventPublisher events.EventPublisher,
	queueManager queue.QueueManager,
) *ContactImportService {
	return &ContactImportService{
		repo:              repo,
		contactRepo:       contactRepo,
		validationService: validationService,
		authService:       authService,
		eventPublisher:    eventPublisher,
		queueManager:      queueManager,
	}
}

// ImportContacts queues an async import job
func (s *ContactImportService) ImportContacts(ctx context.Context, orgID uuid.UUID, req types.ContactImportRequest) (*types.ContactImportJob, error) {
	// Basic authorization check
	userID := ctx.Value("user_id").(uuid.UUID)

	// Validate file format
	if req.FileType != "csv" && req.FileType != "xlsx" {
		return nil, fmt.Errorf("invalid file format: must be 'csv' or 'xlsx'")
	}

	// Create import job
	job := &types.ContactImportJob{
		ID:             uuid.New(),
		OrganizationID: orgID,
		JobID:          uuid.New(),
		Filename:       req.Filename,
		FileType:       req.FileType,
		FieldMapping:   types.JSONBMap(req.FieldMapping),
		Options:        types.JSONBMap{},
		TotalRows:      0,
		ProcessedRows:  0,
		SuccessfulRows: 0,
		FailedRows:     0,
		Status:         "pending",
		CreatedBy:      &userID,
		CreatedAt:      time.Now(),
	}

	err := s.repo.CreateImportJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	// Queue the import job for background processing
	_, err = s.queueManager.EnqueueJob(ctx, "crm", "contact.import", map[string]interface{}{
		"job_id": job.ID.String(),
	}, &queue.ManagerJobOptions{
		Priority: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to queue import job: %w", err)
	}

	// Publish event
	s.eventPublisher.Publish(ctx, "contact.import.started", map[string]interface{}{
		"organization_id": orgID.String(),
		"job_id":          job.ID.String(),
		"file_name":       req.Filename,
		"file_format":     req.FileType,
	})

	return job, nil
}

// ProcessImportJob processes an import job (called by queue worker)
func (s *ContactImportService) ProcessImportJob(ctx context.Context, jobID uuid.UUID, fileData string) error {
	// Get import job
	job, err := s.repo.GetImportJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get import job: %w", err)
	}

	// Update status to processing
	now := time.Now()
	job.Status = "processing"
	job.StartedAt = &now
	err = s.repo.UpdateImportJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update import job status: %w", err)
	}

	// Publish progress event
	s.eventPublisher.Publish(ctx, events.Event{
		Type: "contact.import.progress",
		Data: map[string]interface{}{
			"organization_id": job.OrganizationID.String(),
			"job_id":          job.ID.String(),
			"status":          "processing",
		},
	})

	// Parse file based on format
	var records []map[string]string
	var parseErr error

	if job.FileType == "csv" {
		records, parseErr = s.ParseCSV(fileData)
	} else if job.FileType == "xlsx" {
		records, parseErr = s.ParseXLSX(fileData)
	}

	if parseErr != nil {
		job.Status = "failed"
		errMsg := fmt.Sprintf("Failed to parse file: %v", parseErr)
		job.ErrorMessage = &errMsg
		completed := time.Now()
		job.CompletedAt = &completed
		s.repo.UpdateImportJob(ctx, job)

		s.eventPublisher.Publish(ctx, events.Event{
			Type: "contact.import.failed",
			Data: map[string]interface{}{
				"organization_id": job.OrganizationID.String(),
				"job_id":          job.ID.String(),
				"error":           errMsg,
			},
		})

		return fmt.Errorf("failed to parse file: %w", parseErr)
	}

	job.TotalRows = len(records)

	// Process each record
	for i, record := range records {
		contact, err := s.mapRecordToContact(record, job.FieldMapping, job.OrganizationID)
		if err != nil {
			job.FailedRows++
			continue
		}

		// Handle duplicates
		shouldCreate, err := s.handleDuplicate(ctx, contact, job.Options)
		if err != nil || !shouldCreate {
			job.FailedRows++
			job.ProcessedRows++
			continue
		}

		// Create contact
		err = s.contactRepo.CreateContact(ctx, contact)
		if err != nil {
			job.FailedRows++
		} else {
			job.SuccessfulRows++
		}

		job.ProcessedRows++

		// Periodic progress update (every 100 rows)
		if (i+1)%100 == 0 {
			s.repo.UpdateImportJob(ctx, job)
			s.eventPublisher.Publish(ctx, events.Event{
				Type: "contact.import.progress",
				Data: map[string]interface{}{
					"organization_id": job.OrganizationID.String(),
					"job_id":          job.ID.String(),
					"processed":       job.ProcessedRows,
					"total":           job.TotalRows,
				},
			})
		}
	}

	// Mark as completed
	job.Status = "completed"
	completed := time.Now()
	job.CompletedAt = &completed
	err = s.repo.UpdateImportJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update import job: %w", err)
	}

	// Publish completion event
	s.eventPublisher.Publish(ctx, events.Event{
		Type: "contact.import.completed",
		Data: map[string]interface{}{
			"organization_id": job.OrganizationID.String(),
			"job_id":          job.ID.String(),
			"total_rows":      job.TotalRows,
			"successful":      job.SuccessfulRows,
			"failed":          job.FailedRows,
		},
	})

	return nil
}

// ParseCSV parses CSV file data
func (s *ContactImportService) ParseCSV(fileData string) ([]map[string]string, error) {
	reader := csv.NewReader(strings.NewReader(fileData))

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	var records []map[string]string

	// Read data rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		record := make(map[string]string)
		for i, value := range row {
			if i < len(headers) {
				record[headers[i]] = value
			}
		}
		records = append(records, record)
	}

	return records, nil
}

// ParseXLSX parses Excel file data
func (s *ContactImportService) ParseXLSX(fileData string) ([]map[string]string, error) {
	// In production, fileData would be base64-encoded or a file path
	// For now, we'll assume it's a file path
	f, err := excelize.OpenFile(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Excel rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("Excel file is empty")
	}

	// First row is headers
	headers := rows[0]
	var records []map[string]string

	// Process data rows
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		record := make(map[string]string)

		for j, value := range row {
			if j < len(headers) {
				record[headers[j]] = value
			}
		}
		records = append(records, record)
	}

	return records, nil
}

// GetImportMapping suggests field mappings based on headers
func (s *ContactImportService) GetImportMapping(ctx context.Context, orgID uuid.UUID, req types.GetImportMappingRequest) (*types.GetImportMappingResponse, error) {
	// Authorization check
	userID := ctx.Value("user_id").(uuid.UUID)
	if !s.authService.CanRead(ctx, userID, orgID, "contact") {
		return nil, fmt.Errorf("unauthorized: user does not have read permission for contacts")
	}

	mapping := make(types.JSONBMap)

	// Common field mappings based on header similarity
	fieldMappings := map[string][]string{
		"email":      {"email", "e-mail", "email_address", "mail"},
		"phone":      {"phone", "telephone", "phone_number", "mobile"},
		"first_name": {"first_name", "firstname", "fname", "given_name"},
		"last_name":  {"last_name", "lastname", "lname", "surname", "family_name"},
		"company":    {"company", "organization", "company_name", "org"},
		"title":      {"title", "job_title", "position", "role"},
		"city":       {"city", "town"},
		"state":      {"state", "province", "region"},
		"country":    {"country", "nation"},
	}

	for _, header := range req.Headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))

		// Try to find matching field
		for field, variants := range fieldMappings {
			for _, variant := range variants {
				if headerLower == variant || strings.Contains(headerLower, variant) {
					mapping[header] = field
					break
				}
			}
		}
	}

	return &types.GetImportMappingResponse{
		SuggestedMapping: mapping,
	}, nil
}

// GetImportJob retrieves an import job
func (s *ContactImportService) GetImportJob(ctx context.Context, orgID uuid.UUID, jobID uuid.UUID) (*types.ContactImportJob, error) {
	// Authorization check
	userID := ctx.Value("user_id").(uuid.UUID)
	if !s.authService.CanRead(ctx, userID, orgID, "contact") {
		return nil, fmt.Errorf("unauthorized: user does not have read permission for contacts")
	}

	job, err := s.repo.GetImportJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import job: %w", err)
	}

	// Verify ownership
	if job.OrganizationID != orgID {
		return nil, fmt.Errorf("import job does not belong to the specified organization")
	}

	return job, nil
}

// ListImportJobs lists import jobs
func (s *ContactImportService) ListImportJobs(ctx context.Context, orgID uuid.UUID, filter types.ImportJobFilter) (*types.ListImportJobsResponse, error) {
	// Authorization check
	userID := ctx.Value("user_id").(uuid.UUID)
	if !s.authService.CanRead(ctx, userID, orgID, "contact") {
		return nil, fmt.Errorf("unauthorized: user does not have read permission for contacts")
	}

	filter.OrganizationID = orgID

	jobs, err := s.repo.ListImportJobs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list import jobs: %w", err)
	}

	total, err := s.repo.CountImportJobs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count import jobs: %w", err)
	}

	return &types.ListImportJobsResponse{
		Jobs:   jobs,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

// Helper functions

func (s *ContactImportService) mapRecordToContact(record map[string]string, mapping types.JSONBMap, orgID uuid.UUID) (*types.Contact, error) {
	contact := &types.Contact{
		ID:             uuid.New(),
		OrganizationID: orgID,
	}

	// Map fields based on provided mapping
	for sourceField, targetField := range mapping {
		value, ok := record[sourceField]
		if !ok || value == "" {
			continue
		}

		switch targetField {
		case "email":
			contact.Email = &value
		case "phone":
			contact.Phone = &value
		case "first_name":
			contact.FirstName = &value
		case "last_name":
			contact.LastName = &value
		case "company":
			contact.Company = &value
		case "title":
			contact.Title = &value
		case "address":
			contact.Address = &value
		case "city":
			contact.City = &value
		case "state":
			contact.State = &value
		case "country":
			contact.Country = &value
		case "postal_code":
			contact.PostalCode = &value
		case "website":
			contact.Website = &value
		case "description":
			contact.Description = &value
		}
	}

	return contact, nil
}

func (s *ContactImportService) handleDuplicate(ctx context.Context, contact *types.Contact, handling string) (bool, error) {
	// Check for existing contact by email
	if contact.Email != nil && *contact.Email != "" {
		existing, err := s.findContactByEmail(ctx, contact.OrganizationID, *contact.Email)
		if err == nil && existing != nil {
			// Duplicate found
			switch handling {
			case "skip":
				return false, nil
			case "update":
				// Update existing contact
				contact.ID = existing.ID
				return true, s.contactRepo.UpdateContact(ctx, contact)
			case "create_new":
				// Create as new contact (allow duplicate)
				return true, nil
			default:
				return false, nil
			}
		}
	}

	// No duplicate found, create new
	return true, nil
}

func (s *ContactImportService) findContactByEmail(ctx context.Context, orgID uuid.UUID, email string) (*types.Contact, error) {
	// Simple search by email
	// In production, this would use a more efficient query
	filter := types.ContactFilter{
		OrganizationID: orgID,
		Email:          &email,
	}

	contacts, err := s.contactRepo.ListContacts(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(contacts) > 0 {
		return contacts[0], nil
	}

	return nil, fmt.Errorf("not found")
}

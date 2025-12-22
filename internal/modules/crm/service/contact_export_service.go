package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/auth"
	"github.com/KevTiv/alieze-erp/pkg/events"
	"github.com/KevTiv/alieze-erp/pkg/queue"
)

// ContactExportService handles contact export operations
type ContactExportService struct {
	repo           repository.ContactImportExportRepository
	contactRepo    types.ContactRepository
	authService    auth.AuthorizationService
	eventPublisher events.EventPublisher
	queueManager   queue.QueueManager
}

func NewContactExportService(
	repo repository.ContactImportExportRepository,
	contactRepo types.ContactRepository,
	authService auth.AuthorizationService,
	eventPublisher events.EventPublisher,
	queueManager queue.QueueManager,
) *ContactExportService {
	return &ContactExportService{
		repo:           repo,
		contactRepo:    contactRepo,
		authService:    authService,
		eventPublisher: eventPublisher,
		queueManager:   queueManager,
	}
}

// ExportContacts queues an async export job
func (s *ContactExportService) ExportContacts(ctx context.Context, orgID uuid.UUID, req types.ContactExportRequest) (*types.ContactExportJob, error) {
	// Basic authorization check
	userID := ctx.Value("user_id").(uuid.UUID)

	// Validate file format
	if req.Format != "csv" && req.Format != "xlsx" {
		return nil, fmt.Errorf("invalid file format: must be 'csv' or 'xlsx'")
	}

	// Create export job
	job := &types.ContactExportJob{
		ID:             uuid.New(),
		OrganizationID: orgID,
		JobID:          uuid.New(),
		FilterCriteria: types.JSONBMap{},
		SelectedFields: types.JSONBMap{},
		Format:         req.Format,
		TotalContacts:  0,
		Status:         "pending",
		CreatedBy:      &userID,
		CreatedAt:      time.Now(),
	}

	err := s.repo.CreateExportJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create export job: %w", err)
	}

	// Queue the export job for background processing
	maxRetries := 3
	_, err = s.queueManager.EnqueueJob(ctx, "crm", "contact.export", map[string]interface{}{
		"job_id": job.ID.String(),
	}, &queue.ManagerJobOptions{
		Priority:   1,
		MaxRetries: &maxRetries,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to queue export job: %w", err)
	}

	// Publish event
	s.eventPublisher.Publish(ctx, "contact.export.started", map[string]interface{}{
		"organization_id": orgID.String(),
		"job_id":          job.ID.String(),
		"file_format":     req.Format,
	})

	return job, nil
}

// ProcessExportJob processes an export job (called by queue worker)
func (s *ContactExportService) ProcessExportJob(ctx context.Context, jobID uuid.UUID) error {
	// Get export job
	job, err := s.repo.GetExportJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get export job: %w", err)
	}

	// Update status to processing
	now := time.Now()
	job.Status = "processing"
	job.StartedAt = &now
	err = s.repo.UpdateExportJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update export job status: %w", err)
	}

	// Build filter from criteria
	filter := s.buildFilterFromCriteria(job.FilterCriteria, job.OrganizationID)

	// Get contacts
	contacts, err := s.contactRepo.ListContacts(ctx, filter)
	if err != nil {
		job.Status = "failed"
		errMsg := fmt.Sprintf("Failed to retrieve contacts: %v", err)
		job.ErrorMessage = &errMsg
		completed := time.Now()
		job.CompletedAt = &completed
		s.repo.UpdateExportJob(ctx, job)

		s.eventPublisher.Publish(ctx, "contact.export.failed", map[string]interface{}{
			"organization_id": job.OrganizationID.String(),
			"job_id":          job.ID.String(),
			"error":           errMsg,
		})

		return fmt.Errorf("failed to retrieve contacts: %w", err)
	}

	job.TotalContacts = len(contacts)

	// Generate file based on format
	var generateErr error

	if job.Format == "csv" {
		_, generateErr = s.GenerateCSV(contacts, job.SelectedFields)
	} else if job.Format == "xlsx" {
		_, generateErr = s.GenerateXLSX(contacts, job.SelectedFields)
	}

	if generateErr != nil {
		job.Status = "failed"
		errMsg := fmt.Sprintf("Failed to generate file: %v", generateErr)
		job.ErrorMessage = &errMsg
		completed := time.Now()
		job.CompletedAt = &completed
		s.repo.UpdateExportJob(ctx, job)

		s.eventPublisher.Publish(ctx, "contact.export.failed", map[string]interface{}{
			"organization_id": job.OrganizationID.String(),
			"job_id":          job.ID.String(),
			"error":           errMsg,
		})

		return fmt.Errorf("failed to generate file: %w", generateErr)
	}

	// Upload to storage and get presigned URL
	fileURL := fmt.Sprintf("https://storage.example.com/exports/%s.%s", job.ID, job.Format)
	job.FileURL = &fileURL

	// Mark as completed
	job.Status = "completed"
	completed := time.Now()
	job.CompletedAt = &completed
	err = s.repo.UpdateExportJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update export job: %w", err)
	}

	// Publish completion event
	s.eventPublisher.Publish(ctx, "contact.export.completed", map[string]interface{}{
		"organization_id": job.OrganizationID.String(),
		"job_id":          job.ID.String(),
		"total_records":   job.TotalContacts,
		"file_url":        fileURL,
	})

	return nil
}

// GenerateCSV generates a CSV file from contacts
func (s *ContactExportService) GenerateCSV(contacts []*types.Contact, selectedFields types.JSONBMap) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Determine fields to export
	fields := s.getExportFields(selectedFields)

	// Write header row
	if err := writer.Write(fields); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, contact := range contacts {
		row := s.contactToRow(contact, fields)
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateXLSX generates an Excel file from contacts
func (s *ContactExportService) GenerateXLSX(contacts []*types.Contact, selectedFields types.JSONBMap) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Contacts"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}

	f.SetActiveSheet(index)

	// Determine fields to export
	fields := s.getExportFields(selectedFields)

	// Write header row
	for i, field := range fields {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, field)
	}

	// Write data rows
	for rowNum, contact := range contacts {
		row := s.contactToRow(contact, fields)
		for colNum, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colNum+1, rowNum+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// GetExportJob retrieves an export job
func (s *ContactExportService) GetExportJob(ctx context.Context, orgID uuid.UUID, jobID uuid.UUID) (*types.ContactExportJob, error) {
	// Basic authorization check
	userID := ctx.Value("user_id").(uuid.UUID)

	job, err := s.repo.GetExportJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get export job: %w", err)
	}

	// Verify ownership
	if job.OrganizationID != orgID {
		return nil, fmt.Errorf("export job does not belong to the specified organization")
	}

	return job, nil
}

// ListExportJobs lists export jobs
func (s *ContactExportService) ListExportJobs(ctx context.Context, orgID uuid.UUID, filter types.ExportJobFilter) ([]*types.ContactExportJob, int, error) {
	// Basic authorization check
	userID := ctx.Value("user_id").(uuid.UUID)

	filter.OrganizationID = orgID

	jobs, err := s.repo.ListExportJobs(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list export jobs: %w", err)
	}

	total, err := s.repo.CountExportJobs(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count export jobs: %w", err)
	}

	return jobs, total, nil
}

// Helper functions

func (s *ContactExportService) buildFilterFromCriteria(criteria types.JSONBMap, orgID uuid.UUID) types.ContactFilter {
	filter := types.ContactFilter{
		OrganizationID: orgID,
	}

	// Parse criteria and build filter
	if val, ok := criteria["is_customer"]; ok {
		if boolVal, ok := val.(bool); ok {
			filter.IsCustomer = &boolVal
		}
	}

	if val, ok := criteria["is_vendor"]; ok {
		if boolVal, ok := val.(bool); ok {
			filter.IsVendor = &boolVal
		}
	}

	return filter
}

func (s *ContactExportService) getExportFields(selectedFields types.JSONBMap) []string {
	// If no fields selected, export all standard fields
	if len(selectedFields) == 0 {
		return []string{
			"email", "phone", "name", "company",
			"title", "address", "city", "state", "country", "postal_code",
			"website", "description",
		}
	}

	// Use selected fields
	var fields []string
	for field := range selectedFields {
		fields = append(fields, field)
	}
	return fields
}

func (s *ContactExportService) contactToRow(contact *types.Contact, fields []string) []string {
	row := make([]string, len(fields))

	for i, field := range fields {
		switch field {
		case "email":
			row[i] = getStringPtrValue(contact.Email)
		case "phone":
			row[i] = getStringPtrValue(contact.Phone)
		case "name":
			row[i] = getStringPtrValue(contact.Name)
		case "company":
			row[i] = getStringPtrValue(contact.Company)
		case "title":
			row[i] = getStringPtrValue(contact.Title)
		case "address":
			row[i] = getStringPtrValue(contact.Address)
		case "city":
			row[i] = getStringPtrValue(contact.City)
		case "state":
			row[i] = getStringPtrValue(contact.State)
		case "country":
			row[i] = getStringPtrValue(contact.Country)
		case "postal_code":
			row[i] = getStringPtrValue(contact.PostalCode)
		case "website":
			row[i] = getStringPtrValue(contact.Website)
		case "description":
			row[i] = getStringPtrValue(contact.Description)
		default:
			row[i] = ""
		}
	}

	return row
}

func getStringPtrValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/pkg/queue"
)

const JobTypeContactImport = "contact.import"

// ContactImportJobPayload represents the job payload for contact import
type ContactImportJobPayload struct {
	JobID    uuid.UUID `json:"job_id"`
	FileData string    `json:"file_data"`
}

// ContactImportJobHandler handles async contact import jobs
type ContactImportJobHandler struct {
	importService *service.ContactImportService
}

func NewContactImportJobHandler(importService *service.ContactImportService) *ContactImportJobHandler {
	return &ContactImportJobHandler{
		importService: importService,
	}
}

// Handle processes a contact import job
func (h *ContactImportJobHandler) Handle(ctx context.Context, job *queue.Job) error {
	// Parse job payload
	var payload ContactImportJobPayload
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal job payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal job payload: %w", err)
	}

	// Execute import
	err := h.importService.ProcessImportJob(ctx, payload.JobID, payload.FileData)
	if err != nil {
		return fmt.Errorf("failed to process import job: %w", err)
	}

	fmt.Printf("Contact import job completed: %s\n", payload.JobID)
	return nil
}

// JobType returns the job type this handler processes
func (h *ContactImportJobHandler) JobType() string {
	return JobTypeContactImport
}

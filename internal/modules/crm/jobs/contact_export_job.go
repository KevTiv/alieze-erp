package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/pkg/queue"
)

const JobTypeContactExport = "contact.export"

// ContactExportJobPayload represents the job payload for contact export
type ContactExportJobPayload struct {
	JobID uuid.UUID `json:"job_id"`
}

// ContactExportJobHandler handles async contact export jobs
type ContactExportJobHandler struct {
	exportService *service.ContactExportService
}

func NewContactExportJobHandler(exportService *service.ContactExportService) *ContactExportJobHandler {
	return &ContactExportJobHandler{
		exportService: exportService,
	}
}

// Handle processes a contact export job
func (h *ContactExportJobHandler) Handle(ctx context.Context, job *queue.Job) error {
	// Parse job payload
	var payload ContactExportJobPayload
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal job payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal job payload: %w", err)
	}

	// Execute export
	err := h.exportService.ProcessExportJob(ctx, payload.JobID)
	if err != nil {
		return fmt.Errorf("failed to process export job: %w", err)
	}

	fmt.Printf("Contact export job completed: %s\n", payload.JobID)
	return nil
}

// JobType returns the job type this handler processes
func (h *ContactExportJobHandler) JobType() string {
	return JobTypeContactExport
}

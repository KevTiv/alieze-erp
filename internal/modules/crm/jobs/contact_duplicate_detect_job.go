package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
	"github.com/KevTiv/alieze-erp/pkg/queue"
)

const JobTypeDuplicateDetection = "contact.duplicate.detection"

// DuplicateDetectionJobPayload represents the job payload for duplicate detection
type DuplicateDetectionJobPayload struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	Threshold      int       `json:"threshold"`
	Limit          int       `json:"limit"`
}

// ContactDuplicateDetectJobHandler handles async duplicate detection jobs
type ContactDuplicateDetectJobHandler struct {
	mergeService *service.ContactMergeService
}

func NewContactDuplicateDetectJobHandler(mergeService *service.ContactMergeService) *ContactDuplicateDetectJobHandler {
	return &ContactDuplicateDetectJobHandler{
		mergeService: mergeService,
	}
}

// Handle processes a duplicate detection job
func (h *ContactDuplicateDetectJobHandler) Handle(ctx context.Context, job *queue.Job) error {
	// Parse job payload
	var payload DuplicateDetectionJobPayload
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal job payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal job payload: %w", err)
	}

	// Create request for duplicate detection
	req := types.DetectDuplicatesRequest{
		Threshold: &payload.Threshold,
		Limit:     &payload.Limit,
	}

	// Execute duplicate detection
	response, err := h.mergeService.DetectDuplicates(ctx, payload.OrganizationID, req)
	if err != nil {
		return fmt.Errorf("failed to detect duplicates: %w", err)
	}

	// Log results
	fmt.Printf("Duplicate detection completed for organization %s: found %d duplicates, created %d new records\n",
		payload.OrganizationID, response.TotalFound, response.TotalCreated)

	return nil
}

// JobType returns the job type this handler processes
func (h *ContactDuplicateDetectJobHandler) JobType() string {
	return JobTypeDuplicateDetection
}

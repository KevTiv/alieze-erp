package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/common/service"
	"github.com/KevTiv/alieze-erp/internal/modules/common/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type AttachmentHandler struct {
	service *service.AttachmentService
}

func NewAttachmentHandler(service *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{
		service: service,
	}
}

func (h *AttachmentHandler) RegisterRoutes(router *httprouter.Router) {
	// Attachment management endpoints
	router.POST("/api/v1/attachments", h.Upload)
	router.GET("/api/v1/attachments/:id", h.GetAttachment)
	router.GET("/api/v1/attachments/:id/download", h.Download)
	router.GET("/api/v1/attachments/:id/url", h.GetPublicURL)
	router.PUT("/api/v1/attachments/:id", h.Update)
	router.DELETE("/api/v1/attachments/:id", h.Delete)

	// Resource-based endpoints
	router.GET("/api/v1/attachments/resource/:model/:id", h.ListByResource)
	router.GET("/api/v1/attachments", h.List)

	// Analytics and management
	router.GET("/api/v1/attachments/stats/:organization_id", h.GetStats)
	router.GET("/api/v1/attachments/duplicates/:organization_id", h.FindDuplicates)
	router.POST("/api/v1/attachments/:id/regenerate-token", h.RegenerateToken)

	// Public access (no auth required)
	router.GET("/api/v1/attachments/public/:token", h.DownloadPublic)
}

// Upload handles file upload
func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse multipart form (max 100MB)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		respondError(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		respondError(w, "Failed to read file data", http.StatusInternalServerError)
		return
	}

	// Parse metadata
	name := r.FormValue("name")
	if name == "" {
		name = header.Filename
	}

	description := r.FormValue("description")
	resModel := r.FormValue("res_model")
	resIDStr := r.FormValue("res_id")
	accessTypeStr := r.FormValue("access_type")

	// Validate required fields
	if resModel == "" || resIDStr == "" {
		respondError(w, "res_model and res_id are required", http.StatusBadRequest)
		return
	}

	resID, err := uuid.Parse(resIDStr)
	if err != nil {
		respondError(w, "Invalid res_id", http.StatusBadRequest)
		return
	}

	// Set default access type
	accessType := types.AttachmentAccessPrivate
	if accessTypeStr != "" {
		accessType = types.AttachmentAccessType(accessTypeStr)
	}

	// Parse optional metadata JSON
	var metadata map[string]interface{}
	if metadataStr := r.FormValue("metadata"); metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			respondError(w, "Invalid metadata JSON", http.StatusBadRequest)
			return
		}
	}

	// Get uploaded_by from context (would come from auth middleware)
	// For now, using a placeholder
	uploadedByStr := r.Header.Get("X-User-ID")
	uploadedBy := uuid.Nil
	if uploadedByStr != "" {
		uploadedBy, _ = uuid.Parse(uploadedByStr)
	}

	// Create upload request
	uploadReq := types.AttachmentUploadRequest{
		Name:        name,
		Description: description,
		ResModel:    resModel,
		ResID:       resID,
		AccessType:  accessType,
		Metadata:    metadata,
		FileData:    fileData,
		MimeType:    header.Header.Get("Content-Type"),
		FileSize:    int64(len(fileData)),
	}

	// Upload
	attachment, err := h.service.Upload(r.Context(), uploadReq, uploadedBy)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, attachment, http.StatusCreated)
}

// GetAttachment retrieves attachment metadata
func (h *AttachmentHandler) GetAttachment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	// This would use the repository directly or add a Get method to service
	// For now, we'll download and return metadata only
	response, err := h.service.Download(r.Context(), id, nil)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, response.Attachment, http.StatusOK)
}

// Download downloads a file
func (h *AttachmentHandler) Download(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	// Get accessed_by from context (auth middleware)
	accessedByStr := r.Header.Get("X-User-ID")
	var accessedBy *uuid.UUID
	if accessedByStr != "" {
		uid, _ := uuid.Parse(accessedByStr)
		accessedBy = &uid
	}

	response, err := h.service.Download(r.Context(), id, accessedBy)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", response.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", response.Filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(response.FileData)))

	// Write file data
	w.WriteHeader(http.StatusOK)
	w.Write(response.FileData)
}

// GetPublicURL generates a time-limited public URL
func (h *AttachmentHandler) GetPublicURL(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	// Parse expiry duration (default 1 hour)
	expiryStr := r.URL.Query().Get("expiry")
	expiry := 1 * time.Hour
	if expiryStr != "" {
		if d, err := time.ParseDuration(expiryStr); err == nil {
			expiry = d
		}
	}

	url, err := h.service.GetPublicURL(r.Context(), id, expiry)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"url":        url,
		"expires_in": expiry.String(),
	}, http.StatusOK)
}

// ListByResource lists attachments for a specific resource
func (h *AttachmentHandler) ListByResource(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	model := ps.ByName("model")
	idStr := ps.ByName("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	attachments, err := h.service.ListByResource(r.Context(), model, id)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, attachments, http.StatusOK)
}

// List lists attachments with filters
func (h *AttachmentHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	filters := types.AttachmentFilter{
		Limit:  50,
		Offset: 0,
	}

	// Parse query parameters
	if orgIDStr := r.URL.Query().Get("organization_id"); orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err == nil {
			filters.OrganizationID = &orgID
		}
	}

	if model := r.URL.Query().Get("res_model"); model != "" {
		filters.ResModel = &model
	}

	if resIDStr := r.URL.Query().Get("res_id"); resIDStr != "" {
		resID, err := uuid.Parse(resIDStr)
		if err == nil {
			filters.ResID = &resID
		}
	}

	if attType := r.URL.Query().Get("type"); attType != "" {
		t := types.AttachmentType(attType)
		filters.AttachmentType = &t
	}

	if accessType := r.URL.Query().Get("access_type"); accessType != "" {
		at := types.AttachmentAccessType(accessType)
		filters.AccessType = &at
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	if includeDeleted := r.URL.Query().Get("include_deleted"); includeDeleted == "true" {
		filters.IncludeDeleted = true
	}

	attachments, err := h.service.List(r.Context(), filters)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, attachments, http.StatusOK)
}

// UpdateRequest represents the request body for updating attachment metadata
type UpdateRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	AccessType  types.AttachmentAccessType `json:"access_type,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Update updates attachment metadata
func (h *AttachmentHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing attachment (would need to add Get method to service)
	// For now, creating a partial update
	attachment := types.Attachment{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		AccessType:  req.AccessType,
		Metadata:    req.Metadata,
	}

	updated, err := h.service.Update(r.Context(), attachment)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, updated, http.StatusOK)
}

// Delete soft deletes an attachment
func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	// Get deleted_by from context
	deletedByStr := r.Header.Get("X-User-ID")
	deletedBy := uuid.Nil
	if deletedByStr != "" {
		deletedBy, _ = uuid.Parse(deletedByStr)
	}

	// Check if hard delete requested
	hardDelete := r.URL.Query().Get("hard") == "true"

	if hardDelete {
		err = h.service.HardDelete(r.Context(), id)
	} else {
		err = h.service.Delete(r.Context(), id, deletedBy)
	}

	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"message": "Attachment deleted successfully",
	}, http.StatusOK)
}

// GetStats retrieves attachment statistics
func (h *AttachmentHandler) GetStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	organizationID, err := uuid.Parse(ps.ByName("organization_id"))
	if err != nil {
		respondError(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	stats, err := h.service.GetStats(r.Context(), organizationID)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, stats, http.StatusOK)
}

// FindDuplicates finds duplicate files
func (h *AttachmentHandler) FindDuplicates(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	organizationID, err := uuid.Parse(ps.ByName("organization_id"))
	if err != nil {
		respondError(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	// Parse min_count parameter (default 2)
	minCount := 2
	if minCountStr := r.URL.Query().Get("min_count"); minCountStr != "" {
		if mc, err := strconv.Atoi(minCountStr); err == nil {
			minCount = mc
		}
	}

	duplicates, err := h.service.FindDuplicates(r.Context(), organizationID, minCount)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, duplicates, http.StatusOK)
}

// RegenerateToken regenerates the access token for an attachment
func (h *AttachmentHandler) RegenerateToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	attachment, err := h.service.RegenerateAccessToken(r.Context(), id)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, attachment, http.StatusOK)
}

// DownloadPublic downloads a file using public access token
func (h *AttachmentHandler) DownloadPublic(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token := ps.ByName("token")
	if token == "" {
		respondError(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// This would require adding a FindByToken method to the service/repository
	// For now, returning not implemented
	respondError(w, "Public download not yet implemented", http.StatusNotImplemented)
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, statusCode int) {
	respondJSON(w, map[string]string{"error": message}, statusCode)
}

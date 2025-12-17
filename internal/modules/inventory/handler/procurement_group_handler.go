package handler

import (
	"encoding/json"
	"net/http"

	"alieze-erp/internal/modules/auth/middleware"
	"alieze-erp/internal/modules/inventory/service"
	"alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// ProcurementGroupHandler handles HTTP requests for procurement groups
type ProcurementGroupHandler struct {
	service *service.ProcurementGroupService
}

// NewProcurementGroupHandler creates a new ProcurementGroupHandler
type NewProcurementGroupHandler(service *service.ProcurementGroupService) *ProcurementGroupHandler {
	return &ProcurementGroupHandler{
		service: service,
	}
}

// RegisterRoutes registers procurement group routes
func (h *ProcurementGroupHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/inventory/procurement-groups", h.Create)
	router.GET("/api/inventory/procurement-groups/:id", h.GetByID)
	router.GET("/api/inventory/procurement-groups", h.List)
	router.PUT("/api/inventory/procurement-groups/:id", h.Update)
	router.DELETE("/api/inventory/procurement-groups/:id", h.Delete)
}

// Create handles procurement group creation
func (h *ProcurementGroupHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.ProcurementGroupCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	group, err := h.service.Create(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// GetByID handles retrieving a procurement group by ID
func (h *ProcurementGroupHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	group, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		http.Error(w, "Procurement group not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// List handles listing all procurement groups
func (h *ProcurementGroupHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	orgID, err := auth.GetOrganizationIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groups, err := h.service.List(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// Update handles updating a procurement group
func (h *ProcurementGroupHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req types.ProcurementGroupUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// Delete handles deleting a procurement group
func (h *ProcurementGroupHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

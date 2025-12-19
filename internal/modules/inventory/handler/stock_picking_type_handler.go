package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KevTiv/alieze-erp/internal/modules/auth/middleware"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/service"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// StockPickingTypeHandler handles HTTP requests for stock picking types
type StockPickingTypeHandler struct {
	service *service.StockPickingTypeService
}

// NewStockPickingTypeHandler creates a new StockPickingTypeHandler
type NewStockPickingTypeHandler(service *service.StockPickingTypeService) *StockPickingTypeHandler {
	return &StockPickingTypeHandler{
		service: service,
	}
}

// RegisterRoutes registers stock picking type routes
func (h *StockPickingTypeHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/inventory/stock-picking-types", h.Create)
	router.GET("/api/inventory/stock-picking-types/:id", h.GetByID)
	router.GET("/api/inventory/stock-picking-types", h.List)
	router.PUT("/api/inventory/stock-picking-types/:id", h.Update)
	router.DELETE("/api/inventory/stock-picking-types/:id", h.Delete)
}

// Create handles stock picking type creation
func (h *StockPickingTypeHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.StockPickingTypeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	pickingType, err := h.service.Create(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pickingType)
}

// GetByID handles retrieving a stock picking type by ID
func (h *StockPickingTypeHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	pickingType, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pickingType == nil {
		http.Error(w, "Stock picking type not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pickingType)
}

// List handles listing all stock picking types
func (h *StockPickingTypeHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	orgID, err := auth.GetOrganizationIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pickingTypes, err := h.service.List(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pickingTypes)
}

// Update handles updating a stock picking type
func (h *StockPickingTypeHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req types.StockPickingTypeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pickingType, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pickingType)
}

// Delete handles deleting a stock picking type
func (h *StockPickingTypeHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

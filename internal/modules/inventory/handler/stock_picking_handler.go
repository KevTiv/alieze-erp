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

// StockPickingHandler handles HTTP requests for stock pickings
type StockPickingHandler struct {
	service *service.StockPickingService
}

// NewStockPickingHandler creates a new StockPickingHandler
type NewStockPickingHandler(service *service.StockPickingService) *StockPickingHandler {
	return &StockPickingHandler{
		service: service,
	}
}

// RegisterRoutes registers stock picking routes
func (h *StockPickingHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/inventory/stock-pickings", h.Create)
	router.GET("/api/inventory/stock-pickings/:id", h.GetByID)
	router.GET("/api/inventory/stock-pickings", h.List)
	router.PUT("/api/inventory/stock-pickings/:id", h.Update)
	router.DELETE("/api/inventory/stock-pickings/:id", h.Delete)
}

// Create handles stock picking creation
func (h *StockPickingHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.StockPickingCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	picking, err := h.service.Create(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picking)
}

// GetByID handles retrieving a stock picking by ID
func (h *StockPickingHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	picking, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if picking == nil {
		http.Error(w, "Stock picking not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picking)
}

// List handles listing all stock pickings
func (h *StockPickingHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	orgID, err := auth.GetOrganizationIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pickings, err := h.service.List(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pickings)
}

// Update handles updating a stock picking
func (h *StockPickingHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req types.StockPickingUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	picking, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picking)
}

// Delete handles deleting a stock picking
func (h *StockPickingHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

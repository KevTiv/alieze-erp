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

// StockLotHandler handles HTTP requests for stock lots
type StockLotHandler struct {
	service *service.StockLotService
}

// NewStockLotHandler creates a new StockLotHandler
type NewStockLotHandler(service *service.StockLotService) *StockLotHandler {
	return &StockLotHandler{
		service: service,
	}
}

// RegisterRoutes registers stock lot routes
func (h *StockLotHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/inventory/stock-lots", h.Create)
	router.GET("/api/inventory/stock-lots/:id", h.GetByID)
	router.GET("/api/inventory/stock-lots", h.List)
	router.PUT("/api/inventory/stock-lots/:id", h.Update)
	router.DELETE("/api/inventory/stock-lots/:id", h.Delete)
}

// Create handles stock lot creation
func (h *StockLotHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.StockLotCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	lot, err := h.service.Create(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lot)
}

// GetByID handles retrieving a stock lot by ID
func (h *StockLotHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	lot, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if lot == nil {
		http.Error(w, "Stock lot not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lot)
}

// List handles listing all stock lots
func (h *StockLotHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	orgID, err := auth.GetOrganizationIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lots, err := h.service.List(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lots)
}

// Update handles updating a stock lot
func (h *StockLotHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req types.StockLotUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lot, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lot)
}

// Delete handles deleting a stock lot
func (h *StockLotHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

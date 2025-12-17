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

// StockPackageHandler handles HTTP requests for stock packages
type StockPackageHandler struct {
	service *service.StockPackageService
}

// NewStockPackageHandler creates a new StockPackageHandler
type NewStockPackageHandler(service *service.StockPackageService) *StockPackageHandler {
	return &StockPackageHandler{
		service: service,
	}
}

// RegisterRoutes registers stock package routes
func (h *StockPackageHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/inventory/stock-packages", h.Create)
	router.GET("/api/inventory/stock-packages/:id", h.GetByID)
	router.GET("/api/inventory/stock-packages", h.List)
	router.PUT("/api/inventory/stock-packages/:id", h.Update)
	router.DELETE("/api/inventory/stock-packages/:id", h.Delete)
}

// Create handles stock package creation
func (h *StockPackageHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.StockPackageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	packageModel, err := h.service.Create(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packageModel)
}

// GetByID handles retrieving a stock package by ID
func (h *StockPackageHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	packageModel, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if packageModel == nil {
		http.Error(w, "Stock package not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packageModel)
}

// List handles listing all stock packages
func (h *StockPackageHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	orgID, err := auth.GetOrganizationIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	packages, err := h.service.List(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packages)
}

// Update handles updating a stock package
func (h *StockPackageHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req types.StockPackageUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	packageModel, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packageModel)
}

// Delete handles deleting a stock package
func (h *StockPackageHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

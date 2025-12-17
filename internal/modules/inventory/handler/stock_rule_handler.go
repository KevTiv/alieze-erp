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

// StockRuleHandler handles HTTP requests for stock rules
type StockRuleHandler struct {
	service *service.StockRuleService
}

// NewStockRuleHandler creates a new StockRuleHandler
type NewStockRuleHandler(service *service.StockRuleService) *StockRuleHandler {
	return &StockRuleHandler{
		service: service,
	}
}

// RegisterRoutes registers stock rule routes
func (h *StockRuleHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/inventory/stock-rules", h.Create)
	router.GET("/api/inventory/stock-rules/:id", h.GetByID)
	router.GET("/api/inventory/stock-rules", h.List)
	router.PUT("/api/inventory/stock-rules/:id", h.Update)
	router.DELETE("/api/inventory/stock-rules/:id", h.Delete)
}

// Create handles stock rule creation
func (h *StockRuleHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.StockRuleCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	rule, err := h.service.Create(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// GetByID handles retrieving a stock rule by ID
func (h *StockRuleHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rule, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rule == nil {
		http.Error(w, "Stock rule not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// List handles listing all stock rules
func (h *StockRuleHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	orgID, err := auth.GetOrganizationIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rules, err := h.service.List(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// Update handles updating a stock rule
func (h *StockRuleHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req types.StockRuleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rule, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// Delete handles deleting a stock rule
func (h *StockRuleHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

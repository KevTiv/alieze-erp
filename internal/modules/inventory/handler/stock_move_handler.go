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

// StockMoveHandler handles HTTP requests for stock moves
type StockMoveHandler struct {
	service *service.StockMoveService
}

// NewStockMoveHandler creates a new StockMoveHandler
type NewStockMoveHandler(service *service.StockMoveService) *StockMoveHandler {
	return &StockMoveHandler{
		service: service,
	}
}

// RegisterRoutes registers stock move routes
func (h *StockMoveHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/inventory/stock-moves", h.Create)
	router.GET("/api/inventory/stock-moves/:id", h.GetByID)
	router.GET("/api/inventory/stock-moves", h.List)
	router.PUT("/api/inventory/stock-moves/:id", h.Update)
	router.DELETE("/api/inventory/stock-moves/:id", h.Delete)
}

// Create handles stock move creation
func (h *StockMoveHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.StockMoveCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	move, err := h.service.Create(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(move)
}

// GetByID handles retrieving a stock move by ID
func (h *StockMoveHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	move, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if move == nil {
		http.Error(w, "Stock move not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(move)
}

// List handles listing all stock moves
func (h *StockMoveHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	orgID, err := auth.GetOrganizationIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	moves, err := h.service.List(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(moves)
}

// Update handles updating a stock move
func (h *StockMoveHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req types.StockMoveUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	move, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(move)
}

// Delete handles deleting a stock move
func (h *StockMoveHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

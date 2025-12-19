package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/service"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type InventoryHandler struct {
	service *service.InventoryService
}

func NewInventoryHandler(service *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		service: service,
	}
}

func (h *InventoryHandler) RegisterRoutes(router *httprouter.Router) {
	// Warehouse routes
	router.POST("/api/inventory/warehouses", h.CreateWarehouse)
	router.GET("/api/inventory/warehouses/:id", h.GetWarehouse)
	router.GET("/api/inventory/warehouses", h.ListWarehouses)
	router.PUT("/api/inventory/warehouses/:id", h.UpdateWarehouse)
	router.DELETE("/api/inventory/warehouses/:id", h.DeleteWarehouse)

	// Stock Location routes
	router.POST("/api/inventory/locations", h.CreateLocation)
	router.GET("/api/inventory/locations/:location_id", h.GetLocation)
	router.GET("/api/inventory/locations", h.ListLocations)
	router.PUT("/api/inventory/locations/:location_id", h.UpdateLocation)
	router.DELETE("/api/inventory/locations/:location_id", h.DeleteLocation)

	// Stock Quant routes
	router.GET("/api/inventory/products/:product_id/stock", h.GetProductStock)
	router.GET("/api/inventory/locations/:location_id/stock", h.GetLocationStock)
	router.GET("/api/inventory/products/:product_id/locations/:location_id/available", h.GetAvailableQuantity)

	// Stock Move routes
	router.POST("/api/inventory/moves", h.CreateMove)
	router.GET("/api/inventory/moves/:id", h.GetMove)
	router.GET("/api/inventory/moves", h.ListMoves)
	router.POST("/api/inventory/moves/:id/confirm", h.ConfirmMove)
}

// Warehouse handlers

func (h *InventoryHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.Warehouse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdWarehouse, err := h.service.CreateWarehouse(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdWarehouse)
}

func (h *InventoryHandler) GetWarehouse(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	warehouse, err := h.service.GetWarehouse(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if warehouse == nil {
		http.Error(w, "Warehouse not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(warehouse)
}

func (h *InventoryHandler) ListWarehouses(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	warehouses, err := h.service.ListWarehouses(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(warehouses)
}

func (h *InventoryHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	var req types.Warehouse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = id

	updatedWarehouse, err := h.service.UpdateWarehouse(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedWarehouse)
}

func (h *InventoryHandler) DeleteWarehouse(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteWarehouse(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Stock Location handlers

func (h *InventoryHandler) CreateLocation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.StockLocation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdLocation, err := h.service.CreateLocation(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdLocation)
}

func (h *InventoryHandler) GetLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("location_id"))
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	location, err := h.service.GetLocation(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if location == nil {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(location)
}

func (h *InventoryHandler) ListLocations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	locations, err := h.service.ListLocations(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(locations)
}

func (h *InventoryHandler) UpdateLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("location_id"))
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	var req types.StockLocation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = id

	updatedLocation, err := h.service.UpdateLocation(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLocation)
}

func (h *InventoryHandler) DeleteLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("location_id"))
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteLocation(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Stock Quant handlers

func (h *InventoryHandler) GetProductStock(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	productID, err := uuid.Parse(ps.ByName("product_id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	stock, err := h.service.GetProductStock(r.Context(), orgID, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stock)
}

func (h *InventoryHandler) GetLocationStock(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	locationID, err := uuid.Parse(ps.ByName("location_id"))
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	stock, err := h.service.GetLocationStock(r.Context(), orgID, locationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stock)
}

func (h *InventoryHandler) GetAvailableQuantity(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	productID, err := uuid.Parse(ps.ByName("product_id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	locationID, err := uuid.Parse(ps.ByName("location_id"))
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	available, err := h.service.GetAvailableQuantity(r.Context(), orgID, productID, locationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"product_id":      productID,
		"location_id":     locationID,
		"available":       available,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// Stock Move handlers

func (h *InventoryHandler) CreateMove(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.StockMove
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdMove, err := h.service.CreateMove(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdMove)
}

func (h *InventoryHandler) GetMove(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid move ID", http.StatusBadRequest)
		return
	}

	move, err := h.service.GetMove(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if move == nil {
		http.Error(w, "Move not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(move)
}

func (h *InventoryHandler) ListMoves(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	moves, err := h.service.ListMoves(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(moves)
}

func (h *InventoryHandler) ConfirmMove(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid move ID", http.StatusBadRequest)
		return
	}

	if err := h.service.ConfirmMove(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

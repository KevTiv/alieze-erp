package handler

import (
	"encoding/json"
	"net/http"

	"alieze-erp/internal/modules/common/service"
	"alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// UOMCategoryHandler handles HTTP requests for UOM categories
type UOMCategoryHandler struct {
	service *service.UOMCategoryService
}

func NewUOMCategoryHandler(service *service.UOMCategoryService) *UOMCategoryHandler {
	return &UOMCategoryHandler{service: service}
}

func (h *UOMCategoryHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/v1/uom/categories", h.CreateUOMCategory)
	router.GET("/api/v1/uom/categories/:id", h.GetUOMCategory)
	router.GET("/api/v1/uom/categories", h.ListUOMCategories)
	router.PUT("/api/v1/uom/categories/:id", h.UpdateUOMCategory)
	router.DELETE("/api/v1/uom/categories/:id", h.DeleteUOMCategory)
	router.GET("/api/v1/uom/categories/:id/units", h.GetUOMCategoryWithUnits)
}

func (h *UOMCategoryHandler) CreateUOMCategory(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var req types.UOMCategoryCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := h.service.Create(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *UOMCategoryHandler) GetUOMCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid UOM category ID", http.StatusBadRequest)
		return
	}

	category, err := h.service.GetByID(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if category == nil {
		http.Error(w, "UOM category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *UOMCategoryHandler) ListUOMCategories(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	// Parse query parameters
	filter := types.UOMCategoryFilter{}
	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		// Parse limit
		filter.Limit = 100 // default
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		// Parse offset
		filter.Offset = 0 // default
	}

	categories, err := h.service.List(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *UOMCategoryHandler) UpdateUOMCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid UOM category ID", http.StatusBadRequest)
		return
	}

	var req types.UOMCategoryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := h.service.Update(ctx, id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *UOMCategoryHandler) DeleteUOMCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid UOM category ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UOMCategoryHandler) GetUOMCategoryWithUnits(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid UOM category ID", http.StatusBadRequest)
		return
	}

	categoryWithUnits, err := h.service.GetUOMCategoryWithUnits(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if categoryWithUnits == nil {
		http.Error(w, "UOM category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categoryWithUnits)
}

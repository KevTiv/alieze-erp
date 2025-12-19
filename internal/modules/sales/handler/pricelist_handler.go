package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KevTiv/alieze-erp/internal/modules/sales/types"
	"github.com/KevTiv/alieze-erp/internal/modules/sales/service"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type PricelistHandler struct {
	service *service.PricelistService
}

func NewPricelistHandler(service *service.PricelistService) *PricelistHandler {
	return &PricelistHandler{
		service: service,
	}
}

func (h *PricelistHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/sales/pricelists", h.CreatePricelist)
	router.GET("/api/sales/pricelists/:id", h.GetPricelist)
	router.GET("/api/sales/pricelists", h.ListPricelists)
	router.GET("/api/sales/pricelists/company/:company_id", h.ListPricelistsByCompany)
	router.GET("/api/sales/pricelists/company/:company_id/active", h.ListActivePricelistsByCompany)
	router.PUT("/api/sales/pricelists/:id", h.UpdatePricelist)
	router.DELETE("/api/sales/pricelists/:id", h.DeletePricelist)
}

func (h *PricelistHandler) CreatePricelist(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.Pricelist
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdPricelist, err := h.service.CreatePricelist(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPricelist)
}

func (h *PricelistHandler) GetPricelist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid pricelist ID", http.StatusBadRequest)
		return
	}

	pricelist, err := h.service.GetPricelist(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pricelist == nil {
		http.Error(w, "Pricelist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pricelist)
}

func (h *PricelistHandler) ListPricelists(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (would be set by middleware)
	orgID := uuid.New() // In real implementation, this would come from auth middleware

	pricelists, err := h.service.ListPricelists(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pricelists)
}

func (h *PricelistHandler) ListPricelistsByCompany(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	companyID, err := uuid.Parse(ps.ByName("company_id"))
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	pricelists, err := h.service.ListPricelistsByCompany(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pricelists)
}

func (h *PricelistHandler) ListActivePricelistsByCompany(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	companyID, err := uuid.Parse(ps.ByName("company_id"))
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	pricelists, err := h.service.ListActivePricelistsByCompany(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pricelists)
}

func (h *PricelistHandler) UpdatePricelist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid pricelist ID", http.StatusBadRequest)
		return
	}

	var req types.Pricelist
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = id

	updatedPricelist, err := h.service.UpdatePricelist(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPricelist)
}

func (h *PricelistHandler) DeletePricelist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid pricelist ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePricelist(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

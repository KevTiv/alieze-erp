package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"alieze-erp/internal/modules/accounting/repository"
	"alieze-erp/internal/modules/accounting/service"
	"alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type TaxHandler struct {
	service *service.TaxService
}

func NewTaxHandler(service *service.TaxService) *TaxHandler {
	return &TaxHandler{
		service: service,
	}
}

func (h *TaxHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/accounting/taxes", h.CreateTax)
	router.GET("/api/accounting/taxes/:id", h.GetTax)
	router.GET("/api/accounting/taxes", h.ListTaxes)
	router.PUT("/api/accounting/taxes/:id", h.UpdateTax)
	router.DELETE("/api/accounting/taxes/:id", h.DeleteTax)
	router.GET("/api/accounting/taxes/type/:type", h.GetTaxesByType)
}

func (h *TaxHandler) CreateTax(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.Tax
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdTax, err := h.service.CreateTax(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTax)
}

func (h *TaxHandler) GetTax(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid tax ID", http.StatusBadRequest)
		return
	}

	tax, err := h.service.GetTax(r.Context(), id)
	if err != nil {
		if err.Error() == "tax not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tax)
}

func (h *TaxHandler) ListTaxes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	orgIDStr := r.URL.Query().Get("organization_id")
	if orgIDStr == "" {
		http.Error(w, "organization_id is required", http.StatusBadRequest)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization_id", http.StatusBadRequest)
		return
	}

	filters := repository.TaxFilter{
		OrganizationID: orgID,
	}

	// Optional filters
	if companyIDStr := r.URL.Query().Get("company_id"); companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err == nil {
			filters.CompanyID = &companyID
		}
	}

	if typeTaxUse := r.URL.Query().Get("type_tax_use"); typeTaxUse != "" {
		filters.TypeTaxUse = &typeTaxUse
	}

	if amountType := r.URL.Query().Get("amount_type"); amountType != "" {
		filters.AmountType = &amountType
	}

	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		active := activeStr == "true"
		filters.Active = &active
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	taxes, err := h.service.ListTaxes(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taxes)
}

func (h *TaxHandler) UpdateTax(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid tax ID", http.StatusBadRequest)
		return
	}

	var req types.Tax
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.ID = id

	updatedTax, err := h.service.UpdateTax(r.Context(), req)
	if err != nil {
		if err.Error() == "tax not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTax)
}

func (h *TaxHandler) DeleteTax(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid tax ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteTax(r.Context(), id)
	if err != nil {
		if err.Error() == "tax not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaxHandler) GetTaxesByType(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	typeTaxUse := ps.ByName("type")
	if typeTaxUse == "" {
		http.Error(w, "Tax type is required", http.StatusBadRequest)
		return
	}

	orgIDStr := r.URL.Query().Get("organization_id")
	if orgIDStr == "" {
		http.Error(w, "organization_id is required", http.StatusBadRequest)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization_id", http.StatusBadRequest)
		return
	}

	taxes, err := h.service.GetTaxesByType(r.Context(), orgID, typeTaxUse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taxes)
}

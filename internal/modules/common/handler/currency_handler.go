package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"alieze-erp/internal/modules/common/service"
	"alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// CurrencyHandler handles HTTP requests for currencies
type CurrencyHandler struct {
	service *service.CurrencyService
}

func NewCurrencyHandler(service *service.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{service: service}
}

func (h *CurrencyHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/v1/currencies", h.CreateCurrency)
	router.GET("/api/v1/currencies/:id", h.GetCurrency)
	router.GET("/api/v1/currencies", h.ListCurrencies)
	router.PUT("/api/v1/currencies/:id", h.UpdateCurrency)
	router.DELETE("/api/v1/currencies/:id", h.DeleteCurrency)
	router.GET("/api/v1/currencies/code/:code", h.GetCurrencyByCode)
	router.GET("/api/v1/currencies/default", h.GetDefaultCurrency)
	router.POST("/api/v1/currencies/format", h.FormatAmount)
}

func (h *CurrencyHandler) CreateCurrency(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var req types.CurrencyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currency, err := h.service.Create(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
}

func (h *CurrencyHandler) GetCurrency(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid currency ID", http.StatusBadRequest)
		return
	}

	currency, err := h.service.GetByID(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if currency == nil {
		http.Error(w, "currency not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
}

func (h *CurrencyHandler) GetCurrencyByCode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	code := ps.ByName("code")
	if code == "" {
		http.Error(w, "currency code is required", http.StatusBadRequest)
		return
	}

	currency, err := h.service.GetByCode(ctx, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if currency == nil {
		http.Error(w, "currency not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
}

func (h *CurrencyHandler) ListCurrencies(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	// Parse query parameters
	filter := types.CurrencyFilter{}
	if active := r.URL.Query().Get("active"); active != "" {
		activeBool := active == "true"
		filter.Active = &activeBool
	}
	if code := r.URL.Query().Get("code"); code != "" {
		filter.Code = &code
	}
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

	currencies, err := h.service.List(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currencies)
}

func (h *CurrencyHandler) UpdateCurrency(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid currency ID", http.StatusBadRequest)
		return
	}

	var req types.CurrencyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currency, err := h.service.Update(ctx, id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
}

func (h *CurrencyHandler) DeleteCurrency(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid currency ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CurrencyHandler) GetDefaultCurrency(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	currency, err := h.service.GetDefaultCurrency(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if currency == nil {
		http.Error(w, "default currency not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
}

func (h *CurrencyHandler) FormatAmount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var req struct {
		CurrencyCode string  `json:"currency_code"`
		Amount       float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CurrencyCode == "" {
		http.Error(w, "currency_code is required", http.StatusBadRequest)
		return
	}

	formatted, err := h.service.FormatAmount(ctx, req.CurrencyCode, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"formatted_amount": formatted})
}

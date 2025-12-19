package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/KevTiv/alieze-erp/internal/modules/common/service"
	"github.com/KevTiv/alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// CountryHandler handles HTTP requests for countries
type CountryHandler struct {
	service *service.CountryService
}

func NewCountryHandler(service *service.CountryService) *CountryHandler {
	return &CountryHandler{service: service}
}

func (h *CountryHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/v1/countries", h.CreateCountry)
	router.GET("/api/v1/countries/:id", h.GetCountry)
	router.GET("/api/v1/countries", h.ListCountries)
	router.PUT("/api/v1/countries/:id", h.UpdateCountry)
	router.DELETE("/api/v1/countries/:id", h.DeleteCountry)
	router.GET("/api/v1/countries/code/:code", h.GetCountryByCode)
	router.GET("/api/v1/countries/:id/states", h.GetCountryWithStates)
}

func (h *CountryHandler) CreateCountry(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var req types.CountryCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	country, err := h.service.Create(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(country)
}

func (h *CountryHandler) GetCountry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid country ID", http.StatusBadRequest)
		return
	}

	country, err := h.service.GetByID(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country == nil {
		http.Error(w, "country not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(country)
}

func (h *CountryHandler) GetCountryByCode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	code := ps.ByName("code")
	if code == "" {
		http.Error(w, "country code is required", http.StatusBadRequest)
		return
	}

	country, err := h.service.GetByCode(ctx, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country == nil {
		http.Error(w, "country not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(country)
}

func (h *CountryHandler) ListCountries(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	// Parse query parameters
	filter := types.CountryFilter{}
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

	countries, err := h.service.List(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(countries)
}

func (h *CountryHandler) UpdateCountry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid country ID", http.StatusBadRequest)
		return
	}

	var req types.CountryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	country, err := h.service.Update(ctx, id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(country)
}

func (h *CountryHandler) DeleteCountry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid country ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CountryHandler) GetCountryWithStates(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid country ID", http.StatusBadRequest)
		return
	}

	countryWithStates, err := h.service.GetCountryWithStates(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if countryWithStates == nil {
		http.Error(w, "country not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(countryWithStates)
}

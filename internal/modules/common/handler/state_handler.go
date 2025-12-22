package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KevTiv/alieze-erp/internal/modules/common/service"
	"github.com/KevTiv/alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"

	"github.com/julienschmidt/httprouter"
)

// StateHandler handles HTTP requests for states
type StateHandler struct {
	service *service.StateService
}

func NewStateHandler(service *service.StateService) *StateHandler {
	return &StateHandler{service: service}
}

func (h *StateHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/v1/states", h.CreateState)
	router.GET("/api/v1/states/:id", h.GetState)
	router.GET("/api/v1/states", h.ListStates)
	router.PUT("/api/v1/states/:id", h.UpdateState)
	router.DELETE("/api/v1/states/:id", h.DeleteState)
	router.GET("/api/v1/countries/:country_id/states", h.ListStatesByCountry)
}

func (h *StateHandler) CreateState(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var req types.StateCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	state, err := h.service.Create(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (h *StateHandler) GetState(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid state ID", http.StatusBadRequest)
		return
	}

	state, err := h.service.GetByID(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if state == nil {
		http.Error(w, "state not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (h *StateHandler) ListStates(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	// Parse query parameters
	filter := types.StateFilter{}
	if countryID := r.URL.Query().Get("country_id"); countryID != "" {
		parsedID, err := uuid.Parse(countryID)
		if err != nil {
			http.Error(w, "invalid country_id", http.StatusBadRequest)
			return
		}
		filter.CountryID = &parsedID
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

	states, err := h.service.List(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(states)
}

func (h *StateHandler) ListStatesByCountry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	countryID, err := uuid.Parse(ps.ByName("country_id"))
	if err != nil {
		http.Error(w, "invalid country ID", http.StatusBadRequest)
		return
	}

	states, err := h.service.ListByCountry(ctx, countryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(states)
}

func (h *StateHandler) UpdateState(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid state ID", http.StatusBadRequest)
		return
	}

	var req types.StateUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	state, err := h.service.Update(ctx, id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (h *StateHandler) DeleteState(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid state ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

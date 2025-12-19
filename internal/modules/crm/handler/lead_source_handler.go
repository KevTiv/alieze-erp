package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"
)

type LeadSourceHandler struct {
	service *service.LeadSourceService
}

func NewLeadSourceHandler(service *service.LeadSourceService) *LeadSourceHandler {
	return &LeadSourceHandler{
		service: service,
	}
}

func (h *LeadSourceHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/crm/lead-sources", h.CreateLeadSource)
	router.GET("/api/crm/lead-sources/:id", h.GetLeadSource)
	router.GET("/api/crm/lead-sources", h.ListLeadSources)
	router.PUT("/api/crm/lead-sources/:id", h.UpdateLeadSource)
	router.DELETE("/api/crm/lead-sources/:id", h.DeleteLeadSource)
}

func (h *LeadSourceHandler) CreateLeadSource(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.LeadSourceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateLeadSource(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *LeadSourceHandler) GetLeadSource(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	source, err := h.service.GetLeadSource(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(source)
}

func (h *LeadSourceHandler) ListLeadSources(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	filter := types.LeadSourceFilter{}

	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	sources, err := h.service.ListLeadSources(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sources)
}

func (h *LeadSourceHandler) UpdateLeadSource(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	var req types.LeadSourceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.service.UpdateLeadSource(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *LeadSourceHandler) DeleteLeadSource(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteLeadSource(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

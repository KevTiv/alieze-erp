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

type LeadStageHandler struct {
	service *service.LeadStageService
}

func NewLeadStageHandler(service *service.LeadStageService) *LeadStageHandler {
	return &LeadStageHandler{
		service: service,
	}
}

func (h *LeadStageHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/crm/lead-stages", h.CreateLeadStage)
	router.GET("/api/crm/lead-stages/:id", h.GetLeadStage)
	router.GET("/api/crm/lead-stages", h.ListLeadStages)
	router.PUT("/api/crm/lead-stages/:id", h.UpdateLeadStage)
	router.DELETE("/api/crm/lead-stages/:id", h.DeleteLeadStage)
}

func (h *LeadStageHandler) CreateLeadStage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.LeadStageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateLeadStage(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *LeadStageHandler) GetLeadStage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	stage, err := h.service.GetLeadStage(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stage)
}

func (h *LeadStageHandler) ListLeadStages(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	filter := types.LeadStageFilter{}

	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}

	if isWon := r.URL.Query().Get("is_won"); isWon != "" {
		if won, err := strconv.ParseBool(isWon); err == nil {
			filter.IsWon = &won
		}
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		if id, err := uuid.Parse(teamID); err == nil {
			filter.TeamID = &id
		}
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

	stages, err := h.service.ListLeadStages(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stages)
}

func (h *LeadStageHandler) UpdateLeadStage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	var req types.LeadStageUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.service.UpdateLeadStage(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *LeadStageHandler) DeleteLeadStage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteLeadStage(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

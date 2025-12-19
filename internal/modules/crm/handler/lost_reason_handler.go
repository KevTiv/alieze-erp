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

type LostReasonHandler struct {
	service *service.LostReasonService
}

func NewLostReasonHandler(service *service.LostReasonService) *LostReasonHandler {
	return &LostReasonHandler{
		service: service,
	}
}

func (h *LostReasonHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/crm/lost-reasons", h.CreateLostReason)
	router.GET("/api/crm/lost-reasons/:id", h.GetLostReason)
	router.GET("/api/crm/lost-reasons", h.ListLostReasons)
	router.PUT("/api/crm/lost-reasons/:id", h.UpdateLostReason)
	router.DELETE("/api/crm/lost-reasons/:id", h.DeleteLostReason)
}

func (h *LostReasonHandler) CreateLostReason(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.LostReasonCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateLostReason(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *LostReasonHandler) GetLostReason(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid reason ID", http.StatusBadRequest)
		return
	}

	reason, err := h.service.GetLostReason(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reason)
}

func (h *LostReasonHandler) ListLostReasons(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	filter := types.LostReasonFilter{}

	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}

	if active := r.URL.Query().Get("active"); active != "" {
		if a, err := strconv.ParseBool(active); err == nil {
			filter.Active = &a
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

	reasons, err := h.service.ListLostReasons(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reasons)
}

func (h *LostReasonHandler) UpdateLostReason(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid reason ID", http.StatusBadRequest)
		return
	}

	var req types.LostReasonUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.service.UpdateLostReason(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *LostReasonHandler) DeleteLostReason(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid reason ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteLostReason(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

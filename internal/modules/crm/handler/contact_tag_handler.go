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

type ContactTagHandler struct {
	service *service.ContactTagService
}

func NewContactTagHandler(service *service.ContactTagService) *ContactTagHandler {
	return &ContactTagHandler{
		service: service,
	}
}

func (h *ContactTagHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/crm/contact-tags", h.CreateContactTag)
	router.GET("/api/crm/contact-tags/:id", h.GetContactTag)
	router.GET("/api/crm/contact-tags", h.ListContactTags)
	router.PUT("/api/crm/contact-tags/:id", h.UpdateContactTag)
	router.DELETE("/api/crm/contact-tags/:id", h.DeleteContactTag)
}

func (h *ContactTagHandler) CreateContactTag(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.ContactTag
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateContactTag(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *ContactTagHandler) GetContactTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	tag, err := h.service.GetContactTag(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tag)
}

func (h *ContactTagHandler) ListContactTags(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	filter := types.ContactTagFilter{}

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

	tags, err := h.service.ListContactTags(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tags)
}

func (h *ContactTagHandler) UpdateContactTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	var req types.ContactTag
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.ID = id

	updated, err := h.service.UpdateContactTag(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *ContactTagHandler) DeleteContactTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteContactTag(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/internal/modules/crm/types"
)

type ActivityHandler struct {
	service *service.ActivityService
}

func NewActivityHandler(service *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		service: service,
	}
}

func (h *ActivityHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/crm/activities", h.CreateActivity)
	router.GET("/api/crm/activities/:id", h.GetActivity)
	router.GET("/api/crm/activities", h.ListActivities)
	router.PUT("/api/crm/activities/:id", h.UpdateActivity)
	router.DELETE("/api/crm/activities/:id", h.DeleteActivity)
	router.GET("/api/crm/contacts/:contact_id/activities", h.GetActivitiesByContact)
	router.GET("/api/crm/leads/:lead_id/activities", h.GetActivitiesByLead)
}

func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.ActivityCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateActivity(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *ActivityHandler) GetActivity(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	activity, err := h.service.GetActivity(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(activity)
}

func (h *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	filter := types.ActivityFilter{}

	if activityType := r.URL.Query().Get("activity_type"); activityType != "" {
		at := types.ActivityType(activityType)
		filter.ActivityType = &at
	}

	if state := r.URL.Query().Get("state"); state != "" {
		s := types.ActivityState(state)
		filter.State = &s
	}

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			filter.UserID = &id
		}
	}

	if assignedTo := r.URL.Query().Get("assigned_to"); assignedTo != "" {
		if id, err := uuid.Parse(assignedTo); err == nil {
			filter.AssignedTo = &id
		}
	}

	if resModel := r.URL.Query().Get("res_model"); resModel != "" {
		filter.ResModel = &resModel
	}

	if resID := r.URL.Query().Get("res_id"); resID != "" {
		if id, err := uuid.Parse(resID); err == nil {
			filter.ResID = &id
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

	activities, err := h.service.ListActivities(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(activities)
}

func (h *ActivityHandler) UpdateActivity(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	var req types.ActivityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.service.UpdateActivity(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *ActivityHandler) DeleteActivity(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteActivity(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ActivityHandler) GetActivitiesByContact(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	contactID, err := uuid.Parse(ps.ByName("contact_id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	activities, err := h.service.GetActivitiesByContact(r.Context(), contactID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(activities)
}

func (h *ActivityHandler) GetActivitiesByLead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	leadID, err := uuid.Parse(ps.ByName("lead_id"))
	if err != nil {
		http.Error(w, "Invalid lead ID", http.StatusBadRequest)
		return
	}

	activities, err := h.service.GetActivitiesByLead(r.Context(), leadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(activities)
}

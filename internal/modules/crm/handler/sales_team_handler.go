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

type SalesTeamHandler struct {
	service *service.SalesTeamService
}

func NewSalesTeamHandler(service *service.SalesTeamService) *SalesTeamHandler {
	return &SalesTeamHandler{
		service: service,
	}
}

func (h *SalesTeamHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/crm/sales-teams", h.CreateSalesTeam)
	router.GET("/api/crm/sales-teams/:id", h.GetSalesTeam)
	router.GET("/api/crm/sales-teams", h.ListSalesTeams)
	router.PUT("/api/crm/sales-teams/:id", h.UpdateSalesTeam)
	router.DELETE("/api/crm/sales-teams/:id", h.DeleteSalesTeam)
	router.GET("/api/crm/sales-teams/by-member/:member_id", h.GetSalesTeamsByMember)
}

func (h *SalesTeamHandler) CreateSalesTeam(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.SalesTeamCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateSalesTeam(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *SalesTeamHandler) GetSalesTeam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	team, err := h.service.GetSalesTeam(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(team)
}

func (h *SalesTeamHandler) ListSalesTeams(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	filter := types.SalesTeamFilter{}

	if companyID := r.URL.Query().Get("company_id"); companyID != "" {
		if id, err := uuid.Parse(companyID); err == nil {
			filter.CompanyID = &id
		}
	}

	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}

	if code := r.URL.Query().Get("code"); code != "" {
		filter.Code = &code
	}

	if teamLeaderID := r.URL.Query().Get("team_leader_id"); teamLeaderID != "" {
		if id, err := uuid.Parse(teamLeaderID); err == nil {
			filter.TeamLeaderID = &id
		}
	}

	if isActive := r.URL.Query().Get("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filter.IsActive = &active
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

	teams, err := h.service.ListSalesTeams(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(teams)
}

func (h *SalesTeamHandler) UpdateSalesTeam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var req types.SalesTeamUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.service.UpdateSalesTeam(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *SalesTeamHandler) DeleteSalesTeam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteSalesTeam(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SalesTeamHandler) GetSalesTeamsByMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	memberID, err := uuid.Parse(ps.ByName("member_id"))
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	teams, err := h.service.GetSalesTeamsByMember(r.Context(), memberID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(teams)
}

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

type JournalHandler struct {
	service *service.JournalService
}

func NewJournalHandler(service *service.JournalService) *JournalHandler {
	return &JournalHandler{
		service: service,
	}
}

func (h *JournalHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/accounting/journals", h.CreateJournal)
	router.GET("/api/accounting/journals/:id", h.GetJournal)
	router.GET("/api/accounting/journals", h.ListJournals)
	router.PUT("/api/accounting/journals/:id", h.UpdateJournal)
	router.DELETE("/api/accounting/journals/:id", h.DeleteJournal)
	router.GET("/api/accounting/journals/type/:type", h.GetJournalsByType)
}

func (h *JournalHandler) CreateJournal(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.Journal
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdJournal, err := h.service.CreateJournal(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdJournal)
}

func (h *JournalHandler) GetJournal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid journal ID", http.StatusBadRequest)
		return
	}

	journal, err := h.service.GetJournal(r.Context(), id)
	if err != nil {
		if err.Error() == "journal not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(journal)
}

func (h *JournalHandler) ListJournals(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	filters := repository.JournalFilter{
		OrganizationID: orgID,
	}

	// Optional filters
	if companyIDStr := r.URL.Query().Get("company_id"); companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err == nil {
			filters.CompanyID = &companyID
		}
	}

	if journalType := r.URL.Query().Get("type"); journalType != "" {
		filters.Type = &journalType
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

	journals, err := h.service.ListJournals(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(journals)
}

func (h *JournalHandler) UpdateJournal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid journal ID", http.StatusBadRequest)
		return
	}

	var req types.Journal
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.ID = id

	updatedJournal, err := h.service.UpdateJournal(r.Context(), req)
	if err != nil {
		if err.Error() == "journal not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedJournal)
}

func (h *JournalHandler) DeleteJournal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid journal ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteJournal(r.Context(), id)
	if err != nil {
		if err.Error() == "journal not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *JournalHandler) GetJournalsByType(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	journalType := ps.ByName("type")
	if journalType == "" {
		http.Error(w, "Journal type is required", http.StatusBadRequest)
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

	journals, err := h.service.GetJournalsByType(r.Context(), orgID, journalType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(journals)
}

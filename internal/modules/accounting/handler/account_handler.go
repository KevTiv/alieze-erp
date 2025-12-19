package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KevTiv/alieze-erp/internal/modules/accounting/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/accounting/service"
	"github.com/KevTiv/alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{
		service: service,
	}
}

func (h *AccountHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/accounting/accounts", h.CreateAccount)
	router.GET("/api/accounting/accounts/:id", h.GetAccount)
	router.GET("/api/accounting/accounts", h.ListAccounts)
	router.PUT("/api/accounting/accounts/:id", h.UpdateAccount)
	router.DELETE("/api/accounting/accounts/:id", h.DeleteAccount)
	router.GET("/api/accounting/accounts/type/:type", h.GetAccountsByType)
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.Account
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdAccount, err := h.service.CreateAccount(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAccount)
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	account, err := h.service.GetAccount(r.Context(), id)
	if err != nil {
		if err.Error() == "account not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	filters := repository.AccountFilter{
		OrganizationID: orgID,
	}

	// Optional filters
	if companyIDStr := r.URL.Query().Get("company_id"); companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err == nil {
			filters.CompanyID = &companyID
		}
	}

	if accountType := r.URL.Query().Get("account_type"); accountType != "" {
		filters.AccountType = &accountType
	}

	if deprecatedStr := r.URL.Query().Get("deprecated"); deprecatedStr != "" {
		deprecated := deprecatedStr == "true"
		filters.Deprecated = &deprecated
	}

	if reconcileStr := r.URL.Query().Get("reconcile"); reconcileStr != "" {
		reconcile := reconcileStr == "true"
		filters.Reconcile = &reconcile
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

	accounts, err := h.service.ListAccounts(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
}

func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	var req types.Account
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.ID = id

	updatedAccount, err := h.service.UpdateAccount(r.Context(), req)
	if err != nil {
		if err.Error() == "account not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedAccount)
}

func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteAccount(r.Context(), id)
	if err != nil {
		if err.Error() == "account not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) GetAccountsByType(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	accountType := ps.ByName("type")
	if accountType == "" {
		http.Error(w, "Account type is required", http.StatusBadRequest)
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

	accounts, err := h.service.GetAccountsByType(r.Context(), orgID, accountType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
}

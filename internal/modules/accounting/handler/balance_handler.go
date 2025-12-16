package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type BalanceHandler struct {
	// For now, this is a placeholder for future balance reporting
	// Full implementation would require journal entries and move lines
}

func NewBalanceHandler() *BalanceHandler {
	return &BalanceHandler{}
}

func (h *BalanceHandler) RegisterRoutes(router *httprouter.Router) {
	router.GET("/api/accounting/balances", h.GetTrialBalance)
	router.GET("/api/accounting/balances/:account_id", h.GetAccountBalance)
}

type BalanceResponse struct {
	AccountID   uuid.UUID `json:"account_id"`
	AccountCode string    `json:"account_code"`
	AccountName string    `json:"account_name"`
	Debit       float64   `json:"debit"`
	Credit      float64   `json:"credit"`
	Balance     float64   `json:"balance"`
}

type TrialBalanceResponse struct {
	Accounts      []BalanceResponse `json:"accounts"`
	TotalDebit    float64           `json:"total_debit"`
	TotalCredit   float64           `json:"total_credit"`
	TotalBalance  float64           `json:"total_balance"`
	AsOfDate      string            `json:"as_of_date"`
	OrganizationID uuid.UUID        `json:"organization_id"`
}

func (h *BalanceHandler) GetTrialBalance(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	// TODO: Implement actual trial balance calculation
	// This requires querying journal entries and move lines
	// For now, return a placeholder response
	response := TrialBalanceResponse{
		Accounts:       []BalanceResponse{},
		TotalDebit:     0,
		TotalCredit:    0,
		TotalBalance:   0,
		AsOfDate:       "2025-12-16",
		OrganizationID: orgID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *BalanceHandler) GetAccountBalance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	accountID, err := uuid.Parse(ps.ByName("account_id"))
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual account balance calculation
	// This requires querying journal entries and move lines for this account
	// For now, return a placeholder response
	response := BalanceResponse{
		AccountID:   accountID,
		AccountCode: "PLACEHOLDER",
		AccountName: "Placeholder Account",
		Debit:       0,
		Credit:      0,
		Balance:     0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

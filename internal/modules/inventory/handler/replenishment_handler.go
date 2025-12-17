package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"alieze-erp/internal/modules/inventory/service"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReplenishmentHandler struct {
	replenishmentService *service.ReplenishmentService
}

func NewReplenishmentHandler(
	replenishmentService *service.ReplenishmentService,
) *ReplenishmentHandler {
	return &ReplenishmentHandler{
		replenishmentService: replenishmentService,
	}
}

func (h *ReplenishmentHandler) RegisterRoutes(router chi.Router) {
	router.Route("/replenishment", func(r chi.Router) {
		r.Use(h.authService.Middleware)

		// Replenishment Rules
		r.Route("/rules", func(rr chi.Router) {
			rr.Post("/", h.CreateReplenishmentRule)
			rr.Get("/", h.ListReplenishmentRules)
			rr.Get("/{ruleID}", h.GetReplenishmentRule)
			rr.Put("/{ruleID}", h.UpdateReplenishmentRule)
			rr.Delete("/{ruleID}", h.DeleteReplenishmentRule)
		})

		// Replenishment Orders
		r.Route("/orders", func(ro chi.Router) {
			ro.Post("/", h.CreateReplenishmentOrder)
			ro.Get("/", h.ListReplenishmentOrders)
			ro.Get("/{orderID}", h.GetReplenishmentOrder)
			ro.Put("/{orderID}", h.UpdateReplenishmentOrder)
			ro.Delete("/{orderID}", h.DeleteReplenishmentOrder)
			ro.Get("/status/{status}", h.ListReplenishmentOrdersByStatus)
		})

		// Replenishment Processing
		r.Post("/check", h.CheckAndCreateReplenishmentOrders)
		r.Post("/process", h.ProcessReplenishmentOrders)
		r.Post("/cycle", h.RunReplenishmentCycle)
		r.Get("/statistics", h.GetReplenishmentStatistics)
		r.Post("/manual-check", h.CheckReplenishmentNeeds)
	})
}

// Replenishment Rule Handlers

func (h *ReplenishmentHandler) CreateReplenishmentRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var rule domain.ReplenishmentRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Set organization from context
	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}
	rule.OrganizationID = orgID

	createdRule, err := h.replenishmentService.CreateReplenishmentRule(ctx, rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdRule)
}

func (h *ReplenishmentHandler) GetReplenishmentRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ruleIDStr := chi.URLParam(r, "ruleID")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	rule, err := h.replenishmentService.GetReplenishmentRule(ctx, ruleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, rule)
}

func (h *ReplenishmentHandler) ListReplenishmentRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	rules, err := h.replenishmentService.ListReplenishmentRules(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, rules)
}

func (h *ReplenishmentHandler) UpdateReplenishmentRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ruleIDStr := chi.URLParam(r, "ruleID")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	var rule domain.ReplenishmentRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	rule.ID = ruleID

	updatedRule, err := h.replenishmentService.UpdateReplenishmentRule(ctx, rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedRule)
}

func (h *ReplenishmentHandler) DeleteReplenishmentRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ruleIDStr := chi.URLParam(r, "ruleID")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	err = h.replenishmentService.DeleteReplenishmentRule(ctx, ruleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Replenishment rule deleted successfully"})
}

// Replenishment Order Handlers

func (h *ReplenishmentHandler) CreateReplenishmentOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var order domain.ReplenishmentOrder
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Set organization from context
	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}
	order.OrganizationID = orgID

	createdOrder, err := h.replenishmentService.CreateReplenishmentOrder(ctx, order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdOrder)
}

func (h *ReplenishmentHandler) GetReplenishmentOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderIDStr := chi.URLParam(r, "orderID")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.replenishmentService.GetReplenishmentOrder(ctx, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, order)
}

func (h *ReplenishmentHandler) ListReplenishmentOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	orders, err := h.replenishmentService.ListReplenishmentOrders(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

func (h *ReplenishmentHandler) ListReplenishmentOrdersByStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	status := chi.URLParam(r, "status")

	orders, err := h.replenishmentService.ListReplenishmentOrdersByStatus(ctx, orgID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

func (h *ReplenishmentHandler) UpdateReplenishmentOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderIDStr := chi.URLParam(r, "orderID")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var order domain.ReplenishmentOrder
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	order.ID = orderID

	updatedOrder, err := h.replenishmentService.UpdateReplenishmentOrder(ctx, order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedOrder)
}

func (h *ReplenishmentHandler) DeleteReplenishmentOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderIDStr := chi.URLParam(r, "orderID")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	err = h.replenishmentService.DeleteReplenishmentOrder(ctx, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Replenishment order deleted successfully"})
}

// Replenishment Processing Handlers

func (h *ReplenishmentHandler) CheckAndCreateReplenishmentOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	// Get limit parameter
	limit := 100
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.replenishmentService.CheckAndCreateReplenishmentOrders(ctx, orgID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (h *ReplenishmentHandler) ProcessReplenishmentOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	// Get limit parameter
	limit := 20
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	orders, err := h.replenishmentService.ProcessReplenishmentOrders(ctx, orgID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, orders)
}

func (h *ReplenishmentHandler) RunReplenishmentCycle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	result, err := h.replenishmentService.RunReplenishmentCycle(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

func (h *ReplenishmentHandler) GetReplenishmentStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	stats, err := h.replenishmentService.GetReplenishmentStatistics(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

func (h *ReplenishmentHandler) CheckReplenishmentNeeds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	results, err := h.replenishmentService.CheckReplenishmentNeeds(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, results)
}

// Helper function
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

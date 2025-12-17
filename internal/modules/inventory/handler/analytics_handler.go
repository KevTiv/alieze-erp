package handler

import (
	"encoding/json"
	"net/http"

	"alieze-erp/internal/modules/inventory/service"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
}

func NewAnalyticsHandler(service *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
	}
}

func (h *AnalyticsHandler) RegisterRoutes(router *httprouter.Router) {
	// Valuation endpoints
	router.GET("/api/inventory/analytics/valuation", h.GetInventoryValuation)
	router.GET("/api/inventory/analytics/valuation/:product_id", h.GetValuationByProduct)
	router.GET("/api/inventory/analytics/valuation/summary", h.GetValuationSummary)

	// Turnover endpoints
	router.GET("/api/inventory/analytics/turnover", h.GetInventoryTurnover)
	router.GET("/api/inventory/analytics/turnover/:product_id", h.GetTurnoverByProduct)

	// Aging endpoints
	router.GET("/api/inventory/analytics/aging", h.GetInventoryAging)
	router.GET("/api/inventory/analytics/aging/summary", h.GetAgingSummary)

	// Dead stock endpoints
	router.GET("/api/inventory/analytics/dead-stock", h.GetDeadStock)
	router.GET("/api/inventory/analytics/dead-stock/summary", h.GetDeadStockSummary)

	// Movement summary endpoints
	router.GET("/api/inventory/analytics/movement", h.GetMovementSummary)

	// Reorder analysis endpoints
	router.GET("/api/inventory/analytics/reorder", h.GetReorderAnalysis)
	router.GET("/api/inventory/analytics/reorder/needing", h.GetProductsNeedingReorder)

	// Snapshot endpoints
	router.GET("/api/inventory/analytics/snapshot", h.GetInventorySnapshot)

	// Refresh endpoints
	router.POST("/api/inventory/analytics/refresh", h.RefreshOrganizationAnalytics)

	// Dashboard endpoint
	router.GET("/api/inventory/analytics/dashboard", h.GetInventoryDashboard)
}

// GetInventoryValuation handles requests for inventory valuation data
func (h *AnalyticsHandler) GetInventoryValuation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var request types.AnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Get valuation data
	valuations, err := h.service.GetInventoryValuation(r.Context(), orgID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(valuations)
}

// GetValuationByProduct handles requests for valuation of a specific product
func (h *AnalyticsHandler) GetValuationByProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get product ID from URL
	productID, err := uuid.Parse(ps.ByName("product_id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get valuation data
	valuation, err := h.service.GetValuationByProduct(r.Context(), orgID, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if valuation == nil {
		http.Error(w, "Valuation not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(valuation)
}

// GetValuationSummary handles requests for valuation summary
func (h *AnalyticsHandler) GetValuationSummary(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get summary data
	summary, err := h.service.GetValuationSummary(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

// GetInventoryTurnover handles requests for inventory turnover data
func (h *AnalyticsHandler) GetInventoryTurnover(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var request types.AnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Get turnover data
	turnover, err := h.service.GetInventoryTurnover(r.Context(), orgID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(turnover)
}

// GetTurnoverByProduct handles requests for turnover of a specific product
func (h *AnalyticsHandler) GetTurnoverByProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get product ID from URL
	productID, err := uuid.Parse(ps.ByName("product_id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get turnover data
	turnover, err := h.service.GetTurnoverByProduct(r.Context(), orgID, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if turnover == nil {
		http.Error(w, "Turnover data not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(turnover)
}

// GetInventoryAging handles requests for inventory aging data
func (h *AnalyticsHandler) GetInventoryAging(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var request types.AnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Get aging data
	aging, err := h.service.GetInventoryAging(r.Context(), orgID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(aging)
}

// GetAgingSummary handles requests for aging summary
func (h *AnalyticsHandler) GetAgingSummary(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get aging summary
	summary, err := h.service.GetAgingSummary(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

// GetDeadStock handles requests for dead stock data
func (h *AnalyticsHandler) GetDeadStock(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var request types.AnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Get dead stock data
	deadStock, err := h.service.GetDeadStock(r.Context(), orgID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deadStock)
}

// GetDeadStockSummary handles requests for dead stock summary
func (h *AnalyticsHandler) GetDeadStockSummary(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get dead stock summary
	summary, err := h.service.GetDeadStockSummary(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

// GetMovementSummary handles requests for movement summary data
func (h *AnalyticsHandler) GetMovementSummary(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var request types.AnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Get movement summary data
	summary, err := h.service.GetMovementSummary(r.Context(), orgID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

// GetReorderAnalysis handles requests for reorder analysis data
func (h *AnalyticsHandler) GetReorderAnalysis(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var request types.AnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Get reorder analysis data
	analysis, err := h.service.GetReorderAnalysis(r.Context(), orgID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(analysis)
}

// GetProductsNeedingReorder handles requests for products needing reorder
func (h *AnalyticsHandler) GetProductsNeedingReorder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get products needing reorder
	products, err := h.service.GetProductsNeedingReorder(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

// GetInventorySnapshot handles requests for inventory snapshot
func (h *AnalyticsHandler) GetInventorySnapshot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get inventory snapshot
	snapshot, err := h.service.GetInventorySnapshot(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(snapshot)
}

// RefreshOrganizationAnalytics handles requests to refresh analytics
func (h *AnalyticsHandler) RefreshOrganizationAnalytics(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Refresh analytics
	err := h.service.RefreshOrganizationAnalytics(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Analytics refreshed successfully"})
}

// GetInventoryDashboard handles requests for inventory dashboard data
func (h *AnalyticsHandler) GetInventoryDashboard(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get dashboard data
	dashboard, err := h.service.GetInventoryDashboard(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dashboard)
}

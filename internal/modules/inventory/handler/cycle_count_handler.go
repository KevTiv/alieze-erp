package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/service"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type CycleCountHandler struct {
	service *service.CycleCountService
}

func NewCycleCountHandler(service *service.CycleCountService) *CycleCountHandler {
	return &CycleCountHandler{
		service: service,
	}
}

func (h *CycleCountHandler) RegisterRoutes(router *httprouter.Router) {
	// Cycle count plan endpoints
	router.POST("/api/inventory/cycle-count/plans", h.CreateCycleCountPlan)
	router.GET("/api/inventory/cycle-count/plans/:plan_id", h.GetCycleCountPlan)
	router.GET("/api/inventory/cycle-count/plans", h.ListCycleCountPlans)
	router.PATCH("/api/inventory/cycle-count/plans/:plan_id/status", h.UpdateCycleCountPlanStatus)

	// Cycle count session endpoints
	router.POST("/api/inventory/cycle-count/sessions", h.CreateCycleCountSession)
	router.GET("/api/inventory/cycle-count/sessions/:session_id", h.GetCycleCountSession)
	router.GET("/api/inventory/cycle-count/sessions", h.ListCycleCountSessions)
	router.POST("/api/inventory/cycle-count/sessions/:session_id/complete", h.CompleteCycleCountSession)

	// Cycle count line endpoints
	router.POST("/api/inventory/cycle-count/sessions/:session_id/lines", h.AddCycleCountLine)
	router.GET("/api/inventory/cycle-count/lines/:line_id", h.GetCycleCountLine)
	router.GET("/api/inventory/cycle-count/sessions/:session_id/lines", h.ListCycleCountLines)
	router.POST("/api/inventory/cycle-count/lines/:line_id/verify", h.VerifyCycleCountLine)

	// Cycle count adjustment endpoints
	router.POST("/api/inventory/cycle-count/lines/:line_id/adjustments", h.CreateAdjustmentFromVariance)
	router.GET("/api/inventory/cycle-count/adjustments/:adjustment_id", h.GetCycleCountAdjustment)
	router.GET("/api/inventory/cycle-count/adjustments", h.ListCycleCountAdjustments)
	router.POST("/api/inventory/cycle-count/adjustments/:adjustment_id/approve", h.ApproveCycleCountAdjustment)

	// Cycle count accuracy endpoints
	router.POST("/api/inventory/cycle-count/accuracy/metrics", h.GetCycleCountAccuracyMetrics)
	router.POST("/api/inventory/cycle-count/products/needing", h.GetProductsNeedingCycleCount)
	router.GET("/api/inventory/cycle-count/accuracy/history", h.GetCycleCountAccuracyHistory)

	// Dashboard endpoint
	router.GET("/api/inventory/cycle-count/dashboard", h.GetCycleCountDashboard)
}

// CreateCycleCountPlan handles creation of cycle count plans
func (h *CycleCountHandler) CreateCycleCountPlan(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request types.CreateCycleCountPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Create plan
	plan, err := h.service.CreateCycleCountPlan(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(plan)
}

// GetCycleCountPlan retrieves a cycle count plan
func (h *CycleCountHandler) GetCycleCountPlan(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get plan ID from URL
	planID, err := uuid.Parse(ps.ByName("plan_id"))
	if err != nil {
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	// Get plan
	plan, err := h.service.GetCycleCountPlan(r.Context(), orgID, planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if plan == nil {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(plan)
}

// ListCycleCountPlans retrieves cycle count plans for an organization
func (h *CycleCountHandler) ListCycleCountPlans(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0
	status := r.URL.Query().Get("status")

	if r.URL.Query().Has("limit") {
		fmt.Sscanf(r.URL.Query().Get("limit"), "%d", &limit)
	}
	if r.URL.Query().Has("offset") {
		fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	// Get plans
	plans, err := h.service.ListCycleCountPlans(r.Context(), orgID, statusPtr, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(plans)
}

// UpdateCycleCountPlanStatus updates the status of a cycle count plan
func (h *CycleCountHandler) UpdateCycleCountPlanStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get plan ID from URL
	planID, err := uuid.Parse(ps.ByName("plan_id"))
	if err != nil {
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var request struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update status
	success, err := h.service.UpdateCycleCountPlanStatus(r.Context(), orgID, planID, request.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Plan status updated successfully",
		})
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Plan not found",
		})
	}
}

// CreateCycleCountSession creates a new cycle count session
func (h *CycleCountHandler) CreateCycleCountSession(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request types.CreateCycleCountSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Create session
	session, err := h.service.CreateCycleCountSession(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

// GetCycleCountSession retrieves a cycle count session
func (h *CycleCountHandler) GetCycleCountSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(ps.ByName("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Get session
	session, err := h.service.GetCycleCountSession(r.Context(), orgID, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(session)
}

// ListCycleCountSessions retrieves cycle count sessions for an organization
func (h *CycleCountHandler) ListCycleCountSessions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0
	status := r.URL.Query().Get("status")

	if r.URL.Query().Has("limit") {
		fmt.Sscanf(r.URL.Query().Get("limit"), "%d", &limit)
	}
	if r.URL.Query().Has("offset") {
		fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	// Get sessions
	sessions, err := h.service.ListCycleCountSessions(r.Context(), orgID, statusPtr, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sessions)
}

// CompleteCycleCountSession marks a session as completed
func (h *CycleCountHandler) CompleteCycleCountSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(ps.ByName("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Complete session
	success, err := h.service.CompleteCycleCountSession(r.Context(), orgID, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Session completed successfully",
		})
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Session not found or already completed",
		})
	}
}

// AddCycleCountLine adds a count line to a session
func (h *CycleCountHandler) AddCycleCountLine(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(ps.ByName("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var request types.AddCycleCountLineRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set session and organization IDs
	request.SessionID = sessionID
	request.OrganizationID = orgID

	// Add count line
	line, err := h.service.AddCycleCountLine(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(line)
}

// GetCycleCountLine retrieves a count line
func (h *CycleCountHandler) GetCycleCountLine(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get line ID from URL
	lineID, err := uuid.Parse(ps.ByName("line_id"))
	if err != nil {
		http.Error(w, "Invalid line ID", http.StatusBadRequest)
		return
	}

	// Get count line
	line, err := h.service.GetCycleCountLine(r.Context(), orgID, lineID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if line == nil {
		http.Error(w, "Count line not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(line)
}

// ListCycleCountLines retrieves count lines for a session
func (h *CycleCountHandler) ListCycleCountLines(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(ps.ByName("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Get count lines
	lines, err := h.service.ListCycleCountLines(r.Context(), orgID, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lines)
}

// VerifyCycleCountLine verifies a count line
func (h *CycleCountHandler) VerifyCycleCountLine(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get line ID from URL
	lineID, err := uuid.Parse(ps.ByName("line_id"))
	if err != nil {
		http.Error(w, "Invalid line ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var request types.VerifyCycleCountLineRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set line and organization IDs
	request.LineID = lineID
	request.OrganizationID = orgID

	// Verify count line
	success, err := h.service.VerifyCycleCountLine(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Count line verified successfully",
		})
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Count line not found",
		})
	}
}

// CreateAdjustmentFromVariance creates an adjustment from a count variance
func (h *CycleCountHandler) CreateAdjustmentFromVariance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get line ID from URL
	lineID, err := uuid.Parse(ps.ByName("line_id"))
	if err != nil {
		http.Error(w, "Invalid line ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var request types.CreateAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set line and organization IDs
	request.LineID = lineID
	request.OrganizationID = orgID

	// Create adjustment
	adjustment, err := h.service.CreateAdjustmentFromVariance(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(adjustment)
}

// GetCycleCountAdjustment retrieves an adjustment
func (h *CycleCountHandler) GetCycleCountAdjustment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get adjustment ID from URL
	adjustmentID, err := uuid.Parse(ps.ByName("adjustment_id"))
	if err != nil {
		http.Error(w, "Invalid adjustment ID", http.StatusBadRequest)
		return
	}

	// Get adjustment
	adjustment, err := h.service.GetCycleCountAdjustment(r.Context(), orgID, adjustmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if adjustment == nil {
		http.Error(w, "Adjustment not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adjustment)
}

// ListCycleCountAdjustments retrieves adjustments for an organization
func (h *CycleCountHandler) ListCycleCountAdjustments(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0
	status := r.URL.Query().Get("status")

	if r.URL.Query().Has("limit") {
		fmt.Sscanf(r.URL.Query().Get("limit"), "%d", &limit)
	}
	if r.URL.Query().Has("offset") {
		fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	// Get adjustments
	adjustments, err := h.service.ListCycleCountAdjustments(r.Context(), orgID, statusPtr, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adjustments)
}

// ApproveCycleCountAdjustment approves an adjustment
func (h *CycleCountHandler) ApproveCycleCountAdjustment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get adjustment ID from URL
	adjustmentID, err := uuid.Parse(ps.ByName("adjustment_id"))
	if err != nil {
		http.Error(w, "Invalid adjustment ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var request types.ApproveAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set adjustment and organization IDs
	request.AdjustmentID = adjustmentID
	request.OrganizationID = orgID

	// Approve adjustment
	success, err := h.service.ApproveCycleCountAdjustment(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Adjustment approved and stock updated successfully",
		})
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Adjustment not found or already processed",
		})
	}
}

// GetCycleCountAccuracyMetrics retrieves accuracy metrics
func (h *CycleCountHandler) GetCycleCountAccuracyMetrics(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request types.GetAccuracyMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Get metrics
	metrics, err := h.service.GetCycleCountAccuracyMetrics(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// GetProductsNeedingCycleCount retrieves products that need cycle counting
func (h *CycleCountHandler) GetProductsNeedingCycleCount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request types.GetProductsNeedingCountRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Get products
	products, err := h.service.GetProductsNeedingCycleCount(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

// GetCycleCountAccuracyHistory retrieves accuracy history
func (h *CycleCountHandler) GetCycleCountAccuracyHistory(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0
	productID := r.URL.Query().Get("product_id")

	if r.URL.Query().Has("limit") {
		fmt.Sscanf(r.URL.Query().Get("limit"), "%d", &limit)
	}
	if r.URL.Query().Has("offset") {
		fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	}

	var productIDPtr *uuid.UUID
	if productID != "" {
		parsedID, err := uuid.Parse(productID)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}
		productIDPtr = &parsedID
	}

	// Get history
	history, err := h.service.GetCycleCountAccuracyHistory(r.Context(), orgID, productIDPtr, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}

// GetCycleCountDashboard retrieves a comprehensive dashboard
func (h *CycleCountHandler) GetCycleCountDashboard(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get dashboard data
	dashboard, err := h.service.GetCycleCountDashboard(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dashboard)
}

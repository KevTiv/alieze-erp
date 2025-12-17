package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"alieze-erp/internal/modules/inventory/service"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// BatchOperationHandler handles HTTP requests for batch operations
type BatchOperationHandler struct {
	service service.BatchOperationService
}

func NewBatchOperationHandler(service service.BatchOperationService) *BatchOperationHandler {
	return &BatchOperationHandler{
		service: service,
	}
}

func (h *BatchOperationHandler) RegisterRoutes(router *chi.Mux) {
	// Batch Operation CRUD endpoints
	router.Route("/api/v1/inventory/batch-operations", func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Get("/", h.ListBatchOperations)
		r.Post("/", h.CreateBatchOperation)
		r.Get("/{id}", h.GetBatchOperation)
		r.Put("/{id}", h.UpdateBatchOperation)
		r.Delete("/{id}", h.DeleteBatchOperation)
	})

	// Batch Operation Item endpoints
	router.Route("/api/v1/inventory/batch-operations/{batchOperationID}/items", func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Get("/", h.ListBatchOperationItems)
		r.Post("/", h.CreateBatchOperationItem)
		r.Get("/{id}", h.GetBatchOperationItem)
		r.Put("/{id}", h.UpdateBatchOperationItem)
		r.Delete("/{id}", h.DeleteBatchOperationItem)
	})

	// Specialized Batch Creation endpoints
	router.Route("/api/v1/inventory/batch-operations", func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Post("/stock-adjustment", h.CreateStockAdjustmentBatch)
		r.Post("/stock-transfer", h.CreateStockTransferBatch)
		r.Post("/stock-count", h.CreateStockCountBatch)
		r.Post("/price-update", h.CreatePriceUpdateBatch)
		r.Post("/location-update", h.CreateLocationUpdateBatch)
		r.Post("/status-update", h.CreateStatusUpdateBatch)
	})

	// Processing and Statistics endpoints
	router.Route("/api/v1/inventory/batch-operations", func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Post("/{id}/process", h.ProcessBatchOperation)
		r.Get("/{id}/items", h.ListBatchOperationItems)
		r.Get("/{id}/items/{status}", h.ListBatchOperationItemsByStatus)
		r.Get("/statistics", h.GetBatchOperationStatistics)
	})

	// Filtering endpoints
	router.Route("/api/v1/inventory/batch-operations", func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Get("/by-status/{status}", h.ListBatchOperationsByStatus)
		r.Get("/by-type/{type}", h.ListBatchOperationsByType)
		r.Get("/by-product/{productID}", h.ListBatchOperationsByProduct)
	})
}

func (h *BatchOperationHandler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract organization ID from context (set by auth middleware)
		orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
		if !ok {
			http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
			return
		}

		// Extract user ID from context
		userID, ok := r.Context().Value("userID").(uuid.UUID)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}

		// Add organization ID and user ID to request context for service layer
		ctx := context.WithValue(r.Context(), "organization_id", orgID)
		ctx = context.WithValue(ctx, "user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CRUD Endpoints

func (h *BatchOperationHandler) CreateBatchOperation(w http.ResponseWriter, r *http.Request) {
	var operation types.BatchOperation
	if err := json.NewDecoder(r.Body).Decode(&operation); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	orgID := r.Context().Value("organization_id").(uuid.UUID)
	operation.OrganizationID = orgID

	created, err := h.service.CreateBatchOperation(r.Context(), operation)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create batch operation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *BatchOperationHandler) GetBatchOperation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid batch operation ID", http.StatusBadRequest)
		return
	}

	operation, err := h.service.GetBatchOperation(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get batch operation: %v", err), http.StatusInternalServerError)
		return
	}
	if operation == nil {
		http.Error(w, "Batch operation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(operation)
}

func (h *BatchOperationHandler) ListBatchOperations(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			limit = 100
		}
	}

	operations, err := h.service.ListBatchOperations(r.Context(), orgID, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list batch operations: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(operations)
}

func (h *BatchOperationHandler) UpdateBatchOperation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid batch operation ID", http.StatusBadRequest)
		return
	}

	var operation types.BatchOperation
	if err := json.NewDecoder(r.Body).Decode(&operation); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	operation.ID = id

	updated, err := h.service.UpdateBatchOperation(r.Context(), operation)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update batch operation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *BatchOperationHandler) DeleteBatchOperation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid batch operation ID", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteBatchOperation(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete batch operation: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Batch Operation Item Endpoints

func (h *BatchOperationHandler) CreateBatchOperationItem(w http.ResponseWriter, r *http.Request) {
	batchOperationIDStr := chi.URLParam(r, "batchOperationID")
	batchOperationID, err := uuid.Parse(batchOperationIDStr)
	if err != nil {
		http.Error(w, "Invalid batch operation ID", http.StatusBadRequest)
		return
	}

	var item types.BatchOperationItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	item.BatchOperationID = batchOperationID

	created, err := h.service.CreateBatchOperationItem(r.Context(), item)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create batch operation item: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *BatchOperationHandler) GetBatchOperationItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid batch operation item ID", http.StatusBadRequest)
		return
	}

	item, err := h.service.GetBatchOperationItem(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get batch operation item: %v", err), http.StatusInternalServerError)
		return
	}
	if item == nil {
		http.Error(w, "Batch operation item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *BatchOperationHandler) ListBatchOperationItems(w http.ResponseWriter, r *http.Request) {
	batchOperationIDStr := chi.URLParam(r, "batchOperationID")
	batchOperationID, err := uuid.Parse(batchOperationIDStr)
	if err != nil {
		http.Error(w, "Invalid batch operation ID", http.StatusBadRequest)
		return
	}

	items, err := h.service.ListBatchOperationItems(r.Context(), batchOperationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list batch operation items: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *BatchOperationHandler) UpdateBatchOperationItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid batch operation item ID", http.StatusBadRequest)
		return
	}

	var item types.BatchOperationItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	item.ID = id

	updated, err := h.service.UpdateBatchOperationItem(r.Context(), item)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update batch operation item: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *BatchOperationHandler) DeleteBatchOperationItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid batch operation item ID", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteBatchOperationItem(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete batch operation item: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Specialized Batch Creation Endpoints

func (h *BatchOperationHandler) CreateStockAdjustmentBatch(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	userID := r.Context().Value("user_id").(uuid.UUID)

	var request struct {
		Reference   string                          `json:"reference"`
		Description string                          `json:"description"`
		CompanyID   *uuid.UUID                      `json:"company_id,omitempty"`
		Items       []types.BatchOperationItem     `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	companyID := uuid.Nil
	if request.CompanyID != nil {
		companyID = *request.CompanyID
	}

	operation, err := h.service.CreateStockAdjustmentBatch(
		r.Context(),
		orgID,
		companyID,
		userID,
		request.Reference,
		request.Description,
		request.Items,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create stock adjustment batch: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(operation)
}

func (h *BatchOperationHandler) CreateStockTransferBatch(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	userID := r.Context().Value("user_id").(uuid.UUID)

	var request struct {
		Reference        string                          `json:"reference"`
		Description      string                          `json:"description"`
		CompanyID        *uuid.UUID                      `json:"company_id,omitempty"`
		SourceLocationID uuid.UUID                      `json:"source_location_id"`
		DestLocationID   uuid.UUID                      `json:"dest_location_id"`
		Items            []types.BatchOperationItem     `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	companyID := uuid.Nil
	if request.CompanyID != nil {
		companyID = *request.CompanyID
	}

	operation, err := h.service.CreateStockTransferBatch(
		r.Context(),
		orgID,
		companyID,
		userID,
		request.Reference,
		request.Description,
		request.SourceLocationID,
		request.DestLocationID,
		request.Items,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create stock transfer batch: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(operation)
}

func (h *BatchOperationHandler) CreateStockCountBatch(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	userID := r.Context().Value("user_id").(uuid.UUID)

	var request struct {
		Reference   string                          `json:"reference"`
		Description string                          `json:"description"`
		CompanyID   *uuid.UUID                      `json:"company_id,omitempty"`
		LocationID  uuid.UUID                      `json:"location_id"`
		Items       []types.BatchOperationItem     `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	companyID := uuid.Nil
	if request.CompanyID != nil {
		companyID = *request.CompanyID
	}

	operation, err := h.service.CreateStockCountBatch(
		r.Context(),
		orgID,
		companyID,
		userID,
		request.Reference,
		request.Description,
		request.LocationID,
		request.Items,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create stock count batch: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(operation)
}

func (h *BatchOperationHandler) CreatePriceUpdateBatch(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	userID := r.Context().Value("user_id").(uuid.UUID)

	var request struct {
		Reference  string                          `json:"reference"`
		Description string                          `json:"description"`
		CompanyID  *uuid.UUID                      `json:"company_id,omitempty"`
		CurrencyID uuid.UUID                      `json:"currency_id"`
		Items      []types.BatchOperationItem     `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	companyID := uuid.Nil
	if request.CompanyID != nil {
		companyID = *request.CompanyID
	}

	operation, err := h.service.CreatePriceUpdateBatch(
		r.Context(),
		orgID,
		companyID,
		userID,
		request.Reference,
		request.Description,
		request.CurrencyID,
		request.Items,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create price update batch: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(operation)
}

func (h *BatchOperationHandler) CreateLocationUpdateBatch(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	userID := r.Context().Value("user_id").(uuid.UUID)

	var request struct {
		Reference     string                          `json:"reference"`
		Description   string                          `json:"description"`
		CompanyID     *uuid.UUID                      `json:"company_id,omitempty"`
		NewLocationID uuid.UUID                      `json:"new_location_id"`
		Items         []types.BatchOperationItem     `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	companyID := uuid.Nil
	if request.CompanyID != nil {
		companyID = *request.CompanyID
	}

	operation, err := h.service.CreateLocationUpdateBatch(
		r.Context(),
		orgID,
		companyID,
		userID,
		request.Reference,
		request.Description,
		request.NewLocationID,
		request.Items,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create location update batch: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(operation)
}

func (h *BatchOperationHandler) CreateStatusUpdateBatch(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	userID := r.Context().Value("user_id").(uuid.UUID)

	var request struct {
		Reference  string                          `json:"reference"`
		Description string                          `json:"description"`
		CompanyID  *uuid.UUID                      `json:"company_id,omitempty"`
		NewStatus  string                          `json:"new_status"`
		Items      []types.BatchOperationItem     `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	companyID := uuid.Nil
	if request.CompanyID != nil {
		companyID = *request.CompanyID
	}

	operation, err := h.service.CreateStatusUpdateBatch(
		r.Context(),
		orgID,
		companyID,
		userID,
		request.Reference,
		request.Description,
		request.NewStatus,
		request.Items,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create status update batch: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(operation)
}

// Processing and Statistics Endpoints

func (h *BatchOperationHandler) ProcessBatchOperation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid batch operation ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	result, err := h.service.ProcessBatchOperation(r.Context(), id, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process batch operation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *BatchOperationHandler) ListBatchOperationItemsByStatus(w http.ResponseWriter, r *http.Request) {
	batchOperationIDStr := chi.URLParam(r, "id")
	batchOperationID, err := uuid.Parse(batchOperationIDStr)
	if err != nil {
		http.Error(w, "Invalid batch operation ID", http.StatusBadRequest)
		return
	}

	status := chi.URLParam(r, "status")

	items, err := h.service.ListBatchOperationItemsByStatus(r.Context(), batchOperationID, status)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list batch operation items by status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *BatchOperationHandler) GetBatchOperationStatistics(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)

	// Parse query parameters
	fromTimeStr := r.URL.Query().Get("from_time")
	toTimeStr := r.URL.Query().Get("to_time")
	operationTypeStr := r.URL.Query().Get("operation_type")

	var fromTime, toTime *time.Time
	var operationType *types.BatchOperationType

	if fromTimeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, fromTimeStr)
		if err != nil {
			http.Error(w, "Invalid from_time format. Use RFC3339", http.StatusBadRequest)
			return
		}
		fromTime = &parsedTime
	}

	if toTimeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, toTimeStr)
		if err != nil {
			http.Error(w, "Invalid to_time format. Use RFC3339", http.StatusBadRequest)
			return
		}
		toTime = &parsedTime
	}

	if operationTypeStr != "" {
		opType := types.BatchOperationType(operationTypeStr)
		operationType = &opType
	}

	stats, err := h.service.GetBatchOperationStatistics(r.Context(), orgID, fromTime, toTime, operationType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get batch operation statistics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Filtering Endpoints

func (h *BatchOperationHandler) ListBatchOperationsByStatus(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	status := chi.URLParam(r, "status")

	operations, err := h.service.ListBatchOperationsByStatus(r.Context(), orgID, status)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list batch operations by status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(operations)
}

func (h *BatchOperationHandler) ListBatchOperationsByType(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	operationTypeStr := chi.URLParam(r, "type")
	operationType := types.BatchOperationType(operationTypeStr)

	operations, err := h.service.ListBatchOperationsByType(r.Context(), orgID, operationType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list batch operations by type: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(operations)
}

func (h *BatchOperationHandler) ListBatchOperationsByProduct(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(uuid.UUID)
	productIDStr := chi.URLParam(r, "productID")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	operations, err := h.service.ListBatchOperationsByProduct(r.Context(), orgID, productID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list batch operations by product: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(operations)
}

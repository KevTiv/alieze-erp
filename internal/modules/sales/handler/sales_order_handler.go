package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"alieze-erp/internal/modules/sales/domain"
	"alieze-erp/internal/modules/sales/repository"
	"alieze-erp/internal/modules/sales/service"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type SalesOrderHandler struct {
	service *service.SalesOrderService
}

func NewSalesOrderHandler(service *service.SalesOrderService) *SalesOrderHandler {
	return &SalesOrderHandler{
		service: service,
	}
}

func (h *SalesOrderHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/sales/orders", h.CreateSalesOrder)
	router.GET("/api/sales/orders/:id", h.GetSalesOrder)
	router.GET("/api/sales/orders", h.ListSalesOrders)
	router.PUT("/api/sales/orders/:id", h.UpdateSalesOrder)
	router.DELETE("/api/sales/orders/:id", h.DeleteSalesOrder)
	router.POST("/api/sales/orders/:id/confirm", h.ConfirmSalesOrder)
	router.POST("/api/sales/orders/:id/cancel", h.CancelSalesOrder)
	router.GET("/api/sales/orders/customer/:customer_id", h.GetSalesOrdersByCustomer)
	router.GET("/api/sales/orders/status/:status", h.GetSalesOrdersByStatus)
}

func (h *SalesOrderHandler) CreateSalesOrder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req domain.SalesOrder
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdOrder, err := h.service.CreateSalesOrder(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdOrder)
}

func (h *SalesOrderHandler) GetSalesOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetSalesOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, "Sales order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *SalesOrderHandler) ListSalesOrders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	customerIDStr := r.URL.Query().Get("customer_id")
	statusStr := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	filters := repository.SalesOrderFilter{
		Limit:  10,
		Offset: 0,
	}

	if customerIDStr != "" {
		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			http.Error(w, "Invalid customer ID", http.StatusBadRequest)
			return
		}
		filters.CustomerID = &customerID
	}

	if statusStr != "" {
		status := domain.SalesOrderStatus(statusStr)
		filters.Status = &status
	}

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		filters.Limit = limit
	}

	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
		filters.Offset = offset
	}

	orders, err := h.service.ListSalesOrders(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *SalesOrderHandler) UpdateSalesOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req domain.SalesOrder
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = id

	updatedOrder, err := h.service.UpdateSalesOrder(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}

func (h *SalesOrderHandler) DeleteSalesOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteSalesOrder(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SalesOrderHandler) ConfirmSalesOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	confirmedOrder, err := h.service.ConfirmSalesOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(confirmedOrder)
}

func (h *SalesOrderHandler) CancelSalesOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	cancelledOrder, err := h.service.CancelSalesOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cancelledOrder)
}

func (h *SalesOrderHandler) GetSalesOrdersByCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	customerID, err := uuid.Parse(ps.ByName("customer_id"))
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	orders, err := h.service.GetSalesOrdersByCustomer(r.Context(), customerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *SalesOrderHandler) GetSalesOrdersByStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := domain.SalesOrderStatus(ps.ByName("status"))

	orders, err := h.service.GetSalesOrdersByStatus(r.Context(), status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

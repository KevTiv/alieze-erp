package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"alieze-erp/internal/modules/accounting/types"
	"alieze-erp/internal/modules/accounting/repository"
	"alieze-erp/internal/modules/accounting/service"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type InvoiceHandler struct {
	service *service.InvoiceService
}

func NewInvoiceHandler(service *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{
		service: service,
	}
}

func (h *InvoiceHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/accounting/invoices", h.CreateInvoice)
	router.GET("/api/accounting/invoices/:id", h.GetInvoice)
	router.GET("/api/accounting/invoices", h.ListInvoices)
	router.PUT("/api/accounting/invoices/:id", h.UpdateInvoice)
	router.DELETE("/api/accounting/invoices/:id", h.DeleteInvoice)
	router.POST("/api/accounting/invoices/:id/confirm", h.ConfirmInvoice)
	router.POST("/api/accounting/invoices/:id/cancel", h.CancelInvoice)
	router.POST("/api/accounting/invoices/:id/payments", h.RecordPayment)
	router.GET("/api/accounting/invoices/partner/:partner_id", h.GetInvoicesByPartner)
	router.GET("/api/accounting/invoices/status/:status", h.GetInvoicesByStatus)
	router.GET("/api/accounting/invoices/type/:type", h.GetInvoicesByType)
}

func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.Invoice
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdInvoice, err := h.service.CreateInvoice(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdInvoice)
}

func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.service.GetInvoice(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if invoice == nil {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoice)
}

func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	partnerIDStr := r.URL.Query().Get("partner_id")
	statusStr := r.URL.Query().Get("status")
	typeStr := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	filters := repository.InvoiceFilter{
		Limit:  10,
		Offset: 0,
	}

	if partnerIDStr != "" {
		partnerID, err := uuid.Parse(partnerIDStr)
		if err != nil {
			http.Error(w, "Invalid partner ID", http.StatusBadRequest)
			return
		}
		filters.PartnerID = &partnerID
	}

	if statusStr != "" {
		status := types.InvoiceStatus(statusStr)
		filters.Status = &status
	}

	if typeStr != "" {
		invoiceType := types.InvoiceType(typeStr)
		filters.Type = &invoiceType
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

	invoices, err := h.service.ListInvoices(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoices)
}

func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	var req types.Invoice
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = id

	updatedInvoice, err := h.service.UpdateInvoice(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedInvoice)
}

func (h *InvoiceHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteInvoice(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *InvoiceHandler) ConfirmInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	confirmedInvoice, err := h.service.ConfirmInvoice(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(confirmedInvoice)
}

func (h *InvoiceHandler) CancelInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	cancelledInvoice, err := h.service.CancelInvoice(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cancelledInvoice)
}

func (h *InvoiceHandler) RecordPayment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	invoiceID, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	var req types.Payment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedInvoice, err := h.service.RecordPayment(r.Context(), invoiceID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedInvoice)
}

func (h *InvoiceHandler) GetInvoicesByPartner(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	partnerID, err := uuid.Parse(ps.ByName("partner_id"))
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	invoices, err := h.service.GetInvoicesByPartner(r.Context(), partnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoices)
}

func (h *InvoiceHandler) GetInvoicesByStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := types.InvoiceStatus(ps.ByName("status"))

	invoices, err := h.service.GetInvoicesByStatus(r.Context(), status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoices)
}

func (h *InvoiceHandler) GetInvoicesByType(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	invoiceType := types.InvoiceType(ps.ByName("type"))

	invoices, err := h.service.GetInvoicesByType(r.Context(), invoiceType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoices)
}

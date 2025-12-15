package handler

import (
	"encoding/json"
	"net/http"

	"alieze-erp/internal/modules/accounting/domain"
	"alieze-erp/internal/modules/accounting/repository"
	"alieze-erp/internal/modules/accounting/service"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type PaymentHandler struct {
	service *service.PaymentService
}

func NewPaymentHandler(service *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: service,
	}
}

func (h *PaymentHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/accounting/payments", h.CreatePayment)
	router.GET("/api/accounting/payments/:id", h.GetPayment)
	router.GET("/api/accounting/payments", h.ListPayments)
	router.PUT("/api/accounting/payments/:id", h.UpdatePayment)
	router.DELETE("/api/accounting/payments/:id", h.DeletePayment)
	router.GET("/api/accounting/payments/invoice/:invoice_id", h.GetPaymentsByInvoice)
	router.GET("/api/accounting/payments/partner/:partner_id", h.GetPaymentsByPartner)
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req domain.Payment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdPayment, err := h.service.CreatePayment(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPayment)
}

func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}

	payment, err := h.service.GetPayment(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if payment == nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	invoiceIDStr := r.URL.Query().Get("invoice_id")
	partnerIDStr := r.URL.Query().Get("partner_id")

	filters := repository.PaymentFilter{
		Limit:  10,
		Offset: 0,
	}

	if invoiceIDStr != "" {
		invoiceID, err := uuid.Parse(invoiceIDStr)
		if err != nil {
			http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
			return
		}
		filters.InvoiceID = &invoiceID
	}

	if partnerIDStr != "" {
		partnerID, err := uuid.Parse(partnerIDStr)
		if err != nil {
			http.Error(w, "Invalid partner ID", http.StatusBadRequest)
			return
		}
		filters.PartnerID = &partnerID
	}

	payments, err := h.service.ListPayments(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payments)
}

func (h *PaymentHandler) UpdatePayment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}

	var req domain.Payment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = id

	updatedPayment, err := h.service.UpdatePayment(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPayment)
}

func (h *PaymentHandler) DeletePayment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePayment(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PaymentHandler) GetPaymentsByInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	invoiceID, err := uuid.Parse(ps.ByName("invoice_id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	payments, err := h.service.GetPaymentsByInvoice(r.Context(), invoiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payments)
}

func (h *PaymentHandler) GetPaymentsByPartner(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	partnerID, err := uuid.Parse(ps.ByName("partner_id"))
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	payments, err := h.service.GetPaymentsByPartner(r.Context(), partnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payments)
}

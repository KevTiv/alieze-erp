package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"alieze-erp/internal/modules/crm/types"
	"alieze-erp/internal/modules/crm/service"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type ContactHandler struct {
	service *service.ContactService
}

func NewContactHandler(service *service.ContactService) *ContactHandler {
	return &ContactHandler{
		service: service,
	}
}

func (h *ContactHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/crm/contacts", h.CreateContact)
	router.GET("/api/crm/contacts/:id", h.GetContact)
	router.GET("/api/crm/contacts", h.ListContacts)
	router.PUT("/api/crm/contacts/:id", h.UpdateContact)
	router.DELETE("/api/crm/contacts/:id", h.DeleteContact)
	router.GET("/api/crm/customers/:customer_id/contacts", h.GetContactsByCustomer)
	router.GET("/api/crm/vendors/:vendor_id/contacts", h.GetContactsByVendor)
}

func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.Contact
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdContact, err := h.service.CreateContact(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdContact)
}

func (h *ContactHandler) GetContact(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetContact(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if contact == nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contact)
}

func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")
	phone := r.URL.Query().Get("phone")
	isCustomerStr := r.URL.Query().Get("is_customer")
	isVendorStr := r.URL.Query().Get("is_vendor")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	filters := types.ContactFilter{
		Limit:  10,
		Offset: 0,
	}

	if name != "" {
		filters.Name = &name
	}

	if email != "" {
		filters.Email = &email
	}

	if phone != "" {
		filters.Phone = &phone
	}

	if isCustomerStr != "" {
		isCustomer, err := strconv.ParseBool(isCustomerStr)
		if err != nil {
			http.Error(w, "Invalid is_customer value", http.StatusBadRequest)
			return
		}
		filters.IsCustomer = &isCustomer
	}

	if isVendorStr != "" {
		isVendor, err := strconv.ParseBool(isVendorStr)
		if err != nil {
			http.Error(w, "Invalid is_vendor value", http.StatusBadRequest)
			return
		}
		filters.IsVendor = &isVendor
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

	contacts, _, err := h.service.ListContacts(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contacts)
}

func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req types.Contact
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = id

	updatedContact, err := h.service.UpdateContact(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedContact)
}

func (h *ContactHandler) DeleteContact(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteContact(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) GetContactsByCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := uuid.Parse(ps.ByName("customer_id"))
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	filters := types.ContactFilter{
		IsCustomer: func() *bool { b := true; return &b }(),
	}

	contacts, _, err := h.service.ListContacts(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contacts)
}

func (h *ContactHandler) GetContactsByVendor(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := uuid.Parse(ps.ByName("vendor_id"))
	if err != nil {
		http.Error(w, "Invalid vendor ID", http.StatusBadRequest)
		return
	}

	filters := types.ContactFilter{
		IsVendor: func() *bool { b := true; return &b }(),
	}

	contacts, _, err := h.service.ListContacts(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contacts)
}

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KevTiv/alieze-erp/internal/modules/crm/service"
	"github.com/KevTiv/alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type ContactHandler struct {
	service *service.ContactServiceV2
}

func NewContactHandler(service *service.ContactServiceV2) *ContactHandler {
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

	// ContactRelationship routes
	router.POST("/api/crm/contacts/:contactId/relationships", h.CreateContactRelationship)
	router.GET("/api/crm/contacts/:contactId/relationships", h.ListContactRelationships)
	router.POST("/api/crm/contacts/:contactId/segments", h.AddContactToSegments)
	router.GET("/api/crm/contacts/:contactId/score", h.GetContactScore)
}

func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req service.ContactRequest
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

// ContactRelationship handlers

func (h *ContactHandler) CreateContactRelationship(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	contactID, err := uuid.Parse(ps.ByName("contactId"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req types.ContactRelationshipCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create the relationship
	relationship, err := h.service.CreateRelationship(r.Context(), orgID, contactID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(relationship)
}

func (h *ContactHandler) ListContactRelationships(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	contactID, err := uuid.Parse(ps.ByName("contactId"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	relationshipType := r.URL.Query().Get("type")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	// Get relationships
	relationships, err := h.service.ListRelationships(r.Context(), orgID, contactID, relationshipType, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationships)
}

func (h *ContactHandler) AddContactToSegments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	contactID, err := uuid.Parse(ps.ByName("contactId"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req types.ContactSegmentationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Add to segments/tags
	err = h.service.AddToSegments(r.Context(), orgID, contactID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Contact added to segments successfully",
	})
}

func (h *ContactHandler) GetContactScore(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	contactID, err := uuid.Parse(ps.ByName("contactId"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	// Get contact score
	score, err := h.service.CalculateContactScore(r.Context(), orgID, contactID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(score)
}

func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req service.ContactUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedContact, err := h.service.UpdateContact(r.Context(), id, req)
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

	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// For now, use a general filter approach since specific customer contact methods aren't implemented yet
	// This would need to be enhanced with proper customer-contact relationship handling
	filters := types.ContactFilter{
		OrganizationID: orgID,
		IsCustomer:     func() *bool { b := true; return &b }(),
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

	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// For now, use a general filter approach since specific vendor contact methods aren't implemented yet
	// This would need to be enhanced with proper vendor-contact relationship handling
	filters := types.ContactFilter{
		OrganizationID: orgID,
		IsVendor:       func() *bool { b := true; return &b }(),
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

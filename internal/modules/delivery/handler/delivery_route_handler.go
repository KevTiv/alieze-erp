package handler

import (
	"encoding/json"
	"net/http"

	deliveryservice "alieze-erp/internal/modules/delivery/service"
	deliverytypes "alieze-erp/internal/modules/delivery/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type DeliveryRouteHandler struct {
	service *deliveryservice.DeliveryRouteService
}

func NewDeliveryRouteHandler(service *deliveryservice.DeliveryRouteService) *DeliveryRouteHandler {
	return &DeliveryRouteHandler{
		service: service,
	}
}

func (h *DeliveryRouteHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/delivery/routes", h.CreateDeliveryRoute)
	router.GET("/api/delivery/routes/:id", h.GetDeliveryRoute)
	router.GET("/api/delivery/routes", h.ListDeliveryRoutes)
	router.PUT("/api/delivery/routes/:id", h.UpdateDeliveryRoute)
	router.DELETE("/api/delivery/routes/:id", h.DeleteDeliveryRoute)
	router.POST("/api/delivery/routes/:id/start", h.StartRoute)
	router.POST("/api/delivery/routes/:id/complete", h.CompleteRoute)
	router.GET("/api/delivery/routes/organization/:org_id", h.ListDeliveryRoutesByOrganization)
	router.GET("/api/delivery/routes/organization/:org_id/status/:status", h.ListDeliveryRoutesByStatus)
}

func (h *DeliveryRouteHandler) CreateDeliveryRoute(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req deliverytypes.DeliveryRoute
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdRoute, err := h.service.CreateDeliveryRoute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdRoute)
}

func (h *DeliveryRouteHandler) GetDeliveryRoute(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	route, err := h.service.GetDeliveryRoute(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if route == nil {
		http.Error(w, "Delivery route not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(route)
}

func (h *DeliveryRouteHandler) ListDeliveryRoutes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	orgIDStr := r.URL.Query().Get("organization_id")
	statusStr := r.URL.Query().Get("status")

	if orgIDStr == "" {
		http.Error(w, "organization_id query parameter is required", http.StatusBadRequest)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	var statusFilter *deliverytypes.RouteStatus
	if statusStr != "" {
		status := deliverytypes.RouteStatus(statusStr)
		statusFilter = &status
	}

	routes, err := h.service.ListDeliveryRoutes(r.Context(), orgID, statusFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(routes)
}

func (h *DeliveryRouteHandler) UpdateDeliveryRoute(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	var req deliverytypes.DeliveryRoute
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure the ID from the URL matches the request body
	req.ID = id

	updatedRoute, err := h.service.UpdateDeliveryRoute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedRoute)
}

func (h *DeliveryRouteHandler) DeleteDeliveryRoute(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteDeliveryRoute(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeliveryRouteHandler) StartRoute(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	updatedRoute, err := h.service.StartRoute(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedRoute)
}

func (h *DeliveryRouteHandler) CompleteRoute(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	updatedRoute, err := h.service.CompleteRoute(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedRoute)
}

func (h *DeliveryRouteHandler) ListDeliveryRoutesByOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	orgID, err := uuid.Parse(ps.ByName("org_id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	routes, err := h.service.ListDeliveryRoutes(r.Context(), orgID, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(routes)
}

func (h *DeliveryRouteHandler) ListDeliveryRoutesByStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	orgID, err := uuid.Parse(ps.ByName("org_id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	status := deliverytypes.RouteStatus(ps.ByName("status"))
	routes, err := h.service.ListDeliveryRoutes(r.Context(), orgID, &status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(routes)
}

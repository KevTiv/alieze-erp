package handler

import (
	"encoding/json"
	"net/http"

	deliveryservice "github.com/KevTiv/alieze-erp/internal/modules/delivery/service"
	deliverytypes "github.com/KevTiv/alieze-erp/internal/modules/delivery/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type DeliveryVehicleHandler struct {
	service *deliveryservice.DeliveryVehicleService
}

func NewDeliveryVehicleHandler(service *deliveryservice.DeliveryVehicleService) *DeliveryVehicleHandler {
	return &DeliveryVehicleHandler{
		service: service,
	}
}

func (h *DeliveryVehicleHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/delivery/vehicles", h.CreateDeliveryVehicle)
	router.GET("/api/delivery/vehicles/:id", h.GetDeliveryVehicle)
	router.GET("/api/delivery/vehicles", h.ListDeliveryVehicles)
	router.PUT("/api/delivery/vehicles/:id", h.UpdateDeliveryVehicle)
	router.DELETE("/api/delivery/vehicles/:id", h.DeleteDeliveryVehicle)
	router.GET("/api/delivery/vehicles/organization/:org_id", h.ListDeliveryVehiclesByOrganization)
	router.GET("/api/delivery/vehicles/organization/:org_id/active", h.ListActiveDeliveryVehiclesByOrganization)
}

func (h *DeliveryVehicleHandler) CreateDeliveryVehicle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req deliverytypes.DeliveryVehicle
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdVehicle, err := h.service.CreateDeliveryVehicle(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdVehicle)
}

func (h *DeliveryVehicleHandler) GetDeliveryVehicle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	vehicle, err := h.service.GetDeliveryVehicle(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if vehicle == nil {
		http.Error(w, "Delivery vehicle not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicle)
}

func (h *DeliveryVehicleHandler) ListDeliveryVehicles(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	orgIDStr := r.URL.Query().Get("organization_id")
	activeStr := r.URL.Query().Get("active")

	if orgIDStr == "" {
		http.Error(w, "organization_id query parameter is required", http.StatusBadRequest)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	activeOnly := false
	if activeStr == "true" {
		activeOnly = true
	}

	vehicles, err := h.service.ListDeliveryVehicles(r.Context(), orgID, activeOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

func (h *DeliveryVehicleHandler) UpdateDeliveryVehicle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	var req deliverytypes.DeliveryVehicle
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure the ID from the URL matches the request body
	req.ID = id

	updatedVehicle, err := h.service.UpdateDeliveryVehicle(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedVehicle)
}

func (h *DeliveryVehicleHandler) DeleteDeliveryVehicle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteDeliveryVehicle(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeliveryVehicleHandler) ListDeliveryVehiclesByOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	orgID, err := uuid.Parse(ps.ByName("org_id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	vehicles, err := h.service.ListDeliveryVehicles(r.Context(), orgID, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

func (h *DeliveryVehicleHandler) ListActiveDeliveryVehiclesByOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	orgID, err := uuid.Parse(ps.ByName("org_id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	vehicles, err := h.service.ListDeliveryVehicles(r.Context(), orgID, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

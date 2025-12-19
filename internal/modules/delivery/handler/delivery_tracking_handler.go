package handler

import (
	"encoding/json"
	"net/http"

	deliveryservice "github.com/KevTiv/alieze-erp/internal/modules/delivery/service"
	deliverytypes "github.com/KevTiv/alieze-erp/internal/modules/delivery/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type DeliveryTrackingHandler struct {
	service *deliveryservice.DeliveryTrackingService
}

func NewDeliveryTrackingHandler(service *deliveryservice.DeliveryTrackingService) *DeliveryTrackingHandler {
	return &DeliveryTrackingHandler{
		service: service,
	}
}

func (h *DeliveryTrackingHandler) RegisterRoutes(router *httprouter.Router) {
	// Shipment endpoints
	router.POST("/api/delivery/shipments", h.CreateShipment)
	router.GET("/api/delivery/shipments/:id", h.GetShipment)
	router.GET("/api/delivery/shipments/picking/:picking_id", h.GetShipmentByPickingID)
	router.GET("/api/delivery/shipments/route/:route_id", h.ListShipmentsByRoute)
	router.PUT("/api/delivery/shipments/:id/status", h.UpdateShipmentStatus)

	// Tracking event endpoints
	router.POST("/api/delivery/tracking/events", h.CreateTrackingEvent)
	router.GET("/api/delivery/tracking/events/shipment/:shipment_id", h.GetTrackingEvents)
	router.GET("/api/delivery/tracking/events/shipment/:shipment_id/latest", h.GetLatestTrackingEvent)

	// Route position endpoints
	router.POST("/api/delivery/routes/:route_id/positions", h.CreateRoutePosition)
	router.GET("/api/delivery/routes/:route_id/positions", h.GetRoutePositions)
	router.GET("/api/delivery/routes/:route_id/positions/latest", h.GetLatestRoutePosition)

	// Route assignment endpoints
	router.POST("/api/delivery/routes/:route_id/assignments", h.CreateRouteAssignment)
	router.GET("/api/delivery/routes/:route_id/assignments", h.GetRouteAssignments)

	// Route stop endpoints
	router.POST("/api/delivery/routes/:route_id/stops", h.CreateRouteStop)
	router.GET("/api/delivery/routes/:route_id/stops", h.GetRouteStops)
	router.GET("/api/delivery/shipments/:shipment_id/stop", h.GetRouteStopByShipment)
	router.PUT("/api/delivery/stops/:stop_id/status", h.UpdateRouteStopStatus)
}

func (h *DeliveryTrackingHandler) CreateShipment(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req deliverytypes.DeliveryShipment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdShipment, err := h.service.CreateShipment(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdShipment)
}

func (h *DeliveryTrackingHandler) GetShipment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	shipment, err := h.service.GetShipment(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if shipment == nil {
		http.Error(w, "Shipment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(shipment)
}

func (h *DeliveryTrackingHandler) GetShipmentByPickingID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pickingID, err := uuid.Parse(ps.ByName("picking_id"))
	if err != nil {
		http.Error(w, "Invalid picking ID", http.StatusBadRequest)
		return
	}

	shipment, err := h.service.GetShipmentByPickingID(r.Context(), pickingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if shipment == nil {
		http.Error(w, "Shipment not found for picking", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(shipment)
}

func (h *DeliveryTrackingHandler) ListShipmentsByRoute(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routeID, err := uuid.Parse(ps.ByName("route_id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	shipments, err := h.service.ListShipmentsByRoute(r.Context(), routeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(shipments)
}

func (h *DeliveryTrackingHandler) UpdateShipmentStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	updatedShipment, err := h.service.UpdateShipmentStatus(r.Context(), id, deliverytypes.ShipmentStatus(req.Status))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedShipment)
}

func (h *DeliveryTrackingHandler) CreateTrackingEvent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req deliverytypes.DeliveryTrackingEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdEvent, err := h.service.CreateTrackingEvent(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdEvent)
}

func (h *DeliveryTrackingHandler) GetTrackingEvents(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	shipmentID, err := uuid.Parse(ps.ByName("shipment_id"))
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	events, err := h.service.GetTrackingEvents(r.Context(), shipmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}

func (h *DeliveryTrackingHandler) GetLatestTrackingEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	shipmentID, err := uuid.Parse(ps.ByName("shipment_id"))
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	event, err := h.service.GetLatestTrackingEvent(r.Context(), shipmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if event == nil {
		http.Error(w, "No tracking events found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}

func (h *DeliveryTrackingHandler) CreateRoutePosition(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routeID, err := uuid.Parse(ps.ByName("route_id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	var req deliverytypes.DeliveryRoutePosition
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the route ID from the URL
	req.RouteID = routeID

	createdPosition, err := h.service.CreateRoutePosition(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPosition)
}

func (h *DeliveryTrackingHandler) GetRoutePositions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routeID, err := uuid.Parse(ps.ByName("route_id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	positions, err := h.service.GetRoutePositions(r.Context(), routeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(positions)
}

func (h *DeliveryTrackingHandler) GetLatestRoutePosition(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routeID, err := uuid.Parse(ps.ByName("route_id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	position, err := h.service.GetLatestRoutePosition(r.Context(), routeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if position == nil {
		http.Error(w, "No route positions found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(position)
}

func (h *DeliveryTrackingHandler) CreateRouteAssignment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routeID, err := uuid.Parse(ps.ByName("route_id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	var req deliverytypes.DeliveryRouteAssignment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the route ID from the URL
	req.RouteID = routeID

	createdAssignment, err := h.service.CreateRouteAssignment(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAssignment)
}

func (h *DeliveryTrackingHandler) GetRouteAssignments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routeID, err := uuid.Parse(ps.ByName("route_id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	assignments, err := h.service.GetRouteAssignments(r.Context(), routeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(assignments)
}

func (h *DeliveryTrackingHandler) CreateRouteStop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routeID, err := uuid.Parse(ps.ByName("route_id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	var req deliverytypes.DeliveryRouteStop
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the route ID from the URL
	req.RouteID = routeID

	createdStop, err := h.service.CreateRouteStop(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdStop)
}

func (h *DeliveryTrackingHandler) GetRouteStops(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	routeID, err := uuid.Parse(ps.ByName("route_id"))
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	stops, err := h.service.GetRouteStops(r.Context(), routeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stops)
}

func (h *DeliveryTrackingHandler) GetRouteStopByShipment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	shipmentID, err := uuid.Parse(ps.ByName("shipment_id"))
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	stop, err := h.service.GetRouteStopByShipment(r.Context(), shipmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if stop == nil {
		http.Error(w, "Route stop not found for shipment", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stop)
}

func (h *DeliveryTrackingHandler) UpdateRouteStopStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	stopID, err := uuid.Parse(ps.ByName("stop_id"))
	if err != nil {
		http.Error(w, "Invalid stop ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	updatedStop, err := h.service.UpdateRouteStopStatus(r.Context(), stopID, deliverytypes.StopStatus(req.Status))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedStop)
}

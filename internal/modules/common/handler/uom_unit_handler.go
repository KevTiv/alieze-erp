package handler

import ("json"
	"net/http"

	"github.com/KevTiv/alieze-erp/internal/modules/common/service"
	"github.com/KevTiv/alieze-erp/internal/modules/common/types"
	"github.com/google/uuid"

	"github.com/julienschmidt/httprouter"
)

// UOMUnitHandler handles HTTP requests for UOM units
type UOMUnitHandler struct {
	service *service.UOMUnitService
}

func NewUOMUnitHandler(service *service.UOMUnitService) *UOMUnitHandler {
	return &UOMUnitHandler{service: service}
}

func (h *UOMUnitHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/v1/uom/units", h.CreateUOMUnit)
	router.GET("/api/v1/uom/units/:id", h.GetUOMUnit)
	router.GET("/api/v1/uom/units", h.ListUOMUnits)
	router.PUT("/api/v1/uom/units/:id", h.UpdateUOMUnit)
	router.DELETE("/api/v1/uom/units/:id", h.DeleteUOMUnit)
	router.GET("/api/v1/uom/categories/:category_id/units", h.ListUOMUnitsByCategory)
	router.POST("/api/v1/uom/units/convert", h.ConvertUnit)
}

func (h *UOMUnitHandler) CreateUOMUnit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var req types.UOMUnitCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	unit, err := h.service.Create(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}

func (h *UOMUnitHandler) GetUOMUnit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid UOM unit ID", http.StatusBadRequest)
		return
	}

	unit, err := h.service.GetByID(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if unit == nil {
		http.Error(w, "UOM unit not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}

func (h *UOMUnitHandler) ListUOMUnits(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	// Parse query parameters
	filter := types.UOMUnitFilter{}
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		parsedID, err := uuid.Parse(categoryID)
		if err != nil {
			http.Error(w, "invalid category_id", http.StatusBadRequest)
			return
		}
		filter.CategoryID = &parsedID
	}
	if active := r.URL.Query().Get("active"); active != "" {
		activeBool := active == "true"
		filter.Active = &activeBool
	}
	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		// Parse limit
		filter.Limit = 100 // default
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		// Parse offset
		filter.Offset = 0 // default
	}

	units, err := h.service.List(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

func (h *UOMUnitHandler) ListUOMUnitsByCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	categoryID, err := uuid.Parse(ps.ByName("category_id"))
	if err != nil {
		http.Error(w, "invalid category ID", http.StatusBadRequest)
		return
	}

	units, err := h.service.ListByCategory(ctx, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

func (h *UOMUnitHandler) UpdateUOMUnit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid UOM unit ID", http.StatusBadRequest)
		return
	}

	var req types.UOMUnitUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	unit, err := h.service.Update(ctx, id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}

func (h *UOMUnitHandler) DeleteUOMUnit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "invalid UOM unit ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UOMUnitHandler) ConvertUnit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var req struct {
		FromUnitID uuid.UUID `json:"from_unit_id"`
		ToUnitID   uuid.UUID `json:"to_unit_id"`
		Quantity   float64   `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.FromUnitID == uuid.Nil {
		http.Error(w, "from_unit_id is required", http.StatusBadRequest)
		return
	}
	if req.ToUnitID == uuid.Nil {
		http.Error(w, "to_unit_id is required", http.StatusBadRequest)
		return
	}

	converted, err := h.service.ConvertUnit(ctx, req.FromUnitID, req.ToUnitID, req.Quantity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"from_unit_id": req.FromUnitID,
		"to_unit_id":   req.ToUnitID,
		"quantity":     req.Quantity,
		"converted":    converted,
	})
}

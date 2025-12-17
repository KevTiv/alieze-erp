package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"alieze-erp/internal/modules/inventory/service"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type QualityControlHandler struct {
	qualityControlService *service.QualityControlService
	authService          auth.AuthService
}

func NewQualityControlHandler(
	qualityControlService *service.QualityControlService,
	authService auth.AuthService,
) *QualityControlHandler {
	return &QualityControlHandler{
		qualityControlService: qualityControlService,
		authService:          authService,
	}
}

func (h *QualityControlHandler) RegisterRoutes(router chi.Router) {
	router.Route("/quality-control", func(r chi.Router) {
		r.Use(h.authService.Middleware)

		// Inspection Management
		r.Route("/inspections", func(ri chi.Router) {
			ri.Post("/", h.CreateInspection)
			ri.Get("/", h.ListInspections)
			ri.Get("/{inspectionID}", h.GetInspection)
			ri.Put("/{inspectionID}", h.UpdateInspection)
			ri.Delete("/{inspectionID}", h.DeleteInspection)
			ri.Get("/product/{productID}", h.ListInspectionsByProduct)
			ri.Get("/status/{status}", h.ListInspectionsByStatus)
			ri.Post("/from-stock-move", h.CreateInspectionFromStockMove)
			ri.Post("/{inspectionID}/status", h.UpdateInspectionStatus)
			ri.Post("/{inspectionID}/complete", h.CompleteInspection)
			ri.Post("/{inspectionID}/disposition", h.HandleDisposition)
		})

		// Checklist Management
		r.Route("/checklists", func(rc chi.Router) {
			rc.Post("/", h.CreateChecklist)
			rc.Get("/", h.ListChecklists)
			rc.Get("/{checklistID}", h.GetChecklist)
			rc.Put("/{checklistID}", h.UpdateChecklist)
			rc.Delete("/{checklistID}", h.DeleteChecklist)
			rc.Get("/active", h.ListActiveChecklists)
			rc.Get("/product/{productID}", h.ListChecklistsByProduct)
		})

		// Checklist Item Management
		r.Route("/checklist-items", func(rci chi.Router) {
			rci.Post("/", h.CreateChecklistItem)
			rci.Get("/{itemID}", h.GetChecklistItem)
			rci.Put("/{itemID}", h.UpdateChecklistItem)
			rci.Delete("/{itemID}", h.DeleteChecklistItem)
			rci.Get("/checklist/{checklistID}", h.ListChecklistItems)
			rci.Get("/checklist/{checklistID}/active", h.ListActiveChecklistItems)
		})

		// Inspection Item Management
		r.Route("/inspection-items", func(rii chi.Router) {
			rii.Post("/", h.CreateInspectionItem)
			rii.Get("/{itemID}", h.GetInspectionItem)
			rii.Put("/{itemID}", h.UpdateInspectionItem)
			rii.Delete("/{itemID}", h.DeleteInspectionItem)
			rii.Get("/inspection/{inspectionID}", h.ListInspectionItems)
			rii.Post("/{itemID}/result", h.UpdateInspectionItemResult)
		})

		// Alert Management
		r.Route("/alerts", func(ra chi.Router) {
			ra.Post("/", h.CreateAlert)
			ra.Get("/", h.ListAlerts)
			ra.Get("/{alertID}", h.GetAlert)
			ra.Put("/{alertID}", h.UpdateAlert)
			ra.Delete("/{alertID}", h.DeleteAlert)
			ra.Get("/open", h.ListOpenAlerts)
			ra.Post("/{alertID}/status", h.UpdateAlertStatus)
			ra.Post("/from-inspection", h.CreateAlertFromInspection)
		})

		// Workflow Endpoints
		r.Post("/workflow/start", h.StartQualityControlWorkflow)
		r.Post("/workflow/process", h.ProcessQualityControlResult)

		// Dashboard and Statistics
		r.Get("/statistics", h.GetQualityControlStatistics)
		r.Get("/dashboard", h.GetQualityControlDashboard)
	})
}

// Inspection Handlers

func (h *QualityControlHandler) CreateInspection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var inspection domain.QualityControlInspection
	if err := json.NewDecoder(r.Body).Decode(&inspection); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Set organization from context
	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}
	inspection.OrganizationID = orgID

	createdInspection, err := h.qualityControlService.CreateInspection(ctx, inspection)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdInspection)
}

func (h *QualityControlHandler) GetInspection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inspectionIDStr := chi.URLParam(r, "inspectionID")
	inspectionID, err := uuid.Parse(inspectionIDStr)
	if err != nil {
		http.Error(w, "Invalid inspection ID", http.StatusBadRequest)
		return
	}

	inspection, err := h.qualityControlService.GetInspection(ctx, inspectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, inspection)
}

func (h *QualityControlHandler) ListInspections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	// Get limit parameter
	limit := 50
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	inspections, err := h.qualityControlService.ListInspections(ctx, orgID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, inspections)
}

func (h *QualityControlHandler) ListInspectionsByProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	productIDStr := chi.URLParam(r, "productID")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	inspections, err := h.qualityControlService.ListInspectionsByProduct(ctx, orgID, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, inspections)
}

func (h *QualityControlHandler) ListInspectionsByStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	status := chi.URLParam(r, "status")

	inspections, err := h.qualityControlService.ListInspectionsByStatus(ctx, orgID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, inspections)
}

func (h *QualityControlHandler) UpdateInspection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inspectionIDStr := chi.URLParam(r, "inspectionID")
	inspectionID, err := uuid.Parse(inspectionIDStr)
	if err != nil {
		http.Error(w, "Invalid inspection ID", http.StatusBadRequest)
		return
	}

	var inspection domain.QualityControlInspection
	if err := json.NewDecoder(r.Body).Decode(&inspection); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	inspection.ID = inspectionID

	updatedInspection, err := h.qualityControlService.UpdateInspection(ctx, inspection)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedInspection)
}

func (h *QualityControlHandler) DeleteInspection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inspectionIDStr := chi.URLParam(r, "inspectionID")
	inspectionID, err := uuid.Parse(inspectionIDStr)
	if err != nil {
		http.Error(w, "Invalid inspection ID", http.StatusBadRequest)
		return
	}

	err := h.qualityControlService.DeleteInspection(ctx, inspectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Quality control inspection deleted successfully"})
}

// Checklist Handlers

func (h *QualityControlHandler) CreateChecklist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var checklist domain.QualityControlChecklist
	if err := json.NewDecoder(r.Body).Decode(&checklist); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Set organization from context
	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}
	checklist.OrganizationID = orgID

	createdChecklist, err := h.qualityControlService.CreateChecklist(ctx, checklist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdChecklist)
}

func (h *QualityControlHandler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	checklistIDStr := chi.URLParam(r, "checklistID")
	checklistID, err := uuid.Parse(checklistIDStr)
	if err != nil {
		http.Error(w, "Invalid checklist ID", http.StatusBadRequest)
		return
	}

	checklist, err := h.qualityControlService.GetChecklist(ctx, checklistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, checklist)
}

func (h *QualityControlHandler) ListChecklists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	checklists, err := h.qualityControlService.ListChecklists(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, checklists)
}

func (h *QualityControlHandler) ListActiveChecklists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	checklists, err := h.qualityControlService.ListActiveChecklists(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, checklists)
}

func (h *QualityControlHandler) ListChecklistsByProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	productIDStr := chi.URLParam(r, "productID")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	checklists, err := h.qualityControlService.ListChecklistsByProduct(ctx, orgID, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, checklists)
}

func (h *QualityControlHandler) UpdateChecklist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	checklistIDStr := chi.URLParam(r, "checklistID")
	checklistID, err := uuid.Parse(checklistIDStr)
	if err != nil {
		http.Error(w, "Invalid checklist ID", http.StatusBadRequest)
		return
	}

	var checklist domain.QualityControlChecklist
	if err := json.NewDecoder(r.Body).Decode(&checklist); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	checklist.ID = checklistID

	updatedChecklist, err := h.qualityControlService.UpdateChecklist(ctx, checklist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedChecklist)
}

func (h *QualityControlHandler) DeleteChecklist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	checklistIDStr := chi.URLParam(r, "checklistID")
	checklistID, err := uuid.Parse(checklistIDStr)
	if err != nil {
		http.Error(w, "Invalid checklist ID", http.StatusBadRequest)
		return
	}

	err := h.qualityControlService.DeleteChecklist(ctx, checklistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Quality control checklist deleted successfully"})
}

// Checklist Item Handlers

func (h *QualityControlHandler) CreateChecklistItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var item domain.QualityChecklistItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdItem, err := h.qualityControlService.CreateChecklistItem(ctx, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdItem)
}

func (h *QualityControlHandler) GetChecklistItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.qualityControlService.GetChecklistItem(ctx, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, item)
}

func (h *QualityControlHandler) ListChecklistItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	checklistIDStr := chi.URLParam(r, "checklistID")
	checklistID, err := uuid.Parse(checklistIDStr)
	if err != nil {
		http.Error(w, "Invalid checklist ID", http.StatusBadRequest)
		return
	}

	items, err := h.qualityControlService.ListChecklistItems(ctx, checklistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, items)
}

func (h *QualityControlHandler) ListActiveChecklistItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	checklistIDStr := chi.URLParam(r, "checklistID")
	checklistID, err := uuid.Parse(checklistIDStr)
	if err != nil {
		http.Error(w, "Invalid checklist ID", http.StatusBadRequest)
		return
	}

	items, err := h.qualityControlService.ListActiveChecklistItems(ctx, checklistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, items)
}

func (h *QualityControlHandler) UpdateChecklistItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var item domain.QualityChecklistItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	item.ID = itemID

	updatedItem, err := h.qualityControlService.UpdateChecklistItem(ctx, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedItem)
}

func (h *QualityControlHandler) DeleteChecklistItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	err := h.qualityControlService.DeleteChecklistItem(ctx, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Quality control checklist item deleted successfully"})
}

// Inspection Item Handlers

func (h *QualityControlHandler) CreateInspectionItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var item domain.QualityControlInspectionItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdItem, err := h.qualityControlService.CreateInspectionItem(ctx, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdItem)
}

func (h *QualityControlHandler) GetInspectionItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.qualityControlService.GetInspectionItem(ctx, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, item)
}

func (h *QualityControlHandler) ListInspectionItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inspectionIDStr := chi.URLParam(r, "inspectionID")
	inspectionID, err := uuid.Parse(inspectionIDStr)
	if err != nil {
		http.Error(w, "Invalid inspection ID", http.StatusBadRequest)
		return
	}

	items, err := h.qualityControlService.ListInspectionItems(ctx, inspectionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, items)
}

func (h *QualityControlHandler) UpdateInspectionItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var item domain.QualityControlInspectionItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	item.ID = itemID

	updatedItem, err := h.qualityControlService.UpdateInspectionItem(ctx, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedItem)
}

func (h *QualityControlHandler) UpdateInspectionItemResult(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Result string  `json:"result"`
		Notes  *string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	notes := ""
	if request.Notes != nil {
		notes = *request.Notes
	}

	err = h.qualityControlService.UpdateInspectionItemResult(ctx, itemID, request.Result, notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Inspection item result updated successfully"})
}

func (h *QualityControlHandler) DeleteInspectionItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	err := h.qualityControlService.DeleteInspectionItem(ctx, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Quality control inspection item deleted successfully"})
}

// Alert Handlers

func (h *QualityControlHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var alert domain.QualityControlAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Set organization from context
	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}
	alert.OrganizationID = orgID

	createdAlert, err := h.qualityControlService.CreateAlert(ctx, alert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdAlert)
}

func (h *QualityControlHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alertIDStr := chi.URLParam(r, "alertID")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	alert, err := h.qualityControlService.GetAlert(ctx, alertID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, alert)
}

func (h *QualityControlHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	alerts, err := h.qualityControlService.ListAlerts(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, alerts)
}

func (h *QualityControlHandler) ListOpenAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	alerts, err := h.qualityControlService.ListOpenAlerts(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, alerts)
}

func (h *QualityControlHandler) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alertIDStr := chi.URLParam(r, "alertID")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	var alert domain.QualityControlAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	alert.ID = alertID

	updatedAlert, err := h.qualityControlService.UpdateAlert(ctx, alert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedAlert)
}

func (h *QualityControlHandler) UpdateAlertStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alertIDStr := chi.URLParam(r, "alertID")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Status     string  `json:"status"`
		ResolvedBy *string `json:"resolved_by,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var resolvedBy *uuid.UUID
	if request.ResolvedBy != nil {
		userID, err := uuid.Parse(*request.ResolvedBy)
		if err != nil {
			http.Error(w, "Invalid resolved_by ID", http.StatusBadRequest)
			return
		}
		resolvedBy = &userID
	}

	err := h.qualityControlService.UpdateAlertStatus(ctx, alertID, request.Status, resolvedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Alert status updated successfully"})
}

func (h *QualityControlHandler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alertIDStr := chi.URLParam(r, "alertID")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	err := h.qualityControlService.DeleteAlert(ctx, alertID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Quality control alert deleted successfully"})
}

// Business Logic Handlers

func (h *QualityControlHandler) CreateInspectionFromStockMove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		StockMoveID    uuid.UUID  `json:"stock_move_id"`
		InspectorID    uuid.UUID  `json:"inspector_id"`
		ChecklistID    *uuid.UUID `json:"checklist_id,omitempty"`
		InspectionMethod string    `json:"inspection_method,omitempty"`
		SampleSize     *int       `json:"sample_size,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	inspection, err := h.qualityControlService.CreateInspectionFromStockMove(
		ctx, request.StockMoveID, request.InspectorID, request.ChecklistID,
		request.InspectionMethod, request.SampleSize,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, inspection)
}

func (h *QualityControlHandler) UpdateInspectionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inspectionIDStr := chi.URLParam(r, "inspectionID")
	inspectionID, err := uuid.Parse(inspectionIDStr)
	if err != nil {
		http.Error(w, "Invalid inspection ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Status              string    `json:"status"`
		DefectType          *string   `json:"defect_type,omitempty"`
		DefectDescription   *string   `json:"defect_description,omitempty"`
		DefectQuantity      *float64  `json:"defect_quantity,omitempty"`
		QualityRating       *int      `json:"quality_rating,omitempty"`
		ComplianceNotes     *string   `json:"compliance_notes,omitempty"`
		Disposition         *string   `json:"disposition,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = h.qualityControlService.UpdateInspectionStatus(
		ctx, inspectionID, request.Status,
		getStringValue(request.DefectType), getStringValue(request.DefectDescription),
		request.DefectQuantity, request.QualityRating, getStringValue(request.ComplianceNotes),
		getStringValue(request.Disposition),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Inspection status updated successfully"})
}

func (h *QualityControlHandler) CompleteInspection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inspectionIDStr := chi.URLParam(r, "inspectionID")
	inspectionID, err := uuid.Parse(inspectionIDStr)
	if err != nil {
		http.Error(w, "Invalid inspection ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Status  string                          `json:"status"`
		Results []domain.QualityControlInspectionItem `json:"results"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = h.qualityControlService.CompleteInspection(ctx, inspectionID, request.Status, request.Results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Inspection completed successfully"})
}

func (h *QualityControlHandler) HandleDisposition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inspectionIDStr := chi.URLParam(r, "inspectionID")
	inspectionID, err := uuid.Parse(inspectionIDStr)
	if err != nil {
		http.Error(w, "Invalid inspection ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Disposition string `json:"disposition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	inspection, err := h.qualityControlService.HandleQualityControlDisposition(ctx, inspectionID, request.Disposition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, inspection)
}

func (h *QualityControlHandler) CreateAlertFromInspection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		InspectionID uuid.UUID `json:"inspection_id"`
		AlertType    string    `json:"alert_type"`
		Severity     string    `json:"severity"`
		Title        string    `json:"title"`
		Message      string    `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	alert, err := h.qualityControlService.CreateAlertFromInspection(
		ctx, request.InspectionID, request.AlertType, request.Severity, request.Title, request.Message,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, alert)
}

// Workflow Handlers

func (h *QualityControlHandler) StartQualityControlWorkflow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		StockMoveID uuid.UUID `json:"stock_move_id"`
		InspectorID uuid.UUID `json:"inspector_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	inspection, err := h.qualityControlService.StartQualityControlWorkflow(ctx, request.StockMoveID, request.InspectorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, inspection)
}

func (h *QualityControlHandler) ProcessQualityControlResult(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		InspectionID uuid.UUID                          `json:"inspection_id"`
		Results      []domain.QualityControlInspectionItem `json:"results"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	inspection, err := h.qualityControlService.ProcessQualityControlResult(ctx, request.InspectionID, request.Results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, inspection)
}

// Dashboard and Statistics Handlers

func (h *QualityControlHandler) GetQualityControlStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	// Parse date range parameters
	var fromTime, toTime *time.Time
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			http.Error(w, "Invalid from date format", http.StatusBadRequest)
			return
		}
		fromTime = &parsedTime
	}

	if toStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			http.Error(w, "Invalid to date format", http.StatusBadRequest)
			return
		}
		toTime = &parsedTime
	}

	// Parse product ID parameter
	var productID *uuid.UUID
	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr != "" {
		parsedID, err := uuid.Parse(productIDStr)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}
		productID = &parsedID
	}

	stats, err := h.qualityControlService.GetQualityControlStatistics(ctx, orgID, fromTime, toTime, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

func (h *QualityControlHandler) GetQualityControlDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization not found in context", http.StatusUnauthorized)
		return
	}

	dashboard, err := h.qualityControlService.GetQualityControlDashboard(ctx, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, dashboard)
}

// Helper functions

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

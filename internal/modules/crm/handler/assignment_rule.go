package handler

import (
	"encoding/json"
	"net/http"

	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/internal/modules/crm/types"
	auth "alieze-erp/pkg/auth"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// AssignmentRuleHandler handles HTTP requests for assignment rules
type AssignmentRuleHandler struct {
	service     *service.AssignmentRuleService
	authService auth.Service
}

// NewAssignmentRuleHandler creates a new assignment rule handler
func NewAssignmentRuleHandler(service *service.AssignmentRuleService, authService auth.Service) *AssignmentRuleHandler {
	return &AssignmentRuleHandler{
		service:     service,
		authService: authService,
	}
}

// respondWithJSON sends a JSON response with the given status code, message, and data
func respondWithJSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"data":    data,
	})
}

// respondWithError sends an error response with the given status code and message
func respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	if err != nil {
		http.Error(w, message+": "+err.Error(), statusCode)
	} else {
		http.Error(w, message, statusCode)
	}
}

// RegisterRoutes registers assignment rule routes
func (h *AssignmentRuleHandler) RegisterRoutes(router *httprouter.Router) {
	// Assignment Rule routes
	router.POST("/assignment-rules", h.CreateAssignmentRule)
	router.GET("/assignment-rules/:id", h.GetAssignmentRule)
	router.PUT("/assignment-rules/:id", h.UpdateAssignmentRule)
	router.DELETE("/assignment-rules/:id", h.DeleteAssignmentRule)
	router.GET("/assignment-rules", h.ListAssignmentRules)
	router.POST("/assignment-rules/:id/assign", h.AssignLead)
	router.GET("/assignment-rules/stats/users", h.GetAssignmentStatsByUser)
	router.GET("/assignment-rules/stats/rules", h.GetAssignmentRuleEffectiveness)

	// Territory routes
	router.POST("/territories", h.CreateTerritory)
	router.GET("/territories/:id", h.GetTerritory)
	router.PUT("/territories/:id", h.UpdateTerritory)
	router.DELETE("/territories/:id", h.DeleteTerritory)
	router.GET("/territories", h.ListTerritories)
}

// CreateAssignmentRule handles POST /assignment-rules
func (h *AssignmentRuleHandler) CreateAssignmentRule(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.CreateAssignmentRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	rule, err := h.service.CreateAssignmentRule(r.Context(), &req)
	if err != nil {
		http.Error(w, "Failed to create assignment rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Assignment rule created successfully",
		"data":    rule,
	})
}

// GetAssignmentRule handles GET /assignment-rules/:id
func (h *AssignmentRuleHandler) GetAssignmentRule(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid assignment rule ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	rule, err := h.service.GetAssignmentRule(r.Context(), id)
	if err != nil {
		http.Error(w, "Assignment rule not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Assignment rule retrieved successfully",
		"data":    rule,
	})
}

// UpdateAssignmentRule handles PUT /assignment-rules/:id
func (h *AssignmentRuleHandler) UpdateAssignmentRule(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid assignment rule ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	var req types.UpdateAssignmentRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	rule, err := h.service.UpdateAssignmentRule(r.Context(), id, &req)
	if err != nil {
		http.Error(w, "Failed to update assignment rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Assignment rule updated successfully",
		"data":    rule,
	})
}

// DeleteAssignmentRule handles DELETE /assignment-rules/:id
func (h *AssignmentRuleHandler) DeleteAssignmentRule(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid assignment rule ID", err)
		return
	}

	err = h.service.DeleteAssignmentRule(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete assignment rule", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Assignment rule deleted successfully", nil)
}

// ListAssignmentRules handles GET /assignment-rules
func (h *AssignmentRuleHandler) ListAssignmentRules(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, err := h.authService.GetOrganizationID(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	targetModel := r.URL.Query().Get("target_model")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	rules, err := h.service.ListAssignmentRules(r.Context(), orgID, targetModel, activeOnly)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list assignment rules", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Assignment rules retrieved successfully", rules)
}

// CreateTerritory handles POST /territories
func (h *AssignmentRuleHandler) CreateTerritory(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.CreateTerritoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	territory, err := h.service.CreateTerritory(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create territory", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, "Territory created successfully", territory)
}

// GetTerritory handles GET /territories/:id
func (h *AssignmentRuleHandler) GetTerritory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid territory ID", err)
		return
	}

	territory, err := h.service.GetTerritory(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Territory not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Territory retrieved successfully", territory)
}

// UpdateTerritory handles PUT /territories/:id
func (h *AssignmentRuleHandler) UpdateTerritory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid territory ID", err)
		return
	}

	var req types.UpdateTerritoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	territory, err := h.service.UpdateTerritory(r.Context(), id, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update territory", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Territory updated successfully", territory)
}

// DeleteTerritory handles DELETE /territories/:id
func (h *AssignmentRuleHandler) DeleteTerritory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid territory ID", err)
		return
	}

	err = h.service.DeleteTerritory(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete territory", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Territory deleted successfully", nil)
}

// ListTerritories handles GET /territories
func (h *AssignmentRuleHandler) ListTerritories(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, err := h.authService.GetOrganizationID(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	activeOnly := r.URL.Query().Get("active_only") == "true"

	territories, err := h.service.ListTerritories(r.Context(), orgID, activeOnly)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list territories", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Territories retrieved successfully", territories)
}

// AssignLead handles POST /assignment-rules/:id/assign
func (h *AssignmentRuleHandler) AssignLead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	leadID, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid lead ID", err)
		return
	}

	var conditions map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&conditions); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid conditions payload", err)
		return
	}

	result, err := h.service.AssignLead(r.Context(), leadID, conditions)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to assign lead", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Lead assigned successfully", result)
}

// GetAssignmentStatsByUser handles GET /assignment-rules/stats/users
func (h *AssignmentRuleHandler) GetAssignmentStatsByUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, err := h.authService.GetOrganizationID(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	targetModel := r.URL.Query().Get("target_model")

	stats, err := h.service.GetAssignmentStatsByUser(r.Context(), orgID, targetModel)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get assignment stats", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Assignment stats retrieved successfully", stats)
}

// GetAssignmentRuleEffectiveness handles GET /assignment-rules/stats/rules
func (h *AssignmentRuleHandler) GetAssignmentRuleEffectiveness(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, err := h.authService.GetOrganizationID(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	effectiveness, err := h.service.GetAssignmentRuleEffectiveness(r.Context(), orgID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get assignment rule effectiveness", err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Assignment rule effectiveness retrieved successfully", effectiveness)
}

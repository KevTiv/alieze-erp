package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/alieze-erp/internal/modules/crm/service"
	"github.com/alieze-erp/internal/modules/crm/types"
	"github.com/alieze-erp/pkg/auth"
	"github.com/alieze-erp/pkg/response"
)

// AssignmentRuleHandler handles HTTP requests for assignment rules
type AssignmentRuleHandler struct {
	service *service.AssignmentRuleService
	authService auth.Service
}

// NewAssignmentRuleHandler creates a new assignment rule handler
func NewAssignmentRuleHandler(service *service.AssignmentRuleService, authService auth.Service) *AssignmentRuleHandler {
	return &AssignmentRuleHandler{
		service:    service,
		authService: authService,
	}
}

// RegisterRoutes registers assignment rule routes
func (h *AssignmentRuleHandler) RegisterRoutes(router *gin.RouterGroup) {
	assignmentRoutes := router.Group("/assignment-rules") {
		assignmentRoutes.POST("", h.CreateAssignmentRule)
		assignmentRoutes.GET("/:id", h.GetAssignmentRule)
		assignmentRoutes.PUT("/:id", h.UpdateAssignmentRule)
		assignmentRoutes.DELETE("/:id", h.DeleteAssignmentRule)
		assignmentRoutes.GET("", h.ListAssignmentRules)
	}

	territoryRoutes := router.Group("/territories") {
		territoryRoutes.POST("", h.CreateTerritory)
		territoryRoutes.GET("/:id", h.GetTerritory)
		territoryRoutes.PUT("/:id", h.UpdateTerritory)
		territoryRoutes.DELETE("/:id", h.DeleteTerritory)
		territoryRoutes.GET("", h.ListTerritories)
	}

	assignmentRoutes.POST("/:id/assign", h.AssignLead)
	assignmentRoutes.GET("/stats/users", h.GetAssignmentStatsByUser)
	assignmentRoutes.GET("/stats/rules", h.GetAssignmentRuleEffectiveness)
}

// CreateAssignmentRule handles POST /assignment-rules
func (h *AssignmentRuleHandler) CreateAssignmentRule(c *gin.Context) {
	var req types.CreateAssignmentRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	rule, err := h.service.CreateAssignmentRule(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create assignment rule", err)
		return
	}

	response.Success(c, http.StatusCreated, "Assignment rule created successfully", rule)
}

// GetAssignmentRule handles GET /assignment-rules/:id
func (h *AssignmentRuleHandler) GetAssignmentRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid assignment rule ID", err)
		return
	}

	rule, err := h.service.GetAssignmentRule(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Assignment rule not found", err)
		return
	}

	response.Success(c, http.StatusOK, "Assignment rule retrieved successfully", rule)
}

// UpdateAssignmentRule handles PUT /assignment-rules/:id
func (h *AssignmentRuleHandler) UpdateAssignmentRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid assignment rule ID", err)
		return
	}

	var req types.UpdateAssignmentRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	rule, err := h.service.UpdateAssignmentRule(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update assignment rule", err)
		return
	}

	response.Success(c, http.StatusOK, "Assignment rule updated successfully", rule)
}

// DeleteAssignmentRule handles DELETE /assignment-rules/:id
func (h *AssignmentRuleHandler) DeleteAssignmentRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid assignment rule ID", err)
		return
	}

	err := h.service.DeleteAssignmentRule(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete assignment rule", err)
		return
	}

	response.Success(c, http.StatusOK, "Assignment rule deleted successfully", nil)
}

// ListAssignmentRules handles GET /assignment-rules
func (h *AssignmentRuleHandler) ListAssignmentRules(c *gin.Context) {
	// Get organization ID from context
	orgID, err := h.authService.GetOrganizationID(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	targetModel := c.Query("target_model")
	activeOnly := c.Query("active_only") == "true"

	rules, err := h.service.ListAssignmentRules(c.Request.Context(), orgID, targetModel, activeOnly)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list assignment rules", err)
		return
	}

	response.Success(c, http.StatusOK, "Assignment rules retrieved successfully", rules)
}

// CreateTerritory handles POST /territories
func (h *AssignmentRuleHandler) CreateTerritory(c *gin.Context) {
	var req types.CreateTerritoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	territory, err := h.service.CreateTerritory(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create territory", err)
		return
	}

	response.Success(c, http.StatusCreated, "Territory created successfully", territory)
}

// GetTerritory handles GET /territories/:id
func (h *AssignmentRuleHandler) GetTerritory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid territory ID", err)
		return
	}

	territory, err := h.service.GetTerritory(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Territory not found", err)
		return
	}

	response.Success(c, http.StatusOK, "Territory retrieved successfully", territory)
}

// UpdateTerritory handles PUT /territories/:id
func (h *AssignmentRuleHandler) UpdateTerritory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid territory ID", err)
		return
	}

	var req types.UpdateTerritoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	territory, err := h.service.UpdateTerritory(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update territory", err)
		return
	}

	response.Success(c, http.StatusOK, "Territory updated successfully", territory)
}

// DeleteTerritory handles DELETE /territories/:id
func (h *AssignmentRuleHandler) DeleteTerritory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid territory ID", err)
		return
	}

	err := h.service.DeleteTerritory(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete territory", err)
		return
	}

	response.Success(c, http.StatusOK, "Territory deleted successfully", nil)
}

// ListTerritories handles GET /territories
func (h *AssignmentRuleHandler) ListTerritories(c *gin.Context) {
	// Get organization ID from context
	orgID, err := h.authService.GetOrganizationID(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	activeOnly := c.Query("active_only") == "true"

	territories, err := h.service.ListTerritories(c.Request.Context(), orgID, activeOnly)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list territories", err)
		return
	}

	response.Success(c, http.StatusOK, "Territories retrieved successfully", territories)
}

// AssignLead handles POST /assignment-rules/:id/assign
func (h *AssignmentRuleHandler) AssignLead(c *gin.Context) {
	leadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid lead ID", err)
		return
	}

	var conditions map[string]interface{}
	if err := c.ShouldBindJSON(&conditions); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid conditions payload", err)
		return
	}

	result, err := h.service.AssignLead(c.Request.Context(), leadID, conditions)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to assign lead", err)
		return
	}

	response.Success(c, http.StatusOK, "Lead assigned successfully", result)
}

// GetAssignmentStatsByUser handles GET /assignment-rules/stats/users
func (h *AssignmentRuleHandler) GetAssignmentStatsByUser(c *gin.Context) {
	// Get organization ID from context
	orgID, err := h.authService.GetOrganizationID(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	targetModel := c.Query("target_model")

	stats, err := h.service.GetAssignmentStatsByUser(c.Request.Context(), orgID, targetModel)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get assignment stats", err)
		return
	}

	response.Success(c, http.StatusOK, "Assignment stats retrieved successfully", stats)
}

// GetAssignmentRuleEffectiveness handles GET /assignment-rules/stats/rules
func (h *AssignmentRuleHandler) GetAssignmentRuleEffectiveness(c *gin.Context) {
	// Get organization ID from context
	orgID, err := h.authService.GetOrganizationID(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	effectiveness, err := h.service.GetAssignmentRuleEffectiveness(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get assignment rule effectiveness", err)
		return
	}

	response.Success(c, http.StatusOK, "Assignment rule effectiveness retrieved successfully", effectiveness)
}

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"alieze-erp/internal/modules/crm/service"
	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// LeadHandler handles HTTP requests for leads
type LeadHandler struct {
	leadService *service.LeadService
}

// NewLeadHandler creates a new LeadHandler
func NewLeadHandler(leadService *service.LeadService) *LeadHandler {
	return &LeadHandler{
		leadService: leadService,
	}
}

// RegisterRoutes registers lead routes
func (h *LeadHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/v1/leads", h.CreateLead)
	router.GET("/api/v1/leads/:id", h.GetLead)
	router.PUT("/api/v1/leads/:id", h.UpdateLead)
	router.DELETE("/api/v1/leads/:id", h.DeleteLead)
	router.GET("/api/v1/leads", h.ListLeads)
	router.GET("/api/v1/leads/count", h.CountLeads)

	// Analytics endpoints
	router.GET("/api/v1/leads/pipeline-value", h.GetPipelineValue)
	router.GET("/api/v1/leads/pipeline-value-by-stage", h.GetPipelineValueByStage)
	router.GET("/api/v1/leads/conversion-rate", h.GetConversionRate)
	router.GET("/api/v1/leads/win-rate", h.GetWinRate)
	router.GET("/api/v1/leads/loss-rate", h.GetLossRate)
	router.GET("/api/v1/leads/average-conversion-time", h.GetAverageConversionTime)
	router.GET("/api/v1/leads/average-win-time", h.GetAverageWinTime)
	router.GET("/api/v1/leads/average-loss-time", h.GetAverageLossTime)
	router.GET("/api/v1/leads/average-expected-revenue", h.GetAverageExpectedRevenue)
	router.GET("/api/v1/leads/average-probability", h.GetAverageProbability)
	router.GET("/api/v1/leads/average-recurring-revenue", h.GetAverageRecurringRevenue)
	router.GET("/api/v1/leads/total-expected-revenue", h.GetTotalExpectedRevenue)
	router.GET("/api/v1/leads/total-recurring-revenue", h.GetTotalRecurringRevenue)

	// Filter endpoints
	router.GET("/api/v1/leads/by-contact/:contactID", h.GetLeadsByContact)
	router.GET("/api/v1/leads/by-user/:userID", h.GetLeadsByUser)
	router.GET("/api/v1/leads/by-team/:teamID", h.GetLeadsByTeam)
	router.GET("/api/v1/leads/by-stage/:stageID", h.GetLeadsByStage)
	router.GET("/api/v1/leads/by-source/:sourceID", h.GetLeadsBySource)
	router.GET("/api/v1/leads/by-campaign/:campaignID", h.GetLeadsByCampaign)
	router.GET("/api/v1/leads/by-medium/:mediumID", h.GetLeadsByMedium)
	router.GET("/api/v1/leads/by-tag/:tagID", h.GetLeadsByTag)
	router.GET("/api/v1/leads/by-company/:companyID", h.GetLeadsByCompany)
	router.GET("/api/v1/leads/by-country/:countryID", h.GetLeadsByCountry)
	router.GET("/api/v1/leads/by-state/:stateID", h.GetLeadsByState)
	router.GET("/api/v1/leads/by-city/:city", h.GetLeadsByCity)
	router.GET("/api/v1/leads/by-lost-reason/:lostReasonID", h.GetLeadsByLostReason)
	router.GET("/api/v1/leads/by-created-by/:createdBy", h.GetLeadsByCreatedBy)
	router.GET("/api/v1/leads/by-updated-by/:updatedBy", h.GetLeadsByUpdatedBy)
	router.GET("/api/v1/leads/by-color/:color", h.GetLeadsByColor)
	router.GET("/api/v1/leads/overdue", h.GetOverdueLeads)
	router.GET("/api/v1/leads/high-value", h.GetHighValueLeads)
	router.GET("/api/v1/leads/recent", h.GetRecentLeads)
	router.GET("/api/v1/leads/by-status/:status", h.GetLeadsByStatus)
	router.GET("/api/v1/leads/by-priority/:priority", h.GetLeadsByPriority)
	router.GET("/api/v1/leads/by-type/:leadType", h.GetLeadsByType)
	router.GET("/api/v1/leads/by-won-status/:wonStatus", h.GetLeadsByWonStatus)
	router.GET("/api/v1/leads/by-active-status/:active", h.GetLeadsByActiveStatus)

	// Count endpoints
	router.GET("/api/v1/leads/count-by-stage", h.CountLeadsByStage)
	router.GET("/api/v1/leads/count-by-priority", h.CountLeadsByPriority)
	router.GET("/api/v1/leads/count-by-type", h.CountLeadsByType)
	router.GET("/api/v1/leads/count-by-source", h.CountLeadsBySource)
	router.GET("/api/v1/leads/count-by-medium", h.CountLeadsByMedium)
	router.GET("/api/v1/leads/count-by-campaign", h.CountLeadsByCampaign)
	router.GET("/api/v1/leads/count-by-team", h.CountLeadsByTeam)
	router.GET("/api/v1/leads/count-by-user", h.CountLeadsByUser)
	router.GET("/api/v1/leads/count-by-lost-reason", h.CountLeadsByLostReason)
	router.GET("/api/v1/leads/count-by-won-status", h.CountLeadsByWonStatus)
	router.GET("/api/v1/leads/count-by-country", h.CountLeadsByCountry)
	router.GET("/api/v1/leads/count-by-state", h.CountLeadsByState)
	router.GET("/api/v1/leads/count-by-city", h.CountLeadsByCity)
}

// CreateLead handles lead creation
func (h *LeadHandler) CreateLead(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	var req types.LeadEnhancedCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	lead, err := h.leadService.CreateLead(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

// GetLead handles lead retrieval
func (h *LeadHandler) GetLead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid lead ID", http.StatusBadRequest)
		return
	}

	lead, err := h.leadService.GetLead(r.Context(), orgID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

// UpdateLead handles lead updates
func (h *LeadHandler) UpdateLead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid lead ID", http.StatusBadRequest)
		return
	}

	var req types.LeadEnhancedUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	lead, err := h.leadService.UpdateLead(r.Context(), orgID, id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

// DeleteLead handles lead deletion
func (h *LeadHandler) DeleteLead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid lead ID", http.StatusBadRequest)
		return
	}

	if err := h.leadService.DeleteLead(r.Context(), orgID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListLeads handles lead listing
func (h *LeadHandler) ListLeads(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := types.LeadEnhancedFilter{}

	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}
	if email := r.URL.Query().Get("email"); email != "" {
		filter.Email = &email
	}
	if phone := r.URL.Query().Get("phone"); phone != "" {
		filter.Phone = &phone
	}
	if contactName := r.URL.Query().Get("contact_name"); contactName != "" {
		filter.ContactName = &contactName
	}
	if mobile := r.URL.Query().Get("mobile"); mobile != "" {
		filter.Mobile = &mobile
	}

	// Parse UUID parameters
	if companyID := r.URL.Query().Get("company_id"); companyID != "" {
		parsedID, err := uuid.Parse(companyID)
		if err == nil {
			filter.CompanyID = &parsedID
		}
	}
	if contactID := r.URL.Query().Get("contact_id"); contactID != "" {
		parsedID, err := uuid.Parse(contactID)
		if err == nil {
			filter.ContactID = &parsedID
		}
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		parsedID, err := uuid.Parse(userID)
		if err == nil {
			filter.UserID = &parsedID
		}
	}
	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		parsedID, err := uuid.Parse(teamID)
		if err == nil {
			filter.TeamID = &parsedID
		}
	}
	if stageID := r.URL.Query().Get("stage_id"); stageID != "" {
		parsedID, err := uuid.Parse(stageID)
		if err == nil {
			filter.StageID = &parsedID
		}
	}
	if sourceID := r.URL.Query().Get("source_id"); sourceID != "" {
		parsedID, err := uuid.Parse(sourceID)
		if err == nil {
			filter.SourceID = &parsedID
		}
	}
	if mediumID := r.URL.Query().Get("medium_id"); mediumID != "" {
		parsedID, err := uuid.Parse(mediumID)
		if err == nil {
			filter.MediumID = &parsedID
		}
	}
	if campaignID := r.URL.Query().Get("campaign_id"); campaignID != "" {
		parsedID, err := uuid.Parse(campaignID)
		if err == nil {
			filter.CampaignID = &parsedID
		}
	}
	if lostReasonID := r.URL.Query().Get("lost_reason_id"); lostReasonID != "" {
		parsedID, err := uuid.Parse(lostReasonID)
		if err == nil {
			filter.LostReasonID = &parsedID
		}
	}
	if countryID := r.URL.Query().Get("country_id"); countryID != "" {
		parsedID, err := uuid.Parse(countryID)
		if err == nil {
			filter.CountryID = &parsedID
		}
	}
	if stateID := r.URL.Query().Get("state_id"); stateID != "" {
		parsedID, err := uuid.Parse(stateID)
		if err == nil {
			filter.StateID = &parsedID
		}
	}

	// Parse enum parameters
	if leadType := r.URL.Query().Get("lead_type"); leadType != "" {
		typedLeadType := types.LeadType(leadType)
		filter.LeadType = &typedLeadType
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		typedPriority := types.LeadPriority(priority)
		filter.Priority = &typedPriority
	}
	if wonStatus := r.URL.Query().Get("won_status"); wonStatus != "" {
		typedWonStatus := types.LeadWonStatus(wonStatus)
		filter.WonStatus = &typedWonStatus
	}

	// Parse numeric parameters
	if expectedRevenueMin := r.URL.Query().Get("expected_revenue_min"); expectedRevenueMin != "" {
		if val, err := strconv.ParseFloat(expectedRevenueMin, 64); err == nil {
			filter.ExpectedRevenueMin = &val
		}
	}
	if expectedRevenueMax := r.URL.Query().Get("expected_revenue_max"); expectedRevenueMax != "" {
		if val, err := strconv.ParseFloat(expectedRevenueMax, 64); err == nil {
			filter.ExpectedRevenueMax = &val
		}
	}
	if probabilityMin := r.URL.Query().Get("probability_min"); probabilityMin != "" {
		if val, err := strconv.Atoi(probabilityMin); err == nil {
			filter.ProbabilityMin = &val
		}
	}
	if probabilityMax := r.URL.Query().Get("probability_max"); probabilityMax != "" {
		if val, err := strconv.Atoi(probabilityMax); err == nil {
			filter.ProbabilityMax = &val
		}
	}

	// Parse boolean parameters
	if active := r.URL.Query().Get("active"); active != "" {
		if val, err := strconv.ParseBool(active); err == nil {
			filter.Active = &val
		}
	}

	// Parse pagination parameters
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			filter.Limit = val
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			filter.Offset = val
		}
	}

	leads, err := h.leadService.ListLeads(r.Context(), orgID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// CountLeads handles lead counting
func (h *LeadHandler) CountLeads(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters (similar to ListLeads)
	filter := types.LeadEnhancedFilter{}

	// Parse string parameters
	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}
	if email := r.URL.Query().Get("email"); email != "" {
		filter.Email = &email
	}
	if phone := r.URL.Query().Get("phone"); phone != "" {
		filter.Phone = &phone
	}
	if contactName := r.URL.Query().Get("contact_name"); contactName != "" {
		filter.ContactName = &contactName
	}
	if mobile := r.URL.Query().Get("mobile"); mobile != "" {
		filter.Mobile = &mobile
	}

	// Parse UUID parameters
	if companyID := r.URL.Query().Get("company_id"); companyID != "" {
		parsedID, err := uuid.Parse(companyID)
		if err == nil {
			filter.CompanyID = &parsedID
		}
	}
	if contactID := r.URL.Query().Get("contact_id"); contactID != "" {
		parsedID, err := uuid.Parse(contactID)
		if err == nil {
			filter.ContactID = &parsedID
		}
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		parsedID, err := uuid.Parse(userID)
		if err == nil {
			filter.UserID = &parsedID
		}
	}
	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		parsedID, err := uuid.Parse(teamID)
		if err == nil {
			filter.TeamID = &parsedID
		}
	}
	if stageID := r.URL.Query().Get("stage_id"); stageID != "" {
		parsedID, err := uuid.Parse(stageID)
		if err == nil {
			filter.StageID = &parsedID
		}
	}
	if sourceID := r.URL.Query().Get("source_id"); sourceID != "" {
		parsedID, err := uuid.Parse(sourceID)
		if err == nil {
			filter.SourceID = &parsedID
		}
	}
	if mediumID := r.URL.Query().Get("medium_id"); mediumID != "" {
		parsedID, err := uuid.Parse(mediumID)
		if err == nil {
			filter.MediumID = &parsedID
		}
	}
	if campaignID := r.URL.Query().Get("campaign_id"); campaignID != "" {
		parsedID, err := uuid.Parse(campaignID)
		if err == nil {
			filter.CampaignID = &parsedID
		}
	}
	if lostReasonID := r.URL.Query().Get("lost_reason_id"); lostReasonID != "" {
		parsedID, err := uuid.Parse(lostReasonID)
		if err == nil {
			filter.LostReasonID = &parsedID
		}
	}
	if countryID := r.URL.Query().Get("country_id"); countryID != "" {
		parsedID, err := uuid.Parse(countryID)
		if err == nil {
			filter.CountryID = &parsedID
		}
	}
	if stateID := r.URL.Query().Get("state_id"); stateID != "" {
		parsedID, err := uuid.Parse(stateID)
		if err == nil {
			filter.StateID = &parsedID
		}
	}

	// Parse enum parameters
	if leadType := r.URL.Query().Get("lead_type"); leadType != "" {
		typedLeadType := types.LeadType(leadType)
		filter.LeadType = &typedLeadType
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		typedPriority := types.LeadPriority(priority)
		filter.Priority = &typedPriority
	}
	if wonStatus := r.URL.Query().Get("won_status"); wonStatus != "" {
		typedWonStatus := types.LeadWonStatus(wonStatus)
		filter.WonStatus = &typedWonStatus
	}

	// Parse numeric parameters
	if expectedRevenueMin := r.URL.Query().Get("expected_revenue_min"); expectedRevenueMin != "" {
		if val, err := strconv.ParseFloat(expectedRevenueMin, 64); err == nil {
			filter.ExpectedRevenueMin = &val
		}
	}
	if expectedRevenueMax := r.URL.Query().Get("expected_revenue_max"); expectedRevenueMax != "" {
		if val, err := strconv.ParseFloat(expectedRevenueMax, 64); err == nil {
			filter.ExpectedRevenueMax = &val
		}
	}
	if probabilityMin := r.URL.Query().Get("probability_min"); probabilityMin != "" {
		if val, err := strconv.Atoi(probabilityMin); err == nil {
			filter.ProbabilityMin = &val
		}
	}
	if probabilityMax := r.URL.Query().Get("probability_max"); probabilityMax != "" {
		if val, err := strconv.Atoi(probabilityMax); err == nil {
			filter.ProbabilityMax = &val
		}
	}

	// Parse boolean parameters
	if active := r.URL.Query().Get("active"); active != "" {
		if val, err := strconv.ParseBool(active); err == nil {
			filter.Active = &val
		}
	}

	count, err := h.leadService.CountLeads(r.Context(), orgID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// GetPipelineValue handles pipeline value retrieval
func (h *LeadHandler) GetPipelineValue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	pipelineValue, err := h.leadService.GetLeadPipelineValue(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"pipeline_value": pipelineValue})
}

// GetPipelineValueByStage handles pipeline value by stage retrieval
func (h *LeadHandler) GetPipelineValueByStage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	pipelineValueByStage, err := h.leadService.GetLeadPipelineValueByStage(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pipelineValueByStage)
}

// GetConversionRate handles conversion rate retrieval
func (h *LeadHandler) GetConversionRate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	conversionRate, err := h.leadService.GetLeadConversionRate(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"conversion_rate": conversionRate})
}

// GetWinRate handles win rate retrieval
func (h *LeadHandler) GetWinRate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	winRate, err := h.leadService.GetLeadWinRate(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"win_rate": winRate})
}

// GetLossRate handles loss rate retrieval
func (h *LeadHandler) GetLossRate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	lossRate, err := h.leadService.GetLeadLossRate(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"loss_rate": lossRate})
}

// GetAverageConversionTime handles average conversion time retrieval
func (h *LeadHandler) GetAverageConversionTime(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	avgConversionTime, err := h.leadService.GetLeadAverageConversionTime(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"average_conversion_time": avgConversionTime})
}

// GetAverageWinTime handles average win time retrieval
func (h *LeadHandler) GetAverageWinTime(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	avgWinTime, err := h.leadService.GetLeadAverageWinTime(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"average_win_time": avgWinTime})
}

// GetAverageLossTime handles average loss time retrieval
func (h *LeadHandler) GetAverageLossTime(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	avgLossTime, err := h.leadService.GetLeadAverageLossTime(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"average_loss_time": avgLossTime})
}

// GetAverageExpectedRevenue handles average expected revenue retrieval
func (h *LeadHandler) GetAverageExpectedRevenue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	avgExpectedRevenue, err := h.leadService.GetLeadAverageExpectedRevenue(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"average_expected_revenue": avgExpectedRevenue})
}

// GetAverageProbability handles average probability retrieval
func (h *LeadHandler) GetAverageProbability(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	avgProbability, err := h.leadService.GetLeadAverageProbability(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"average_probability": avgProbability})
}

// GetAverageRecurringRevenue handles average recurring revenue retrieval
func (h *LeadHandler) GetAverageRecurringRevenue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	avgRecurringRevenue, err := h.leadService.GetLeadAverageRecurringRevenue(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"average_recurring_revenue": avgRecurringRevenue})
}

// GetTotalExpectedRevenue handles total expected revenue retrieval
func (h *LeadHandler) GetTotalExpectedRevenue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	totalExpectedRevenue, err := h.leadService.GetLeadTotalExpectedRevenue(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"total_expected_revenue": totalExpectedRevenue})
}

// GetTotalRecurringRevenue handles total recurring revenue retrieval
func (h *LeadHandler) GetTotalRecurringRevenue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	totalRecurringRevenue, err := h.leadService.GetLeadTotalRecurringRevenue(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"total_recurring_revenue": totalRecurringRevenue})
}

// GetLeadsByContact handles leads by contact retrieval
func (h *LeadHandler) GetLeadsByContact(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	contactID, err := uuid.Parse(ps.ByName("contactID"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByContact(r.Context(), orgID, contactID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByUser handles leads by user retrieval
func (h *LeadHandler) GetLeadsByUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(ps.ByName("userID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByUser(r.Context(), orgID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByTeam handles leads by team retrieval
func (h *LeadHandler) GetLeadsByTeam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	teamID, err := uuid.Parse(ps.ByName("teamID"))
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByTeam(r.Context(), orgID, teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByStage handles leads by stage retrieval
func (h *LeadHandler) GetLeadsByStage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	stageID, err := uuid.Parse(ps.ByName("stageID"))
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByStage(r.Context(), orgID, stageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsBySource handles leads by source retrieval
func (h *LeadHandler) GetLeadsBySource(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	sourceID, err := uuid.Parse(ps.ByName("sourceID"))
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsBySource(r.Context(), orgID, sourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByCampaign handles leads by campaign retrieval
func (h *LeadHandler) GetLeadsByCampaign(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	campaignID, err := uuid.Parse(ps.ByName("campaignID"))
	if err != nil {
		http.Error(w, "Invalid campaign ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByCampaign(r.Context(), orgID, campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByMedium handles leads by medium retrieval
func (h *LeadHandler) GetLeadsByMedium(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	mediumID, err := uuid.Parse(ps.ByName("mediumID"))
	if err != nil {
		http.Error(w, "Invalid medium ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByMedium(r.Context(), orgID, mediumID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByTag handles leads by tag retrieval
func (h *LeadHandler) GetLeadsByTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	tagID, err := uuid.Parse(ps.ByName("tagID"))
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByTag(r.Context(), orgID, tagID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByCompany handles leads by company retrieval
func (h *LeadHandler) GetLeadsByCompany(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	companyID, err := uuid.Parse(ps.ByName("companyID"))
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByCompany(r.Context(), orgID, companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByCountry handles leads by country retrieval
func (h *LeadHandler) GetLeadsByCountry(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	countryID, err := uuid.Parse(ps.ByName("countryID"))
	if err != nil {
		http.Error(w, "Invalid country ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByCountry(r.Context(), orgID, countryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByState handles leads by state retrieval
func (h *LeadHandler) GetLeadsByState(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	stateID, err := uuid.Parse(ps.ByName("stateID"))
	if err != nil {
		http.Error(w, "Invalid state ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByState(r.Context(), orgID, stateID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByCity handles leads by city retrieval
func (h *LeadHandler) GetLeadsByCity(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	city := ps.ByName("city")
	if city == "" {
		http.Error(w, "Invalid city", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByCity(r.Context(), orgID, city)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByLostReason handles leads by lost reason retrieval
func (h *LeadHandler) GetLeadsByLostReason(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	lostReasonID, err := uuid.Parse(ps.ByName("lostReasonID"))
	if err != nil {
		http.Error(w, "Invalid lost reason ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByLostReason(r.Context(), orgID, lostReasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByCreatedBy handles leads by created by retrieval
func (h *LeadHandler) GetLeadsByCreatedBy(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	createdBy, err := uuid.Parse(ps.ByName("createdBy"))
	if err != nil {
		http.Error(w, "Invalid created by ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByCreatedBy(r.Context(), orgID, createdBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByUpdatedBy handles leads by updated by retrieval
func (h *LeadHandler) GetLeadsByUpdatedBy(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	updatedBy, err := uuid.Parse(ps.ByName("updatedBy"))
	if err != nil {
		http.Error(w, "Invalid updated by ID", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByUpdatedBy(r.Context(), orgID, updatedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByColor handles leads by color retrieval
func (h *LeadHandler) GetLeadsByColor(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	color, err := strconv.Atoi(ps.ByName("color"))
	if err != nil {
		http.Error(w, "Invalid color", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByColor(r.Context(), orgID, strconv.Itoa(color))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetOverdueLeads handles overdue leads retrieval
func (h *LeadHandler) GetOverdueLeads(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	leads, err := h.leadService.GetOverdueLeads(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetHighValueLeads handles high-value leads retrieval
func (h *LeadHandler) GetHighValueLeads(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	minExpectedRevenue, err := strconv.ParseFloat(r.URL.Query().Get("min_expected_revenue"), 64)
	if err != nil {
		minExpectedRevenue = 10000.0 // Default threshold
	}

	leads, err := h.leadService.GetHighValueLeads(r.Context(), orgID, minExpectedRevenue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetRecentLeads handles recent leads retrieval
func (h *LeadHandler) GetRecentLeads(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	days, err := strconv.Atoi(r.URL.Query().Get("days"))
	if err != nil {
		days = 7 // Default to 7 days
	}

	leads, err := h.leadService.GetRecentLeads(r.Context(), orgID, days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByStatus handles leads by status retrieval
func (h *LeadHandler) GetLeadsByStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	status := ps.ByName("status")
	if status == "" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByStatus(r.Context(), orgID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByPriority handles leads by priority retrieval
func (h *LeadHandler) GetLeadsByPriority(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	priority := types.LeadPriority(ps.ByName("priority"))
	if priority == "" {
		http.Error(w, "Invalid priority", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByPriority(r.Context(), orgID, priority)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByType handles leads by type retrieval
func (h *LeadHandler) GetLeadsByType(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	leadType := types.LeadType(ps.ByName("leadType"))
	if leadType == "" {
		http.Error(w, "Invalid lead type", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByType(r.Context(), orgID, leadType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByWonStatus handles leads by won status retrieval
func (h *LeadHandler) GetLeadsByWonStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	wonStatus := types.LeadWonStatus(ps.ByName("wonStatus"))
	if wonStatus == "" {
		http.Error(w, "Invalid won status", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByWonStatus(r.Context(), orgID, wonStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// GetLeadsByActiveStatus handles leads by active status retrieval
func (h *LeadHandler) GetLeadsByActiveStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	active, err := strconv.ParseBool(ps.ByName("active"))
	if err != nil {
		http.Error(w, "Invalid active status", http.StatusBadRequest)
		return
	}

	leads, err := h.leadService.GetLeadsByActiveStatus(r.Context(), orgID, active)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// CountLeadsByStage handles leads count by stage
func (h *LeadHandler) CountLeadsByStage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByStage(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByPriority handles leads count by priority
func (h *LeadHandler) CountLeadsByPriority(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByPriority(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByType handles leads count by type
func (h *LeadHandler) CountLeadsByType(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByType(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsBySource handles leads count by source
func (h *LeadHandler) CountLeadsBySource(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsBySource(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByMedium handles leads count by medium
func (h *LeadHandler) CountLeadsByMedium(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByMedium(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByCampaign handles leads count by campaign
func (h *LeadHandler) CountLeadsByCampaign(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByCampaign(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByTeam handles leads count by team
func (h *LeadHandler) CountLeadsByTeam(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByTeam(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByUser handles leads count by user
func (h *LeadHandler) CountLeadsByUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByUser(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByLostReason handles leads count by lost reason
func (h *LeadHandler) CountLeadsByLostReason(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByLostReason(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByWonStatus handles leads count by won status
func (h *LeadHandler) CountLeadsByWonStatus(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByWonStatus(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByCountry handles leads count by country
func (h *LeadHandler) CountLeadsByCountry(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByCountry(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByState handles leads count by state
func (h *LeadHandler) CountLeadsByState(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByState(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// CountLeadsByCity handles leads count by city
func (h *LeadHandler) CountLeadsByCity(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context (set by auth middleware)
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	counts, err := h.leadService.CountLeadsByCity(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

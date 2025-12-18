package service

import (
	"context"
	"fmt"
	"time"

	"alieze-erp/internal/modules/crm/types"

	"github.com/google/uuid"
)

// GetLeadPipelineValue calculates the total pipeline value
func (s *LeadService) GetLeadPipelineValue(ctx context.Context, orgID uuid.UUID) (float64, error) {
	// Calculate pipeline value by summing expected revenue of all active leads
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	active := true
	filter.Active = &active

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get leads for pipeline calculation: %w", err)
	}

	var pipelineValue float64
	for _, lead := range leads {
		if lead.ExpectedRevenue != nil {
			pipelineValue += *lead.ExpectedRevenue
		}
	}

	return pipelineValue, nil
}

// GetLeadPipelineValueByStage calculates pipeline value by stage
func (s *LeadService) GetLeadPipelineValueByStage(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]float64, error) {
	// Get counts by stage first (currently unused but kept for future reference)
	counts, err := s.leadRepository.CountByStage(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead counts by stage: %w", err)
	}
	if len(counts) == 0 {
		return nil, fmt.Errorf("no leads found for pipeline calculation")
	}

	// Get all active leads
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	active := true
	filter.Active = &active

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for pipeline calculation: %w", err)
	}

	// Calculate pipeline value by stage
	pipelineByStage := make(map[uuid.UUID]float64)
	for _, lead := range leads {
		if lead.StageID != nil && lead.ExpectedRevenue != nil {
			pipelineByStage[*lead.StageID] += *lead.ExpectedRevenue
		}
	}

	return pipelineByStage, nil
}

// GetLeadConversionRate calculates the lead conversion rate
func (s *LeadService) GetLeadConversionRate(ctx context.Context, orgID uuid.UUID) (float64, error) {
	// Get total leads
	totalFilter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	totalLeads, err := s.leadRepository.FindAll(ctx, totalFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to get total leads: %w", err)
	}

	if len(totalLeads) == 0 {
		return 0.0, nil
	}

	// Get converted leads (won status)
	convertedFilter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	wonStatus := types.LeadWonStatusWon
	convertedFilter.WonStatus = &wonStatus
	convertedLeads, err := s.leadRepository.FindAll(ctx, convertedFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to get converted leads: %w", err)
	}

	conversionRate := (float64(len(convertedLeads)) / float64(len(totalLeads))) * 100
	return conversionRate, nil
}

// GetLeadWinRate calculates the lead win rate
func (s *LeadService) GetLeadWinRate(ctx context.Context, orgID uuid.UUID) (float64, error) {
	// Get closed leads (won or lost)
	closedFilter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	// We need to get leads that are either won or lost
	// This is a simplified approach - in a real implementation, we'd need a more sophisticated query
	closedLeads, err := s.leadRepository.FindAll(ctx, closedFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to get closed leads: %w", err)
	}

	if len(closedLeads) == 0 {
		return 0.0, nil
	}

	// Count won leads
	var wonCount int
	for _, lead := range closedLeads {
		if lead.WonStatus != nil && *lead.WonStatus == types.LeadWonStatusWon {
			wonCount++
		}
	}

	winRate := (float64(wonCount) / float64(len(closedLeads))) * 100
	return winRate, nil
}

// GetLeadLossRate calculates the lead loss rate
func (s *LeadService) GetLeadLossRate(ctx context.Context, orgID uuid.UUID) (float64, error) {
	// Get closed leads (won or lost)
	closedFilter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	closedLeads, err := s.leadRepository.FindAll(ctx, closedFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to get closed leads: %w", err)
	}

	if len(closedLeads) == 0 {
		return 0.0, nil
	}

	// Count lost leads
	var lostCount int
	for _, lead := range closedLeads {
		if lead.WonStatus != nil && *lead.WonStatus == types.LeadWonStatusLost {
			lostCount++
		}
	}

	lossRate := (float64(lostCount) / float64(len(closedLeads))) * 100
	return lossRate, nil
}

// GetLeadAverageExpectedRevenue calculates the average expected revenue
func (s *LeadService) GetLeadAverageExpectedRevenue(ctx context.Context, orgID uuid.UUID) (float64, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	active := true
	filter.Active = &active

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get leads for average revenue calculation: %w", err)
	}

	if len(leads) == 0 {
		return 0, nil
	}

	var totalRevenue float64
	var count int
	for _, lead := range leads {
		if lead.ExpectedRevenue != nil {
			totalRevenue += *lead.ExpectedRevenue
			count++
		}
	}

	if count == 0 {
		return 0.0, nil
	}

	return totalRevenue / float64(count), nil
}

// GetLeadAverageProbability calculates the average probability
func (s *LeadService) GetLeadAverageProbability(ctx context.Context, orgID uuid.UUID) (float64, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	active := true
	filter.Active = &active

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get leads for average probability calculation: %w", err)
	}

	if len(leads) == 0 {
		return 0, nil
	}

	var totalProbability int
	for _, lead := range leads {
		totalProbability += lead.Probability
	}

	return float64(totalProbability) / float64(len(leads)), nil
}

// GetLeadTotalExpectedRevenue calculates the total expected revenue
func (s *LeadService) GetLeadTotalExpectedRevenue(ctx context.Context, orgID uuid.UUID) (float64, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	active := true
	filter.Active = &active

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get leads for total revenue calculation: %w", err)
	}

	var totalRevenue float64
	for _, lead := range leads {
		if lead.ExpectedRevenue != nil {
			totalRevenue += *lead.ExpectedRevenue
		}
	}

	return totalRevenue, nil
}

// GetLeadTotalRecurringRevenue calculates the total recurring revenue
func (s *LeadService) GetLeadTotalRecurringRevenue(ctx context.Context, orgID uuid.UUID) (float64, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	active := true
	filter.Active = &active

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get leads for total recurring revenue calculation: %w", err)
	}

	var totalRecurringRevenue float64
	for _, lead := range leads {
		if lead.RecurringRevenue != nil {
			totalRecurringRevenue += *lead.RecurringRevenue
		}
	}

	return totalRecurringRevenue, nil
}

// GetLeadsBySource retrieves leads by source
func (s *LeadService) GetLeadsBySource(ctx context.Context, orgID uuid.UUID, sourceID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if sourceID == uuid.Nil {
		return nil, fmt.Errorf("invalid source ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		SourceID:       &sourceID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by source: %w", err)
	}

	return leads, nil
}

// GetLeadsByCampaign retrieves leads by campaign
func (s *LeadService) GetLeadsByCampaign(ctx context.Context, orgID uuid.UUID, campaignID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if campaignID == uuid.Nil {
		return nil, fmt.Errorf("invalid campaign ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		CampaignID:     &campaignID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by campaign: %w", err)
	}

	return leads, nil
}

// GetLeadsByMedium retrieves leads by medium
func (s *LeadService) GetLeadsByMedium(ctx context.Context, orgID uuid.UUID, mediumID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if mediumID == uuid.Nil {
		return nil, fmt.Errorf("invalid medium ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		MediumID:       &mediumID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by medium: %w", err)
	}

	return leads, nil
}

// GetLeadsByTag retrieves leads by tag
func (s *LeadService) GetLeadsByTag(ctx context.Context, orgID uuid.UUID, tagID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if tagID == uuid.Nil {
		return nil, fmt.Errorf("invalid tag ID")
	}

	// Note: This is a simplified approach. In a real implementation, we'd need
	// to query a separate lead_tags table or use JSON functions to search within tag_ids
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for tag filtering: %w", err)
	}

	// Filter in memory (not ideal for large datasets)
	var filteredLeads []*types.LeadEnhanced
	for _, lead := range leads {
		for _, id := range lead.TagIDs {
			if id == tagID {
				filteredLeads = append(filteredLeads, lead)
				break
			}
		}
	}

	return filteredLeads, nil
}

// GetLeadsByCompany retrieves leads by company
func (s *LeadService) GetLeadsByCompany(ctx context.Context, orgID uuid.UUID, companyID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if companyID == uuid.Nil {
		return nil, fmt.Errorf("invalid company ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		CompanyID:      &companyID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by company: %w", err)
	}

	return leads, nil
}

// GetLeadsByCountry retrieves leads by country
func (s *LeadService) GetLeadsByCountry(ctx context.Context, orgID uuid.UUID, countryID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if countryID == uuid.Nil {
		return nil, fmt.Errorf("invalid country ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		CountryID:      &countryID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by country: %w", err)
	}

	return leads, nil
}

// GetLeadsByState retrieves leads by state
func (s *LeadService) GetLeadsByState(ctx context.Context, orgID uuid.UUID, stateID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if stateID == uuid.Nil {
		return nil, fmt.Errorf("invalid state ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		StateID:        &stateID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by state: %w", err)
	}

	return leads, nil
}

// GetLeadsByCity retrieves leads by city
func (s *LeadService) GetLeadsByCity(ctx context.Context, orgID uuid.UUID, city string) ([]*types.LeadEnhanced, error) {
	if city == "" {
		return nil, fmt.Errorf("city cannot be empty")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		City:           &city,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by city: %w", err)
	}

	return leads, nil
}

// GetLeadAverageConversionTime calculates the average conversion time for leads
func (s *LeadService) GetLeadAverageConversionTime(ctx context.Context, orgID uuid.UUID) (time.Duration, error) {
	// Get all converted leads
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	wonStatus := types.LeadWonStatusWon
	filter.WonStatus = &wonStatus

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get converted leads: %w", err)
	}

	if len(leads) == 0 {
		return 0, nil
	}

	var totalDuration time.Duration
	var count int

	for _, lead := range leads {
		if lead.CreatedAt.IsZero() || lead.DateClosed.IsZero() {
			continue
		}
		duration := lead.DateClosed.Sub(lead.CreatedAt)
		totalDuration += duration
		count++
	}

	if count == 0 {
		return 0, nil
	}

	return totalDuration / time.Duration(count), nil
}

// GetLeadAverageWinTime calculates the average win time for leads
func (s *LeadService) GetLeadAverageWinTime(ctx context.Context, orgID uuid.UUID) (time.Duration, error) {
	// Get all won leads
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	wonStatus := types.LeadWonStatusWon
	filter.WonStatus = &wonStatus

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get won leads: %w", err)
	}

	if len(leads) == 0 {
		return 0, nil
	}

	var totalDuration time.Duration
	var count int

	for _, lead := range leads {
		if lead.CreatedAt.IsZero() || lead.DateClosed.IsZero() {
			continue
		}
		duration := lead.DateClosed.Sub(lead.CreatedAt)
		totalDuration += duration
		count++
	}

	if count == 0 {
		return 0, nil
	}

	return totalDuration / time.Duration(count), nil
}

// GetLeadAverageLossTime calculates the average loss time for leads
func (s *LeadService) GetLeadAverageLossTime(ctx context.Context, orgID uuid.UUID) (time.Duration, error) {
	// Get all lost leads
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	lostStatus := types.LeadWonStatusLost
	filter.WonStatus = &lostStatus

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get lost leads: %w", err)
	}

	if len(leads) == 0 {
		return 0, nil
	}

	var totalDuration time.Duration
	var count int

	for _, lead := range leads {
		if lead.CreatedAt.IsZero() || lead.DateClosed.IsZero() {
			continue
		}
		duration := lead.DateClosed.Sub(lead.CreatedAt)
		totalDuration += duration
		count++
	}

	if count == 0 {
		return 0, nil
	}

	return totalDuration / time.Duration(count), nil
}

// GetLeadAverageRecurringRevenue calculates the average recurring revenue
func (s *LeadService) GetLeadAverageRecurringRevenue(ctx context.Context, orgID uuid.UUID) (float64, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}
	active := true
	filter.Active = &active

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get leads for average recurring revenue calculation: %w", err)
	}

	if len(leads) == 0 {
		return 0, nil
	}

	var totalRecurringRevenue float64
	var count int
	for _, lead := range leads {
		if lead.RecurringRevenue != nil {
			totalRecurringRevenue += *lead.RecurringRevenue
			count++
		}
	}

	if count == 0 {
		return 0.0, nil
	}

	return totalRecurringRevenue / float64(count), nil
}

// GetLeadsByContact retrieves leads by contact
func (s *LeadService) GetLeadsByContact(ctx context.Context, orgID uuid.UUID, contactID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if contactID == uuid.Nil {
		return nil, fmt.Errorf("invalid contact ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		ContactID:       &contactID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by contact: %w", err)
	}

	return leads, nil
}

// GetLeadsByUser retrieves leads by user
func (s *LeadService) GetLeadsByUser(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		UserID:         &userID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by user: %w", err)
	}

	return leads, nil
}

// GetLeadsByTeam retrieves leads by team
func (s *LeadService) GetLeadsByTeam(ctx context.Context, orgID uuid.UUID, teamID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if teamID == uuid.Nil {
		return nil, fmt.Errorf("invalid team ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		TeamID:         &teamID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by team: %w", err)
	}

	return leads, nil
}

// GetLeadsByStage retrieves leads by stage
func (s *LeadService) GetLeadsByStage(ctx context.Context, orgID uuid.UUID, stageID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if stageID == uuid.Nil {
		return nil, fmt.Errorf("invalid stage ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		StageID:        &stageID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by stage: %w", err)
	}

	return leads, nil
}

// GetLeadsByLostReason retrieves leads by lost reason
func (s *LeadService) GetLeadsByLostReason(ctx context.Context, orgID uuid.UUID, lostReasonID uuid.UUID) ([]*types.LeadEnhanced, error) {
	if lostReasonID == uuid.Nil {
		return nil, fmt.Errorf("invalid lost reason ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		LostReasonID:   &lostReasonID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by lost reason: %w", err)
	}

	return leads, nil
}

// GetOverdueLeads retrieves overdue leads
func (s *LeadService) GetOverdueLeads(ctx context.Context, orgID uuid.UUID) ([]*types.LeadEnhanced, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for overdue filtering: %w", err)
	}

	// Filter in memory for overdue leads
	var overdueLeads []*types.LeadEnhanced
	now := time.Now()
	for _, lead := range leads {
		if lead.DateDeadline != nil && lead.DateDeadline.Before(now) && lead.WonStatus == nil {
			overdueLeads = append(overdueLeads, lead)
		}
	}

	return overdueLeads, nil
}

// GetHighValueLeads retrieves high-value leads
func (s *LeadService) GetHighValueLeads(ctx context.Context, orgID uuid.UUID, minExpectedRevenue float64) ([]*types.LeadEnhanced, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID:          orgID,
		ExpectedRevenueMin:     &minExpectedRevenue,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get high-value leads: %w", err)
	}

	return leads, nil
}

// GetRecentLeads retrieves recently created/modified leads
func (s *LeadService) GetRecentLeads(ctx context.Context, orgID uuid.UUID, days int) ([]*types.LeadEnhanced, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for recent filtering: %w", err)
	}

	// Filter in memory for recent leads
	var recentLeads []*types.LeadEnhanced
	cutoff := time.Now().AddDate(0, 0, -days)
	for _, lead := range leads {
		if lead.CreatedAt.After(cutoff) || lead.UpdatedAt.After(cutoff) {
			recentLeads = append(recentLeads, lead)
		}
	}

	return recentLeads, nil
}

// GetLeadsByCreatedBy retrieves leads by created by user
func (s *LeadService) GetLeadsByCreatedBy(ctx context.Context, orgID uuid.UUID, createdBy uuid.UUID) ([]*types.LeadEnhanced, error) {
	if createdBy == uuid.Nil {
		return nil, fmt.Errorf("invalid created by ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		CreatedBy:      &createdBy,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by created by: %w", err)
	}

	return leads, nil
}

// GetLeadsByUpdatedBy retrieves leads by updated by user
func (s *LeadService) GetLeadsByUpdatedBy(ctx context.Context, orgID uuid.UUID, updatedBy uuid.UUID) ([]*types.LeadEnhanced, error) {
	if updatedBy == uuid.Nil {
		return nil, fmt.Errorf("invalid updated by ID")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		UpdatedBy:      &updatedBy,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by updated by: %w", err)
	}

	return leads, nil
}

// GetLeadsByColor retrieves leads by color
func (s *LeadService) GetLeadsByColor(ctx context.Context, orgID uuid.UUID, color string) ([]*types.LeadEnhanced, error) {
	if color == "" {
		return nil, fmt.Errorf("color cannot be empty")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		Color:          &color,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by color: %w", err)
	}

	return leads, nil
}

// GetLeadsByStatus retrieves leads by status
func (s *LeadService) GetLeadsByStatus(ctx context.Context, orgID uuid.UUID, status string) ([]*types.LeadEnhanced, error) {
	if status == "" {
		return nil, fmt.Errorf("status cannot be empty")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for status filtering: %w", err)
	}

	// Filter in memory for status (simplified approach)
	var filteredLeads []*types.LeadEnhanced
	for _, lead := range leads {
		// This is a simplified approach - in a real implementation, you'd need
		// to define what "status" means for leads (e.g., won/lost, active/inactive, etc.)
		if lead.WonStatus != nil && string(*lead.WonStatus) == status {
			filteredLeads = append(filteredLeads, lead)
		} else if status == "active" && lead.Active {
			filteredLeads = append(filteredLeads, lead)
		} else if status == "inactive" && !lead.Active {
			filteredLeads = append(filteredLeads, lead)
		}
	}

	return filteredLeads, nil
}

// GetLeadsByActiveStatus retrieves leads by active status
func (s *LeadService) GetLeadsByActiveStatus(ctx context.Context, orgID uuid.UUID, active bool) ([]*types.LeadEnhanced, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		Active:         &active,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by active status: %w", err)
	}

	return leads, nil
}

// GetLeadsByPriority retrieves leads by priority
func (s *LeadService) GetLeadsByPriority(ctx context.Context, orgID uuid.UUID, priority types.LeadPriority) ([]*types.LeadEnhanced, error) {
	if priority == "" {
		return nil, fmt.Errorf("priority cannot be empty")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		Priority:       &priority,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by priority: %w", err)
	}

	return leads, nil
}

// GetLeadsByType retrieves leads by type
func (s *LeadService) GetLeadsByType(ctx context.Context, orgID uuid.UUID, leadType types.LeadType) ([]*types.LeadEnhanced, error) {
	if leadType == "" {
		return nil, fmt.Errorf("lead type cannot be empty")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		LeadType:       &leadType,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by type: %w", err)
	}

	return leads, nil
}

// GetLeadsByWonStatus retrieves leads by won status
func (s *LeadService) GetLeadsByWonStatus(ctx context.Context, orgID uuid.UUID, wonStatus types.LeadWonStatus) ([]*types.LeadEnhanced, error) {
	if wonStatus == "" {
		return nil, fmt.Errorf("won status cannot be empty")
	}

	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
		WonStatus:      &wonStatus,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads by won status: %w", err)
	}

	return leads, nil
}

// CountLeadsByStage counts leads by stage
func (s *LeadService) CountLeadsByStage(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	counts, err := s.leadRepository.CountByStage(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count leads by stage: %w", err)
	}

	return counts, nil
}

// CountLeadsByPriority counts leads by priority
func (s *LeadService) CountLeadsByPriority(ctx context.Context, orgID uuid.UUID) (map[types.LeadPriority]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for priority counting: %w", err)
	}

	counts := make(map[types.LeadPriority]int)
	for _, lead := range leads {
		counts[lead.Priority]++
	}

	return counts, nil
}

// CountLeadsByType counts leads by type
func (s *LeadService) CountLeadsByType(ctx context.Context, orgID uuid.UUID) (map[types.LeadType]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for type counting: %w", err)
	}

	counts := make(map[types.LeadType]int)
	for _, lead := range leads {
		counts[lead.LeadType]++
	}

	return counts, nil
}

// CountLeadsBySource counts leads by source
func (s *LeadService) CountLeadsBySource(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for source counting: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, lead := range leads {
		if lead.SourceID != nil {
			counts[*lead.SourceID]++
		}
	}

	return counts, nil
}

// CountLeadsByMedium counts leads by medium
func (s *LeadService) CountLeadsByMedium(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for medium counting: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, lead := range leads {
		if lead.MediumID != nil {
			counts[*lead.MediumID]++
		}
	}

	return counts, nil
}

// CountLeadsByCampaign counts leads by campaign
func (s *LeadService) CountLeadsByCampaign(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for campaign counting: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, lead := range leads {
		if lead.CampaignID != nil {
			counts[*lead.CampaignID]++
		}
	}

	return counts, nil
}

// CountLeadsByTeam counts leads by team
func (s *LeadService) CountLeadsByTeam(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for team counting: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, lead := range leads {
		if lead.TeamID != nil {
			counts[*lead.TeamID]++
		}
	}

	return counts, nil
}

// CountLeadsByUser counts leads by user
func (s *LeadService) CountLeadsByUser(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for user counting: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, lead := range leads {
		if lead.UserID != nil {
			counts[*lead.UserID]++
		}
	}

	return counts, nil
}

// CountLeadsByLostReason counts leads by lost reason
func (s *LeadService) CountLeadsByLostReason(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for lost reason counting: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, lead := range leads {
		if lead.LostReasonID != nil {
			counts[*lead.LostReasonID]++
		}
	}

	return counts, nil
}

// CountLeadsByWonStatus counts leads by won status
func (s *LeadService) CountLeadsByWonStatus(ctx context.Context, orgID uuid.UUID) (map[types.LeadWonStatus]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for won status counting: %w", err)
	}

	counts := make(map[types.LeadWonStatus]int)
	for _, lead := range leads {
		if lead.WonStatus != nil {
			counts[*lead.WonStatus]++
		}
	}

	return counts, nil
}

// CountLeadsByCountry counts leads by country
func (s *LeadService) CountLeadsByCountry(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for country counting: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, lead := range leads {
		if lead.CountryID != nil {
			counts[*lead.CountryID]++
		}
	}

	return counts, nil
}

// CountLeadsByState counts leads by state
func (s *LeadService) CountLeadsByState(ctx context.Context, orgID uuid.UUID) (map[uuid.UUID]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for state counting: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, lead := range leads {
		if lead.StateID != nil {
			counts[*lead.StateID]++
		}
	}

	return counts, nil
}

// CountLeadsByCity counts leads by city
func (s *LeadService) CountLeadsByCity(ctx context.Context, orgID uuid.UUID) (map[string]int, error) {
	filter := types.LeadEnhancedFilter{
		OrganizationID: orgID,
	}

	leads, err := s.leadRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads for city counting: %w", err)
	}

	counts := make(map[string]int)
	for _, lead := range leads {
		if lead.City != nil {
			counts[*lead.City]++
		}
	}

	return counts, nil
}

package repository

import (
	"context"
	"time"
	"database/sql"


	"github.com/google/uuid"
	"alieze-erp/internal/modules/crm/types"
)

// EnhancedLeadRepository defines the interface for enhanced lead data access
type EnhancedLeadRepository interface {
	// FindByID retrieves a lead by its ID
	FindByID(ctx context.Context, db *sql.DB, id uuid.UUID) (*types.LeadEnhanced, error)

	// FindAll retrieves all leads with optional filters
	FindAll(ctx context.Context, db *sql.DB, filter types.LeadFilter) ([]*types.LeadEnhanced, error)

	// Create inserts a new lead
	Create(ctx context.Context, db *sql.DB, lead *types.LeadEnhanced) error

	// Update modifies an existing lead
	Update(ctx context.Context, db *sql.DB, lead *types.LeadEnhanced) error

	// Delete removes a lead
	Delete(ctx context.Context, db *sql.DB, id uuid.UUID) error

	// FindByContact retrieves leads associated with a contact
	FindByContact(ctx context.Context, db *sql.DB, contactID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByUser retrieves leads assigned to a user
	FindByUser(ctx context.Context, db *sql.DB, userID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByTeam retrieves leads assigned to a team
	FindByTeam(ctx context.Context, db *sql.DB, teamID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByStage retrieves leads in a specific stage
	FindByStage(ctx context.Context, db *sql.DB, stageID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindBySource retrieves leads from a specific source
	FindBySource(ctx context.Context, db *sql.DB, sourceID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByDateRange retrieves leads within a date range
	FindByDateRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) ([]*types.LeadEnhanced, error)

	// CountByStage counts leads by stage for pipeline analytics
	CountByStage(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// FindOverdue retrieves leads that are overdue
	FindOverdue(ctx context.Context, db *sql.DB) ([]*types.LeadEnhanced, error)

	// FindHighValue retrieves high-value leads
	FindHighValue(ctx context.Context, db *sql.DB, minValue float64) ([]*types.LeadEnhanced, error)

	// FindRecent retrieves recently created/modified leads
	FindRecent(ctx context.Context, db *sql.DB, days int) ([]*types.LeadEnhanced, error)

	// FindByStatus retrieves leads by their status (open, closed, won, lost)
	FindByStatus(ctx context.Context, db *sql.DB, status string) ([]*types.LeadEnhanced, error)

	// FindByPriority retrieves leads by priority
	FindByPriority(ctx context.Context, db *sql.DB, priority types.LeadPriority) ([]*types.LeadEnhanced, error)

	// FindByType retrieves leads by type
	FindByType(ctx context.Context, db *sql.DB, leadType types.LeadType) ([]*types.LeadEnhanced, error)

	// FindByLostReason retrieves leads by lost reason
	FindByLostReason(ctx context.Context, db *sql.DB, lostReasonID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByCampaign retrieves leads by campaign
	FindByCampaign(ctx context.Context, db *sql.DB, campaignID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByMedium retrieves leads by medium
	FindByMedium(ctx context.Context, db *sql.DB, mediumID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByTag retrieves leads by tag
	FindByTag(ctx context.Context, db *sql.DB, tagID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByCompany retrieves leads by company
	FindByCompany(ctx context.Context, db *sql.DB, companyID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByCountry retrieves leads by country
	FindByCountry(ctx context.Context, db *sql.DB, countryID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByState retrieves leads by state
	FindByState(ctx context.Context, db *sql.DB, stateID uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByCity retrieves leads by city
	FindByCity(ctx context.Context, db *sql.DB, city string) ([]*types.LeadEnhanced, error)

	// FindByExpectedRevenueRange retrieves leads by expected revenue range
	FindByExpectedRevenueRange(ctx context.Context, db *sql.DB, minRevenue, maxRevenue float64) ([]*types.LeadEnhanced, error)

	// FindByProbabilityRange retrieves leads by probability range
	FindByProbabilityRange(ctx context.Context, db *sql.DB, minProbability, maxProbability int) ([]*types.LeadEnhanced, error)

	// FindByRecurringRevenueRange retrieves leads by recurring revenue range
	FindByRecurringRevenueRange(ctx context.Context, db *sql.DB, minRevenue, maxRevenue float64) ([]*types.LeadEnhanced, error)

	// FindByDateOpenRange retrieves leads by date open range
	FindByDateOpenRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) ([]*types.LeadEnhanced, error)

	// FindByDateClosedRange retrieves leads by date closed range
	FindByDateClosedRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) ([]*types.LeadEnhanced, error)

	// FindByDateDeadlineRange retrieves leads by date deadline range
	FindByDateDeadlineRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) ([]*types.LeadEnhanced, error)

	// FindByDateLastStageUpdateRange retrieves leads by date last stage update range
	FindByDateLastStageUpdateRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) ([]*types.LeadEnhanced, error)

	// FindByActiveStatus retrieves leads by active status
	FindByActiveStatus(ctx context.Context, db *sql.DB, active bool) ([]*types.LeadEnhanced, error)

	// FindByWonStatus retrieves leads by won status
	FindByWonStatus(ctx context.Context, db *sql.DB, wonStatus types.LeadWonStatus) ([]*types.LeadEnhanced, error)

	// FindByCustomField retrieves leads by custom field
	FindByCustomField(ctx context.Context, db *sql.DB, fieldName string, fieldValue interface{}) ([]*types.LeadEnhanced, error)

	// FindByMetadata retrieves leads by metadata
	FindByMetadata(ctx context.Context, db *sql.DB, metadataKey string, metadataValue interface{}) ([]*types.LeadEnhanced, error)

	// FindBySearchTerm retrieves leads by search term (name, email, phone, etc.)
	FindBySearchTerm(ctx context.Context, db *sql.DB, searchTerm string) ([]*types.LeadEnhanced, error)

	// FindByColor retrieves leads by color
	FindByColor(ctx context.Context, db *sql.DB, color int) ([]*types.LeadEnhanced, error)

	// FindByCreatedBy retrieves leads by created by user
	FindByCreatedBy(ctx context.Context, db *sql.DB, createdBy uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByUpdatedBy retrieves leads by updated by user
	FindByUpdatedBy(ctx context.Context, db *sql.DB, updatedBy uuid.UUID) ([]*types.LeadEnhanced, error)

	// FindByCreatedAtRange retrieves leads by created at range
	FindByCreatedAtRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) ([]*types.LeadEnhanced, error)

	// FindByUpdatedAtRange retrieves leads by updated at range
	FindByUpdatedAtRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) ([]*types.LeadEnhanced, error)

	// CountTotal counts all leads
	CountTotal(ctx context.Context, db *sql.DB) (int, error)

	// CountActive counts active leads
	CountActive(ctx context.Context, db *sql.DB) (int, error)

	// CountInactive counts inactive leads
	CountInactive(ctx context.Context, db *sql.DB) (int, error)

	// CountByPriority counts leads by priority
	CountByPriority(ctx context.Context, db *sql.DB) (map[types.LeadPriority]int, error)

	// CountByType counts leads by type
	CountByType(ctx context.Context, db *sql.DB) (map[types.LeadType]int, error)

	// CountBySource counts leads by source
	CountBySource(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// CountByMedium counts leads by medium
	CountByMedium(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// CountByCampaign counts leads by campaign
	CountByCampaign(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// CountByTeam counts leads by team
	CountByTeam(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// CountByUser counts leads by user
	CountByUser(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// CountByLostReason counts leads by lost reason
	CountByLostReason(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// CountByWonStatus counts leads by won status
	CountByWonStatus(ctx context.Context, db *sql.DB) (map[types.LeadWonStatus]int, error)

	// CountByCountry counts leads by country
	CountByCountry(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// CountByState counts leads by state
	CountByState(ctx context.Context, db *sql.DB) (map[uuid.UUID]int, error)

	// CountByCity counts leads by city
	CountByCity(ctx context.Context, db *sql.DB) (map[string]int, error)

	// CountByExpectedRevenueRange counts leads by expected revenue range
	CountByExpectedRevenueRange(ctx context.Context, db *sql.DB, minRevenue, maxRevenue float64) (int, error)

	// CountByProbabilityRange counts leads by probability range
	CountByProbabilityRange(ctx context.Context, db *sql.DB, minProbability, maxProbability int) (int, error)

	// CountByRecurringRevenueRange counts leads by recurring revenue range
	CountByRecurringRevenueRange(ctx context.Context, db *sql.DB, minRevenue, maxRevenue float64) (int, error)

	// CountByDateRange counts leads by date range
	CountByDateRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (int, error)

	// CountByDateOpenRange counts leads by date open range
	CountByDateOpenRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (int, error)

	// CountByDateClosedRange counts leads by date closed range
	CountByDateClosedRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (int, error)

	// CountByDateDeadlineRange counts leads by date deadline range
	CountByDateDeadlineRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (int, error)

	// CountByDateLastStageUpdateRange counts leads by date last stage update range
	CountByDateLastStageUpdateRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (int, error)

	// CountByCreatedAtRange counts leads by created at range
	CountByCreatedAtRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (int, error)

	// CountByUpdatedAtRange counts leads by updated at range
	CountByUpdatedAtRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (int, error)

	// GetConversionRate calculates lead conversion rate
	GetConversionRate(ctx context.Context, db *sql.DB) (float64, error)

	// GetWinRate calculates lead win rate
	GetWinRate(ctx context.Context, db *sql.DB) (float64, error)

	// GetLossRate calculates lead loss rate
	GetLossRate(ctx context.Context, db *sql.DB) (float64, error)

	// GetAverageConversionTime calculates average lead conversion time
	GetAverageConversionTime(ctx context.Context, db *sql.DB) (time.Duration, error)

	// GetAverageWinTime calculates average lead win time
	GetAverageWinTime(ctx context.Context, db *sql.DB) (time.Duration, error)

	// GetAverageLossTime calculates average lead loss time
	GetAverageLossTime(ctx context.Context, db *sql.DB) (time.Duration, error)

	// GetAverageExpectedRevenue calculates average expected revenue
	GetAverageExpectedRevenue(ctx context.Context, db *sql.DB) (float64, error)

	// GetAverageProbability calculates average probability
	GetAverageProbability(ctx context.Context, db *sql.DB) (float64, error)

	// GetAverageRecurringRevenue calculates average recurring revenue
	GetAverageRecurringRevenue(ctx context.Context, db *sql.DB) (float64, error)

	// GetTotalExpectedRevenue calculates total expected revenue
	GetTotalExpectedRevenue(ctx context.Context, db *sql.DB) (float64, error)

	// GetTotalRecurringRevenue calculates total recurring revenue
	GetTotalRecurringRevenue(ctx context.Context, db *sql.DB) (float64, error)

	// GetPipelineValue calculates total pipeline value
	GetPipelineValue(ctx context.Context, db *sql.DB) (float64, error)

	// GetPipelineValueByStage calculates pipeline value by stage
	GetPipelineValueByStage(ctx context.Context, db *sql.DB) (map[uuid.UUID]float64, error)

	// GetPipelineValueByUser calculates pipeline value by user
	GetPipelineValueByUser(ctx context.Context, db *sql.DB) (map[uuid.UUID]float64, error)

	// GetPipelineValueByTeam calculates pipeline value by team
	GetPipelineValueByTeam(ctx context.Context, db *sql.DB) (map[uuid.UUID]float64, error)

	// GetPipelineValueBySource calculates pipeline value by source
	GetPipelineValueBySource(ctx context.Context, db *sql.DB) (map[uuid.UUID]float64, error)

	// GetPipelineValueByCampaign calculates pipeline value by campaign
	GetPipelineValueByCampaign(ctx context.Context, db *sql.DB) (map[uuid.UUID]float64, error)

	// GetPipelineValueByMedium calculates pipeline value by medium
	GetPipelineValueByMedium(ctx context.Context, db *sql.DB) (map[uuid.UUID]float64, error)

	// GetPipelineValueByCountry calculates pipeline value by country
	GetPipelineValueByCountry(ctx context.Context, db *sql.DB) (map[uuid.UUID]float64, error)

	// GetPipelineValueByState calculates pipeline value by state
	GetPipelineValueByState(ctx context.Context, db *sql.DB) (map[uuid.UUID]float64, error)

	// GetPipelineValueByCity calculates pipeline value by city
	GetPipelineValueByCity(ctx context.Context, db *sql.DB) (map[string]float64, error)

	// GetPipelineValueByPriority calculates pipeline value by priority
	GetPipelineValueByPriority(ctx context.Context, db *sql.DB) (map[types.LeadPriority]float64, error)

	// GetPipelineValueByType calculates pipeline value by type
	GetPipelineValueByType(ctx context.Context, db *sql.DB) (map[types.LeadType]float64, error)

	// GetPipelineValueByDateRange calculates pipeline value by date range
	GetPipelineValueByDateRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (float64, error)

	// GetPipelineValueByDateOpenRange calculates pipeline value by date open range
	GetPipelineValueByDateOpenRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (float64, error)

	// GetPipelineValueByDateClosedRange calculates pipeline value by date closed range
	GetPipelineValueByDateClosedRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (float64, error)

	// GetPipelineValueByDateDeadlineRange calculates pipeline value by date deadline range
	GetPipelineValueByDateDeadlineRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (float64, error)

	// GetPipelineValueByDateLastStageUpdateRange calculates pipeline value by date last stage update range
	GetPipelineValueByDateLastStageUpdateRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (float64, error)

	// GetPipelineValueByCreatedAtRange calculates pipeline value by created at range
	GetPipelineValueByCreatedAtRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (float64, error)

	// GetPipelineValueByUpdatedAtRange calculates pipeline value by updated at range
	GetPipelineValueByUpdatedAtRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time) (float64, error)

	// GetPipelineValueByActiveStatus calculates pipeline value by active status
	GetPipelineValueByActiveStatus(ctx context.Context, db *sql.DB, active bool) (float64, error)

	// GetPipelineValueByWonStatus calculates pipeline value by won status
	GetPipelineValueByWonStatus(ctx context.Context, db *sql.DB, wonStatus types.LeadWonStatus) (float64, error)

	// GetPipelineValueByLostReason calculates pipeline value by lost reason
	GetPipelineValueByLostReason(ctx context.Context, db *sql.DB, lostReasonID uuid.UUID) (float64, error)

	// GetPipelineValueByCompany calculates pipeline value by company
	GetPipelineValueByCompany(ctx context.Context, db *sql.DB, companyID uuid.UUID) (float64, error)

	// GetPipelineValueByContact calculates pipeline value by contact
	GetPipelineValueByContact(ctx context.Context, db *sql.DB, contactID uuid.UUID) (float64, error)

	// GetPipelineValueByTag calculates pipeline value by tag
	GetPipelineValueByTag(ctx context.Context, db *sql.DB, tagID uuid.UUID) (float64, error)

	// GetPipelineValueByColor calculates pipeline value by color
	GetPipelineValueByColor(ctx context.Context, db *sql.DB, color int) (float64, error)

	// GetPipelineValueByCreatedBy calculates pipeline value by created by user
	GetPipelineValueByCreatedBy(ctx context.Context, db *sql.DB, createdBy uuid.UUID) (float64, error)

	// GetPipelineValueByUpdatedBy calculates pipeline value by updated by user
	GetPipelineValueByUpdatedBy(ctx context.Context, db *sql.DB, updatedBy uuid.UUID) (float64, error)

	// GetPipelineValueByCustomField calculates pipeline value by custom field
	GetPipelineValueByCustomField(ctx context.Context, db *sql.DB, fieldName string, fieldValue interface{}) (float64, error)

	// GetPipelineValueByMetadata calculates pipeline value by metadata
	GetPipelineValueByMetadata(ctx context.Context, db *sql.DB, metadataKey string, metadataValue interface{}) (float64, error)

	// GetPipelineValueBySearchTerm calculates pipeline value by search term
	GetPipelineValueBySearchTerm(ctx context.Context, db *sql.DB, searchTerm string) (float64, error)

	// GetPipelineValueByExpectedRevenueRange calculates pipeline value by expected revenue range
	GetPipelineValueByExpectedRevenueRange(ctx context.Context, db *sql.DB, minRevenue, maxRevenue float64) (float64, error)

	// GetPipelineValueByProbabilityRange calculates pipeline value by probability range
	GetPipelineValueByProbabilityRange(ctx context.Context, db *sql.DB, minProbability, maxProbability int) (float64, error)

	// GetPipelineValueByRecurringRevenueRange calculates pipeline value by recurring revenue range
	GetPipelineValueByRecurringRevenueRange(ctx context.Context, db *sql.DB, minRevenue, maxRevenue float64) (float64, error)

	// GetPipelineValueByDateRangeAndStage calculates pipeline value by date range and stage
	GetPipelineValueByDateRangeAndStage(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndUser calculates pipeline value by date range and user
	GetPipelineValueByDateRangeAndUser(ctx context.Context, db *sql.DB, startDate, endDate time.Time, userID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndTeam calculates pipeline value by date range and team
	GetPipelineValueByDateRangeAndTeam(ctx context.Context, db *sql.DB, startDate, endDate time.Time, teamID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndSource calculates pipeline value by date range and source
	GetPipelineValueByDateRangeAndSource(ctx context.Context, db *sql.DB, startDate, endDate time.Time, sourceID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndCampaign calculates pipeline value by date range and campaign
	GetPipelineValueByDateRangeAndCampaign(ctx context.Context, db *sql.DB, startDate, endDate time.Time, campaignID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndMedium calculates pipeline value by date range and medium
	GetPipelineValueByDateRangeAndMedium(ctx context.Context, db *sql.DB, startDate, endDate time.Time, mediumID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndCountry calculates pipeline value by date range and country
	GetPipelineValueByDateRangeAndCountry(ctx context.Context, db *sql.DB, startDate, endDate time.Time, countryID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndState calculates pipeline value by date range and state
	GetPipelineValueByDateRangeAndState(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stateID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndCity calculates pipeline value by date range and city
	GetPipelineValueByDateRangeAndCity(ctx context.Context, db *sql.DB, startDate, endDate time.Time, city string) (float64, error)

	// GetPipelineValueByDateRangeAndPriority calculates pipeline value by date range and priority
	GetPipelineValueByDateRangeAndPriority(ctx context.Context, db *sql.DB, startDate, endDate time.Time, priority types.LeadPriority) (float64, error)

	// GetPipelineValueByDateRangeAndType calculates pipeline value by date range and type
	GetPipelineValueByDateRangeAndType(ctx context.Context, db *sql.DB, startDate, endDate time.Time, leadType types.LeadType) (float64, error)

	// GetPipelineValueByDateRangeAndActiveStatus calculates pipeline value by date range and active status
	GetPipelineValueByDateRangeAndActiveStatus(ctx context.Context, db *sql.DB, startDate, endDate time.Time, active bool) (float64, error)

	// GetPipelineValueByDateRangeAndWonStatus calculates pipeline value by date range and won status
	GetPipelineValueByDateRangeAndWonStatus(ctx context.Context, db *sql.DB, startDate, endDate time.Time, wonStatus types.LeadWonStatus) (float64, error)

	// GetPipelineValueByDateRangeAndLostReason calculates pipeline value by date range and lost reason
	GetPipelineValueByDateRangeAndLostReason(ctx context.Context, db *sql.DB, startDate, endDate time.Time, lostReasonID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndCompany calculates pipeline value by date range and company
	GetPipelineValueByDateRangeAndCompany(ctx context.Context, db *sql.DB, startDate, endDate time.Time, companyID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndContact calculates pipeline value by date range and contact
	GetPipelineValueByDateRangeAndContact(ctx context.Context, db *sql.DB, startDate, endDate time.Time, contactID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndTag calculates pipeline value by date range and tag
	GetPipelineValueByDateRangeAndTag(ctx context.Context, db *sql.DB, startDate, endDate time.Time, tagID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndColor calculates pipeline value by date range and color
	GetPipelineValueByDateRangeAndColor(ctx context.Context, db *sql.DB, startDate, endDate time.Time, color int) (float64, error)

	// GetPipelineValueByDateRangeAndCreatedBy calculates pipeline value by date range and created by user
	GetPipelineValueByDateRangeAndCreatedBy(ctx context.Context, db *sql.DB, startDate, endDate time.Time, createdBy uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndUpdatedBy calculates pipeline value by date range and updated by user
	GetPipelineValueByDateRangeAndUpdatedBy(ctx context.Context, db *sql.DB, startDate, endDate time.Time, updatedBy uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndCustomField calculates pipeline value by date range and custom field
	GetPipelineValueByDateRangeAndCustomField(ctx context.Context, db *sql.DB, startDate, endDate time.Time, fieldName string, fieldValue interface{}) (float64, error)

	// GetPipelineValueByDateRangeAndMetadata calculates pipeline value by date range and metadata
	GetPipelineValueByDateRangeAndMetadata(ctx context.Context, db *sql.DB, startDate, endDate time.Time, metadataKey string, metadataValue interface{}) (float64, error)

	// GetPipelineValueByDateRangeAndSearchTerm calculates pipeline value by date range and search term
	GetPipelineValueByDateRangeAndSearchTerm(ctx context.Context, db *sql.DB, startDate, endDate time.Time, searchTerm string) (float64, error)

	// GetPipelineValueByDateRangeAndExpectedRevenueRange calculates pipeline value by date range and expected revenue range
	GetPipelineValueByDateRangeAndExpectedRevenueRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time, minRevenue, maxRevenue float64) (float64, error)

	// GetPipelineValueByDateRangeAndProbabilityRange calculates pipeline value by date range and probability range
	GetPipelineValueByDateRangeAndProbabilityRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time, minProbability, maxProbability int) (float64, error)

	// GetPipelineValueByDateRangeAndRecurringRevenueRange calculates pipeline value by date range and recurring revenue range
	GetPipelineValueByDateRangeAndRecurringRevenueRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time, minRevenue, maxRevenue float64) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndUser calculates pipeline value by date range, stage, and user
	GetPipelineValueByDateRangeAndStageAndUser(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, userID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndTeam calculates pipeline value by date range, stage, and team
	GetPipelineValueByDateRangeAndStageAndTeam(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, teamID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndSource calculates pipeline value by date range, stage, and source
	GetPipelineValueByDateRangeAndStageAndSource(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, sourceID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndCampaign calculates pipeline value by date range, stage, and campaign
	GetPipelineValueByDateRangeAndStageAndCampaign(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, campaignID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndMedium calculates pipeline value by date range, stage, and medium
	GetPipelineValueByDateRangeAndStageAndMedium(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, mediumID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndCountry calculates pipeline value by date range, stage, and country
	GetPipelineValueByDateRangeAndStageAndCountry(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, countryID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndState calculates pipeline value by date range, stage, and state
	GetPipelineValueByDateRangeAndStageAndState(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, stateID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndCity calculates pipeline value by date range, stage, and city
	GetPipelineValueByDateRangeAndStageAndCity(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, city string) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndPriority calculates pipeline value by date range, stage, and priority
	GetPipelineValueByDateRangeAndStageAndPriority(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, priority types.LeadPriority) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndType calculates pipeline value by date range, stage, and type
	GetPipelineValueByDateRangeAndStageAndType(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, leadType types.LeadType) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndActiveStatus calculates pipeline value by date range, stage, and active status
	GetPipelineValueByDateRangeAndStageAndActiveStatus(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, active bool) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndWonStatus calculates pipeline value by date range, stage, and won status
	GetPipelineValueByDateRangeAndStageAndWonStatus(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, wonStatus types.LeadWonStatus) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndLostReason calculates pipeline value by date range, stage, and lost reason
	GetPipelineValueByDateRangeAndStageAndLostReason(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, lostReasonID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndCompany calculates pipeline value by date range, stage, and company
	GetPipelineValueByDateRangeAndStageAndCompany(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, companyID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndContact calculates pipeline value by date range, stage, and contact
	GetPipelineValueByDateRangeAndStageAndContact(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, contactID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndTag calculates pipeline value by date range, stage, and tag
	GetPipelineValueByDateRangeAndStageAndTag(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, tagID uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndColor calculates pipeline value by date range, stage, and color
	GetPipelineValueByDateRangeAndStageAndColor(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, color int) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndCreatedBy calculates pipeline value by date range, stage, and created by user
	GetPipelineValueByDateRangeAndStageAndCreatedBy(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, createdBy uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndUpdatedBy calculates pipeline value by date range, stage, and updated by user
	GetPipelineValueByDateRangeAndStageAndUpdatedBy(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, updatedBy uuid.UUID) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndCustomField calculates pipeline value by date range, stage, and custom field
	GetPipelineValueByDateRangeAndStageAndCustomField(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, fieldName string, fieldValue interface{}) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndMetadata calculates pipeline value by date range, stage, and metadata
	GetPipelineValueByDateRangeAndStageAndMetadata(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, metadataKey string, metadataValue interface{}) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndSearchTerm calculates pipeline value by date range, stage, and search term
	GetPipelineValueByDateRangeAndStageAndSearchTerm(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, searchTerm string) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndExpectedRevenueRange calculates pipeline value by date range, stage, and expected revenue range
	GetPipelineValueByDateRangeAndStageAndExpectedRevenueRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, minRevenue, maxRevenue float64) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndProbabilityRange calculates pipeline value by date range, stage, and probability range
	GetPipelineValueByDateRangeAndStageAndProbabilityRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, minProbability, maxProbability int) (float64, error)

	// GetPipelineValueByDateRangeAndStageAndRecurringRevenueRange calculates pipeline value by date range, stage, and recurring revenue range
	GetPipelineValueByDateRangeAndStageAndRecurringRevenueRange(ctx context.Context, db *sql.DB, startDate, endDate time.Time, stageID uuid.UUID, minRevenue, maxRevenue float64) (float64, error)
}

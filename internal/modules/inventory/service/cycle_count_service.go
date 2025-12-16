package service

import (
	"context"
	"fmt"

	"alieze-erp/internal/modules/inventory/repository"
	"alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
)

type CycleCountService struct {
	cycleCountRepo repository.CycleCountRepository
}

func NewCycleCountService(cycleCountRepo repository.CycleCountRepository) *CycleCountService {
	return &CycleCountService{
		cycleCountRepo: cycleCountRepo,
	}
}

// CreateCycleCountPlan creates a new cycle count plan
func (s *CycleCountService) CreateCycleCountPlan(ctx context.Context, request domain.CreateCycleCountPlanRequest) (*domain.CycleCountPlan, error) {
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if request.Frequency == "" {
		return nil, fmt.Errorf("frequency is required")
	}
	if request.ABCClass == "" {
		return nil, fmt.Errorf("abc_class is required")
	}
	if request.CreatedBy == uuid.Nil {
		return nil, fmt.Errorf("created_by is required")
	}

	return s.cycleCountRepo.CreateCycleCountPlan(ctx, request)
}

// GetCycleCountPlan retrieves a cycle count plan
func (s *CycleCountService) GetCycleCountPlan(ctx context.Context, orgID uuid.UUID, planID uuid.UUID) (*domain.CycleCountPlan, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if planID == uuid.Nil {
		return nil, fmt.Errorf("plan_id is required")
	}

	return s.cycleCountRepo.GetCycleCountPlan(ctx, orgID, planID)
}

// ListCycleCountPlans retrieves cycle count plans for an organization
func (s *CycleCountService) ListCycleCountPlans(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]domain.CycleCountPlan, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.cycleCountRepo.ListCycleCountPlans(ctx, orgID, status, limit, offset)
}

// UpdateCycleCountPlanStatus updates the status of a cycle count plan
func (s *CycleCountService) UpdateCycleCountPlanStatus(ctx context.Context, orgID uuid.UUID, planID uuid.UUID, status string) (bool, error) {
	if orgID == uuid.Nil {
		return false, fmt.Errorf("organization_id is required")
	}
	if planID == uuid.Nil {
		return false, fmt.Errorf("plan_id is required")
	}
	if status == "" {
		return false, fmt.Errorf("status is required")
	}

	return s.cycleCountRepo.UpdateCycleCountPlanStatus(ctx, orgID, planID, status)
}

// CreateCycleCountSession creates a new cycle count session
func (s *CycleCountService) CreateCycleCountSession(ctx context.Context, request domain.CreateCycleCountSessionRequest) (*domain.CycleCountSession, error) {
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if request.UserID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required")
	}
	if request.CountMethod == "" {
		return nil, fmt.Errorf("count_method is required")
	}

	return s.cycleCountRepo.CreateCycleCountSession(ctx, request)
}

// GetCycleCountSession retrieves a cycle count session
func (s *CycleCountService) GetCycleCountSession(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) (*domain.CycleCountSession, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if sessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id is required")
	}

	return s.cycleCountRepo.GetCycleCountSession(ctx, orgID, sessionID)
}

// ListCycleCountSessions retrieves cycle count sessions for an organization
func (s *CycleCountService) ListCycleCountSessions(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]domain.CycleCountSession, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.cycleCountRepo.ListCycleCountSessions(ctx, orgID, status, limit, offset)
}

// CompleteCycleCountSession marks a session as completed
func (s *CycleCountService) CompleteCycleCountSession(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) (bool, error) {
	if orgID == uuid.Nil {
		return false, fmt.Errorf("organization_id is required")
	}
	if sessionID == uuid.Nil {
		return false, fmt.Errorf("session_id is required")
	}

	return s.cycleCountRepo.CompleteCycleCountSession(ctx, orgID, sessionID)
}

// AddCycleCountLine adds a count line to a session
func (s *CycleCountService) AddCycleCountLine(ctx context.Context, request domain.AddCycleCountLineRequest) (*domain.CycleCountLine, error) {
	if request.SessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id is required")
	}
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product_id is required")
	}
	if request.LocationID == uuid.Nil {
		return nil, fmt.Errorf("location_id is required")
	}
	if request.CountedQuantity <= 0 {
		return nil, fmt.Errorf("counted_quantity must be positive")
	}
	if request.CountedBy == uuid.Nil {
		return nil, fmt.Errorf("counted_by is required")
	}

	return s.cycleCountRepo.AddCycleCountLine(ctx, request)
}

// GetCycleCountLine retrieves a count line
func (s *CycleCountService) GetCycleCountLine(ctx context.Context, orgID uuid.UUID, lineID uuid.UUID) (*domain.CycleCountLine, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if lineID == uuid.Nil {
		return nil, fmt.Errorf("line_id is required")
	}

	return s.cycleCountRepo.GetCycleCountLine(ctx, orgID, lineID)
}

// ListCycleCountLines retrieves count lines for a session
func (s *CycleCountService) ListCycleCountLines(ctx context.Context, orgID uuid.UUID, sessionID uuid.UUID) ([]domain.CycleCountLine, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if sessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id is required")
	}

	return s.cycleCountRepo.ListCycleCountLines(ctx, orgID, sessionID)
}

// VerifyCycleCountLine verifies a count line
func (s *CycleCountService) VerifyCycleCountLine(ctx context.Context, request domain.VerifyCycleCountLineRequest) (bool, error) {
	if request.LineID == uuid.Nil {
		return false, fmt.Errorf("line_id is required")
	}
	if request.OrganizationID == uuid.Nil {
		return false, fmt.Errorf("organization_id is required")
	}
	if request.VerifiedBy == uuid.Nil {
		return false, fmt.Errorf("verified_by is required")
	}
	if request.Status == "" {
		return false, fmt.Errorf("status is required")
	}

	return s.cycleCountRepo.VerifyCycleCountLine(ctx, request)
}

// CreateAdjustmentFromVariance creates an adjustment from a count variance
func (s *CycleCountService) CreateAdjustmentFromVariance(ctx context.Context, request domain.CreateAdjustmentRequest) (*domain.CycleCountAdjustment, error) {
	if request.LineID == uuid.Nil {
		return nil, fmt.Errorf("line_id is required")
	}
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.AdjustmentType == "" {
		return nil, fmt.Errorf("adjustment_type is required")
	}
	if request.AdjustedBy == uuid.Nil {
		return nil, fmt.Errorf("adjusted_by is required")
	}

	return s.cycleCountRepo.CreateAdjustmentFromVariance(ctx, request)
}

// GetCycleCountAdjustment retrieves an adjustment
func (s *CycleCountService) GetCycleCountAdjustment(ctx context.Context, orgID uuid.UUID, adjustmentID uuid.UUID) (*domain.CycleCountAdjustment, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if adjustmentID == uuid.Nil {
		return nil, fmt.Errorf("adjustment_id is required")
	}

	return s.cycleCountRepo.GetCycleCountAdjustment(ctx, orgID, adjustmentID)
}

// ListCycleCountAdjustments retrieves adjustments for an organization
func (s *CycleCountService) ListCycleCountAdjustments(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]domain.CycleCountAdjustment, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.cycleCountRepo.ListCycleCountAdjustments(ctx, orgID, status, limit, offset)
}

// ApproveCycleCountAdjustment approves an adjustment and updates stock
func (s *CycleCountService) ApproveCycleCountAdjustment(ctx context.Context, request domain.ApproveAdjustmentRequest) (bool, error) {
	if request.AdjustmentID == uuid.Nil {
		return false, fmt.Errorf("adjustment_id is required")
	}
	if request.OrganizationID == uuid.Nil {
		return false, fmt.Errorf("organization_id is required")
	}
	if request.ApprovedBy == uuid.Nil {
		return false, fmt.Errorf("approved_by is required")
	}

	return s.cycleCountRepo.ApproveCycleCountAdjustment(ctx, request)
}

// GetCycleCountAccuracyMetrics retrieves accuracy metrics
func (s *CycleCountService) GetCycleCountAccuracyMetrics(ctx context.Context, request domain.GetAccuracyMetricsRequest) (*domain.CycleCountMetrics, error) {
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.cycleCountRepo.GetCycleCountAccuracyMetrics(ctx, request)
}

// GetProductsNeedingCycleCount retrieves products that need cycle counting
func (s *CycleCountService) GetProductsNeedingCycleCount(ctx context.Context, request domain.GetProductsNeedingCountRequest) ([]domain.ProductNeedingCycleCount, error) {
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.cycleCountRepo.GetProductsNeedingCycleCount(ctx, request)
}

// GetCycleCountAccuracyHistory retrieves accuracy history
func (s *CycleCountService) GetCycleCountAccuracyHistory(ctx context.Context, orgID uuid.UUID, productID *uuid.UUID, limit, offset int) ([]domain.CycleCountAccuracy, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	return s.cycleCountRepo.GetCycleCountAccuracyHistory(ctx, orgID, productID, limit, offset)
}

// GetCycleCountDashboard retrieves a comprehensive dashboard for cycle counting
func (s *CycleCountService) GetCycleCountDashboard(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	dashboard := make(map[string]interface{})

	// Get active plans
	plans, err := s.ListCycleCountPlans(ctx, orgID, func() *string { s := "active"; return &s }(), 5, 0)
	if err != nil {
		return nil, err
	}
	dashboard["active_plans"] = plans

	// Get in-progress sessions
	sessions, err := s.ListCycleCountSessions(ctx, orgID, func() *string { s := "in_progress"; return &s }(), 10, 0)
	if err != nil {
		return nil, err
	}
	dashboard["in_progress_sessions"] = sessions

	// Get accuracy metrics (last 30 days)
	metrics, err := s.GetCycleCountAccuracyMetrics(ctx, domain.GetAccuracyMetricsRequest{
		OrganizationID: orgID,
		DateFrom:       func() *time.Time { d := time.Now().AddDate(0, 0, -30); return &d }(),
	})
	if err != nil {
		return nil, err
	}
	dashboard["accuracy_metrics"] = metrics

	// Get products needing attention
	products, err := s.GetProductsNeedingCycleCount(ctx, domain.GetProductsNeedingCountRequest{
		OrganizationID: orgID,
		DaysSinceLastCount: func() *int { d := 30; return &d }(),
		MinVariancePercentage: func() *float64 { v := 5.0; return &v }(),
	})
	if err != nil {
		return nil, err
	}
	dashboard["products_needing_attention"] = products

	// Get pending adjustments
	adjustments, err := s.ListCycleCountAdjustments(ctx, orgID, func() *string { s := "pending"; return &s }(), 10, 0)
	if err != nil {
		return nil, err
	}
	dashboard["pending_adjustments"] = adjustments

	return dashboard, nil
}

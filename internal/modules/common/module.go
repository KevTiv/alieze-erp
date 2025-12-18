package common

import (
	"context"
	"log/slog"

	"alieze-erp/internal/modules/common/handler"
	"alieze-erp/internal/modules/common/repository"
	"alieze-erp/internal/modules/common/service"
	"alieze-erp/pkg/registry"
	"github.com/julienschmidt/httprouter"
)

// CommonModule represents the common module for foundation/reference data
type CommonModule struct {
	attachmentHandler      *handler.AttachmentHandler
	currencyHandler        *handler.CurrencyHandler
	countryHandler         *handler.CountryHandler
	stateHandler           *handler.StateHandler
	uomCategoryHandler     *handler.UOMCategoryHandler
	uomUnitHandler         *handler.UOMUnitHandler
	paymentTermHandler     *handler.PaymentTermHandler
	fiscalPositionHandler  *handler.FiscalPositionHandler
	analyticAccountHandler *handler.AnalyticAccountHandler
	industryHandler        *handler.IndustryHandler
	utmCampaignHandler     *handler.UTMCampaignHandler
	utmMediumHandler       *handler.UTMMediumHandler
	utmSourceHandler       *handler.UTMSourceHandler
	logger                 *slog.Logger
}

// NewCommonModule creates a new common module
func NewCommonModule() *CommonModule {
	return &CommonModule{}
}

// Name returns the module name
func (m *CommonModule) Name() string {
	return "common"
}

// Init initializes the common module
func (m *CommonModule) Init(ctx context.Context, deps registry.Dependencies) error {
	// Initialize logger
	m.logger = deps.Logger.With("module", "common")
	m.logger.Info("Initializing common module")

	// Create repositories
	attachmentRepo := repository.NewAttachmentRepository(deps.DB)
	currencyRepo := repository.NewCurrencyRepository(deps.DB)
	countryRepo := repository.NewCountryRepository(deps.DB)
	stateRepo := repository.NewStateRepository(deps.DB)
	uomCategoryRepo := repository.NewUOMCategoryRepository(deps.DB)
	uomUnitRepo := repository.NewUOMUnitRepository(deps.DB)
	paymentTermRepo := repository.NewPaymentTermRepository(deps.DB)
	fiscalPositionRepo := repository.NewFiscalPositionRepository(deps.DB)
	analyticAccountRepo := repository.NewAnalyticAccountRepository(deps.DB)
	industryRepo := repository.NewIndustryRepository(deps.DB)
	utmCampaignRepo := repository.NewUTMCampaignRepository(deps.DB)
	utmMediumRepo := repository.NewUTMMediumRepository(deps.DB)
	utmSourceRepo := repository.NewUTMSourceRepository(deps.DB)

	// Create services
	attachmentService := service.NewAttachmentService(attachmentRepo)
	currencyService := service.NewCurrencyService(currencyRepo)
	countryService := service.NewCountryService(countryRepo)
	stateService := service.NewStateService(stateRepo)
	uomCategoryService := service.NewUOMCategoryService(uomCategoryRepo)
	uomUnitService := service.NewUOMUnitService(uomUnitRepo)
	paymentTermService := service.NewPaymentTermService(paymentTermRepo)
	fiscalPositionService := service.NewFiscalPositionService(fiscalPositionRepo)
	analyticAccountService := service.NewAnalyticAccountService(analyticAccountRepo)
	industryService := service.NewIndustryService(industryRepo)
	utmCampaignService := service.NewUTMCampaignService(utmCampaignRepo)
	utmMediumService := service.NewUTMMediumService(utmMediumRepo)
	utmSourceService := service.NewUTMSourceService(utmSourceRepo)

	// Create handlers
	m.attachmentHandler = handler.NewAttachmentHandler(attachmentService)
	m.currencyHandler = handler.NewCurrencyHandler(currencyService)
	m.countryHandler = handler.NewCountryHandler(countryService)
	m.stateHandler = handler.NewStateHandler(stateService)
	m.uomCategoryHandler = handler.NewUOMCategoryHandler(uomCategoryService)
	m.uomUnitHandler = handler.NewUOMUnitHandler(uomUnitService)
	m.paymentTermHandler = handler.NewPaymentTermHandler(paymentTermService)
	m.fiscalPositionHandler = handler.NewFiscalPositionHandler(fiscalPositionService)
	m.analyticAccountHandler = handler.NewAnalyticAccountHandler(analyticAccountService)
	m.industryHandler = handler.NewIndustryHandler(industryService)
	m.utmCampaignHandler = handler.NewUTMCampaignHandler(utmCampaignService)
	m.utmMediumHandler = handler.NewUTMMediumHandler(utmMediumService)
	m.utmSourceHandler = handler.NewUTMSourceHandler(utmSourceService)

	m.logger.Info("Common module initialized successfully")
	return nil
}

// RegisterRoutes registers common module routes
func (m *CommonModule) RegisterRoutes(router interface{}) {
	if router == nil {
		return
	}

	if r, ok := router.(*httprouter.Router); ok {
		if m.attachmentHandler != nil {
			m.attachmentHandler.RegisterRoutes(r)
		}
		if m.currencyHandler != nil {
			m.currencyHandler.RegisterRoutes(r)
		}
		if m.countryHandler != nil {
			m.countryHandler.RegisterRoutes(r)
		}
		if m.stateHandler != nil {
			m.stateHandler.RegisterRoutes(r)
		}
		if m.uomCategoryHandler != nil {
			m.uomCategoryHandler.RegisterRoutes(r)
		}
		if m.uomUnitHandler != nil {
			m.uomUnitHandler.RegisterRoutes(r)
		}
		if m.paymentTermHandler != nil {
			m.paymentTermHandler.RegisterRoutes(r)
		}
		if m.fiscalPositionHandler != nil {
			m.fiscalPositionHandler.RegisterRoutes(r)
		}
		if m.analyticAccountHandler != nil {
			m.analyticAccountHandler.RegisterRoutes(r)
		}
		if m.industryHandler != nil {
			m.industryHandler.RegisterRoutes(r)
		}
		if m.utmCampaignHandler != nil {
			m.utmCampaignHandler.RegisterRoutes(r)
		}
		if m.utmMediumHandler != nil {
			m.utmMediumHandler.RegisterRoutes(r)
		}
		if m.utmSourceHandler != nil {
			m.utmSourceHandler.RegisterRoutes(r)
		}
	}
}

// Health checks the health of the common module
func (m *CommonModule) Health() error {
	return nil
}

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/KevTiv/alieze-erp/internal/database"
	"github.com/KevTiv/alieze-erp/pkg/audit"
	authmodule "github.com/KevTiv/alieze-erp/internal/modules/auth"
	commonmodule "github.com/KevTiv/alieze-erp/internal/modules/common"
	crmmodule "github.com/KevTiv/alieze-erp/internal/modules/crm"
	accountingmodule "github.com/KevTiv/alieze-erp/internal/modules/accounting"
	inventorymodule "github.com/KevTiv/alieze-erp/internal/modules/inventory"
	productsmodule "github.com/KevTiv/alieze-erp/internal/modules/products"
	salesmodule "github.com/KevTiv/alieze-erp/internal/modules/sales"
	deliverymodule "github.com/KevTiv/alieze-erp/internal/modules/delivery"
	"github.com/KevTiv/alieze-erp/pkg/events"
	"github.com/KevTiv/alieze-erp/pkg/policy"
	"github.com/KevTiv/alieze-erp/pkg/registry"
	"github.com/KevTiv/alieze-erp/pkg/rules"
	"github.com/KevTiv/alieze-erp/pkg/workflow"
)

type Server struct {
	port              int

	db               database.Service
	authModule       *authmodule.AuthModule
	registry         *registry.Registry
	eventBus         *events.Bus
	ruleEngine       *rules.RuleEngine
	policyEngine     *policy.Engine
	stateMachineFactory *workflow.StateMachineFactory
	logger           *slog.Logger
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialize logger
	logger := slog.Default()

	// Initialize database
	dbService := database.New()

	// Initialize core infrastructure
	eventBus := events.NewBus(false) // Use synchronous event processing for now

	// Initialize rule engine and load configurations
	ruleEngine := rules.NewRuleEngine(nil)
	if err := ruleEngine.LoadConfigFromFile("config/rules/crm.yaml"); err != nil {
		logger.Error("Failed to load CRM rules", "error", err)
		// Continue without rules - they're optional for now
	}

	// Initialize policy engine with Casbin
	// Build connection string from environment variables
	dbHost := os.Getenv("BLUEPRINT_DB_HOST")
	dbPort := os.Getenv("BLUEPRINT_DB_PORT")
	dbUser := os.Getenv("BLUEPRINT_DB_USERNAME")
	dbPass := os.Getenv("BLUEPRINT_DB_PASSWORD")
	dbName := os.Getenv("BLUEPRINT_DB_DATABASE")

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	casbinEnforcer, err := policy.NewCasbinEnforcer(connString, "")
	if err != nil {
		logger.Warn("Failed to initialize Casbin enforcer, using mock mode", "error", err)
		casbinEnforcer, _ = policy.NewCasbinEnforcer("", "") // Fallback to mock mode
	}
	policyEngine := policy.NewEngineWithCasbin(casbinEnforcer)
	logger.Info("Policy engine initialized")

	// Create audit logger with memory repository
	auditRepo := audit.NewMemoryAuditLogRepository()
	auditLogger := audit.NewAuditLogger(auditRepo)

	// Create permission-aware database connection with audit logging
	permissionDB := database.NewPermissionDB(dbService.GetDB(), policyEngine, auditLogger)

	// Initialize state machine factory
	stateMachineFactory := workflow.NewStateMachineFactory()
	if err := stateMachineFactory.LoadFromDirectory("config/workflows"); err != nil {
		logger.Error("Failed to load workflow configurations", "error", err)
		// Continue without workflows - they're optional for now
	}

	// Initialize base dependencies
	baseDeps := registry.Dependencies{
		DB:                  permissionDB,
		EventBus:            eventBus,
		RuleEngine:          ruleEngine,
		PolicyEngine:        policyEngine,
		StateMachineFactory: stateMachineFactory,
		Logger:              logger,
	}

	// Create registry with base dependencies
	repoRegistry := registry.NewRegistry(baseDeps)

	// Register all modules
	authMod := authmodule.NewAuthModule()
	commonMod := commonmodule.NewCommonModule()
	crmMod := crmmodule.NewCRMModule()
	inventoryMod := inventorymodule.NewInventoryModule()
	accountingMod := accountingmodule.NewAccountingModule()
	productsMod := productsmodule.NewProductsModule()
	salesMod := salesmodule.NewSalesModule()
	deliveryMod := deliverymodule.NewDeliveryModule()

	repoRegistry.Register(authMod)
	repoRegistry.Register(commonMod)
	repoRegistry.Register(crmMod)
	repoRegistry.Register(inventoryMod)
	repoRegistry.Register(accountingMod)
	repoRegistry.Register(productsMod)
	repoRegistry.Register(salesMod)
	repoRegistry.Register(deliveryMod)

	// Phase 1: Initialize auth, common, and products modules first (needed by inventory)
	ctx := context.Background()
	if err := authMod.Init(ctx, baseDeps); err != nil {
		logger.Error("Failed to initialize auth module", "error", err)
		os.Exit(1)
	}
	if err := commonMod.Init(ctx, baseDeps); err != nil {
		logger.Error("Failed to initialize common module", "error", err)
		os.Exit(1)
	}

	// Get AuthService and ProductRepo for dependencies
	baseDeps.AuthService = authMod.GetAuthService()
	baseDeps.ProductRepo = productsMod // Products module will be init with ProductRepo=nil initially

	if err := productsMod.Init(ctx, baseDeps); err != nil {
		logger.Error("Failed to initialize products module", "error", err)
		os.Exit(1)
	}

	// Phase 2: Initialize inventory module to get integration service
	if err := inventoryMod.Init(ctx, baseDeps); err != nil {
		logger.Error("Failed to initialize inventory module", "error", err)
		os.Exit(1)
	}

	// Get inventory integration service and add to dependencies
	baseDeps.InventoryService = inventoryMod.GetIntegrationService()

	// Update registry dependencies
	repoRegistry.UpdateDependencies(baseDeps)

	// Phase 3: Initialize remaining modules with full dependencies
	if err := crmMod.Init(ctx, baseDeps); err != nil {
		logger.Error("Failed to initialize CRM module", "error", err)
		os.Exit(1)
	}
	if err := accountingMod.Init(ctx, baseDeps); err != nil {
		logger.Error("Failed to initialize accounting module", "error", err)
		os.Exit(1)
	}
	if err := salesMod.Init(ctx, baseDeps); err != nil {
		logger.Error("Failed to initialize sales module", "error", err)
		os.Exit(1)
	}
	if err := deliveryMod.Init(ctx, baseDeps); err != nil {
		logger.Error("Failed to initialize delivery module", "error", err)
		os.Exit(1)
	}

	// Register event handlers for all modules
	repoRegistry.RegisterAllEventHandlers(eventBus)
	logger.Info("Event handlers registered for all modules")

	NewServer := &Server{
		port:              port,
		db:                dbService,
		authModule:        authMod,
		registry:          repoRegistry,
		eventBus:          eventBus,
		ruleEngine:        ruleEngine,
		policyEngine:      policyEngine,
		stateMachineFactory: stateMachineFactory,
		logger:            logger,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

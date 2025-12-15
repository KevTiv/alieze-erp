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

	"alieze-erp/internal/database"
	authmodule "alieze-erp/internal/modules/auth"
	crmmodule "alieze-erp/internal/modules/crm"
	accountingmodule "alieze-erp/internal/modules/accounting"
	productsmodule "alieze-erp/internal/modules/products"
	salesmodule "alieze-erp/internal/modules/sales"
	"alieze-erp/pkg/events"
	"alieze-erp/pkg/policy"
	"alieze-erp/pkg/registry"
	"alieze-erp/pkg/rules"
	"alieze-erp/pkg/workflow"
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

	// Initialize policy engine (Casbin)
	policyEngine := policy.NewEngine(nil) // Will initialize Casbin later
	if err := policyEngine.LoadConfigFromFile("config/policy/rules.yaml"); err != nil {
		logger.Error("Failed to load policy rules", "error", err)
		// Continue without policy rules - they're optional for now
	}

	// Initialize state machine factory
	stateMachineFactory := workflow.NewStateMachineFactory()
	if err := stateMachineFactory.LoadFromDirectory("config/workflows"); err != nil {
		logger.Error("Failed to load workflow configurations", "error", err)
		// Continue without workflows - they're optional for now
	}

	// Create registry with dependencies
	repoRegistry := registry.NewRegistry(registry.Dependencies{
		DB:               dbService.GetDB(),
		EventBus:         eventBus,
		RuleEngine:       ruleEngine,
		PolicyEngine:     policyEngine,
		StateMachineFactory: stateMachineFactory,
		Logger:           logger,
	})

	// Register all modules
	repoRegistry.Register(authmodule.NewAuthModule())
	repoRegistry.Register(crmmodule.NewCRMModule())
	repoRegistry.Register(accountingmodule.NewAccountingModule())
	repoRegistry.Register(productsmodule.NewProductsModule())
	repoRegistry.Register(salesmodule.NewSalesModule())

	// Initialize all modules
	if err := repoRegistry.InitAll(context.Background()); err != nil {
		logger.Error("Failed to initialize modules", "error", err)
		os.Exit(1)
	}

	// Register event handlers for all modules
	repoRegistry.RegisterAllEventHandlers(eventBus)
	logger.Info("Event handlers registered for all modules")

	// Get auth module for middleware
	authModule, ok := repoRegistry.GetModule("auth")
	if !ok {
		logger.Error("Auth module not found")
		os.Exit(1)
	}
	authMod, ok := authModule.(*authmodule.AuthModule)
	if !ok {
		logger.Error("Failed to cast auth module")
		os.Exit(1)
	}

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

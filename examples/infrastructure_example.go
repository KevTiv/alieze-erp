package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"alieze-erp/pkg/events"
	"alieze-erp/pkg/policy"
	"alieze-erp/pkg/registry"
	"alieze-erp/pkg/rules"
	"alieze-erp/pkg/workflow"
)

// ExampleModule demonstrates how to implement a module
type ExampleModule struct {
	name string
}

func (m *ExampleModule) Name() string {
	return m.name
}

func (m *ExampleModule) Init(ctx context.Context, deps registry.Dependencies) error {
	slog.Info("Initializing example module", "name", m.name)
	return nil
}

func (m *ExampleModule) RegisterRoutes(router interface{}) {
	slog.Info("Registering routes for example module", "name", m.name)
}

func (m *ExampleModule) RegisterEventHandlers(bus interface{}) {
	slog.Info("Registering event handlers for example module", "name", m.name)

	if eventBus, ok := bus.(*events.Bus); ok {
		eventBus.Subscribe("example.event", func(ctx context.Context, event events.Event) error {
			slog.Info("Received event", "type", event.Type, "payload", event.Payload)
			return nil
		})
	}
}

func (m *ExampleModule) Health() error {
	return nil
}

func main() {
	ctx := context.Background()

	// Initialize dependencies
	logger := slog.Default()

	// Create mock database connection (in real app, this would be a real connection)
	var db *pgxpool.Pool // nil for example

	// Create event bus
	eventBus := events.NewBus(true) // async

	// Create rule engine with empty config
	ruleConfig := &rules.RuleConfig{
		Modules: make(map[string]rules.ModuleRules),
	}
	ruleEngine := rules.NewRuleEngine(ruleConfig)

	// Create policy engine with mock Casbin enforcer
	casbinEnforcer, _ := policy.NewCasbinEnforcer("", "")
	policyEngine := policy.NewEngine(casbinEnforcer)

	// Create dependencies
	deps := registry.Dependencies{
		DB:           db,
		EventBus:     eventBus,
		RuleEngine:   ruleEngine,
		PolicyEngine: policyEngine,
		Logger:       logger,
	}

	// Create registry
	registry := registry.NewRegistry(deps)

	// Register modules
	exampleModule := &ExampleModule{name: "example"}
	registry.Register(exampleModule)

	// Initialize all modules
	if err := registry.InitAll(ctx); err != nil {
		logger.Error("Failed to initialize modules", "error", err)
		return
	}

	// Register event handlers
	for _, module := range registry.GetAllModules() {
		module.RegisterEventHandlers(eventBus)
	}

	// Publish a test event
	eventBus.Publish(ctx, "example.event", map[string]string{"message": "Hello from infrastructure!"})

	// Test rule engine
	ruleEngine.RegisterValidator("test_validator", func(ctx context.Context, entity interface{}) error {
		logger.Info("Test validator called", "entity", entity)
		return nil
	})

	if err := ruleEngine.Validate(ctx, "test_validator", "test_entity"); err != nil {
		logger.Error("Validation failed", "error", err)
	}

	// Test policy engine
	if allowed, err := policyEngine.CheckPermission(ctx, "user1", "resource1", "read"); err != nil {
		logger.Error("Permission check failed", "error", err)
	} else {
		logger.Info("Permission check result", "allowed", allowed)
	}

	// Test state machine
	workflowConfig := &workflow.StateMachineConfig{
		WorkflowID: "example.workflow",
		Model:      "example.model",
		Initial:    "draft",
		States:     []string{"draft", "published", "archived"},
		Transitions: []workflow.StateTransition{
			{
				Name:      "publish",
				From:      []string{"draft"},
				To:        "published",
				Validator: "can_publish",
			},
		},
	}

	stateMachine := workflow.NewStateMachine(workflowConfig)
	stateMachine.RegisterValidator("can_publish", func(ctx context.Context, entity interface{}) error {
		logger.Info("Can publish validator called", "entity", entity)
		return nil
	})

	logger.Info("Current state", "state", stateMachine.CurrentState())

	if err := stateMachine.Transition(ctx, "publish", "example_entity"); err != nil {
		logger.Error("State transition failed", "error", err)
	} else {
		logger.Info("State transition successful", "new_state", stateMachine.CurrentState())
	}

	fmt.Println("Infrastructure example completed successfully!")
}

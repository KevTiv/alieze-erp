package registry

import (
	"database/sql"
	"log/slog"

	"alieze-erp/pkg/events"
	"alieze-erp/pkg/policy"
	"alieze-erp/pkg/rules"
	"alieze-erp/pkg/workflow"
)

// Dependencies contains the shared dependencies for all modules
type Dependencies struct {
	DB                  *sql.DB
	EventBus            *events.Bus
	RuleEngine          *rules.RuleEngine
	PolicyEngine        *policy.Engine
	StateMachineFactory *workflow.StateMachineFactory
	Logger              *slog.Logger
	ProductRepo         interface{} // Product repository for inventory module
	AuthService         interface{} // Auth service for quality control
	InventoryService    interface{} // Inventory integration service for delivery module
}

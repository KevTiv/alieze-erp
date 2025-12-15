package registry

import (
	"database/sql"
	"log/slog"
)

// Dependencies contains the shared dependencies for all modules
type Dependencies struct {
	DB               *sql.DB
	EventBus         interface{}
	RuleEngine       interface{}
	PolicyEngine     interface{}
	StateMachineFactory interface{}
	Logger           *slog.Logger
}

package rules

// RuleConfig defines the configuration for validation rules and state machines
type RuleConfig struct {
	Modules map[string]ModuleRules `yaml:"modules"`
}

// ModuleRules contains validation and state machine rules for a module
type ModuleRules struct {
	Validation   map[string][]ValidationRule `yaml:"validation"`
	StateMachine map[string]StateMachine     `yaml:"state_machine"`
	Calculations map[string]CalculationRule  `yaml:"calculations"`
}

// ValidationRule defines a validation rule
type ValidationRule struct {
	Name      string                 `yaml:"name"`
	Validator string                 `yaml:"validator"`
	Field     string                 `yaml:"field"`
	Params    map[string]interface{} `yaml:"params"`
}

// StateMachine defines a state machine configuration
type StateMachine struct {
	Initial     string                 `yaml:"initial"`
	States      []string               `yaml:"states"`
	Transitions []StateTransition      `yaml:"transitions"`
}

// StateTransition defines a state transition
type StateTransition struct {
	Name       string                 `yaml:"name"`
	From       []string               `yaml:"from"`
	To         string                 `yaml:"to"`
	Validator  string                 `yaml:"validator"`
	Permission string                 `yaml:"permission"`
}

// CalculationRule defines a calculation rule
type CalculationRule struct {
	Formula string `yaml:"formula"`
}

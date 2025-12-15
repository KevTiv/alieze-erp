package workflow

// StateMachineConfig defines the configuration for a state machine
type StateMachineConfig struct {
	WorkflowID string            `yaml:"workflow_id"`
	Model      string            `yaml:"model"`
	Initial    string            `yaml:"initial"`
	States     []string          `yaml:"states"`
	Transitions []StateTransition `yaml:"transitions"`
}

// StateTransition defines a state transition
type StateTransition struct {
	Name       string `yaml:"name"`
	From       []string `yaml:"from"`
	To         string `yaml:"to"`
	Validator  string `yaml:"validator"`
	Permission string `yaml:"permission"`
}

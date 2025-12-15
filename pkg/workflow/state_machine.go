package workflow

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v3"
)

// StateMachine manages state transitions for business entities
type StateMachine struct {
	currentState string
	config       *StateMachineConfig
	validators   map[string]TransitionValidator
	mu           sync.RWMutex
}

// TransitionValidator validates state transitions
type TransitionValidator func(ctx context.Context, entity interface{}) error

// NewStateMachine creates a new state machine
func NewStateMachine(config *StateMachineConfig) *StateMachine {
	return &StateMachine{
		currentState: config.Initial,
		config:       config,
		validators:   make(map[string]TransitionValidator),
	}
}

// NewStateMachineFromFile creates a state machine from a YAML configuration file
func NewStateMachineFromFile(filePath string) (*StateMachine, error) {
	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state machine config file: %w", err)
	}

	// Parse YAML
	var config StateMachineConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse state machine config: %w", err)
	}

	return NewStateMachine(&config), nil
}

// StateMachineFactory manages multiple state machines for different workflows
type StateMachineFactory struct {
	stateMachines map[string]*StateMachine
	mu            sync.RWMutex
}

// NewStateMachineFactory creates a new state machine factory
func NewStateMachineFactory() *StateMachineFactory {
	return &StateMachineFactory{
		stateMachines: make(map[string]*StateMachine),
	}
}

// LoadFromDirectory loads all state machine configurations from a directory
func (smf *StateMachineFactory) LoadFromDirectory(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read workflow config directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := dirPath + "/" + file.Name()
		if filePath == "." || filePath == ".." {
			continue
		}

		sm, err := NewStateMachineFromFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load state machine from %s: %w", file.Name(), err)
		}

		smf.stateMachines[sm.config.WorkflowID] = sm
	}

	return nil
}

// GetStateMachine returns a state machine by workflow ID
func (smf *StateMachineFactory) GetStateMachine(workflowID string) (*StateMachine, bool) {
	smf.mu.RLock()
	defer smf.mu.RUnlock()

	sm, exists := smf.stateMachines[workflowID]
	return sm, exists
}

// GetAllStateMachines returns all registered state machines
func (smf *StateMachineFactory) GetAllStateMachines() map[string]*StateMachine {
	smf.mu.RLock()
	defer smf.mu.RUnlock()

	return smf.stateMachines
}

// Transition moves the state machine to a new state
func (sm *StateMachine) Transition(ctx context.Context, transitionName string, entity interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find the transition
	var transition *StateTransition
	for _, t := range sm.config.Transitions {
		if t.Name == transitionName {
			transition = &t
			break
		}
	}

	if transition == nil {
		return fmt.Errorf("transition %s not found", transitionName)
	}

	// Check if current state allows this transition
	if !sm.canTransition(transition) {
		return fmt.Errorf("cannot transition from %s using %s", sm.currentState, transitionName)
	}

	// Validate transition
	if transition.Validator != "" {
		if validator, exists := sm.validators[transition.Validator]; exists {
			if err := validator(ctx, entity); err != nil {
				return fmt.Errorf("transition validation failed: %w", err)
			}
		}
	}

	// Perform transition
	sm.currentState = transition.To
	return nil
}

// canTransition checks if a transition is allowed from current state
func (sm *StateMachine) canTransition(transition *StateTransition) bool {
	for _, fromState := range transition.From {
		if fromState == sm.currentState || fromState == "*" {
			return true
		}
	}
	return false
}

// CurrentState returns the current state
func (sm *StateMachine) CurrentState() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// RegisterValidator adds a transition validator
func (sm *StateMachine) RegisterValidator(name string, validator TransitionValidator) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.validators[name] = validator
}

// GetConfig returns the state machine configuration
func (sm *StateMachine) GetConfig() *StateMachineConfig {
	return sm.config
}

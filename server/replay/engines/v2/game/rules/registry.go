package rules

import "fmt"

// Registry holds multiple game rulesets identified by mode names.
type Registry struct {
	modes map[string]*GameRules
}

// NewRegistry creates a new rules registry.
func NewRegistry() *Registry {
	return &Registry{
		modes: make(map[string]*GameRules),
	}
}

// Register adds a new ruleset to the registry.
func (r *Registry) Register(mode string, rules *GameRules) {
	r.modes[mode] = rules
}

// Get retrieves a ruleset by its mode name.
func (r *Registry) Get(mode string) (*GameRules, error) {
	rules, ok := r.modes[mode]
	if !ok {
		return nil, fmt.Errorf("ruleset %s not found", mode)
	}
	return rules, nil
}

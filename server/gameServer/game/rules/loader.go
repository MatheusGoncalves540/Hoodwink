package rules

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadRules reads and parses a YAML file to create a GameRules instance.
func LoadRules(path string) (*GameRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rules GameRules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, err
	}

	return &rules, nil
}

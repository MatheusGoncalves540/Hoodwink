package rules

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadRulesFromBytes interpreta YAML em memória e monta uma instância de GameRules.
// É o caminho puro do motor (sem arquivo nem log.Fatalf).
func LoadRulesFromBytes(data []byte) (*GameRules, error) {
	var rules GameRules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, err
	}
	return &rules, nil
}

// LoadRules lê e interpreta um arquivo YAML (fallback para testes/dev).
func LoadRules(path string) (*GameRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return LoadRulesFromBytes(data)
}

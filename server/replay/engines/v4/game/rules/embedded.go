package rules

import (
	"embed"
	"fmt"
	"strings"
)

// modesFS embute os YAML de regras no binário. O motor (WASM) carrega daqui,
// sem arquivo em disco e sem log.Fatalf.
//
//go:embed modes/*.yaml
var modesFS embed.FS

// LoadEmbeddedRegistry carrega todos os modos embutidos no binário e devolve o
// Registry pronto. Retorna erro (sem log.Fatalf) para o caminho do motor.
func LoadEmbeddedRegistry() (*Registry, error) {
	registry := NewRegistry()

	entries, err := modesFS.ReadDir("modes")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		mode := strings.TrimSuffix(entry.Name(), ".yaml")
		data, err := modesFS.ReadFile("modes/" + entry.Name())
		if err != nil {
			return nil, err
		}

		r, err := LoadRulesFromBytes(data)
		if err != nil {
			return nil, fmt.Errorf("failed to load rules %s: %w", mode, err)
		}
		if err := r.Validate(); err != nil {
			return nil, fmt.Errorf("invalid rules %s: %w", mode, err)
		}

		registry.Register(mode, r)
	}

	return registry, nil
}

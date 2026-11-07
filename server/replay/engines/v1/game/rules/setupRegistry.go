package rules

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func SetupRulesRegistry(dir string) *Registry {
	registry := NewRegistry()

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read rules directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		mode := strings.TrimSuffix(entry.Name(), ".yaml")

		r, err := LoadRules(filepath.Join(dir, entry.Name()))
		if err != nil {
			log.Fatalf("failed to load rules %s: %v", mode, err)
		}

		if err := r.Validate(); err != nil {
			log.Fatalf("invalid rules %s: %v", mode, err)
		}

		registry.Register(mode, r)
		if os.Getenv("DEBUG") == "true" {
			log.Println("✅ " + mode + " loaded successfully")
		}
	}

	return registry
}

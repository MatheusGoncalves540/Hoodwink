package utils

import (
	"regexp"
	"testing"
)

func TestGenerateNewIdFormato(t *testing.T) {
	id := GenerateNewId()
	if matched, _ := regexp.MatchString(`^[0-9a-f]{16}$`, id); !matched {
		t.Fatalf("GenerateNewId() = %q, esperado hex de 16 caracteres", id)
	}
}

func TestGenerateNewIdUnico(t *testing.T) {
	const n = 10000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := GenerateNewId()
		if _, dup := seen[id]; dup {
			t.Fatalf("colisão no id %q", id)
		}
		seen[id] = struct{}{}
	}
}

package utils

import (
	"regexp"
	"testing"
)

// 6 caracteres do alfabeto sem ambíguos (fora 0/O e 1/I/L).
var roomCodeFormat = regexp.MustCompile(`^[2-9A-HJ-NP-Z]{6}$`)

func TestGenerateNewIdFormato(t *testing.T) {
	for i := 0; i < 1000; i++ {
		if id := GenerateNewId(); !roomCodeFormat.MatchString(id) {
			t.Fatalf("GenerateNewId() = %q, esperado 6 caracteres do alfabeto sem ambíguos", id)
		}
	}
}

func TestGenerateNewIdUnico(t *testing.T) {
	// 1e3 em 32^6 possibilidades: uma colisão aqui seria bug, não azar.
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := GenerateNewId()
		if _, dup := seen[id]; dup {
			t.Fatalf("colisão no id %q", id)
		}
		seen[id] = struct{}{}
	}
}

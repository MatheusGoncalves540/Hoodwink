package http

import "testing"

func TestCanCreateRoom(t *testing.T) {
	cases := map[string]bool{
		"google": true,
		"":       true, // token antigo, sem provider
		"guest":  false,
		"local":  false, // nome antigo do login de visitante
	}

	for provider, want := range cases {
		if got := canCreateRoom(provider); got != want {
			t.Fatalf("provider %q: canCreateRoom = %v, esperado %v", provider, got, want)
		}
	}
}

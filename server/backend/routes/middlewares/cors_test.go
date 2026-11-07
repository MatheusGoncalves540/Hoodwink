package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func corsRequest(t *testing.T, origin string) *httptest.ResponseRecorder {
	t.Helper()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/alive", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rr := httptest.NewRecorder()
	CORSMiddleware(next).ServeHTTP(rr, req)
	return rr
}

func TestCORSMiddlewareAllowlist(t *testing.T) {
	t.Setenv("FRONTEND_URL", "https://hoodwink.hermis.com.br")
	t.Setenv("GAME_SERVER_URL", "https://hoodwink.hermis.com.br/gs")
	t.Setenv("CORS", "false")

	t.Run("origem permitida tem credenciais", func(t *testing.T) {
		rr := corsRequest(t, "https://hoodwink.hermis.com.br")
		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://hoodwink.hermis.com.br" {
			t.Fatalf("Allow-Origin = %q, esperado origem permitida", got)
		}
		if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Fatalf("Allow-Credentials = %q, esperado true", got)
		}
		if got := rr.Header().Get("Vary"); !strings.Contains(got, "Origin") {
			t.Fatalf("Vary = %q, esperado conter Origin", got)
		}
	})

	t.Run("origem estranha é bloqueada", func(t *testing.T) {
		rr := corsRequest(t, "https://evil.example.com")
		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("Allow-Origin = %q, esperado vazio", got)
		}
		if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "" {
			t.Fatalf("Allow-Credentials = %q, esperado vazio", got)
		}
	})

	t.Run("origem vazia não ganha cabeçalho", func(t *testing.T) {
		rr := corsRequest(t, "")
		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("Allow-Origin = %q, esperado vazio", got)
		}
	})
}

func TestCORSMiddlewareDevAllowAnyWithoutCredentials(t *testing.T) {
	t.Setenv("FRONTEND_URL", "https://hoodwink.hermis.com.br")
	t.Setenv("GAME_SERVER_URL", "https://hoodwink.hermis.com.br/gs")
	t.Setenv("CORS", "true")

	rr := corsRequest(t, "http://localhost:5173")
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Allow-Origin = %q, esperado origem dev", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Allow-Credentials = %q, esperado vazio no modo dev", got)
	}
}

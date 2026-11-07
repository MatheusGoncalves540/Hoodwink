// middlewares/cors.go
package middlewares

import (
	"net/http"
	"os"
)

// CORSMiddleware aplica a política de origem do backend.
// - Sempre usa allowlist (FRONTEND_URL e GAME_SERVER_URL).
// - Allow-Credentials só para origem da allowlist.
// - CORS=="true" é opt-in explícito de dev: libera qualquer origem, SEM credenciais.
// - Vary: Origin evita cache errado entre origens diferentes.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		frontendURL := os.Getenv("FRONTEND_URL")
		gameServerURL := os.Getenv("GAME_SERVER_URL")
		allowAny := os.Getenv("CORS") == "true"

		allowed := false
		allowCredentials := false

		if origin != "" {
			if allowAny {
				// Dev opt-in: qualquer origem, mas nunca com credenciais.
				allowed = true
			} else if origin == frontendURL || origin == gameServerURL {
				allowed = true
				allowCredentials = true
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			if allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

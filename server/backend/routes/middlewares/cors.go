// middlewares/cors.go
package middlewares

import (
	"net/http"
	"os"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		// Permitir qualquer origem (modo dev)
		if os.Getenv("CORS") == "false" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// validar origem
			if origin == os.Getenv("FRONTEND_URL") ||
				origin == os.Getenv("GAME_SERVER_URL") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

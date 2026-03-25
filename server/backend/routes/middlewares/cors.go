// middlewares/cors.go
package middlewares

import (
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("CORS") == "false" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
		} else {
			// Configura cabeçalhos CORS
			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("GAME_SERVER_URL"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Responde a preflight diretamente
		if r.Method == http.MethodOptions {
			utils.SendError(w, "No Content", http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

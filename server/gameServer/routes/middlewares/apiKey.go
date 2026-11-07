package middlewares

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// APIKeyMiddleware valida o header x-api-key em tempo constante contra
// BACKEND_API_KEY. Chave ausente (na requisição ou na configuração do servidor)
// é sempre rejeitada com 401 (fail closed).
func APIKeyMiddleware(next http.Handler) http.Handler {
	expected := os.Getenv("BACKEND_API_KEY")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := r.Header.Get("x-api-key")
		if provided == "" {
			utils.SendError(w, "API Key não fornecida", http.StatusUnauthorized)
			return
		}
		if expected == "" {
			utils.SendError(w, "API Key não configurada no servidor", http.StatusUnauthorized)
			return
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			utils.SendError(w, "API Key inválida", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

package middlewares

import (
	"context"
	"crypto/subtle"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

type contextKey string

const ApiKeyContextKey contextKey = "apiKey"

// APIKeyMiddleware valida o header x-api-key em tempo constante e injeta o nome
// do sistema externo no contexto. Sem a chave ou com chave diferente responde 401.
func APIKeyMiddleware(next http.Handler, handler handlers.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("x-api-key")
		if authHeader == "" {
			utils.SendError(w, "API Key não fornecida", http.StatusUnauthorized)
			return
		}

		if subtle.ConstantTimeCompare([]byte(authHeader), []byte(handler.APIKey)) != 1 {
			utils.SendError(w, "API Key inválida", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ApiKeyContextKey, "internal")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

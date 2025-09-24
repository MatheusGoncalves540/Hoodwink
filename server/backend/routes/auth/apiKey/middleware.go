package apiKey

import (
	"context"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

type contextKey string

const ApiKeyContextKey contextKey = "apiKey"

// APIKeyMiddleware verifica o header x-api-key, valida, e injeta o nome do sistema externo no contexto
func APIKeyMiddleware(next http.Handler, handler handlers.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("x-api-key")
		if authHeader == "" {
			utils.SendError(w, "API Key não fornecida", http.StatusBadRequest)
			return
		}

		systemName, err := handler.DBService.VerifyApiKey(authHeader)
		if err != nil {
			utils.SendError(w, "API Key inválida", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ApiKeyContextKey, systemName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

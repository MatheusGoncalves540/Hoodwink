package auth

import (
	"context"
	"net/http"
)

type contextKey string

const userEmailKey contextKey = "userEmail"

// JWTMiddleware é um middleware que verifica se o token JWT está presente e válido no header da solicitação.
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}

		email, err := ValidateJWT(authHeader)
		if err != nil {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		// Passa o e-mail do usuário no contexto
		ctx := context.WithValue(r.Context(), userEmailKey, email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

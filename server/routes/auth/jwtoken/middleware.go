package jwtoken

import (
	"context"
	"net/http"
)

type contextKey string

const UserContextKey contextKey = "user"

// JWTMiddleware verifica o cookie JWT, valida, e injeta o struct UserClaims no contexto
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			http.Error(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}

		user, err := ValidateJWT(cookie.Value)
		if err != nil {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

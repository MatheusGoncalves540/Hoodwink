package jwtoken

import (
	"context"
	"net/http"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

// JWTMiddleware verifica o header Authorization, valida, e injeta o struct UserClaims no contexto
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.SendError(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := ValidateJWT(tokenStr)
		if err != nil {
			utils.SendError(w, "Token inválido", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

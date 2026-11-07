package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

// JWTMiddleware verifica o header Authorization, valida, e injeta o struct UserClaims no contexto
func JWTMiddleware(next http.Handler, h *handlers.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.SendError(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := h.JWTService.ValidateJWT(tokenStr)
		if err != nil {
			utils.SendError(w, "Token inválido", http.StatusUnauthorized)
			return
		}
		if info, ok := r.Context().Value(reqInfoContextKey).(*ReqInfo); ok {
			info.UserID = user.Id
			info.Username = user.Username
		}
		ctx := context.WithValue(r.Context(), handlers.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

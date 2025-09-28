package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/rHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// JWTBackendMiddleware verifica o header Authorization, valida, e injeta o struct UserClaims no contexto
func JWTBackendMiddleware(next http.Handler, h *rHandlers.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.SendError(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claimsRaw, err := h.JWTService.ParseToken(tokenStr, true)
		if err != nil {
			utils.SendError(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		claims := claimsRaw
		player := &endpointStructures.ClaimsBackend{
			Id:       claims["id"].(string),
			Username: claims["username"].(string),
			Provider: claims["provider"].(string),
			Email:    claims["email"].(string),
			Temp:     claims["temp"].(bool),
		}

		switch v := claims["exp"].(type) {
		case float64:
			player.Exp = int64(v)
		case int64:
			player.Exp = v
		}

		ctx := context.WithValue(r.Context(), contextKeys.PlayerContextKey, player)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

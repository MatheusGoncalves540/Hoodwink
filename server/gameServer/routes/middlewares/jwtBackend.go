package middlewares

import (
	"context"
	"net/http"
	"strings"

	httpHandlers "github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/http"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// JWTBackendMiddleware valida o token de login com o backend via
// /internal/validateToken e injeta ClaimsBackend{Id, Username} no contexto.
func JWTBackendMiddleware(next http.Handler, h *httpHandlers.HTTPHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.SendError(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		id, username, status := h.BackendService.ValidateToken(r.Context(), tokenStr)
		if status != http.StatusOK {
			if status == http.StatusServiceUnavailable {
				utils.SendError(w, "Servidor de login indisponível", http.StatusServiceUnavailable)
				return
			}
			utils.SendError(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		player := &endpointStructures.ClaimsBackend{
			Id:       id,
			Username: username,
		}

		if info, ok := r.Context().Value(reqInfoContextKey).(*ReqInfo); ok {
			info.UserID = id
			info.Username = username
		}

		ctx := context.WithValue(r.Context(), contextKeys.PlayerContextKey, player)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

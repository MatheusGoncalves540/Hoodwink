package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func (h *HTTPHandler) GetTicket(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomId := chi.URLParam(r, "RoomId")
		playerCtx := r.Context().Value(contextKeys.PlayerContextKey)
		playerClaims, ok := playerCtx.(*endpointStructures.ClaimsBackend)
		if !ok || playerClaims == nil {
			utils.SendError(w, "dados de player inválidos no token", http.StatusUnauthorized)
			return
		}

		// Valida se o player pode entrar na sala (regra local; o backend só revalida o token).
		isValid, err := h.RoomService.ValidatePlayerEntry(r, rdb, roomId, playerClaims.Id)
		if !isValid {
			utils.SendError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// O ticket é assinado pelo backend; reenviamos o token de login recebido.
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.SendError(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		ticket, status := h.BackendService.RequestTicket(r.Context(), tokenStr, roomId, utils.ClientIP(r), utils.UserAgent(r))
		if status != http.StatusOK {
			if status == http.StatusUnauthorized {
				utils.SendError(w, "Token inválido", http.StatusUnauthorized)
				return
			}
			utils.SendError(w, "Servidor de login indisponível", http.StatusServiceUnavailable)
			return
		}

		utils.LogDebug(fmt.Sprintf("Ticket gerado para player %s, sala %s", playerClaims.Id, roomId))
		utils.SendJSON(w, http.StatusOK, utils.APIResponse{
			Message: "Ticket gerado com sucesso",
			Data:    map[string]string{"ticket": ticket},
		})
	}
}

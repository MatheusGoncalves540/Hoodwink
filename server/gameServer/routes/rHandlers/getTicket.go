package rHandlers

import (
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func (h *Handler) GetTicket(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomId := chi.URLParam(r, "RoomId")
		playerCtx := r.Context().Value(contextKeys.PlayerContextKey)
		playerClaims, ok := playerCtx.(*endpointStructures.ClaimsBackend)
		if !ok || playerClaims == nil {
			utils.SendError(w, "dados de player inválidos no token", http.StatusUnauthorized)
			return
		}

		isValid, errMsg := h.RoomService.ValidatePlayerEntry(r, rdb, h.BackendService, roomId, playerClaims.Id)
		if !isValid {
			utils.SendError(w, errMsg, http.StatusNotFound)
			return
		}

		// Geração do ticket JWT
		ticket, err := h.JWTService.GenerateToken(playerClaims, roomId)
		if err != nil {
			utils.SendError(w, "Erro ao gerar ticket", http.StatusInternalServerError)
			return
		}

		utils.LogDebug(fmt.Sprintf("Ticket gerado para player %s, sala %s", playerClaims.Id, roomId))
		utils.SendJSON(w, http.StatusOK, utils.APIResponse{
			Message: "Ticket gerado com sucesso",
			Data:    map[string]string{"ticket": ticket},
		})
	}
}

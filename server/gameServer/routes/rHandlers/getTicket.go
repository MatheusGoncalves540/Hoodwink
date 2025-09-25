package rHandlers

import (
	"encoding/json"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func (h *Handler) GetTicket(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomId := chi.URLParam(r, "RoomId")

		var reqBody endpointStructures.GetTicketBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			utils.SendError(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		utils.LogDebug(roomId)
		utils.LogDebug(reqBody)

		isValid, errMsg := h.RoomService.ValidatePlayerEntry(r, rdb, h.BackendService, roomId, reqBody.PlayerId)
		if !isValid {
			utils.SendError(w, errMsg, http.StatusNotFound)
			return
		}

		// Geração do ticket JWT
		ticket, err := h.JWTService.GenerateToken(reqBody.PlayerId, roomId)
		if err != nil {
			utils.SendError(w, "Erro ao gerar ticket", http.StatusInternalServerError)
			return
		}

		utils.SendJSON(w, http.StatusOK, utils.APIResponse{
			Message: "Ticket gerado com sucesso",
			Data:    map[string]string{"ticket": ticket},
		})
	}
}

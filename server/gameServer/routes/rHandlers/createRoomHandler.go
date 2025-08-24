// gameServer/routes/rHandlers/createRoomHandler.go
package rHandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var reqBody endpointStructures.CreateRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if !utils.ValidateInfos(w, reqBody) {
		return
	}

	newRoomData, err := h.RoomService.CreateNewRoom(r, reqBody)
	if err != nil {
		utils.SendError(w, "Erro ao criar a sala", http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, http.StatusCreated, endpointStructures.CreateRoomResponse{
		RoomId: newRoomData.ID,
		Msg:    fmt.Sprintf("Sala '%s' criada com sucesso.", reqBody.RoomName),
	})
}

func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "RoomId")

	var reqBody endpointStructures.GetTicketBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	utils.PrintDebug(roomId)
	utils.PrintDebug(reqBody)

	//TODO: Validar se a sala existe e se o playerId é válido e apto para entrar na sala

	// Geração do ticket JWT
	ticket, err := h.JWTService.GenerateToken(reqBody.PlayerId, roomId)
	if err != nil {
		utils.PrintDebug(err)
		utils.SendError(w, "Erro ao gerar ticket", http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, http.StatusOK, utils.APIResponse{
		Message: "Ticket gerado com sucesso",
		Data:    map[string]string{"ticket": ticket},
	})
}

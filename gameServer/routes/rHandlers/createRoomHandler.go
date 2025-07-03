// gameServer/routes/rHandlers/createRoomHandler.go
package rHandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var reqBody endpointStructures.CreateRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if !utils.ValidateInfos(w, reqBody) {
		return
	}

	newRoomData, err := h.RoomService.CreateNewRoom(r, reqBody)
	if err != nil {
		http.Error(w, "Erro ao criar a sala", http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, http.StatusCreated, endpointStructures.CreateRoomResponse{
		RoomID: newRoomData.ID,
		Msg:    fmt.Sprintf("Sala '%s' criada com sucesso.", reqBody.RoomName),
	})
}

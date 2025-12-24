package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func (h *HTTPHandler) CreateCustomRoom(w http.ResponseWriter, r *http.Request) {
	var reqBody endpointStructures.CreateRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if !utils.ValidateInfos(w, reqBody) {
		return
	}

	newRoomData, err := h.RoomService.CreateNewRoom(r, reqBody, true)
	if err != nil {
		utils.SendError(w, "Erro ao criar a sala", http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, http.StatusCreated, endpointStructures.CreateRoomResponse{
		RoomId: newRoomData.ID,
		Msg:    fmt.Sprintf("Sala '%s' criada com sucesso.", reqBody.RoomName),
	})
}

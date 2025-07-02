package rHandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

type CreateRoomRequest struct {
	RoomName string  `json:"roomName" validate:"required,max=30,min=3"`
	Password *string `json:"password" validate:"omitempty,max=24"`
}

// Create a new room with the given data
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var reqBody CreateRoomRequest

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if !utils.ValidateInfos(w, reqBody) {
		return
	}

	// Aqui está tudo validado
	msg := fmt.Sprintf("Criou uma sala chamada %s", reqBody.RoomName)
	if reqBody.Password != nil {
		msg += " com senha definida."
	} else {
		msg += " sem senha."
	}

	utils.SendJSON(w, http.StatusCreated, map[string]string{"message": msg})
}

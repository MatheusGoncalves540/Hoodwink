package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// canCreateRoom: visitante (login só com nome) entra pelo código de uma sala já
// criada, mas não cria a própria. "local" é o nome antigo do mesmo login.
func canCreateRoom(provider string) bool {
	return provider != "guest" && provider != "local"
}

func (h *HTTPHandler) CreateCustomRoom(w http.ResponseWriter, r *http.Request) {
	playerCtx := r.Context().Value(contextKeys.PlayerContextKey)
	playerClaims, ok := playerCtx.(*endpointStructures.ClaimsBackend)
	if !ok || playerClaims == nil {
		utils.SendError(w, "dados de player inválidos no token", http.StatusUnauthorized)
		return
	}
	if !canCreateRoom(playerClaims.Provider) {
		utils.SendError(w, "Entre com sua conta para criar uma sala.", http.StatusForbidden)
		return
	}

	var reqBody endpointStructures.CreateRoomRequest

	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			utils.SendError(w, "Corpo da requisição muito grande (máx. 16 KB)", http.StatusRequestEntityTooLarge)
			return
		}
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if !utils.ValidateInfos(w, reqBody) {
		return
	}

	newRoomData, err := h.RoomService.CreateNewRoom(r, h.RulesRegistry, reqBody, true)
	if err != nil {
		utils.SendError(w, "Erro ao criar a sala", http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, http.StatusCreated, endpointStructures.CreateRoomResponse{
		RoomId: newRoomData.ID,
		Msg:    fmt.Sprintf("Sala '%s' criada com sucesso.", reqBody.RoomName),
	})
}

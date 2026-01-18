package protocolsValidation

import (
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateContestProtocol(roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	// Contest só pode ser usado durante um evento de carta
	if !strings.Contains(string(roomData.GameEvent.Type), "CARD_PLAYED") {
		utils.LogInvldPlyrReq("Contest só pode ser usado durante um evento de carta", playerPlay.PlayerId)
		return false
	}

	return true
}

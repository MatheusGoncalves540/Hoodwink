package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateTakeCoinsProtocol(roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Ação de coletar moedas só pode ser usada durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou coletar moedas fora do seu turno", playerPlay.PlayerId)
		return false
	}

	return true
}

package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

func ValidateTakeCoinsProtocol(roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		pure.LogInvldPlyrReq("Ação de coletar moedas só pode ser usada durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou coletar moedas fora do seu turno", playerPlay.PlayerId)
		return false
	}

	return true
}

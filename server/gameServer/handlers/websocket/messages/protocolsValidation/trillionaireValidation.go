package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateTrillionaireProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	// Trillionaire só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Trillionaire só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	return true
}

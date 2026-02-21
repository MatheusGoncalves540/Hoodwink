package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidatePoliticalProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	// Political só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Political só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		utils.LogInvldPlyrReq(err, playerPlay.PlayerId)
		return false
	}

	return true
}

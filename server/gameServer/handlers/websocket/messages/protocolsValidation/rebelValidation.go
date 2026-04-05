package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateRebelProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	generalRules, err := roomData.GetGeneralRules(registryRules)
	if err != nil {
		utils.LogError(err)
		return false
	}

	// Rebel só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Rebel só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou usar Rebel fora do seu turno", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		utils.LogInvldPlyrReq(err, playerPlay.PlayerId)
		return false
	}

	// verifica se a taxa mínima já foi atingida
	if roomData.Tax <= *generalRules.MinTax {
		utils.LogInvldPlyrReq("Taxa mínima atingida", playerPlay.PlayerId)
		return false
	}

	return true
}

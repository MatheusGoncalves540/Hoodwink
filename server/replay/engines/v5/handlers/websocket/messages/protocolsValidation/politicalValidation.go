package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

func ValidatePoliticalProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		pure.LogError(err)
		return false
	}

	generalRules, err := roomData.GetGeneralRules(registryRules)
	if err != nil {
		pure.LogError(err)
		return false
	}

	// Political só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		pure.LogInvldPlyrReq("Political só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou usar Political fora do seu turno", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		pure.LogInvldPlyrReq(err, playerPlay.PlayerId)
		return false
	}

	// verifica se a taxa máxima já foi atingida
	if generalRules.MaxTax != nil && roomData.Tax >= *generalRules.MaxTax {
		pure.LogInvldPlyrReq("Taxa máxima atingida", playerPlay.PlayerId)
		return false
	}

	return true
}

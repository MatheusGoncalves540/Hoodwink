package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateTricksterProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, tricksterPayload *structs.TricksterPayload) bool {
	// Trickster só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Trickster só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	targetPlayer, err := roomData.GetPlayer(tricksterPayload.TargetPlayer)
	if err != nil {
		utils.LogInvldPlyrReq("Jogador alvo não encontrado", playerPlay.PlayerId)
		return false
	}

	// verifica se player alvo do roubo tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(targetPlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		utils.LogInvldPlyrReq(err.Error(), playerPlay.PlayerId)
		return false
	}

	return true
}

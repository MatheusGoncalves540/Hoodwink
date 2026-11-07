package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

func ValidateTricksterProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, tricksterPayload *structs.TricksterPayload) bool {
	// Trickster só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		pure.LogInvldPlyrReq("Trickster só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou usar Trickster fora do seu turno", playerPlay.PlayerId)
		return false
	}

	if tricksterPayload.TargetPlayer == playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Trickster só pode roubar moedas de um adversário", playerPlay.PlayerId)
		return false
	}

	targetPlayer, err := roomData.GetPlayer(tricksterPayload.TargetPlayer)
	if err != nil || !targetPlayer.Alive {
		pure.LogInvldPlyrReq("Jogador alvo não encontrado", playerPlay.PlayerId)
		return false
	}

	// verifica se player alvo do roubo tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(targetPlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		pure.LogInvldPlyrReq(err.Error(), playerPlay.PlayerId)
		return false
	}

	return true
}

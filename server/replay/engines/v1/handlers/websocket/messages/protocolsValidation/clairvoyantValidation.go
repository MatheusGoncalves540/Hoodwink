package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

func ValidateClairvoyantProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, clairvoyantPayload structs.ClairvoyantPayload) bool {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		pure.LogError(err)
		return false
	}

	// Clairvoyant só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		pure.LogInvldPlyrReq("Clairvoyant só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou usar Clairvoyant fora do seu turno", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		pure.LogInvldPlyrReq(err, playerPlay.PlayerId)
		return false
	}

	if clairvoyantPayload.TargetPlayer == playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Clairvoyant só pode revelar a carta de um adversário", playerPlay.PlayerId)
		return false
	}

	targetPlayer, err := roomData.GetPlayer(clairvoyantPayload.TargetPlayer)
	if err != nil || !targetPlayer.Alive {
		pure.LogInvldPlyrReq("Jogador alvo inválido para o Clairvoyant", playerPlay.PlayerId)
		return false
	}

	card, err := targetPlayer.GetCardByIndex(clairvoyantPayload.TargetCardIndex)
	if err != nil || card.Dead {
		pure.LogInvldPlyrReq("Clairvoyant tentou revelar uma carta morta", playerPlay.PlayerId)
		return false
	}

	return true
}

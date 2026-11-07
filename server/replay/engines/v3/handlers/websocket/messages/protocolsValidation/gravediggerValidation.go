package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
)

func ValidateGravediggerProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		pure.LogError(err)
		return false
	}

	// Gravedigger só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		pure.LogInvldPlyrReq("Gravedigger só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou usar Gravedigger fora do seu turno", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		pure.LogInvldPlyrReq(err, playerPlay.PlayerId)
		return false
	}

	gravediggerPayload, ok := playerPlay.Payload.(structs.GravediggerPayload)
	if !ok {
		pure.LogInvldPlyrReq("payload does not match GravediggerPayload structure", playerPlay.PlayerId)
		return false
	}

	if gravediggerPayload.TargetPlayer == nil || *gravediggerPayload.TargetPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Gravedigger só pode ressuscitar uma carta do próprio jogador", playerPlay.PlayerId)
		return false
	}
	if gravediggerPayload.TargetCardIndex == nil {
		pure.LogInvldPlyrReq("targetCardIndex é obrigatório", playerPlay.PlayerId)
		return false
	}

	card, err := sourcePlayer.GetCardByIndex(*gravediggerPayload.TargetCardIndex)
	if err != nil || !card.Dead {
		pure.LogInvldPlyrReq("Gravedigger só pode ressuscitar uma carta morta da mão do jogador", playerPlay.PlayerId)
		return false
	}

	return true
}

package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/rooms"
)

func ValidateBasicKillProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, payload structs.AssassinPayload) bool {
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		pure.LogInvldPlyrReq("Ação de matar carta só pode ser usada durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou matar uma carta fora do seu turno", playerPlay.PlayerId)
		return false
	}

	if payload.TargetPlayer == nil || *payload.TargetPlayer == playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou matar uma carta de si mesmo com a ação básica", playerPlay.PlayerId)
		return false
	}

	targetPlayer, err := roomData.GetPlayer(*payload.TargetPlayer)
	if err != nil {
		pure.LogError(err)
		return false
	}

	card, err := targetPlayer.GetCardByIndex(*payload.TargetCardIndex)
	if err != nil || card.Dead {
		pure.LogInvldPlyrReq("Player tentou matar uma carta morta", playerPlay.PlayerId)
		return false
	}

	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		pure.LogError(err)
		return false
	}

	price, err := roomData.GetBasicKillCardCost(registryRules)
	if err != nil {
		pure.LogError(err)
		return false
	}
	if sourcePlayer.Coins < price {
		pure.LogInvldPlyrReq("Moedas insuficientes para matar uma carta", playerPlay.PlayerId)
		return false
	}

	return true
}

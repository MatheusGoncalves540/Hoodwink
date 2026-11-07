package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/rooms"
)

func ValidateSwapCardsProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, payload structs.SwapCardsPayload) bool {
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		pure.LogInvldPlyrReq("Ação de trocar cartas só pode ser usada durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou trocar cartas fora do seu turno", playerPlay.PlayerId)
		return false
	}

	if payload.TargetPlayer != nil && *payload.TargetPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou trocar cartas de outro jogador", playerPlay.PlayerId)
		return false
	}

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

	minCoins := 6
	if generalRules.BasicSwapMinCoins != nil {
		minCoins = *generalRules.BasicSwapMinCoins
	}
	if sourcePlayer.Coins < minCoins {
		pure.LogInvldPlyrReq("É preciso ter no mínimo 6 moedas para trocar cartas", playerPlay.PlayerId)
		return false
	}

	if payload.TargetCardIndex == nil {
		pure.LogInvldPlyrReq("targetCardIndex é obrigatório", playerPlay.PlayerId)
		return false
	}

	if _, err := sourcePlayer.GetCardByIndex(*payload.TargetCardIndex); err != nil {
		pure.LogInvldPlyrReq("targetCardIndex inválido", playerPlay.PlayerId)
		return false
	}

	if payload.UseOnTwoCards != nil && *payload.UseOnTwoCards {
		otherIndex := 1 - *payload.TargetCardIndex
		if _, err := sourcePlayer.GetCardByIndex(otherIndex); err != nil {
			pure.LogInvldPlyrReq("O jogador não possui duas cartas para trocar", playerPlay.PlayerId)
			return false
		}
	}

	return true
}

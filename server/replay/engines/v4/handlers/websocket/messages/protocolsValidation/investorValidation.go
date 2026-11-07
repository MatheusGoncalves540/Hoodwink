package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

func ValidateInvestorProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	// Investor só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		pure.LogInvldPlyrReq("Investor só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou usar Investor fora do seu turno", playerPlay.PlayerId)
		return false
	}

	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		pure.LogError(err)
		return false
	}

	// você investe 1 moeda da mão
	if sourcePlayer.Coins < 1 {
		pure.LogInvldPlyrReq("Moedas insuficientes para investir", playerPlay.PlayerId)
		return false
	}

	// O valor da carta representa o número máximo de moedas investidas (4 + taxas, limitado entre 2 e 6)
	investmentMax, err := roomData.GetCardValue(registryRules, playerPlay.Type, 0)
	if err != nil {
		pure.LogError(err)
		return false
	}

	if len(sourcePlayer.Investments) >= investmentMax {
		pure.LogInvldPlyrReq("Número máximo de investimentos ativos alcançado", playerPlay.PlayerId)
		return false
	}

	return true
}

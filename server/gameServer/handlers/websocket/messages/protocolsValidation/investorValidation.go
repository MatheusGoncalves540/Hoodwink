package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateInvestorProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		utils.LogError("Erro ao obter regras da carta Guardian: " + err.Error())
		return false
	}

	// Investor só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Investor só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	// Verifica se o jogador tem o máximo de investimentos ativos
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	investmentMax := *cardRules.FixedValue

	if len(sourcePlayer.Investments) >= investmentMax {
		utils.LogInvldPlyrReq("Número máximo de investimentos ativos alcançado", playerPlay.PlayerId)
		return false
	}

	return true
}

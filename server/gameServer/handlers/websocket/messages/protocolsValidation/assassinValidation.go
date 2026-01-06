package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateAssassinProtocol(roomData *roomStructs.Room, playerPlay *roomStructs.PlayerPlay, payload roomStructs.AssassinPayload, RegistryRules *rules.Registry) bool {
	cardRules, err := roomData.GetCardRules(RegistryRules, string(playerPlay.Type))
	if err != nil {
		utils.LogError("Erro ao obter regras da carta Assassin: " + err.Error())
		return false
	}

	// verifica se player tem moedas pra isso
	if roomData.Players[playerPlay.PlayerId].Coins >= *cardRules.Price {
		utils.LogInvldPlyrReq(playerPlay.PlayerId + " não tem moedas suficientes para jogar Assassin")
		return false
	}

	// verifica se a carta sendo jurada de morte, ja nao esta morta
	targetPlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}
	card, err := targetPlayer.GetCardByIndex(*payload.TargetCardIndex)
	if err != nil {
		utils.LogError(err)
		return false
	}
	if card.Dead {
		utils.LogError(err)
		return false
	}

	return true
}

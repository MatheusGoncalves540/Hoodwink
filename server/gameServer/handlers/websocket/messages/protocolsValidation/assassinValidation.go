package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateAssassinProtocol(roomData *rooms.Room, RegistryRules *rules.Registry, playerPlay *structs.PlayerPlay, payload structs.AssassinPayload) bool {
	cardRules, err := roomData.GetCardRules(RegistryRules, string(playerPlay.Type))
	if err != nil {
		utils.LogError("Erro ao obter regras da carta Assassin: " + err.Error())
		return false
	}
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	// Assassin só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Assassin só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	if sourcePlayer.Coins < *cardRules.Price {
		utils.LogInvldPlyrReq("player não tem moedas suficientes para jogar Assassin", playerPlay.PlayerId)
		return false
	}

	// verifica se o jogador está tentando jurar de morte uma carta dele mesmo
	if payload.TargetPlayer == &playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou jurar de morte uma carta dele mesmo", playerPlay.PlayerId)
		return false
	}

	// verifica se a carta sendo jurada de morte, ja nao esta morta
	targetPlayer, err := roomData.GetPlayer(*payload.TargetPlayer)
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
		utils.LogInvldPlyrReq("Player tentou jurar de morte uma carta já morta", playerPlay.PlayerId)
		return false
	}

	return true
}

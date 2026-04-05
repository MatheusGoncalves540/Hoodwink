package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateAssassinProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, payload structs.AssassinPayload) bool {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Assassin só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou usar Assassin fora do seu turno", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		utils.LogInvldPlyrReq(err, playerPlay.PlayerId)
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

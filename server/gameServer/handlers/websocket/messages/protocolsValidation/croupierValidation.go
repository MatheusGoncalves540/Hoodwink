package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateCroupierProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	// Croupier só pode ser usado durante o evento de espera de primeira ação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventWaitingFirstAction {
		utils.LogInvldPlyrReq("Croupier só pode ser usado durante o evento de espera de primeira ação", playerPlay.PlayerId)
		return false
	}

	if roomData.CurrentPlayer != playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou usar Croupier fora do seu turno", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, 0)
	if err != nil {
		utils.LogInvldPlyrReq(err, playerPlay.PlayerId)
		return false
	}

	croupierPayload, ok := playerPlay.Payload.(structs.CroupierPayload)
	if !ok {
		utils.LogInvldPlyrReq("payload does not match CroupierPayload structure", playerPlay.PlayerId)
		return false
	}

	if croupierPayload.TargetPlayer == nil || *croupierPayload.TargetPlayer != playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Croupier só pode trocar uma carta do próprio jogador", playerPlay.PlayerId)
		return false
	}
	if croupierPayload.TargetCardIndex == nil {
		utils.LogInvldPlyrReq("targetCardIndex é obrigatório", playerPlay.PlayerId)
		return false
	}
	if croupierPayload.UseOnTwoCards != nil && *croupierPayload.UseOnTwoCards {
		utils.LogInvldPlyrReq("Croupier só troca uma carta por vez", playerPlay.PlayerId)
		return false
	}

	// O Croupier pode descartar uma carta viva ou morta da própria mão
	if _, err := sourcePlayer.GetCardByIndex(*croupierPayload.TargetCardIndex); err != nil {
		utils.LogInvldPlyrReq("Índice de carta inválido para o Croupier", playerPlay.PlayerId)
		return false
	}

	return true
}

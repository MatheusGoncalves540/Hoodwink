package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateKamikazeProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, payload *structs.KamikazePayload) bool {
	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		utils.LogError("Erro ao obter regras da carta Kamikaze: " + err.Error())
		return false
	}

	if roomData.GameEvent == nil {
		utils.LogInvldPlyrReq("Kamikaze só pode ser usado durante o evento de carta morta ou de primeira ação", playerPlay.PlayerId)
		return false
	}

	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	switch roomData.GameEvent.Type {
	case structs.EventWaitingFirstAction:
		if roomData.CurrentPlayer != playerPlay.PlayerId {
			utils.LogInvldPlyrReq("Player tentou usar suícidio do Kamikaze fora do seu turno", playerPlay.PlayerId)
			return false
		}

		if cardRules.CanKillSelf == nil || !*cardRules.CanKillSelf {
			utils.LogInvldPlyrReq("Player tentou usar kamikaze para matar a própria carta na primeira ação, mas a carta não permite isso", playerPlay.PlayerId)
			return false
		}

		if payload.TargetAllyCardIndex == nil {
			utils.LogInvldPlyrReq("Player tentou usar kamikaze para matar a própria carta na primeira ação, mas não forneceu o índice da carta aliada", playerPlay.PlayerId)
			return false
		}

		targetAllyCard, err := sourcePlayer.GetCardByIndex(*payload.TargetAllyCardIndex)
		if err != nil || targetAllyCard.Dead || targetAllyCard.Name != structs.CardKamikaze {
			utils.LogInvldPlyrReq("Player tentou usar kamikaze para matar uma carta que não é um Kamikaze vivo", playerPlay.PlayerId)
			return false
		}

		payload.KilledHimSelf = true

	case structs.EventCardKilled:
		payloadMap, ok := roomData.GameEvent.Payload.(map[string]interface{})
		if !ok {
			utils.LogError("payload does not match map structure")
			return false
		}

		targetPlayerStr, ok := payloadMap["targetPlayer"].(string)
		if !ok {
			utils.LogInvldPlyrReq("targetPlayer not found or invalid in payload", playerPlay.PlayerId)
			return false
		}
		if targetPlayerStr != playerPlay.PlayerId {
			utils.LogInvldPlyrReq("Player tentou usar kamikaze em um evento que não é dele", playerPlay.PlayerId)
			return false
		}

		// A carta que morreu precisa ser um Kamikaze vivo do jogador.
		deadCardIndex, ok := payloadMap["targetCardIndex"].(float64)
		if !ok {
			utils.LogInvldPlyrReq("targetCardIndex not found or invalid in payload", playerPlay.PlayerId)
			return false
		}
		deadCard, err := sourcePlayer.GetCardByIndex(int(deadCardIndex))
		if err != nil || deadCard.Dead || deadCard.Name != structs.CardKamikaze {
			utils.LogInvldPlyrReq("A carta que morreu não é um Kamikaze vivo", playerPlay.PlayerId)
			return false
		}

		// Se a carta está protegida e o ataque é bloqueável, ela não morre de verdade:
		// o Kamikaze não deve ser acionado.
		cause, _ := payloadMap["cause"].(string)
		unblockable := structs.EffectCause(cause) == structs.EffectKamikaze ||
			structs.EffectCause(cause) == structs.EffectContestPenalty ||
			structs.EffectCause(cause) == structs.EffectGreed
		if deadCard.Protected && !unblockable {
			utils.LogInvldPlyrReq("O Kamikaze está protegido e não vai morrer com esse ataque", playerPlay.PlayerId)
			return false
		}

	default:
		utils.LogInvldPlyrReq("Kamikaze só pode ser usado durante o evento de carta morta ou de primeira ação", playerPlay.PlayerId)
		return false
	}

	// verifica se o jogador está tentando mirar em uma carta dele mesmo
	if payload.TargetPlayer == nil || *payload.TargetPlayer == playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou usar kamikaze em uma carta dele mesmo", playerPlay.PlayerId)
		return false
	}

	if payload.TargetCardIndex == nil {
		utils.LogInvldPlyrReq("targetCardIndex é obrigatório", playerPlay.PlayerId)
		return false
	}

	// verifica se a carta sendo mirada já não está morta
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
		utils.LogInvldPlyrReq("Player tentou usar kamikaze em uma carta já morta", playerPlay.PlayerId)
		return false
	}

	return true
}

package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateGuardianProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, payload *structs.GuardianPayload) bool {
	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		utils.LogError("Erro ao obter regras da carta Guardian: " + err.Error())
		return false
	}

	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	if roomData.GameEvent == nil {
		utils.LogInvldPlyrReq("Guardian só pode ser usado durante o evento de carta morta ou de primeira ação", playerPlay.PlayerId)
		return false
	}

	// Se usado na primeira ação do jogo, marca que o jogador matou a propria carta
	switch roomData.GameEvent.Type {
	case structs.EventWaitingFirstAction:
		// verifica se o jogador forneceu um índice válido da carta aliada
		targetAllyCard, err := sourcePlayer.GetCardByIndex(payload.TargetCardIndex)
		if err != nil {
			utils.LogError(err)
			return false
		}

		if targetAllyCard.Dead {
			utils.LogInvldPlyrReq("Player tentou usar Guardian para proteger uma carta na primeira ação, mas não forneceu um índice valido da carta", playerPlay.PlayerId)
			return false
		}
		if targetAllyCard.Protected {
			utils.LogInvldPlyrReq("Player tentou usar Guardian para proteger uma carta na primeira ação, mas a carta já está protegida", playerPlay.PlayerId)
			return false
		}
	case structs.EventCardKilled:
		payloadMap, ok := roomData.GameEvent.Payload.(map[string]interface{})
		if !ok {
			utils.LogError("payload does not match map structure")
			return false
		}

		// Extrair targetPlayer do mapa
		targetPlayerIdStr, ok := payloadMap["targetPlayer"].(string)
		if !ok {
			utils.LogInvldPlyrReq("targetPlayer não encontrado ou inválido no payload", playerPlay.PlayerId)
			return false
		}

		if targetPlayerIdStr != playerPlay.PlayerId {
			utils.LogInvldPlyrReq("Player tentou usar Guardian em um evento que não é dele", playerPlay.PlayerId)
			return false
		}

		cardKilledCause, ok := payloadMap["cause"].(string)
		if !ok {
			utils.LogError("Causa não validada para o uso de carta morta de Guardian")
			return false
		}

		// verifica se o evento de carta morta foi causado por ContestPenalty ou Kamikaze, caso tenha sido, o jogador não pode usar Guardian
		switch cardKilledCause {
		case string(structs.EffectContestPenalty):
			utils.LogInvldPlyrReq("Player tentou usar Guardian em um evento que não é bloqueado por Guardian", playerPlay.PlayerId)
			return false
		case string(structs.EffectKamikaze):
			utils.LogInvldPlyrReq("Player tentou usar Guardian em um evento que não é bloqueado por Guardian", playerPlay.PlayerId)
			return false
		}

	case structs.EventCardPlayedClairvoyant:
		if cardRules.CanDefendClairvoyant == nil || !*cardRules.CanDefendClairvoyant {
			utils.LogInvldPlyrReq("Player tentou usar Guardian para bloquear clairvoyant, mas a carta não permite isso", playerPlay.PlayerId)
			return false
		}

	default:
		utils.LogInvldPlyrReq("Guardian só pode ser usado durante o evento de carta morta, clairvoyant ou de primeira ação", playerPlay.PlayerId)
		return false
	}

	// verifica se o jogador está tentando proteger uma carta que não é dele
	if payload.TargetPlayer != playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou usar Guardian em uma carta que não é dele", playerPlay.PlayerId)
		return false
	}

	// se só tiver uma carta viva, o jogador pode usar Guardian por um desconto
	discount := 0
	aliveCards := sourcePlayer.GetAliveCardsIndexes()
	if len(aliveCards) == 1 {
		discount = 1
	}

	// verifica se player tem moedas pra isso, considerando o desconto caso só tenha uma carta viva
	err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, discount)
	if err != nil {
		utils.LogInvldPlyrReq(err, playerPlay.PlayerId)
		return false
	}

	return true
}

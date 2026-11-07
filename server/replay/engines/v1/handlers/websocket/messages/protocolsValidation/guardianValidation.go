package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

func ValidateGuardianProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, payload *structs.GuardianPayload) bool {
	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		pure.LogError("Erro ao obter regras da carta Guardian: " + err.Error())
		return false
	}

	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		pure.LogError(err)
		return false
	}

	if roomData.GameEvent == nil {
		pure.LogInvldPlyrReq("Guardian só pode ser usado durante o evento de carta morta, clairvoyant ou de primeira ação", playerPlay.PlayerId)
		return false
	}

	if payload.TargetPlayer != playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou usar Guardian em uma carta que não é dele", playerPlay.PlayerId)
		return false
	}

	targetAllyCard, err := sourcePlayer.GetCardByIndex(payload.TargetCardIndex)
	if err != nil {
		pure.LogError(err)
		return false
	}
	if targetAllyCard.Dead {
		pure.LogInvldPlyrReq("Player tentou usar Guardian em uma carta morta", playerPlay.PlayerId)
		return false
	}

	protectMode := ""

	switch roomData.GameEvent.Type {
	case structs.EventWaitingFirstAction:
		if roomData.CurrentPlayer != playerPlay.PlayerId {
			pure.LogInvldPlyrReq("Player tentou usar Guardian fora do seu turno", playerPlay.PlayerId)
			return false
		}
		protectMode = "future"

	case structs.EventCardKilled:
		payloadMap, ok := roomData.GameEvent.Payload.(map[string]interface{})
		if !ok {
			pure.LogError("payload does not match map structure")
			return false
		}

		targetPlayerIdStr, ok := payloadMap["targetPlayer"].(string)
		if !ok {
			pure.LogInvldPlyrReq("targetPlayer não encontrado ou inválido no payload", playerPlay.PlayerId)
			return false
		}
		if targetPlayerIdStr != playerPlay.PlayerId {
			pure.LogInvldPlyrReq("Player tentou usar Guardian em um evento que não é dele", playerPlay.PlayerId)
			return false
		}

		killedCardIndex, ok := payloadMap["targetCardIndex"].(float64)
		if !ok || int(killedCardIndex) != payload.TargetCardIndex {
			pure.LogInvldPlyrReq("Guardian deve ser usada na carta que está sofrendo o ataque", playerPlay.PlayerId)
			return false
		}

		cardKilledCause, ok := payloadMap["cause"].(string)
		if !ok {
			pure.LogError("Causa não validada para o uso de carta morta de Guardian")
			return false
		}

		// Kamikaze, contestação e perda de carta por limite de moedas não podem ser impedidos
		switch structs.EffectCause(cardKilledCause) {
		case structs.EffectContestPenalty, structs.EffectKamikaze, structs.EffectGreed:
			pure.LogInvldPlyrReq("Player tentou usar Guardian em um evento que não é bloqueado por Guardian", playerPlay.PlayerId)
			return false
		}

		// A proteção existente já bloqueia o ataque; não é necessário gastar outra Guardian
		if targetAllyCard.Protected {
			pure.LogInvldPlyrReq("A carta já está protegida contra o próximo ataque", playerPlay.PlayerId)
			return false
		}

		protectMode = "death"

	case structs.EventCardPlayedClairvoyant:
		if cardRules.CanDefendClairvoyant == nil || !*cardRules.CanDefendClairvoyant {
			pure.LogInvldPlyrReq("Player tentou usar Guardian para bloquear clairvoyant, mas a carta não permite isso", playerPlay.PlayerId)
			return false
		}

		payloadMap, ok := roomData.GameEvent.Payload.(map[string]interface{})
		if !ok {
			pure.LogError("payload does not match map structure")
			return false
		}
		targetPlayerIdStr, _ := payloadMap["targetPlayer"].(string)
		targetCardIndex, _ := payloadMap["targetCardIndex"].(float64)
		if targetPlayerIdStr != playerPlay.PlayerId || int(targetCardIndex) != payload.TargetCardIndex {
			pure.LogInvldPlyrReq("Guardian deve ser usada na própria carta que está sendo revelada", playerPlay.PlayerId)
			return false
		}
		protectMode = "sight"

	default:
		pure.LogInvldPlyrReq("Guardian só pode ser usado durante o evento de carta morta, clairvoyant ou de primeira ação", playerPlay.PlayerId)
		return false
	}

	// verifica se player tem moedas pra isso
	cardPrice, err := roomData.GetGuardianValue(registryRules, protectMode)
	if err != nil {
		pure.LogError(err)
		return false
	}
	if sourcePlayer.Coins < cardPrice {
		pure.LogInvldPlyrReq("Moedas insuficientes para usar Guardian", playerPlay.PlayerId)
		return false
	}

	return true
}

package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateKamikazeProtocol(roomData *rooms.Room, RegistryRules *rules.Registry, playerPlay *structs.PlayerPlay, payload *structs.KamikazePayload) bool {
	cardRules, err := roomData.GetCardRules(RegistryRules, string(playerPlay.Type))
	if err != nil {
		utils.LogError("Erro ao obter regras da carta Kamikaze: " + err.Error())
		return false
	}

	// Se usado na primeira ação do jogo, marca que o jogador matou a propria carta
	switch roomData.GameEvent.Type {
	case structs.EventWaitingFirstAction:
		if cardRules.CanKillSelf == nil || !*cardRules.CanKillSelf {
			utils.LogInvldPlyrReq("Player tentou usar kamikaze para matar a própria carta na primeira ação, mas a carta não permite isso", playerPlay.PlayerId)
			return false
		}

		// verifica se o jogador forneceu um índice válido da carta aliada
		player, _ := roomData.GetPlayer(playerPlay.PlayerId)
		targetAllyCard, _ := player.GetCardByIndex(*payload.TargetAllyCardIndex)

		if targetAllyCard.Dead {
			// Verifica se o índice da carta aliada é válido e não está morta
			utils.LogInvldPlyrReq("Player tentou usar kamikaze para matar a própria carta na primeira ação, mas não forneceu um índice valido da carta aliada", playerPlay.PlayerId)
			return false
		}
		payload.KilledHimSelf = true
	case structs.EventCardKilled:
		payloadMap, ok := roomData.GameEvent.Payload.(map[string]interface{})
		if !ok {
			utils.LogError("payload does not match map structure")
			return false
		}

		// Extrair targetPlayer do mapa
		targetPlayerStr, ok := payloadMap["targetPlayer"].(string)
		if !ok {
			utils.LogInvldPlyrReq("targetPlayer not found or invalid in payload", playerPlay.PlayerId)
			return false
		}

		if targetPlayerStr != playerPlay.PlayerId {
			utils.LogInvldPlyrReq("Player tentou usar kamikaze em um evento que não é dele", playerPlay.PlayerId)
			return false
		}
	default:
		utils.LogInvldPlyrReq("Kamikaze só pode ser usado durante o evento de carta morta ou de primeira ação", playerPlay.PlayerId)
		return false
	}

	// verifica se o jogador está tentando jurar de morte uma carta dele mesmo
	if payload.TargetPlayer == &playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou usar kamikaze em uma carta dele mesmo", playerPlay.PlayerId)
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
		utils.LogInvldPlyrReq("Player tentou usar kamikaze em uma carta já morta", playerPlay.PlayerId)
		return false
	}

	return true
}

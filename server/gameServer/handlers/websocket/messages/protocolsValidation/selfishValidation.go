package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateSelfishProtocol(roomData *rooms.Room, registryRules *rules.Registry, playerPlay *structs.PlayerPlay) bool {
	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		utils.LogError("Erro ao obter regras da carta Selfish: " + err.Error())
		return false
	}

	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	if roomData.GameEvent == nil {
		utils.LogInvldPlyrReq("Selfish só pode ser usado em reação a um evento", playerPlay.PlayerId)
		return false
	}

	// Egoísta só pode ser usada fora do próprio turno
	if roomData.CurrentPlayer == playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Selfish só pode ser usado fora do seu turno", playerPlay.PlayerId)
		return false
	}

	// Não é possível impedir a si mesmo
	if roomData.GameEvent.PlayerID == playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Selfish não pode ser usado contra a própria jogada", playerPlay.PlayerId)
		return false
	}

	payloadMap, ok := roomData.GameEvent.Payload.(map[string]interface{})
	if !ok {
		utils.LogError("payload does not match map structure")
		return false
	}

	// Custo padrão: 1 moeda fixa (negar moedas)
	cardPrice := 0
	if cardRules.FixedValue != nil {
		cardPrice = *cardRules.FixedValue
	}

	switch roomData.GameEvent.Type {
	case structs.EventEarnCoins:
		// A receita passiva do Investidor e a ação básica de coletar não podem ser impedidas.
		cause, _ := payloadMap["cause"].(string)
		if structs.EffectCause(cause) != structs.EffectTrillionaire {
			utils.LogInvldPlyrReq("Selfish só pode impedir o recebimento de moedas de cartas", playerPlay.PlayerId)
			return false
		}

		earnedCoins, ok := payloadMap["earnedCoins"].(float64)
		if !ok || earnedCoins <= 0 {
			utils.LogInvldPlyrReq("Quantia inválida para usar Selfish", playerPlay.PlayerId)
			return false
		}

		// Só pode impedir o recebimento de (2 + taxas) moedas ou mais
		thresholdBase := 2
		if cardRules.CoinDenialThreshold != nil {
			thresholdBase = *cardRules.CoinDenialThreshold
		}
		threshold := thresholdBase + roomData.Tax
		if int(earnedCoins) < threshold {
			utils.LogInvldPlyrReq("Quantia recebida abaixo do limite do Selfish", playerPlay.PlayerId)
			return false
		}

	case structs.EventStealCoins:
		lostCoins, ok := payloadMap["lostCoins"].(float64)
		if !ok || lostCoins <= 0 {
			utils.LogInvldPlyrReq("Quantia inválida para usar Selfish", playerPlay.PlayerId)
			return false
		}

		// O Trapaceiro recebe a metade maior; o Selfish só pode impedir se receber mais de 1 moeda.
		if utils.CeilDiv(int(lostCoins), 2) <= 1 {
			utils.LogInvldPlyrReq("Valor do roubo deve ser maior que 1 para usar Selfish", playerPlay.PlayerId)
			return false
		}

	case structs.EventInvestment:
		// negar investimento custa 1 moeda fixa

	case structs.EventChangedCard:
		cause, _ := payloadMap["cause"].(string)
		if structs.EffectCause(cause) != structs.EffectCroupier {
			utils.LogInvldPlyrReq("Selfish só pode impedir a troca de carta do Croupier", playerPlay.PlayerId)
			return false
		}
		cardPrice, err = roomData.GetCardValue(registryRules, playerPlay.Type, 0)
		if err != nil {
			utils.LogError(err)
			return false
		}

	case structs.EventRevivedCard:
		cause, _ := payloadMap["cause"].(string)
		if structs.EffectCause(cause) != structs.EffectGravedigger {
			utils.LogInvldPlyrReq("Selfish só pode impedir a ressurreição do Coveiro", playerPlay.PlayerId)
			return false
		}
		cardPrice, err = roomData.GetCardValue(registryRules, playerPlay.Type, 0)
		if err != nil {
			utils.LogError(err)
			return false
		}

	default:
		utils.LogInvldPlyrReq("Selfish só pode ser usado durante eventos de ganho de moeda, investimento, roubo, troca ou ressurreição de carta", playerPlay.PlayerId)
		return false
	}

	if sourcePlayer.Coins < cardPrice {
		utils.LogInvldPlyrReq("Moedas insuficientes para usar Selfish", playerPlay.PlayerId)
		return false
	}

	return true
}

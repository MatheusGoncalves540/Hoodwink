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
		utils.LogError("Erro ao obter regras da carta Guardian: " + err.Error())
		return false
	}

	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	// Selfish só pode ser usado durante eventos de ganho de moeda, investimento ou troca de carta
	if roomData.GameEvent == nil ||
		(roomData.GameEvent.Type != structs.EventEarnCoins &&
			roomData.GameEvent.Type != structs.EventInvestment &&
			roomData.GameEvent.Type != structs.EventChangedCard &&
			roomData.GameEvent.Type != structs.EventStealCoins) {
		utils.LogInvldPlyrReq("Selfish só pode ser usado durante eventos de ganho de moeda, investimento, roubo e troca de carta", playerPlay.PlayerId)
		return false
	}

	switch roomData.GameEvent.Type {
	case structs.EventChangedCard:
		// verifica se player tem moedas pra isso (calcula o custo com base no valor NÃO FIXO definido nas regras da carta)
		err = roomData.VerifyPlayerHasEnoughCoins(sourcePlayer, registryRules, playerPlay.Type, 0)
		if err != nil {
			utils.LogInvldPlyrReq(err, playerPlay.PlayerId)
			return false
		}
	case structs.EventStealCoins:
		// verifica se o valor do roubo é maior que 1 (caso for, pode usar selfish)
		earnedCoins := roomData.GameEvent.Payload.(structs.StealCoinsPayload).LostCoins / 2

		if earnedCoins <= 1 {
			utils.LogInvldPlyrReq("Valor do roubo deve ser maior que 1 para usar Selfish", playerPlay.PlayerId)
			return false
		}
	default:
		// verifica se player tem moedas pra isso (calcula o custo com base no valor FIXO definido nas regras da carta)
		if sourcePlayer.Coins < *cardRules.FixedValue {
			utils.LogInvldPlyrReq("Moedas insuficientes para usar Selfish", playerPlay.PlayerId)
			return false
		}
	}

	return true
}

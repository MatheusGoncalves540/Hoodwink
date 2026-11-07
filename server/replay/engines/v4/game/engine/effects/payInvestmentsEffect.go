package effects

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

// PayInvestmentsForPlayer paga a receita passiva do Investidor no início do turno do jogador
// e decrementa o tempo restante de cada moeda investida.
func PayInvestmentsForPlayer(registryRules *rules.Registry, roomData *rooms.Room, player *players.Player) {
	if player == nil || !player.Alive || !player.HasActiveInvestment() {
		return
	}

	generalRules, err := roomData.GetGeneralRules(registryRules)
	if err != nil {
		pure.LogError(err)
		return
	}

	payPlayerInvestments(roomData, player, *generalRules.MaxCoins)
	countdownInvestment(roomData, player, 1, *generalRules.MaxCoins)
}

// payPlayerInvestments paga a receita passiva do Investidor (metade dos investimentos, apenas moedas inteiras).
func payPlayerInvestments(roomData *rooms.Room, sourcePlayer *players.Player, maxCoins int) {
	earnedCoins := len(sourcePlayer.Investments) / 2

	// Com número ímpar de investimentos, alterna o pagamento da metade restante
	// para que o total recebido ao longo do tempo seja equivalente à metade.
	if len(sourcePlayer.Investments)%2 != 0 {
		if sourcePlayer.PendingInvestmentCoin {
			earnedCoins++
			sourcePlayer.PendingInvestmentCoin = false
		} else {
			sourcePlayer.PendingInvestmentCoin = true
		}
	}

	if earnedCoins <= 0 {
		return
	}

	if sourcePlayer.AddCoins(earnedCoins, maxCoins) {
		applyGreedCardLoss(roomData, sourcePlayer)
	}
}

// countdownInvestment decrementa os investimentos do jogador, devolvendo a moeda investida quando expira.
func countdownInvestment(roomData *rooms.Room, player *players.Player, amount int, maxCoins int) {
	for i := 0; i < len(player.Investments); i++ {
		player.Investments[i] -= amount
		if player.Investments[i] > 0 {
			continue
		}

		// Remove o investimento expirado
		player.Investments = append(player.Investments[:i], player.Investments[i+1:]...)
		i--

		// Devolve a moeda investida para a mão do jogador
		if player.AddCoins(1, maxCoins) {
			applyGreedCardLoss(roomData, player)
		}
	}

	if len(player.Investments) == 0 {
		player.PendingInvestmentCoin = false
	}
}

// applyGreedCardLoss aplica a perda de carta por ultrapassar o limite de moedas.
// Essa perda não pode ser impedida por nenhuma carta.
func applyGreedCardLoss(roomData *rooms.Room, player *players.Player) {
	aliveIndexes := player.GetAliveCardsIndexes()
	if len(aliveIndexes) == 0 {
		return
	}

	if err := player.KillCard(aliveIndexes[0]); err != nil {
		pure.LogError(err)
		return
	}
	roomData.VerifyIfPlayerDead(player)
}

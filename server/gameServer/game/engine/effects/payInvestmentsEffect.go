package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// PayInvestments paga os investimentos do jogador e gera as moedas correspondentes
// Chamado no início de cada novo turno. Essa função paga todos os investimentos de todos os players de uma vez
// TODO chamar essa função no início de cada novo turno (não rodada, mas sim novo turno, ou seja, depois de todos os players jogarem)
func PayInvestments(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room) {
	for _, player := range roomData.Players {
		if player.HasActiveInvestment() {
			payPlayerInvestments(ctx, rdb, registryRules, roomData, player)
			player.CountdownInvestment(1)
		}
	}
}

// payPlayerInvestments paga os investimentos de um jogador específico e gera as moedas correspondentes
func payPlayerInvestments(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, soucePlayer *players.Player) {
	// Calcula as moedas a serem ganhas com base no número de investimentos ativos
	earnedCoins := len(soucePlayer.Investments) / 2

	// Caso o número de investimentos seja ímpar, o jogador tem um investimento pendente que ainda não gerou moedas
	// então marca o investimento pendente e só gera as moedas na próxima vez que pagar os investimentos
	if len(soucePlayer.Investments)%2 != 0 {
		if soucePlayer.PendingInvestmentCoin == true {
			// se já havia um investimento pendente, gera a moeda correspondente e remove o status de investimento pendente
			earnedCoins++
			soucePlayer.PendingInvestmentCoin = false
		} else {
			// se não havia um investimento pendente, marca o investimento atual como pendente para gerar a moeda na próxima vez
			soucePlayer.PendingInvestmentCoin = true
		}
	}

	earnCoinsPayload := structs.NewEarnCoinsPayload(string(structs.EffectInvestor), earnedCoins, nil)

	investmentsEffect := structs.NewEffect(structs.EffectInvestor, soucePlayer.Id, earnCoinsPayload)

	err := EarnCoinsAnnouncer(ctx, rdb, registryRules, roomData, investmentsEffect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

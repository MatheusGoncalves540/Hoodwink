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
func PayInvestments(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room) {
	for _, player := range roomData.Players {
		if player.HasActiveInvestment() {
			payPlayerInvestments(ctx, rdb, roomData, player)
			countdownInvestment(ctx, rdb, registryRules, roomData, player, 1)
		}
	}
}

// payPlayerInvestments paga os investimentos de um jogador específico e gera as moedas correspondentes
func payPlayerInvestments(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, soucePlayer *players.Player) {
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

	err := EarnCoinsAnnouncer(ctx, rdb, roomData, investmentsEffect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

// countdownInvestment decrementa os investimentos do jogador, removendo os que chegam a zero (chamar a cada rodada para atualizar os investimentos)
func countdownInvestment(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, player *players.Player, amount int) {
	generalRules, err := roomData.GetGeneralRules(registryRules)
	if err != nil {
		utils.LogError(err)
		return
	}

	// Loop para percorrer os investimentos do jogador
	for i := 0; i < len(player.Investments); i++ {
		// Decrementa o investimento pelo valor do amount
		player.Investments[i] -= amount
		// Verifica se o investimento expirou
		if player.Investments[i] <= 0 {
			// Remove o investimento expirado
			player.Investments = append(player.Investments[:i], player.Investments[i+1:]...)
			// Decrementa o índice para verificar o próximo investimento corretamente
			i--

			// Gera as moedas correspondentes ao investimento expirado
			// adiciona as moedas ao jogador e retorna breakLimit, que indica se o limite de moedas será ultrapassado
			breakLimit := player.AddCoins(1, *generalRules.MaxCoins)

			if breakLimit {
				// se o limite de moedas for ultrapassado, mata a carta viva de menor índice
				aliveIndexes := player.GetAliveCardsIndexes()

				greedPayload := structs.NewKillCardPayload(string(structs.EffectGreed), &player.Id, &aliveIndexes[0])

				killEffect := structs.Effect{
					Cause:        structs.EffectGreed,
					SourcePlayer: player.Id,
					Payload:      greedPayload,
				}

				KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect)
			}
		}
	}
}

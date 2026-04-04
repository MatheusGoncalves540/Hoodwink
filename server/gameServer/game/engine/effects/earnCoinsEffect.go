package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func EarnCoinsEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	earnCoinsPayload, ok := effect.Payload.(structs.EarnCoinsPayload)
	if !ok {
		utils.LogError("payload inválido para EffectEarnCoins")
		return
	}

	generalRules, err := roomData.GetGeneralRules(registryRules)
	if err != nil {
		utils.LogError(err)
		return
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// adiciona as moedas ao jogador e retorna breakLimit, que indica se o limite de moedas será ultrapassado
	breakLimit := sourcePlayer.AddCoins(earnCoinsPayload.EarnedCoins, *generalRules.MaxCoins)

	if breakLimit {
		// se o limite de moedas for ultrapassado, mata a carta viva de menor índice
		aliveIndexes := sourcePlayer.GetAliveCardsIndexes()

		greedPayload := structs.NewKillCardPayload(string(structs.EffectGreed), &sourcePlayer.Id, &aliveIndexes[0])

		killEffect := structs.Effect{
			Cause:        structs.EffectGreed,
			SourcePlayer: sourcePlayer.Id,
			Payload:      greedPayload,
		}

		KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect)
	}
}

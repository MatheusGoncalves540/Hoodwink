package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func StealCoinsEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	stealCoinsPayload, ok := effect.Payload.(structs.StealCoinsPayload)
	if !ok {
		utils.LogError("payload inválido para StealCoinsPayload")
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
	targetPlayer, err := roomData.GetPlayer(*stealCoinsPayload.TargetPlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// o Trapaceiro retira a quantia do alvo e recebe a metade maior em caso de valor ímpar
	lostCoins := stealCoinsPayload.LostCoins
	if lostCoins > targetPlayer.Coins {
		lostCoins = targetPlayer.Coins
	}
	earnedCoins := utils.CeilDiv(lostCoins, 2)

	// subtrai as moedas do jogador
	targetPlayer.RemoveCoins(lostCoins)

	// adiciona as moedas ao jogador
	breakLimit := sourcePlayer.AddCoins(earnedCoins, *generalRules.MaxCoins)

	if breakLimit {
		// se o limite de moedas for ultrapassado, mata a carta viva de menor índice
		aliveIndexes := sourcePlayer.GetAliveCardsIndexes()
		if len(aliveIndexes) == 0 {
			return
		}

		greedPayload := structs.NewKillCardPayload(string(structs.EffectGreed), &sourcePlayer.Id, &aliveIndexes[0])

		killEffect := structs.Effect{
			Cause:        structs.EffectGreed,
			SourcePlayer: sourcePlayer.Id,
			Payload:      greedPayload,
		}

		KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect)
	}
}

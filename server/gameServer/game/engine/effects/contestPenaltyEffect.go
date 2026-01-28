package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ContestPenaltyEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	contestPenaltyPayload, ok := effect.Payload.(structs.ContestPenaltyPayload)
	if !ok {
		utils.LogError("payload inválido para EffectContestPenalty")
		return
	}

	// obtém o jogador que iniciou contestou
	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// obtém o jogador que está sendo contestado
	contestedPlayer, err := roomData.GetPlayer(contestPenaltyPayload.ContestedPlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	if contestPenaltyPayload.HasCard {
		// mata uma carta aleatória do jogador que iniciou o contestamento

		// Seleciona uma carta viva aleatória do jogador que iniciou o contestamento
		possibleIndexes := sourcePlayer.GetAliveCardsIndexes()
		indexChosen := utils.GetRandomElementFromSlice(possibleIndexes)

		killPayload := structs.NewKillCardPayload(string(structs.EffectContestPenalty), &effect.SourcePlayer, &indexChosen)

		killEffect := structs.NewEffect(structs.EffectContestPenalty, contestedPlayer.Id, killPayload)

		err := KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect)
		if err != nil {
			utils.LogError(err)
			return
		}
	} else {
		// mata uma carta aleatória do jogador que foi contestado

		// Seleciona uma carta viva aleatória do jogador que foi contestado
		possibleIndexes := contestedPlayer.GetAliveCardsIndexes()
		indexChosen := utils.GetRandomElementFromSlice(possibleIndexes)

		killPayload := structs.NewKillCardPayload(string(structs.EffectContestPenalty), &contestedPlayer.Id, &indexChosen)

		killEffect := structs.NewEffect(structs.EffectContestPenalty, effect.SourcePlayer, killPayload)

		err := KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect)
		if err != nil {
			utils.LogError(err)
			return
		}
	}
}

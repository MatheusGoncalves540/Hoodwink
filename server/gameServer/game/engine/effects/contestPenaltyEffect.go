package effects

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ContestPenaltyEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, rawEffect structs.Effect) {
	// extrai o payload do efeito
	raw, ok := rawEffect.Payload.(map[string]interface{})
	if !ok {
		utils.LogError("payload inválido")
		return
	}

	b, _ := json.Marshal(raw)

	var contestPenaltyPayload structs.ContestPenaltyPayload
	if err := json.Unmarshal(b, &contestPenaltyPayload); err != nil {
		utils.LogError(err)
		return
	}

	// obtém o jogador que iniciou contestou
	sourcePlayer, err := roomData.GetPlayer(rawEffect.SourcePlayer)
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

		effect := structs.Effect{
			Cause:        structs.EffectContestPenalty,
			SourcePlayer: contestedPlayer.Id,
			Payload: structs.KillCardPayload{
				Cause:           string(structs.EffectContestPenalty),
				TargetPlayer:    &rawEffect.SourcePlayer,
				TargetCardIndex: &indexChosen,
			},
		}

		err := KillCard(ctx, rdb, registryRules, roomData, effect)
		if err != nil {
			utils.LogError(err)
			return
		}
	} else {
		// mata uma carta aleatória do jogador que foi contestado

		// Seleciona uma carta viva aleatória do jogador que foi contestado
		possibleIndexes := contestedPlayer.GetAliveCardsIndexes()
		indexChosen := utils.GetRandomElementFromSlice(possibleIndexes)

		effect := structs.Effect{
			Cause:        structs.EffectContestPenalty,
			SourcePlayer: rawEffect.SourcePlayer,
			Payload: structs.KillCardPayload{
				Cause:           string(structs.EffectContestPenalty),
				TargetPlayer:    &contestedPlayer.Id,
				TargetCardIndex: &indexChosen,
			},
		}

		err := KillCard(ctx, rdb, registryRules, roomData, effect)
		if err != nil {
			utils.LogError(err)
			return
		}
	}
}

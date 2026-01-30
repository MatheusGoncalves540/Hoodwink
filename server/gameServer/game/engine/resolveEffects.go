package engine

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func resolveNextEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room) {
	effectDto, ok := roomData.PopLastPendingEffect()
	if !ok {
		utils.LogError(fmt.Errorf("no pending effects to resolve"))
		return
	}

	effect, err := BuildEffect(effectDto)
	if err != nil {
		utils.LogError(err)
		return
	}

	switch effect.Cause {
	case structs.EffectCardKilled:
		effects.KillCard(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectEarnCoins:
		effects.EarnCoins(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectContest:
		effects.ContestEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectContestPenalty:
		effects.ContestPenaltyEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectAssassin:
		effects.AssassinEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectKamikaze:
		effects.KamikazeEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectTrillionaire:
		effects.TrillionaireEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectPolitical:
		effects.PoliticalEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectRebel:
		effects.RebelEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectClairvoyant:
		effects.ClairvoyantEffect(ctx, rdb, registryRules, roomData, effect)
	}

	roomData.SaveRoom(ctx, rdb)
	roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{})
}

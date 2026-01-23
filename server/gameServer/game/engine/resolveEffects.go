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

func resolveNextEffect(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry, roomData *rooms.Room) {
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
	case structs.EffectContest:
		effects.ContestEffect(ctx, rdb, RegistryRules, roomData, effect)

	case structs.EffectContestPenalty:
		effects.ContestPenaltyEffect(ctx, rdb, RegistryRules, roomData, effect)

	case structs.EffectAssassin:
		effects.AssassinEffect(ctx, rdb, RegistryRules, roomData, effect)

	case structs.EffectKamikaze:
		effects.KamikazeEffect(ctx, rdb, RegistryRules, roomData, effect)

	case structs.EffectTrillionaire:
		effects.TrillionaireEffect(ctx, rdb, RegistryRules, roomData, effect)
	}
}

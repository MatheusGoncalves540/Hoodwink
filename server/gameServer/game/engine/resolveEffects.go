package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func resolveNextEffect(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry, roomData *rooms.Room) {
	effect := roomData.PendingEffects[len(roomData.PendingEffects)-1]
	// remove o efeito que está sendo resolvido
	roomData.CancelLastPendingEffect(ctx, rdb)

	switch effect.Cause {
	case structs.EffectContest:
		effects.ContestEffect(ctx, rdb, RegistryRules, roomData, effect)

	case structs.EffectContestPenalty:
		effects.ContestPenaltyEffect(ctx, rdb, RegistryRules, roomData, effect)

	case structs.EffectAssassin:
		effects.AssassinEffect(ctx, rdb, RegistryRules, roomData, effect)

	case structs.EffectKamikaze:
		effects.KamikazeEffect(ctx, rdb, RegistryRules, roomData, effect)
	}
}

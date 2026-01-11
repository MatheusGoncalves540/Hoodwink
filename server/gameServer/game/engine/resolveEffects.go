package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/redis/go-redis/v9"
)

func resolveNextEffect(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry, roomData *rooms.Room) {
	effect := roomData.PendingEffects[0]

	switch effect.Cause {

	case roomStructs.EffectAssassin:
		effects.AssassinEffect(ctx, rdb, RegistryRules, roomData, effect)
	}
	// remove o efeito que está sendo resolvido
	roomData.PendingEffects = roomData.PendingEffects[1:]
}

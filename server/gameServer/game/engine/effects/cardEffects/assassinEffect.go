package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func AssassinEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	// resolve o efeito de matar carta
	assassinPayload, ok := effect.Payload.(structs.AssassinPayload)
	if !ok {
		utils.LogError("payload inválido para AssassinPayload")
		return
	}

	killPayload := structs.NewKillCardPayload(string(effect.Cause), assassinPayload.TargetPlayer, assassinPayload.TargetCardIndex)

	killEffect := structs.Effect{
		Cause:        structs.EffectAssassin,
		SourcePlayer: effect.SourcePlayer,
		Payload:      killPayload,
	}

	err := effects.KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

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

func KamikazeEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	// resolve o efeito de matar carta
	kamikazePayload, ok := effect.Payload.(structs.KamikazePayload)
	if !ok {
		utils.LogError("payload inválido para KamikazePayload")
		return
	}

	killPayload := structs.NewKillCardPayload(string(effect.Cause), kamikazePayload.TargetPlayer, kamikazePayload.TargetCardIndex)

	killEffect := structs.Effect{
		Cause:        structs.EffectKamikaze,
		SourcePlayer: effect.SourcePlayer,
		Payload:      killPayload,
	}

	err := effects.KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

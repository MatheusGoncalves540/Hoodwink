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

func GravediggerEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	gravediggerPayload, ok := effect.Payload.(structs.GravediggerPayload)
	if !ok {
		utils.LogError("payload inválido para GravediggerPayload")
		return
	}

	reviveCardPayload := structs.NewReviveCardPayload(string(effect.Cause), gravediggerPayload.TargetPlayer, gravediggerPayload.TargetCardIndex)

	reviveCardEffect := structs.Effect{
		Cause:        structs.EffectGravedigger,
		SourcePlayer: effect.SourcePlayer,
		Payload:      reviveCardPayload,
	}

	// chama o ReviveCardAnnouncer para ressuscitar a carta morta do jogador
	if err := effects.ReviveCardAnnouncer(ctx, rdb, registryRules, roomData, reviveCardEffect); err != nil {
		utils.LogError(err)
	}
}

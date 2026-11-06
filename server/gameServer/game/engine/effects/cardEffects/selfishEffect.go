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

func SelfishEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	selfishPayload, ok := effect.Payload.(structs.SelfishPayload)
	if !ok {
		utils.LogError("payload inválido para SelfishPayload")
		return
	}

	effect.Payload = structs.NewRichnessDeniedPayload(string(effect.Cause), *selfishPayload.WhatWasDenied, selfishPayload.TargetPlayer)

	err := effects.RichnessDeniedAnnouncer(ctx, rdb, registryRules, roomData, effect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

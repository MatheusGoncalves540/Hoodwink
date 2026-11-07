package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func SelfishEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	selfishPayload, ok := effect.Payload.(structs.SelfishPayload)
	if !ok {
		pure.LogError("payload inválido para SelfishPayload")
		return
	}

	effect.Payload = structs.NewRichnessDeniedPayload(string(effect.Cause), *selfishPayload.WhatWasDenied, selfishPayload.TargetPlayer)

	err := effects.RichnessDeniedAnnouncer(ctx, sink, registryRules, roomData, effect)
	if err != nil {
		pure.LogError(err)
		return
	}
}

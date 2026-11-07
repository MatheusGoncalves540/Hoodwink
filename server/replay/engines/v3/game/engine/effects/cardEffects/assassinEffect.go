package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
)

func AssassinEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	// resolve o efeito de matar carta
	assassinPayload, ok := effect.Payload.(structs.AssassinPayload)
	if !ok {
		pure.LogError("payload inválido para AssassinPayload")
		return
	}

	killPayload := structs.NewKillCardPayload(string(effect.Cause), assassinPayload.TargetPlayer, assassinPayload.TargetCardIndex)

	killEffect := structs.Effect{
		Cause:        structs.EffectAssassin,
		SourcePlayer: effect.SourcePlayer,
		Payload:      killPayload,
	}

	err := effects.KillAnnouncerCard(ctx, sink, registryRules, roomData, killEffect)
	if err != nil {
		pure.LogError(err)
		return
	}
}

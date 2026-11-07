package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

func TricksterEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	tricksterPayload, ok := effect.Payload.(structs.TricksterPayload)
	if !ok {
		pure.LogError("payload inválido para TricksterPayload")
		return
	}

	effect.Payload = structs.NewStealCoinsPayload(string(effect.Cause), tricksterPayload.StealedCoins, &tricksterPayload.TargetPlayer)

	err := effects.StealCoinsAnnouncer(ctx, sink, registryRules, roomData, effect)
	if err != nil {
		pure.LogError(err)
		return
	}
}

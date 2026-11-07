package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

func TrillionaireEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	trillionairePayload, ok := effect.Payload.(structs.TrillionairePayload)
	if !ok {
		pure.LogError("payload inválido para TrillionairePayload")
		return
	}

	effect.Payload = structs.NewEarnCoinsPayload(string(effect.Cause), trillionairePayload.EarnedCoins, nil)

	err := effects.EarnCoinsAnnouncer(ctx, sink, roomData, effect)
	if err != nil {
		pure.LogError(err)
		return
	}
}

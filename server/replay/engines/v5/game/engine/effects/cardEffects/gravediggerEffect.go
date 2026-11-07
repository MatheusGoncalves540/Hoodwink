package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

func GravediggerEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	gravediggerPayload, ok := effect.Payload.(structs.GravediggerPayload)
	if !ok {
		pure.LogError("payload inválido para GravediggerPayload")
		return
	}

	reviveCardPayload := structs.NewReviveCardPayload(string(effect.Cause), gravediggerPayload.TargetPlayer, gravediggerPayload.TargetCardIndex)

	reviveCardEffect := structs.Effect{
		Cause:        structs.EffectGravedigger,
		SourcePlayer: effect.SourcePlayer,
		Payload:      reviveCardPayload,
	}

	// chama o ReviveCardAnnouncer para ressuscitar a carta morta do jogador
	if err := effects.ReviveCardAnnouncer(ctx, sink, registryRules, roomData, reviveCardEffect); err != nil {
		pure.LogError(err)
	}
}

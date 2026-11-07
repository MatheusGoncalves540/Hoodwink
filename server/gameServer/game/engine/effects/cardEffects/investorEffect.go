package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func InvestorEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	investorPayload, ok := effect.Payload.(structs.InvestorPayload)
	if !ok {
		pure.LogError("payload inválido para InvestorPayload")
		return
	}

	gracePeriod := 6
	if investorPayload.GracePeriod != nil {
		gracePeriod = *investorPayload.GracePeriod
	}

	effect.Payload = structs.NewInvestmentPayload(string(effect.Cause), &gracePeriod)

	err := effects.InvestmentAnnouncer(ctx, sink, registryRules, roomData, effect)
	if err != nil {
		pure.LogError(err)
		return
	}
}

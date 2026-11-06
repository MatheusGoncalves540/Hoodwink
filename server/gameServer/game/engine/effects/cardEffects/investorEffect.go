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

func InvestorEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	investorPayload, ok := effect.Payload.(structs.InvestorPayload)
	if !ok {
		utils.LogError("payload inválido para InvestorPayload")
		return
	}

	gracePeriod := 6
	if investorPayload.GracePeriod != nil {
		gracePeriod = *investorPayload.GracePeriod
	}

	effect.Payload = structs.NewInvestmentPayload(string(effect.Cause), &gracePeriod)

	err := effects.InvestmentAnnouncer(ctx, rdb, registryRules, roomData, effect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

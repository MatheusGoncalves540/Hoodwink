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

	effect.Payload = structs.NewInvestmentPayload(string(effect.Cause), investorPayload.GracePeriod)

	err := effects.InvestmentAnnouncer(ctx, rdb, registryRules, roomData, effect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

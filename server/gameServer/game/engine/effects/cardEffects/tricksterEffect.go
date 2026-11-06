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

func TricksterEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	tricksterPayload, ok := effect.Payload.(structs.TricksterPayload)
	if !ok {
		utils.LogError("payload inválido para TricksterPayload")
		return
	}

	effect.Payload = structs.NewStealCoinsPayload(string(effect.Cause), tricksterPayload.StealedCoins, &tricksterPayload.TargetPlayer)

	err := effects.StealCoinsAnnouncer(ctx, rdb, registryRules, roomData, effect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

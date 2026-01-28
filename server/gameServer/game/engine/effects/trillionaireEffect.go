package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func TrillionaireEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	trillionairePayload, ok := effect.Payload.(structs.TrillionairePayload)
	if !ok {
		utils.LogError("payload inválido para TrillionairePayload")
		return
	}

	effect.Payload = structs.NewEarnCoinsPayload(string(effect.Cause), trillionairePayload.EarnedCoins, nil)

	err := EarnCoinsAnnouncer(ctx, rdb, registryRules, roomData, effect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

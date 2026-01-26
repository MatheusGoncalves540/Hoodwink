package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func AssassinEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	// resolve o efeito de matar carta
	err := KillCard(ctx, rdb, registryRules, roomData, effect)
	if err != nil {
		utils.LogError(err)
		return
	}
}

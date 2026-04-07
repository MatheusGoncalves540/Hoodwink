package commands

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func CommandStartGame(ctx context.Context, rdb *redis.Client, roomData *rooms.Room) {
	effects.StartGameEffect(ctx, rdb, roomData)
}

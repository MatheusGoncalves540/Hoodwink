package redisHandlers

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func SaveRoom(ctx context.Context, rdb *redis.Client, room *roomStructs.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, "room:"+room.ID, data, 0).Err()
}

package redisHandlers

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func LoadRoom(ctx context.Context, rdb *redis.Client, roomID string) (*roomStructs.Room, error) {
	val, err := rdb.Get(ctx, "room:"+roomID).Result()
	if err != nil {
		return nil, err
	}
	var room roomStructs.Room
	err = json.Unmarshal([]byte(val), &room)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

package redisHandlers

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/redis/go-redis/v9"
)

func PushEvent(ctx context.Context, rdb *redis.Client, roomID string, evt eventQueue.Event) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return rdb.LPush(ctx, "room:"+roomID+":eventQueue", data).Err()
}

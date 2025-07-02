package redisHandlers

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/redis/go-redis/v9"
)

func PopEvent(ctx context.Context, rdb *redis.Client, roomID string) (*eventQueue.Event, error) {
	res, err := rdb.BRPop(ctx, 0, "room:"+roomID+":eventQueue").Result()
	if err != nil {
		return nil, err
	}
	if len(res) < 2 {
		return nil, nil
	}
	var evt eventQueue.Event
	err = json.Unmarshal([]byte(res[1]), &evt)
	if err != nil {
		return nil, err
	}
	return &evt, nil
}

package room

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type Event struct {
	Type       string      `json:"type"`
	PlayerUUID string      `json:"player_uuid"`
	Payload    interface{} `json:"payload"`
}

func PushEvent(ctx context.Context, rdb *redis.Client, roomID string, evt Event) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return rdb.LPush(ctx, "room:"+roomID+":eventQueue", data).Err()
}

func PopEvent(ctx context.Context, rdb *redis.Client, roomID string) (*Event, error) {
	res, err := rdb.BRPop(ctx, 0, "room:"+roomID+":eventQueue").Result()
	if err != nil {
		return nil, err
	}
	if len(res) < 2 {
		return nil, nil
	}
	var evt Event
	err = json.Unmarshal([]byte(res[1]), &evt)
	if err != nil {
		return nil, err
	}
	return &evt, nil
}

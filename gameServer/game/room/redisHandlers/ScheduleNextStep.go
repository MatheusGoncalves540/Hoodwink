package redisHandlers

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/redis/go-redis/v9"
)

func ScheduleNextStep(ctx context.Context, rdb *redis.Client, roomID string, evt eventQueue.Event) {
	duration := time.Duration(evt.TimeoutMillis) * time.Millisecond
	time.AfterFunc(duration, func() {
		_ = PushEvent(ctx, rdb, roomID, evt)
	})
}

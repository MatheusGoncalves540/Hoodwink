package redisHandlers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// TryLockRoom tenta adquirir um lock para a sala especificada.
func TryLockRoom(ctx context.Context, rdb *redis.Client, roomID string) bool {
	return rdb.SetNX(
		ctx,
		"lock:room:"+roomID,
		"1",
		2*time.Second,
	).Val()
}

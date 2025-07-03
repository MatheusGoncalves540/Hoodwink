package redisHandlers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Tenta adquirir um lock distribuído para uma sala
func AcquireRoomLock(ctx context.Context, rdb *redis.Client, roomID, instanceID string, ttl time.Duration) (bool, error) {
	return rdb.SetNX(ctx, "lock:room:"+roomID, instanceID, ttl).Result()
}

// Remove o lock se ainda for seu
func ReleaseRoomLock(ctx context.Context, rdb *redis.Client, roomID, instanceID string) error {
	val, err := rdb.Get(ctx, "lock:room:"+roomID).Result()
	if err == nil && val == instanceID {
		return rdb.Del(ctx, "lock:room:"+roomID).Err()
	}
	return nil
}

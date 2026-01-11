package player

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/redis/go-redis/v9"
)

// AcquirePlayerLock tenta adquirir um lock distribuído para um player.
// Isso garante que ele não seja registrado em duas salas ao mesmo tempo por instâncias diferentes.
func AcquirePlayerLock(ctx context.Context, rdb *redis.Client, playerId string, ttl time.Duration) (bool, error) {
	return rdb.SetNX(ctx, "lock:player:"+playerId, config.InstanceID, ttl).Result()
}

// ReleasePlayerLock remove o lock do player se ainda pertencer à instância atual.
func ReleasePlayerLock(ctx context.Context, rdb *redis.Client, playerId string) error {
	val, err := rdb.Get(ctx, "lock:player:"+playerId).Result()
	if err == nil && val == config.InstanceID {
		return rdb.Del(ctx, "lock:player:"+playerId).Err()
	}
	// não remove se não for o dono do lock ou se não existir
	return nil
}

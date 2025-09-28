package redisHandlers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// SetRoomTTL define um TTL para uma sala específica
func SetRoomTTL(ctx context.Context, rdb *redis.Client, roomId string, ttl time.Duration) error {
	return rdb.Expire(ctx, "room:"+roomId, ttl).Err()
}

// RemoveRoomTTL remove o TTL de uma sala (torna ela persistente)
func RemoveRoomTTL(ctx context.Context, rdb *redis.Client, roomId string) error {
	return rdb.Persist(ctx, "room:"+roomId).Err()
}

// CheckIfRoomIsEmpty verifica se uma sala está vazia (sem jogadores conectados)
func CheckIfRoomIsEmpty(ctx context.Context, rdb *redis.Client, roomId string) (bool, error) {
	room, err := LoadRoom(ctx, rdb, roomId)
	if err != nil {
		return false, err
	}

	// Verifica se há jogadores conectados
	for _, player := range room.Players {
		if player.Connected {
			return false, nil
		}
	}

	return true, nil
}

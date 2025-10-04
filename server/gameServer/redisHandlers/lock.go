package redisHandlers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// AcquireRoomLock tenta adquirir um lock distribuído para uma sala usando SetNX.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	RoomId: identificador da sala
//	instanceID: identificador único da instância/processo
//	ttl: tempo de expiração do lock
//
// Retorno:
//
//	bool: true se lock foi adquirido, false se já existe
//	error: erro de comunicação com Redis
func AcquireRoomLock(ctx context.Context, rdb *redis.Client, RoomId, instanceID string, ttl time.Duration) (bool, error) {
	// Tenta criar a chave de lock com TTL
	return rdb.SetNX(ctx, "lock:room:"+RoomId, instanceID, ttl).Result()
}

// ReleaseRoomLock remove o lock da sala se ainda pertencer à instância atual.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	RoomId: identificador da sala
//	instanceID: identificador único da instância/processo
//
// Retorno:
//
//	error: erro de comunicação com Redis
func ReleaseRoomLock(ctx context.Context, rdb *redis.Client, RoomId, instanceID string) error {
	// Busca o valor atual do lock
	val, err := rdb.Get(ctx, "lock:room:"+RoomId).Result()
	if err == nil && val == instanceID {
		// Só remove se o lock for da instância
		return rdb.Del(ctx, "lock:room:"+RoomId).Err()
	}
	// Não remove se não for o dono ou se não existir
	return nil
}

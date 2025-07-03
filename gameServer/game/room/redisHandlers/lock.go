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
//	roomID: identificador da sala
//	instanceID: identificador único da instância/processo
//	ttl: tempo de expiração do lock
//
// Retorno:
//
//	bool: true se lock foi adquirido, false se já existe
//	error: erro de comunicação com Redis
func AcquireRoomLock(ctx context.Context, rdb *redis.Client, roomID, instanceID string, ttl time.Duration) (bool, error) {
	// Tenta criar a chave de lock com TTL
	return rdb.SetNX(ctx, "lock:room:"+roomID, instanceID, ttl).Result()
}

// ReleaseRoomLock remove o lock da sala se ainda pertencer à instância atual.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	roomID: identificador da sala
//	instanceID: identificador único da instância/processo
//
// Retorno:
//
//	error: erro de comunicação com Redis
func ReleaseRoomLock(ctx context.Context, rdb *redis.Client, roomID, instanceID string) error {
	// Busca o valor atual do lock
	val, err := rdb.Get(ctx, "lock:room:"+roomID).Result()
	if err == nil && val == instanceID {
		// Só remove se o lock for da instância
		return rdb.Del(ctx, "lock:room:"+roomID).Err()
	}
	// Não remove se não for o dono ou se não existir
	return nil
}

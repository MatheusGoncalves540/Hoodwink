package room

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// AcquireRoomLock tenta adquirir ou renovar um lock distribuído para uma sala.
//
// A função possui comportamento reentrante por instância:
//   - Se o lock não existir, ele é criado com o TTL informado.
//   - Se o lock já existir e pertencer à mesma instância, o TTL é renovado e
//     a função retorna sucesso.
//   - Se o lock existir e pertencer a outra instância, o lock não é adquirido.
//
// Parâmetros:
//
//	ctx: contexto para controle de timeout e cancelamento
//	rdb: cliente Redis
//	roomID: identificador da sala
//	instanceID: identificador único da instância/processo
//	ttl: tempo de expiração do lock
//
// Retorno:
//
//	bool: true se o lock foi adquirido ou renovado pela instância atual;
//	      false se o lock pertence a outra instância
//	error: erro de comunicação com o Redis
func AcquireRoomLock(
	ctx context.Context,
	rdb *redis.Client,
	roomID string,
	instanceID string,
	ttl time.Duration,
) (bool, error) {
	key := "lock:room:" + roomID

	// 1. Tenta adquirir o lock normalmente
	ok, err := rdb.SetNX(ctx, key, instanceID, ttl).Result()
	if err != nil {
		return false, err
	}

	if ok {
		// Lock adquirido com sucesso
		return true, nil
	}

	// 2. Lock já existe — verificar se é da mesma instância
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// A chave sumiu entre o SetNX e o Get (condição rara)
			return false, nil
		}
		return false, err
	}

	if val != instanceID {
		// Lock pertence a outra instância
		return false, nil
	}

	// 3. Lock é da mesma instância → renovar TTL
	_, err = rdb.Expire(ctx, key, ttl).Result()
	if err != nil {
		return false, err
	}

	return true, nil
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

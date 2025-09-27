package redisHandlers

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/redis/go-redis/v9"
)

// RegisterPlayerInRoom registra no Redis que o player está conectado em uma sala.
// Não adiciona o player na struct da sala, apenas mantém o vínculo player → room.
// O TTL pode ser usado para expirar automaticamente caso a conexão seja perdida.
func RegisterPlayerInRoom(ctx context.Context, rdb *redis.Client, playerId, roomId string) error {
	return rdb.Set(ctx, "player:"+playerId+":room", roomId, 0).Err()
}

// UnregisterPlayerFromRoom remove do Redis o vínculo de um player com uma sala.
// Normalmente chamado quando o player sai ou é desconectado.
func UnregisterPlayerFromRoom(ctx context.Context, rdb *redis.Client, playerId string) error {
	return rdb.Del(ctx, "player:"+playerId+":room").Err()
}

// GetRegisteredRoomForPlayer retorna a sala em que o player está registrado.
// Retorno: roomId, bool (true se está registrado), erro do Redis.
func GetRegisteredRoomForPlayer(ctx context.Context, rdb *redis.Client, playerId string) (string, bool, error) {
	roomId, err := rdb.Get(ctx, "player:"+playerId+":room").Result()
	if err == redis.Nil {
		return "", false, nil // player não está em nenhuma sala
	}
	if err != nil {
		return "", false, err // erro de comunicação com Redis
	}
	return roomId, true, nil
}

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

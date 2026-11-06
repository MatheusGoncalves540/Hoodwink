package playerRedis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// AcquirePlayerLock tenta adquirir um lock distribuído para um player.
// Isso garante que ele não seja registrado em duas salas ao mesmo tempo por instâncias diferentes.
func AcquirePlayerLock(ctx context.Context, rdb *redis.Client, playerId string, ttl time.Duration) (bool, error) {
	return rdb.SetNX(ctx, "lock:player:"+playerId, utils.GetInstanceID(), ttl).Result()
}

// ReleasePlayerLock remove o lock do player se ainda pertencer à instância atual.
func ReleasePlayerLock(ctx context.Context, rdb *redis.Client, playerId string) error {
	val, err := rdb.Get(ctx, "lock:player:"+playerId).Result()
	if err == nil && val == utils.GetInstanceID() {
		return rdb.Del(ctx, "lock:player:"+playerId).Err()
	}
	// não remove se não for o dono do lock ou se não existir
	return nil
}

// GetRegisteredRoomForPlayer retorna a sala em que o player está registrado.
// Agora lida com o formato "roomId:instanceId" e verifica se a instância está viva.
// Retorno: roomId, bool (true se está registrado e instância viva), erro do Redis.
func GetRegisteredRoomForPlayer(ctx context.Context, rdb *redis.Client, playerId string) (string, bool, error) {
	value, err := rdb.Get(ctx, "player:"+playerId+":room").Result()
	if err == redis.Nil {
		return "", false, nil // player não está em nenhuma sala
	}
	if err != nil {
		return "", false, err // erro de comunicação com Redis
	}

	// Parse do formato "roomId:instanceId"
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		// Formato antigo ou inválido, remove registro
		UnregisterPlayerFromRoom(ctx, rdb, playerId)
		return "", false, nil
	}

	roomId := parts[0]
	instanceId := parts[1]

	// Verifica se a instância ainda está viva
	instanceKey := fmt.Sprintf("instance:%s:alive", instanceId)
	_, err = rdb.Get(ctx, instanceKey).Result()
	if err == redis.Nil {
		// Instância morreu, remove registro do player
		UnregisterPlayerFromRoom(ctx, rdb, playerId)
		return "", false, nil
	}
	if err != nil {
		return "", false, err // erro de comunicação com Redis
	}

	return roomId, true, nil
}

// RegisterPlayerInRoom registra no Redis que o player está conectado em uma sala.
// O formato inclui a instanceID: "roomId:instanceId"
// Com TTL sincronizado com o heartbeat da instância.
func RegisterPlayerInRoom(ctx context.Context, rdb *redis.Client, roomId string, playerId string) error {
	key := fmt.Sprintf("player:%s:room", playerId)
	value := fmt.Sprintf("%s:%s", roomId, utils.GetInstanceID())
	// TTL igual ao heartbeat (10s)
	return rdb.Set(ctx, key, value, 10*time.Second).Err()
}

// UnregisterPlayerFromRoom remove do Redis o vínculo de um player com uma sala.
// Normalmente chamado quando o player sai ou é desconectado.
func UnregisterPlayerFromRoom(ctx context.Context, rdb *redis.Client, playerId string) error {
	return rdb.Del(ctx, "player:"+playerId+":room").Err()
}

// GetPlayerRegistrationInfo retorna informações completas do registro do player
// Retorno: roomId, instanceId, bool (registrado), erro
func GetPlayerRegistrationInfo(ctx context.Context, rdb *redis.Client, playerId string) (string, string, bool, error) {
	value, err := rdb.Get(ctx, "player:"+playerId+":room").Result()
	if err == redis.Nil {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, err
	}

	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return "", "", false, nil
	}

	return parts[0], parts[1], true, nil
}

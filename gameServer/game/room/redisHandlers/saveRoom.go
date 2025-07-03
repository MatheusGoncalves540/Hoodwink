package redisHandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// SaveRoom salva o estado da sala no Redis de forma segura usando lock distribuído.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	room: ponteiro para struct Room a ser salva
//
// Retorno:
//
//	error: erro de lock, serialização ou comunicação com Redis
func SaveRoom(ctx context.Context, rdb *redis.Client, room *roomStructs.Room) error {
	// Gera um ID único para a instância atual
	instanceID := utils.GetInstanceID()
	// Tenta adquirir o lock da sala
	ok, err := AcquireRoomLock(ctx, rdb, room.ID, instanceID, 5*time.Second)
	if err != nil {
		// Retorna erro se falhar ao tentar lock
		return err
	}
	if !ok {
		// Retorna erro se outra instância já está modificando
		return fmt.Errorf("a sala %s está sendo modificada por outra instância, tente novamente", room.ID)
	}
	// Libera o lock ao final da função
	defer ReleaseRoomLock(ctx, rdb, room.ID, instanceID)

	// Serializa a sala para JSON
	data, err := json.Marshal(room)
	if err != nil {
		// Retorna erro se não conseguir serializar
		return err
	}
	// Salva o JSON no Redis (sem expiração)
	return rdb.Set(ctx, "room:"+room.ID, data, 0).Err()
}

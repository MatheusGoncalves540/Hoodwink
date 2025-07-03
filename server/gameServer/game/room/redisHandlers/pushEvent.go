package redisHandlers

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/redis/go-redis/v9"
)

// PushEvent adiciona um novo evento à fila de eventos da sala no Redis.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	roomID: identificador da sala
//	evt: evento a ser adicionado (struct Event)
//
// Retorno:
//
//	error: erro de serialização ou comunicação com Redis
func PushEvent(ctx context.Context, rdb *redis.Client, roomID string, evt eventQueue.Event) error {
	// Serializa o evento para JSON
	data, err := json.Marshal(evt)
	if err != nil {
		// Retorna erro se não conseguir serializar
		return err
	}
	// Adiciona o evento no início da fila de eventos (LPUSH)
	return rdb.LPush(ctx, "room:"+roomID+":eventQueue", data).Err()
}

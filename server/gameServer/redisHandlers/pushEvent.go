package redisHandlers

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// PushPlayerPlay adiciona um novo evento à fila de eventos da sala no Redis.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	RoomId: identificador da sala
//	playerPlay: evento a ser adicionado (struct Event)
//
// Retorno:
//
//	error: erro de serialização ou comunicação com Redis
func PushPlayerPlay(ctx context.Context, rdb *redis.Client, RoomId string, playerPlay roomStructs.PlayerPlay) error {
	// Serializa o evento para JSON
	data, err := json.Marshal(playerPlay)
	if err != nil {
		// Retorna erro se não conseguir serializar
		return err
	}
	// Adiciona o evento no início da fila de eventos (LPUSH)
	return rdb.LPush(ctx, "room:"+RoomId+":playQueue", data).Err()
}

package redisHandlers

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// LoadRoom busca o estado de uma sala no Redis e retorna como struct Room.
// Parâmetros:
//
//	ctx: contexto para controle de timeout/cancelamento
//	rdb: cliente Redis
//	roomID: identificador da sala
//
// Retorno:
//
//	*roomStructs.Room: ponteiro para struct da sala carregada
//	error: erro caso não encontre ou não consiga desserializar
func LoadRoom(ctx context.Context, rdb *redis.Client, roomID string) (*roomStructs.Room, error) {
	// Busca o valor da chave da sala no Redis (formato: room:<roomID>)
	val, err := rdb.Get(ctx, "room:"+roomID).Result()
	if err != nil {
		// Retorna erro se a chave não existir ou houver falha de conexão
		return nil, err
	}
	var room roomStructs.Room
	// Converte o JSON armazenado em struct Room
	err = json.Unmarshal([]byte(val), &room)
	if err != nil {
		// Retorna erro se o JSON estiver inválido
		return nil, err
	}
	// Retorna a struct da sala carregada
	return &room, nil
}

package redisHandlers

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// LoadRoomField busca apenas um campo específico da sala no Redis.
// Parâmetros:
// ctx: contexto para controle de timeout/cancelamento
// rdb: cliente Redis
// RoomId: identificador da sala
// field: nome do campo desejado
// Retorno:
// interface{}: valor do campo
// error: erro caso não encontre ou não consiga desserializar
func LoadRoomField(ctx context.Context, rdb *redis.Client, RoomId string, field string) (interface{}, error) {
	val, err := rdb.Get(ctx, "room:"+RoomId).Result()
	if err != nil {
		return nil, err
	}
	var room roomStructs.Room
	err = json.Unmarshal([]byte(val), &room)
	if err != nil {
		return nil, err
	}
	v := reflect.ValueOf(room)
	f := v.FieldByName(field)
	if !f.IsValid() {
		return nil, &FieldNotFoundError{Field: field}
	}
	return f.Interface(), nil
}

// FieldNotFoundError representa erro de campo não encontrado
type FieldNotFoundError struct {
	Field string
}

func (e *FieldNotFoundError) Error() string {
	return "Campo '" + e.Field + "' não encontrado na struct Room"
}

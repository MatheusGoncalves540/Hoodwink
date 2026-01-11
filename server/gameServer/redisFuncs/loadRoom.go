package redisFuncs

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/redis/go-redis/v9"
)

// LoadRoom busca o estado de uma sala no Redis e retorna como struct Room
func LoadRoom(ctx context.Context, rdb *redis.Client, RoomId string) (*rooms.Room, error) {
	// Busca o valor da chave da sala no Redis (formato: room:<RoomId>)
	val, err := rdb.Get(ctx, "room:"+RoomId).Result()
	if err != nil {
		// Retorna erro se a chave não existir ou houver falha de conexão
		return nil, err
	}
	var room rooms.Room
	// Converte o JSON armazenado em struct Room
	err = json.Unmarshal([]byte(val), &room)
	if err != nil {
		// Retorna erro se o JSON estiver inválido
		return nil, err
	}
	// Retorna a struct da sala carregada
	return &room, nil
}

// LoadRoomField busca apenas um campo específico da sala no Redis
func LoadRoomField(ctx context.Context, rdb *redis.Client, RoomId string, field string) (any, error) {
	val, err := rdb.Get(ctx, "room:"+RoomId).Result()
	if err != nil {
		return nil, err
	}
	var room rooms.Room
	err = json.Unmarshal([]byte(val), &room)
	if err != nil {
		return nil, err
	}
	v := reflect.ValueOf(room)
	f := v.FieldByName(field)
	if !f.IsValid() {
		return nil, fmt.Errorf("%s", "Campo '"+field+"' não encontrado na struct Room")
	}
	return f.Interface(), nil
}

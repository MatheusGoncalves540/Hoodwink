package room

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// LoadRoom busca o estado de uma sala no Redis e retorna como struct Room.
// Parâmetros:
//
//	ctx: contexto para controle de timeout/cancelamento
//	rdb: cliente Redis
//	RoomId: identificador da sala
//
// Retorno:
//
//	*roomStructs.Room: ponteiro para struct da sala carregada
//	error: erro caso não encontre ou não consiga desserializar
func LoadRoom(ctx context.Context, rdb *redis.Client, RoomId string) (*roomStructs.Room, error) {
	// Busca o valor da chave da sala no Redis (formato: room:<RoomId>)
	val, err := rdb.Get(ctx, "room:"+RoomId).Result()
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
	roomAfkTtlLimit := utils.MustEnvInt("ROOM_AFK_TTL_LIMIT", 180) // 3 minutos padrão
	return SaveRoomWithTTL(ctx, rdb, room, time.Duration(roomAfkTtlLimit)*time.Second)
}

// SaveRoomWithTTL salva o estado da sala no Redis com TTL específico
func SaveRoomWithTTL(ctx context.Context, rdb *redis.Client, room *roomStructs.Room, ttl time.Duration) error {
	// Pega o ID da instância atual
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
	// Salva o JSON no Redis com TTL específico
	err = rdb.Set(ctx, "room:"+room.ID, data, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

// LoadRoomField busca apenas um campo específico da sala no Redis.
// Parâmetros:
// ctx: contexto para controle de timeout/cancelamento
// rdb: cliente Redis
// RoomId: identificador da sala
// field: nome do campo desejado
// Retorno:
// any: valor do campo
// error: erro caso não encontre ou não consiga desserializar
func LoadRoomField(ctx context.Context, rdb *redis.Client, RoomId string, field string) (any, error) {
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

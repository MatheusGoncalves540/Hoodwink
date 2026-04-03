package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// SaveRoomWithTTL salva o estado da sala no Redis com TTL específico
func (r *Room) SaveRoomWithTTL(ctx context.Context, rdb *redis.Client, ttl time.Duration) error {
	// Pega o ID da instância atual
	instanceID := utils.GetInstanceID()
	// Tenta adquirir o lock da sala
	ok, err := r.AcquireRoomLock(ctx, rdb, instanceID, 5*time.Second)
	if err != nil {
		// Retorna erro se falhar ao tentar lock
		utils.LogError(err)
		return err
	}
	if !ok {
		// Retorna erro se outra instância já está modificando
		return fmt.Errorf("não foi possível adquirir lock para a sala %s", r.ID)
	}
	// Libera o lock ao final da função
	defer r.ReleaseRoomLock(ctx, rdb, instanceID)

	// Serializa a sala para JSON
	data, err := json.Marshal(r)
	if err != nil {
		// Retorna erro se não conseguir serializar
		return err
	}
	// Salva o JSON no Redis com TTL específico
	err = rdb.Set(ctx, "room:"+r.ID, data, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

// SaveRoom salva o estado da sala no Redis de forma segura usando lock distribuído.
func (r *Room) SaveRoom(ctx context.Context, rdb *redis.Client) error {
	roomAfkTtlLimit := utils.MustEnvInt("ROOM_AFK_TTL_LIMIT", 180) // 3 minutos padrão
	err := r.SaveRoomWithTTL(ctx, rdb, time.Duration(roomAfkTtlLimit)*time.Second)
	if err != nil {
		return err
	}
	return nil
}

// AcquireRoomLock tenta adquirir ou renovar um lock distribuído no redis para uma sala.
func (r *Room) AcquireRoomLock(ctx context.Context, rdb *redis.Client, instanceID string, ttl time.Duration) (bool, error) {
	key := "lock:room:" + r.ID

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

// ReleaseRoomLock remove o lock redis da sala se ainda pertencer à instância atual.
func (r *Room) ReleaseRoomLock(ctx context.Context, rdb *redis.Client, instanceID string) error {
	// Busca o valor atual do lock
	val, err := rdb.Get(ctx, "lock:room:"+r.ID).Result()
	if err == nil && val == instanceID {
		// Só remove se o lock for da instância
		return rdb.Del(ctx, "lock:room:"+r.ID).Err()
	}
	// Não remove se não for o dono ou se não existir
	return nil
}

// SetTTL define um TTL para uma sala específica
func (r *Room) SetTTL(ctx context.Context, rdb *redis.Client, ttl time.Duration) error {
	return rdb.Expire(ctx, "room:"+r.ID, ttl).Err()
}

// RemoveTTL remove o TTL de uma sala (torna ela persistente)
func (r *Room) RemoveTTL(ctx context.Context, rdb *redis.Client) error {
	return rdb.Persist(ctx, "room:"+r.ID).Err()
}

// RegistryTimeout registra o timeout no redis
func (r *Room) RegistryTimeout(rdb *redis.Client, ctx context.Context, expiresAt time.Time) {
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: r.ID,
	})
}

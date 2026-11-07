package redisFuncs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// SaveRoomWithTTL salva o estado da sala no Redis com TTL específico.
func SaveRoomWithTTL(ctx context.Context, rdb *redis.Client, room *rooms.Room, ttl time.Duration) error {
	return persistRoom(ctx, rdb, room, ttl, nil)
}

// SaveRoom salva o estado da sala no Redis de forma segura usando lock distribuído.
func SaveRoom(ctx context.Context, rdb *redis.Client, room *rooms.Room) error {
	roomAfkTtlLimit := utils.MustEnvInt("ROOM_AFK_TTL_LIMIT", 180) // 3 minutos padrão
	return SaveRoomWithTTL(ctx, rdb, room, time.Duration(roomAfkTtlLimit)*time.Second)
}

// SaveRoomAndReplay grava o estado da sala e a entrada de replay no MESMO MULTI
// (SET room + RPUSH replay + EXPIRE), preservando a atomicidade. Usado pelo
// WithRoom quando o fn aplicou uma mudança replayável.
func SaveRoomAndReplay(ctx context.Context, rdb *redis.Client, room *rooms.Room, replayEntry []byte) error {
	roomAfkTtlLimit := utils.MustEnvInt("ROOM_AFK_TTL_LIMIT", 180) // 3 minutos padrão
	return persistRoom(ctx, rdb, room, time.Duration(roomAfkTtlLimit)*time.Second, replayEntry)
}

// persistRoom resolve o lock (do WithRoom ou um novo) e delega ao saveRoom.
func persistRoom(ctx context.Context, rdb *redis.Client, room *rooms.Room, ttl time.Duration, replayEntry []byte) error {
	// Sem Redis (motor rodando offline, nos testes de cenário) não há o que persistir.
	if rdb == nil {
		return nil
	}

	// Dentro do WithRoom o lock já é da instância: não pega um segundo lock e, se a
	// renovação perdeu o lock, falha em vez de gravar por cima de outro pod.
	if lock := room.Lock(); lock != nil {
		if err := lock.Err(); err != nil {
			return fmt.Errorf("lock da sala %s perdido; estado não gravado: %w", room.ID, err)
		}
		return saveRoom(lock.Context(), rdb, room, ttl, replayEntry)
	}

	// Fora do WithRoom (ex.: criação de sala): serializa com um lock próprio.
	instanceID := utils.GetInstanceID()
	lock, err := AcquireRoomLock(ctx, rdb, room.ID, instanceID, 5*time.Second)
	if err != nil {
		utils.LogError(err)
		return err
	}
	if lock == nil {
		return fmt.Errorf("não foi possível adquirir lock para a sala %s", room.ID)
	}
	defer lock.Release()

	return saveRoom(ctx, rdb, room, ttl, replayEntry)
}

// saveRoom serializa e grava a sala. Com replayEntry != nil, grava a entrada de
// replay no MESMO MULTI (SET room + RPUSH replay + EXPIRE), nunca depois.
func saveRoom(ctx context.Context, rdb *redis.Client, room *rooms.Room, ttl time.Duration, replayEntry []byte) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	if len(replayEntry) == 0 {
		return rdb.Set(ctx, "room:"+room.ID, data, ttl).Err()
	}

	roomKey := "room:" + room.ID
	replayKey := ReplayListKey(room.MatchID)
	_, err = rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, roomKey, data, ttl)
		pipe.RPush(ctx, replayKey, replayEntry)
		pipe.Expire(ctx, replayKey, replayListTTL)
		return nil
	})
	return err
}

// SetTTL define um TTL para uma sala específica.
func SetTTL(ctx context.Context, rdb *redis.Client, room *rooms.Room, ttl time.Duration) error {
	return rdb.Expire(ctx, "room:"+room.ID, ttl).Err()
}

// RemoveTTL remove o TTL de uma sala (torna ela persistente).
func RemoveTTL(ctx context.Context, rdb *redis.Client, room *rooms.Room) error {
	return rdb.Persist(ctx, "room:"+room.ID).Err()
}

// RedisSink implementa rooms.Sink sobre um client do Redis. É a saída de produção
// do motor; quando rdb é nil, NewSink devolve o no-op de rooms.
type RedisSink struct {
	rdb *redis.Client
}

// NewSink cria o Sink de produção. Com rdb nil devolve o no-op (replay/testes).
func NewSink(rdb *redis.Client) rooms.Sink {
	if rdb == nil {
		return rooms.NoopSink()
	}
	return &RedisSink{rdb: rdb}
}

func (s *RedisSink) Broadcast(ctx context.Context, room *rooms.Room, confidencial *rooms.Room, confidencialPlayerIds []string) error {
	return SendUpdatedRoomData(ctx, s.rdb, room, confidencial, confidencialPlayerIds)
}

func (s *RedisSink) RegistryTimeout(ctx context.Context, room *rooms.Room, expiresAt time.Time) {
	if s.rdb == nil {
		return
	}
	s.rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: room.ID,
	})
}

func (s *RedisSink) ClearTimeout(ctx context.Context, room *rooms.Room) {
	if s.rdb == nil {
		return
	}
	s.rdb.ZRem(ctx, "rooms:timeouts", room.ID)
}

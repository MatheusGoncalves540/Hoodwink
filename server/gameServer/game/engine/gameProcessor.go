package engine

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// StartGameProcessor inicia o processador de eventos do jogo.
func StartGameProcessor(rdb *redis.Client) {
	log.Println("🔄 Inicializando gameProcessorEngine...")
	ctx := context.Background()
	go func() {
		intervalMs := utils.MustEnvInt("PROCESSOR_INTERVAL_MS", 300)
		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			processExpiredRoomsEvents(ctx, rdb)
		}
	}()
}

// processExpiredRoomsEvents verifica eventos em salas com timeout expirado e avança o estado do jogo.
func processExpiredRoomsEvents(ctx context.Context, rdb *redis.Client) error {
	now := time.Now().UnixNano()

	roomIDs, err := rdb.ZRangeByScore(ctx, "rooms:timeouts", &redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatInt(now, 10),
		Offset: 0,
		Count:  int64(utils.MustEnvInt("PROCESSOR_BATCH_SIZE", 5)),
	}).Result()
	if err != nil {
		utils.LogError(err)
		return err
	}

	utils.LogDebug("Salas com timeout expirado: " + strconv.Itoa(len(roomIDs)))

	for _, roomID := range roomIDs {
		processRoomWithLock(ctx, rdb, roomID)
	}
	return nil
}

func processRoomWithLock(ctx context.Context, rdb *redis.Client, roomID string) {
	ok, err := room.AcquireRoomLock(ctx, rdb, roomID, config.InstanceID, 2*time.Second)
	if err != nil || !ok {
		return
	}
	defer room.ReleaseRoomLock(ctx, rdb, roomID, config.InstanceID)

	roomData, err := room.LoadRoom(ctx, rdb, roomID)
	if err != nil {
		utils.LogError(err)
		rdb.ZRem(ctx, "rooms:timeouts", roomID)
		return
	}

	if roomData.PendingEvent == nil {
		rdb.ZRem(ctx, "rooms:timeouts", roomID)
		return
	}

	AdvanceByTimeout(ctx, rdb, roomData)
}

func WaitingFirstAction(ctx context.Context, rdb *redis.Client, room *roomStructs.Room) {
	room.State = roomStructs.StateWaitingFirstAction

	expiresAt := time.Now().Add(5 * time.Second)
	room.PendingEvent = &roomStructs.PendingEvent{
		Type:      roomStructs.TypeDisplayingMessage,
		ExpiresAt: expiresAt,
	}

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixNano()),
		Member: room.ID,
	})
}

func NextTurn(room *roomStructs.Room, rdb *redis.Client, ctx context.Context) {
	room.Turn++

	// limpa evento anterior
	room.PendingEvent = nil

	// inicia novo turno
	WaitingFirstAction(ctx, rdb, room)
}

package game

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// StartGameProcessor inicia o processador de eventos do jogo.
func StartGameProcessor(rdb *redis.Client) {
	utils.LogDebug("✅ Processador iniciado")
	ctx := context.Background()
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			processExpiredRooms(ctx, rdb)
		}
	}()
}

// processExpiredRooms verifica eventos em salas com timeout expirado e avança o estado do jogo.
func processExpiredRooms(ctx context.Context, rdb *redis.Client) error {
	now := time.Now().UnixNano()

	roomIDs, err := rdb.ZRangeByScore(ctx, "rooms:timeouts", &redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatInt(now, 10),
		Offset: 0,
		Count:  5,
	}).Result()
	if err != nil {
		return err
	}

	utils.LogDebug("Salas com timeout expirado: " + strconv.Itoa(len(roomIDs)))

	for _, roomID := range roomIDs {
		if !redisHandlers.TryLockRoom(ctx, rdb, roomID) {
			continue
		}

		room, err := redisHandlers.LoadRoom(ctx, rdb, roomID)
		if err != nil {
			utils.LogError(err)
			rdb.ZRem(ctx, "rooms:timeouts", roomID)
			continue
		}
		if room.PendingEvent == nil {
			rdb.ZRem(ctx, "rooms:timeouts", roomID)
			continue
		}

		AdvanceByTimeout(ctx, rdb, room)

		redisHandlers.SaveRoom(ctx, rdb, room)
	}
	return nil
}

// AdvanceByTimeout avança o estado do jogo baseado no estado atual da sala.
func AdvanceByTimeout(ctx context.Context, rdb *redis.Client, room *roomStructs.Room) {
	switch room.State {

	case roomStructs.StateWaitTurn:
		NextTurn(room, rdb, ctx)
		data, _ := json.MarshalIndent(room, "", "  ")
		fmt.Println(string(data))

	case roomStructs.StateWaitAction:
		// AdvanceTurn(room)

	case roomStructs.StateWaitContest:
		// ApplyAction(room)
		// StartKamikaze(room)
	}
}

// OnPlayerEvent processa quando uma ação é enviada por um jogador.
func OnPlayerEvent(ctx context.Context, rdb *redis.Client, roomID string, event roomStructs.PendingEvent) error {
	if !redisHandlers.TryLockRoom(ctx, rdb, roomID) {
		return nil
	}

	room, err := redisHandlers.LoadRoom(ctx, rdb, roomID)
	if err != nil {
		return err
	}

	switch room.State {
	case roomStructs.StateWaitAction:
		// handleUseCard(room, event)

	case roomStructs.StateWaitContest:
		// handleContest(room, event)
	}

	err = redisHandlers.SaveRoom(ctx, rdb, room)
	if err != nil {
		return err
	}
	return nil
}

func StartTurn(ctx context.Context, rdb *redis.Client, room *roomStructs.Room) {
	room.State = roomStructs.StateWaitTurn

	expiresAt := time.Now().Add(5 * time.Second)
	room.PendingEvent = &roomStructs.PendingEvent{
		Type:      roomStructs.TypeTimeout,
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
	StartTurn(ctx, rdb, room)
}

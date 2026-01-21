package engine

import (
	"context"
	"log"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// StartGameProcessor inicia o processador de eventos do jogo.
func StartGameProcessor(rdb *redis.Client, RegistryRules *rules.Registry) {
	log.Println("🔄 Inicializando gameProcessorEngine...")
	ctx := context.Background()

	go func() {
		// Adiciona um atraso aleatório para evitar picos de carga e melhor distribuição entre instâncias
		delay := time.Duration(rand.IntN(1000)) * time.Millisecond
		time.Sleep(delay)
		log.Printf("⏱️ gameProcessorEngine iniciado após atraso de %v\n", delay)

		intervalMs := utils.MustEnvInt("PROCESSOR_INTERVAL_MS", 300)
		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			processExpiredRoomsEvents(ctx, rdb, RegistryRules)
		}
	}()
}

// processExpiredRoomsEvents verifica eventos em salas com timeout expirado e avança o estado do jogo.
func processExpiredRoomsEvents(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry) error {
	now := time.Now().UnixMilli()

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

	// utils.LogDebug("Salas com timeout expirado: " + strconv.Itoa(len(roomIDs)))

	for _, roomID := range roomIDs {
		processRoomWithLock(ctx, rdb, RegistryRules, roomID)
	}
	return nil
}

func processRoomWithLock(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry, roomID string) {
	roomData, err := redisFuncs.LoadRoom(ctx, rdb, roomID)
	if err != nil {
		rdb.ZRem(ctx, "rooms:timeouts", roomID)
		return
	}

	ok, err := roomData.AcquireRoomLock(ctx, rdb, utils.GetInstanceID(), 2*time.Second)
	if err != nil || !ok {
		return
	}
	defer roomData.ReleaseRoomLock(ctx, rdb, utils.GetInstanceID())

	// Fecha janela atual (se existir)
	if roomData.GameEvent != nil {
		roomData.GameEvent = nil
	}

	// Resolve próximo efeito, se existir
	if len(roomData.PendingEffects) > 0 {
		resolveNextEffect(ctx, rdb, RegistryRules, roomData)
		roomData.SaveRoom(ctx, rdb)
		roomData.SendUpdatedRoomData(ctx, rdb)
		return
	}

	// Nada pendente → próximo turno
	if roomData.GameEvent == nil && len(roomData.PendingEffects) == 0 {
		NextTurn(roomData, rdb, ctx)
	}
}

func WaitingFirstAction(ctx context.Context, rdb *redis.Client, roomData *rooms.Room) error {
	expiresAt := time.Now().Add(15 * time.Second).UTC()
	roomData.GameEvent = &structs.GameEvent{
		Type:      structs.EventWaitingFirstAction,
		ExpiresAt: expiresAt,
	}

	if err := roomData.SaveRoom(ctx, rdb); err != nil {
		return err
	}

	if err := roomData.SendUpdatedRoomData(ctx, rdb); err != nil {
		return err
	}

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
	return nil
}

func NextTurn(roomData *rooms.Room, rdb *redis.Client, ctx context.Context) error {
	roomData.Turn++

	// limpa evento anterior
	roomData.GameEvent = nil

	if err := roomData.SaveRoom(ctx, rdb); err != nil {
		return err
	}

	// inicia novo turno
	if err := WaitingFirstAction(ctx, rdb, roomData); err != nil {
		return err
	}
	return nil
}

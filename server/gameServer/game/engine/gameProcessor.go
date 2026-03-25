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
func StartGameProcessor(rdb *redis.Client, registryRules *rules.Registry) {
	log.Println("🔄 Inicializando gameProcessorEngine...")

	go func() {
		// Adiciona um atraso aleatório para evitar picos de carga e melhor distribuição entre instâncias
		delay := time.Duration(rand.IntN(1000)) * time.Millisecond
		time.Sleep(delay)
		log.Printf("⏱️ gameProcessorEngine iniciado após atraso de %v\n", delay)

		intervalMs := utils.MustEnvInt("PROCESSOR_INTERVAL_MS", 300)
		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			processExpiredRoomsEvents(context.Background(), rdb, registryRules)
		}
	}()
}

// processExpiredRoomsEvents verifica eventos em salas com timeout expirado e avança o estado do jogo.
func processExpiredRoomsEvents(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry) error {
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
		processRoomWithLock(ctx, rdb, registryRules, roomID)
	}
	return nil
}

func processRoomWithLock(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomID string) {
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

	if roomData.HasPendingPresentationEvent() {
		createGameEventFromPresentationEvent(ctx, rdb, registryRules, roomData)
		return
	}

	// Resolve próximo efeito, se existir
	if roomData.HasPendingLogicEffect() {
		resolveNextEffect(ctx, rdb, registryRules, roomData)
		scheduleImmediateProcessing(ctx, rdb, roomData)
		return
	}

	// Nada pendente → próximo turno
	if roomData.GameEvent == nil && !roomData.HasPendingLogicEffect() && !roomData.HasPendingPresentationEvent() {
		NextTurn(roomData, rdb, ctx)
	}
}

func createGameEventFromPresentationEvent(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room) {
	presentationEvent, ok := roomData.PopNextPendingPresentationEvent()
	if !ok {
		return
	}

	timeoutField := presentationEvent.TimeoutField
	if timeoutField == "" {
		timeoutField = "DisplayMessage"
	}

	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, timeoutField)
	if err != nil {
		utils.LogError(err)
		return
	}

	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()
	roomData.GameEvent = structs.NewGameEvent(presentationEvent.PlayerID, presentationEvent.Type, expiresAt, presentationEvent.Payload)

	if err := roomData.SaveRoom(ctx, rdb); err != nil {
		utils.LogError(err)
		return
	}

	var (
		confidencialRoomData *rooms.Room
		playersThatCanSee    []string
	)

	if presentationEvent.ConfidencialPayload != nil && len(presentationEvent.ConfidencialPlayerIds) > 0 {
		confidencialRoomData = roomData.Clone()
		confidencialRoomData.GameEvent.Payload = presentationEvent.ConfidencialPayload
		playersThatCanSee = append(playersThatCanSee, presentationEvent.ConfidencialPlayerIds...)
	}

	if err := roomData.SendUpdatedRoomData(ctx, rdb, confidencialRoomData, playersThatCanSee); err != nil {
		utils.LogError(err)
		return
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
}

func scheduleImmediateProcessing(ctx context.Context, rdb *redis.Client, roomData *rooms.Room) {
	if roomData.GameEvent != nil {
		return
	}
	if !roomData.HasPendingPresentationEvent() && !roomData.HasPendingLogicEffect() {
		return
	}

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(time.Now().UnixMilli()),
		Member: roomData.ID,
	})
}

func WaitingFirstAction(ctx context.Context, rdb *redis.Client, roomData *rooms.Room) error {
	expiresAt := time.Now().Add(15 * time.Second).UTC()
	roomData.GameEvent = structs.NewGameEvent("", structs.EventWaitingFirstAction, expiresAt, nil)

	if err := roomData.SaveRoom(ctx, rdb); err != nil {
		return err
	}

	if err := roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{}); err != nil {
		return err
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
	return nil
}

func NextTurn(roomData *rooms.Room, rdb *redis.Client, ctx context.Context) error {
	roomData.Turn++

	// limpa evento anterior
	roomData.GameEvent = nil

	// decrementa os rounds restantes para redução dos valores dobrados
	roomData.DecreaseDoubledCardValuesRounds()

	if err := roomData.SaveRoom(ctx, rdb); err != nil {
		return err
	}

	// inicia novo turno
	if err := WaitingFirstAction(ctx, rdb, roomData); err != nil {
		return err
	}
	return nil
}

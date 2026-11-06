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
	now := time.Now().UTC().UnixMilli()

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
	err := redisFuncs.WithRoom(ctx, rdb, roomID, func(roomData *rooms.Room) {
		// O lote de salas vencidas foi lido antes do mutex: se uma jogada abriu outra janela
		// nesse meio-tempo (ex.: contestação aos 9,9 s), ela ainda não venceu e não pode ser fechada.
		if roomData.GameEvent != nil && roomData.GameEvent.ExpiresAt.After(time.Now()) {
			return
		}

		ok, err := roomData.AcquireRoomLock(ctx, rdb, utils.GetInstanceID(), 2*time.Second)
		if err != nil || !ok {
			return
		}
		defer roomData.ReleaseRoomLock(ctx, rdb, utils.GetInstanceID())

		Tick(ctx, rdb, registryRules, roomData)
	})
	if err != nil {
		rdb.ZRem(ctx, "rooms:timeouts", roomID)
	}
}

// Tick fecha a janela atual e avança a sala um passo: próximo evento de apresentação,
// próximo efeito lógico ou próximo turno. Sem rdb, roda só em memória (testes de cenário).
func Tick(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room) {
	roomData.GameEvent = nil

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

	// Nada pendente → próximo turno (ou fim de jogo)
	if roomData.GameEvent == nil && !roomData.HasPendingLogicEffect() && !roomData.HasPendingPresentationEvent() {
		roomData.UpdateGameOver()
		if roomData.GameOver {
			if rdb != nil {
				rdb.ZRem(ctx, "rooms:timeouts", roomData.ID)
			}
			roomData.SaveRoom(ctx, rdb)
			roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{})
			return
		}
		NextRound(ctx, rdb, roomData, registryRules)
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

// scheduleImmediateProcessing verifica se a sala tem um evento de apresentação ou efeito lógico pendente e registra um timeout imediato para processamento.
func scheduleImmediateProcessing(ctx context.Context, rdb *redis.Client, roomData *rooms.Room) {
	if rdb == nil || roomData.GameEvent != nil {
		return
	}
	if !roomData.HasPendingPresentationEvent() && !roomData.HasPendingLogicEffect() {
		return
	}

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(time.Now().UTC().UnixMilli()),
		Member: roomData.ID,
	})
}

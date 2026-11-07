// Package runtime é a camada de produção (Redis/WebSocket) que dirige o motor puro.
// Fica fora do fecho do motor: o processador, o Sink de Redis, o lock de sala e as
// operações de lobby/conexão vivem aqui (ou em redisFuncs).
package runtime

import (
	"context"
	"log"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
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
// Drena lotes de salas vencidas até a fila esvaziar ou até o orçamento de tempo do
// tick (PROCESSOR_TICK_BUDGET_MS, padrão 200ms) acabar, para que um acúmulo de salas
// não espere um tick por lote.
func processExpiredRoomsEvents(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry) error {
	_, err := drainExpiredRooms(ctx, rdb, func(roomID string) {
		processRoomWithLock(ctx, rdb, registryRules, roomID)
	})
	return err
}

// drainExpiredRooms processa lotes de até PROCESSOR_BATCH_SIZE salas vencidas, em
// ordem de timeout, até a fila esvaziar ou o orçamento do tick acabar. PROCESSOR_TICK_BUDGET_MS
// <= 0 mantém o comportamento antigo (um lote por tick). Devolve quantas salas
// passaram pelo process.
func drainExpiredRooms(ctx context.Context, rdb *redis.Client, process func(roomID string)) (int, error) {
	batchSize := int64(utils.MustEnvInt("PROCESSOR_BATCH_SIZE", 5))
	if batchSize <= 0 {
		batchSize = 1
	}

	budgetMs := utils.MustEnvInt("PROCESSOR_TICK_BUDGET_MS", 200)
	if budgetMs < 0 {
		budgetMs = 0
	}
	deadline := time.Now().Add(time.Duration(budgetMs) * time.Millisecond)

	processed := 0
	for {
		if err := ctx.Err(); err != nil {
			return processed, err
		}

		now := time.Now().UTC().UnixMilli()
		roomIDs, err := rdb.ZRangeByScore(ctx, "rooms:timeouts", &redis.ZRangeBy{
			Min:    "0",
			Max:    strconv.FormatInt(now, 10),
			Offset: 0,
			Count:  batchSize,
		}).Result()
		if err != nil {
			utils.LogError(err)
			return processed, err
		}
		if len(roomIDs) == 0 {
			return processed, nil
		}

		for _, roomID := range roomIDs {
			if budgetMs == 0 {
				// Orçamento zerado: mantém o comportamento antigo (1 lote por tick).
				if processed >= int(batchSize) {
					return processed, nil
				}
			} else if time.Now().After(deadline) {
				return processed, nil
			}

			process(roomID)
			processed++
		}
	}
}

func processRoomWithLock(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomID string) {
	err := redisFuncs.WithRoom(ctx, rdb, roomID, redisFuncs.CauseTick, func(roomData *rooms.Room) bool {
		// O lote de salas vencidas foi lido antes do mutex: se uma jogada abriu outra janela
		// nesse meio-tempo (ex.: contestação aos 9,9 s), ela ainda não venceu e não pode ser fechada.
		if roomData.GameEvent != nil && roomData.GameEvent.ExpiresAt.After(time.Now()) {
			return false
		}

		engine.Tick(ctx, redisFuncs.NewSink(rdb), registryRules, roomData)
		return true
	})
	if err != nil {
		rdb.ZRem(ctx, "rooms:timeouts", roomID)
	}
}

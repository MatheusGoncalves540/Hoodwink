package redisHandlers

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/redis/go-redis/v9"
)

// ScheduleNextStep agenda a execução de um evento para o futuro, usando time.AfterFunc.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	roomID: identificador da sala
//	evt: evento a ser agendado (struct Event)
//
// Não retorna erro, apenas agenda a execução.
func ScheduleNextStep(ctx context.Context, rdb *redis.Client, roomID string, evt eventQueue.Event) {
	// Converte o timeout do evento para duração em milissegundos
	duration := time.Duration(evt.TimeoutMillis) * time.Millisecond
	// Agenda a função para adicionar o evento à fila após o timeout
	time.AfterFunc(duration, func() {
		_ = PushEvent(ctx, rdb, roomID, evt)
	})
}

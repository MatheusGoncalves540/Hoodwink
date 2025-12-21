package redisHandlers

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// ScheduleNextStep agenda a execução de um evento para o futuro, usando time.AfterFunc.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	RoomId: identificador da sala
//	playerPlay: evento a ser agendado (struct Event)
//
// Não retorna erro, apenas agenda a execução.
func ScheduleNextStep(ctx context.Context, rdb *redis.Client, RoomId string, playerPlay roomStructs.PlayerPlay) {
}
